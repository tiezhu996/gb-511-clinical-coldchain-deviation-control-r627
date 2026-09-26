
# 临床冷链温度偏差处置

医药运输容器、温控规则、偏差事件与质量处置平台。项目采用前后端分离和明确的领域分层，重点保证状态迁移、RBAC、审计日志、请求追踪与限流在各层保持一致。

## Docker Compose 快速启动

```bash
cp .env.example .env
docker compose up -d --build
docker compose ps
```

- Web 工作台：http://127.0.0.1:18511
- 后端健康检查：http://127.0.0.1:19511/healthz
- 后端 API：http://127.0.0.1:19511/api
- 演示账号：`viewer`、`operator`、`reviewer`、`admin`，密码均为 `Admin123!`（仅限本地演示）

停止并清理本项目容器与数据卷：

```bash
docker compose down -v --remove-orphans
```

## 主要功能

| 业务模块 | 后端实体 | API 前缀 | 状态流 |
|---|---|---|---|
| 运输容器 | `TransportContainer` | `/api/containers` | ready, in_transit, quarantine, cleared |
| 温控规则 | `TemperatureWindow` | `/api/windows` | draft, active, expired, superseded |
| 偏差事件 | `ExcursionEvent` | `/api/excursions` | open, in_review, decided, closed |
| 处置决定 | `DispositionDecision` | `/api/dispositions` | draft, release, quarantine, discard |
| 传感器证据 | `SensorEvidence` | `/api/evidence` | 不可变登记记录 |

- JWT 登录和 viewer/operator/reviewer/admin 四级 RBAC。
- 用户与启用角色在每次请求时回查角色表，令牌中的旧角色不能绕过停用或降权。
- 偏差与处置状态变化使用乐观锁，并与 request ID、前后状态、证据一起原子写入审计日志。
- 传感器证据拥有独立实体和五层后端实现，通过 MinIO 生成限时上传地址并保存 SHA-256 元数据。
- 处置决定实行双人复核：提议人不能批准自己的提议，最终决定不可编辑或反向迁移；偏差没有最终处置时不能关闭。
- 请求 ID、结构化日志、全局错误映射和 Redis 分布式限流。
- 提供脱敏运行配置、当前会话、审计汇总和单实体审计历史接口。
- 业务工作台支持查询、新建、状态推进、风险标识及操作审计查看。

## 技术栈

| 层次 | 技术 |
|---|---|
| 前端 | React 18 + TypeScript + Vite + Material UI |
| 后端 | Go 1.22 + Gin + GORM |
| 数据 | PostgreSQL + Redis、MinIO |
| 部署 | Docker Compose + Nginx |

## 本地开发

后端可使用 SQLite 开发模式，不需要先启动数据库：

```bash
cd backend
go mod download
DATABASE_DRIVER=sqlite DATABASE_DSN=local.db REDIS_ADDR='' \
JWT_SECRET=local-development-secret PORT=8080 go run ./cmd/server
```

前端开发服务器：

```bash
cd frontend
npm install
npm run dev
```

质量检查：

```bash
cd backend && go test ./... && go build ./...
cd ../frontend && npm run typecheck && npm run build
cd .. && docker compose config --quiet
```

也可以从项目根目录执行 `./scripts/validate.sh`，脚本会构建、启动、检查健康接口和鉴权 API，并在结束时关闭容器。

## 目录结构

```text
.
├── backend/
│   ├── cmd/server/                 # 服务入口与优雅退出
│   └── internal/
│       ├── config/                 # 环境配置
│       ├── constants/              # 状态枚举与迁移图
│       ├── database/               # 连接、迁移与演示数据
│       ├── dto/                    # 输入契约
│       ├── handler/                # HTTP 接口
│       ├── middleware/             # JWT、追踪、限流
│       ├── model/                  # GORM 实体
│       ├── repository/             # 持久化边界
│       ├── router/                 # 路由装配
│       ├── service/                # 业务规则与审计
│       └── util/                   # 统一 HTTP 响应
├── frontend/src/
│   ├── api/                        # 按实体拆分的 API
│   ├── components/common/          # 共享业务组件
│   ├── hooks/                      # 认证与分页 hooks
│   ├── pages/                      # 五个路由页面
│   ├── router/                     # 路由配置
│   ├── stores/                     # 按实体拆分的状态仓库
│   ├── types/                      # 共享类型与枚举
│   └── utils/                      # 格式化与状态工具
├── docker-compose.yml
└── runtime_smoke.json
```

## 共享枚举位置

| 枚举 | 值 | 前后端出现位置 |
|---|---|---|
| `ContainerState` | `ready, in_transit, quarantine, cleared` | `backend/internal/constants/status.go`、`frontend/src/types/status.ts` |
| `ExcursionState` | `open, in_review, decided, closed` | `backend/internal/constants/status.go`、`frontend/src/types/status.ts` |

每个实体自己的完整迁移图同样位于 `backend/internal/constants/status.go`；页面使用的状态列表位于 `frontend/src/types/status.ts`。修改状态时必须同步两处并更新对应服务测试。

## 环境变量

| 变量 | 说明 |
|---|---|
| `COMPOSE_PROJECT_NAME` | 固定英文 Compose 项目名，支持中文父目录 |
| `DB_NAME/DB_USER/DB_PASSWORD` | 数据库名称与业务账号 |
| `DB_ROOT_PASSWORD` | MySQL 管理员密码（PostgreSQL 项目保留统一模板字段） |
| `JWT_SECRET` | JWT 签名密钥，生产环境必须替换 |
| `FRONTEND_PORT/BACKEND_PORT/DB_PORT` | 宿主机端口映射 |
| `REDIS_PORT` | Redis 宿主机端口 |
| `MINIO_*` | 证据对象存储配置（启用 MinIO 的项目） |

## API 使用示例

```bash
token=$(curl -sS -X POST http://127.0.0.1:19511/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Admin123!"}' | jq -r '.data.token')

curl -sS http://127.0.0.1:19511/api/overview \
  -H "Authorization: Bearer $token"
```

## License

MIT
