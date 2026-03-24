# 接口设计 (API Specification)

> **路径**：[`docs/design/api-spec.md`](api-spec.md)  
> **用途**：REST API 全量接口表、WS 消息类型一览、Payload 结构定义、鉴权 Middleware 方案、API错误码。  
> **关联文档**：[技术设计总纲](./tech-design.md) · [WS 架构](./ws-design.md)

---

## 1. REST API

基础路径：`/api/v1`。所有需要鉴权的接口须在 Header 携带 `Authorization: Bearer <access_token>`。

### 鉴权 Middleware

浏览器原生 **`WebSocket` API 不支持自定义 Header**，无法在 WS Upgrade 请求里携带 `Authorization`。  
统一解决方案：登录时将 `access_token` **同时写入 `access_token` Cookie**（`SameSite=Lax`，非 httpOnly 以允许 JS 读取用于 REST 请求），WS Upgrade 时浏览器自动携带 Cookie，Middleware 优先读 Header，其次 fallback 到 Cookie。

> **前端策略**：REST 请求手动在 Header 附带 token；WS 升级直接 `new WebSocket(url)`，不传任何参数，浏览器自动附带 Cookie 完成鉴权。WS 地址不再需要 `?token=xxx`。具体实现见 [附录 A](#附录-a-鉴权-middleware-实现)。


**统一响应格式**：
```json
{ "code": 20000, "msg": "success",          "data": {} }
{ "code": 40001, "msg": "参数不满足要求",    "data": null }
{ "code": 50001, "msg": "internal server error", "data": null }
```

### 分页与过滤标准 (Pagination & Filtering)
- **统一参数**：所有列表类接口（`/list`）统一使用 `page` (从1开始) 和 `page_size` (默认 20，最大 100) 参数。
- **响应结构**：分页响应必须包含 `total`, `page`, `page_size` 及 `items` 数组。

### 幂等性与安全 (Idempotency)
- **安全方法**：`GET`, `HEAD`, `OPTIONS` 必须是安全的（无副作用）。
- **幂等方法**：`PUT`, `DELETE` 必须是幂等的。对于非幂等的 `POST`（如创建房间），客户端应考虑在 Header 携带 `X-Idempotency-Key`（UUID）以防止重复提交。

**注意**：所有业务错误码在 `pkg/e/` 中统一定义，Agent 遇到新错误场景必须在此注册，同时更新本文档**附录 A** 中的错误码和代码中的错误码，不得自行造码。

### 认证模块 `/auth`

| 方法 | 路径 | 说明 | Auth |
|:---|:---|:---|:---|
| `POST` | `/auth/register` | 注册（username/password/nickname/rank）| 否 |
| `POST` | `/auth/login` | 登录，返回 access_token；refresh_token 写 httpOnly Cookie | 否 |
| `POST` | `/auth/refresh` | 用 Cookie 中的 refresh_token 换新 access_token | 否 |
| `POST` | `/auth/logout` | 清除 refresh_token Cookie | 是 |

> **Token 策略**：access_token 有效期 1h，refresh_token 有效期 7d（存 httpOnly Cookie，防 XSS 读取）。

### 用户模块 `/user`

| 方法 | 路径 | 说明 | Auth |
|:---|:---|:---|:---|
| `GET` | `/user/:uid` | 获取用户公开信息（昵称/段位/战绩）| 是 |
| `PUT` | `/user` | 更新个人资料（nickname/bio/avatar_url）| 是 |
| `GET` | `/user/preferences` | 获取当前用户偏好（dnd/theme/greetings）| 是 |
| `PUT` | `/user/preferences` | 更新当前用户偏好 | 是 |
| `GET` | `/user/list` | 获取在线用户列表（分页，供大厅初始化）| 是 |
| `GET` | `/user/search` | 按用户名搜索 `?username=` | 是 |
| `GET` | `/user/:uid/games` | 获取用户历史对局列表（分页）| 是 |

### 房间模块 `/room`

| 方法 | 路径 | 说明 | Auth |
|:---|:---|:---|:---|
| `POST` | `/room` | 创建房间（description/is_public/password）| 是 |
| `GET` | `/room/list` | 获取公开房间列表（供大厅初始化）| 是 |
| `GET` | `/room/:roomId` | 获取单个房间信息 | 是 |
| `DELETE` | `/room/:roomId` | 关闭/解散房间（仅房主）| 是 |

### 对局记录模块 `/game`

| 方法 | 路径 | 说明 | Auth |
|:---|:---|:---|:---|
| `GET` | `/game/:gameId` | 获取对局详情（结算数据 + config_snap）| 是 |
| `GET` | `/game/:gameId/sgf` | 下载 SGF 文件 | 是 |

### WebSocket 升级入口

| 路径 | 说明 |
|:---|:---|
| `GET /ws` | 统一 WS 入口，升级后通过消息区分大厅/房间上下文 |

> **认证方式**：WS Upgrade 走上述 Middleware 的 Cookie Fallback 路径，service 端统一处理，无需 Query 参数。验证失败时 HTTP 握手返回 `401`，不升级连接。

---

## 2. WebSocket 消息协议

### 2.1 基础报文格式 (Envelope)

```json
{
  "type":      "STRING",     // 消息类型，如 "MOVE", "PROPOSAL", "CHAT"
  "seq":       100,          // 序列号，用于前端排序和防乱序
  "room_id":   "UUID",       // 所属房间（大厅消息则为 "hall"）
  "payload":   {},           // 具体业务数据
  "timestamp": 1709000000000
}
```

> **为什么 `type` 字段使用 `UPPER_SNAKE_CASE`（全大写）？**
> 1. **区分指令与数据**：在 JSON 中，普通数据字段通常是 `snake_case`（如 `room_id`），把 `type` 定义为全大写能一眼看出它是一个"动作指令 (Action/Event)"，借鉴了 Redux Action 的设计惯例。
> 2. **日志可读性**：在后端控制台和浏览器 Network 面板里追 WS 满屏的日志流时，全大写的 `JOIN_ROOM` 具有极高的视觉辨识度。
> 
> *注意：这指的是 JSON 传输中的字符串值。在 Go 后端代码里，常量名依然遵循 Go 的驼峰规范，即 `const MsgTypeJoinRoom = "JOIN_ROOM"`。*

### 2.2 完整消息类型一览

| type | 方向 | 路由 | 说明 |
|:---|:---|:---|:---|
| `HEARTBEAT` | C→S | Hub | 心跳（每 1s 一次，无 payload）|
| `JOIN_ROOM` | C→S | Hub | 进入房间 |
| `LEAVE_ROOM` | C→S | Hub | 离开房间 |
| `INVITE` | C→S | Hub | 大厅邀请另一用户 |
| `INVITE_REPLY` | C→S | Hub | 被邀请方回复（accept \| reject）|
| `PROPOSAL` | C→S / S→C | Hub | 对局配置提议（协商阶段）|
| `MOVE` | C→S | Room | 落子 |
| `PASS` | C→S | Room | 停手 |
| `RESIGN` | C→S | Room | 认输 |
| `CHAT` | C→S | Room | 聊天消息 |
| `SYNC_BOARD` | S→C | Room broadcast | 棋盘状态同步（落子确认/快照）|
| `GAME_START` | S→C | Room broadcast | 对局激活通知 |
| `GAME_END` | S→C | Room broadcast | 终局结算通知 |
| `MEMBER_UPDATE` | S→C | Room broadcast | 成员列表变更 |
| `LOBBY_UPDATE` | S→C | Hub broadcast | 大厅用户/房间列表增量推送 |
| `NOTIFICATION` | S→C | Hub unicast | 系统通知（邀请/提议被拒等）|
| `ERROR` | S→C | Hub/Room unicast | 业务错误提示（劫争违规等）|

### 2.3 所有消息 Payload 结构 (TypeScript)

对于 WebSocket 来说，比起臃肿的 AsyncAPI 标准规范，**TypeScript Interface** 是业内公认最清晰、前端消费最直接的结构化文档方式。

```typescript
// --- C->S (客户端发给服务端) ---

// JOIN_ROOM / LEAVE_ROOM / HEARTBEAT / PASS / RESIGN payload
// 此类消息没有额外 payload 数据结构，仅改变外层 Envelope 的 type (与 room_id)
null | {}

// INVITE payload (大厅邀请挑战)
{
  target_uid: number,
  room_id:    number    // 自己创建好但在"等待中"的房间
}

// INVITE_REPLY payload
{
  invite_uid: number,   // 给你发邀请的人的 UID
  accept:     boolean   // true=同意并加入房间, false=拒绝
}

// PROPOSAL payload (协商双向流转，C->S 及 S->C 相同)
{
  action:     'propose' | 'accept' | 'reject' | 'counter',
  target_uid: number,   // 接收方 UID
  config: {
    board_size:        number,   // 棋盘路数（如 19/13/9）。支持任意数字（如5-25），异形棋盘晚于P2迭代。
    komi:              number,   // 7.5 | 6.5
    rule_type:         0 | 1,    // 0=中国 1=日本
    time_system:       'byoyomi' | 'absolute',
    main_time_seconds: number,
    byoyomi_periods:   number,
    byoyomi_seconds:   number,
    plays_black_id:    number    // 0=随机, 否则为指定的 UID
  }
}

// MOVE payload
{ 
  x: number, 
  y: number, 
  step: number          // 客户端视角的手数，用于服务端乱序防抖
}

// CHAT payload
{
  content: string       // 纯文本或者特定格式的表情包码
}


// --- S->C (服务端发给客户端) ---

// SYNC_BOARD payload (房间内快照同步)
{
  board:     number[][],        // [size][size] 二维数组, 0=空 1=黑 2=白
  next_turn: 'black' | 'white',
  last_move: { x: number, y: number, step: number } | null,
  timers: {
    black: { remaining_ms: number, byoyomi_count: number },
    white: { remaining_ms: number, byoyomi_count: number }
  },
  seq: number                   // 单调递增序列号，防乱序
}

// GAME_START payload
{
  game_id:   number,
  black_uid: number,
  white_uid: number,
  config:    any                // 这里的 config 结构同 PROPOSAL payload 的 config
}

// GAME_END payload
{
  winner_uid: number | null,    // null = 平局
  end_type:   'counting' | 'resignation' | 'timeout' | 'disconnection_forfeit',
  score_diff: number,
  elo_delta:  { black: number, white: number }
}

// MEMBER_UPDATE payload (房间内成员变动同步)
{
  members: Array<{
    uid: number,
    nickname: string,
    elo: number,
    status: 'player' | 'spectator' // 是对局者还是观战者
  }>
}

// LOBBY_UPDATE payload (大厅级别的增量推送)
{
  users_add:    OnlineUserEntry[],
  users_remove: number[],       // uid 列表
  rooms_add:    RoomEntry[],
  rooms_remove: string[]        // room_id 列表
}

// NOTIFICATION payload (大厅级别的系统通知、被拒通知等)
{
  type:    'info' | 'warning' | 'invite_rejected',
  message: string
}

// ERROR payload (各类操作越权检查、棋盘规则违例)
{
  code:    number,              // 如 40304 (打劫禁手)
  message: string               // 给前端直接 Toast 的文本提示
}
```

---
## 附录

### 附录 A：错误码注册表

所有业务错误码在 `pkg/e/` 中统一定义，Agent 遇到新错误场景必须在此注册，同时更新文档和代码，不得自行造码。

| 错误码 | HTTP 状态码 | message | 触发场景 | 例子 |
|:---|:---|:---|:---|:---|
| **200xx 成功** | | | |
| `20000` | 200 | success | 正常成功 |
| `20001` | 200 | no records found | 查询结果为空（非错误） | 查询好友列表为空 |
| **400xx 客户端错误** | | | |
| `40001` | 400 | request does not satisfy requirements | 业务前置条件不满足（如房间已在对局中） |
| `40002` | 400 | request field error | 参数格式/类型错误（binding 失败） |
| `40100` | 401 | unauthorized | 未携带 token 或 token 解析失败 |
| `40101` | 401 | token expired | access_token 已过期（前端应自动刷新）|
| `40102` | 401 | refresh token expired | refresh_token 已过期（需重新登录）|
| `40300` | 403 | forbidden | 无权限（如非房主关闭房间）|
| `40400` | 404 | resource not found | 请求的资源不存在（user/room/game）|
| `40900` | 409 | conflict | 资源冲突（如用户名已注册）|
| **401xx 认证模块** | | | |
| `40110` | 400 | password too short | 密码长度不足 |
| `40111` | 400 | username already exists | 用户名已被注册 |
| `40112` | 400 | invalid credentials | 用户名或密码错误 |
| **402xx 房间模块** | | | |
| `40200` | 400 | room is full | 房间人数已满 |
| `40201` | 400 | room not in waiting state | 房间已开始对局，无法进行该操作 |
| `40202` | 403 | not room owner | 非房主，无权执行此操作 |
| `40203` | 400 | wrong room password | 房间密码错误 |
| **403xx 对局/规则引擎** | | | |
| `40300` | 400 | game not active | 对局未激活 |
| `40301` | 400 | not your turn | 非当前行棋方 |
| `40302` | 400 | illegal move: occupied | 目标交叉点已有棋子 |
| `40303` | 400 | illegal move: suicide | 禁止自杀 |
| `40304` | 400 | illegal move: ko | 打劫禁手 |
| `40305` | 400 | invalid step number | 手数不连续（乱序检测）|
| `40306` | 400 | invalid position | 坐标越界 |
| **500xx 服务端错误** | | | |
| `50001` | 500 | internal server error | 未预期的服务端错误 |
| `50002` | 500 | database error | 数据库操作失败 |
| `50003` | 500 | redis error | Redis 操作失败 |


## 附录 B 鉴权 Middleware 实现

```go
// middleware/auth.go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 优先从 Authorization Header 读取（REST API 场景）
        raw := c.GetHeader("Authorization")
        if strings.HasPrefix(raw, "Bearer ") {
            raw = strings.TrimPrefix(raw, "Bearer ")
        }
        // 2. Fallback：从 Cookie 读取（WebSocket Upgrade 场景）
        if raw == "" {
            raw, _ = c.Cookie("access_token")
        }
        if raw == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized,
                gin.H{"code": 40100, "msg": "unauthorized"})
            return
        }
        claims, err := jwt.Parse(raw, keyFunc)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized,
                gin.H{"code": 40101, "msg": "invalid token"})
            return
        }
        c.Set("uid", claims.UserID)
        c.Next()
    }
}
```
