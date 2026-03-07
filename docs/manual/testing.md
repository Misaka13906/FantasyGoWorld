# 测试规范 (Testing Strategy)

> **路径**：[`docs/manual/testing.md`](testing.md)  
> **用途**：规范全栈项目的测试方法、覆盖率要求及工具链。

---

## 1. 概览

本项目严格遵循经典的测试金字塔模型：
- **单元测试 (Unit Tests)**：绝大部分的业务组件，尤其是**围棋核心规则引擎**的独立算法隔离验证。
- **集成测试 (Integration Tests)**：覆盖 Web 后端的 Router 接口到 Repository 持久层存储链路。
- **端到端测试 (E2E Tests) / 压测**：系统全流程模拟及大并发下的性能瓶颈测试。

---

## 2. 后端测试规范 (Go)

### 2.1 基础工具链
- 原生 `testing` 包，推荐配合 `t.Run` 进行子测试管理。
- 使用 `github.com/stretchr/testify/assert` 和 `require` 进行断言式测试，取代啰嗦的 `if err != nil` 报错手写。
- 对外依赖网络与缓存的服务使用 `go.uber.org/mock` 自动生成 mock 文件（配合 `go generate`）。

### 2.2 核心单元测试（规则引擎）
对于 `internal/engine` 下的围棋引擎逻辑（如打劫、气数计算、自杀禁止等），**要求达到 100% 的分支覆盖率**。
此包不依赖任何网络或存储，必须大量编写**表驱动测试 (Table-Driven Tests)**，例如定义多组残局 SGF 文本 -> 解析 -> 模拟一子 -> 断言提子数和合法性。

```bash
# 执行单元测试并查看覆盖率报表
go test -v -coverprofile=coverage.out ./internal/engine/...
go tool cover -html=coverage.out
```

### 2.3 接口集成测试
- **测试范围**：`internal/api` 等基于 HTTP 或短线 WS 通信的封装入口。
- **运行环境**：依赖真实的 MySQL 和 Redis 环境，在本地开发中直接连通 `docker-compose` 起的容器环境。
- **数据隔离**：每次执行前可以通过建立沙盒 DB、`Truncate` 洗数据，或在测试代码首尾利用 `Transaction` 回滚，确保测试状态无副作用互不干扰。

```bash
# 用 -tags 隔离集成测试，避免日常跑单测时因为没起容器报错
go test -tags=integration ./test/integration/...
```

### 2.4 测试编写进阶规范 (Best Practices)
- **命名规范**：遵循 `Test<被测函数><场景><预期结果>`，如 `TestPlaceStoneSuicideReturnsError`。
- **代码结构**：严格遵守 `Arrange-Act-Assert` (准备-执行-断言) 三段式结构。
- **杜绝 Flaky Tests (闪烁测试)**：异步测试中严禁使用 `time.Sleep` 进行固定时长盲等。必须利用 `sync.WaitGroup`、Channel 信号或 `testify` 提供的 `require.Eventually` 进行确定性等待。

---

## 3. 前端测试规范 (TypeScript/React)

### 3.1 基础工具链
- 放弃 Jest，使用与 Vite 完美集成的 **Vitest**。
- 组件渲染与 DOM 断言使用 **React Testing Library (RTL)**。
- 强行 Mock 后端请求用 **msw (Mock Service Worker)**。

### 3.2 重点测试区域
- **状态管理 (Zustand Store)**：不用挂载组件，直接测试业务逻辑是否成立，如：`GameStore.dispatch(SyncBoardPayload)` 后，Store 内的倒计时、当前回合数和 19x19 的状态二维数组是否符合预期。
- **棋盘交互 (Board)**：通过 RTL 点击特定坐标，验证 onClick 输出给服务端的落子坐标（如将 DOM 界面的 XY 转化为 0-18 的 `x,y` 索引）是否正确。

```bash
# 前端运行（支持热重启）
npm run test
# 覆盖率检查
npm run test -- --coverage
```

### 3.3 前端测试进阶规范
- **语义化断言**：使用 `describe` 分组模块，使用 `it` 描述用例（如 `it('should render loading spinner when fetching')`）。
- **用户视角选择器 (A11y Priority)**：在使用 RTL 时，强制优先使用 `getByRole`、`getByLabelText` 等无障碍选择器定位元素。极力避免使用依赖 UI 实现细节的 CSS 类名选择器或滥用 `data-testid`。

---

## 4. 并发压力测试 (Load Testing)

对于以**并发大厅 + 对弈实时通信**为核心的游戏业务，性能基准测试与日常压测一样重要。

- **压测工具**：`k6` (采用 Go 开发的高性能压测框架，非常适合压测 WebSocket 持续通信，相比 JMeter 更轻量敏捷)。
- **核心关注场景**：
  1. **巨量连接心跳存活**：模拟单节点 5000+ 个 User 同时在线闲置下发 PingPong 心跳，检查 Hub 单一 goroutine 的消费速度与内存占用。
  2. **高频收发验证**：模拟 1000 个房间（2000 个玩家）以极高频的速度落子，关注平均响应延迟 (Latency) 以及引擎 `PlaceStone` 判断时引起的 CPU Profiling 尖峰。
