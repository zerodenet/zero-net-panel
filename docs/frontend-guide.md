# 前端项目开发指南

本文档面向需要对接 Zero Network Panel（ZNP）后端的前端团队，覆盖环境配置、鉴权策略、接口映射和常见坑位。

## 1. 项目定位与角色

- **管理端（Admin Console）**：节点、模板、套餐、公告、订单、第三方安全配置等运维/运营功能。
- **用户端（User Portal）**：订阅与套餐展示、公告、余额、订单购买与取消。

后端区分角色访问：管理端接口要求 `admin` 角色，用户端接口要求 `user` 角色。

## 2. 环境配置建议

建议在前端通过环境变量管理 API 地址与路由前缀：

- `API_BASE_URL`：如 `http://localhost:8888`
- `API_PREFIX`：固定为 `/api/v1`
- `ADMIN_PREFIX`：默认 `admin`（需要与后端 `Admin.RoutePrefix` 一致）

示例拼接规则：

- 管理端：`${API_BASE_URL}/api/v1/${ADMIN_PREFIX}`
- 用户端：`${API_BASE_URL}/api/v1/user`

## 3. API 客户端设计

推荐封装统一的请求层：

- 自动拼接 base URL 与前缀
- 自动注入 `Authorization: Bearer <token>`
- 全局处理 `401/403/429/5xx`
- 支持单飞刷新（避免并发刷新导致令牌覆盖）

示例（伪代码）：

```ts
async function request(url, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) };
  const token = getAccessToken();
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_BASE_URL}${url}`, { ...options, headers });
  if (res.status === 401) {
    await refreshTokenOnce();
    return request(url, options);
  }
  return res.json();
}
```

## 4. 鉴权与刷新

- 登录接口：`POST /api/v1/auth/login`
- 刷新接口：`POST /api/v1/auth/refresh`

建议流程：

1. 登录成功后缓存 `access_token` 与 `refresh_token`。
2. 访问接口时注入 `Authorization` 头。
3. 遇到 `401` 时用 `refresh_token` 换新令牌，再重试一次原请求。

> 建议 `access_token` 存于内存（减少 XSS 风险），`refresh_token` 放在更安全的存储（如 HttpOnly Cookie 或受控存储）。

## 5. 页面与接口映射

### 管理端

- 仪表盘：`GET /api/v1/{adminPrefix}/dashboard`
- 节点列表/同步：`GET /nodes`、`POST /nodes/{id}/kernels/sync`
- 订阅模板：`GET/POST/PATCH /subscription-templates`、`POST /subscription-templates/{id}/publish`
- 套餐管理：`GET/POST/PATCH /plans`
- 公告管理：`GET/POST /announcements`、`POST /announcements/{id}/publish`
- 安全配置：`GET/PATCH /security-settings`
- 订单管理：`GET /orders`、`GET /orders/{id}`、`POST /orders/{id}/pay|cancel|refund`

### 用户端

- 订阅列表/预览/切换模板：`GET /subscriptions`、`GET /subscriptions/{id}/preview`、`POST /subscriptions/{id}/template`
- 套餐列表：`GET /plans`
- 公告列表：`GET /announcements`
- 余额与流水：`GET /account/balance`
- 订单：`POST /orders`、`GET /orders`、`GET /orders/{id}`、`POST /orders/{id}/cancel`

完整字段说明请参考 `docs/api-reference.md`，或使用 `./scripts/gen-api-docs.sh` 生成的 `docs/api-generated/`。

## 6. 数据格式与展示建议

- **金额**：`*_cents` 为分单位，展示时建议 `amount_cents / 100` 并配合 `currency`。
- **流量**：`traffic_limit_bytes`、`traffic_used_bytes` 建议使用二进制单位（GB/TB）。
- **时间**：所有 `*_at` 字段为 Unix 秒（UTC），前端需本地化显示。
- **订单状态**：
  - `status`：`pending_payment`、`paid`、`payment_failed`、`cancelled`、`partially_refunded`、`refunded`
  - `payment_status`：`pending`、`succeeded`、`failed`
- **套餐状态**：`draft`、`active`（未激活套餐前端可隐藏）

## 7. 订单与支付流程提示

- `POST /user/orders` 支持 `payment_method=balance|external`。
- `payment_method=external` 且金额大于 0 时，需要传 `payment_channel`，响应会带 `payment_intent_id` 与 `payments`。
- 推荐前端传 `idempotency_key`（如点击下单时生成 UUID），避免重复下单。

## 8. 第三方签名开关

用户端路由统一挂载第三方签名中间件：

- 若 `security_settings.third_party_api_enabled=true` 且 `api_key/api_secret` 生效，前端必须附带签名头。
- 浏览器端不适合存储 `api_secret`，建议在后台关闭该开关或通过 BFF 服务代签名。

## 9. 管理端访问限制

管理端可能开启 IP 白名单与速率限制（`Admin.Access`）：

- 前端部署地址需在允许网段内。
- 被限流时返回 `429`，可做提示与退避重试。

## 10. 本地联调建议

1. 启动后端：

```bash
go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --apply --seed-demo
go run ./cmd/znp serve --config etc/znp-sqlite.yaml --migrate-to latest
```

2. 默认账号：

- 管理员：`admin@example.com` / `P@ssw0rd!`
- 用户：`user@example.com` / `P@ssw0rd!`

3. API Base：`http://localhost:8888/api/v1`

## 11. 常见问题排查

- `401`：检查 token 过期、角色是否匹配。
- `403`：检查角色、IP 白名单或第三方签名是否启用。
- `409`：通常是并发更新/状态冲突，前端应提示重试。
- `422/400`：检查必填字段与格式是否匹配。
