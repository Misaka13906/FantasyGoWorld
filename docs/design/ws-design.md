# WebSocket 连接管理架构

> **路径**：[`docs/design/ws-design.md`](ws-design.md)  
> **用途**：Hub/Client/Room 三层架构最终设计、各实体内存结构与操作接口、关键时序图。  
> **关联文档**：[技术设计总纲](./tech-design.md) · [接口协议](./api-spec.md) · [数据存储](./data-schema.md)  
> **选型原因**：见 [`docs/decision/why.md](../decision/why.md) §WebSocket` 和 [`docs/decision/log.md](../decision/log.md) ADR-11/12`

---

## 1. 整体分层

```
[Browser] ──WS── [Client goroutines: ReadPump / WritePump]
                       │
              ReadPump 解析 msg.Type
                  ┌────┴─────────────┐
             游戏动作                控制消息
             (MOVE/PASS/RESIGN/CHAT) (JOIN/LEAVE/PROPOSAL/INVITE)
                  │                       │
          Client.roomInbound          Hub.Inbound
                  │                       │
            Room.Run()               Hub.Run()
```

- **Hub**：控制平面，全局唯一，与进程同寿。管理连接注册/注销、房间加入/离开、邀请、协商等控制消息。
- **Room**：数据平面，每个活跃房间一个 goroutine。处理落子/虚手/认输，调用规则引擎，持有对局状态，运行倒计时 Ticker。
- **Client**：一条 WS 连接的物理抽象。ReadPump 负责路由；WritePump 是唯一写 WS 的 goroutine。

---

## 2. 数据控制权原则

> **单一 Writer 原则**：控制权 = 单一 Writer。谁是唯一写入 goroutine，谁就"控制"这个数据结构。
> **跨 Goroutine 读取原则**：其他 goroutine 严禁直接通过指针读取（防止并发读写引发 panic）。读操作只允许有两种路径：
> 1. 通过 Channel 发送附带 `Reply chan Snapshot` 的查询消息，由 Writer 返回深拷贝快照。
> 2. Writer 主动将状态深拷贝并序列化为不可变的 JSON (`[]byte`)，通过 Channel 推送给其他 goroutine（本项目主要采用此方式）。

| 数据结构 | 唯一 Writer | 读安全策略 / 理由 |
|:---|:---|:---|
| `sessions map[UserID]*UserSession` | `Hub.Run()` | 跨连接全局视图。外部若需读取（如 API 查询在线状态），必须向 Hub 发 Query 消息，Hub 返回深拷贝。 |
| `rooms map[RoomID]*Room`（索引） | `Hub.Run()` | 仅 Hub 内部路由使用，无外部并发读需求。 |
| `Room.clients map[*Client]bool` | `Room.Run()` | 房间内在线连接列表。完全在 Room goroutine 内访问，天然无竞争。 |
| `MemRoom`（房间/对局业务状态）| `Room.Run()` | 写入在 Room 内。Room 主动序列化并广播，客户端收到的是不可变纯文本，彻底隔离内存指针。 |
| `GameInstance`（棋盘/计时）| `Room.Run()` | 完全独占。同样依靠主动广播 `[]byte` 快照，外部 goroutine 不持指针。 |
| `Client.roomInbound`（路由桥） | `Hub.Run()`（写一次） | 使用 `atomic.Pointer`。Store 与 Load 均为原子级操作，完全无锁并发安全。 |

---

## 3. 消息路由决策

### 3.1 所有消息类型路由表

| 消息类型 | 路由目标 | 路由原因 | 不走另一个的原因 |
|:---|:---|:---|:---|
| **CONNECT**（连接建立）| Hub | 连接建立本身就是 Hub 的注册事件 | 此时不存在任何 Room 上下文 |
| **DISCONNECT**（断线清理）| Hub（ReadPump defer 触发）| Hub 决策：对局中→启动 GRACE_PERIOD；非对局→直接清理 UserSession | Room 没有 UserSession 的访问权，无法单独决策 |
| **HEARTBEAT** | Hub | `UserSession.LastPulse` 是 Hub-owned 数据，ReadPump 不能跨 goroutine 直接写 | 若 ReadPump 直写 LastPulse，与 Hub.Run() 产生数据竞争（无论 GAMING 还是 IDLE） |
| **JOIN_ROOM** | Hub | **先有鸡还是先有蛋**：Client 要投递到 Room.Inbound，但 `roomInbound` 是 Hub 在处理 JOIN_ROOM **之后**才注入的。要走 Room 需要先拿到 Room.Inbound channel，但这只有 Hub 能提供 | 无法绕过。Client 此时 `roomInbound == nil` |
| **LEAVE_ROOM** | Hub | 需要同时：① 向 Room.leave channel 发命令（Hub 持有 Room 引用），② atomic store nil 清除 `Client.roomInbound`，③ 更新 `UserSession.ActiveGameID`。三件事横跨两个 goroutine 边界 | 若走 Room：Room 无法反向清除 Client.roomInbound（Client 是 Hub 管理的数据） |
| **INVITE** | Hub | 邀请目标是 UserID，需查 `sessions[targetID]` 拿到目标 Client 的 SendCh。Room 无全局 sessions map。邀请可在大厅发起（Client 不在任何 Room） | Room 无权访问其他用户的 Session |
| **PROPOSAL propose/counter/reject** | **Room** | PROPOSAL 发生在房间内，双方已通过 JOIN_ROOM 成为 Room 成员，`roomInbound` 早已注入。Room 的 `clients` map 直接持有双方 `*Client` 引用，可做房间内单播，无需 Hub 的全局 sessions map。PROPOSAL 走 Hub 只会把房间内操作不必要地序列化进全局事件循环，白白阻塞其他房间的 JOIN/LEAVE | INVITE（大厅邀请）才走 Hub，PROPOSAL 前提是双方已在同一 Room |
| **PROPOSAL accept** | **Room** | 同上。roomInbound 在 JOIN_ROOM 时已注入，accept 时 Room 直接激活 GameInstance、分配角色、广播 GAME_START。激活完成后通过 `Hub.notify <- GameActivatedEvent` 异步通知 Hub 更新 `UserSession.ActiveGameID`（Hub 不需要是同步路径的一部分） | — |
| **MOVE / PASS / RESIGN / CHAT** | Room（直达，绕过 Hub）| 纯对局内事件，只需要规则引擎和棋盘状态，Room 完全独立处理。Hub 作中转会成为所有对局的串行瓶颈 | 不走 Hub 是本设计的核心性能保证 |

### 3.2 PROPOSAL accept 的处理路径

```
Room.Run() 处理 PROPOSAL accept：
  1. 校验发起方和接受方都在 r.clients（前提保证，不满足则拒绝）
  2. 激活 GameInstance，分配黑白角色，更新 MemRoom.Status = "ongoing"
  3. 广播 GAME_START 给房间内所有成员（含观战者）
  4. hub.notify <- GameActivatedEvent{blackUserID, whiteUserID, gameID}  ← 异步通知
```

步骤 4 是非阻塞的单向 channel 发送，Room.Run() 不等待 Hub 的确认。Hub.Run() 异步收到后更新 `UserSession.ActiveGameID`。  
这样 Hub 不在 GAME_START 的关键路径上，广播延迟不受 Hub 当前负载影响。



---

## 4. 对象引用拓扑

```
Hub（全局唯一）
 ├── sessions map[UserID]*UserSession
 ├── clients  map[UserID]*Client
 └── rooms    map[RoomID]*Room
               │
              Room
               ├── register chan *Client     ← Hub 通过此 channel 加人，不直接写 MemRoom
               ├── leave    chan *Client     ← Hub 通过此 channel 减人，不直接写 MemRoom
               ├── Inbound  chan *Envelope   ← 游戏消息入口（ReadPump 投递）
               ├── clients  map[*Client]bool ← Room 内部维护，单一 Writer: Room.Run()
               └── state    *MemRoom         ← 房间/对局数据，单一 Writer: Room.Run()
                              └── Instance *GameInstance
```

---

## 5. 各实体定义与并发访问策略

### 5.1 UserSession

```go
type UserSession struct {
    UserID       int
    Nickname     string
    Elo          int
    Status       string    // "idle" | "gaming" | "dnd"
    ActiveGameID int       // 0=未参与对局（含观战）; >0=当前对局者（单一参与规则）
    LastPulse    time.Time // 由 Hub.Run() 在收到 HEARTBEAT 消息时更新
}
```

**并发策略**：`Hub.Run()` 是**唯一 Writer**。
- `Client.ReadPump` 收到心跳帧 → 投递 `HEARTBEAT` 消息到 `Hub.Inbound` → `Hub.Run()` 更新 `LastPulse`。
- `ReadPump` 不直接写 Session 任何字段。`Room.Run()` 只读 Nickname/Elo（广播用）。

---

### 5.2 MemRoom

```go
type MemRoom struct {
    ID       int
    OwnerID  int
    IsPublic bool
    Status   string          // "waiting" | "ongoing"
    Members  map[int]*Member // UserID -> Member
    Config   *GameConfig     // 当前协商配置契约
    Instance *GameInstance   // Status="ongoing" 后非 nil
}

type Member struct {
    UserID int
    Role   string // "spectator" | "black" | "white"
}
```

**并发策略**：`Room.Run()` 是**唯一 Writer**。
- Hub 加人：向 `Room.register` channel 发送 `*Client`，`Room.Run()` 收到后写 `Members`。
- Hub 减人：向 `Room.leave` channel 发送 `*Client`，`Room.Run()` 收到后写 `Members`。
- Hub 不直接访问 `MemRoom` 任何字段，避免与 `Room.Run()` 产生数据竞争。

---

### 5.3 GameInstance

```go
type GameInstance struct {
    GameID     int          // 对应 games 表 ID
    Config     GameConfig   // 本局生效的配置快照
    Board      [19][19]int8 // 实时棋盘：0=空 1=黑 2=白
    StepSeq    int          // 当前手数，防乱序
    LastMoveAt time.Time    // 最近落子时间，超时判断用
    Timer      *GameTimer   // 封装读秒/包干逻辑
}
```

**并发策略**：`Room.Run()` **完全独占读写**，无任何并发访问问题。

---

### 5.4 Client

```go
type Client struct {
    Hub         *Hub
    Conn        *websocket.Conn
    UserID      int
    SendCh      chan []byte                // WritePump 消费；其他 goroutine 只 send，不 receive
    roomInbound atomic.Pointer[chan<- *Envelope] // Hub 写一次，ReadPump 持续读
}
```

**并发策略**：
- `SendCh`：多写单读，channel 本身线程安全。
- `roomInbound`：Hub.Run() 在 `JOIN_ROOM` 时写入一次，ReadPump 持续读取。跨 goroutine，使用 `atomic.Pointer` 保证 store/load 的可见性，无需加锁。

---

## 4. Hub 处理的操作

| 消息类型 | 读取的状态 | 修改的状态 |
|:---|:---|:---|
| `CONNECT` | — | 新建 `UserSession`，写 `clients` |
| `DISCONNECT` | `sessions[userID]` | 删 `clients`，启动断线宽限期计时 |
| `HEARTBEAT` | — | 写 `session.LastPulse` |
| `JOIN_ROOM` | `rooms[roomID]` | 向 `room.register` 发命令；写 `client.roomInbound` |
| `LEAVE_ROOM` | — | 向 `room.leave` 发命令；清空 `client.roomInbound` |
| `INVITE` | `sessions[targetID]`（查 Status/ActiveGameID） | 单播 INVITE 给目标 Client |
| `PROPOSAL` | `rooms[roomID]`（查存在性） | 转发给目标 Client，不修改 MemRoom |

---

## 5. Room 处理的操作

| 触发来源 | 操作 | 读取的状态 | 修改的状态 |
|:---|:---|:---|:---|
| `register` channel | 成员加入 | — | `clients`，`state.Members`（Role="spectator"） |
| `leave` channel | 成员离开 | — | `clients`，`state.Members` |
| `PROPOSAL accept` | 激活对局 | `state.Config` | 新建 `GameInstance`，设 `state.Status="ongoing"`，`GAME_START` 广播 |
| `MOVE` | 落子 | `Instance.Board`，`Instance.StepSeq` | `Instance.Board`（规则引擎写），`Instance.StepSeq`，写 Redis，广播 `SYNC_BOARD` |
| `PASS` | 虚手 | `Instance.StepSeq` | `Instance.StepSeq`，广播 |
| `RESIGN` | 认输 | `Instance.GameID` | 触发终局路径，落库，清理 Redis，广播 `GAME_END` |
| `Timer Tick` | 每秒计时 | `Instance.Timer` | `Instance.Timer`（减少持子方时间），超时时触发终局 |

---

## 6. 倒计时实现

`time.Ticker`（1s）内置于 `Room.Run()` 的 select，不开独立 goroutine：

```
Room.Run() select：
  case c := <-r.register:   // 加入成员
  case c := <-r.leave:      // 离开成员
  case msg := <-r.Inbound:  // 游戏消息
  case <-ticker.C:          // 每 1 秒
      if r.state.Instance == nil { continue }  // 对局未激活，跳过
      持子方时间 -= 1s
      if 时间 == 0 → 走 RESIGN 同一条终局路径
      else if 进入读秒区 → 广播 TIMER_UPDATE（含读秒次数、警告标志）
```

落子成功后，Room 切换持子方并重置读秒次数，Client 无需任何通知。

---

## 7. 广播策略（观战者高连接数场景）

观战者在赛事期间可能达几千甚至几万人。当前设计的同步 for-loop 广播**锁住 Room.Run()**，遇到大量观战者会副作用事件循环。

### 7.1 异步快照广播（当前方案）

Room.Run() 内先拥取 clients 快照，然后开 goroutine 广播，不阻塞事件循环：

```go
// Room.Run() 内，处理完 MOVE 后：
result := marshalSyncBoard(...)

// 1. 在 Room goroutine 内拥取快照（安全）
recipients := make([]*Client, 0, len(r.clients))
for c := range r.clients {
    recipients = append(recipients, c)
}

// 2. 开独立 goroutine 广播，不阻塞 Room.Run()
go func(payload []byte, clients []*Client) {
    for _, c := range clients {
        select {
        case c.SendCh <- payload:
        default:
            // SendCh 缓冲满：记录日志，不强制断开（由心跳超时处理）
        }
    }
}(result, recipients)
```

**为什么可行**：围棋落子频率极低（每手间隔常达数秒），即使一次广播 1 万人耗时漏到几 ms，下一手不会在几 ms 内到达。

### 7.2 SpectatorHub（远期方案）

如果观战者规模达到 10 万+，异步快照推播需要进一步分层：

```
Room 只广播给：
  - 对局者 A.SendCh
  - 对局者 B.SendCh
  - spectatorRelay chan []byte  ← 只发一条。SpectatorHub 进行扁出

SpectatorHub
  └─订阅 Room.spectatorRelay
  └─ 面向所有观战者 Client.SendCh 广播（分批并发）
```

SpectatorHub 是单独的扩展单元，可以单独水平扩展。当前暂不实现，当观战者规模不过条件时再拆分。

### 7.3 方案对比

| 方案 | 适合人数 | Room.Run() 是否阻塞 | 复杂度 |
|:---|:---|:---|:---|
| 同步 for-loop | ≤ 100 | 是 | 最低 |
| **异步快照广播**（当前）| ≤ 1 万 | 否 | 中 |
| SpectatorHub | ≤ 10 万 | 否 | 较高 |
| SSE/HTTP streaming | 10 万+ | 否 | 高（客户端需区分连接类型）|

---

## 7. 关键业务时序图

见 [diagrams.md §4-§6](./diagrams.md)：
- **§4** 对局协商流程（PROPOSAL 多轮）
- **§5** 落子流程（MOVE → 规则引擎 → 广播）
- **§6** 断线保护与恢复



---

## 8. 原版代码关键修正点

| 原设计问题 | 修正方案 |
|:---|:---|
| 落子消息全部走 Hub，Room 收不到 | Client.ReadPump 按 type 分流，游戏动作直投 Room.Inbound |
| `boardState map[string]string` 结构随意 | 改为 `[19][19]int8`，由规则引擎维护 |
| `broadcast` 开 goroutine pool | 保留异步广播思路，改为快照 + 单一 goroutine（先拥快照再 fork，开销最小），见 §7.1 |
| `maxMessageSize = 512` | 调整为 65536（64KB）|
| `Client.rooms` 字段未维护 | 改为 `roomInbound atomic.Pointer`，由 Hub 在 JOIN_ROOM 时注入 |
| Hub 直接修改 MemRoom 字段 | Hub 只向 Room.register/leave channel 发命令，Room.Run() 统一写 |

---

## 9. 资源生命周期与有序性 (Lifecycle & Guarantees)

### 9.1 优雅退出与泄漏预防 (Graceful Shutdown)
- **Room 退出条件**：当 `len(clients) == 0` 且对局已结束（或对局中但断线保护期已过期）时，`Room.Run()` 必须结束循环并关闭其 `Inbound` channel，允许 GC 回收资源。
- **防止僵尸连接**：所有 `SendCh` 必须具备缓冲区。若发送缓冲区满且持续时间超过阈值，WritePump 必须主动断开连接，防止慢速客户端撑爆内存。

### 9.2 消息有序性与幂等性 (Ordering)
- **Sequence Number (Seq)**：对局报文必须包含自增的 `Seq`。客户端通过 `Seq` 判断指令是否乱序或丢失；服务端在 `MOVE` 冲突时利用 `Seq` 进行乐观锁判定。
- **幂等重试**：如果客户端未收到落子确认报文并进行重试，服务端应能识别重复的 `Seq` 并返回上一次的结果，而非再次执行业务逻辑。
