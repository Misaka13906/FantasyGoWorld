# Go 代码规范

> **路径**：[`docs/spec/code-style-go.md`](code-style-go.md)  
> **用途**：后端 Go 代码的命名、格式、错误处理、并发等编码规范，供 Agent 实现代码时遵循。  
> **参考**：[Effective Go](https://go.dev/doc/effective_go) · [Google Go Style Guide](https://google.github.io/styleguide/go) · [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

---

## 1. 格式化（强制）

- **唯一规则**：所有代码提交前必须通过 `gofmt` 格式化。不讨论缩进（Tab）、括号位置等风格问题，`gofmt` 说了算。
- 推荐使用 `goimports`（`gofmt` 超集，同时自动整理 `import` 分组）。
- IDE 配置保存时自动运行 `goimports`，不依赖人工记忆。

---

## 2. 命名规范

### 2.1 可见性

| 标识符类型 | 规则 | 示例 |
|:---|:---|:---|
| 导出（包外可用） | `PascalCase` | `PlaceStone`、`GameStore`、`ErrKo` |
| 未导出（包内私有） | `camelCase` | `countLiberty`、`roomInbound` |

### 2.2 包名

- 全小写，单个单词，不用下划线或混合大小写。
- 名称应能直接说明包的职责：`engine`、`ws`、`repository` ✅，`util`、`common`、`helper` ❌
- 包名是调用侧的前缀，避免冗余：`engine.PlaceStone()`，而不是 `engine.EnginePlace()`。

### 2.3 函数与变量

- 局部变量短小精悍：循环计数 `i`/`j`，错误 `err`，上下文 `ctx` 都是惯例。
- 作用域越大，名字越具体：全局 `gameInstance` > 函数局部 `g`。
- **布尔量**以 `is`/`has`/`can` 开头：`isValid`、`hasLiberty`、`canPlace`。
- **不要**在变量名里重复类型：`userCount` ✅，`countInt` ❌。

### 2.4 常量

- 导出常量用 `PascalCase`：`MaxBoardSize = 19`。
- 包内常量视可读性选 `camelCase` 或 `PascalCase`，同一文件保持一致。
- **不用** `UPPER_SNAKE_CASE`，这不是 Go 惯例。
- **未导出全局变量**：对于未导出的顶层常量和变量，使用 `_` 作为前缀以明确其包作用域（如 `_defaultPort`）。*例外：未导出的错误值可以用 `err` 前缀。*

### 2.5 接口

- **单方法接口**以方法名 + `-er` 后缀命名：`Reader`、`Writer`、`Closer`、`Stringer`。
- 多方法接口用描述性名词：`GameEngine`、`RoomStore`。

### 2.6 缩写与首字母缩略词

- 缩写保持全大写或全小写，不拆分：`HTTPServer`、`urlPath`、`apiClient`。
- ✅ `HTTPRequest`、`userID`、`parseJSON`
- ❌ `HttpRequest`、`userId`、`parseJson`

### 2.7 接收者（Receiver）

- 接收者变量名用类型名的首字母缩写，**1~2 个字母**，且同一类型的所有方法保持一致。

```go
// ✅ 接收者名取类型首字母，简洁一致
func (o *Object) Method() error { ... }
func (o *Object) Update() int { ... }

// ❌ 接收者名不一致或过长
func (obj *Object) Method(...) { ... }
func (o *Object) Update(...) { ... }
```

### 2.9 导入别名
如果包名与路径最后一部分不匹配，必须使用别名。其余情况除命名冲突外，避免使用别名。
```go
import (
    client "example.com/client-go" // 路径与包名不符，必须别名
)
```

### 2.8 文件名

- 全小写，多词用下划线分隔：`user_state.go`、`game_moves.go`。
- 测试文件以 `_test.go` 结尾：`rule_test.go`。

---

## 3. 错误处理

### 3.1 基本原则

- **显式检查**：每个返回 `error` 的调用都必须检查，不允许用 `_` 忽略。
- **尽早返回（Early Return）**：检查到错误立即 `return`，保持主干逻辑左对齐。

```go
// ✅ 尽早返回
func (o *Object) Action(arg int) error {
    if arg < 0 {
        return ErrInvalidArg
    }
    if o.isLocked() {
        return ErrLocked
    }
    // ... 主干逻辑
    return nil
}

// ❌ 嵌套地狱
func (o *Object) Action(arg int) error {
    if arg >= 0 {
        if !o.isLocked() {
            // ... 主干逻辑
        }
    }
    return nil
}
```

### 3.2 错误包裹（Wrapping）

使用 `fmt.Errorf` + `%w` 添加上下文，保留原始错误链，便于 `errors.Is`/`errors.As` 查询。

```go
// ✅
if err := service.DoSomething(ctx, data); err != nil {
    return fmt.Errorf("pkg.Caller: failed to do something: %w", err)
}

// ❌ 丢失原始错误
return fmt.Errorf("保存失败: %v", err)  // %v 不可 unwrap
```

### 3.3 哨兵错误（Sentinel Errors）

预定义的、调用方需要显式处理的错误，在包级别导出，以 `Err` 开头。

```go
// pkg/errors.go
var (
    ErrNotFound = errors.New("resource not found")
    ErrTimeout  = errors.New("operation timeout")
)

// 调用方
if errors.Is(err, pkg.ErrNotFound) {
    // 专项处理
}
```

### 3.4 错误只记录一次

不要既 `log` 又 `return err`——这会导致同一个错误被打印多次。
- **底层**（repository/engine）：直接返回 wrapped error，不打日志。
- **最顶层**（controller/Hub）：记录日志并决定如何响应。

### 3.5 Panic 的使用边界

**业务代码里几乎不使用 `panic`。**

- 可预期的错误（参数非法、状态不符）→ **返回 `error`**，哪怕是在初始化函数里。
- 进程无法继续运行（DB 连不上、配置文件缺失）→ **`log.Fatal`**（内部调用 `os.Exit(1)`，干净退出）。
- `panic` 仅保留给真正"不该发生、发生了就是代码有 bug"的场景，在本项目中预计**几乎为零**。

```go
// ✅ 调用方传非法参数 → return error，由调用侧决定如何处理
func NewObject(config Config) (*Object, error) {
    if config.Value == "" {
        return nil, fmt.Errorf("NewObject: missing value")
    }
    return &Object{value: config.Value}, nil
}

// ✅ 启动时 DB 连不上，进程无意义继续 → log.Fatal
func main() {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to connect database: %v", err)
    }
    ...
}

// ❌ 任何业务路径里都不应出现 panic
func (o *Object) Method(...) {
    if somethingWrong { panic("oops") }  // 应 return error
}
```

### 3.6 日志级别决策

写日志前先问自己：**这行日志是写给谁看的？**

```
给开发者看？
├─ 不需要记录变量状态 → TRACE
└─ 需要记录变量状态   → DEBUG

给运维人员看：
因异常状态触发？
├─ 否 → INFO（如 gin 自带的请求日志）
└─ 是 → 进程能带着这个异常继续运行？
        ├─ 是 → WARN（降级、重试等容忍场景）
        └─ 否 → 应用还能继续处理其他请求？
                ├─ 是 → ERROR（当前请求/任务失败，服务整体存活）
                └─ 否 → FATAL（进程无法继续，log.Fatal 退出）
```

**本项目场景映射**：

| 场景 | 级别 |
|:---|:---|
| 开发阶段打印入参、中间变量 | `DEBUG` |
| WebSocket 客户端心跳超时断线 | `WARN` |
| 某次落子 Redis 写入失败（单局受影响，服务存活） | `ERROR` |
| 启动时 MySQL / Redis 连接失败 | `FATAL`（`log.Fatal`） |

---

## 4. 注释规范


- **导出符号必须有文档注释**，以符号名开头，完整句子结尾。
- 包注释写在 `doc.go` 或当前包的任意一个文件顶部。
- 函数内部的逻辑注释说明 **why**，不是 what。

```go
// PlaceStone 在棋盘 (x, y) 处落子，并执行提子逻辑。
// 若落子违反禁入或打劫规则，返回对应哨兵错误。
func (b *Board) PlaceStone(x, y int, c Color) error { ... }

// 并查集根节点合并后，必须重新计算气数，
// 因为提子操作可能使相邻棋串重新获得气。
b.mergeStrings(x, y)
```

---

## 5. 并发规范

### 5.1 Goroutine 生命周期管理

- **严禁 fire-and-forget**：所有 `go func()` 必须有明确的退出机制，防止 goroutine 泄漏。
- 使用 `context.Context` 控制生命周期；使用 `sync.WaitGroup` 等待退出。

```go
// ✅
func (w *Worker) start(ctx context.Context, wg *sync.WaitGroup) {
    defer wg.Done()
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            w.execute()
        case <-ctx.Done():
            return
        }
    }
}

// ❌ 缺失退出机制
go func() {
    for {
        time.Sleep(time.Second)
        doSomething()
    }
}()  // 永远不会退出，导致内存泄漏
```

### 5.2 Channel 尺寸
Channel 的容量要么是 **0 (无缓冲)**，要么是 **1 (带缓冲)**。任何超过 1 的容量必须经过严格设计审查。
- `ch := make(chan int)` ✅
- `ch := make(chan int, 1)` ✅
- `ch := make(chan int, 64)` ❌ (除非有极端性能理由)
```

### 5.2 Channel 方向声明

函数参数中的 channel 应声明方向，明确收发意图。

```go
func producer(out chan<- Message) { ... }  // 只写
func consumer(in <-chan Message)  { ... }  // 只读
```

### 5.3 避免数据竞争

- 共享状态的读写必须有 `sync.Mutex` 或 `sync.RWMutex` 保护。
- 所有测试须加 `-race` flag：`go test -race ./...`。

### 5.4 Mutex 规范
- **禁止嵌入**：即使是私有结构体，也不要直接嵌入 `sync.Mutex`。应作为私有字段使用。
- **零值有效**：`var mu sync.Mutex` 即可使用，不需要 `new(sync.Mutex)`。
- **保护边界**：在函数边界（接收或返回）处，必须对 Slice 和 Map 进行**防御性拷贝**，防止外部逻辑意外修改内部状态。

```go
// ✅ 正确做法
type Stats struct {
    mu       sync.Mutex
    counters map[string]int
}

func (s *Stats) Snapshot() map[string]int {
    s.mu.Lock()
    defer s.mu.Unlock()
    res := make(map[string]int, len(s.counters))
    for k, v := range s.counters { res[k] = v } // 防御性拷贝
    return res
}
```

---

## 6. 核心习惯用法 (Go Idioms)

### 6.1 Context 的使用
- **首个参数**：如果函数需要 `context.Context`，它必须是跨层级调用的**第一个参数**，且命名为 `ctx`。
- **禁止存储**：严禁将 `context.Context` 存储在结构体字段中。`Context` 应该显式地随着函数调用链传递。

```go
// ✅
func (s *Service) DoWork(ctx context.Context, arg string) error { ... }

// ❌ 错误做法：存入 struct
type Service struct {
    ctx context.Context
}
```

### 6.2 接口与结构体设计
- **Accept Interfaces, Return Structs**（接收接口，返回结构体）。
- 返回具体结构体让调用方可以直接访问数据或方法，并且未来可以随时在此结构体上实现新的接口。
- 将接口定义在**调用方**（使用者）包内，而不是实现方包内。这有助于保持低耦合和真正的按需抽象。

### 6.3 表格驱动测试 (Table-Driven Tests)
- 对于含有多种边界条件的逻辑校验函数，强制使用表格驱动测试模式。
- 利用切片和匿名结构体组织测试用例，配合 `t.Run(tc.name, ...)` 分离失败输出。

---

### 6.4 接口合理性验证 (Compile-time Check)
在导出类型时，使用 `var _ Interface = (*Type)(nil)` 确保类型在编译期确实实现了目标接口。
```go
type Handler struct{}
var _ http.Handler = (*Handler)(nil) // 编译期校验
```

### 6.5 避免裸参数 (Avoid Naked Parameters)
当函数参数（尤其是布尔值或整数）语义不明时，使用 C 风格注释或自定义类型。
- `printInfo("foo", true /* isLocal */, true /* done */)` ✅
- `printInfo("foo", true, true)` ❌ (语义不明)

### 6.6 功能选项模式 (Functional Options)
对于具有多个可选参数的构造函数，优先使用 Functional Options 模式以保持扩展性。

---

## 7. 工具链配置（待项目初始化时添加）

```
golangci-lint   # 静态分析（集成 errcheck、govet、staticcheck 等）
goimports       # 格式化 + import 整理
go test -race   # 带竞态检测的测试
```
