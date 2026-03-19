package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentOrderRepository struct {
	client *dbent.Client
}

func NewPaymentOrderRepository(client *dbent.Client) service.PaymentOrderRepository {
	return &paymentOrderRepository{client: client}
}

func (r *paymentOrderRepository) Create(ctx context.Context, order *service.PaymentOrder) error {
	created, err := r.client.PaymentOrder.Create().
		SetOrderNo(order.OrderNo).
		SetUserID(order.UserID).
		SetType(order.Type).
		SetNillablePlanID(order.PlanID).
		SetAmount(order.Amount).
		SetNillableCreditAmount(order.CreditAmount).
		SetCurrency(order.Currency).
		SetStatus(order.Status).
		SetNillableProvider(order.Provider).
		SetNillableProviderOrderNo(order.ProviderOrderNo).
		SetNillablePaidAt(order.PaidAt).
		SetNillableCompletedAt(order.CompletedAt).
		SetNillableRefundedAt(order.RefundedAt).
		SetExpiredAt(order.ExpiredAt).
		SetNillableCallbackRaw(order.CallbackRaw).
		SetNillableAdminNote(order.AdminNote).
		Save(ctx)
	if err != nil {
		return err
	}
	order.ID = created.ID
	order.CreatedAt = created.CreatedAt
	order.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *paymentOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.PaymentOrder, error) {
	m, err := r.client.PaymentOrder.Query().
		Where(paymentorder.OrderNoEQ(orderNo)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrPaymentOrderNotFound
		}
		return nil, err
	}
	return toPaymentOrder(m), nil
}

func (r *paymentOrderRepository) GetByID(ctx context.Context, id int64) (*service.PaymentOrder, error) {
	m, err := r.client.PaymentOrder.Query().
		Where(paymentorder.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrPaymentOrderNotFound
		}
		return nil, err
	}
	return toPaymentOrder(m), nil
}

func (r *paymentOrderRepository) UpdateStatusAtomically(ctx context.Context, orderNo string, fromStatuses []string, toStatus string, updates map[string]any) (int, error) {
	up := r.client.PaymentOrder.Update().
		Where(
			paymentorder.OrderNoEQ(orderNo),
			paymentorder.StatusIn(fromStatuses...),
		).
		SetStatus(toStatus)

	for k, v := range updates {
		switch k {
		case "paid_at":
			if t, ok := v.(time.Time); ok {
				up.SetPaidAt(t)
			}
		case "completed_at":
			if t, ok := v.(time.Time); ok {
				up.SetCompletedAt(t)
			}
		case "refunded_at":
			if t, ok := v.(time.Time); ok {
				up.SetRefundedAt(t)
			}
		case "provider_order_no":
			if s, ok := v.(string); ok {
				up.SetProviderOrderNo(s)
			}
		case "callback_raw":
			if s, ok := v.(string); ok {
				up.SetCallbackRaw(s)
			}
		case "admin_note":
			if s, ok := v.(string); ok {
				up.SetAdminNote(s)
			}
		case "credit_amount":
			if f, ok := v.(float64); ok {
				up.SetCreditAmount(f)
			}
		}
	}

	affected, err := up.Save(ctx)
	return affected, err
}

func (r *paymentOrderRepository) ListByUser(ctx context.Context, userID int64, filter service.OrderFilter, params pagination.PaginationParams) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	q := r.client.PaymentOrder.Query().
		Where(paymentorder.UserIDEQ(userID))

	q = applyOrderFilter(q, filter)

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(paymentorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return toPaymentOrders(orders), paginationResultFromTotal(int64(total), params), nil
}

func (r *paymentOrderRepository) ListAll(ctx context.Context, filter service.OrderFilter, params pagination.PaginationParams) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	q := r.client.PaymentOrder.Query()
	q = applyOrderFilter(q, filter)

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(paymentorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return toPaymentOrders(orders), paginationResultFromTotal(int64(total), params), nil
}

func (r *paymentOrderRepository) ExpirePendingOrders(ctx context.Context) (int, error) {
	affected, err := r.client.PaymentOrder.Update().
		Where(
			paymentorder.StatusEQ(domain.PaymentStatusPending),
			paymentorder.ExpiredAtLT(time.Now()),
		).
		SetStatus(domain.PaymentStatusExpired).
		Save(ctx)
	return affected, err
}

func (r *paymentOrderRepository) Stats(ctx context.Context, filter service.StatsFilter) (*service.OrderStats, error) {
	q := r.client.PaymentOrder.Query().
		Where(paymentorder.StatusEQ(domain.PaymentStatusCompleted))

	if filter.StartDate != "" {
		if t, err := time.Parse("2006-01-02", filter.StartDate); err == nil {
			q = q.Where(paymentorder.CreatedAtGTE(t))
		}
	}
	if filter.EndDate != "" {
		if t, err := time.Parse("2006-01-02", filter.EndDate); err == nil {
			// end of day
			q = q.Where(paymentorder.CreatedAtLT(t.Add(24 * time.Hour)))
		}
	}

	orders, err := q.All(ctx)
	if err != nil {
		return nil, err
	}

	// Aggregate in Go
	layout := "2006-01-02"
	if filter.GroupBy == "month" {
		layout = "2006-01"
	}

	type bucket struct {
		amount float64
		count  int
	}
	buckets := make(map[string]*bucket)
	var keyOrder []string

	totalAmount := 0.0
	for _, o := range orders {
		key := o.CreatedAt.Format(layout)
		if _, exists := buckets[key]; !exists {
			buckets[key] = &bucket{}
			keyOrder = append(keyOrder, key)
		}
		buckets[key].amount += o.Amount
		buckets[key].count++
		totalAmount += o.Amount
	}

	breakdown := make([]service.StatsBreakdown, 0, len(keyOrder))
	for _, key := range keyOrder {
		b := buckets[key]
		breakdown = append(breakdown, service.StatsBreakdown{
			Date:   key,
			Amount: b.amount,
			Count:  b.count,
		})
	}

	return &service.OrderStats{
		TotalAmount: totalAmount,
		TotalOrders: len(orders),
		Breakdown:   breakdown,
	}, nil
}

// applyOrderFilter applies OrderFilter predicates to a PaymentOrder query.
func applyOrderFilter(q *dbent.PaymentOrderQuery, filter service.OrderFilter) *dbent.PaymentOrderQuery {
	if filter.Status != "" {
		q = q.Where(paymentorder.StatusEQ(filter.Status))
	}
	if filter.Type != "" {
		q = q.Where(paymentorder.TypeEQ(filter.Type))
	}
	if filter.UserID != nil {
		q = q.Where(paymentorder.UserIDEQ(*filter.UserID))
	}
	return q
}

func toPaymentOrder(e *dbent.PaymentOrder) *service.PaymentOrder {
	if e == nil {
		return nil
	}
	return &service.PaymentOrder{
		ID:              e.ID,
		OrderNo:         e.OrderNo,
		UserID:          e.UserID,
		Type:            e.Type,
		PlanID:          e.PlanID,
		Amount:          e.Amount,
		CreditAmount:    e.CreditAmount,
		Currency:        e.Currency,
		Status:          e.Status,
		Provider:        e.Provider,
		ProviderOrderNo: e.ProviderOrderNo,
		PaidAt:          e.PaidAt,
		CompletedAt:     e.CompletedAt,
		RefundedAt:      e.RefundedAt,
		ExpiredAt:       e.ExpiredAt,
		CallbackRaw:     e.CallbackRaw,
		AdminNote:       e.AdminNote,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func toPaymentOrders(models []*dbent.PaymentOrder) []service.PaymentOrder {
	out := make([]service.PaymentOrder, 0, len(models))
	for _, m := range models {
		if o := toPaymentOrder(m); o != nil {
			out = append(out, *o)
		}
	}
	return out
}
