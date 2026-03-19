package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentplan"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentPlanRepository struct {
	client *dbent.Client
}

func NewPaymentPlanRepository(client *dbent.Client) service.PaymentPlanRepository {
	return &paymentPlanRepository{client: client}
}

func (r *paymentPlanRepository) Create(ctx context.Context, plan *service.PaymentPlan) error {
	created, err := r.client.PaymentPlan.Create().
		SetName(plan.Name).
		SetDescription(plan.Description).
		SetNillableBadge(plan.Badge).
		SetGroupID(plan.GroupID).
		SetDurationDays(plan.DurationDays).
		SetPrice(plan.Price).
		SetNillableOriginalPrice(plan.OriginalPrice).
		SetSortOrder(plan.SortOrder).
		SetIsActive(plan.IsActive).
		Save(ctx)
	if err != nil {
		return err
	}
	plan.ID = created.ID
	plan.CreatedAt = created.CreatedAt
	plan.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *paymentPlanRepository) Update(ctx context.Context, id int64, updates map[string]any) (*service.PaymentPlan, error) {
	up := r.client.PaymentPlan.UpdateOneID(id)

	for k, v := range updates {
		switch k {
		case "name":
			if s, ok := v.(string); ok {
				up.SetName(s)
			}
		case "description":
			if s, ok := v.(string); ok {
				up.SetDescription(s)
			}
		case "badge":
			if s, ok := v.(string); ok {
				up.SetBadge(s)
			} else if v == nil {
				up.ClearBadge()
			}
		case "group_id":
			if i, ok := v.(int64); ok {
				up.SetGroupID(i)
			}
		case "duration_days":
			if i, ok := v.(int); ok {
				up.SetDurationDays(i)
			}
		case "price":
			if f, ok := v.(float64); ok {
				up.SetPrice(f)
			}
		case "original_price":
			if f, ok := v.(float64); ok {
				up.SetOriginalPrice(f)
			} else if v == nil {
				up.ClearOriginalPrice()
			}
		case "sort_order":
			if i, ok := v.(int); ok {
				up.SetSortOrder(i)
			}
		case "is_active":
			if b, ok := v.(bool); ok {
				up.SetIsActive(b)
			}
		}
	}

	updated, err := up.Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrPaymentPlanNotFound
		}
		return nil, err
	}
	return toPaymentPlan(updated), nil
}

func (r *paymentPlanRepository) GetByID(ctx context.Context, id int64) (*service.PaymentPlan, error) {
	ctxSkip := mixins.SkipSoftDelete(ctx)
	m, err := r.client.PaymentPlan.Query().
		Where(paymentplan.IDEQ(id)).
		WithGroup().
		Only(ctxSkip)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrPaymentPlanNotFound
		}
		return nil, err
	}
	return toPaymentPlan(m), nil
}

func (r *paymentPlanRepository) GetByIDActive(ctx context.Context, id int64) (*service.PaymentPlan, error) {
	m, err := r.client.PaymentPlan.Query().
		Where(
			paymentplan.IDEQ(id),
			paymentplan.IsActive(true),
		).
		WithGroup().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrPaymentPlanNotFound
		}
		return nil, err
	}
	return toPaymentPlan(m), nil
}

func (r *paymentPlanRepository) ListActive(ctx context.Context) ([]service.PaymentPlan, error) {
	plans, err := r.client.PaymentPlan.Query().
		Where(paymentplan.IsActive(true)).
		WithGroup().
		Order(paymentplan.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.PaymentPlan, 0, len(plans))
	for _, p := range plans {
		if s := toPaymentPlan(p); s != nil {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (r *paymentPlanRepository) ListAll(ctx context.Context, params pagination.PaginationParams) ([]service.PaymentPlan, *pagination.PaginationResult, error) {
	ctxSkip := mixins.SkipSoftDelete(ctx)

	q := r.client.PaymentPlan.Query()

	total, err := q.Count(ctxSkip)
	if err != nil {
		return nil, nil, err
	}

	plans, err := r.client.PaymentPlan.Query().
		WithGroup().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(paymentplan.FieldCreatedAt)).
		All(ctxSkip)
	if err != nil {
		return nil, nil, err
	}

	out := make([]service.PaymentPlan, 0, len(plans))
	for _, p := range plans {
		if s := toPaymentPlan(p); s != nil {
			out = append(out, *s)
		}
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (r *paymentPlanRepository) SoftDelete(ctx context.Context, id int64) error {
	now := time.Now()
	err := r.client.PaymentPlan.UpdateOneID(id).
		SetDeletedAt(now).
		Exec(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrPaymentPlanNotFound
		}
		return err
	}
	return nil
}

func toPaymentPlan(e *dbent.PaymentPlan) *service.PaymentPlan {
	if e == nil {
		return nil
	}
	p := &service.PaymentPlan{
		ID:           e.ID,
		Name:         e.Name,
		Description:  e.Description,
		Badge:        e.Badge,
		GroupID:      e.GroupID,
		DurationDays: e.DurationDays,
		Price:        e.Price,
		SortOrder:    e.SortOrder,
		IsActive:     e.IsActive,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
	if e.OriginalPrice != nil {
		p.OriginalPrice = e.OriginalPrice
	}
	if g := e.Edges.Group; g != nil {
		p.GroupName = g.Name
		if g.DailyLimitUsd != nil {
			p.DailyLimitUSD = *g.DailyLimitUsd
		}
		if g.WeeklyLimitUsd != nil {
			p.WeeklyLimitUSD = *g.WeeklyLimitUsd
		}
		if g.MonthlyLimitUsd != nil {
			p.MonthlyLimitUSD = *g.MonthlyLimitUsd
		}
	}
	return p
}
