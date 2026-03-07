# 本地开发环境启动指南

> **路径**：[`docs/manual/local-dev.md`](local-dev.md)  
> **用途**：本地开发环境搭建、配置文件字段定义、前后端启动命令、常用开发命令速查。
> **关联手册**：
> - 测试规范与要求，请参阅 [`testing.md`](./testing.md)
> - 生产部署、CI/CD 与可观测性，请参阅 [`devops.md`](./devops.md)

---

## 1. 前置依赖

| 工具 | 版本要求 | 用途 |
|:---|:---|:---|
| Go | 1.22+ | 后端运行时 |
| Node.js | 20+ | 前端运行时 |
| Docker + Docker Compose | 任意稳定版 | 启动 MySQL + Redis |
| Git | - | 版本控制 |

---

## 2. 基础设施启动方案

开发者可根据习惯选择：**Docker Compose (方案 A)** 或 **本机服务 (方案 B)**。

### 💡 极简启动：直接使用本地服务 (localhost) —— 方案 B
如果你本地已安装并运行了 MySQL 和 Redis，这是最快的方式：

1. **MySQL 准备**:
   - 确认运行在 `3306`。
   - 创建数据库：`CREATE DATABASE fantasy_go_world CHARACTER SET utf8mb4;`。
   - (可选) 创建专用账号并授权，或使用 root（生产不可行，本地测试自由）。
2. **Redis 准备**: 确认运行在 `6379`。
3. **配置文件**: 复制并创建 `config.yaml`（见第 3 节）。我们将该文件中的 `host` 默认设为 `127.0.0.1`。

### 🐳 推荐方案：使用 Docker Compose —— 方案 A
适用于不想在宿主机安装数据库，或追求环境严格一致的场景：

```bash
# 自动启动 MySQL + Redis（配置已固化在 docker-compose.yml 中）
docker compose -f local/docker-compose.yml up -d
```
> **注意**：方案 A 启动后，数据库的数据会持久化在 `./local/mysql-data`。

---

## 3. 配置文件 (config.yaml) 设置
复制 `config.example.yaml` 为 `config.yaml`（此文件已在 `.gitignore` 中，防止泄露个人账号）。

---

## 3. config.yaml 完整字段定义

本地开发时需要手动创建完整的配置文件（不提交到 Git）。
标准流程：将根目录的 `config.example.yaml` 复制为 `config.yaml` 并根据本地环境按需修改。

```yaml
server:
  port: 8080           # HTTP 服务端口
  mode: debug          # debug | release | test（传给 Gin）

database:
  host: 127.0.0.1
  port: 3306
  name: fantasy_go_world
  user: fgw
  password: fgw123
  charset: utf8mb4
  max_open_conns: 20   # 最大开放连接数
  max_idle_conns: 10   # 最大空闲连接数

redis:
  host: 127.0.0.1
  port: 6379
  password: ""         # 本地无密码
  db: 0                # 使用的 DB 编号

jwt:
  access_secret: "your-access-secret-key-local"   # access_token 签名密钥（≥32字符）
  refresh_secret: "your-refresh-secret-key-local" # refresh_token 签名密钥（≥32字符）
  access_expire: 3600        # access_token 有效期（秒），1h
  refresh_expire: 604800     # refresh_token 有效期（秒），7d

cors:
  allowed_origins:
    - "http://localhost:5173"  # Vite 前端开发服务器
  allow_credentials: true

websocket:
  heartbeat_timeout: 3        # 无心跳判定离线阈值（秒）
  disconnect_protect: 180     # 断线保护时长（秒）
  max_message_size: 65536     # 单条消息最大字节（64KB）
```

**配置文件加载方式**：使用 `viper`，按以下顺序加载（后者覆盖前者）：
1. `config.yaml`（默认，开发环境）
2. 同名环境变量（生产环境注入，如 `DATABASE_HOST`）

---

## 4. 后端启动 (FantasyGoWorld-BE)

> **注意**：后端代码独立存放在 `FantasyGoWorld-BE/` 子目录下，所有 Go 相关的操作均需在该目录执行。

```bash
# 1. 复制配置
cd FantasyGoWorld-BE
cp local/config.example.yaml local/config.yaml

# 2. 安装依赖 (初次运行)
go mod tidy 

# 3. 启动开发服务器 (自动重载推荐用 air)
# 注：若代码中通过 --config 参数指定路径，需传参 local/config.yaml
go run ./cmd/server

# 或使用 air 热重载
air

# 4. 执行代码检查 (Lint)
golangci-lint run
```

---

## 5. 前端启动

```bash
# 进入前端目录
cd fantasy-go-world-fe

# 安装依赖
npm install

# 启动开发服务器（默认 http://localhost:5173）
npm run dev

# 运行单元测试（vitest）
npm run test

# 构建生产包
npm run build
```

前端环境变量（`.env.local`，不提交 Git）：

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
```

---

## 6. 数据库迁移

```bash
# 第一次初始化（Docker Compose 启动时自动执行 migrations/001_init.sql）
# 后续新增迁移文件按版本号命名：002_add_xxx.sql、003_alter_xxx.sql ...

# 手动执行（必要时）
docker exec -i <mysql-container> mysql -ufgw -pfgw123 fantasy_go_world < migrations/002_xxx.sql
```

---

## 7. 常用速查

```bash
# 查看所有运行的容器
docker ps

# 连接 MySQL（调试用）
docker exec -it <container-id> mysql -ufgw -pfgw123 fantasy_go_world

# 连接 Redis（调试用）
docker exec -it <container-id> redis-cli

# 生成 Go 接口 mock（依赖 mockgen）
go generate ./internal/...
```

---

## 8. 常见排错与问题诊断 (Troubleshooting)

| 报错现象 | 常见原因 | 解决办法 |
|:---|:---|:---|
| `bind: address already in use` (端口 3306 或 6379 冲突) | 本地已常驻 MySQL / Redis 进程 | 停止本地服务，或在 `docker-compose.yml` 中修改映射端口（如 `3307:3306`），并同步修改 `config.yaml`。 |
| 前端调用后端接口跨域拦截 (CORS Error) | `.env.local` 配置错误或后端 CORS 未放行域名 | 检查 `VITE_API_BASE_URL` 端口是否为后端实际启动端口，检查后端 `config.yaml` 中的 `allowed_origins` 是否包含前端 dev server 的地址。 |
| 后端启动读取不到配置 (panic: Fatal error config file) | 根目录缺少 `config.yaml` | 确认已执行 `cp config.example.yaml config.yaml`。 |
| Docker 挂载数据卷没有写权限 | Linux/macOS 用户侧权限问题 | 赋予挂载目录权限：`sudo chown -R 1001:1001 ./mysql-data` 或清理后重建。 |
