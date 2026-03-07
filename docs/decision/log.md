# 架构决策日志 (Decision Log)

> **路径**：[`docs/decision/log.md`](log.md)  
> **用途**：面向 AI Agent 的决策记录索引，汇总所有架构选型和改进决策及其原因，防止 Agent 重复提出已否决的方案。

---

## ADR-01 · 主键从 VARCHAR 改为 BIGINT AUTO_INCREMENT

**文件**：[`docs/design/data-schema.md`](../design/data-schema.md)  
**原版**：`users.uid VARCHAR(20)`  
**决策**：改为 `BIGINT UNSIGNED AUTO_INCREMENT`，直接作为对外用户 ID（洛谷/CF 风格）  
**原因**：
- VARCHAR 主键导致 B-Tree 索引体积膨胀，JOIN 需要逐字节字符串比较，外键列全部变宽
- 用户量级不需要分布式主键，AUTO_INCREMENT 性能最优、实现最简单
- 注册量不敏感，游戏社区平台直接暴露数字 UID 反而对用户友好

**否决方案**：UUID v4（随机插入页分裂，性能最差）；UUID v7（秒内非严格递增）；雪花 ID（初期单节点无需 worker ID 协调）

---

## ADR-02 · 删除 Moves[]Move 内联 JSON，改为终局 SGF

**文件**：[`docs/design/data-schema.md`](../design/data-schema.md)  
**原版**：`games.moves JSONB` 实时追加落子  
**决策**：Redis List (`game_moves:{room_id}`) 实时存落子流，对局结束后一次性生成 `sgf_data MEDIUMTEXT` 写入 DB  
**原因**：
- 每次落子更新整列 JSONB 产生大量写放大
- Redis List RPUSH 是 O(1)，对局期间零 DB 写压力
- SGF 是业界标准格式，便于回放和导出，无需自定义 Move 序列化

---

## ADR-03 · 新增 config_snap JSON 快照

**文件**：[`docs/design/data-schema.md`](../design/data-schema.md)  
**决策**：`games` 表保留 `rule_id / timer_id` 外键的同时新增 `config_snap JSON`  
**原因**：预设规则表内容可能随版本更新，若历史对局只存外键，回放时参数会随预设修改而漂移。快照在写入时固定参数，保证历史记录不可变性。

---

## ADR-04 · Redis 双语义分层（store vs cache）

**文件**：[`docs/design/data-schema.md`](../design/data-schema.md)  
**决策**：
- `store/`：无 DB 对应的瞬态主存储（房间成员、用户在线状态、实时落子流），无 TTL
- `cache/`：有 DB 对应的加速缓存（用户名片，TTL: 5min），Cache Miss 回退查 DB

**判断原则**：
- DB 里是否有对应表？有 → `cache`；无 → `store`
- 数据丢失是否可接受？可接受（进程重启后可重建）→ `store`；不可接受 → `cache + DB 双写`

---

## ADR-05 · WS 消息按 type 分流，游戏动作绕过 Hub 直投 Room

**文件**：[`docs/design/ws-design.md`](../design/ws-design.md)  
**原版问题**：所有 WS 消息都经 Hub 路由，Room 收不到落子消息  
**决策**：`Client.ReadPump` 按 type 判断路由目标（`roomBoundTypes` 白名单）；Hub 在 `JOIN_ROOM` 时将 `Room.Inbound` 注入 `Client.roomInbound`，之后游戏期间消息完全绕过 Hub  
**收益**：Hub 从热路径上移除，对局延迟降低；Hub 和 Room 职责清晰分离

---

## ADR-06 · access_token 只存内存，不写 localStorage

**文件**：[`docs/design/frontend-design.md`](../design/frontend-design.md)  
**决策**：前端 `AuthStore.token` 仅在 JS 内存中持有，页面刷新后通过 refresh_token（httpOnly Cookie）静默换取新 token  
**原因**：localStorage 可被同域 XSS 读取；httpOnly Cookie 无法被 JS 访问，安全边界更清晰。代价是页面刷新后需要一次 `/auth/refresh` 请求（< 100ms，体验无感知）

---

## ADR-07 · 数据库从 PostgreSQL 改为 MySQL 8.0+

**文件**：[`docs/design/data-schema.md`](../design/data-schema.md)  
**决策**：使用 MySQL 8.0+，同步调整语法：`SERIAL` → `INT UNSIGNED AUTO_INCREMENT`，`JSONB` → `JSON`，`BOOLEAN` → `TINYINT(1)`，外键改为独立 `CONSTRAINT` 声明  
**原因**：用户主动选择，MySQL 生态工具链更熟悉，GORM 对 MySQL 支持完善

---

## ADR-08 · WS 鉴权改为 Cookie Fallback，不用 Query 参数

**文件**：[`docs/design/api-spec.md`](../design/api-spec.md)  
**问题**：浏览器原生 WebSocket API 不支持自定义 Header，无法携带 `Authorization`  
**决策**：登录时将 `access_token` 同时写入 `SameSite=Lax` Cookie；Gin Middleware 优先读 `Authorization` Header，空则 fallback 读 Cookie；WS 地址无需 `?token=xxx`  
**否决方案**：Query 参数（token 明文出现在日志和浏览器历史）；握手后首条消息鉴权（实现复杂，有空窗期）

---

## ADR-09 · 房间成员列表迁至 Redis，DB 只存元数据

**文件**：[`docs/design/data-schema.md`](../design/data-schema.md)  
**原版**：DB 维护 room_members 多对多表  
**决策**：成员列表存 Redis Hash (`room:{room_id}`)，DB rooms 表只存房间元数据（owner_id、状态、描述）  
**原因**：成员进出频率高（聊天室场景），每次变更都写 DB 开销过大；成员列表是瞬态数据，进程重启后可以按需重建；符合 ADR-04 store 语义

---

## ADR-10 · 采用规格驱动开发（SDD）工作流

**文件**：[`docs/spec/specification.md`](../spec/specification.md)  
**决策**：以文档/规格为第一公民，AI 以文档为输入生成代码；四阶段循环：Specify → Plan → Implement → Validate  
**原因**：相比 vibe coding，SDD 产出更稳定、可回溯，决策上下文不丢失；AI 输出质量与给定上下文质量直接挂钩

---

## ADR-11 · WS 并发模型：Actor 模型 + 分实体单一 Writer 策略

**文件**：[`docs/design/ws-design.md`](../design/ws-design.md)  
**决策**：Hub 和 Room 各自维护一个 `Run()` goroutine 作为事件循环。每个数据实体有且只有一个 goroutine 写入，不使用 `sync.Mutex`。

**分实体策略**：

| 实体 | Writer | 具体方案 | 否决方案 |
|:---|:---|:---|:---|
| `UserSession` | `Hub.Run()` | ReadPump 收心跳 → 投 HEARTBEAT 消息到 Hub.Inbound → Hub 写 LastPulse | ReadPump 直写：跨 goroutine 竞态 |
| `MemRoom` | `Room.Run()` | Hub 通过 `register/leave` channel 发加减成员命令，Room 统一写 Members | Hub 直写 + RWMutex：嵌套锁易死锁 |
| `GameInstance` | `Room.Run()` | 完全独占，无并发访问 | — |
| `Client.roomInbound` | `Hub.Run()` 写一次 | `atomic.Pointer`，Store/Load 保证跨 goroutine 可见性 | channel happens-before：ReadPump select 无自然等待点，易读到 nil |

**核心原因**：
- 围棋对局是严格的顺序状态机（第 N 手必须在第 N-1 手之后），强行并发化收益为零
- `sync.RWMutex`：MemRoom 写操作频繁，RW 锁优势消失；嵌套调用极易死锁
- `sync.Map`：只适合 KV，无法覆盖 GameInstance 复杂嵌套
- 无锁/CAS：围棋状态更新是复合操作，CAS 无法原子覆盖

---

## ADR-12 · 倒计时选用 Room 内置 Ticker，不开独立 goroutine

**文件**：[`docs/design/ws-design.md`](../design/ws-design.md)  
**决策**：在 `Room.Run()` 的 `select` 中增加 `time.Ticker` case，Tick 时由 Room 直接处理计时、超时判断和广播，不为计时单独开 goroutine。  
**原因**：
- Room 自身的 goroutine 已经是完美的"定时任务容器"，Ticker 内置是零成本扩展
- 独立 Timer goroutine 需要额外管理生命周期（context 取消、WaitGroup），带来不对等的复杂度
- 全局 Ticker + Hub 广播会让 Hub 重新成为瓶颈，且 Hub 不应访问 GameInstance 内部状态

**细节**：Ticker 随 Room 创建而存在，通过 `Instance == nil` 判断是否处理，不需要动态开关。

---

## ADR-13 · 原有 WS 实现对比与修正

**参考实现**：`FantasyGoWorld-BE-new/internal/pkg/ws`  
**文件**：[`docs/design/ws-design.md](../design/ws-design.md) §2-§3`

### 原实现的三个核心问题

**① Room.Run() 是死代码**  
`room.go` 定义了完整的 `Room.Run()` 事件循环，但 `hub.go` 从未启动它。Hub 用自己的 `h.Rooms map[string]map[*Client]bool` 管理成员——这是 Hub 内部的普通 map，不是 Room goroutine 的引用。Room 实际上只是个数据结构，从未以独立 goroutine 运行。

**② `WsMessage.Type` 混淆传输层与业务层语义**  
原实现的 Type 取值为 `"broadcast"/"unicast"/"join_room"`——前两者是**传输方式**，后者才是**业务意图**。广播/单播是服务端的实现细节，不应由客户端声明。改为：Type 描述业务动作（`MOVE`/`PROPOSAL`/`INVITE`），广播/单播由服务端根据动作类型决定。

**③ `broadcast` goroutine pool 实现有并发安全隐患（但异步广播思路是对的）**  
原实现在 `go func()` 内访问 `recipients`（一个从 `r.clients` 临时取出的切片），同时 Room.Run() 可能在下一条 MOVE 到来时修改 `r.clients`。实际上切片已经是快照，但 `r.clients` 是 map，map 遍历和修改并发会 panic——原代码在 broadcast goroutine 外先 `for client := range r.clients` 拷贝为切片再传入，因此是安全的。但 5 个 worker goroutine 的设计对于"快照已经是切片，直接 for-each send channel"这个场景来说是额外复杂度。

> **修正**：观战者可达数万，异步广播方向正确，改为：Room.Run() 内拿客户端切片快照（在 Room goroutine，安全），然后 `go func(payload, snapshot)` 单一 goroutine 非阻塞 send，无需 worker pool。详见 ws-design.md §7。


### 消息路由决策（各类型逐条）

| 消息 | 路由 | 核心理由 |
|:---|:---|:---|
| CONNECT | Hub | 连接建立是 Hub 注册事件，此时无 Room 上下文 |
| DISCONNECT | Hub（defer）| Hub 决定是否启动 GRACE_PERIOD，需访问 UserSession |
| HEARTBEAT | Hub | `UserSession.LastPulse` 是 Hub-owned，ReadPump 不可跨 goroutine 直写 |
| JOIN_ROOM | Hub | 先有鸡先有蛋：`roomInbound` 由 Hub 在 JOIN_ROOM 后才注入，走 Room 需要先拿到 Room.Inbound，只有 Hub 能提供 |
| LEAVE_ROOM | Hub | 需同时：① Room.leave channel 命令，② 清除 Client.roomInbound，③ 更新 UserSession。横跨两个 goroutine 边界，只有 Hub 能协调 |
| INVITE | Hub | 需按 UserID 查 sessions map 做单播，Room 无全局视图，且邀请可发生在大厅（无 Room 上下文） |
| PROPOSAL propose/counter/reject | **Room** | 双方已是 Room 成员，roomInbound 早已注入，Room.clients 持有双方 Client 引用，可房间内单播，不需 Hub sessions map；走 Hub 会把房间内操作序列化进全局事件循环阻塞其他操作 |
| PROPOSAL accept | **Room** | 同上；accept 呠 Room 激活 GameInstance、广播 GAME_START；完成后异步向 Hub 发 GameActivatedEvent 更新 UserSession.ActiveGameID |
| MOVE/PASS/RESIGN/CHAT | Room（绕过 Hub）| 纯对局内事件，Hub 中转=所有对局串行化，是核心性能瓶颈 |

---

## ADR-14 · WS 升级路由：双端点 vs 单端点

**参考实现**：`internal/api/router/router.go`
```
GET /ws/hall         → ConnectHub
GET /ws/room/:roomID → ConnectRoom
```

### 原设计的优点（值得保留的思路）

- **HTTP 层前置校验**：WS 升级前在 HTTP handler 里做 Redis roomExists 检查，房间不存在直接返回 404，比升级后通过 WS 消息报错更省资源、语义更清晰。
- **URL 携带意图**：`/ws/room/101` 本身表达了"连接到房间 101"，不需要升级后再发 JOIN_ROOM 消息。
- **中间件层可做房间级鉴权**：roomID 在 Gin middleware 里直接可用。

### 原设计的缺点

- **名不副实**：`ConnectRoom` 最终把消息发到 `hub.Inbound`，Room goroutine 从未启动。"房间 WS"实际只是"自动发 join_room 的 Hub WS"，没有真正的 Room 隔离。
- **用户同时持有两条 WS 连接**：大厅连接 + 房间连接 = 4 个 goroutine pair，双倍资源消耗。
- **前端管理两条 WS 生命周期**：两连接状态需要协调，增加前端复杂度。
- **广播无法区分两类连接**：两者都注册进同一个 `hub.Clients` map，Hub 看到的是同一类 Client。

### 决策：单 WS 端点，Query Param 保留前置校验

**采用单端点**：`GET /ws`，一条连接处理大厅和房间所有消息。

**前置校验通过 Query Param 保留**：

```
GET /ws           ← 纯大厅连接（默认）
GET /ws?room=101  ← 直接进某房间，upgrade 前做 Redis roomExists 校验
```

服务端在 HTTP upgrade handler 里读 `room` 参数，校验通过后升级，升级后自动触发 JOIN_ROOM 注入 roomInbound。对前端透明，整个生命周期只有一条连接。

**为什么不用专用管道（`/ws/room/:roomID`）**：

前提错误：该方案成立需要"大厅不需要实时推送"。但本项目大厅需要推送邀请、在线用户状态、房间变更通知，没有大厅 WS 这些都要降级为轮询。

若同时保留 `/ws/hall` + `/ws/room/:id`：
- 用户在房间里持有两条连接，双倍 goroutine
- 前端需要协调两条 WS 的 open/close/reconnect 生命周期
- Hub 里同一用户有两个 Client，广播重复

**围棋场景的关键特征**：**每条 WS 连接**的状态是顺序的（此连接要么在大厅，要么在某个房间），但用户可通过多 Tab 并发持有多条连接（一边下棋一边看大厅）。这与单端点设计完全兼容——两个 Tab 各自连 `/ws`，各自管理自己的 roomInbound 状态；Hub 广播时对同一 UserID 的所有活跃 Client 全部推送。
