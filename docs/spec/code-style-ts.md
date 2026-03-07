# TypeScript / React 代码规范

> **路径**：[`docs/spec/code-style-ts.md`](code-style-ts.md)  
> **用途**：前端 TypeScript + React 代码的命名、组件、Hook、类型等编码规范，供 Agent 实现代码时遵循。  
> **参考**：[react.dev](https://react.dev/learn) · [@typescript-eslint 官方文档](https://typescript-eslint.io) · [Airbnb React Style Guide](https://github.com/airbnb/javascript/tree/master/react)

---

## 1. 格式化（强制）

- **唯一规则**：所有代码提交前必须通过 **Prettier** 格式化。不讨论单引号/双引号、分号、缩进等风格，Prettier 说了算。
- ESLint 负责**逻辑规则**（未使用变量、Hook 规则等），Prettier 负责**格式规则**，两者不重叠。
- IDE 配置保存时自动运行 Prettier，不依赖人工记忆。

---

## 2. 命名规范

### 2.1 总览

| 标识符类型 | 规则 | 示例 |
|:---|:---|:---|
| React 组件 | `PascalCase` | `GoBoard`、`ProposalPanel`、`GameTimer` |
| 组件文件 | `PascalCase.tsx` | `GoBoard.tsx`、`ChatBox.tsx` |
| 自定义 Hook | `use` + `PascalCase` | `useGameStore`、`useWsClient` |
| 非组件文件 | `camelCase.ts` | `wsClient.ts`、`handlers.ts` |
| 变量 / 函数 | `camelCase` | `nextTurn`、`applyServerBoard` |
| 常量（模块级） | `UPPER_SNAKE_CASE` | `MAX_RECONNECT_ATTEMPTS`、`WS_TIMEOUT_MS` |
| 类型 / 接口 | `PascalCase` | `BoardSnapshot`、`TimerState` |
| Props 类型 | `XxxProps`（组件名 + Props） | `GoBoardProps`、`ProposalPanelProps` |
| 枚举成员 | `PascalCase` | `Stone.Black`、`Stone.White` |

### 2.2 组件命名细则

```tsx
// ✅ 组件名 = 文件名（PascalCase）
// 文件：UserCard.tsx
export const UserCard: React.FC<UserCardProps> = ({ user, isAdmin }) => { ... }

// ❌ 不一致
```

### 2.3 Props 类型

- 使用 `type`（而非 `interface`）定义 Props，以 `Props` 后缀结尾。
- **不加 `I` 前缀**（`IGoBoardProps` ❌）。

```tsx
// ✅
type UserCardProps = {
  user:    User
  isAdmin: boolean
  onEdit:  (id: string) => void
  theme?:  'dark' | 'light'
}

// ❌
interface IUserCardProps { ... }
```

### 2.4 布尔 Props

布尔类型的 prop 以 `is`/`has`/`can`/`show` 开头：

```tsx
type ButtonProps = {
  isActive:  boolean
  hasIcon:   boolean
  canClick:  boolean
  showLabel: boolean
}
```

### 2.5 导出规范 (Exports)

- **统一使用命名导出 (Named Exports)**，禁止使用默认导出 (`export default`)，除非框架机制（如 React.lazy 路由懒加载）强制要求。
- 相比默认导出，命名导出在跨文件重构、IDE 自动导入时更安全，严格保证了消费者名字和提供者一致。

---

## 3. 类型系统规范

### 3.1 禁止 `any`

- 禁止使用 `any`，用 `unknown` + 类型守卫代替，或定义具体类型。
- `tsconfig.json` 开启 `"strict": true`，包含 `noImplicitAny`。

```ts
// ❌
const handleMessage = (msg: any) => { ... }

// ✅
const handleMessage = (msg: unknown) => {
  if (!isEnvelope(msg)) return
  // ... msg 在此处被收窄为 Envelope
}
```

### 3.2 类型 vs 接口

- **Props / 本地数据形态**：用 `type`（语义更精确，支持联合类型）。
- **可扩展的模块契约（如 Repository 层接口）**：用 `interface`。

### 3.3 枚举

- 数字枚举不直观，**优先使用字符串字面量联合类型**：

```ts
// ✅ 字符串联合（可读、可序列化）
type Status = 'pending' | 'active' | 'closed'
type Step = 1 | 2 | 3

// ❌ 数字枚举（日志里是数字，难以调试）
enum Status { Pending = 0, Active = 1 }
```

---

## 4. 组件规范

### 4.1 只使用函数组件

- 不使用 Class Component，统一使用函数组件 + Hooks。

### 4.2 组件职责单一

- 一个组件只做一件事。超过 **150 行**考虑拆分为子组件。
- 避免深层 JSX 嵌套（超过 3 层考虑拆子组件）。

### 4.3 Props 解构

- 在函数参数处解构 Props，不在函数体内用 `props.xxx`。

```tsx
// ✅
const UserCard = ({ user, isAdmin, onEdit }: UserCardProps) => { ... }

// ❌
const UserCard = (props: UserCardProps) => {
  const user = props.user  // 冗余
}
```

### 4.4 条件渲染

- 优先使用逻辑与 `&&` 或三元表达式，避免过深 `if/else` 嵌套。
- 复杂条件抽取为变量或子组件。

```tsx
// ✅
const MessageList = ({ messages }: MessageListProps) => {
  if (messages.length === 0) return null  // 提前返回，避免嵌套
  return <ul>...</ul>
}
```

### 4.5 语义化与可访问性 (a11y)

- **优先使用语义化 HTML 标签**：可交互的点击元素必须使用 `<button>` 或 `<a>`，严禁用 `<div onClick={...}>` 模拟（会导致键盘 Tab 焦点和屏幕阅读器失效）。
- 图像必须包含 `alt` 属性。

### 4.6 性能优化 (Memoization)

- **避免过早优化**：绝大多数 React 正常组件的重渲染开销极小，不需要包裹 `React.memo`、`useMemo` 或 `useCallback`。
- 仅在渲染图表、巨型列表等**真实遭遇性能瓶颈**的昂贵组件处，或需要向下传递极其稳定的对象引用以避免死循环时，才引入记忆化。

---

## 5. Hook 规范

### 5.1 自定义 Hook 命名

- 必须以 `use` 开头（React Lint 规则强制）：`useWsClient`、`useGameTimer`。

### 5.2 单一职责

- 每个 Hook 只封装一个关注点。如果一个 Hook 超过 **80 行**，考虑拆分。

### 5.3 依赖数组

- `useEffect`/`useMemo`/`useCallback` 的依赖数组必须完整，不允许残留警告。
- 使用 `eslint-plugin-react-hooks` 的 `exhaustive-deps` 规则检查。

```tsx
// ✅
useEffect(() => {
  const subscription = api.subscribe(id)
  return () => subscription.unsubscribe()  // 必须清理
}, [id])                                   // 依赖完整

// ❌ 空依赖数组但内部用了外部变量
useEffect(() => {
  log(id, data)               // id 是外部变量
}, [])
```

### 5.4 Zustand Store 订阅

遵循 `frontend-design.md §1.4`，必须用 **Selector** 精准订阅，不全量订阅：

```ts
// ✅ 只在 timers 变化时重渲染
const timers = useGameStore(s => s.timers)

// ❌ 任何 store 字段变化都触发重渲染
const state = useGameStore()
```

---

## 6. 禁止事项（Anti-patterns）

| 禁止 | 原因 | 替代方案 |
|:---|:---|:---|
| `var` | 函数级作用域，行为不直观 | `const` / `let` |
| `any` | 绕过类型检查 | `unknown` + 类型守卫 / 具体类型 |
| `I` 前缀接口（`IUser`） | TypeScript 社区约定不加前缀 | `User`、`UserProps` |
| 全量 Store 订阅 | 导致无效重渲染（见 §5.4） | Selector 精准订阅 |
| 火忘式副作用（无清理） | 内存泄漏、组件卸载后仍执行 | `useEffect` 返回清理函数 |
| 在 JSX 里写复杂逻辑 | 可读性差 | 抽出变量或子组件 |
| `React.FC` 泛型（过度使用） | 隐含 `children` 类型，语义不明确 | 直接写函数签名 + Props 类型 |

---
 
## 7. 注释规范 (Annotation Standards)

- **组件与 Hook**：复杂的业务组件或逻辑 Hook 必须提供简要的 JSDoc，说明其核心职责。
- **AI 友好型注释**：在非直观的副作用处理、复杂的 RxJS/Zustand 状态同步或特定的 UI 权衡处，**必须**使用结构化标签（详见 [`specification.md`](./specification.md#15-ai-辅助注释规范-ai-optimized-annotation)）。

```tsx
/**
 * useGameTimer - 处理对局读秒逻辑
 * @logic-hint: 
 * 1. 优先使用本地时间差计算，每秒与后端 SYNC_BOARD 校准一次。
 * 2. 读秒小于 10s 时触发音频预加载，防止网络延迟导致铃声卡顿。
 */
export const useGameTimer = (gameId: string) => {
  // ...
  // @ai-fix: 强制在卸载时 clearAll 所有的 window.requestAnimationFrame
  useEffect(() => { ... }, []);
}
```

---

## 8. 工具链配置（待项目初始化时添加）

```
eslint                    # 逻辑规则检查
@typescript-eslint        # TypeScript 专项规则
eslint-plugin-react-hooks # Hook 规则（exhaustive-deps）
prettier                  # 格式化
```

推荐 `tsconfig.json` 必须开启的选项：

```json
{
  "compilerOptions": {
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true
  }
}
```
