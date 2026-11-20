package service

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

func TestPaymentOrderDisplayNoNeverUsesInternalID(t *testing.T) {
	require.Equal(t, "sub2_20261118RQ9Wj6o", paymentOrderDisplayNo(&dbent.PaymentOrder{ID: 207, OutTradeNo: " sub2_20261118RQ9Wj6o "}))
	require.Equal(t, "-", paymentOrderDisplayNo(&dbent.PaymentOrder{ID: 207, OutTradeNo: ""}))
	require.Equal(t, "-", paymentOrderDisplayNo(&dbent.PaymentOrder{ID: 207, OutTradeNo: "  "}))
}

func TestPaymentRefundDefaultReasonUsesOrderNo(t *testing.T) {
	s := &PaymentService{}
	plan := s.refundFinalizePlan(&dbent.PaymentOrder{ID: 207, OutTradeNo: "sub2_20261118RQ9Wj6o"})
	require.Equal(t, "refund order:sub2_20261118RQ9Wj6o", plan.Reason)
}

func TestPaymentSuccessEmailsShowOrderNoNotInternalID(t *testing.T) {
	order := &dbent.PaymentOrder{
		ID:         207,
		OutTradeNo: "sub2_20261118RQ9Wj6o",
		UserID:     42,
		UserEmail:  "user@example.com",
		UserName:   "User",
		Amount:     10,
	}
	sends := map[string]func(*PaymentService, context.Context, *dbent.PaymentOrder) error{
		NotificationEmailEventBalanceRechargeSuccess:      (*PaymentService).sendBalanceRechargeSuccessNotification,
		NotificationEmailEventSubscriptionPurchaseSuccess: (*PaymentService).sendSubscriptionPurchaseSuccessNotification,
	}
	for event, send := range sends {
		t.Run(event, func(t *testing.T) {
			ctx := context.Background()
			repo := newNotificationEmailMemorySettingRepo()
			smtpServer := startNotificationEmailTestSMTPServer(t)
			require.NoError(t, repo.SetMultiple(ctx, smtpServer.settings()))
			s := &PaymentService{notificationEmailService: NewNotificationEmailService(repo, NewEmailService(repo, nil))}

			require.NoError(t, send(s, ctx, order))
			require.Equal(t, int64(1), smtpServer.messageCount())
			body := smtpServer.lastMessageBody(t)
			require.Contains(t, body, "Order ID: sub2_20261118RQ9Wj6o")
			require.NotContains(t, body, "207")

			// The dedup key still uses the internal ID so a resend is suppressed.
			_, err := repo.GetValue(ctx, notificationEmailDeliveryKey(event, "payment_order", "207", order.UserEmail, ""))
			require.NoError(t, err)
		})
	}
}
