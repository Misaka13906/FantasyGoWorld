---
description: 执行完整闭环、由规格驱动原子变更的标准代码编写与提交流程
---

1. **研究领域上下文 (Mandatory Research)**
   - 对于架构方案和前置依赖需要变更的时候：必须使用 `view_file` 工具读取 `docs/decision/log.md`（确保避开已踩坑的方案）以及 `docs/decision/plan.md`（明确当前位于哪个开发阶段）。
   - 对于具体数据结构与接口交互：强制读取并比对 `docs/design/api-spec.md` 及 `docs/design/data-schema.md`。必须严禁依靠大模型幻觉编造字段名和接口。

2. **文档同步更新 (Spec Update)**
   - 如果开发过程发现有必要修改 API 约定、数据库表增加新字段、或者新增 WebSocket Payload 格式，**绝不能直接修改代码**。
   - 必须先分析修改范围，并主动更新 `docs/design/` 下的相关协议设计文件，确保**文档作为 Single Source of Truth 先于代码变更**。

3. **编码实现与原子隔离 (Implementation & Isolation)**
   - 根据规范和设计撰写业务代码、组件或控制层，保持一个逻辑解决一件事（即一原子）。
   - 如果遇到通用、经常易错或者需多次提醒的复杂业务规则，在代码中添加如 `// @ai-context:` 或 `// @logic-hint:` 等 AI 优化注解（Why > What），方便之后检索上下文。

4. **全面测试门禁测试 (Mandatory Testing)**
   - 开始任何 Git 提交前，**必须执行门禁工具链**验证代码有效性。
   - **后端：** 切换到 `FantasyGoWorld-BE` 目录，执行 `go test ./internal/...` 以及 `golangci-lint run ./...` 或 `go mod tidy`，必须保证 0 错误（全绿）。若带红，立刻中止后续步骤并分析错误。
   - **前端：** 对应目录执行状态机单元测试或 E2E 集成测试（如 `npm run test`）。

5. **总结规约与精确反馈持久化 (Persistence of Human Wisdom)**
   - 在 Code Review 阶段收到人类的反馈或者被指出 Bug、坏点子（Bad Smell）时，除了修复代码外，**必须**主动判定该反馈是否具备通用规律，并在修复代码后将其精准持久化到对应的文档中，切勿写错位置：
     - **全局防雷与核心背景约束**：写入 `AGENTS.md`。例如：“本项目直接对外暴露数字 UID，不要脑补 UUID”。
     - **架构决策与选型否决**：追加记录至 `docs/decision/log.md` (ADR)。例如：“人类否决了使用 Redis Pub/Sub 广播消息的提议，因为单机连接数已够用”。
     - **语言级别最佳实践与反模式**：写入 `docs/spec/code-style-go.md` 或是 `code-style-ts.md`。例如：“禁止在 `useEffect` 里执行无清理的轮询”、“Go 规则引擎要求表驱动测试全覆盖”。
     - **AI 协作流程与注释要求变更**：写入 `docs/spec/specification.md`。例如：“人类要求必须标记 `@logic-hint` 或者新增某类自定义 Tag”。

6. **原子化打包与自修复提交 (Git Commit)**
   - 确认测试绿灯，且所有反馈都已在代码和文档侧双向闭环。
   - 使用严谨的 Conventional Commits 格式执行提交（如 `feat(scope): 描述`、`fix(scope): 描述`、`docs(scope): 描述`）。代码变更和相关的文案应包含在同一次原子提交中。