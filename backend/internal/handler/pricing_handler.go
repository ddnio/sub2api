package handler

import (
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// PricingHandler handles pricing-related HTTP requests.
type PricingHandler struct {
	billingService *service.BillingService
	gatewayService *service.GatewayService
	apiKeyService  *service.APIKeyService
}

// NewPricingHandler creates a new PricingHandler.
func NewPricingHandler(
	billingService *service.BillingService,
	gatewayService *service.GatewayService,
	apiKeyService *service.APIKeyService,
) *PricingHandler {
	return &PricingHandler{
		billingService: billingService,
		gatewayService: gatewayService,
		apiKeyService:  apiKeyService,
	}
}

// GetModelPricing returns model pricing information.
// GET /api/v1/pricing/models?group_id=<optional>
func (h *PricingHandler) GetModelPricing(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse optional group_id query parameter.
	var requestedGroupID *int64
	if gidStr := c.Query("group_id"); gidStr != "" {
		gid, err := strconv.ParseInt(gidStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid group_id")
			return
		}
		requestedGroupID = &gid
	}

	// Load user's available groups for authorization and rate info.
	groups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.InternalError(c, "failed to load groups")
		return
	}

	// Build a lookup map of authorized groups.
	groupMap := make(map[int64]*service.Group, len(groups))
	for i := range groups {
		groupMap[groups[i].ID] = &groups[i]
	}

	// If a specific group_id was requested, validate that the user has access.
	if requestedGroupID != nil {
		if _, authorized := groupMap[*requestedGroupID]; !authorized {
			response.BadRequest(c, "group not available")
			return
		}
	}

	// Collect models: either from one group or all user groups.
	modelSet := make(map[string]struct{})
	var targetGroups []service.Group

	if requestedGroupID != nil {
		g := groupMap[*requestedGroupID]
		targetGroups = []service.Group{*g}
	} else {
		targetGroups = groups
	}

	for i := range targetGroups {
		gid := targetGroups[i].ID
		platform := targetGroups[i].Platform
		models := h.gatewayService.GetAvailableModels(c.Request.Context(), &gid, "")
		if len(models) == 0 {
			// C3 fix: platform-aware default fallback (matching /v1/models behavior).
			addDefaultModelsForPlatform(modelSet, platform)
		} else {
			for _, m := range models {
				modelSet[m] = struct{}{}
			}
		}
	}

	// Sort model IDs for deterministic output.
	modelIDs := make([]string, 0, len(modelSet))
	for id := range modelSet {
		modelIDs = append(modelIDs, id)
	}
	sort.Strings(modelIDs)

	// Load user-specific group rate overrides (C2 fix: use per-user rates).
	// F2: log error but fallback to group default rate (non-blocking).
	userGroupRates, userRateErr := h.apiKeyService.GetUserGroupRates(c.Request.Context(), subject.UserID)
	if userRateErr != nil {
		log.Printf("[Pricing] failed to load user group rates for user %d, falling back to group defaults: %v", subject.UserID, userRateErr)
	}

	// Resolve effective rate multiplier for the selected group.
	var effectiveRate float64 = 1.0
	if requestedGroupID != nil {
		g := groupMap[*requestedGroupID]
		effectiveRate = g.RateMultiplier
		// User-specific override takes precedence.
		if userGroupRates != nil {
			if userRate, ok := userGroupRates[*requestedGroupID]; ok {
				effectiveRate = userRate
			}
		}
		if effectiveRate < 0 {
			effectiveRate = 1.0
		}
	}

	// Build pricing items.
	items := make([]dto.ModelPricingItem, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		pricing, err := h.billingService.GetModelPricing(modelID)

		item := dto.ModelPricingItem{
			ID:          modelID,
			DisplayName: modelID,
			OwnedBy:     categorizeModel(modelID),
			Category:    categorizeModel(modelID),
		}

		// C1 fix: nil pricing → null fields (not zero = "free").
		if err != nil || pricing == nil {
			item.Pricing = dto.PriceSet{} // all nil pointers
		} else {
			item.Pricing = buildPriceSet(pricing)
		}

		// Compute effective pricing when a single group is selected.
		if requestedGroupID != nil {
			if err != nil || pricing == nil {
				item.EffectivePricing = &dto.EffectivePriceSet{RateMultiplier: effectiveRate}
			} else {
				item.EffectivePricing = buildEffectivePriceSet(pricing, effectiveRate)
			}
		}

		items = append(items, item)
	}

	// Build group info for the response (use effective rate, not just group default).
	var groupInfo *dto.GroupInfo
	if requestedGroupID != nil {
		g := groupMap[*requestedGroupID]
		groupInfo = &dto.GroupInfo{
			ID:             g.ID,
			Name:           g.Name,
			RateMultiplier: effectiveRate,
		}
	}

	resp := dto.ModelPricingResponse{
		Models: items,
		Group:  groupInfo,
	}
	if requestedGroupID != nil {
		resp.Notice = "Effective prices are estimates based on group rate multiplier. Actual billing may vary."
	}

	response.Success(c, resp)
}

// perMillion converts a per-token price to per-million-token price.
const perMillion = 1_000_000

// buildPriceSet converts ModelPricing (per-token) to dto.PriceSet (per-million).
// Returns all-nil PriceSet if p is nil (pricing unavailable).
func buildPriceSet(p *service.ModelPricing) dto.PriceSet {
	if p == nil {
		return dto.PriceSet{}
	}
	inputPM := p.InputPricePerToken * perMillion
	outputPM := p.OutputPricePerToken * perMillion
	ps := dto.PriceSet{
		InputPerMillion:  &inputPM,
		OutputPerMillion: &outputPM,
	}
	if p.CacheReadPricePerToken > 0 {
		v := p.CacheReadPricePerToken * perMillion
		ps.CacheReadPerMillion = &v
	}
	if p.CacheCreationPricePerToken > 0 {
		v := p.CacheCreationPricePerToken * perMillion
		ps.CacheCreationPerMillion = &v
	}
	return ps
}

// buildEffectivePriceSet applies a rate multiplier to base pricing.
// Returns all-nil fields (except RateMultiplier) if p is nil.
func buildEffectivePriceSet(p *service.ModelPricing, rate float64) *dto.EffectivePriceSet {
	if p == nil {
		return &dto.EffectivePriceSet{RateMultiplier: rate}
	}
	effInput := p.InputPricePerToken * perMillion * rate
	effOutput := p.OutputPricePerToken * perMillion * rate
	eff := &dto.EffectivePriceSet{
		InputPerMillion:  &effInput,
		OutputPerMillion: &effOutput,
		RateMultiplier:   rate,
	}
	if p.CacheReadPricePerToken > 0 {
		v := p.CacheReadPricePerToken * perMillion * rate
		eff.CacheReadPerMillion = &v
	}
	if p.CacheCreationPricePerToken > 0 {
		v := p.CacheCreationPricePerToken * perMillion * rate
		eff.CacheCreationPerMillion = &v
	}
	return eff
}

// addDefaultModelsForPlatform adds platform-specific default models when no
// model_mapping is configured (C3 fix: matches /v1/models fallback behavior).
func addDefaultModelsForPlatform(modelSet map[string]struct{}, platform string) {
	switch platform {
	case service.PlatformAnthropic:
		for _, id := range claude.DefaultModelIDs() {
			modelSet[id] = struct{}{}
		}
	case service.PlatformOpenAI:
		for _, id := range openai.DefaultModelIDs() {
			modelSet[id] = struct{}{}
		}
	case service.PlatformGemini:
		for _, id := range []string{"gemini-2.5-pro", "gemini-2.5-flash"} {
			modelSet[id] = struct{}{}
		}
	case service.PlatformAntigravity:
		// Antigravity supports Claude + Gemini models (not OpenAI).
		for _, id := range claude.DefaultModelIDs() {
			modelSet[id] = struct{}{}
		}
		for _, id := range []string{"gemini-2.5-pro", "gemini-2.5-flash"} {
			modelSet[id] = struct{}{}
		}
	case service.PlatformSora:
		// Sora models are per-request priced (not token-based); skip for pricing page.
	default:
		// Unknown platform: add Claude defaults as safe fallback.
		for _, id := range claude.DefaultModelIDs() {
			modelSet[id] = struct{}{}
		}
	}
}

// categorizeModel returns a provider category based on model ID prefix.
func categorizeModel(modelID string) string {
	lower := strings.ToLower(modelID)
	switch {
	case strings.HasPrefix(lower, "claude"):
		return "anthropic"
	case strings.HasPrefix(lower, "gpt") ||
		strings.HasPrefix(lower, "o1") ||
		strings.HasPrefix(lower, "o3") ||
		strings.HasPrefix(lower, "o4"):
		return "openai"
	case strings.HasPrefix(lower, "gemini"):
		return "google"
	default:
		return "other"
	}
}
