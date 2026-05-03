package dto

// ModelPricingItem represents a single model's pricing information.
type ModelPricingItem struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"display_name"`
	OwnedBy     string   `json:"owned_by"`
	Category    string   `json:"category"`
	Pricing     PriceSet `json:"pricing"`
	// EffectivePricing is non-nil only when group_id is provided.
	EffectivePricing *EffectivePriceSet `json:"effective_pricing"`
}

// PriceSet holds base prices in USD per million tokens.
// nil means "not supported / pricing unavailable" (render as "—").
// Explicit 0 means "free" (render as "$0").
type PriceSet struct {
	InputPerMillion         *float64 `json:"input_per_million"`
	OutputPerMillion        *float64 `json:"output_per_million"`
	CacheReadPerMillion     *float64 `json:"cache_read_per_million"`
	CacheCreationPerMillion *float64 `json:"cache_creation_per_million"`
}

// EffectivePriceSet holds prices after applying group/user rate multiplier.
// nil fields inherit nil from base PriceSet (pricing unavailable).
type EffectivePriceSet struct {
	InputPerMillion         *float64 `json:"input_per_million"`
	OutputPerMillion        *float64 `json:"output_per_million"`
	CacheReadPerMillion     *float64 `json:"cache_read_per_million"`
	CacheCreationPerMillion *float64 `json:"cache_creation_per_million"`
	RateMultiplier          float64  `json:"rate_multiplier"`
}

// GroupInfo holds basic group information for the pricing response.
type GroupInfo struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	RateMultiplier float64 `json:"rate_multiplier"`
}

// ModelPricingResponse is the response for GET /api/v1/pricing/models.
type ModelPricingResponse struct {
	Models []ModelPricingItem `json:"models"`
	Group  *GroupInfo         `json:"group"`
	// Notice reminds users that effective prices are standard-tier estimates.
	Notice string `json:"notice,omitempty"`
}
