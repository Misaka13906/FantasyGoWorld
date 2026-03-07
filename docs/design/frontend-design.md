# 前端设计

> **路径**：[`docs/design/frontend-design.md`](frontend-design.md)  
> **用途**：Zustand 全局状态设计（AuthStore/LobbyStore/GameStore）、乐观更新策略、前端路由方案。  
> **关联文档**：[技术设计总纲](./tech-design.md) · [接口协议](./api-spec.md)

---

## 1. 核心设计原则 (Design Principles)

为确保 Agent 开发不偏离航道，前端实现须遵循以下"状态机"原则：

### 1.1 状态驱动 UI (Data-Driven)
*   **所有 UI 变化必须源自 Store 的状态更新**。
*   组件仅作为视图（View），负责将 `state` 渲染为 HTML 片段。

### 1.2 资源回收 (Cleanup)
由于单页应用（SPA）不间断运行，必须严格管理外部资源：
*   **定时器**：所有 `setInterval/setTimeout` 必须在组件卸载（Unmount）时清除。
*   **监听器**：进入房间订阅的 WS 频道或事件监听，在离开房间组件时必须注销，防止内存泄露和逻辑串线。

### 1.3 乐观更新与回滚 (Optimistic UI)
为满足 SLA 中的极低延迟感：
1.  用户点击落子 → **立即**通过 `optimisticMove` 更新本地 Board 呈现。
2.  异步发送 `MOVE` 消息。
3.  若收到服务端 `ERROR` 消息或超时未确认 → 调用 `rollbackMove` **撤销**该落子并弹出提示。

### 1.4 选择器优化 (Selector Optimization)
禁止在组件中全量订阅 Store（如 `const state = useStore()`）。
*   **必须使用 Selector**（如 `const timer = useStore(s => s.timers)`）。这样当计时器每秒跳动时，只有计时器组件重绘，而棋盘组件不会受到波动。

---

## 2. 全局状态管理 (Zustand)

前端使用 **Zustand** 管理全局客户端状态，按页面职责拆分为三个 Store，避免单一大树造成无效重渲染。

```typescript
// ── 1. AuthStore：登录态与当前用户信息 ──────────────────────────────────
interface AuthStore {
  token:     string | null   // access_token（内存持有，不写 localStorage）
  currentUser: {
    id:       number
    username: string
    nickname: string
    elo:      number
    rank:     string         // '18K' ~ '9D'
    dnd:      boolean
  } | null
  status: 'idle' | 'gaming' | 'dnd'  // 本端用户三态，由 WS 推送更新
  setToken:       (token: string) => void
  setStatus:      (s: UserStatus)  => void
  clearAuth:      () => void
}

// ── 2. LobbyStore：大厅页面瞬态数据 ─────────────────────────────────────
interface LobbyStore {
  onlineUsers:   OnlineUserEntry[]       // 在线用户列表（WS 推送增量更新）
  publicRooms:   RoomEntry[]             // 公开房间列表（WS 推送）
  pendingInvite: InvitePayload | null    // 待处理的入局邀请弹窗
  upsertUser:       (u: OnlineUserEntry) => void
  removeUser:       (uid: number)        => void
  upsertRoom:       (r: RoomEntry)       => void
  removeRoom:       (roomId: string)     => void
  setPendingInvite: (inv: InvitePayload | null) => void
}

// ── 3. GameStore：对局室页面核心状态 ─────────────────────────────────────
type Stone = 0 | 1 | 2   // 0=空 1=黑 2=白
type Board = Stone[][]    // [size][size]，行优先

interface GameStore {
  // 房间成员
  members: Member[]                // 含 role: 'black'|'white'|'spectator'
  myRole:  'black' | 'white' | 'spectator'

  // 协商阶段
  proposal:          ProposalPayload | null
  proposalDirection: 'incoming' | 'outgoing' | null

  // 对局阶段
  gameActive:  boolean
  board:       Board     // 乐观渲染：落子后立即更新，被服务端拒绝则回滚
  nextTurn:    'black' | 'white'
  moveHistory: Move[]    // 手顺列表
  stepSeq:     number    // 已渲染的最大 seq，用于防乱序
  timers: {
    black: TimerState
    white: TimerState
  }

  // Actions
  applyServerBoard: (snap: BoardSnapshot) => void  // 收到 SYNC_BOARD 时强制覆盖
  optimisticMove:   (x: number, y: number) => void
  rollbackMove:     () => void
  setProposal:      (p: ProposalPayload | null, dir: 'incoming'|'outgoing'|null) => void
  activateGame:     (config: GameConfig) => void
}
```

> **乐观更新策略**：玩家点击落子时 `optimisticMove` 立即渲染（延迟约 0ms），若服务端返回错误（非法落子/劫争）则在 16ms 内调用 `rollbackMove` 还原，并弹出提示。符合 SLA §3.2 要求。

---

## 3. 路由方案 (Routing)

前端使用 **React Router v6**，采用基于鉴权钩子的装饰器模式。

### 路由表

| 路径 | 页面组件 | 访问权限 | 说明 |
| :--- | :--- | :--- | :--- |
| `/login` | `LoginPage` | Public | 未登录可见；已登录自动重定向至 `/lobby` |
| `/lobby` | `LobbyPage` | Protected | 需登录；大厅在线用户与房间列表 |
| `/room/:id` | `GameRoomPage` | Protected | 需登录；核心对局室 |
| `/` | - | - | 重定向至 `/lobby` |

### 导航守护 (Auth Guards)

1.  **`ProtectedRoute`**：拦截未登录请求。检测 `AuthStore.token` 是否为空，若为空则重定向至 `/login`。
2.  **`PublicRoute`**：防止已登录用户重复登录。若 `token` 已存在，访问 `/login` 会自动重定向回 `/lobby`。

### 页面切换与连接保持

1.  **App 顶层**：负责静默登录请求（`/auth/refresh`）和全局 WebSocket（大厅通知）的维护。
2.  **状态持久化**：路由切换不会触发页面刷新，因此全局 WS 连接在 `/lobby` 和 `/room` 之间切换时保持活跃，无需重新握手。

---

## 4. 页面生命周期与 WS 连接管理

### 4.1 WS 连接归属

WS 连接由 **`<App>` 顶层组件**持有，不属于任何页面组件。这是 SPA 的核心特征：路由切换不销毁 App，WS 连接全程存活。

```
<App>  ← useEffect 在这里建立 WS，并订阅 ws/handlers.ts 分发逻辑
  <AuthProvider>
    /login      ← 无 WS，登录成功后 App 建立连接
    /lobby      ← 消费 LobbyStore（WS 推送填充）
    /room/:id   ← 消费 GameStore（WS 推送填充），mount 时发 JOIN_ROOM
  </AuthProvider>
```

### 4.2 各阶段生命周期

#### App 层（全局，只执行一次）

| 时机 | 动作 |
|:---|:---|
| 用户打开页面 | 调用 `GET /auth/refresh` 静默续期。成功 → 写 AuthStore.token；失败 → 清空 token，重定向 `/login` |
| token 有效（含静默续期成功） | `new WebSocket(/ws)` 建立连接，注册 `ws.onmessage → handlers.ts` |
| WS 断线（网络抖动） | 自动重连（指数退避），重连成功后若在房间页面则发 JOIN_ROOM 恢复上下文 |
| 退出登录 | `ws.close()`，清空 AuthStore / LobbyStore / GameStore |

#### LoginPage（`/login`）

| 时机 | 动作 |
|:---|:---|
| Mount | 若 `AuthStore.token` 已存在，重定向 `/lobby`（PublicRoute 守卫） |
| 登录成功 | 写 token → App 层 `useEffect` 依赖 token 变化，自动建立 WS |
| Unmount | 无需任何清理 |

#### LobbyPage（`/lobby`）

| 时机 | 动作 |
|:---|:---|
| Mount | 无需显式操作。WS 已在 App 层建立，LobbyStore 由 `LOBBY_UPDATE` 推送实时填充 |
| 点击进入房间 | `navigate('/room/101')`，触发 GameRoomPage mount |
| Unmount | 无需清理（WS 继续，LobbyStore 继续接收推送） |

#### GameRoomPage（`/room/:id`）

| 时机 | 动作 |
|:---|:---|
| Mount | ① 清空上一局 GameStore 残留状态；② 发送 `JOIN_ROOM {roomID}` WS 消息 |
| 服务端响应 `MEMBER_UPDATE` | 初始化 GameStore.members、myRole |
| 服务端响应 `SYNC_BOARD` | 填充 GameStore.board（断线重连恢复场景） |
| 用户点击落子 | `optimisticMove` → 发 MOVE → 等待 SYNC_BOARD 确认或 ERROR 回滚 |
| Unmount ① 正常离开（点返回） | 发 `LEAVE_ROOM` WS 消息，清空 GameStore |
| Unmount ② 关闭 Tab | Tab 关闭 → WS 断开 → 服务端心跳超时 → 启动断线保护（对局中）或直接清理 Session |

### 4.3 页面刷新（F5）行为

刷新 = 整个 JS 上下文销毁重建，WS 断开，Store 清空。

```
用户在 /room/101 按 F5：
  1. 新页面加载，App mount
  2. GET /auth/refresh → token 恢复（httpOnly Cookie 自动携带）
  3. new WebSocket(/ws?room=101)  ← URL 带 roomID，query param
  4. 服务端 upgrade 时校验 room 存在，升级成功后自动触发 JOIN_ROOM
  5. 服务端推送 SYNC_BOARD（完整棋盘快照）→ GameStore 恢复
```

前端读取当前路由参数（`useParams().id`）并在 WS 建立时附带，确保刷新后无缝恢复。

### 4.4 房间内切换（`/room/101` → `/room/102`）

React Router 的路由切换会 unmount 旧 GameRoomPage 并 mount 新 GameRoomPage：

```
Unmount /room/101：发 LEAVE_ROOM，清空 GameStore
Mount   /room/102：发 JOIN_ROOM {roomID: 102}，服务端注入新 roomInbound
```

同一条 WS 连接，通过消息切换房间上下文，不重新握手。

### 4.5 多 Tab 行为

每个 Tab 是独立的 JS 上下文，各自建立一条 `/ws` 连接：

```
Tab 1（/lobby）：连 /ws        → roomInbound = nil，接收大厅推送
Tab 2（/room/101）：连 /ws?room=101 → roomInbound 指向 Room 101
```

两个 Tab 各自独立管理生命周期。服务端 Hub 对同一 UserID 的所有活跃连接全部广播大厅消息；Room 消息只发给有对应 roomInbound 的连接。

### 4.6 资源清理契约

| 资源 | 建立时机 | 清理时机 | 清理方式 |
|:---|:---|:---|:---|
| WS 连接 | App mount + token 写入 | 退出登录 / Tab 关闭 | `ws.close()` |
| GameStore 状态 | GameRoomPage mount | GameRoomPage unmount | `resetGameStore()` |
| 计时器 UI（本地 interval） | GAME_START 后 | GAME_END / unmount | `clearInterval` in useEffect cleanup |
| 邀请弹窗 | 收到 INVITE 消息 | 接受/拒绝/超时 | `setPendingInvite(null)` |



### 4.1 Axios 拦截器 (Interceptors)

前端维护单一 `apiClient` 实例，集成以下逻辑：

1.  **请求拦截 (Request)**：
    *   自动从 `AuthStore` 读取 `token`。
    *   若 `token` 存在，注入 `Authorization: Bearer <token>`。
2.  **响应拦截 (Response)**：
    *   **200 (Business Error)**：若后端返回业务错误码（如 `40302` 劫争），拦截器不抛出异常，而是记录日志并返回给调用方处理特殊逻辑。
    *   **401 (Unauthorized)**：**核心逻辑**。
        *   暂停所有后续请求（入队列）。
        *   调用 `POST /auth/refresh`。
        *   若刷新成功，更新 `AuthStore.token` 并重试队列中的请求。
        *   若刷新失败（401/403），清空 Store 并跳转 `/login`。
    *   **500/Net Error**：统一调用全局 Toast 提示。

### 4.2 服务端推送消息路由 (WS Inbound)

所有的 WebSocket 消息在进入 Store 前，由 `ws/handlers.ts` 进行分发：

```typescript
// ws/handlers.ts
export const handleInboundMessage = (msg: Envelope) => {
  switch (msg.type) {
    case 'LOBBY_UPDATE':
      useLobbyStore.getState().handleLobbyUpdate(msg.payload);
      break;
    case 'SYNC_BOARD':
      useGameStore.getState().applyServerBoard(msg.payload);
      break;
    case 'ERROR':
      toast.error(msg.payload.message); // 全局业务报错
      break;
    // ... 其他类型
  }
};
```

---

## 5. 全局 UI 反馈架构

### 5.1 全局 Toast & Modal
*   **不重复造轮子**：使用 `react-hot-toast` 或 `react-toastify` 处理非阻塞通知。
*   **Modal 管理**：重要的协商弹窗（如收到 PROPOSAL）由 `LobbyStore.pendingInvite` 或 `GameStore.proposal` 驱动渲染。

### 5.2 全局 Loading 态
*   **页面级**：路由切换时的 `Suspense` 回退。
*   **按钮级**：提交请求时的 `pending` 状态由组件内部 `useState` 管理，防止重复提交。

---

## 6. 项目文件结构

(已按最新架构调整，见 [`docs/design/tech-design.md`](tech-design.md))

---
