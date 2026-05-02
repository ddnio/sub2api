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

	for _, key := range []string{"amount", "pay_amount", "fee_rate", "refund_amount"} {
		value, ok := got[key]
		if !ok {
			t.Fatalf("expected %q to be present in response: %#v", key, got)
		}
		if value != float64(0) {
			t.Fatalf("expected %q to be 0, got %#v", key, value)
		}
	}
	if _, ok := got["provider_snapshot"]; ok {
		t.Fatalf("provider_snapshot should not be exposed: %#v", got)
	}
}
