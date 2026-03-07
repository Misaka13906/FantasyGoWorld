# 疑惑与解答记录

> **路径**：[`docs/decision/why.md`](why.md)  
> **用途**：开发过程中遇到的疑问、备选方案对比及最终选择原因，防止决策上下文丢失，面向人类技术学习。

注：不仅需要记录最终选择的原因，也需要记录不同解决方案的优缺点。


---

## 前端架构

### SPA 和 MPA 的区别是什么？

| | MPA（多页应用） | SPA（单页应用） | SSR（服务端渲染） |
|:---|:---|:---|:---|
| **原理** | 每次跳转发 HTTP 请求，服务端返回新 HTML，页面完全重建 | 只加载一次 `index.html`，路由跳转由 JS 控制，页面从不刷新 | 服务端渲染首屏 HTML，客户端激活后表现为 SPA |
| **JS 状态** | 跳转时销毁，WS/变量无法跨页面保持 | 全程存活，WS 连接可跨路由维持 | 同 SPA |
| **SEO（搜索引擎优化）** | ✅ 天然友好，服务端直出 HTML | ❌ 需额外处理（预渲染或 SSR） | ✅ 友好 |
| **首屏速度** | 快（服务端返回即可渲染） | 慢（需先下载完整 JS bundle） | 快 |
| **开发复杂度** | 低，不依赖前端框架 | 中，需框架和路由配置 | 高，需处理服务端渲染与客户端激活的差异 |
| **适用场景** | 内容型网站、文档、博客 | 管理后台、实时应用、游戏平台 | 电商、内容平台（兼顾 SEO 和交互性） |
| **代表框架** | 传统三件套、PHP 模板 | React/Vue + Vite | Next.js、Nuxt.js |

**本项目选择 SPA（React + Vite）**：围棋平台核心是实时对局体验，需跨路由维持 WS 连接，SEO 不是优先考量。


---

### SPA 里"全局 WebSocket"怎么工作？

SPA 有一个 App 实例（整个应用的根组件），它的生命周期从用户登录到退出登录。

```
<App>   ← 在这里 useEffect 建 WS，绑定登录态
  <Router>
    /lobby     → 用 App 层的 WS
    /room/:id  → 额外建对局 WS（组件 unmount 时关闭）
    /profile   → 纯 HTTP，不需要 WS
  </Router>
</App>
```

所谓"全局"是指在**同一个 Tab 的 App 生命周期内全局**，不跨 Tab。

---

### 新标签页能不能共享 WebSocket？

每个 Tab 是独立的 JS 上下文（独立进程/线程），内存完全隔离，Tab A 的 WS 连接对 Tab B 不可见。

**有三种处理方式：**

| 方案 | 原理 | 优点 | 缺点 |
|:---|:---|:---|:---|
| 每 Tab 独立连接 | 每个 Tab 各自建 WS，服务端广播给用户的全部连接 | 最简单，无跨 Tab 通信复杂度 | 同一用户可能有多个活跃连接 |
| SharedWorker | 浏览器内置 API，Worker 跨 Tab 共享 JS 上下文 | 真正共享同一个 WS，节省连接数 | 开发复杂，Safari 兼容差，几乎无生产项目使用 |
| BroadcastChannel | Tab 间广播消息（不共享连接本身） | 可以同步数据状态 | 无法共享 WS 连接，每 Tab 仍需独立连接 |

**本项目选择每 Tab 独立连接**：服务端对"同一用户多个连接"做广播处理，实现最简单、可靠。多 Tab 属于边界场景，不值得引入 SharedWorker 的复杂度。

### 我们用的这个意思是单页应用但是一个 tab一个连接？
对，理解完全正确。这两个概念不矛盾：

"单页" = 一个 Tab 内只加载一次 HTML，路由切换不刷新页面，JS 状态全程存活。
"一个 Tab 一个连接" = 每个 Tab 独立运行一套 SPA 实例，各自建自己的 WS 连接。
所以更准确的描述是：

每个 Tab
  └─ 一个 SPA 实例（App 载入一次，never reload）
       ├─ App 层大厅 WS（登录后建立，随 App 存活）
       └─ 进对局室路由时建立对局 WS（离开路由时关闭）
"单页"强调的是页面内切换不重载，和 Tab 数量无关。开多个 Tab 就是多套独立的 SPA 实例，每套各管各的 WS，彼此完全隔离。服务端视角看是同一个用户的多条连接，广播时全部推送即可。


---

## WebSocket

### WebSocket URL 是什么，和 HTTP URL 有什么区别？

WebSocket 连接**必须从一个 HTTP 请求开始**。浏览器发送一个带有特殊头的 HTTP GET 请求，服务端同意后把这条 HTTP 连接"升级"成 WebSocket 双工连接：

```
客户端 HTTP 请求：
  GET /ws HTTP/1.1
  Connection: Upgrade
  Upgrade: websocket
  Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==

服务端响应：
  HTTP/1.1 101 Switching Protocols
  Upgrade: websocket
  Connection: Upgrade
```

这个 HTTP GET 的 URL（包括路径参数和 Query Param）在握手期间都是可用的。Go 里用 `c.Param()` 和 `c.Query()` 读取，读完后 `upgrader.Upgrade()` 就建立了 WS 连接。

---

### WS 端点设计：单管道 vs 专用管道

**单管道（大管道，本项目选用）**：全局一个 `GET /ws`，所有状态（大厅/房间）都在一条连接上通过消息 type 路由。

**专用管道**：不同上下文用不同 URL，如 `GET /ws/hall`（大厅）+ `GET /ws/room/:roomID`（房间），每个上下文一条独立连接。

| | 单管道 | 专用管道 |
|:---|:---|:---|
| 连接数 | 每 Tab 一条 | 每 Tab 可能多条（大厅 + 房间并存） |
| 前端复杂度 | 低，管一条 WS | 高，协调多条 WS 生命周期 |
| HTTP 层前置校验 | 用 Query Param 实现（`/ws?room=101`） | 天然，升级路由即带上下文 |
| 适用场景 | 状态顺序切换的应用（游戏、聊天室）| 上下文真正相互独立的服务 |

**本项目用单管道的原因**：大厅需要实时推送（邀请、在线状态），如果只有 `/ws/room/:id`，大厅连接就没了，要么跑轮询，要么还是要补一个 `/ws/hall`——两条连接的协调成本得不偿失。

---

### 但用户可以同时下棋和在大厅聊天啊，怎么办？

用多 Tab。每个 Tab 是独立的 JS 上下文，各自建一条 `/ws` 连接：

```
Tab 1（大厅）：连 /ws          → roomInbound = nil，停在大厅
Tab 2（对局室）：连 /ws?room=101 → Hub 在 upgrade 时注入 roomInbound
```

Hub 收到大厅广播（如 MEMBER_UPDATE）时，对同一 UserID 的所有活跃 Client **全部推送**。Hub 的 `clients` 是 `map[ConnID]*Client`，允许同一用户有多个 ConnID。

这和多管道方案达到的效果完全一样，但不需要前端管理多条 WS 的 open/close/error 生命周期。

---

### `/ws?room=101` 的 Query Param 起什么作用？

WS 升级是一次 HTTP 请求，这次请求**可以做完整的 HTTP 层校验**。服务端在 upgrade handler 里：

```go
func wsHandler(hub *Hub) gin.HandlerFunc {
    return func(c *gin.Context) {
        roomID := c.Query("room")   // 可能为空（纯大厅连接）
        if roomID != "" {
            exists, _ := roomRedisRepo.CheckRoomExists(ctx, roomID)
            if !exists {
                c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
                return  // 直接返回 HTTP 404，不推进握手
            }
        }
        conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
        client := InitClient(hub, conn, userID)
        if roomID != "" {
            // 升级成功后自动发 JOIN_ROOM
            hub.Inbound <- &Envelope{Type: "JOIN_ROOM", RoomID: roomID, Sender: client}
        }
    }
}
```

好处：**房间不存在直接返回 HTTP 404**，比升级成功后再通过 WS 消息报错更省资源，前端也可以根据 HTTP 状态码直接展示错误，而不需要解析 WS 消息。

---

### 后端 Go 的 Hub+Client 模式是什么原理？

标准的 Gorilla WebSocket 服务端模式：

| 概念 | 职责 |
|:---|:---|
| `Client` | 代表一个 WS 连接，持有 Conn 和发送 channel |
| `Hub` | 全局连接注册中心，负责广播、单播、连接生命周期 |
| `Room` | 房间级事件循环，处理对局内的落子/聊天等消息 |

每个 `Client` 开两个 goroutine：
- **ReadPump**：从 WS 读消息 → 按 type 路由到 Hub 或 Room。
- **WritePump**：从 `SendCh` channel 读 → 写入 WS（保证单 goroutine 写，避免并发写冲突）。

`Hub.Run()` 是一个 select 循环，串行处理 Register/Unregister/Broadcast，不需要加锁。

---

### 为什么客户端发消息要分流到 Hub 还是 Room？

- **Hub（控制平面）**：低频的控制类消息——JOIN_ROOM、INVITE、PROPOSAL 协商、全局广播。需要全局视图。
- **Room（数据平面）**：高频的游戏类消息——MOVE、PASS、RESIGN、CHAT。不经过 Hub 这个串行瓶颈，直接进 Room 事件循环。

分流机制：`Client` 加入房间时，Hub 把 `Room.Inbound` channel 引用注入 Client，之后游戏消息绕过 Hub 直投 Room。

---

### 能不能全用 Hub，不分 Room？

不行。Hub 是单个 goroutine 的 `select` 循环。如果把所有消息都放进 Hub：

- 规则引擎的并查集计算会**阻塞所有其他用户的消息响应**
- 计时器 Tick 必须在单独的粒度触发，如果 Hub 的 select 被一次落子计算卡住，所有房间的计时都会抖动
- 房间 A 的对局与房间 B 完全无关联，强制串行化是纯粹的性能浪费

**Room 的本质**：给一个时间上隔离的、围棋对局专属的 goroutine 事件循环。

---

### Hub 只有一个 goroutine 处理所有操作，会不会阻塞？

比如 1 万人同时登录建立连接，Hub 会不会卡死导致其他用户的邀请和房间操作超时？

**答案是：几乎不可能阻塞。** 只要严格遵守代码规范，Go 的单个 goroutine 处理 10 万 qps 的**纯内存操作**轻而易举。

这里有一个关键的**剥离慢速 I/O**的设计原理：

1. **连接建立的"慢"操作在进入 Hub 前就完成了。**
   当 1 万人同时尝试连接时：TCP 三次握手、HTTP 解析、JWT Token 解析与鉴权、Redis 查缓存确认房间有效性——这些全都是在 Gin 的 **HTTP Worker goroutine 中并发执行的**。HTTP 层可以同时有几千个 goroutine 在处理这些耗时 I/O。
2. **Hub.Run() 只做"纯内存操作"。**
   只有当上述所有校验全通过，连接成功 Upgrade 成 WebSocket 后，才会把这个准备就绪的 `*Client` 发入 `Hub.register` channel。
   接下来 Hub 每收到一个 `Register` 操作，只做两件事：
   - 往 `Hub.clients` map 里写一条记录
   - 往 `Hub.sessions` map 里写一条记录
   Go 的 map 插入只需要十几**纳秒（ns）**。连续处理 1 万次 map 插入不到 1 毫秒（ms）。

#### 唯一的潜在瓶颈与防范契约

Hub 唯一会阻塞的可能性，是开发者**违反了 Actor 模型契约**：在 `Hub.Run()` 的 case 代码块里，写了同步查询 DB/Redis 的代码，或者发起了一个同步 HTTP 请求。

> **绝对禁言令**：在 `Hub.Run()` 和 `Room.Run()` 的事件循环函数体内，**严禁写任何阻塞型 I/O 操作（包括查 MySQL、查 Redis、发 HTTP 请求）**。
>
> 如果必须查 DB：必须新开一个 goroutine `go func(){}` 去查，查完后再把结果 `[类型]Channel` 发回到 Hub 的 Inbound 来处理下一步。 

对于广播消息给 1 万人是否会阻塞 Hub，由于我们采用了"快照 + 异步写"的策略（发 channel 是非阻塞的），这也完全不会卡死主要事件循环。

---

### WS 状态并发怎么处理？（分实体讨论）

Actor 模型的核心原则：**每个数据实体有且只有一个 Writer goroutine**。不同实体的 Writer 不同，方案也不同。

#### `UserSession`

- **Writer**：`Hub.Run()` 唯一写。
- **问题**：`Client.ReadPump` 收到心跳帧时，能否直接写 `session.LastPulse`？
- **不能**：ReadPump 是独立 goroutine，直接写会与 Hub.Run() 产生数据竞争。
- **解决方案**：ReadPump 收到心跳帧 → 投递 `HEARTBEAT` 消息到 `Hub.Inbound` → Hub.Run() 统一更新 `LastPulse`。Session 所有字段的写入都在 Hub 单一 goroutine 内，无竞争，不需要锁。

#### `MemRoom`

- **问题**：Hub 需要加减成员（JOIN/LEAVE），Room 需要更新配置和激活对局。两个 goroutine 都想写 MemRoom，是真实的并发冲突。
- **弃用方案**：给 MemRoom 加 `sync.RWMutex`——引入锁后在 Hub.Run() 和 Room.Run() 各处都要加锁，嵌套调用易死锁，破坏 Actor 模型的推理简洁性。
- **解决方案**：**Room.Run() 是唯一 Writer**。Hub 通过 `Room.register chan *Client` 和 `Room.leave chan *Client` 两个专用 channel 发命令，Room.Run() select 到后自己写 Members。Hub 不直接访问任何 MemRoom 字段。

#### `GameInstance`

- **Writer**：`Room.Run()` 完全独占，不存在任何并发访问，Actor 模型自然覆盖。

#### `Client.roomInbound`

- **问题**：Hub.Run() 在 JOIN_ROOM 时写入一次，Client.ReadPump 持续读取。是系统中**唯一一次跨 goroutine 赋值**。
- **弃用方案**：依赖 channel happens-before—— ReadPump 的 select 没有一个自然的"等待 roomInbound 注入完成"的点，稍有不慎会读到 nil 且无编译警告。
- **解决方案**：`roomInbound` 声明为 `atomic.Pointer[chan<- *Envelope]`，Hub Store，ReadPump Load。`atomic.Pointer` 的 Store/Load 保证跨 goroutine 可见性，语义明确，无锁，性能接近裸指针。

#### 那读操作怎么办？会不会遇到并发读？

既然是单一 Writer，意味着**只有这个 Writer goroutine 允许直接读取它拥有的状态**。如果是其他 goroutine（如 HTTP Handler）直接通过指针去读，必定会发生 Data Race（Go 的 map 被并发读写会直接 panic）。

**解决并发读的两种标准方案：**

1. **通过 Channel 异步查快照**：给控制该状态的 goroutine 发一条自带返回通道的消息。
   ```go
   type QueryRoomMsg struct {
       Reply chan RoomSnapshot
   }
   // 其他 goroutine 这样查询：
   replyCh := make(chan RoomSnapshot)
   room.Inbound <- QueryRoomMsg{Reply: replyCh}
   snapshot := <-replyCh // 阻塞等待 Room 返回最新的深拷贝快照
   ```
2. **深拷贝与消息传递**：在本项目的核心循环中，几乎没有"外部主动来读"的场景。当状态发生变更（落子、超时）时，Writer goroutine（如 `Room.Run()`）会**主动**将当前状态深拷贝，序列化成 JSON (`[]byte`)，然后把这些不可变的消息放入各 Client 的 channel 中供 `WritePump` 发送。由于传递的是彻底脱离原结构的字节流快照，根本没有暴露内存指针，自然杜绝了并发读写问题。

#### 全局并发风险总结

| 实体 | Writer | 跨 goroutine 写方案 | 跨 goroutine 读方案 |
|:---|:---|:---|:---|
| `UserSession` | Hub.Run() | 心跳通过 channel 转发 Hub 写 | 必须转为向 Hub 发查询 Channel 消息 |
| `MemRoom` | Room.Run() | Hub 通过 register/leave channel 发命令 | Room 组装为字节流快照广播给客户端 |
| `GameInstance` | Room.Run() | 完全独占 | 同上，主动推送快照，不开放直接读取 |
| `Client.roomInbound` | Hub.Run() 写一次 | `atomic.Pointer` | 读也通过 `atomic.Pointer`，完全无锁安全 |

---

### 倒计时怎么做？

三种方案：

| 方案 | 弃用原因 |
|:---|:---|
| 独立 Timer goroutine | 需要管理 context 取消和 WaitGroup；Room 本身已是完美容器，额外 goroutine 不对等 |
| 全局 Ticker + Hub 广播 | Hub 不应访问 GameInstance 内部；Hub 重新成为热路径瓶颈 |
| **Room 内置 Ticker** ✅ | Ticker 作为 `Room.Run()` select 的一个 case，零成本扩展，完全符合"GameInstance 单一 Writer"原则 |

**实现要点**：Ticker 随 Room 创建（不动态开关），通过 `Instance == nil` 判断是否处理；落子成功后 Room 切换持子方并重置读秒，Client 无需通知。

---

### WebSocket Upgrade 为什么不能用 Authorization Header？

**三种解决方案对比：**

| 方案 | 做法 | 优点 | 缺点 |
|:---|:---|:---|:---|
| Query 参数 | `ws://host/ws?token=xxx` | 实现最简单 | token 明文出现在服务端日志、代理日志、浏览器历史中，有泄漏风险 |
| Cookie Fallback | 登录时额外写 Cookie，WS 自动携带，Middleware fallback 读 | 安全，前端无感知，Middleware 统一覆盖 REST 和 WS | 需配置 `SameSite`，跨域场景需额外处理 |
| 握手后首条消息鉴权 | WS 建立后客户端发 `AUTH` 消息含 token | 安全，不依赖 Cookie | 实现复杂，握手到鉴权之间有极短空窗期 |

**本项目选择 Cookie Fallback**：安全性好，前端对 WS 建立无感知（不需要拼参数），Middleware 一套逻辑同时覆盖 REST 和 WS。

---

## 认证与会话

### JWT 和 Session 的区别？

**JWT（JSON Web Token）**：token 自包含用户信息和签名，服务端无需存储，验证只靠签名密钥。

**Server Session（Redis 存储）**：服务端在 Redis 存 `{session_id → user_data}`，客户端只持有不透明的 session_id。

| | JWT | Server Session（Redis） |
|:---|:---|:---|
| 服务端存储 | 无（无状态） | Redis 存 session 数据 |
| 撤销 token | 困难（需维护 token 黑名单） | 直接删 Redis key |
| 水平扩展 | 轻松（无共享状态） | 需要所有节点共享同一 Redis |
| 信息量 | token 内自带 | 只有 ID，其余查 Redis |
| 安全性 | 依赖签名密钥不泄漏 | 依赖 session_id 难以猜测 |

**本项目选择 JWT**：单体架构初期零额外存储，与 WS 升级配合好（Upgrade 时一次验证，后续连接不再鉴权）。如果后期需要强制踢人，维护一个 JWT 黑名单（Redis Set，TTL 与 token 过期时间对齐）即可。
本项目用 JWT：`access_token`（1h）+ `refresh_token`（7d，httpOnly Cookie）。


---

### Cookie 是什么，为什么 logout 要请求接口？

Cookie 是浏览器存储并自动随请求携带的键值对。分两种：

**普通 Cookie**（非 httpOnly）：JS 可读写（`document.cookie`），前端自己可以删除。

**httpOnly Cookie**：对 JS 完全隐藏，JS 既不能读也不能删。删除它的唯一方式是让服务端在响应里设置同名 Cookie 且 `Max-Age=0`：

```http
Set-Cookie: refresh_token=; Max-Age=0; HttpOnly; Path=/
```

所以 `POST /auth/logout` 的本质是：**帮前端删一个它自己删不了的 Cookie**。

本项目中：
- `access_token`：非 httpOnly Cookie，前端自己可以清除。
- `refresh_token`：httpOnly Cookie，必须通过 logout 接口由服务端清除。

---

## 数据存储

### Redis 作为"缓存"和作为"主存储"有什么区别？

| | Redis 缓存（Cache） | Redis 主存储（Store） |
|:---|:---|:---|
| 数据来源 | DB 数据的副本 | Redis 本身是 Source of Truth |
| DB 对应表 | 有 | 无 |
| 数据丢失 | 不可接受，丢了回查 DB | 可接受（瞬态数据） |
| TTL | 必须设（否则不会失效） | 无 TTL 或由业务生命周期控制 |
| 代码模式 | 先查 Redis，miss 则查 DB 回填 | 直接读写 Redis，无回退 |

本项目的 Repository 分层：
- `store/`：房间成员、用户在线状态、实时落子流（瞬态，服务重启可重建）
- `cache/`：用户昵称、Elo、段位（DB 数据的加速副本，供大厅列表渲染）

---

### Redis Store 选型：AOF（每秒刷盘）与"混合持久化"哪个更好？

其实这是一个经典的**概念混淆陷阱**。被问到"AOF 每秒刷盘和混合持久化哪个好"时，正确答案是：**它们不仅不互斥，而且应该同时开启，且解决的是完全不同的问题。**

1. **AOF 每秒刷盘 (`appendfsync everysec`) 解决的是"性能 vs 数据安全性"的权衡。**
   - Redis 将命令先写内存，然后每秒异步 fsync 到磁盘。宕机最多丢 1 秒数据，对我们的对局和房间状态（瞬态数据）来说完全在可接受范围内。
   - 这比 `always`（每条命令都刷盘，极慢）快得多，比 `no`（交由操作系统决定，极其不安全）安全得多。

2. **混合持久化 (`aof-use-rdb-preamble yes`，Redis 5.0 默认开启) 解决的是"AOF 文件体积膨胀与重启恢复极慢"的问题。**
   - AOF 的致命缺点是记录了所有的历史过程（例如对同一个 key `INCR` 了 1 万次，AOF 就会有 1 万条命令）。虽然有 AOF Rewrite 机制，但纯 AOF 的重写和加载都很慢。
   - **混合持久化机制**：当 AOF 执行 Rewrite 时，Redis 会**先用 RDB 的二进制格式把当前的内存全量快照写到新的 AOF 文件头部**，此时处于进行中的增量命令依然以 AOF 的文本格式追加到文件尾部。
   - **结果**：产生了一个"头部是二进制 RDB，尾部是增量 AOF"的混合文件。

**完美配合方案：**
- 日常运行中，增量数据靠 `appendfsync everysec` 每秒刷盘（保证最多丢 1 秒）。
- 当文件变大时，靠 `aof-use-rdb-preamble yes` 做重写（极大地压缩历史脏数据，且重启加载时先走二进制 RDB 读取，速度提升十几倍）。
- **所以，"我打算用 AOF 每秒刷盘，并开启混合持久化用于重写"，这是满分答案。**

---

### 为什么 VARCHAR 主键比 INT 主键性能差，以及如何选 ID 策略？

**VARCHAR 主键的问题：**
1. **索引体积**：VARCHAR(255) 每项最大 255 字节，INT 只有 4 字节。索引树更高，查找需要更多磁盘 IO。
2. **JOIN 开销**：字符串比较需要逐字节处理并计算 collation，远慢于整数比较。
3. **外键存储**：所有引用该表的外键列都与主键等宽，VARCHAR 主键导致整个 schema 的外键列全部变宽。
4. **随机插入（UUID v4）**：随机 ID 导致 B-Tree 随机位置插入，频繁引发页分裂，写性能极差。

**ID 生成策略对比（UUID v4 基本都比 v7 差，不列入）：**

| | AUTO_INCREMENT | UUID v7 | 雪花 ID（Snowflake） |
|:---|:---|:---|:---|
| **类型** | INT/BIGINT | 128-bit，时间戳前缀 | 64-bit 整数 |
| **大小** | 4/8 字节 | 16 字节（字符串存更大）| 8 字节 |
| **单调性** | 严格递增 | 秒级递增，秒内随机 | 毫秒级严格递增（同机器同毫秒内序列号保证）|
| **B-Tree 性能** | 最优，顺序追加 | 较好但偶有秒内乱序 | 等同 AUTO_INCREMENT |
| **分布式** | ❌ 多写节点冲突 | ✅ 无需协调 | ✅ 需分配 worker ID |
| **信息泄漏** | 暴露记录总量和写入速度 | 无 | 可推算时间戳，不暴露总量 |

**本项目的选择：**

直接用 **`BIGINT AUTO_INCREMENT` 作为对外用户 ID**（洛谷、CF 等竞技平台都这样做），用户可以直接用数字 ID 加好友搜索。简单、直接，注册量泄漏对非商业化的游戏社区平台没有商业敏感性。

如果后期真的需要防泄漏（比如转型商业敏感场景），再加 `user_no` 列做双 ID 分离。

---

**📌 八股：注册量泄漏什么时候真的有问题？**

| 产品类型 | 危害 | 是否需要隐藏 |
|:---|:---|:---|
| SaaS 创业公司 | 竞对通过注册 UID 差值精确估算新增用户速度，用于竞争情报 | ✅ 需要 |
| 电商/交易平台 | 订单号自增 → 外部可估算日单量/GMV，是商业机密 | ✅ 需要 |
| 游戏/社区平台 | 注册量通常是公开营销数据，暴露了反而是自我宣传 | ❌ 不需要 |
| 内部工具/个人项目 | 无商业竞争，暴露无意义 | ❌ 不需要 |

---

### 为什么 `games` 表要用 `config_snap` JSON 存快照，而不是直接用外键关联字典表？

在之前的完全关系化思路中：存在 `game_rules` 和 `game_timers`，如果用户用了默认设置，直接引 ID；如果用户自定义了设置，就在字典表新插一条记录，返回新 ID 给这局游戏用。

但本项目并没有采用这种"全外键关联"模式，而是在 `games` 表中引入了 `config_snap`（JSON）保存当时的完整参数，仅仅把外键 `rule_id` 等作为"可选的元数据参考"（甚至是 NULL）。

**弃用完全外键关联模式的原因：**

1. **历史记录的绝对不可变性（Immutability）**
   对局记录（类似于财务账单）一经生成，**永远不能因为外部字典的变动而改变语义**。
   假设运营在后台把 `id=1` 的默认"中国规则"贴目从 7.5 改成了 6.5，如果在 `games` 表只存 `rule_id=1`，那么去年的比赛再渲染出来，就会错误地按 6.5 贴目结算。
   > 一种补救思路是："如果要改，那就添加新记录而不去改旧的"。但这要求做**软删除**或**版本控制**（把旧的 id=1 隐藏，新建 id=3 的中国规则），这会大幅增加后台运营 CRUD 的心智负担。

2. **防止字典表无限膨胀（Garbage Pollution）**
   如果支持高度自定义（比如玩家 A 设 15 秒读秒，玩家 B 设 16 秒，玩家 C 设 17 秒），每次自定义都在 `game_timers` 表里无脑 `INSERT` 一条新记录。
   由于这些自定义设置**仅仅被这一次对局专属使用**，`game_timers` 表很快会膨胀出几万条一次性的垃圾数据。字典表失去了"字典"归类的意义，变成了另一个事实上的大表。

**采用快照模式（Event Sourcing / Snapshot）的优势：**

- `config_snap (JSON)`：作为对局运转的**绝对事实来源**（Source of Truth）。对局引擎从这里读取参数，复盘页面从这里读取贴目，不论世界怎么变，这场棋的贴目永远是快照里定死的数字。
- `rule_id / timer_id (INT)`：降级为可选的**归类标签**。如果这场对局使用了预设方案，就记下 ID，纯粹为了方便日后做数据统计（比如"统计全站有多少比例的对局使用常规时限"）。如果是完全自定义的对局，这两个字段直接填 `NULL` 即可。

这是一种结合了关系型数据库（统计查询，Normalization）和文档型数据库（不可变事实快照，Denormalization）优点的经典业务架构模式，在电商订单表（快照商品当时的名称、价格）中也极为常见。

---

### `users` 主表和 `user_preferences` 偏好表为什么要分离？偏好要存 Redis 吗？

**为什么要分表（主从表策略）：**

1. **冷热指责分离**：`users` 主表存的是登录鉴权、基础身份、匹配核心数据（Username, Password, Elo 等）。这部分数据**高频读**（如大厅列表渲染）但极低频修改。
2. **减少主表体积**：用户偏好（DND 拒接开关、主题色、聊天常用意图/语录、棋子皮肤等）往往是扩展性极强的长文本（如 JSON 配置），而且可能会跟着版本频繁迭代加字段。如果全都塞在 `users` 表里，每次用户进大厅查列表，取出的数据页非常宽，造成巨大的内存和 I/O 浪费。
3. 把这些抽进 `user_preferences`，保证主表纯粹且精简，后期加再多偏好字段也不会影响核心链路（登录和匹配）的性能。

**偏好信息需要存进 Redis 吗？**

**结论：不需要全存，绝大多数偏好连 Redis 都不用进。唯一需要的是将 DND 状态投影到运行时内存中。**

1. **不需要全存的影响**："默认主题"、"自定义招呼语"这种配置确实需要持久化（MySQL），但它们只有**用户自己打开浏览器时加载一次**即可。大厅里的其他 1000 个用户根本不关心你给自己的棋盘配了什么皮肤。查 MySQL 并把结果放到前端的 `Zustand Store` 内存里缓存就足够了，后端完全不需要用 Redis 为这些只读一次的数据浪费宝贵的内存空间。
2. **唯一的高频判断场景：DND（免打扰）**
   - **痛点**：如果开了 DND，别人邀请你时应该立刻被拒。如果每次有人发邀请，后端都在判断时去查 MySQL / Redis 的偏好表，这就是极大的高并发瓶颈。
   - **解法**：在用户**登录建立 WS 连接**的那一刻，服务端顺带查一次 MySQL 的 `user_preferences` 拿到 DND 开关。如果 `dnd=1`，就在建立 `*Client` 时直接将其状态 `ws.UserSession.Status` 设置为 `"dnd"`。
   - 这样，所有的权限和状态判断全都在 Server 的纯内存（Hub 管理的 `map[UserID]*UserSession`）里完成了。

这样设计，**偏好的归偏好（重装 DB，慢加载），状态的归状态（重内存，快判断）**。
