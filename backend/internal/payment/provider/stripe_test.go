package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	stripe "github.com/stripe/stripe-go/v85"
)

func TestNewStripeConfigAndMetadata(t *testing.T) {
	t.Parallel()

	provider, err := NewStripe("stripe-1", map[string]string{
		"secretKey":      "sk_test_123",
		"publishableKey": "pk_test_123",
	})
	if err != nil {
		t.Fatalf("NewStripe returned error: %v", err)
	}
	if provider.Name() != "Stripe" {
		t.Fatalf("Name = %q, want Stripe", provider.Name())
	}
	if provider.ProviderKey() != payment.TypeStripe {
		t.Fatalf("ProviderKey = %q, want %q", provider.ProviderKey(), payment.TypeStripe)
	}
	if got := provider.GetPublishableKey(); got != "pk_test_123" {
		t.Fatalf("GetPublishableKey = %q, want pk_test_123", got)
	}
	if got := provider.SupportedTypes(); len(got) != 1 || got[0] != payment.TypeStripe {
		t.Fatalf("SupportedTypes = %#v, want [stripe]", got)
	}

	_, err = NewStripe("stripe-1", map[string]string{"publishableKey": "pk_test_123"})
	if err == nil || !strings.Contains(err.Error(), "secretKey") {
		t.Fatalf("missing secretKey error = %v, want mention secretKey", err)
	}
}

func TestResolveStripeMethodTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty defaults to card", input: "", want: []string{"card"}},
		{name: "unknown defaults to card", input: "unknown", want: []string{"card"}},
		{name: "stripe base defaults to card", input: payment.TypeStripe, want: []string{"card"}},
		{name: "card and link", input: "card, link", want: []string{"card", "link"}},
		{name: "wechat and alipay", input: "wxpay,alipay", want: []string{"wechat_pay", "alipay"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := resolveStripeMethodTypes(tt.input)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("resolveStripeMethodTypes(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}

	if !hasStripeMethod([]string{"card", "wechat_pay"}, "wechat_pay") {
		t.Fatal("hasStripeMethod should find wechat_pay")
	}
	if hasStripeMethod([]string{"card"}, "wechat_pay") {
		t.Fatal("hasStripeMethod should not find absent method")
	}
}

func TestStripeVerifyNotification(t *testing.T) {
	t.Parallel()

	const webhookSecret = "whsec_test_secret"
	provider, err := NewStripe("stripe-1", map[string]string{
		"secretKey":     "sk_test_123",
		"webhookSecret": webhookSecret,
	})
	if err != nil {
		t.Fatalf("NewStripe returned error: %v", err)
	}

	tests := []struct {
		name       string
		eventType  string
		wantStatus string
		wantNil    bool
	}{
		{
			name:       "payment succeeded",
			eventType:  stripeEventPaymentSuccess,
			wantStatus: payment.ProviderStatusSuccess,
		},
		{
			name:       "payment failed",
			eventType:  stripeEventPaymentFailed,
			wantStatus: payment.ProviderStatusFailed,
		},
		{
			name:      "unrelated event is ignored",
			eventType: "charge.refunded",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := stripeWebhookPayload(tt.eventType)
			notification, err := provider.VerifyNotification(context.Background(), payload, map[string]string{
				"stripe-signature": signStripeWebhookPayload(t, payload, webhookSecret),
			})
			if err != nil {
				t.Fatalf("VerifyNotification returned error: %v", err)
			}
			if tt.wantNil {
				if notification != nil {
					t.Fatalf("notification = %#v, want nil", notification)
				}
				return
			}
			if notification == nil {
				t.Fatal("notification is nil")
			}
			if notification.TradeNo != "pi_test_123" {
				t.Fatalf("TradeNo = %q, want pi_test_123", notification.TradeNo)
			}
			if notification.OrderID != "order-123" {
				t.Fatalf("OrderID = %q, want order-123", notification.OrderID)
			}
			if notification.Amount != 12.34 {
				t.Fatalf("Amount = %v, want 12.34", notification.Amount)
			}
			if notification.Status != tt.wantStatus {
				t.Fatalf("Status = %q, want %q", notification.Status, tt.wantStatus)
			}
			if notification.RawData != payload {
				t.Fatal("RawData should preserve original payload")
			}
		})
	}
}

func TestStripeVerifyNotificationErrors(t *testing.T) {
	t.Parallel()

	providerWithoutSecret, err := NewStripe("stripe-1", map[string]string{"secretKey": "sk_test_123"})
	if err != nil {
		t.Fatalf("NewStripe returned error: %v", err)
	}
	if _, err := providerWithoutSecret.VerifyNotification(context.Background(), "{}", nil); err == nil || !strings.Contains(err.Error(), "webhookSecret") {
		t.Fatalf("missing webhookSecret error = %v, want mention webhookSecret", err)
	}

	provider, err := NewStripe("stripe-1", map[string]string{
		"secretKey":     "sk_test_123",
		"webhookSecret": "whsec_test_secret",
	})
	if err != nil {
		t.Fatalf("NewStripe returned error: %v", err)
	}
	if _, err := provider.VerifyNotification(context.Background(), "{}", nil); err == nil || !strings.Contains(err.Error(), "stripe-signature") {
		t.Fatalf("missing signature error = %v, want mention stripe-signature", err)
	}
}

func stripeWebhookPayload(eventType string) string {
	return fmt.Sprintf(`{
  "id": "evt_test_123",
  "object": "event",
  "api_version": %q,
  "type": %q,
  "data": {
    "object": {
      "id": "pi_test_123",
      "object": "payment_intent",
      "amount": 1234,
      "currency": "cny",
      "metadata": {
        "orderId": "order-123"
      }
    }
  }
}`, stripe.APIVersion, eventType)
}

func signStripeWebhookPayload(t *testing.T, payload string, secret string) string {
	t.Helper()
	timestamp := time.Now().Unix()
	signedPayload := fmt.Sprintf("%d.%s", timestamp, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(signedPayload)); err != nil {
		t.Fatalf("write hmac payload: %v", err)
	}
	return fmt.Sprintf("t=%d,v1=%s", timestamp, hex.EncodeToString(mac.Sum(nil)))
}
