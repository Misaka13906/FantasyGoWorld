# 架构时序图与流程图 (Diagrams)

> **路径**：[`docs/design/diagrams.md`](diagrams.md)  
> **用途**：项目核心业务流程的可视化呈现，包括前后端交互、分层调用和关键状态机转换。  
> **关联文档**：[技术设计总纲](./tech-design.md) · [前端设计](./frontend-design.md) · [WS 架构](./ws-design.md)

---

## 1. 认证流：登录与 Token 自动刷新

为了保证用户体验，我们采用 `access_token` (短期) + `refresh_token` (长期) 的双 Token 方案。前端 Axios 拦截器须实现无感刷新。

```mermaid
sequenceDiagram
    participant User as 用户
    participant FE as 前端 (Axios/App)
    participant BE as 后端 (Auth Middleware)
    participant DB as 数据库/Redis

    Note over User, DB: 登录阶段
    User->>FE: 输入用户名密码
    FE->>BE: POST /auth/login
    BE->>DB: 验证用户
    BE-->>FE: HTTP 200 (access_token in JSON, refresh_token in HttpOnly Cookie)
    FE->>FE: 写入 AuthStore

    Note over User, DB: 业务请求 & Token 过期处理
    FE->>BE: GET /api/v1/user/list (带 access_token)
    BE->>BE: 验证过期?
    BE-->>FE: HTTP 401 (token_expired)
    
    Note right of FE: 触发 Axios 响应拦截器
    FE->>BE: POST /auth/refresh (带 Cookie)
    BE->>BE: 验证 refresh_token
    BE-->>FE: HTTP 200 (new access_token)
    FE->>FE: 更新 AuthStore.token
    
    Note right of FE: 使用新 Token 自动重试刚才失败的请求
    FE->>BE: GET /api/v1/user/list (new access_token)
    BE-->>FE: HTTP 200 (Success)
```

---

## 2. 对局全生命周期

从大厅进入房间，到对局结束并记录 SGF 的全套逻辑。

```mermaid
sequenceDiagram
    participant P1 as 玩家 A (房主)
    participant P2 as 玩家 B
    participant Hub as 服务端 Hub
    participant Room as 服务端 Room
    participant DB as MySQL/Redis/Engine

    Note over P1, DB: 准备阶段
    P1->>Hub: WS: JOIN_ROOM (RoomId: 101)
    Hub->>Room: register <- P1
    Room-->>P1: WS: MEMBER_UPDATE

    P2->>Hub: WS: JOIN_ROOM (RoomId: 101)
    Hub->>Room: register <- P2
    Room-->>P1: WS: MEMBER_UPDATE (P2 加入)
    Room-->>P2: WS: MEMBER_UPDATE (同步成员列表)

    Note over P1, P2: 协商阶段 (Proposal)
    P1->>Hub: WS: PROPOSAL {action:propose, target:P2, config:19路/中国规则}
    Hub-->>P2: WS: PROPOSAL {action:propose, from:P1}
    P2->>Hub: WS: PROPOSAL {action:counter, config:13路/中国规则}
    Hub-->>P1: WS: PROPOSAL {action:counter, from:P2}
    P1->>Hub: WS: PROPOSAL {action:accept}
    Hub->>Hub: 为所有对局方 Client 注入 roomInbound（先于广播，消除竞态）
    Hub->>DB: INSERT games（初始化对局记录）
    Hub-->>P1: WS: GAME_START {black, white, config}
    Hub-->>P2: WS: GAME_START {black, white, config}

    Note over P1, P2: 对局阶段 (MOVE 流，绕过 Hub 直达 Room)
    P1->>Room: WS: MOVE {x, y, seq:1}
    Room->>DB: engine.ValidateMove
    DB-->>Room: NewBoard + captures
    Room->>DB: Redis RPUSH game_moves
    Room-->>P1: WS: SYNC_BOARD {seq:2}
    Room-->>P2: WS: SYNC_BOARD {seq:2}

    Note over P1, P2: 终局结算
    P2->>Room: WS: RESIGN
    Room->>DB: repository.SaveGame（同步写库，含 SGF 生成）
    Room->>DB: Redis DEL game_moves（写库成功后清理）
    Room-->>P1: WS: GAME_END {winner:P1}
    Room-->>P2: WS: GAME_END {winner:P1}
    Room->>Hub: 通知销毁 GameInstance
```

---

## 3. 连接与对局状态机

### 3.1 客户端状态机（前端视角）

描述用户在浏览器内的页面/连接状态流转。

```mermaid
stateDiagram-v2
    [*] --> LOBBY: 登录成功，WS 建立
    LOBBY --> ROOM_WAITING: JOIN_ROOM 成功

    ROOM_WAITING --> GAMING: 收到 GAME_START
    ROOM_WAITING --> RECONNECTING: WS 断开
    RECONNECTING --> ROOM_WAITING: 重连成功（等待阶段无保护期，直接回来）

    GAMING --> SCORING: 双方 PASS
    GAMING --> LOBBY: 收到 GAME_END（认输/超时/解散）
    GAMING --> RECONNECTING_GAME: WS 断开

    RECONNECTING_GAME --> GAMING: 重连成功 + 收到 SYNC_BOARD（在保护期内）
    RECONNECTING_GAME --> LOBBY: 重连成功 + 收到 GAME_END（保护期已过，已判负）

    SCORING --> LOBBY: 确认结果
```

---

### 3.2 服务端状态机（UserSession / 断线保护视角）

描述服务端对单个用户会话生命周期的管理。

```mermaid
stateDiagram-v2
    [*] --> ONLINE_IDLE: WS 握手成功，注册 UserSession

    ONLINE_IDLE --> ONLINE_GAMING: 对局激活（GAME_START 发出）
    ONLINE_IDLE --> OFFLINE_CLEAN: WS 断开（非对局中）
    OFFLINE_CLEAN --> [*]: 直接清理 UserSession，无保护期

    ONLINE_GAMING --> GRACE_PERIOD: WS 断开（对局中）
    note right of GRACE_PERIOD
        暂停断线方计时器
        通知对方"对手断线"
        启动 180s 倒计时
    end note

    GRACE_PERIOD --> ONLINE_GAMING: 保护期内重连（发送 SYNC_BOARD，恢复计时器）
    GRACE_PERIOD --> FORFEITED: 180s 超时
    FORFEITED --> [*]: 触发弃权负（GAME_END），落库，清理 UserSession
```


---

## 4. 对局协商流程（PROPOSAL 多轮）

```mermaid
sequenceDiagram
    participant A as 用户A (Browser)
    participant Room as 服务端 Room
    participant B as 用户B (Browser)

    Note over A,B: 前提：A 和 B 均已 JOIN_ROOM，roomInbound 已注入

    A->>Room: PROPOSAL {action:propose, target:B, config:...}
    Room-->>B: PROPOSAL {action:propose, from:A, config:...}

    alt B 接受
        B->>Room: PROPOSAL {action:accept, target:A}
        Room->>Room: 激活 GameInstance，分配黑白方，Status="ongoing"
        Room-->>A: GAME_START {black, white, config}
        Room-->>B: GAME_START {black, white, config}
        Room-->>All: MEMBER_UPDATE {roles updated}
        Room-)Hub: GameActivatedEvent（异步，更新 UserSession.ActiveGameID）
    else B 反向修改（counter）
        B->>Room: PROPOSAL {action:counter, target:A, config:modified}
        Room-->>A: PROPOSAL {action:counter, from:B, config:modified}
        Note over A,B: 可多轮往返，直到一方 accept 或 reject
    else B 拒绝
        B->>Room: PROPOSAL {action:reject, target:A}
        Room-->>A: NOTIFICATION {type:proposal_rejected}
    end
```

---

## 5. 落子流程（MOVE → 规则引擎 → 广播）

```mermaid
sequenceDiagram
    participant P as 黑方 (Player)
    participant R as Room.Run()
    participant E as 规则引擎
    participant All as 全体成员

    P->>R: MOVE {x:3, y:4, step:5}
    R->>E: ValidateMove(board, x, y, step)
    alt 合法
        E-->>R: NewBoard + captures
        R->>R: 追加 game_moves Redis List
        R-->>All: SYNC_BOARD {board, next_turn, timers, seq:6}
    else 非法（禁着/劫争/乱序）
        E-->>R: ErrIllegalMove
        R-->>P: ERROR {code, reason}
        Note over P: 前端 rollbackMove() 还原预渲染
    end
```

---

## 6. 断线保护与恢复

```mermaid
sequenceDiagram
    participant P as 断线方
    participant S as Server
    participant O as 对方

    P-xS: (连接断开)
    S->>S: 3s 无心跳 → 标记离线
    S-->>O: NOTIFICATION {type:opponent_disconnected}
    S->>S: 暂停 P 的计时器，启动 180s 保护倒计时

    alt 保护期内重连
        P->>S: WS 握手 + JWT
        S->>S: 验证 JWT，恢复 UserSession
        S-->>P: SYNC_BOARD (完整快照)
        S->>S: 恢复 P 的计时器
        S-->>O: NOTIFICATION {type:opponent_reconnected}
    else 超过 180s 仍未重连
        S->>S: 触发弃权负，生成 GAME_END
        S-->>O: GAME_END {end_type:disconnection_forfeit, winner:O}
        S->>S: 写 games 表，生成 SGF，清理 Redis
    end
```

