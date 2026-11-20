//go:build unit

package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
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

func TestPrepareRefundDefaultReasonUsesOrderNo(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().SetEmail("refund-reason@example.com").
		SetPasswordHash("hash").SetUsername("refund-reason").Save(ctx)
	require.NoError(t, err)
	inst, err := client.PaymentProviderInstance.Create().SetProviderKey(payment.TypeEasyPay).
		SetName("refund-reason").SetConfig("{}").SetSupportedTypes("wxpay").
		SetEnabled(true).SetRefundEnabled(true).Save(ctx)
	require.NoError(t, err)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).SetUserEmail(user.Email).SetUserName(user.Username).
		SetAmount(100).SetPayAmount(100).SetFeeRate(0).
		SetRechargeCode("REFUND-REASON").SetOutTradeNo("sub2_20261118RQ9Wj6o").
		SetPaymentType(payment.TypeWxpay).SetPaymentTradeNo("trade-reason").
		SetOrderType(payment.OrderTypeBalance).SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").SetSrcHost("api.example.com").
		SetProviderInstanceID(strconv.FormatInt(inst.ID, 10)).Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client, userRepo: &mockUserRepo{getByIDUser: &User{Balance: 200}}}
	plan, early, err := svc.PrepareRefund(ctx, order.ID, 0, "", false, false)
	require.NoError(t, err)
	require.Nil(t, early)
	require.Equal(t, "refund order:sub2_20261118RQ9Wj6o", plan.Reason)
}

func TestPaymentSuccessEmailsShowOrderNoNotInternalID(t *testing.T) {
	order := &dbent.PaymentOrder{
		ID:         987654321,
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
			require.NotContains(t, body, "987654321")

			// The dedup key still uses the internal ID so a resend is suppressed.
			_, err := repo.GetValue(ctx, notificationEmailDeliveryKey(event, "payment_order", "987654321", order.UserEmail, ""))
			require.NoError(t, err)
		})
	}
}
