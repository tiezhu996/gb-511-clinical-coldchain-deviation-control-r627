# 验收记录

验收日期：2026-08-22（Asia/Shanghai）

## 静态检查

- `go test ./...`：通过
- `go vet ./...`：通过
- `go build ./...`：通过
- `npm run typecheck`：通过
- `npm run build`：通过，Vite 生产包生成成功
- `docker compose config --quiet`：通过
- Go 源文件 40 个，共 3526 行，符合 28–40 个文件、2800–4000 行要求

## 空库部署

执行 `docker compose down -v --remove-orphans` 后重新运行 `docker compose up -d --build`。PostgreSQL、Redis、MinIO、backend、frontend 五个服务均达到 `healthy`，`GET /healthz` 返回数据库和 Redis `ready`。迁移与种子数据在空卷环境完成。

## API 红线

- viewer 可读但写接口返回 403；operator 不能审批处置
- operator 尝试 `quarantine -> in_transit` 返回 422，不能绕过质量放行
- admin 自提议后自审批返回 422；reviewer 独立审批成功
- 已审批处置的 PUT 和二次状态迁移均返回 422
- 偏差完成证据登记与评估后，没有最终处置仍不能关闭；独立审批处置后可关闭
- `/api/evidence` 返回 MinIO 15 分钟预签名上传 URL，并保存 SHA-256、对象键、容器和偏差关联
- 审计记录包含 request ID、前态、后态、证据、提议人和复核人
- 运行时停用 viewer 角色后，既有 JWT 立即返回 403；恢复角色后同一 JWT 恢复只读访问
- 运行时停用 viewer 用户后，下一次页面请求返回 401，React 会话立即回到登录页

## 内置 Browser

仅使用内置 Browser 验证，未使用外部 Chrome：

- 五个页面均可访问：运输容器、温控规则、偏差处理、处置审核、审计追踪
- 实际完成容器质量放行、温控规则生效、偏差闭环和独立处置审批
- 温度组件按规则显示正常/越界，偏差页读取独立 SensorEvidence 元数据
- viewer 看不到写按钮和审计导航，直接访问 `/audit` 会重定向到 `/containers`
- 桌面与 390×844 移动视口无横向溢出，移动端仅渲染一个详情面板
- 浏览器控制台 error 和 warning 均为 0

## 清理

验收结束后关闭 Browser 标签，并执行 `docker compose down -v --remove-orphans`；本项目容器、网络和三个命名卷均已移除。
