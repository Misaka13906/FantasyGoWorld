# 实现顺序计划

> **路径**：[`docs/decision/plan.md`](plan.md)  
> **用途**：面向 Agent 的实现顺序指南，规定各模块的开发顺序和依赖关系，防止依赖模块未就绪就被调用。

---

## 开发原则

- **依赖先行**：被依赖的模块必须先实现并通过测试，再开始依赖它的模块
- **垂直切片**：每个阶段完成后系统应处于可运行状态，不留半成品
- **测试随行**：每个模块完成时同步补齐单元测试，不攒到最后

---

## 阶段一：项目脚手架（P0）

> 目标：空项目能跑起来，基础框架就位

```
[x] 初始化 Go 模块（go mod init）
[x] 搭建目录结构（cmd/ internal/ pkg/ migrations/）
[x] config.yaml 加载（viper）
[x] MySQL 连接初始化（GORM）
[x] Redis 连接初始化（go-redis）
[x] Gin 路由基础框架（router.go，含 CORS、health check /ping）
[x] 统一响应封装（pkg/response）
[x] 统一错误码定义（pkg/e）
[x] JWT 工具（pkg/jwtauth：生成/解析 access + refresh token）
[x] 数据库初始化文件（migrations/001_init.sql）
[x] 前端：初始化 Vite + React + TypeScript 项目
[x] 前端：配置 axios instance（pkg/api/http.ts，含拦截器）
```

---

## 阶段二：用户认证（P0）

> 目标：用户可以注册、登录、刷新 token、退出

**后端**：
```
[x] GORM 模型（repository/model/user.go）
[x] Repository 层（repository/db/user.go）：CreateUser、GetByUsername、GetByID
[x] Auth Biz 层（biz/auth.go）：Register、Login、Refresh、Logout
[x] Auth Controller（api/controller/auth.go）
[x] 鉴权 Middleware（api/middleware/auth.go）：Header → Cookie fallback
[x] 接口测试：POST /auth/register、/auth/login、/auth/refresh、/auth/logout
```

**前端**：
```
[x] AuthStore（store/authStore.ts）
[x] Login 页面（pages/LoginPage.tsx）
[x] api/auth.ts（register、login、refresh、logout 请求封装）
[x] 401 自动刷新 token 拦截器（http.ts）
```

---

## 阶段三：大厅 HTTP 初始化（P0）

> 目标：登录后能获取在线用户列表和公开房间列表（HTTP，不含 WS 推送）

**后端**：
```
[x] GORM 模型（model/room.go）
[x] Repository 层（db/user.go 扩展：GetOnlineList；db/room.go：Create、List、GetByID、Delete）
[x] User Biz 层（biz/user.go）：GetProfile、UpdateProfile、GetOnlineList
[x] Room Biz 层（biz/room.go）：CreateRoom、ListPublicRooms、CloseRoom
[x] User Controller（/user）
[x] Room Controller（/room）
[x] 接口测试：GET /user/list、POST /room、GET /room/list
```

**前端**：
```
[x] LobbyStore（store/lobbyStore.ts）
[x] LobbyPage 骨架（HTTP 初始化数据，暂无 WS 推送）
[x] api/user.ts、api/room.ts
```

---

## 阶段四：WebSocket 核心连接（P0）

> 目标：客户端能建立 WS 连接，Hub 管理连接生命周期，心跳正常

**后端**：
```
[ ] Envelope 定义（ws/envelope.go）
[ ] Client 结构体（ws/client.go）：ReadPump、WritePump、心跳检测
[ ] Hub（ws/hub.go）：Register/Unregister、连接映射、心跳扫描
[ ] WS 升级入口（router 注册 GET /ws）
[ ] Redis Store：user_state（store/user_state.go）
```

**前端**：
```
[ ] wsClient.ts：连接管理、心跳发送、自动重连
[ ] App.tsx 登录后建立 WS 连接
```

---

## 阶段五：大厅实时推送（P0）

> 目标：用户上线/下线、房间创建/关闭实时推送到大厅所有客户端

**后端**：
```
[ ] Hub 处理 LOBBY_UPDATE 广播（Hub.broadcastLobbyUpdate）
[ ] 用户上线/下线触发 LOBBY_UPDATE
[ ] 房间创建/关闭触发 LOBBY_UPDATE（增量推送：users_add/remove、rooms_add/remove）
```

**前端**：
```
[ ] ws/handlers.ts：处理 LOBBY_UPDATE → upsertUser/removeUser/upsertRoom/removeRoom
[ ] LobbyPage 接入 WS 增量更新
```

---

## 阶段六：房间与对局协商（P0）

> 目标：用户能进入/离开房间，两人能完成对局配置协商流程

**后端**：
```
[ ] Room 事件循环（ws/room.go）：Register、Inbound 消息分发
[ ] Hub 处理 JOIN_ROOM：注入 roomInbound 到 Client
[ ] Hub 处理 LEAVE_ROOM、PROPOSAL（转发协商消息）
[ ] Redis Store：room_store（store/room_store.go）：成员列表、配置快照
[ ] MEMBER_UPDATE 广播
[ ] PROPOSAL 协商流：propose → counter → accept/reject → GAME_START
[ ] games 表 GORM 模型（model/game.go）
```

**前端**：
```
[ ] GameStore（store/gameStore.ts）协商阶段字段
[ ] GameRoomPage 骨架
[ ] ProposalPanel（components/Proposal/ProposalPanel.tsx）
[ ] WS handlers：JOIN_ROOM、MEMBER_UPDATE、PROPOSAL、GAME_START
```

---

## 阶段七：围棋规则引擎（P0）

> 目标：规则引擎通过所有单元测试，可被 Room 调用

```
[ ] internal/engine/board.go：Board 类型、Color 枚举、邻格/边界工具
[ ] internal/engine/string_set.go：并查集（Union-Find）
[ ] internal/engine/rule.go：PlaceStone、CountLiberty、RemoveString（禁入+Ko）
[ ] internal/engine/score.go：CalculateScore（**仅中国规则**：子空皆地，贴目 7.5）
[ ] internal/engine/dead.go：MarkDead、ConfirmDead
[ ] 单元测试（全部通过后才进入下一阶段）：
    - 普通落子+提子
    - 禁入（自杀）检测
    - 简单劫（Ko）检测
    - 多子同时提
    - 中国规则计分
```

---

## 阶段八：对局核心流程（P1）

> 目标：两人能完整下完一局围棋（落子、提子、认输）

**后端**：
```
[ ] Room 集成规则引擎：处理 MOVE → ValidateMove → SYNC_BOARD 广播
[ ] Redis Store：game_moves（store/game_moves.go）：实时落子流 RPUSH
[ ] PASS 处理：连续两次 Pass 进入死子结算
[ ] RESIGN 处理：立即终局，生成 GAME_END
[ ] GameInstance 内存结构（含黑白方分配）
[ ] 乱序检测（step 字段校验）
```

**前端**：
```
[ ] GoBoard.tsx：SVG/Canvas 棋盘渲染
[ ] Stone.tsx：棋子渲染
[ ] 乐观落子（optimisticMove → rollbackMove）
[ ] WS handlers：MOVE、SYNC_BOARD、PASS、RESIGN、GAME_END
```

---

## 阶段九：计时系统（P1）

> 目标：读秒制和包干制计时正常工作，超时自动判负

```
[ ] 服务端时钟（goroutine 定时器，每秒 tick）
[ ] TimerState 数据结构（黑白方各自 remaining_ms、byoyomi_count）
[ ] SYNC_BOARD 中携带 timers 字段
[ ] 超时判负逻辑 → GAME_END
[ ] 前端 GameTimer.tsx：倒计时显示、读秒闪烁
```

---

## 阶段十：断线保护（P1）

> 目标：断线后有 180 秒保护期，重连可恢复

```
[ ] 心跳超时检测（3秒无心跳标记离线）
[ ] 暂停计时器，向对方推送 NOTIFICATION{opponent_disconnected}
[ ] 断线保护倒计时（180 秒 goroutine）
[ ] 重连后下发完整 SYNC_BOARD 快照
[ ] 保护期超时：GAME_END{disconnection_forfeit}
```

---

## 阶段十一：SGF 生成与对局记录（P1）

> 目标：对局结束后生成 SGF 文件，写入数据库

```
[ ] internal/engine/sgf.go：落子序列 []Move + 对局配置元信息 → .sgf 文本
[ ] 对局结束时：从 Redis List 读取完整落子流（game_moves:{room_id}）→ 生成 SGF → 写 games.sgf_data
[ ] 清理 Redis game_moves 数据
[ ] GET /game/:gameId 接口
[ ] GET /game/:gameId/sgf 下载接口
[ ] GET /user/:uid/games 历史对局列表
```

---

## 阶段十二：Elo 等级分（P1）

```
[ ] Elo 计算函数（pkg/elo）
[ ] 对局结算时更新双方 elo 和 rank
[ ] GAME_END payload 携带 elo_delta 字段
```

---

## 暂不实现（P2，后续版本）

- 日本规则计分（score.go 补充 RuleJapanese 分支）
- 悔棋协商（REQ-UT-01）
- 棋谱回放（REQ-SG-03）
- 快速匹配（REQ-RN-05 匹配部分）
- 大厅邀请（REQ-RN-05 邀请部分）