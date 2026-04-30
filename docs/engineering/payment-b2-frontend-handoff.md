# Payment B-2 交接文档（前端改造 — 已完成）

**日期**：2026-04-30
**当前状态**：Frontend 改造全部完成，Build 通过，Review 通过（Kimi + Claude），Codex 待补
**Worktree**：`.claude/worktrees/payment-b2/`
**PR**：https://github.com/ddnio/sub2api/pull/18
**最新 Commit**：`6ee63fc1`

---

## 1. 已完成的工作

### 1.1 前端字段同步（upstream payment v2）

所有前端 API 层和视图层已同步新后端字段：

| 旧字段 | 新字段 | 状态 |
|---|---|---|
| `order_no` | `out_trade_no` | 完成 |
| `type` | `order_type`（`subscription` / `balance`）| 完成 |
| `expired_at` | `expires_at` | 完成 |
| `provider` | `payment_type` | 完成 |
| `qr_code_url` | `qr_code` | 完成 |
| `duration_days` | `validity_days` | 完成 |
| `is_active` | `for_sale` | 完成 |
| `badge` | **已移除** | 完成 |

### 1.2 新增接口/API

- `frontend/src/api/admin/payment.ts`: 新增 `ProviderInstance`、`PaymentConfig` 接口及 CRUD API
- `frontend/src/api/admin/index.ts`: 导出新类型

### 1.3 视图改动

- **PaymentView.vue**: 移除 iframe 购买页，改为直接 API 下单 + 二维码展示
- **PaymentOrdersView.vue**: 列同步新字段，状态值大写处理（PENDING/PAID/COMPLETED/FAILED/EXPIRED/REFUNDED）
- **PaymentPlansView.vue**: 表单字段同步（validity_days, for_sale），移除 badge
- **SettingsView.vue**: 新增"支付管理" tab
  - Provider 列表（增删改）
  - 支付全局配置（启用/禁用、最小/最大金额、禁用余额充值）
  - `savePaymentConfig` 已加 500ms debounce
- **Router**: `/purchase` 重定向到 `/payment`

### 1.4 i18n

- `zh.ts` / `en.ts`: 新增 `orderType.subscription/balance`、`adminPayment.*` 系列键
- 已移除 orphan `planBadge` 键
- Provider Dialog 硬编码文本已转 i18n（providerKey, providerConfig, provider）

### 1.5 部署手册

- `docs/engineering/payment-b2-deploy.md`: 新增前端变更章节（第 4 节）

---

## 2. 当前状态

### 2.1 Build 验证

```bash
pnpm --dir frontend build
# 结果: 0 错误, exit 0, built in ~17s
```

### 2.2 TypeScript

```bash
pnpm --dir frontend vue-tsc --noEmit
# 结果: 无错误
```

### 2.3 Git 状态

```
worktree-payment-b2 分支，领先 main 若干 commit
所有改动已 push 到 origin/worktree-payment-b2
PR #18 自动更新
```

---

## 3. Review 结果

### 3.1 Claude code-reviewer agent
- **结论**: APPROVED with suggestions
- Critical: 无
- 已处理: group_name 类型补充、orphan i18n 清理、debounce 添加

### 3.2 Kimi CLI Review
- **结论**: 通过，详细 review 报告已生成
- 验证确认:
  - AdminPaymentPlan/AdminPaymentOrder 通过 `extends` 自动继承新字段
  - CreateOrderResponse 已从 `frontend/src/api/index.ts` 导出
  - PaymentView.vue 的 `loadOrders` 不传类型参数，不受影响
- 已处理: i18n 硬编码文本、providerForm.id 严格判断（`> 0`）

### 3.3 Codex CLI Review
- **状态**: 未完成
- **原因**: Codex CLI OAuth token 过期（`unauthorized`），需运行 `codex login` 完成浏览器授权
- **建议**: 如需 Codex review，先完成登录授权，再重新启动 review

---

## 4. 已知问题 / 待办

| # | 事项 | 优先级 | 说明 |
|---|---|---|---|
| 1 | Codex review | 中 | 需 `codex login` 后重新运行 |
| 2 | Provider Dialog 字段完整度 | 低 | 当前只暴露 name/config/enabled，其他字段（supported_types, payment_mode 等）是可选参数，如需 UI 编辑后续可补充 |
| 3 | PurchaseSubscriptionView.vue 死代码 | 低 | 路由已重定向，但原文件可能还在项目中，可清理 |
| 4 | 端到端测试 | 高 | 部署到测试环境后，需完成一笔真实支付订单验证全流程 |

---

## 5. 关键文件路径

```
frontend/src/api/payment.ts                    # PaymentPlan, PaymentOrder, CreateOrderResponse
frontend/src/api/admin/payment.ts              # AdminPaymentPlan, ProviderInstance, PaymentConfig
frontend/src/api/admin/index.ts                # Barrel export
frontend/src/views/user/PaymentView.vue        # 用户支付页面（已重写）
frontend/src/views/admin/PaymentOrdersView.vue # 订单管理
frontend/src/views/admin/PaymentPlansView.vue  # 套餐管理
frontend/src/views/admin/SettingsView.vue      # 设置页（新增支付管理 tab）
frontend/src/router/index.ts                   # /purchase 重定向
frontend/src/i18n/locales/zh.ts                # 中文 i18n
frontend/src/i18n/locales/en.ts                # 英文 i18n
docs/engineering/payment-b2-deploy.md          # 部署手册
docs/engineering/payment-b2-frontend-handoff.md # 本文件
```

---

## 6. 环境信息

- **Worktree**: `/Users/nio/project/nanafox/sub2api/.claude/worktrees/payment-b2/`
- **分支**: `worktree-payment-b2`
- **测试环境**: `https://router-test.nanafox.com`（→ 127.0.0.1:8081）
- **生产环境**: `https://router.nanafox.com`（→ 127.0.0.1:8080）
- **服务器**: `ssh nio@108.160.133.141`（密码 nio2026.）

---

## 7. 下一步建议

1. **如需 Codex review**: 运行 `codex login` 完成浏览器授权，然后重新启动 review
2. **合并前**: 部署到测试环境，完成端到端支付验证（见部署手册 §6）
3. **合并后**: 按部署手册 §5 配置 wxpay Provider
