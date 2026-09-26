请生成 `clinical-coldchain-deviation-control`「临床冷链温度偏差处置」Go 全栈项目，面向医药配送中心管理运输容器、温度窗口、偏差事件和放行/隔离决定。不要实现药品销售、采购订单、库存台账或费用结算。

## 项目主要需求

复杂度下限：核心实体不少于 3 个、核心页面不少于 4 个、横切关注点不少于 2 个、共享前端组件不少于 3 个、自定义 hooks/utils 不少于 2 个、后端中间件不少于 2 个。

### 核心实体

`TransportContainer`（容器和传感器）、`TemperatureWindow`（温控规则）、`ExcursionEvent`（偏差事件）、`DispositionDecision`（放行/隔离/报废决定）贯穿数据库、Go 分层和前端。

### 核心页面

`/containers` 容器；`/windows` 温控规则；`/excursions` 偏差处理；`/dispositions` 决定审核；`/audit` 审计。`TemperatureBadge` 在容器和偏差页共用，`DecisionPanel` 在偏差和决定页共用。

### 横切关注点

RBAC 联动角色表、认证中间件、前端路由守卫和按钮显隐；偏差处置审计保存传感器证据、状态迁移和 request ID；全局错误处理与 Redis 限流独立实现。

### 共享枚举/组件

同步 `ContainerState`（ready/in_transit/quarantine/cleared）与 `ExcursionState`（open/in_review/decided/closed）。共享 `StatusBadge`、`EvidenceList`、`ConfirmDialog`，hooks 为 `useAuth`、`usePolling`。

### 技术与规模要求

前端 React 18 + TypeScript + Vite + Material UI；后端 Go 1.22 + Gin + GORM；PostgreSQL、Redis、MinIO。目标 2800–4000 行、28–40 个 `.go` 文件。

### 文件结构强制清单

前端固定 `api/stores/types/components/common/hooks/pages/router/utils`；后端固定 `model/dto/repository/service/handler/router/middleware/constants/util`，禁止单文件堆叠。

### 结构红线

严禁合并职责到单一文件；偏差处置、证据和决定必须各自分层。

### 部署与交付

根目录必须提供 `docker-compose.yml`（顶层 `name: clinical-coldchain-deviation-control`，且不写 `version:`）、`.env` 和 `.env.example`（均含 `COMPOSE_PROJECT_NAME=clinical-coldchain-deviation-control`）、`README.md`、`frontend/Dockerfile`、`backend/Dockerfile` 和 `frontend/nginx.conf`。前端端口 `18511`、后端端口 `19511`；Nginx 反代 `/api`，依赖 healthcheck、命名卷和 `condition: service_healthy`，提供真实 `/healthz` 与 Git 初始化。
