# 技术设计文档（总纲）

> **路径**：[`docs/design/tech-design.md`](tech-design.md)  
> **用途**：技术架构总纲，包含技术选型、架构分层图、子文档索引和关键架构决策记录（ADR）、前后端代码文件夹结构。

---

## 技术选型

| 层次 | 技术 |
|:---|:---|
| 前端 | React + TypeScript（Vite）|
| 后端 | Golang · Gin + GORM |
| 主数据库 | MySQL 8.0+ |
| 缓存 / 实时状态 | Redis |
| 实时通信 | WebSocket（每个对局室页面 = 一个连接）|

---

## 架构分层概览

```
┌─────────────────────────────────────┐
│               Browser               │
│   React + Zustand + WS Client       │
└────────────────┬────────────────────┘
                 │  HTTP / WebSocket
┌────────────────▼────────────────────┐
│           Gin HTTP Server           │
│  controller → biz → repository      │
├──────────────────────────────────────┤
│         WebSocket Hub / Room        │
│   Hub（控制平面） Room（数据平面）     │
└──────┬──────────────────┬───────────┘
       │                  │
┌──────▼──────┐   ┌───────▼────────┐
│    MySQL    │   │     Redis       │
│  持久化存储  │   │  store / cache  │
└─────────────┘   └────────────────┘
```

### 交叉领域关注点 (Cross-Cutting Concerns)
现代架构不仅关注层次调用，也必须考虑横向贯穿全局的非功能性约束（NFR，Non-Functional Requirements）。相关的详细设计已拆分至专属文档：
- **服务分级与容错**：SLA 延迟目标、数据一致性模型及断线保护，见 [`docs/spec/sla.md`](../spec/sla.md)。
- **安全与反作弊**：JWT Token 续期策略、同 IP 过滤防御，部分见 [`docs/design/frontend-design.md`](frontend-design.md)（乐观更新回滚）及 SLA 规范。
- **可观测性与部署**：普罗米修斯监控埋点、滚动更新发布方案，见 [`docs/manual/devops.md`](../manual/devops.md)。

---

## 核心文档索引

全局项目所有相关设计报告及设计决策，现已统一收敛至项目根目录的[`AGENTS.md`](../../AGENTS.md) (面向 AI 和人类的双视角入口)，请作为第一入口访问。历史散落的子文档均可通过该总索引触达。

---

## 关键架构决策记录 (ADR)

> 详细的决断过程与防重蹈覆辙建议，详见 [`docs/decision/log.md`](../decision/log.md)。

## 项目文件结构

### 前端 (React + TypeScript · Vite)

```
fantasy-go-world-fe/
├── public/
├── src/
│   ├── api/                      # HTTP 请求封装（axios instance + 各模块 API）
│   │   ├── http.ts               # axios 拦截器：token 注入、401 刷新
│   │   ├── auth.ts
│   │   ├── user.ts
│   │   ├── room.ts
│   │   └── game.ts
│   ├── ws/                       # WebSocket 客户端管理
│   │   ├── wsClient.ts           # 连接管理、心跳、重连策略
│   │   └── handlers.ts           # 消息 type 分发 → 更新对应 store
│   ├── store/                    # Zustand 状态管理
│   │   ├── authStore.ts          # 登录态 & 当前用户
│   │   ├── lobbyStore.ts         # 大厅在线用户 & 房间列表
│   │   └── gameStore.ts          # 对局室棋盘、协商、计时
│   ├── pages/
│   │   ├── LoginPage.tsx
│   │   ├── LobbyPage.tsx         # 大厅：在线列表 + 房间列表
│   │   └── GameRoomPage.tsx      # 对局室：棋盘 + 协商面板 + 聊天
│   ├── components/
│   │   ├── Board/
│   │   │   ├── GoBoard.tsx       # SVG/Canvas 棋盘渲染
│   │   │   └── Stone.tsx
│   │   ├── Lobby/
│   │   │   ├── UserListItem.tsx
│   │   │   └── RoomListItem.tsx
│   │   ├── Proposal/
│   │   │   └── ProposalPanel.tsx # 协商配置面板（弹窗）
│   │   ├── Timer/
│   │   │   └── GameTimer.tsx     # 计时器显示（读秒 + 包干）
│   │   └── Chat/
│   │       └── ChatBox.tsx
│   ├── types/                    # 全局 TypeScript 类型定义
│   │   ├── ws.ts                 # Envelope, Payload 类型
│   │   ├── game.ts               # Board, Stone, GameConfig ...
│   │   └── user.ts
│   ├── App.tsx                   # 路由配置（React Router v6）
│   └── main.tsx
├── index.html
├── vite.config.ts
└── package.json
```


### 附录：后端 (Go · Gin + GORM)

```
fantasy-go-world-be/
├── cmd/
│   └── server/
│       └── main.go               # 程序入口，依赖注入组装
├── internal/
│   ├── api/
│   │   ├── middleware/
│   │   │   ├── auth.go           # JWT 解析 & 注入 uid 到 gin.Context
│   │   │   └── cors.go
│   │   ├── controller/
│   │   │   ├── auth.go           # 注册/登录/刷新/登出
│   │   │   ├── user.go           # 用户信息 CRUD
│   │   │   ├── room.go           # 房间创建/列表/关闭
│   │   │   └── game.go           # 对局详情/SGF 下载/历史
│   │   └── router/
│   │       └── router.go         # 路由注册（gin.Engine 组装）
│   ├── biz/                      # 业务逻辑层（Service，跨 repo 编排）
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── room.go
│   │   └── game.go
│   ├── repository/
│   │   ├── model/                # GORM 模型结构体
│   │   │   ├── user.go
│   │   │   ├── game.go
│   │   │   ├── room.go
│   │   │   ├── game_rule.go
│   │   │   └── game_timer.go
│   │   ├── db/                   # MySQL 交互（GORM）
│   │   │   ├── user.go
│   │   │   ├── room.go
│   │   │   └── game.go
│   │   ├── store/                # Redis 主存储（Source of Truth，无 TTL）
│   │   │   ├── user_state.go
│   │   │   ├── room_store.go
│   │   │   └── game_moves.go
│   │   └── cache/                # Redis 缓存（DB 副本，有 TTL）
│   │       └── user.go           # user_profile:{uid} Hash, TTL:5min
│   ├── ws/                       # WebSocket 核心
│   │   ├── hub.go                # Hub：连接注册/注销，控制消息路由
│   │   ├── client.go             # Client：ReadPump / WritePump
│   │   ├── room.go               # Room：成员管理，消息广播
│   │   └── envelope.go           # 统一报文 Envelope 定义
│   ├── engine/                   # 规则引擎（逻辑、计分、合法性）
│   │   ├── board.go
│   │   └── rule.go
│   └── config/
│       └── config.go             # 配置结构体 + Viper 加载
├── pkg/
│   ├── jwt/                      # JWT 生成与解析工具
│   ├── e/                        # 统一错误码定义
│   ├── sgf/                      # SGF 解析器与生成器（可独立开源）
│   │   ├── parser.go             # 接口定义 (Parser & Formatter)
│   │   ├── linear.go             # 线性单链处理实现 (P1)
│   │   └── tree.go               # 树状 AST 分析实现 (P2)
│   └── response/                 # gin 统一响应封装
├── migrations/                   # SQL 迁移文件（按版本命名）
│   └── 001_init.sql
├── local/                   # 部署文件
│   └── config.yaml
└── go.mod
```