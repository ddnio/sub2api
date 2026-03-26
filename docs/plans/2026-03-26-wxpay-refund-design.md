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

### 退款单号生成策略

使用确定性格式：`"R" + order.OrderNo`（如 `R20260324161606776db5077ed00663f2`）。

好处：
- 同一订单多次调用退款 API 会产生相同的 `out_refund_no`，微信侧自动幂等去重
- 无需额外的并发锁即可防止双击退款

### AdminRefundOrder 改动

当前流程：
```
扣余额 → 标记 refunded（未调微信退款）
```

改后流程：
```
1. 校验订单状态（paid / completed）
2. 获取 provider_order_no（若为空，使用 order_no 兜底，wxpay 退款 API 支持两者）
3. 生成退款单号："R" + orderNo（确定性，天然幂等）
4. 调用 provider.Refund()
5. 微信返回 SUCCESS 或 PROCESSING 均视为成功（PROCESSING 表示已受理，退款失败极罕见）
6. 标记 refunded + 存储 refund_no（先持久化状态，因为微信退款已不可逆）
7. 扣回系统余额（topup）/ 记日志（plan）—— 若此步失败，订单已标记 refunded，日志记录，需人工补扣
8. 失败 → 返回错误，不做任何操作
```

步骤 6 在步骤 7 之前的原因：微信退款一旦受理即不可逆，应优先持久化退款状态，避免进程崩溃导致本地状态与微信不一致。余额扣除是本地 DB 操作，失败概率极低，即使失败也可人工补救。

### DB 变更

新增 migration `078_add_refund_no.sql`：

```sql
ALTER TABLE payment_orders ADD COLUMN refund_no VARCHAR(64) DEFAULT NULL;
COMMENT ON COLUMN payment_orders.refund_no IS '商户退款单号';
```

### 前端变更

无。现有退款按钮和 API 调用已完备。

### 需要修改的文件

| 文件 | 改动 |
|------|------|
| `backend/internal/service/payment_service.go` | 新增 RefundRequest/RefundResult 类型，PaymentOrder 结构体加 RefundNo 字段，Refund 加入接口，改写 AdminRefundOrder |
| `backend/internal/repository/wxpay_provider.go` | 实现 Refund 方法 |
| `backend/internal/repository/easypay_provider.go` | 空实现 Refund 方法 |
| `backend/internal/repository/wire_gen.go` | 无需改动（Provider 初始化不变） |
| `backend/migrations/078_add_refund_no.sql` | 新增 refund_no 字段 |
| `backend/ent/schema/paymentorder.go` | 新增 refund_no 字段 |
| `backend/ent/` | 重新生成 ent 代码 |

### 错误处理

- 微信退款 API 调用失败：返回错误给管理员，订单状态不变，不扣余额
- 退款成功但扣余额失败：订单已标记 refunded，日志记录错误，需人工补扣（概率极低，本地 DB 操作）
- 双击退款：确定性 refund_no 保证微信侧幂等，不会重复退款
- PROCESSING 状态：微信退款为异步处理，返回 PROCESSING 表示已受理，视为成功进行本地操作

### 不做的事

- 退款回调通知（微信退款结果通知）：一期不处理，管理员可通过微信商户后台查看退款状态
- 部分退款
- 用户自助退款
