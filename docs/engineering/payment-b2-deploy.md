# Payment B-2 部署执行手册

**文档目的**：将 fork payment 模块迁移到 upstream payment v2 架构的部署步骤清单。  
**适用范围**：测试环境（test）和生产环境（prod），步骤一致，参数不同。  
**关联 PR**：[#18](https://github.com/ddnio/sub2api/pull/18)  
**最后更新**：2026-04-30

---

## ⚠️ 部署前必读

1. **此次部署不可回滚到代码层面**——一旦 091a 执行后旧表已被 DROP，只能从 pg_dump 恢复
2. **回调 URL 不在微信商户平台配置**——由代码生成，存在 `payment_provider_instances.config.notifyUrl` 字段
3. **部署期间服务约 30 秒不可用**（migration 执行 091a→092→092b→...→120a）
4. **生产部署前必须先用测试环境验证一遍**

---

## 0. 环境信息

| 环境 | 主机 | 端口 | 域名 | DB |
|---|---|---|---|---|
| test | 108.160.133.141 | 8081 | router-test.nanafox.com | sub2api_test |
| prod | 108.160.133.141 | 8080 | router.nanafox.com | sub2api |

容器名（共享 PG 实例）：`sub2api-postgres`  
登录：`ssh nio@108.160.133.141`（密码：nio2026.）

---

## 1. 部署前置准备（本地）

```bash
# 确认 PR 已合并到 main（或确定要部署的分支）
gh pr view 18 --repo ddnio/sub2api
```

**当前部署使用分支**：`worktree-payment-b2`（未 merge 时）或 `main`（已 merge 后）

---

## 2. 服务器端：备份 + Preflight

```bash
ssh nio@108.160.133.141
cd /data/service/sub2api
```

### 2.1 pg_dump 备份（必须，不可省）

**测试环境**：
```bash
TS=$(date +%Y%m%d-%H%M)
docker exec sub2api-postgres pg_dump -U sub2api sub2api_test \
  > /data/backups/sub2api_test_pre_b2_${TS}.sql
ls -lh /data/backups/sub2api_test_pre_b2_${TS}.sql
# 必须非空，至少几 MB
chmod 600 /data/backups/sub2api_test_pre_b2_${TS}.sql
```

**生产环境**（部署 prod 时执行）：
```bash
TS=$(date +%Y%m%d-%H%M)
docker exec sub2api-postgres pg_dump -U sub2api sub2api \
  > /data/backups/sub2api_prod_pre_b2_${TS}.sql
ls -lh /data/backups/sub2api_prod_pre_b2_${TS}.sql
chmod 600 /data/backups/sub2api_prod_pre_b2_${TS}.sql
```

> 备份文件包含支付密钥（如果之前已迁移到 DB），chmod 600 限制访问。

### 2.2 Preflight SQL 检查（4 条，全部必须返回 0）

**测试环境**：
```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
-- 1. 金额精度：所有订单 amount 必须 ≤ 2 位小数
SELECT COUNT(*) AS bad_amount FROM payment_orders WHERE amount != ROUND(amount::numeric, 2);

-- 2. expired_at 不能为 NULL（092 新表 expires_at NOT NULL 依赖此字段）
SELECT COUNT(*) AS null_expired FROM payment_orders WHERE expired_at IS NULL;

-- 3. 没有孤儿订单（user_id 必须在 users 表中存在）
SELECT COUNT(*) AS orphan_orders FROM payment_orders po
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = po.user_id);

-- 4. 没有其他表的 FK 指向 payment_orders（DROP 安全性）
SELECT COUNT(*) AS fk_to_payment FROM pg_constraint
WHERE confrelid = 'payment_orders'::regclass AND contype = 'f';
"
```

**生产环境**：将 `sub2api_test` 替换为 `sub2api`。

> ⚠️ 若任一检查 > 0，**停止部署**，先解决数据问题。

---

## 3. 切换分支并部署

### 3.1 切换分支
```bash
cd /data/service/sub2api
git fetch origin
git checkout worktree-payment-b2  # 或 main（若已 merge）
git pull
git log -1 --oneline   # 确认是最新 commit
```

### 3.2 执行部署脚本

**测试环境**：
```bash
bash deploy/deploy-server.sh test
```

**生产环境**：
```bash
bash deploy/deploy-server.sh prod
```

> 脚本自动：`git pull` → `docker build` → 重启容器  
> 容器启动时 migration 自动执行（091a → 092 → 092b → 093 → 093a → ... → 120a → 102a）

### 3.3 监控 migration 日志

```bash
# 测试环境
docker logs -f sub2api-test 2>&1 | tail -100

# 生产环境
docker logs -f sub2api-prod 2>&1 | tail -100
```

**关键观察**：
- 应看到 25 个 payment migration 文件按顺序执行
- 不能有 `panic`、`ERROR`、`POSTCHECK FAILED`、`PREFLIGHT FAILED`
- 容器启动成功 = 所有 migration 通过

**若日志中有错误，立即跳到 §6 回滚**。

---

## 4. 部署后验证（必做）

### 4.1 容器健康检查
```bash
# 测试
docker ps | grep sub2api-test
curl -s http://127.0.0.1:8081/health | head -5

# 生产
docker ps | grep sub2api-prod
curl -s http://127.0.0.1:8080/health | head -5
```

### 4.2 数据完整性验证

**测试环境**：
```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
-- 1. payment_orders 行数应等于 backup（迁移前的 56 条）
SELECT COUNT(*) AS new_orders FROM payment_orders;
SELECT COUNT(*) AS backup_orders FROM payment_orders_v1_backup;
-- 这两个数字必须相等

-- 2. status 全部大写（092b 已转换）
SELECT DISTINCT status FROM payment_orders;
-- 期望：PENDING / PAID / COMPLETED / EXPIRED / REFUNDED 等大写值

-- 3. expires_at 全部非空
SELECT COUNT(*) AS null_expires FROM payment_orders WHERE expires_at IS NULL;
-- 期望：0

-- 4. out_trade_no 已回填（102a 完成）
SELECT COUNT(*) AS empty_otn FROM payment_orders WHERE out_trade_no = '';
-- 应等于历史中 order_no 为空的订单数（应为 0 或很少）

-- 5. 新表已创建
SELECT COUNT(*) FROM payment_audit_logs;          -- 表存在即可
SELECT COUNT(*) FROM payment_provider_instances;  -- 应为 0（待手动配置）
SELECT COUNT(*) FROM subscription_plans;          -- 应为 0 或继承 payment_plans 数据

-- 6. 唯一约束已建
SELECT conname FROM pg_constraint WHERE conrelid='payment_provider_instances'::regclass;
-- 期望：包含 uq_payment_provider_instances_provider_name

SELECT indexname FROM pg_indexes WHERE tablename='payment_audit_logs';
-- 期望：包含 uq_payment_audit_logs_affiliate_claim
"
```

**生产环境**：将 `sub2api_test` 替换为 `sub2api`。

### 4.3 ID 序列验证
```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
SELECT last_value FROM payment_orders_id_seq;
SELECT MAX(id) FROM payment_orders;
"
# last_value 应 ≥ MAX(id)，下一次 INSERT 不会冲突
```

---

## 5. 配置 wxpay Provider（首次部署必做）

### 5.1 准备配置数据

从服务器 yaml 文件读取（**仅查看，不修改**）：
```bash
# 测试环境
sudo cat /etc/sub2api/test.yaml | grep wxpay

# 生产环境
sudo cat /etc/sub2api/prod.yaml | grep wxpay
```

记录以下字段值：
- `wxpay_app_id`
- `wxpay_mch_id`
- `wxpay_api_v3_key`
- `wxpay_private_key`（PEM 格式，多行）
- `wxpay_serial_no`
- `wxpay_public_key`（PEM 格式，多行）
- `wxpay_public_key_id`

### 5.2 在 admin 后台录入

打开管理后台：
- 测试：https://router-test.nanafox.com/admin
- 生产：https://router.nanafox.com/admin

进入 **支付管理 → Provider 配置 → 新增**，填入：

| 字段 | 值 | 说明 |
|---|---|---|
| provider_key | `wxpay` | 固定值 |
| name | `wxpay-default` | 实例标识，自定义 |
| supported_types | `wxpay` | 支持的支付类型 |
| enabled | true | 启用 |
| refund_enabled | true | 允许退款 |
| **config（JSON）** | 见下 | 加密存储 |

config JSON 内容：
```json
{
  "appId": "<wxpay_app_id>",
  "mchId": "<wxpay_mch_id>",
  "apiV3Key": "<wxpay_api_v3_key>",
  "privateKey": "<wxpay_private_key 完整 PEM>",
  "certSerial": "<wxpay_serial_no>",
  "publicKey": "<wxpay_public_key 完整 PEM>",
  "publicKeyId": "<wxpay_public_key_id>",
  "notifyUrl": "https://router-test.nanafox.com/api/v1/payment/webhook/wxpay"
}
```

> ⚠️ **`notifyUrl` 必填**：测试环境用 `router-test`，生产环境用 `router`。  
> ⚠️ 微信商户平台**不需要任何配置**，回调 URL 由代码下单时传给微信。

### 5.3 验证 provider 配置成功
```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
SELECT id, provider_key, name, enabled, refund_enabled FROM payment_provider_instances;
"
# 应有 1 行 wxpay-default
```

---

## 6. 端到端功能测试（测试环境必做）

### 6.1 创建一笔小额测试订单
1. 用普通用户账号登录前台
2. 访问购买页面，选择最低价套餐或最低充值金额（建议 ¥0.01）
3. 选择微信支付
4. 用扫码完成支付
5. 等待 30 秒后查看订单状态

### 6.2 验证订单流转
```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
SELECT id, user_id, amount, status, payment_type, payment_trade_no, paid_at, completed_at
FROM payment_orders
ORDER BY created_at DESC LIMIT 3;
"
```

**期望流转**：`PENDING` → `PAID` → `COMPLETED`

### 6.3 验证 audit log
```bash
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
SELECT order_id, action, operator, created_at FROM payment_audit_logs
ORDER BY created_at DESC LIMIT 10;
"
# 应能看到 ORDER_PAID, ORDER_RECOVERED 等 action
```

---

## 6. 回滚预案（如部署失败）

### 6.1 容器启动失败（migration 报错）

```bash
# 在服务器执行
docker stop sub2api-test  # 或 sub2api-prod
docker rm sub2api-test

# 从 pg_dump 恢复
TS=<之前记录的备份时间戳>
cat /data/backups/sub2api_test_pre_b2_${TS}.sql | \
  docker exec -i sub2api-postgres psql -U sub2api -d sub2api_test

# 切回旧分支并重新部署
cd /data/service/sub2api
git checkout main  # 或上一个稳定 commit
bash deploy/deploy-server.sh test
```

### 6.2 容器启动成功但功能异常

```bash
# 切回旧分支
cd /data/service/sub2api
git checkout <previous-stable-commit>
bash deploy/deploy-server.sh test
```

> ⚠️ 代码回滚后 DB 已是新 schema，旧代码无法读写新字段 → 必须同步做 §6.1 DB 回滚。

### 6.3 数据问题（已部署但数据异常）

如果 `payment_orders_v1_backup` 表还在（部署后 1 周内不应删除）：
```bash
# 临时方案：从 backup 表查历史数据
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
SELECT * FROM payment_orders_v1_backup WHERE id = <order_id>;
"
```

完整回滚 → §6.1 从 pg_dump 恢复。

---

## 7. 部署后清理（部署 1 周后，且确认稳定）

```bash
# 删除 backup 表（仅在生产稳定运行 1 周以上才执行）
docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c "
DROP TABLE IF EXISTS payment_orders_v1_backup;
"
```

> ⚠️ 不要急着删！只要 backup 表存在，就能用 §6.3 快速对照历史数据。

---

## 8. 部署执行记录表

每次部署填写并提交本表到 git（`docs/engineering/payment-b2-deploy-log.md`）：

### 测试环境部署记录
| 字段 | 值 |
|---|---|
| 部署时间 | 待填 |
| 部署人 | 待填 |
| 部署分支/commit | 待填 |
| pg_dump 备份文件 | 待填 |
| Preflight 检查结果 | 全部 0 ✓ |
| Migration 日志关键节点 | 待填 |
| 数据完整性验证 | 待填（new_orders 数 / backup 数） |
| Provider 配置时间 | 待填 |
| 端到端测试结果 | 待填（测试订单 ID） |
| 异常 / 备注 | 待填 |

### 生产环境部署记录
（同上结构）

---

## 9. 已知风险与注意事项

| 项 | 描述 | 缓解 |
|---|---|---|
| Migration 30s 停服 | 091a→092→092b→...→120a 串行执行约 30s | 选低峰部署 |
| Provider 配置后才能下单 | 部署后到完成 §5 之间用户无法支付 | 部署完立即配置（< 5 min） |
| 容器 healthcheck 期间 503 | 启动期 `/health` 返回 503 | docker-compose `restart: unless-stopped` 自动恢复 |
| backup 表占用空间 | 暂不删除（用于回滚） | 7 天后清理（§7） |

---

## 10. 故障联系

- PR：https://github.com/ddnio/sub2api/pull/18
- 备份目录：`/data/backups/sub2api_*_pre_b2_*.sql`
- 服务器：`ssh nio@108.160.133.141`
- 测试库：`docker exec sub2api-postgres psql -U sub2api -d sub2api_test`
- 生产库：`docker exec sub2api-postgres psql -U sub2api -d sub2api`
