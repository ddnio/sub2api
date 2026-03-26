# 微信支付退款功能设计

日期：2026-03-26
分支：feature/wxpay-refund
状态：已确认

## 背景

当前 `AdminRefundOrder` 仅做本地记账（扣回系统余额 + 标记 refunded），不调用微信支付退款 API。管理员点退款后用户不会收到真实退款，资金"消失"。

## 目标

管理员发起退款时，系统调用微信支付退款 API 将资金退回用户微信账户，成功后再扣回系统余额并标记订单状态。

## 范围

- 仅管理员后台退款，不做用户自助退款
- 仅全额退款，不支持部分退款
- 仅 wxpay provider 实现退款；easypay provider 返回 not supported

## 设计

### 接口变更

`PaymentProvider` 新增 `Refund` 方法：

```go
type PaymentProvider interface {
    CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResult, error)
    ParseCallback(r *http.Request) (*CallbackResult, error)
    Refund(ctx context.Context, req RefundRequest) (*RefundResult, error)
}

type RefundRequest struct {
    OrderNo         string  // 商户订单号（out_trade_no）
    ProviderOrderNo string  // 微信支付订单号（transaction_id）
    RefundNo        string  // 商户退款单号（out_refund_no）
    Amount          float64 // 退款金额（元）
    Reason          string  // 退款原因
}

type RefundResult struct {
    ProviderRefundNo string // 微信退款单号
    Status           string // 退款状态（SUCCESS / PROCESSING）
}
```

### wxpayProvider.Refund 实现

- 使用 SDK `services/refunddomestic` 包
- 调用 `RefundsApi.Create`
- 退款金额 = 订单全额，元→分转换
- 退款单号由 service 层生成并传入
- 微信退款为异步处理，`Create` 成功表示受理，资金 1-3 工作日退回

### easypayProvider.Refund 实现

返回 `fmt.Errorf("easypay: refund not supported")`。

### AdminRefundOrder 改动

当前流程：
```
扣余额 → 标记 refunded（未调微信退款）
```

改后流程：
```
1. 校验订单状态（paid / completed）
2. 校验 provider_order_no 存在
3. 生成退款单号
4. 调用 provider.Refund()
5. 成功 → 扣回系统余额（topup）/ 记日志（plan）
6. 标记 refunded，存储 refund_no
7. 失败 → 返回错误，不做任何操作
```

### DB 变更

新增 migration `078_add_refund_no.sql`：

```sql
ALTER TABLE payment_orders ADD COLUMN refund_no VARCHAR(64) DEFAULT NULL;
```

### 前端变更

无。现有退款按钮和 API 调用已完备。

### 需要修改的文件

| 文件 | 改动 |
|------|------|
| `backend/internal/service/payment_service.go` | 新增 RefundRequest/RefundResult 类型，Refund 加入接口，改写 AdminRefundOrder |
| `backend/internal/repository/wxpay_provider.go` | 实现 Refund 方法 |
| `backend/internal/repository/easypay_provider.go` | 空实现 Refund 方法 |
| `backend/internal/repository/wire_gen.go` | 无需改动（Provider 初始化不变） |
| `backend/migrations/078_add_refund_no.sql` | 新增 refund_no 字段 |
| `backend/ent/schema/paymentorder.go` | 新增 refund_no 字段 |
| `backend/ent/` | 重新生成 ent 代码 |

### 错误处理

- 微信退款 API 调用失败：返回错误给管理员，订单状态不变，不扣余额
- 退款成功但扣余额失败：订单已标记 refunded，日志记录错误，需人工处理（概率极低，本地 DB 操作）

### 不做的事

- 退款回调通知（微信退款结果通知）：一期不处理，管理员可通过微信商户后台查看退款状态
- 部分退款
- 用户自助退款
