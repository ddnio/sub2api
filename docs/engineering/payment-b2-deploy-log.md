# Payment B-2 部署记录

## 2026-05-02 生产环境：部署前只读复核

| 字段 | 值 |
|---|---|
| 环境 | prod |
| 复核时间 | 2026-05-02 09:58-10:05 Asia/Shanghai |
| 复核人 | Codex |
| 当前服务器工作树 | `/data/service/sub2api` 已在 `worktree-payment-b2` / `d36a1b18`，这是测试部署拉取后的工作树状态，不代表生产容器已经运行该 commit |
| 当前生产容器 | `sub2api-prod` healthy，镜像 `sub2api:prod`，启动时间约 2026-05-01 09:18 Asia/Shanghai，静态资源为 `/assets/index-7w3Ly_Qz.js`、`/assets/index-0_l6BwS5.css` |
| 当前测试容器 | `sub2api-test` healthy，已部署最终测试 commit `d36a1b18`，静态资源为 `/assets/index-DmPQiWlR.js`、`/assets/index-sWXccaJK.css` |
| 生产代码判断 | 生产前端资源仍不是最终测试资源；生产尚未部署 `d36a1b18` 最终版 |
| 生产 DB migration 状态 | `schema_migrations` 已包含 payment migrations 至 `120b_backfill_subscription_plans_from_payment_plans.sql`，应用时间为 2026-05-01 09:01 Asia/Shanghai |
| 现有生产备份 | `/home/nio/backups/sub2api_prod_pre_b2_20260501-0101.sql`，约 996M；这是上一次 B-2 生产操作前备份，正式部署最终版前仍必须重新备份 |
| 生产只读 Preflight | `bad_amount=0`、`null_expired=0`、`orphan_orders=0`、`invalid_payment_order_index=0`、`null_expires=0`、`empty_otn=0`、`duplicate_out_trade_no=0` |
| 生产数据摘要 | `payment_orders=57`、`payment_orders_v1_backup=56`、`payment_orders_id_seq=57`、`MAX(payment_orders.id)=57` |
| Provider 配置 | `wxpay-default` enabled，`payment_mode=qrcode`，`refund_enabled=true`，`allow_user_refund=false`，`limits={"wxpay":{"singleMin":1,"singleMax":10000}}` |
| 全局充值设置 | `MIN_RECHARGE_AMOUNT=1.00`、`MAX_RECHARGE_AMOUNT=10000.00`；当前生产维持 1 元起付，和 provider `singleMin=1` 一致 |
| 套餐验证 | `subscription_plans=4`，4 个在售套餐：`Codex专用`、`ClaudeCode-Pro`、`ClaudeCode-Lite`、`cc-kimi`；`invalid_for_sale_plans=0` |
| 生产日志 | `docker logs --tail 240 sub2api-prod` 未发现 panic/fatal/preflight/postcheck/migration 失败；grep 命中 `/error-distribution` 只是正常 admin API 路径，HTTP 200 |
| 结论 | 可以进入生产最终版部署准备；正式执行前必须重新 pg_dump 生产库、确认备份大小和权限、确认 wxpay 生产 notifyUrl 与起付金额策略 |

## 2026-05-02 测试环境：final provider review gaps 修复

| 字段 | 值 |
|---|---|
| 环境 | test |
| 部署时间 | 2026-05-02 01:39-01:45 Asia/Shanghai |
| 部署分支/commit | `worktree-payment-b2` / `d36a1b18` |
| pg_dump 备份文件 | `/home/nio/backups/sub2api_test_pre_payment_b2_final_review_20260502-013941.sql`，57M，权限 `600` |
| 变更范围 | Kimi final follow-up：补齐 Stripe provider 单测；补齐 wxpay `QueryOrder`、`VerifyNotification`、`Refund`、`CancelPayment` provider 单测；修复微信内浏览器无 OpenID 时误走 JSAPI/OAuth 的 wxpay 拦截 |
| 本地验证 | `go test -count=1 ./...` passed；`go test -tags unit -count=1 ./internal/payment/provider ...` passed；service JSAPI/OAuth targeted tests passed；`git diff --check` passed |
| 部署命令 | `bash deploy/deploy-server.sh test` |
| HTTP /health | test/prod 均返回 `{"status":"ok"}` |
| 容器状态 | `sub2api-test` healthy，`127.0.0.1:8081->8080/tcp`；`sub2api-prod` healthy，`127.0.0.1:8080->8080/tcp` |
| Postcheck | `bad_amount=0`、`null_expired=0`、`orphan_orders=0`、`invalid_payment_order_index=0`、`null_expires=0`、`empty_otn=0`、`duplicate_out_trade_no=0` |
| Provider 配置结果 | `wxpay-default` enabled，`limits={"wxpay":{"singleMax":10000,"singleMin":0.1}}` |
| 套餐验证 | `subscription_plans=3`，`invalid_for_sale_plans=0` |
| 静态前端验证 | `/purchase`、`/orders` HTTP 200；入口资源 `/assets/index-DmPQiWlR.js`、`/assets/index-sWXccaJK.css` |
| 端到端测试结果 | 本次未重新扫码；前一轮同一测试环境真实微信支付订单 `35` balance 0.10 和订单 `36` subscription 0.10 均已 `COMPLETED`，本次改动为 provider gating/test coverage，不改前端支付点击链路 |
| 生产影响 | 未部署生产；仅确认 `sub2api-prod` health 正常 |
| 异常 / 备注 | 浏览器自动化会话卡顿，最终验证改用 health、静态 route、SQL postcheck、日志扫描；生产前建议人工快速打开关键页面复核视觉 |

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
