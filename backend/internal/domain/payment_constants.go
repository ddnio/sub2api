// backend/internal/domain/payment_constants.go
package domain

// Payment order types
const (
	PaymentOrderTypePlan  = "plan"
	PaymentOrderTypeTopup = "topup"
)

// Payment order status
const (
	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusExpired   = "expired"
	PaymentStatusRefunded  = "refunded"
)

// Payment providers
const (
	PaymentProviderAlipay = "alipay"
	PaymentProviderWxpay  = "wxpay"
)

// Default currency
const (
	PaymentCurrencyCNY = "CNY"
)
