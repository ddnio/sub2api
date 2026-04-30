# Payment B-2 交接文档（前端改造阶段）

**日期**：2026-04-30  
**当前状态**：Backend 已完成并验证（测试环境已回滚保留），Frontend 待实施  
**Worktree**：`.claude/worktrees/payment-b2/`  
**PR**：https://github.com/ddnio/sub2api/pull/18 （13 commits，未 merge）

---

## 1. 当前进度（已完成）

### Backend 完整迁移到 upstream payment v2 ✅
- **Migration**：4 个 fork patch（091a/092b/093a/095a/096a/102a）+ 18 个 upstream payment migration（092~120a，113 已剔除是 auth 模块）
- **代码**：`internal/payment/`、`internal/payment/provider/`、`internal/service/payment_*.go` 全部从 upstream 复制
- **删除**：fork 自研 12 个 payment 文件
- **ent schema**：payment_order/audit_log/provider_instance/subscription_plan
- **wire_gen.go**：手工更新为新 DI 结构
- **routes**：admin `/payment/providers`+`/payment/config`，user webhook 路径变 `/payment/webhook/wxpay`
- **fork patches**：
  - `getWeChatPaymentOAuthCredential` → JSAPI not supported stub（fork 用 Native Pay）
  - `AffiliateService` → 最小 stub（nil-guarded）
  - `WeChatConnectConfig` → fork config 追加
  - `normalizeVisibleMethodSettingSource` → 内联

### 测试环境部署验证 ✅（然后回滚）
- 备份：`/data/backups/sub2api_test_pre_b2_20260430-1054.sql`（57MB）
- migration 全部成功（27 个）
- 18 条历史订单完整迁移（id 保留、状态大写、字段映射正确）
- HTTP /health 200
- **回滚原因**：发现前端没改造，UI 出现：
  - 订单管理显示 `payment.orderType.undefined`（i18n key 未补齐）
  - 系统设置无 Provider 配置入口（admin 旧 SettingsView）
  - 充值/订阅是旧 iframe 版（user 旧 PurchaseSubscriptionView）
- **当前状态**：测试环境已恢复旧版（main 分支 + DB 从 pg_dump 恢复）

---

## 2. 待办（Frontend 改造，1-2 天）

### 2.1 引入 upstream 前端组件
**目标位置**：`frontend/src/`

```bash
# 在 worktree 里执行
cd /Users/nio/project/nanafox/sub2api/.claude/worktrees/payment-b2

# upstream 新增组件清单（已 grep 确认 fork 无）
git ls-tree -r upstream/main frontend/src/components/admin/payment/ | awk '{print $4}'
git ls-tree -r upstream/main frontend/src/components/payment/ | awk '{print $4}'
```

需复制：
- `components/admin/payment/` 7 个文件（AdminOrderDetail/Table/RefundDialog + 4 个 chart）
- `components/payment/` 13 个文件（含 SubscriptionPlanCard、PaymentMethodSelector、PaymentQRDialog 等）

```bash
# 批量复制（参考 backend migration 的复制脚本）
mkdir -p frontend/src/components/admin/payment frontend/src/components/payment
for f in $(git ls-tree -r upstream/main frontend/src/components/admin/payment/ | awk '{print $4}'); do
  git show "upstream/main:$f" > "$f"
done
for f in $(git ls-tree -r upstream/main frontend/src/components/payment/ | awk '{print $4}'); do
  git show "upstream/main:$f" > "$f"
done
```

### 2.2 重写主要页面

| 文件 | 操作 | 上游版本 |
|---|---|---|
| `frontend/src/views/admin/PaymentOrdersView.vue` | 重写：用新组件 | `git show upstream/main:frontend/src/views/admin/PaymentOrdersView.vue` |
| `frontend/src/views/admin/PaymentPlansView.vue` | 重命名为 `SubscriptionPlansView.vue` 并重写 | 上游叫 `SubscriptionsView.vue` |
| `frontend/src/views/admin/SettingsView.vue` | **关键**：加入 Provider 配置 + Payment Config 区块（参考上游） | 上游有 |
| `frontend/src/views/user/PurchaseSubscriptionView.vue` | 重写：用 SubscriptionPlanCard + AmountInput | 上游 `SubscribeView.vue` |
| `frontend/src/views/user/PaymentView.vue` | 重写：PaymentMethodSelector + PaymentQRDialog + PaymentStatusPanel | 上游有 |

### 2.3 i18n 补齐（关键问题）

测试环境看到的 `payment.orderType.undefined` 来自：
- `frontend/src/locales/zh-CN.json`
- `frontend/src/locales/en-US.json`

需要补齐的 key（从上游 locales 文件直接拿）：
```bash
# 看上游 i18n 中 payment 部分
git show "upstream/main:frontend/src/locales/zh-CN.json" | python3 -c "
import json, sys
d = json.load(sys.stdin)
print(json.dumps(d.get('payment', {}), ensure_ascii=False, indent=2))
"
```

需要的 key 至少包括：
- `payment.orderType.balance`、`payment.orderType.subscription`
- `payment.orderStatus.PENDING`、`PAID`、`COMPLETED`、`EXPIRED`、`FAILED`、`REFUNDED` 等

### 2.4 router + sidebar
- `frontend/src/router/index.ts`：添加新页面路由（如 admin payment provider 配置页）
- `frontend/src/components/layout/AppSidebar.vue`：
  - 旧：`purchase_subscription_enabled` → 隐藏/显示充值入口
  - 新：使用 `custom_menu_items`（来自 098 migration）

### 2.5 stores 适配新 API
- `frontend/src/stores/subscriptions.ts`：可能需要改字段名（plan_id → 仍然 plan_id，但其他字段如 expired_at → expires_at 等）
- `frontend/src/api/payment.ts`：检查响应类型是否匹配新 backend

### 2.6 验证
```bash
cd /Users/nio/project/nanafox/sub2api/.claude/worktrees/payment-b2
pnpm --dir frontend build  # 必须 0 错误
pnpm --dir frontend dev    # 本地启动看 UI
```

---

## 3. 重要技术约束

### 3.1 Backend 是新版（不要回退）
- DB schema 已是 v2（order_no→out_trade_no, type→order_type, expired_at→expires_at, status 大写）
- API 端点：
  - 用户：`GET /payment/plans`、`POST /payment/orders`、`GET /payment/orders`、`GET /payment/orders/:id`
  - 管理员：`GET /admin/payment/orders`、`/admin/payment/providers`、`/admin/payment/config`
  - Webhook：`POST /payment/webhook/wxpay`（旧路径 `/callback/:provider` 已不存在）
- 前端必须按新 API 字段名调用

### 3.2 测试环境状态
- main 分支已部署，DB 是旧 schema（已从 pg_dump 恢复）
- 完整 PR worktree-payment-b2 13 commits 在 GitHub，**不要 merge 到 main**
- 前端工作完成后追加 commit 到同一分支，再合并

### 3.3 配置 Provider 的方式
- **不是**通过 yaml 文件
- **不是**在微信商户平台
- 是通过 admin 后台 UI（页面待前端实现）写入 `payment_provider_instances` 表
- 关键字段：`config` JSON 含 `notifyUrl`（值如 `https://router-test.nanafox.com/api/v1/payment/webhook/wxpay`）

---

## 4. 关键文件路径

| 文件 | 用途 |
|---|---|
| `docs/engineering/payment-b2-deploy.md` | 部署手册（backend 部分完整，frontend 部分需补） |
| `backend/migrations/091a_payment_orders_backup.sql` | 备份+DROP 旧表 |
| `backend/migrations/092b_payment_orders_restore_history.sql` | 历史订单迁入新表（DO 块） |
| `backend/migrations/093a_payment_audit_logs_unique_constraint.sql` | 部分唯一索引 |
| `backend/migrations/096a_payment_provider_instances_unique.sql` | (provider_key, name) UNIQUE |
| `backend/migrations/102a_backfill_out_trade_no_from_backup.sql` | 回填 out_trade_no |
| `backend/internal/service/affiliate_service_stub.go` | Affiliate 最小 stub |
| `backend/internal/service/setting_service_wechat_payment_ext.go` | WeChatConnect helpers + normalizeVisibleMethodSettingSource |
| `backend/internal/service/payment_order.go` (line 549-557) | getWeChatPaymentOAuthCredential JSAPI stub |
| `backend/internal/service/payment_fulfillment.go` (line 461) | ON CONFLICT WHERE 谓词 |
| `backend/cmd/server/wire_gen.go` (line 214-225) | 手工维护的新 payment DI |

---

## 5. 三轮 Review 历史（已修复问题清单）

| 轮次 | 来源 | 问题 | 状态 |
|---|---|---|---|
| R1 | Codex+Kimi | 092b INSERT `out_trade_no` 不存在 | ✅ 修复（092b 移除 + 102a backfill） |
| R1 | Kimi | fresh-install 无 backup 表会报错 | ✅ DO 块 IF NOT EXISTS |
| R1 | Kimi | INNER JOIN 丢孤儿订单 | ✅ 改 LEFT JOIN |
| R2 | Codex+Kimi | 092b INSERT FROM 在解析期就报错 | ✅ 整体移入 DO 块 |
| R2 | Codex+Kimi | audit_logs 无 UNIQUE 约束 | ✅ 093a 部分索引 |
| R3 | Codex | 093a `ADD CONSTRAINT IF NOT EXISTS` 语法错 | ✅ DO 块 + pg_constraint 检查 |
| R3 | Codex | UNIQUE 约束过宽 | ✅ 部分索引 + ON CONFLICT WHERE |
| R3 | Kimi | postcheck COUNT(*) 包含其他数据误报 | ✅ COUNT WHERE id IN backup |
| R4 | Codex | 096a 无幂等保护 | ✅ DO 块 + pg_constraint 检查 |
| 部署中 | 实测 | 113 引用 auth_identities 不存在 | ✅ 删除（auth 模块非 payment） |
| 部署中 | 用户发现 | **frontend 完全没改造** | ⏳ **待做** |

---

## 6. 下一步行动

1. 重新进入 worktree：`cd /Users/nio/project/nanafox/sub2api/.claude/worktrees/payment-b2`
2. 按 §2 顺序做 frontend 改造
3. `pnpm --dir frontend build` 0 错误
4. 本地 `pnpm dev` 启动测试 UI
5. 推送到同一分支 `worktree-payment-b2`，PR #18 自动更新
6. 重新双模型 review（Codex + Kimi 关注 frontend 部分）
7. **测试环境重新部署**（重复 §2-§4 的 backend 部署 + 新前端验证 UI）
8. 用户做端到端测试一笔订单
9. 生产部署

---

## 7. 关键账号/路径

- 服务器：`ssh nio@108.160.133.141`（密码 nio2026.）
- 测试库：`docker exec sub2api-postgres psql -U sub2api -d sub2api_test`
- 生产库：`docker exec sub2api-postgres psql -U sub2api -d sub2api`
- 测试域名：https://router-test.nanafox.com
- 生产域名：https://router.nanafox.com
- 备份：`/data/backups/sub2api_test_pre_b2_*.sql`（写需要 sudo）

---

## 8. 给下一个 session 的提醒

1. **不要继续执行 backend 改造**——已经完成且经过 4 轮 review
2. **专注 frontend**——这是 100% 的剩余工作量
3. **PR #18 已是干净状态**，只追加 commit 不要 reset
4. **测试环境已回滚**，前端改完再重新部署验证
5. **生产环境从未碰过**（只测试环境）
6. **遵循 worktree 工作流**（CLAUDE.md 强制）
7. **每次推送前必须 pnpm build 0 错误**
