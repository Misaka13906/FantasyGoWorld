# DevOps 与可观测性指南

> **路径**：[`docs/manual/devops.md`](devops.md)  
> **用途**：指导项目的容器化封装、CI/CD 自动化流水线配置及生产上线的可观测性（监控与告警）搭建。

---

## 1. 容器化方案 (Docker)

本工程不论是后端还是前端，全部贯彻 **Multi-stage build (多阶段构建)** 原则，极大地压缩产物镜像体积并确保安全。

### 1.1 后端 Dockerfile
基于 Go 的静态编译优势，使用 `alpine` 甚至是 `scratch`（空白系统）运行最终产物。
```dockerfile
# --------- 构建阶段 ---------
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# 禁用 CGO 进行完全静态链接
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# --------- 运行阶段 ---------
FROM alpine:latest
# 安装时区、CA 根证书
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
ENTRYPOINT ["./server"]
```

### 1.2 前端 Dockerfile
基于 Vite 编译产生静态资源（HTML/JS/CSS），最终挂载于轻量级 Web 容器中即可。
```dockerfile
# --------- 构建阶段 ---------
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# --------- 运行阶段 ---------
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
# 建议：额外拷贝一份自定义的 nginx.conf 用于配置前端路由的 SPA fallback (即所有 404 回退给 /index.html)
EXPOSE 80
```

---

## 2. CI/CD 流水线 (Github Actions 等)

### 2.1 PR 门禁自动检查 (CI)
针对针对 `main` 分支提交的 Pull Request 自动执行，杜绝"格式丑陋及破损构建"进库。
- **后端任务**：并行执行 `go fmt`, `golangci-lint` 进行偏执级别代码检查，紧接着运行完整 `go test -cover`，如果覆盖率出现跌落则阻断合并。
- **前端任务**：执行 `npm run lint`, `npm run test`，并尝试 build 验证是否有 TypeScript 编译期类型错误。

### 2.2 发布流水线 (CD)
当研发合并并打上 `v1.x.x` 的 Tag 后触发发布发布流程：
1. 自动注入 Git Commit ID 和版本号进入编译的 `-ldflags` 参数。
2. 构建并打包出上文的双端镜像产物。
3. 把镜像推送到镜像中心（如 DockerHub 或 阿里云 ACR）。
4. （可选）触发远程生产环境服务器拉取 `docker-compose pull` 或直接部署进 Kubernetes 的 `Deployments` 中滚更。

### 2.3 部署策略与密钥管理 (Deployment & Secrets)
- **零停机部署**：生产环境推荐使用 **蓝绿部署 (Blue-Green)** 或 **滚动更新 (Rolling Update)**，确保在长连接对局较少时进行平滑升级，或由负载均衡器优雅排出 (Drain) 旧连接。
- **凭证安全**：严禁在代码仓库中硬编码任何 Password/Secret/Token。所有涉密配置（如 Redis 密码、JWT Secret）必须通过**环境变量**注入或使用单独的凭证管理工具（如 HashiCorp Vault、AWS Secrets Manager、GitHub Actions Secrets）在 CI 阶段动态传递。

---

## 3. 运维生产级可观测性 (Observability)

为了避免“黑盒运行”，尤其是确保实时对局在发生阻塞、内存爆炸前得到预警，系统原生需搭建三大分支。

### 3.1 业务与系统指标 (Metrics)
建议向基础代码里引入 Prometheus 采集器（如 `prometheus/client_golang`），暴露 HTTP `/metrics` 端口：
*   **长连接健康度**：`ws_online_users` (在线玩家)、`active_rooms` (活跃棋室)。这些指标直接证明业务当前“有多火热”。
*   **引擎性能**：`rule_engine_duration_seconds` (核心逻辑计算耗时统方图)，监控围棋引擎是否有递归死活带来的 CPU 计算毛刺现象。
*   **基础设施内存**：基于默认暴露出来的 `go_memstats_alloc_bytes`（Goroutine 与内存分配监控），严防内存泄漏。

然后配置一个高大上的 **Grafana Dashboard** 持续监控大盘。

### 3.2 结构化日志 (Logs)
- **全面摒弃 `fmt.Printf`**：使用 Go 原生最新的 `log/slog`（或经典的 `uber-go/zap`）并强制统一输出为 **JSON Logging** 格式。这样可以在排查时将长长的结构化栈拆解过滤。
- **日志采集池**：可以通过轻量的 `Promtail + Loki` 或者传统的 `Filebeat + ELK` 汇入收集池集中查询业务报错及 `Level = ERROR` 的信息。

### 3.3 离线告警 (Alerting)
对接 Grafana 的 Alertmanager，发送通知到指定的告警群（如飞书群）：
*   当错误日志（5xx Server Error）突然激增每分钟超过 50 个。
*   当 Go 后端的活跃 Goroutine 数量持续稳定突破阈值无法 GC 被释放。
*   当 MySQL 或 Redis 请求延时持续走高。
