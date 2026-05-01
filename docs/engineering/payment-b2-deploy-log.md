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
