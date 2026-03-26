# 微信支付退款功能实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 管理员退款时调用微信支付退款 API 真实退款，而非仅做本地记账。

**Architecture:** 在 `PaymentProvider` 接口新增 `Refund` 方法，`wxpayProvider` 实现调用微信 SDK `refunddomestic` 包，`AdminRefundOrder` 先调微信退款 → 成功后标记状态 → 再扣回余额。使用确定性退款单号 `"R" + orderNo` 保证幂等。

**Tech Stack:** Go, wechatpay-go SDK (v0.2.21), ent ORM, PostgreSQL

**Design Spec:** `docs/plans/2026-03-26-wxpay-refund-design.md`

---

### Task 1: DB Migration — 新增 refund_no 字段

**Files:**
- Create: `backend/migrations/078_add_refund_no.sql`

- [ ] **Step 1: 创建 migration 文件**

```sql
ALTER TABLE payment_orders ADD COLUMN refund_no VARCHAR(64) DEFAULT NULL;
COMMENT ON COLUMN payment_orders.refund_no IS '商户退款单号';
```

- [ ] **Step 2: Commit**

```bash
git add backend/migrations/078_add_refund_no.sql
git commit -m "migration: add refund_no column to payment_orders"
```

---

### Task 2: Ent Schema — 新增 refund_no 字段并重新生成

**Files:**
- Modify: `backend/ent/schema/payment_order.go:85` (在 `admin_note` 字段后添加)

- [ ] **Step 1: 在 ent schema 添加 refund_no 字段**

在 `backend/ent/schema/payment_order.go` 的 `Fields()` 方法中，在 `admin_note` 字段之后、`created_at` 之前添加：

```go
		field.String("refund_no").
			MaxLen(64).
			Optional().
			Nillable(),
```

- [ ] **Step 2: 重新生成 ent 代码**

```bash
cd backend && go generate ./ent
```

验证：生成成功，`backend/ent/paymentorder/` 目录下文件包含 `refund_no` 相关代码。

- [ ] **Step 3: Commit**

```bash
git add backend/ent/
git commit -m "ent: add refund_no field to PaymentOrder schema"
```

---

### Task 3: Service 层 — 扩展接口和类型

**Files:**
- Modify: `backend/internal/service/payment_service.go`

- [ ] **Step 1: PaymentOrder 结构体添加 RefundNo 字段**

在 `backend/internal/service/payment_service.go` 的 `PaymentOrder` 结构体中，在 `AdminNote *string` 后添加：

```go
	RefundNo        *string
```

- [ ] **Step 2: 添加 RefundRequest 和 RefundResult 类型**

在 `PaymentResult` 结构体之后添加：

```go
type RefundRequest struct {
	OrderNo         string  // 商户订单号（out_trade_no）
	ProviderOrderNo string  // 微信支付订单号（transaction_id），可为空
	RefundNo        string  // 商户退款单号（out_refund_no）
	Amount          float64 // 退款金额（元）
	Reason          string  // 退款原因
}

type RefundResult struct {
	ProviderRefundNo string // 微信退款单号
	Status           string // 退款状态（SUCCESS / PROCESSING）
}
```

- [ ] **Step 3: PaymentProvider 接口添加 Refund 方法**

修改 `PaymentProvider` 接口，添加第三个方法：

```go
type PaymentProvider interface {
	CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResult, error)
	ParseCallback(r *http.Request) (*CallbackResult, error)
	Refund(ctx context.Context, req RefundRequest) (*RefundResult, error)
}
```

- [ ] **Step 4: 编译检查**

```bash
cd backend && go build ./...
```

预期：编译失败，因为 `wxpayProvider` 和 `easypayProvider` 还没实现 `Refund` 方法。确认错误信息包含 `does not implement PaymentProvider (missing method Refund)`。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/payment_service.go
git commit -m "feat: add Refund to PaymentProvider interface and RefundRequest/Result types"
```

---

### Task 4: easypayProvider — 空实现 Refund

**Files:**
- Modify: `backend/internal/repository/easypay_provider.go`

- [ ] **Step 1: 在 easypayProvider 添加 Refund 方法**

在文件末尾（`ParseCallback` 方法之后）添加：

```go
func (p *easyPayProvider) Refund(ctx context.Context, req service.RefundRequest) (*service.RefundResult, error) {
	return nil, fmt.Errorf("easypay: refund not supported")
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/repository/easypay_provider.go
git commit -m "feat: add Refund stub to easypay provider (not supported)"
```

---

### Task 5: wxpayProvider — 实现 Refund 方法

**Files:**
- Modify: `backend/internal/repository/wxpay_provider.go`

- [ ] **Step 1: 添加 refunddomestic import**

在 `wxpay_provider.go` 的 import 中添加：

```go
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
```

- [ ] **Step 2: 实现 Refund 方法**

在 `ParseCallback` 方法之后添加：

```go
// Refund 调用微信支付退款 API（国内退款）。
// 微信退款为异步处理，Create 成功（返回 SUCCESS 或 PROCESSING）表示退款已受理。
func (p *wxpayProvider) Refund(ctx context.Context, req service.RefundRequest) (*service.RefundResult, error) {
	svc := refunddomestic.RefundsApiService{Client: p.client}

	amountFen := yuanToFen(req.Amount)

	createReq := refunddomestic.CreateRequest{
		OutTradeNo:  core.String(req.OrderNo),
		OutRefundNo: core.String(req.RefundNo),
		Reason:      core.String(req.Reason),
		Amount: &refunddomestic.AmountReq{
			Refund:   core.Int64(amountFen),
			Total:    core.Int64(amountFen), // 全额退款：refund == total
			Currency: core.String("CNY"),
		},
	}

	// 优先使用微信支付订单号（更精确）
	if req.ProviderOrderNo != "" {
		createReq.TransactionId = core.String(req.ProviderOrderNo)
	}

	resp, _, err := svc.Create(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("wxpay: refund: %w", err)
	}

	result := &service.RefundResult{}
	if resp.RefundId != nil {
		result.ProviderRefundNo = *resp.RefundId
	}
	if resp.Status != nil {
		result.Status = string(*resp.Status) // Status 是 refunddomestic.Status 枚举类型，转为 string
	}

	return result, nil
}
```

- [ ] **Step 3: 编译检查**

```bash
cd backend && go build ./...
```

预期：编译成功。

- [ ] **Step 4: Commit**

```bash
git add backend/internal/repository/wxpay_provider.go
git commit -m "feat: implement wxpay Refund using refunddomestic API"
```

---

### Task 6: Repository — UpdateStatusAtomically 支持 refund_no

**Files:**
- Modify: `backend/internal/repository/payment_order_repo.go:84-114` (switch 语句内)

- [ ] **Step 1: 在 UpdateStatusAtomically 的 switch 中添加 refund_no case**

在 `credit_amount` case 之后（`}` 之前）添加：

```go
		case "refund_no":
			if s, ok := v.(string); ok {
				up.SetRefundNo(s)
			}
```

- [ ] **Step 2: 更新 repository 的 toServiceOrder 映射**

找到 `payment_order_repo.go` 中将 ent entity 映射到 `service.PaymentOrder` 的函数，在其中添加 `RefundNo` 字段映射：

```go
RefundNo:        e.RefundNo,
```

- [ ] **Step 3: 编译检查**

```bash
cd backend && go build ./...
```

预期：编译成功。

- [ ] **Step 4: Commit**

```bash
git add backend/internal/repository/payment_order_repo.go
git commit -m "feat: add refund_no support to payment order repository"
```

---

### Task 7: Service 层 — 改写 AdminRefundOrder

**Files:**
- Modify: `backend/internal/service/payment_service.go:564-601`

- [ ] **Step 1: 改写 AdminRefundOrder 方法**

用以下代码替换 `payment_service.go` 中 `AdminRefundOrder` 方法的完整实现：

```go
func (s *PaymentService) AdminRefundOrder(ctx context.Context, orderID int64, adminNote string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrPaymentOrderNotFound
	}
	if order.Status != domain.PaymentStatusCompleted && order.Status != domain.PaymentStatusPaid {
		return ErrPaymentInvalidStatus
	}

	// 1. Build refund request
	refundNo := "R" + order.OrderNo
	providerOrderNo := ""
	if order.ProviderOrderNo != nil {
		providerOrderNo = *order.ProviderOrderNo
	}

	reason := "管理员退款"
	if adminNote != "" {
		reason = adminNote
	}

	// 2. Call payment provider refund API
	_, err = s.provider.Refund(ctx, RefundRequest{
		OrderNo:         order.OrderNo,
		ProviderOrderNo: providerOrderNo,
		RefundNo:        refundNo,
		Amount:          order.Amount,
		Reason:          reason,
	})
	if err != nil {
		log.Printf("[Payment] Refund API failed for order %s: %v", order.OrderNo, err)
		return ErrPaymentProviderError
	}

	// 3. Mark as refunded + store refund_no (before balance deduction — wxpay refund is irreversible)
	refundedAt := time.Now()
	affected, err := s.orderRepo.UpdateStatusAtomically(ctx, order.OrderNo,
		[]string{domain.PaymentStatusCompleted, domain.PaymentStatusPaid},
		domain.PaymentStatusRefunded,
		map[string]any{
			"refunded_at": refundedAt,
			"admin_note":  adminNote,
			"refund_no":   refundNo,
		},
	)
	if err != nil {
		log.Printf("[Payment] CRITICAL: Refund API succeeded but DB update failed for order %s: %v", order.OrderNo, err)
		return err
	}
	if affected == 0 {
		// Order status already changed (concurrent request), wxpay idempotent refund_no prevents double refund
		return nil
	}

	// 4. Reverse benefits (after status is persisted)
	if order.Status == domain.PaymentStatusCompleted && order.Type == domain.PaymentOrderTypeTopup {
		creditAmount := order.Amount
		if order.CreditAmount != nil {
			creditAmount = *order.CreditAmount
		}
		if err := s.userService.UpdateBalance(ctx, order.UserID, -creditAmount); err != nil {
			log.Printf("[Payment] WARN: Refund succeeded but balance deduction failed for order %s: %v — requires manual fix", order.OrderNo, err)
			// Don't return error — refund already committed, balance issue needs manual resolution
		}
		s.asyncInvalidateCache(order.UserID)
	} else if order.Status == domain.PaymentStatusCompleted && order.Type == domain.PaymentOrderTypePlan {
		log.Printf("[Payment] WARN: Refunding plan order %s - subscription benefits not reversed automatically, handle manually", order.OrderNo)
	}

	return nil
}
```

- [ ] **Step 2: 编译检查**

```bash
cd backend && go build ./...
```

预期：编译成功。

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/payment_service.go
git commit -m "feat: AdminRefundOrder now calls wxpay refund API before local bookkeeping"
```

---

### Task 8: 端到端验证

**Files:** 无新文件

- [ ] **Step 1: 完整编译检查**

```bash
cd backend && go build ./...
```

预期：编译成功，无错误。

- [ ] **Step 2: 部署到测试环境**

在服务器上：

```bash
ssh nio@108.160.133.141
cd /data/service/sub2api
git fetch origin feature/wxpay-refund
git checkout feature/wxpay-refund
bash deploy/deploy-server.sh test
```

验证：
- 容器启动成功（`docker logs sub2api-test --tail 20` 无报错）
- migration 078 自动执行（日志中应有 `078_add_refund_no.sql` 相关信息）

- [ ] **Step 3: 创建测试订单并验证退款**

1. 在测试环境 `https://router-test.nanafox.com` 创建一个 ¥1.00 的充值订单
2. 完成支付
3. 在管理后台点击退款
4. 验证：微信退款 API 调用成功（查看日志无 `Refund API failed` 错误）
5. 验证：订单状态变为 `refunded`
6. 验证：用户余额已扣回
7. 验证：微信退款到账（1-3 个工作日，可在微信商户后台立即确认退款状态）

- [ ] **Step 4: 验证错误场景**

1. 对已退款的订单再次点击退款 → 应返回状态错误（按钮不应显示）
2. 对 `failed` 状态的订单 → 退款按钮不显示（前端已处理）

- [ ] **Step 5: Commit 确认**

确认所有代码已提交，分支干净：

```bash
git status
git log --oneline origin/main..HEAD
```
