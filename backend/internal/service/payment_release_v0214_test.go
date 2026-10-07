//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestReleaseV0214BonusOrderFulfillmentAndReplay(t *testing.T) {
	for _, status := range []string{OrderStatusPaid, OrderStatusRecharging, OrderStatusCompleted} {
		t.Run(status, func(t *testing.T) {
			ctx := context.Background()
			client := newPaymentConfigServiceTestClient(t)
			ensurePaymentAuditOrderActionUniqueIndex(t, ctx, client)
			staleAt := time.Now().Add(-paymentFulfillmentLeaseDuration - time.Minute)
			order := createPaymentFulfillmentSubscriptionOrder(t, ctx, client, status, staleAt)
			order, err := client.PaymentOrder.UpdateOneID(order.ID).
				SetOrderType(payment.OrderTypeBalance).
				SetAmount(110).
				SetPayAmount(100).
				SetBonusAmount(10).
				ClearPlanID().
				ClearSubscriptionGroupID().
				ClearSubscriptionDays().
				SetUpdatedAt(staleAt).
				Save(ctx)
			require.NoError(t, err)
			require.Equal(t, 100.0, affiliateRebateBaseAmount(order))

			redeemRepo := &paymentFulfillmentRedeemRepo{}
			credited := 0.0
			userRepo := &mockUserRepo{getByIDUser: &User{ID: order.UserID}}
			userRepo.updateBalanceFn = func(_ context.Context, id int64, amount float64) error {
				require.Equal(t, order.UserID, id)
				credited += amount
				return nil
			}
			cache := &paymentFulfillmentRedeemCacheStub{}
			redeemService := NewRedeemService(redeemRepo, userRepo, nil, cache, nil, client, nil, nil, nil)
			if status == OrderStatusRecharging {
				usedBy := order.UserID
				redeemService = &RedeemService{redeemRepo: &redeemCodeRepoStub{codesByCode: map[string]*RedeemCode{
					order.RechargeCode: {ID: 101, Code: order.RechargeCode, Type: RedeemTypeBalance, Value: 110, Status: StatusUsed, UsedBy: &usedBy},
				}}}
			}
			svc := &PaymentService{entClient: client, redeemService: redeemService, userRepo: userRepo}
			if status != OrderStatusCompleted {
				require.NoError(t, svc.ExecuteBalanceFulfillment(ctx, order.ID))
			}
			notification := &payment.PaymentNotification{
				TradeNo: "release-bonus-replay", OrderID: order.OutTradeNo,
				Amount: 100, Status: payment.NotificationStatusSuccess,
			}
			require.NoError(t, svc.HandlePaymentNotification(ctx, notification, payment.TypeAlipay))
			require.NoError(t, svc.HandlePaymentNotification(ctx, notification, payment.TypeAlipay))
			if status == OrderStatusPaid {
				require.Equal(t, 110.0, credited)
				require.Len(t, redeemRepo.useCalls, 1)
			} else {
				require.Zero(t, credited)
				require.Empty(t, redeemRepo.useCalls)
			}
			reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
			require.NoError(t, err)
			require.Equal(t, OrderStatusCompleted, reloaded.Status)
			require.Equal(t, 110.0, reloaded.Amount)
			require.Equal(t, 10.0, reloaded.BonusAmount)
			require.Equal(t, 100.0, affiliateRebateBaseAmount(reloaded))
		})
	}
}
