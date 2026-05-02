# Payment B-2 前端与部署复核交接

**日期**：2026-05-02
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
- Kimi final follow-up 发现微信内浏览器在无 OpenID 时会被误导入 JSAPI/OAuth，而本 fork 不支持 JSAPI。已收窄判断：只有真实 OpenID/JSAPI 请求才要求 JSAPI-compatible provider；普通微信内 H5/native wxpay 不再被 `WECHAT_PAYMENT_MP_NOT_CONFIGURED` 拦截。
- 已补齐 Stripe provider 单测，以及 wxpay provider `QueryOrder`、`VerifyNotification`、`Refund`、`CancelPayment` 单测。

## 2026-05-02 测试环境最终部署记录

- Commit：`d36a1b18 fix(payment-b2): close final provider review gaps`
- 测试库备份：`/home/nio/backups/sub2api_test_pre_payment_b2_final_review_20260502-013941.sql`，大小 57M，权限 `600`。
- 测试环境：`sub2api-test`，镜像 `sub2api:test`，部署后容器 health 为 `healthy`，`http://127.0.0.1:8081/health` 返回 `{"status":"ok"}`。
- 生产环境未部署：`sub2api-prod` 仍运行在 `127.0.0.1:8080->8080/tcp`。
- 静态资源验收：
  - `/purchase` 和 `/orders` 返回 HTTP 200。
  - 当前构建入口为 `/assets/index-DmPQiWlR.js`，当前 CSS 为 `/assets/index-sWXccaJK.css`。
  - CSS 包含 `btn-wxpay`、`btn-alipay`、`btn-stripe`、`btn-outline-danger`、`btn-xs`，覆盖此前按钮样式漏迁问题。
- 数据验收：
  - Provider：`wxpay-default` enabled，`limits={"wxpay":{"singleMax":10000,"singleMin":0.1}}`。
  - `subscription_plans=3`，`invalid_for_sale_plans=0`。
  - `bad_amount=0`、`null_expired=0`、`orphan_orders=0`、`invalid_payment_order_index=0`、`null_expires=0`、`empty_otn=0`、`duplicate_out_trade_no=0`。
  - 测试库此前真实支付完成记录仍有效：`id=35` 为 balance 0.10 wxpay `COMPLETED`，`id=36` 为 subscription 0.10 wxpay `COMPLETED`；两笔均有 QR、`paid_at`、`completed_at`。
  - 审计日志包含 `ORDER_CREATED`、`ORDER_PAID`、`RECHARGE_SUCCESS`、`SUBSCRIPTION_SUCCESS`。
- 备注：浏览器自动化工具不稳定，最终改用 health、静态路由、SQL postcheck、日志扫描和真实 wxpay 订单证据验证。生产前仍建议人工快速打开 `/purchase`、`/orders`、`/admin/payment/orders` 复核视觉。

## 仍需生产前确认

1. 生产部署前重新执行 pg_dump，并确认备份文件大小正常、权限为 `600`。
2. 生产 Provider 的 `notifyUrl` 必须使用 `https://router.nanafox.com/api/v1/payment/webhook/wxpay`，不要沿用测试域名。
3. 生产 Provider `limits` 需要与后台全局最低充值金额一致；如果允许 0.1 元支付，`wxpay.singleMin` 不能高于 `0.1`。
4. 生产部署前再次确认用户开放套餐均绑定 active subscription 分组，避免 payment v2 下单时拒绝无效分组套餐。
5. 如果生产启用 Alipay 或 Stripe，需要先在测试环境分别完成真实或 sandbox 端到端验证；本轮真实支付只覆盖 wxpay。
6. 如果生产启用退款或用户自助退款，需要先在测试环境完成真实退款验证；本轮仅有退款单元测试覆盖。
7. 浏览器自动化工具本轮不稳定，最终 UI 视觉只做了静态资源和真实订单链路验证；生产前建议人工快速打开 `/purchase`、`/orders`、`/admin/payment/orders` 复核页面渲染。

## 关键文件

```text
backend/internal/server/routes/payment.go
backend/internal/server/routes/payment_routes_test.go
backend/internal/payment/provider/stripe_test.go
backend/internal/payment/provider/wxpay.go
backend/internal/payment/provider/wxpay_test.go
backend/internal/service/payment_order.go
backend/internal/service/payment_order_result_test.go
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
GOCACHE="$PWD/../.cache/go-build" go test ./internal/payment ./internal/payment/provider ./internal/handler/admin ./internal/handler/dto ./internal/server/routes
GOCACHE="$PWD/../.cache/go-build" go test ./internal/service -run 'Test.*Payment|Test.*Wechat|Test.*WeChat|Test.*Provider|Test.*Order|Test.*Refund|Test.*Fulfillment|Test.*Config'
```

说明：`internal/service` 全包测试在本地沙箱可能被非 payment 的 `httptest.NewServer` 用例阻断；复核 payment 相关测试时使用更窄的 `-run` 范围。
