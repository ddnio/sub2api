# Payment B-2 部署记录

## 2026-05-01 测试环境：支付页浮层修复

| 字段 | 值 |
|---|---|
| 环境 | test |
| 部署时间 | 2026-05-01 18:57-19:00 Asia/Shanghai |
| 部署分支/commit | `worktree-payment-b2` / `50824a2b` |
| pg_dump 备份文件 | `/home/nio/backups/sub2api_test_pre_contact_payment_flow_20260501-105700.sql` |
| 变更范围 | fork 新增的 `FloatingContactButton` 在 `/purchase`、`/payment/*` 不渲染；支付页主体继续保留 upstream 实现 |
| 本地验证 | `pnpm exec vitest run ...` 40 tests passed；`pnpm exec vue-tsc --noEmit` passed；`pnpm build` passed |
| 部署命令 | `bash deploy/deploy-server.sh test` |
| HTTP /health | `{"status":"ok"}` |
| 容器状态 | `sub2api-test` healthy，`127.0.0.1:8081->8080/tcp` |
| 生产影响 | 未操作生产库；`sub2api-prod` 仍 healthy，`127.0.0.1:8080->8080/tcp` |
| 端到端测试结果 | `/purchase` 不再显示“联系我们”浮层；余额充值微信支付创建订单 `28`；订阅套餐微信支付创建订单 `29` |
| 数据库验证 | 订单 `28` 为 `balance/PENDING/wxpay/10.00`；订单 `29` 为 `subscription/PENDING/wxpay/0.10` |
| 异常 / 备注 | 订单停留 `PENDING` 是未扫码支付的预期状态 |

## 2026-05-01 测试环境：upstream audit gaps 修复

| 字段 | 值 |
|---|---|
| 环境 | test |
| 部署时间 | 2026-05-01 23:31 Asia/Shanghai |
| 部署分支/commit | `worktree-payment-b2` / `05593a1b` |
| pg_dump 备份文件 | `/home/nio/backups/sub2api_test_pre_payment_b2_audit_20260501-153107.sql`，59M |
| 变更范围 | upstream payment audit gaps：provider config 明文写入 + legacy AES 读取兜底、API error `reason/metadata` 透传、移除旧 purchase subscription 前端设置 |
| 本地验证 | `go test -count=1 ./...` passed after Stripe test addition；`go test -count=1 ./internal/payment/provider` passed；payment/service targeted tests passed；unit-tag webhook/result/Stripe tests passed；frontend targeted vitest 12 files / 99 tests passed；`vue-tsc --noEmit` passed；`pnpm build` passed；`git diff --check` passed |
| 部署命令 | `bash deploy/deploy-server.sh test` |
| HTTP /health | test/prod 均返回 `{"status":"ok"}` |
| 容器状态 | `sub2api-test` healthy，`127.0.0.1:8081->8080/tcp`；`sub2api-prod` healthy，`127.0.0.1:8080->8080/tcp` |
| Preflight / Postcheck | `bad_amount=0`、`null_expired=0`、`orphan_orders=0`、`fk_to_payment=0`、`invalid_payment_order_index=0`、`duplicate_out_trade_no=0` |
| Provider 配置结果 | `wxpay-default` enabled，`limits={"wxpay":{"singleMax":10000,"singleMin":0.1}}` |
| 数据完整性验证 | `payment_orders=36`、`payment_orders_v1_backup=18`；差值来自部署后测试订单；`null_expires=0`、`empty_otn=0`、`paymentorder_out_trade_no` valid/ready；`payment_orders_id_seq=36`、`MAX(id)=36` |
| 套餐验证 | `subscription_plans=3`，其中在售 2 个；`invalid_for_sale_plans=0`；旧 `payment_plans` 未迁移残留 0 个 |
| 静态前端验证 | `/purchase`、`/orders` HTTP 200；入口 chunk 引用 `PaymentView-DEvP5jRX.js`、`UserOrdersView-CvrL0lfQ.js`、`PaymentResultView-DeW8TWLC.js`；CSS 包含 `btn-wxpay`、`btn-alipay`、`btn-stripe`、`btn-outline-danger`、`btn-xs` |
| 端到端测试结果 | 真实微信支付订单 `35` 为 balance 0.10 wxpay `COMPLETED`；订单 `36` 为 subscription 0.10 wxpay `COMPLETED`；两笔均有 QR、`paid_at`、`completed_at` |
| 审计日志 | 订单 `35` 有 `ORDER_CREATED`、`ORDER_PAID`、`AFFILIATE_REBATE_SKIPPED`、`RECHARGE_SUCCESS`；订单 `36` 有 `ORDER_CREATED`、`ORDER_PAID`、`SUBSCRIPTION_SUCCESS` |
| 生产影响 | 未部署生产；仅确认 `sub2api-prod` health 正常 |
| Kimi final review | 无 P0；Stripe provider 单测缺口已补；微信内浏览器无 OpenID 的 JSAPI/OAuth 误拦截已修；wxpay QueryOrder/VerifyNotification/Refund/CancelPayment 单测已补；Alipay/Stripe 真实链路和真实退款记录为启用前 gate |
| 异常 / 备注 | 浏览器自动化会话本轮不稳定，改用静态资源 + SQL + 真实订单链路验证；生产前建议人工快速打开关键页面复核视觉 |
