# 数据存储设计

> **路径**：[`docs/design/data-schema.md`](data-schema.md)  
> **用途**：MySQL DDL 建表语句、Go 服务端内存结构体、Repository 分层设计、Redis Key 规范。  
> **关联文档**：[技术设计总纲](./tech-design.md)

---

## 1. 数据库设计 (MySQL 8.0+)

基于原版模型的改进版，保留表结构，调整主键类型和数据冗余点。

```sql
-- 全局配置
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 用户主表
-- 改进：UID 从 VARCHAR 改为 INT AUTO_INCREMENT
-- 改进：删除 TotalGames/Wins/Losses 冗余统计字段
CREATE TABLE users (
    id            INT UNSIGNED NOT NULL AUTO_INCREMENT,
    username      VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname      VARCHAR(50) NOT NULL,
    email         VARCHAR(255),
    phone         VARCHAR(20),
    bio           TEXT,
    rank  VARCHAR(5) NOT NULL COMMENT '棋力 18K-9D',
    elo   INT NOT NULL DEFAULT 1500 COMMENT 'Elo 分，用于匹配',
    avatar_url    TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at    DATETIME NULL DEFAULT NULL COMMENT '软删除',
    PRIMARY KEY (id),
    UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX idx_users_elo ON users(current_elo);

-- 用户偏好设置表（分表策略）
-- 改动说明：将不影响登录的核心偏好提取为一比一从表，减少主表负担，方便后期扩展如"常用招呼语"等。
CREATE TABLE user_preferences (
    user_id       INT UNSIGNED NOT NULL,
    dnd           TINYINT(1) NOT NULL DEFAULT 0 COMMENT '拒绝邀请开关（0=接受 1=拒绝）',
    theme         VARCHAR(20) NOT NULL DEFAULT 'light' COMMENT '界面主题',
    board_theme   VARCHAR(20) NOT NULL DEFAULT 'default' COMMENT '棋盘/棋子皮肤',
    greetings     JSON COMMENT '常用招呼语（预设或自定义）',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id),
    CONSTRAINT fk_user_prefs FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 规则预设表
CREATE TABLE game_rules (
    id         INT UNSIGNED NOT NULL AUTO_INCREMENT,
    is_custom  TINYINT(1) NOT NULL DEFAULT 0,
    name       VARCHAR(100) NOT NULL,
    rule_type  TINYINT NOT NULL DEFAULT 0 COMMENT '0=中国规则 1=日本规则',
    komi       DECIMAL(3,1) NOT NULL DEFAULT 7.5,
    handicap   TINYINT NOT NULL DEFAULT 0 COMMENT '0=分先 1=让先 2+=让子数',
    handicap_positions JSON COMMENT '让子位置数组',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 计时器预设表
CREATE TABLE game_timers (
    id            INT UNSIGNED NOT NULL AUTO_INCREMENT,
    is_custom     TINYINT(1) NOT NULL DEFAULT 0,
    name          VARCHAR(100) NOT NULL,
    timer_type    TINYINT NOT NULL DEFAULT 0 COMMENT '0=读秒制 1=包干制',
    base_time     INT NOT NULL DEFAULT 0 COMMENT '基础时间(秒)',
    byoyomi_time  INT NOT NULL DEFAULT 30 COMMENT '读秒时间(秒)',
    byoyomi_count INT NOT NULL DEFAULT 3 COMMENT '读秒次数',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 对局记录表
-- 改进：黑白方 FK 改为 INT
-- 改进：删除 Moves JSON 内联，改为终局生成 sgf_data
-- 改进：新增 config_snap JSON 快照
-- ⚠️ 重点契约：对局的核心配置事实（贴目、基本时间等）必须从 config_snap 中读取，以保证历史不可变。
-- rule_id 和 timer_id 仅作为预设配置的分类外键，对于用户高度自由自定义的非预设配置，直接将 rule_id 和 timer_id 置为 NULL，同时所有参数记录进 config_snap。详见 why.md 中的数据库决策。
CREATE TABLE games (
    id            INT UNSIGNED NOT NULL AUTO_INCREMENT,
    black_user_id INT UNSIGNED NOT NULL,
    white_user_id INT UNSIGNED NOT NULL,
    rule_id       INT UNSIGNED,
    timer_id      INT UNSIGNED,
    config_snap   JSON NOT NULL COMMENT '规则+计时参数快照',
    permit_undo   TINYINT(1) NOT NULL DEFAULT 0,
    is_finished   TINYINT(1) NOT NULL DEFAULT 0,
    end_type      VARCHAR(30) COMMENT 'counting|resignation|timeout|disconnection_forfeit',
    winner_id     INT UNSIGNED COMMENT 'NULL=未结束或平局',
    points_diff   DECIMAL(5,1) NOT NULL DEFAULT 0 COMMENT '未计贴目目差',
    score_diff    DECIMAL(5,1) NOT NULL DEFAULT 0 COMMENT '计贴目后分差(正=白胜 负=黑胜)',
    max_move_num  INT NOT NULL DEFAULT 0 COMMENT '总手数',
    sgf_data      MEDIUMTEXT COMMENT '终局生成的完整 SGF',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    CONSTRAINT fk_games_black FOREIGN KEY (black_user_id) REFERENCES users(id),
    CONSTRAINT fk_games_white FOREIGN KEY (white_user_id) REFERENCES users(id),
    CONSTRAINT fk_games_rule  FOREIGN KEY (rule_id)  REFERENCES game_rules(id),
    CONSTRAINT fk_games_timer FOREIGN KEY (timer_id) REFERENCES game_timers(id),
    CONSTRAINT fk_games_winner FOREIGN KEY (winner_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX idx_games_black ON games(black_user_id);
CREATE INDEX idx_games_white ON games(white_user_id);

-- 房间表
-- 改进：成员列表迁至 Redis，这里只存元数据
CREATE TABLE rooms (
    id              INT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id        INT UNSIGNED NOT NULL,
    description     TEXT,
    is_public       TINYINT(1) NOT NULL DEFAULT 1,
    password        VARCHAR(100),
    status          TINYINT NOT NULL DEFAULT 0 COMMENT '0=等待中 1=进行中 2=已结束',
    current_game_id INT UNSIGNED,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    CONSTRAINT fk_rooms_owner FOREIGN KEY (owner_id) REFERENCES users(id),
    CONSTRAINT fk_rooms_game  FOREIGN KEY (current_game_id) REFERENCES games(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;
```

### 1.1 数据安全与隐私合规 (Data Privacy & Security)
- **密码存储**：必须使用单向不可逆且包含盐值的哈希算法（如 `bcrypt`），严禁明文或使用单纯的 MD5 存储 `password_hash`。
- **PII (个人身份信息) 脱敏**：对于 `email` 和 `phone` 字段，落库前若无强检索需求，可考虑加密存储；日志打印、错误抛出及返回给非用户本人的前端 API 响应中，必须执行脱敏（如 `138****0000` 掩码）。

---

## 2. 服务端内存存储 (Go Structs & Ownership)

运行时热路径数据，存放在 `internal/ws` 或直接由 `Hub`/`Room` 对象持有。各字段定义对齐 [business-model.md](../requirement/business-model.md)。

### 2.1 用户会话 (UserSession)

对应业务模型 §1.1 User，描述用户在服务端的**运行时状态**。

*   **所有权**: `ws.Hub` 维护 `map[int]*UserSession`。
*   **生命周期**: WS 连接注册时创建；明确注销时删除（断线宽限期期间保留，宽限期结束仍未重连则删除）。
*   **读写权限**:
    *   `Hub.Run()`: 写（Status 转换、ActiveGameID 记录）。
    *   `Client`: 写（LastPulse 心跳续命）。
    *   `Room.Run()`: 读（广播时获取 Nickname/Elo）。

```go
// UserSession 对应业务模型 User 的运行时快照。
// 业务规则"单一参与规则"（§4）体现在 ActiveGameID：
//   - 0 表示不在任何对局中（空闲或观战）
//   - >0 表示当前对局者角色，此时系统拒绝再次作为对局者加入新对局
type UserSession struct {
    UserID       int
    Nickname     string
    Elo          int
    Status       string    // "idle" | "gaming" | "dnd"，对应业务模型用户状态机
    ActiveGameID int       // 当前参与对局的 GameInstance ID，0 = 未参与（观战不算）
    LastPulse    time.Time // 心跳时间，心跳超时则标记断线
}
```

### 2.2 房间运行态 (MemRoom)

对应业务模型 §1.2 Room。

*   **所有权**: `ws.Hub` 维护 `map[int]*MemRoom` 索引。
*   **生命周期**: 房主建房时创建；房间关闭（最后一人离开 or 房主主动关闭）时销毁。
*   **读写权限**:
    *   `Hub.Run()`: 写（添加/移除 Members）、读（可见性查询）。
    *   `Room.Run()`: 写（Config 协商更新、Status 切换、Instance 激活）、读（规则校验）。

```go
// MemRoom 对应业务模型 Room。
// 业务约束：
//   - Status="waiting"  时，所有成员 Role="spectator"（§1.2）
//   - Status="ongoing"  时，新加入者自动赋予 Role="spectator"（§1.2）
//   - Instance 非 nil 当且仅当 Status="ongoing"（§1.4）
type MemRoom struct {
    ID         int
    OwnerID    int              // 引用用户 ID（§1.2 房主）
    IsPublic   bool             // 公开 / 私密（§1.2 可见性）
    Status     string           // "waiting" | "ongoing"（§1.2 状态机）
    Members    map[int]*Member  // UserID -> Member（§1.2 成员集合）
    Config     *GameConfig      // 当前协商中的配置契约（§1.3）
    Instance   *GameInstance    // Status="ongoing" 后非 nil（§1.4）
}

// Member 对应业务模型 §1.2 成员，引用用户 ID + 房间内角色。
type Member struct {
    UserID int
    Role   string    // "spectator"（等待中） | "black" | "white"（对局激活后）
}
```

### 2.3 对局配置契约 (GameConfig)

对应业务模型 §1.3 Game Configuration。在协商阶段由 `Hub` 转发流转，达成一致后作为激活 `GameInstance` 的唯一入参。

```go
// GameConfig 对应业务模型 §1.3 的所有"内容属性"。
type GameConfig struct {
    BoardSize       int     `json:"board_size"`         // 棋盘路数（如 19/13/9/N），异形棋盘P2后再考虑支持
    RuleType        int     `json:"rule_type"`          // 规则标准（§1.3）: 0=中国 1=日本
    Komi            float64 `json:"komi"`               // 贴目（§1.3 补偿方案）
    TimeSystem      string  `json:"time_system"`        // 计时制度（§1.3）: "byoyomi" | "absolute"
    MainTimeSeconds int     `json:"main_time_seconds"`
    ByoyomiPeriods  int     `json:"byoyomi_periods"`
    ByoyomiSeconds  int     `json:"byoyomi_seconds"`
    FirstBlackID    int     `json:"first_black_id"`     // 指定黑方（§1.3 初始执子），0=随机
}
```

### 2.4 对局运行态 (GameInstance)

对应业务模型 §1.4 Game Instance。

*   **所有权**: 被对应 `MemRoom.Instance` 强引用。`GameInstance` 不冗余存储黑白方 UserID，通过 `MemRoom.Members` 按 Role 查询以保证一致性。
*   **生命周期**: 协商达成一致时创建（§3.1 锁定步骤）；终局（认输/超时/点目完成）后落库 `games` 表并销毁。
*   **读写权限**: `Room.Run()` 全权读写。

```go
// GameInstance 对应业务模型 §1.4。
// 不存黑白方 UserID，从 MemRoom.Members 里按 Role 获取，避免数据双写。
// 对局状态的两态（进行中/已结束，§1.4）通过 MemRoom.Status 体现，GameInstance 本身只存"进行中"快照。
type GameInstance struct {
    GameID     int             // 对应 games 表的 ID，落库时用
    Config     GameConfig      // 本局生效的配置快照
    Board      [19][19]int8    // 棋盘快照（§1.4 竞技数据）: 0=空 1=黑 2=白
    StepSeq    int             // 当前手数（§1.4 指令序列）
    LastMoveAt time.Time       // 最近一次有效落子时间，用于超时判断（§3.3）
    Timer      *GameTimer      // 计时容器（§1.4），封装读秒/包干逻辑
}
```

---

## 3. Repository 层分层设计

Redis 有两种语义不同的用法，通过目录分离避免混淆：

```
repository/
  model/          ← GORM 数据库模型（users, games, rooms ...）
  db/             ← MySQL 交互（只在此层使用 GORM）
    user.go
    room.go
    game.go
  store/          ← Redis 作为主存储（Source of Truth，无 DB 对应，无 TTL 或由心跳控制）
    room_store.go    ← 房间 Hash：成员列表、房间状态、配置快照
    user_state.go    ← 用户三态（idle/gaming/dnd）及所在房间
    game_moves.go    ← 实时落子流 List（对局结束后落库 SGF，Redis 数据清除）
  cache/          ← Redis 作为缓存（DB 数据加速副本，有 TTL，Cache Miss 回退 DB）
    user_cache.go    ← 大厅列表渲染用：缓存昵称、Elo、段位（TTL: 5min）
```

**Store vs Cache 判断原则**：
- DB 里是否有对应表？有 → Cache；无 → Store。
- 数据丢失是否可接受？可接受（瞬态） → Store；不可接受 → Cache + DB 双写。

### Redis Key 设计

```
# Store 层（无 TTL，由业务生命周期控制）
user_state:{user_id}          Hash   {status, room_id, last_active}
room:{room_id}                Hash   {owner_id, status, members_json, config_json}
game_moves:{room_id}          List   [{step,type,x,y,user_id,timestamp}, ...]

# Cache 层（需设 TTL）
user_profile:{user_id}        Hash   {nickname, elo, rank}   TTL: 5min
```
