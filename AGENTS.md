# Fantasy Go World — Agent Context

一句话简介：在线围棋对弈平台，Go + React，单体架构。

---

## 🚫 严禁脑补协议 (Anti-Hallucination Protocol)

作为 Agent，你**严禁**基于通用训练数据猜测本项目特有的 API 结构、字段名或错误码。**必须执行以下“两步走”策略**：

1.  **研究阶段 (Mandatory Research)**：在任何代码改动前，必须使用 `view_file` 读取：
    -   [`docs/design/api-spec.md`](docs/design/api-spec.md)（获取接口定义）
    -   [`docs/design/data-schema.md`](docs/design/data-schema.md)（获取数据库定义）
    -   [`docs/spec/specification.md`](docs/spec/specification.md)（获取框架习惯）
2.  **严控输出 (Verification)**：如果文档中没有定义，**严禁自行发明**，必须先更新设计文档（或询问人类确认）后再编写代码。

---

## ⚡ 每次任务开始前必须过的 Checklist

> 完整工作流规范见 [`docs/spec/specification.md`](docs/spec/specification.md)，以下为内联摘要，**开始任何代码工作前先确认这几条**。

- [ ] **涉及架构/技术选型变更**？→ 先读 [`docs/decision/log.md`](docs/decision/log.md)，确认没有提出已否决的方案
- [ ] **开始新功能实现**？→ 先查 [`docs/decision/plan.md`](docs/decision/plan.md) 确认当前阶段和前置依赖是否就绪
- [ ] **涉及接口或数据结构变更**？→ 先更新对应 `docs/design/` 文档，再写代码
- [ ] **完成一个逻辑变更或脚手架搭建后**？→ **必须**运行 `go test ./...`。即便没有编写测试文件，也必须确保代码能通过编译且 `go mod tidy` 无错输出。
- [ ] **涉及依赖变更**？→ **禁止**使用多个 `go get` 分次安装，统一使用 `go mod tidy` 进行依赖梳理和清理。
- [ ] **commit message 格式**：`<type>(<scope>): <desc>`，例如 `feat(ws): inject roomInbound on JOIN_ROOM`

如果你已确认上述原则已遵守，请在与人类对话的最结尾加上一句话：喵喵，已完成！
---

## 工作流规范摘要

### 提交粒度
"原子"指**逻辑范围**，不是文件数量。同一问题的改动可横跨多个文件，合为一个 commit；不同问题的改动绝不混在一起。

### 测试门禁

| 场景 | 需要跑的测试 | 时机 |
|:---|:---|:---|
| 改业务逻辑 / Repository 层 | 单元测试 `go test ./internal/...` | 每次 commit 前 |
| 改 controller / middleware / router | 单元测试 + 接口测试（httptest） | 每次 commit 前 |
| PR 合并到 main | 集成测试（真实 MySQL + Redis） | 发版前，手动执行 |
| 纯文档变更 | 无需跑测试 | — |

---

## 重要文档索引 (全局唯一索引站)

> **目录结构与文档说明**：本项目的所有结构化设计、需求及操作文档均统一归档于 `docs/` 目录下。作为 AI 和人类开发者的**全局唯一索引栈 (Single Source of Truth)**，请查阅以下分类导航。

### 1. 核心规范与落地手册 (`docs/spec/` & `docs/manual/`)
| 文件 | 说明 |
|:---|:---|
| [`docs/spec/specification.md`](docs/spec/specification.md) | ⭐ **项目规范**（Git 提交、测试工作流、AI 工作流）|
| [`docs/spec/code-style-go.md`](docs/spec/code-style-go.md) | Go 语言后端代码格式及模式规范 |
| [`docs/spec/code-style-ts.md`](docs/spec/code-style-ts.md) | TypeScript 语言前端组件状态规范 |
| [`docs/spec/sla.md`](docs/spec/sla.md) | API 响应时间、打点与系统可用性标准指标 |
| [`docs/manual/local-dev.md`](docs/manual/local-dev.md) | 本地环境 Docker 起步、config.yaml 字段与常用命令 |
| [`docs/manual/testing.md`](docs/manual/testing.md) | 全栈自动化测试策略（单测/集成/E2E）与并发压测大纲 |
| [`docs/manual/devops.md`](docs/manual/devops.md) | Docker 构建方案、CI/CD 与可观测性 (Metrics) 指南 |

### 2. 架构与设计 (`docs/design/`)
| 文件 | 说明 |
|:---|:---|
| [`docs/design/tech-design.md`](docs/design/tech-design.md) | 总纲：技术栈概览、分层架构拓扑与**前后端代码目录结构** |
| [`docs/design/rule-engine.md`](docs/design/rule-engine.md) | 围棋引擎：算法选用、提子/自杀等合法性实现及 SGF 序列化演进 |
| [`docs/design/data-schema.md`](docs/design/data-schema.md) | 数据层：MySQL 物理表及字段、Redis Key 体系与缓存分层 |
| [`docs/design/ws-design.md`](docs/design/ws-design.md) | WebSocket：高并发架构、单线程锁控制、Room 长生命周期流转 |
| [`docs/design/api-spec.md`](docs/design/api-spec.md) | 接口字典：REST API (非结构化) 及所有 WS 结构 Payload 汇总 |
| [`docs/design/openapi.yaml`](docs/design/openapi.yaml) | 机器可读：全量 REST API 的 OpenAPI 3.0 (Swagger) 严密定义 |
| [`docs/design/asyncapi.yaml`](docs/design/asyncapi.yaml) | 机器可读：全量 WebSocket 消息协议的 AsyncAPI 2.6 定义 |
| [`docs/design/frontend-design.md`](docs/design/frontend-design.md) | 前端架构：Zustand 数据流编排、乐观更新、组件渲染树 |
| [`docs/design/diagrams.md`](docs/design/diagrams.md) | UML图库：跨端认证/鉴权时序、对弈断线重连时序与核心状态机 |

### 3. 需求与领域备忘 (`docs/requirement/` & `docs/reference/`)
| 文件 | 说明 |
|:---|:---|
| [`docs/requirement/requirement.md`](docs/requirement/requirement.md) | 极简 PRD：围绕大厅及对局室提炼的原子化功能点 |
| [`docs/requirement/business-model.md`](docs/requirement/business-model.md) | 领域建模：核心实体属性与彼此的流转依赖生命周期 |
| [`docs/reference/sgf.md`](docs/reference/sgf.md) | 格式指南：SGF(棋谱) 的结构解析与 Go 盘面专用坐标系换算 |
| [`docs/reference/game-rule.md`](docs/reference/game-rule.md) | 补录法理：真实世界中围棋日规与中规的对弈边缘判例说明 |

### 4. 架构决策与发版推进 (`docs/decision/`)
| 文件 | 说明 |
|:---|:---|
| [`docs/decision/log.md`](docs/decision/log.md) | ⭐ **架构决策记录 (ADR)**，所有历史定论归档，防重复提出已否决方案 |
| [`docs/decision/plan.md`](docs/decision/plan.md) | **实现分镜计划**，界定从零到 P1/P2/P3 的里程碑和开发编码次序 |
| [`docs/decision/why.md`](docs/decision/why.md) | ⭐ 痛点 QA 汇总：记录项目发展期针对特殊难点选型的思辨与解答 |
--

## 已确定的关键决策
- 技术选型：前端 React + Vite + TypeScript，后端 Go + Gin + GORM + MySQL + Redis
- 主键：`BIGINT AUTO_INCREMENT`，对外直接暴露数字 UID
- 认证：JWT（access_token 1h + refresh_token 7d httpOnly Cookie）
- WS 鉴权：Gin Middleware 优先读 Authorization Header，fallback 读 Cookie
- Repository 分层：`db/`（MySQL）、`store/`（Redis 主存储，无 TTL）、`cache/`（Redis 缓存，有 TTL）
