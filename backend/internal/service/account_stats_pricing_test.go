//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestResolveAccountStatsCostCustomRuleUsesUpstreamModel(t *testing.T) {
	price := 0.000002
	groupID := int64(11)
	accountID := int64(22)
	channelSvc := newTestChannelService(makeStandardRepo(Channel{
		ID:       1,
		Status:   StatusActive,
		GroupIDs: []int64{groupID},
		AccountStatsPricingRules: []AccountStatsPricingRule{
			{
				AccountIDs: []int64{accountID},
				Pricing: []ChannelModelPricing{
					{
						Platform:    PlatformOpenAI,
						Models:      []string{"gpt-5.1-upstream"},
						BillingMode: BillingModeToken,
						InputPrice:  &price,
					},
				},
			},
		},
	}, map[int64]string{groupID: PlatformOpenAI}))

	got := resolveAccountStatsCost(context.Background(), channelSvc, nil, accountID, groupID, "gpt-5.1-upstream", UsageTokens{
		InputTokens: 100,
	}, 1, 9.9)

	require.NotNil(t, got)
	require.InDelta(t, 0.0002, *got, 1e-12)
}

func TestResolveAccountStatsCostApplyPricingUsesTotalCost(t *testing.T) {
	groupID := int64(11)
	channelSvc := newTestChannelService(makeStandardRepo(Channel{
		ID:                         1,
		Status:                     StatusActive,
		GroupIDs:                   []int64{groupID},
		ApplyPricingToAccountStats: true,
	}, map[int64]string{groupID: PlatformOpenAI}))

	got := resolveAccountStatsCost(context.Background(), channelSvc, NewBillingService(&config.Config{}, nil), 22, groupID, "gpt-5.1", UsageTokens{
		InputTokens: 100,
	}, 1, 1.25)

	require.NotNil(t, got)
	require.Equal(t, 1.25, *got)
}

func TestResolveAccountStatsCostFallsBackToModelFilePricing(t *testing.T) {
	groupID := int64(11)
	channelSvc := newTestChannelService(makeStandardRepo(Channel{
		ID:       1,
		Status:   StatusActive,
		GroupIDs: []int64{groupID},
	}, map[int64]string{groupID: PlatformOpenAI}))

	got := resolveAccountStatsCost(context.Background(), channelSvc, NewBillingService(&config.Config{}, nil), 22, groupID, "gpt-5.1", UsageTokens{
		InputTokens: 100,
	}, 1, 9.9)

	require.NotNil(t, got)
	require.InDelta(t, 0.000125, *got, 1e-12)
}

func TestApplyAccountStatsCostFallsBackToRequestedModel(t *testing.T) {
	groupID := int64(11)
	channelSvc := newTestChannelService(makeStandardRepo(Channel{
		ID:       1,
		Status:   StatusActive,
		GroupIDs: []int64{groupID},
	}, map[int64]string{groupID: PlatformOpenAI}))
	log := &UsageLog{}

	applyAccountStatsCost(context.Background(), log, channelSvc, NewBillingService(&config.Config{}, nil), 22, groupID, "", "gpt-5.1", UsageTokens{
		InputTokens: 100,
	}, 9.9)

	require.NotNil(t, log.AccountStatsCost)
	require.InDelta(t, 0.000125, *log.AccountStatsCost, 1e-12)
}
