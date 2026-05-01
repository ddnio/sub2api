# Payment B-2 前端与部署复核交接

**日期**：2026-04-30  
**Worktree**：`.claude/worktrees/payment-b2/`  
**分支**：`worktree-payment-b2`  
**PR**：https://github.com/ddnio/sub2api/pull/18

## 当前状态

Payment B-2 已从“仅前端字段同步”扩展为一次完整复核修正：

- 用户 payment API 已改为 upstream payment v2 路径：订单列表使用 `/payment/orders/my`，订单状态轮询读取 `/payment/orders/:id`。
- 后端已补齐 payment v2 route surface，包括 `/payment/config`、`/payment/checkout-info`、`/payment/limits`、public verify/resolve、webhook、admin dashboard、admin retry/refund/provider/plan/config。
- 管理端套餐页已改为读取 `/admin/payment/plans` 的数组响应，不再误用分页 loader。
- 管理端订单页已改为调用 `/admin/payment/dashboard` 和 `/admin/payment/orders/:id/retry`，退款请求体按后端 `{ amount, reason, deduct_balance, force }` 发送。
- 部署文档已修正命令、章节、Provider 配置方式，并移除服务器密码。

## 仍需测试环境验证

本地编译只能证明契约和类型正确，不能替代真实支付验收。合并前必须在测试环境完成：

1. 配置 wxpay Provider（优先按部署手册使用 Admin API 创建完整字段）。
2. 创建/编辑订阅套餐。注意 payment v2 读取 `subscription_plans`；旧 `payment_plans` 会由 `120b` migration 按 `group_id + name` 幂等补齐，但只回填绑定 active subscription 分组的旧套餐。
3. 普通用户进入 `/payment` 下单。
4. 微信扫码支付。
5. 确认订单从 `PENDING` 流转到 `PAID` / `COMPLETED`。
6. 确认 `payment_audit_logs` 有对应记录。

## 关键文件

```text
backend/internal/server/routes/payment.go
backend/internal/server/routes/payment_routes_test.go
frontend/src/api/payment.ts
frontend/src/api/admin/payment.ts
frontend/src/api/__tests__/payment-contract.spec.ts
frontend/src/views/user/PaymentView.vue
frontend/src/views/admin/PaymentOrdersView.vue
frontend/src/views/admin/PaymentPlansView.vue
frontend/src/views/admin/SettingsView.vue
docs/engineering/payment-b2-deploy.md
```

## 验证命令

```bash
cd frontend
pnpm exec vue-tsc --noEmit
pnpm build
pnpm vitest run src/api/__tests__/payment-contract.spec.ts

cd ../backend
GOCACHE="$PWD/../.cache/go-build" go test ./internal/payment ./internal/handler/admin ./internal/handler/dto ./internal/server/routes
GOCACHE="$PWD/../.cache/go-build" go test ./internal/service -run 'Test.*Payment|Test.*Wechat|Test.*WeChat|Test.*Provider|Test.*Order|Test.*Refund|Test.*Fulfillment|Test.*Config'
```

说明：`internal/service` 全包测试在本地沙箱可能被非 payment 的 `httptest.NewServer` 用例阻断；复核 payment 相关测试时使用更窄的 `-run` 范围。
