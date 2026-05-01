package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func TestPaymentConfigServiceCreatePlanRejectsStandardGroup(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	standardGroup := client.Group.Create().
		SetName("standard").
		SetSubscriptionType(domain.SubscriptionTypeStandard).
		SaveX(ctx)
	svc := &PaymentConfigService{entClient: client}

	_, err := svc.CreatePlan(ctx, CreatePlanRequest{
		GroupID:      int64(standardGroup.ID),
		Name:         "Pro",
		Price:        10,
		ValidityDays: 30,
		ValidityUnit: "day",
		ForSale:      true,
	})

	if err == nil {
		t.Fatal("expected standard group to be rejected")
	}
	if got := infraerrors.Reason(err); got != "PLAN_GROUP_INVALID" {
		t.Fatalf("error reason = %q, want PLAN_GROUP_INVALID", got)
	}
}

func TestPaymentConfigServiceCreatePlanAllowsSubscriptionGroup(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	subGroup := client.Group.Create().
		SetName("subscription").
		SetSubscriptionType(domain.SubscriptionTypeSubscription).
		SaveX(ctx)
	svc := &PaymentConfigService{entClient: client}

	plan, err := svc.CreatePlan(ctx, CreatePlanRequest{
		GroupID:      int64(subGroup.ID),
		Name:         "Pro",
		Price:        10,
		ValidityDays: 30,
		ValidityUnit: "day",
		ForSale:      true,
	})

	if err != nil {
		t.Fatalf("CreatePlan returned error: %v", err)
	}
	if plan.GroupID != int64(subGroup.ID) {
		t.Fatalf("plan.GroupID = %d, want %d", plan.GroupID, subGroup.ID)
	}
}

func TestPaymentConfigServiceUpdatePlanRejectsStandardGroup(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	subGroup := client.Group.Create().
		SetName("subscription").
		SetSubscriptionType(domain.SubscriptionTypeSubscription).
		SaveX(ctx)
	standardGroup := client.Group.Create().
		SetName("standard").
		SetSubscriptionType(domain.SubscriptionTypeStandard).
		SaveX(ctx)
	plan := client.SubscriptionPlan.Create().
		SetGroupID(int64(subGroup.ID)).
		SetName("Pro").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("day").
		SaveX(ctx)
	svc := &PaymentConfigService{entClient: client}
	newGroupID := int64(standardGroup.ID)

	_, err := svc.UpdatePlan(ctx, plan.ID, UpdatePlanRequest{GroupID: &newGroupID})

	if err == nil {
		t.Fatal("expected standard group to be rejected")
	}
	if got := infraerrors.Reason(err); got != "PLAN_GROUP_INVALID" {
		t.Fatalf("error reason = %q, want PLAN_GROUP_INVALID", got)
	}
}
