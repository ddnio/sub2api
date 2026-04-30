# Payment B-2 部署执行手册

**文档目的**：将 fork payment 模块迁移到 upstream payment v2 架构的部署步骤清单。  
**适用范围**：测试环境（test）和生产环境（prod），步骤一致，参数不同。  
**关联 PR**：[#18](https://github.com/ddnio/sub2api/pull/18)  
**最后更新**：2026-04-30

---

## 0. 部署前必读

1. **必须先部署测试环境并完成真实支付端到端验证**，再部署生产。
2. **必须先做 pg_dump 备份**。091a 会备份并重建旧 `payment_orders`，失败回滚依赖 pg_dump。
3. 微信支付回调 URL 由 provider config 的 `notifyUrl` 控制，不在微信商户平台单独配置。
4. 部署期间容器会重启并执行 migration，预期有短暂不可用窗口。
5. 文档不得保存服务器密码、支付密钥、token；凭据从团队密码库或运维交接渠道获取。

## 1. 环境信息

| 环境 | 主机 | 端口 | 域名 | DB |
|---|---|---|---|---|
| test | 108.160.133.141 | 8081 | router-test.nanafox.com | sub2api_test |
| prod | 108.160.133.141 | 8080 | router.nanafox.com | sub2api |

容器名（共享 PG 实例）：`sub2api-postgres`  
登录方式：`ssh nio@108.160.133.141`，密码或私钥从团队安全渠道获取。

## 2. 本地验证

在待部署 commit 上执行：

```bash
cd frontend
pnpm exec vue-tsc --noEmit
pnpm build

cd ../backend
GOCACHE="$PWD/../.cache/go-build" go test ./internal/payment ./internal/handler/admin ./internal/handler/dto ./internal/server/routes
GOCACHE="$PWD/../.cache/go-build" go test ./internal/service -run 'Test.*Payment|Test.*Wechat|Test.*WeChat|Test.*Provider|Test.*Order|Test.*Refund|Test.*Fulfillment|Test.*Config'
```

说明：
- 前端 build 产物由 Dockerfile 的 multi-stage build 生成并复制到镜像内的 `backend/internal/web/dist`，不需要把 dist 提交进 PR。
- `go test ./internal/service` 全包测试在本地沙箱中可能被非 payment 的 `httptest.NewServer` 用例阻断；至少必须单独验证 payment 相关用例和编译状态。

## 3. 服务器备份与 Preflight

```bash
ssh nio@108.160.133.141
cd /data/service/sub2api
```

测试环境备份：

```bash
TS=$(date +%Y%m%d-%H%M)
docker exec sub2api-postgres pg_dump -U sub2api sub2api_test \
  > /data/backups/sub2api_test_pre_b2_${TS}.sql
ls -lh /data/backups/sub2api_test_pre_b2_${TS}.sql
chmod 600 /data/backups/sub2api_test_pre_b2_${TS}.sql
```

生产环境将库名和文件名前缀替换为 `sub2api` / `sub2api_prod`。

Preflight SQL（全部必须为 0）：

```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
SELECT COUNT(*) AS bad_amount
FROM payment_orders
WHERE amount != ROUND(amount::numeric, 2);

SELECT COUNT(*) AS null_expired
FROM payment_orders
WHERE expired_at IS NULL;

SELECT COUNT(*) AS orphan_orders
FROM payment_orders po
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = po.user_id);

SELECT COUNT(*) AS fk_to_payment
FROM pg_constraint
WHERE confrelid = 'payment_orders'::regclass AND contype = 'f';
"
```

生产环境将 `sub2api_test` 替换为 `sub2api`。

## 4. 部署

```bash
cd /data/service/sub2api
git fetch origin
git checkout worktree-payment-b2  # 已 merge 后使用 main
git pull --ff-only
git log -1 --oneline
```

`deploy/deploy-server.sh` 内部还会执行一次 `git pull`。服务器工作树必须保持干净；建议在服务器仓库配置 `git config pull.ff only`，避免部署时产生 merge commit。

测试环境：

```bash
bash deploy/deploy-server.sh test
docker logs -f sub2api-test
```

生产环境：

```bash
bash deploy/deploy-server.sh prod
docker logs -f sub2api-prod
```

日志检查：
- 不能出现 `panic`、`ERROR`、`POSTCHECK FAILED`、`PREFLIGHT FAILED`。
- payment migration 应包含 `091a`、`092`、`092b`、`093`、`093a`、`094`、`095`、`095a`、`096`、`096a`、`098`、`099`、`100`、`101`、`102`、`102a`、`103`、`111`、`112`、`117`、`119`、`120`、`120a`。
- `113_normalize_legacy_wechat_provider_key.sql` 属于 upstream auth identity 迁移，不在本 PR 中引入。

## 5. 部署后验证

健康检查：

```bash
docker ps | grep sub2api-test
curl -s http://127.0.0.1:8081/health | head -5
```

数据完整性：

```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
SELECT COUNT(*) AS new_orders FROM payment_orders;
SELECT COUNT(*) AS backup_orders FROM payment_orders_v1_backup;

SELECT DISTINCT status FROM payment_orders ORDER BY status;

SELECT COUNT(*) AS null_expires
FROM payment_orders
WHERE expires_at IS NULL;

SELECT COUNT(*) AS empty_otn
FROM payment_orders
WHERE out_trade_no = '';

SELECT COUNT(*) FROM payment_audit_logs;
SELECT COUNT(*) FROM payment_provider_instances;
SELECT COUNT(*) FROM subscription_plans;

SELECT conname
FROM pg_constraint
WHERE conrelid='payment_provider_instances'::regclass;

SELECT indexname
FROM pg_indexes
WHERE tablename='payment_audit_logs';

SELECT last_value FROM payment_orders_id_seq;
SELECT MAX(id) FROM payment_orders;
"
```

期望：
- `new_orders` 与 `backup_orders` 相等。
- `status` 全部为大写状态。
- `null_expires = 0`。
- `payment_provider_instances` 首次部署后通常为 0，配置 Provider 后应增加。
- `payment_orders_id_seq.last_value >= MAX(payment_orders.id)`。

## 6. 配置 wxpay Provider

当前后台 UI 可以录入 `provider_key`、`name`、`enabled`、`config`，但不完整暴露 `supported_types`、`payment_mode`、`refund_enabled`、`allow_user_refund`。首次部署建议使用 Admin API 创建，避免字段缺失。

从服务器配置读取 wxpay 值（只读，不写入文档）：

```bash
sudo grep wxpay /etc/sub2api/test.yaml
```

准备 `provider-wxpay.json`：

```json
{
  "provider_key": "wxpay",
  "name": "wxpay-default",
  "supported_types": ["wxpay"],
  "enabled": true,
  "payment_mode": "qrcode",
  "refund_enabled": true,
  "allow_user_refund": false,
  "config": {
    "appId": "<wxpay_app_id>",
    "mchId": "<wxpay_mch_id>",
    "apiV3Key": "<wxpay_api_v3_key>",
    "privateKey": "<wxpay_private_key 完整 PEM>",
    "certSerial": "<wxpay_serial_no>",
    "publicKey": "<wxpay_public_key 完整 PEM>",
    "publicKeyId": "<wxpay_public_key_id>",
    "notifyUrl": "https://router-test.nanafox.com/api/v1/payment/webhook/wxpay"
  }
}
```

调用 Admin API：

```bash
ADMIN_TOKEN=<从浏览器或安全渠道获取的管理员 JWT>
curl -sS -X POST https://router-test.nanafox.com/api/v1/admin/payment/providers \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  --data @provider-wxpay.json
```

生产环境将域名和 `notifyUrl` 改为 `router.nanafox.com`。

验证：

```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
SELECT id, provider_key, name, supported_types, payment_mode, enabled, refund_enabled, allow_user_refund
FROM payment_provider_instances;
"
```

## 7. 端到端测试

1. 普通用户登录 `https://router-test.nanafox.com`。
2. 进入 `/payment`。
3. 选择最低价套餐或小额余额充值。
4. 选择微信支付并扫码完成支付。
5. 等待订单从 `PENDING` 流转到 `PAID` / `COMPLETED`。

数据库复核：

```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
SELECT id, user_id, amount, pay_amount, status, payment_type, payment_trade_no, paid_at, completed_at
FROM payment_orders
ORDER BY created_at DESC LIMIT 3;

SELECT order_id, action, operator, created_at
FROM payment_audit_logs
ORDER BY created_at DESC LIMIT 10;
"
```

## 8. 回滚预案

容器启动失败或 migration 报错：

```bash
docker stop sub2api-test
docker rm sub2api-test

TS=<备份时间戳>
cat /data/backups/sub2api_test_pre_b2_${TS}.sql | \
  docker exec -i sub2api-postgres psql -U sub2api -d sub2api_test

cd /data/service/sub2api
git checkout <previous-stable-commit>
bash deploy/deploy-server.sh test
```

如果容器启动成功但功能异常，旧代码不能直接读写新 schema；仍需按上面步骤恢复 DB。

## 9. 部署后清理

稳定运行至少 7 天后，且确认不再需要历史对照，再删除 backup 表：

```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
DROP TABLE IF EXISTS payment_orders_v1_backup;
"
```

## 10. 执行记录模板

每次部署把记录提交到 `docs/engineering/payment-b2-deploy-log.md`：

| 字段 | 值 |
|---|---|
| 环境 | test / prod |
| 部署时间 |  |
| 部署人 |  |
| 部署分支/commit |  |
| pg_dump 备份文件 |  |
| Preflight 结果 | bad_amount=0, null_expired=0, orphan_orders=0, fk_to_payment=0 |
| Migration 结果 |  |
| 数据完整性验证 |  |
| HTTP /health |  |
| Provider 配置结果 |  |
| 端到端测试结果 |  |
| 异常 / 备注 |  |

## 11. 故障信息

- PR：https://github.com/ddnio/sub2api/pull/18
- 备份目录：`/data/backups/sub2api_*_pre_b2_*.sql`
- 测试库：`docker exec sub2api-postgres psql -U sub2api -d sub2api_test`
- 生产库：`docker exec sub2api-postgres psql -U sub2api -d sub2api`
