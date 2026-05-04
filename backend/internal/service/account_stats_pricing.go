package service

import (
	"context"
	"strings"
)

// resolveAccountStatsCost returns nil when account stats should use the default
// formula: total_cost * account_rate_multiplier.
func resolveAccountStatsCost(
	ctx context.Context,
	channelService *ChannelService,
	billingService *BillingService,
	accountID int64,
	groupID int64,
	upstreamModel string,
	tokens UsageTokens,
	requestCount int,
	totalCost float64,
) *float64 {
	if channelService == nil || upstreamModel == "" {
		return nil
	}
	channel, err := channelService.GetChannelForGroup(ctx, groupID)
	if err != nil || channel == nil {
		return nil
	}

	platform := channelService.GetGroupPlatform(ctx, groupID)
	if cost := tryAccountStatsCustomRules(channel, accountID, groupID, platform, upstreamModel, tokens, requestCount); cost != nil {
		return cost
	}

	if channel.ApplyPricingToAccountStats {
		cost := totalCost
		if cost <= 0 {
			return nil
		}
		return &cost
	}

	if billingService != nil {
		return tryAccountStatsModelFilePricing(billingService, upstreamModel, tokens)
	}
	return nil
}

func tryAccountStatsModelFilePricing(billingService *BillingService, model string, tokens UsageTokens) *float64 {
	pricing, err := billingService.GetModelPricing(model)
	if err != nil || pricing == nil {
		return nil
	}
	cost := float64(tokens.InputTokens)*pricing.InputPricePerToken +
		float64(tokens.OutputTokens)*pricing.OutputPricePerToken +
		float64(tokens.CacheCreationTokens)*pricing.CacheCreationPricePerToken +
		float64(tokens.CacheReadTokens)*pricing.CacheReadPricePerToken +
		float64(tokens.ImageOutputTokens)*pricing.ImageOutputPricePerToken
	if cost <= 0 {
		return nil
	}
	return &cost
}

func tryAccountStatsCustomRules(
	channel *Channel,
	accountID int64,
	groupID int64,
	platform string,
	model string,
	tokens UsageTokens,
	requestCount int,
) *float64 {
	modelLower := strings.ToLower(model)
	for _, rule := range channel.AccountStatsPricingRules {
		if !matchAccountStatsRule(&rule, accountID, groupID) {
			continue
		}
		pricing := findAccountStatsPricingForModel(rule.Pricing, platform, modelLower)
		if pricing == nil {
			continue
		}
		return calculateAccountStatsCost(pricing, tokens, requestCount)
	}
	return nil
}

func matchAccountStatsRule(rule *AccountStatsPricingRule, accountID int64, groupID int64) bool {
	if len(rule.AccountIDs) == 0 && len(rule.GroupIDs) == 0 {
		return false
	}
	for _, id := range rule.AccountIDs {
		if id == accountID {
			return true
		}
	}
	for _, id := range rule.GroupIDs {
		if id == groupID {
			return true
		}
	}
	return false
}

func findAccountStatsPricingForModel(pricingList []ChannelModelPricing, platform string, modelLower string) *ChannelModelPricing {
	for i := range pricingList {
		p := &pricingList[i]
		if !isAccountStatsPlatformMatch(platform, p.Platform) {
			continue
		}
		for _, model := range p.Models {
			if strings.ToLower(model) == modelLower {
				return p
			}
		}
	}
	for i := range pricingList {
		p := &pricingList[i]
		if !isAccountStatsPlatformMatch(platform, p.Platform) {
			continue
		}
		for _, model := range p.Models {
			normalized := strings.ToLower(model)
			if !strings.HasSuffix(normalized, "*") {
				continue
			}
			if strings.HasPrefix(modelLower, strings.TrimSuffix(normalized, "*")) {
				return p
			}
		}
	}
	return nil
}

func isAccountStatsPlatformMatch(queryPlatform string, pricingPlatform string) bool {
	if queryPlatform == "" || pricingPlatform == "" {
		return true
	}
	return queryPlatform == pricingPlatform
}

func calculateAccountStatsCost(pricing *ChannelModelPricing, tokens UsageTokens, requestCount int) *float64 {
	if pricing == nil {
		return nil
	}
	switch pricing.BillingMode {
	case BillingModePerRequest, BillingModeImage:
		return calculateAccountStatsPerRequestCost(pricing, requestCount)
	default:
		return calculateAccountStatsTokenCost(pricing, tokens)
	}
}

func calculateAccountStatsPerRequestCost(pricing *ChannelModelPricing, requestCount int) *float64 {
	if pricing.PerRequestPrice == nil || *pricing.PerRequestPrice <= 0 {
		return nil
	}
	cost := *pricing.PerRequestPrice * float64(requestCount)
	return &cost
}

func calculateAccountStatsTokenCost(pricing *ChannelModelPricing, tokens UsageTokens) *float64 {
	selected := pricing
	if len(pricing.Intervals) > 0 {
		totalTokens := tokens.InputTokens + tokens.OutputTokens + tokens.CacheCreationTokens + tokens.CacheReadTokens
		if interval := FindMatchingInterval(pricing.Intervals, totalTokens); interval != nil {
			selected = &ChannelModelPricing{
				InputPrice:      interval.InputPrice,
				OutputPrice:     interval.OutputPrice,
				CacheWritePrice: interval.CacheWritePrice,
				CacheReadPrice:  interval.CacheReadPrice,
				PerRequestPrice: interval.PerRequestPrice,
			}
		}
	}
	value := func(ptr *float64) float64 {
		if ptr == nil {
			return 0
		}
		return *ptr
	}
	cost := float64(tokens.InputTokens)*value(selected.InputPrice) +
		float64(tokens.OutputTokens)*value(selected.OutputPrice) +
		float64(tokens.CacheCreationTokens)*value(selected.CacheWritePrice) +
		float64(tokens.CacheReadTokens)*value(selected.CacheReadPrice) +
		float64(tokens.ImageOutputTokens)*value(selected.ImageOutputPrice)
	if cost <= 0 {
		return nil
	}
	return &cost
}

func applyAccountStatsCost(
	ctx context.Context,
	usageLog *UsageLog,
	channelService *ChannelService,
	billingService *BillingService,
	accountID int64,
	groupID int64,
	upstreamModel string,
	requestedModel string,
	tokens UsageTokens,
	totalCost float64,
) {
	model := upstreamModel
	if model == "" {
		model = requestedModel
	}
	usageLog.AccountStatsCost = resolveAccountStatsCost(
		ctx, channelService, billingService, accountID, groupID, model, tokens, 1, totalCost,
	)
}
