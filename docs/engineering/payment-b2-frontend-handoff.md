# Payment B-2 前端与部署复核交接

**日期**：2026-05-01
**Worktree**：`.claude/worktrees/payment-b2/`
**分支**：`worktree-payment-b2`
**PR**：https://github.com/ddnio/sub2api/pull/18

## 当前状态

Payment B-2 已从“仅前端字段同步”扩展为一次完整 payment v2 前端迁移与复核修正：

- 用户 payment API 已改为 upstream payment v2 路径：订单列表使用 `/payment/orders/my`，订单状态轮询读取 `/payment/orders/:id`。
- 用户充值/订阅页已迁移到 upstream `PaymentView`，主入口为 `/purchase`；旧 `/payment` 路由保留兼容但不再作为侧边栏入口。
- 用户订单中心已迁移到 upstream `UserOrdersView`，入口为 `/orders`。
- 支付结果、二维码、Stripe、微信支付回调页面已迁移：`/payment/result`、`/payment/qrcode`、`/payment/stripe`、`/payment/stripe-popup`、`/auth/wechat/payment/callback`。
- 后端已补齐 payment v2 route surface，包括 `/payment/config`、`/payment/checkout-info`、`/payment/limits`、public verify/resolve、webhook、admin dashboard、admin retry/refund/provider/plan/config。
- 后端 `/payment/limits` / `/payment/checkout-info` 保持 upstream 语义：前端展示的可用支付方式范围来自 provider instance 的 `limits`；全局 `MIN_RECHARGE_AMOUNT` / `MAX_RECHARGE_AMOUNT` 仍在创建余额充值订单时校验。部署配置时如果希望 0.1 元可付，必须同时确认对应 provider 的 `singleMin` 为空/0 或不高于 0.1。
- 管理端支付看板、订单页、套餐页已迁移到 upstream `views/admin/orders/*`，兼容现有 fork 路由 `/admin/payment/dashboard`、`/admin/payment/orders`、`/admin/payment/plans`。
- 管理端 Provider 配置组件已迁移到 upstream，支持 `supported_types`、`payment_mode`、退款开关、限额、Stripe / wxpay / alipay 配置提示。
- 部署文档已修正命令、章节、Provider 配置方式，并移除服务器密码。
- 2026-05-01 测试环境复核发现 `/purchase` 充值/订阅点击“确认支付”容易表现为无响应：后端创建订单接口正常，支付页主体与 upstream 保持一致；风险来自 fork 新增的全局 `FloatingContactButton` 在支付流程页底部浮动，可能遮挡或干扰支付确认区域。已将该浮层从 `/purchase` 和 `/payment/*` 排除，并补充组件测试。
- 2026-05-01 再次复核发现一次真实迁移漏项：`PaymentView.vue` 已引用 upstream 的 `btn-wxpay` / `btn-alipay` / `btn-stripe`，但 fork 的 `frontend/src/style.css` 没有同步这些全局按钮类，导致“确认支付”渲染成接近纯文本，用户感知为点击无响应。已按 upstream 补齐支付按钮样式，并补齐 `btn-outline-danger` / `btn-xs`。
- 已新增 `frontend/src/__tests__/buttonClasses.spec.ts`，扫描 `frontend/src` 下所有 `btn-*` 引用，要求在 `frontend/src/style.css` 中存在定义，避免后续再出现“迁移了模板但漏迁全局样式”的同类问题。
- 支付相关按钮已统一补充 `type="button"`，并在订单创建进入付款态后滚动到页面顶部二维码区域。`type="button"` 不是本次点击无响应的根因，但可降低后续表单重构引入隐性提交行为的风险。

## 2026-05-01 测试环境部署记录

- Commit：`2cafd049 fix(payment-b2): restore payment button styles`
- 测试库备份：`/home/nio/backups/sub2api_test_pre_payment_button_styles_20260501-135458.sql`，大小 58M。
- 测试环境：`sub2api-test`，镜像 `sub2api:test`，部署后容器 health 为 `healthy`，`http://127.0.0.1:8081/health` 返回 `{"status":"ok"}`。
- 生产环境未部署：`sub2api-prod` 仍运行在 `127.0.0.1:8080->8080/tcp`。
- 浏览器验收：
  - `/purchase` 充值页加载新构建产物 `PaymentView-CWE25KYc.js`。
  - 充值选择 10 元后，确认按钮 class 为 `btn w-full py-3 text-base font-medium btn-wxpay`，背景色 `rgb(43, 183, 65)`，未禁用；点击后进入微信二维码付款态，二维码 canvas 为 `220x220`。
  - 订阅页展示两个套餐：`codex 月`、`cc-月`；选择 `codex 月` 后确认按钮为同一 wxpay 样式，点击后进入微信二维码付款态，二维码 canvas 为 `220x220`。
  - 测试库生成订单：`id=33` balance 10.00 wxpay PENDING、有二维码；`id=34` subscription 15.80 wxpay PENDING、有二维码。

## 仍需测试环境验证

本地编译只能证明契约和类型正确，不能替代真实支付验收。合并前必须在测试环境完成：

1. 配置 wxpay Provider（优先按部署手册使用 Admin API 创建完整字段）。
2. 复核 wxpay Provider 的 `limits`：测试环境曾出现 `{"wxpay":{"singleMin":1,"singleMax":10000}}`，即使全局最小金额改为 `0.10`，前台仍会按实际 provider 最低额显示 1 元起付；需要改为 `{"wxpay":{"singleMin":0.1,"singleMax":10000}}` 或留空使用无 provider 限制。
3. 创建/编辑订阅套餐。注意 payment v2 读取 `subscription_plans`；旧 `payment_plans` 会由 `120b` migration 按 `group_id + name` 幂等补齐，`123` 会把绑定普通分组、停用分组或缺失分组的在售套餐下架。
4. 普通用户进入 `/purchase` 下单，确认侧边栏有“我的订单”入口 `/orders`。
5. 微信扫码支付。
6. 确认订单从 `PENDING` 流转到 `PAID` / `COMPLETED`。
7. 确认 `payment_audit_logs` 有对应记录。

## 关键文件

```text
backend/internal/server/routes/payment.go
backend/internal/server/routes/payment_routes_test.go
frontend/src/api/payment.ts
frontend/src/api/admin/payment.ts
frontend/src/api/__tests__/payment-contract.spec.ts
frontend/src/__tests__/buttonClasses.spec.ts
frontend/src/views/user/PaymentView.vue
frontend/src/views/user/UserOrdersView.vue
frontend/src/views/user/PaymentQRCodeView.vue
frontend/src/views/user/StripePaymentView.vue
frontend/src/components/payment/StripePaymentInline.vue
frontend/src/views/admin/orders/AdminOrdersView.vue
frontend/src/views/admin/orders/AdminPaymentDashboardView.vue
frontend/src/views/admin/orders/AdminPaymentPlansView.vue
frontend/src/views/admin/SettingsView.vue
frontend/src/components/FloatingContactButton.vue
docs/engineering/payment-b2-deploy.md
```

## 验证命令

```bash
cd frontend
pnpm exec vue-tsc --noEmit
pnpm build
pnpm vitest run src/api/__tests__/payment-contract.spec.ts src/components/payment/__tests__ src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/paymentUx.spec.ts src/views/user/__tests__/paymentWechatResume.spec.ts src/utils/__tests__/device.spec.ts

cd ../backend
GOCACHE="$PWD/../.cache/go-build" go test ./internal/payment ./internal/handler/admin ./internal/handler/dto ./internal/server/routes
GOCACHE="$PWD/../.cache/go-build" go test ./internal/service -run 'Test.*Payment|Test.*Wechat|Test.*WeChat|Test.*Provider|Test.*Order|Test.*Refund|Test.*Fulfillment|Test.*Config'
```

说明：`internal/service` 全包测试在本地沙箱可能被非 payment 的 `httptest.NewServer` 用例阻断；复核 payment 相关测试时使用更窄的 `-run` 范围。
