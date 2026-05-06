//go:build unit

package handler

import (
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

func TestSanitizePaymentOrderForResponseKeepsZeroMoneyFields(t *testing.T) {
	order := &dbent.PaymentOrder{
		ID:          55,
		UserID:      3,
		PaymentType: "wxpay",
		OrderType:   "subscription",
		Status:      "FAILED",
		ExpiresAt:   time.Date(2026, 4, 29, 13, 21, 9, 0, time.FixedZone("CST", 8*60*60)),
		CreatedAt:   time.Date(2026, 4, 29, 12, 51, 9, 0, time.FixedZone("CST", 8*60*60)),
		UpdatedAt:   time.Date(2026, 4, 29, 12, 51, 9, 0, time.FixedZone("CST", 8*60*60)),
	}

	got := sanitizePaymentOrderForResponse(order)

	if got.Amount != 0 {
		t.Fatalf("expected amount to be 0, got %#v", got.Amount)
	}
	if got.PayAmount != 0 {
		t.Fatalf("expected pay_amount to be 0, got %#v", got.PayAmount)
	}
	if got.FeeRate != 0 {
		t.Fatalf("expected fee_rate to be 0, got %#v", got.FeeRate)
	}
	if got.RefundAmount != 0 {
		t.Fatalf("expected refund_amount to be 0, got %#v", got.RefundAmount)
	}
	if got.ProviderSnapshot != nil {
		t.Fatalf("provider_snapshot should not be exposed: %#v", got.ProviderSnapshot)
	}
}
