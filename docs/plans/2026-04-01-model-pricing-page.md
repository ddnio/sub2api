# Model Pricing Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `/pricing` page that shows model input/output/cache prices before API calls, with group-based effective pricing.

**Architecture:** New `PricingHandler` in backend serves model prices by combining existing `GetAvailableModels()` + `BillingService.GetModelPricing()` + group rate multipliers. Frontend adds a new `PricingView.vue` page with table layout, group filter, and search. Pure additive — no existing code behavior changes.

**Tech Stack:** Go (gin), Vue 3 (Composition API), Tailwind CSS, vue-i18n

---

## File Structure

### Backend (new files)
- `backend/internal/handler/pricing_handler.go` — HTTP handler for pricing API
- `backend/internal/handler/pricing_handler_test.go` — Handler tests
- `backend/internal/handler/dto/pricing.go` — Pricing DTO types

### Backend (modify)
- `backend/internal/handler/handler.go:38-55` — Add `Pricing` field to `Handlers` struct
- `backend/internal/handler/wire.go:80-99` — Add `pricingHandler` param to `ProvideHandlers`
- `backend/internal/server/routes/user.go:17-97` — Add pricing route group
- `backend/cmd/server/wire_gen.go:82,243` — Wire up `PricingHandler`

### Frontend (new files)
- `frontend/src/views/user/PricingView.vue` — Pricing page component
- `frontend/src/api/pricing.ts` — Pricing API client

### Frontend (modify)
- `frontend/src/router/index.ts:153` — Add `/pricing` route (between `/keys` and `/usage`)
- `frontend/src/components/layout/AppSidebar.vue:489` — Add pricing nav item (between keys and usage)
- `frontend/src/i18n/locales/en.ts` — Add English translations
- `frontend/src/i18n/locales/zh.ts` — Add Chinese translations
- `frontend/src/types/index.ts` — Add pricing types

---

### Task 1: Backend DTO Types

**Files:**
- Create: `backend/internal/handler/dto/pricing.go`

- [ ] **Step 1: Create pricing DTO file**

```go
package dto

// ModelPricingItem represents a single model's pricing information.
type ModelPricingItem struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	OwnedBy     string  `json:"owned_by"`
	Category    string  `json:"category"`
	Pricing     PriceSet `json:"pricing"`
	// EffectivePricing is non-nil only when group_id is provided.
	EffectivePricing *EffectivePriceSet `json:"effective_pricing"`
}

// PriceSet holds base prices in USD per million tokens.
// nil means "not supported" (render as "—"), 0 means "free" (render as "$0").
type PriceSet struct {
	InputPerMillion         float64  `json:"input_per_million"`
	OutputPerMillion        float64  `json:"output_per_million"`
	CacheReadPerMillion     *float64 `json:"cache_read_per_million"`
	CacheCreationPerMillion *float64 `json:"cache_creation_per_million"`
}

// EffectivePriceSet holds prices after applying group/user rate multiplier.
type EffectivePriceSet struct {
	InputPerMillion         float64  `json:"input_per_million"`
	OutputPerMillion        float64  `json:"output_per_million"`
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/handler/dto/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/dto/pricing.go
git commit -m "feat(pricing): add DTO types for model pricing API"
```

---

### Task 2: Backend Pricing Handler

**Files:**
- Create: `backend/internal/handler/pricing_handler.go`

- [ ] **Step 1: Create handler file**

```go
package handler

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	response "github.com/Wei-Shaw/sub2api/internal/handler/dto"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const tokensPerMillion = 1_000_000

// PricingHandler handles model pricing endpoints.
type PricingHandler struct {
	gatewayService *service.GatewayService
	billingService *service.BillingService
	apiKeyService  *service.APIKeyService
}

// NewPricingHandler creates a new PricingHandler.
func NewPricingHandler(
	gatewayService *service.GatewayService,
	billingService *service.BillingService,
	apiKeyService *service.APIKeyService,
) *PricingHandler {
	return &PricingHandler{
		gatewayService: gatewayService,
		billingService: billingService,
		apiKeyService:  apiKeyService,
	}
}

// ListModelPricing returns model pricing information.
// GET /api/v1/pricing/models?group_id=<optional>
func (h *PricingHandler) ListModelPricing(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}

	// Parse optional group_id
	var groupID *int64
	if gidStr := c.Query("group_id"); gidStr != "" {
		gid, err := strconv.ParseInt(gidStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid group_id")
			return
		}
		groupID = &gid
	}

	// Get user's available groups to scope model visibility
	availableGroups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.InternalError(c, "failed to get available groups")
		return
	}

	// If group_id provided, verify user has access
	var selectedGroup *service.Group
	if groupID != nil {
		for i := range availableGroups {
			if availableGroups[i].ID == *groupID {
				selectedGroup = &availableGroups[i]
				break
			}
		}
		if selectedGroup == nil {
			response.BadRequest(c, "group not available")
			return
		}
	}

	// Collect available models
	modelSet := make(map[string]struct{})
	if groupID != nil {
		// Single group: get models for this group
		models := h.gatewayService.GetAvailableModels(c.Request.Context(), groupID, "")
		for _, m := range models {
			modelSet[m] = struct{}{}
		}
		// Fallback: if no model_mapping configured, include default models
		if len(modelSet) == 0 {
			h.addDefaultModels(modelSet)
		}
	} else {
		// No group selected: aggregate models across all user's available groups
		for _, g := range availableGroups {
			gid := g.ID
			models := h.gatewayService.GetAvailableModels(c.Request.Context(), &gid, "")
			for _, m := range models {
				modelSet[m] = struct{}{}
			}
		}
		// Fallback: include defaults if no model_mapping found
		if len(modelSet) == 0 {
			h.addDefaultModels(modelSet)
		}
	}

	// Get effective rate multiplier
	var rateMultiplier float64 = 1.0
	if selectedGroup != nil {
		rateMultiplier = selectedGroup.RateMultiplier
		// Check for user-specific rate override
		if userRate, err := h.apiKeyService.GetUserGroupRate(c.Request.Context(), subject.UserID, selectedGroup.ID); err == nil && userRate != nil {
			rateMultiplier = *userRate
		}
	}

	// Build pricing items
	modelIDs := make([]string, 0, len(modelSet))
	for m := range modelSet {
		modelIDs = append(modelIDs, m)
	}
	sort.Strings(modelIDs)

	items := make([]dto.ModelPricingItem, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		item := h.buildPricingItem(modelID, rateMultiplier, selectedGroup != nil)
		if item != nil {
			items = append(items, *item)
		}
	}

	resp := dto.ModelPricingResponse{
		Models: items,
	}
	if selectedGroup != nil {
		resp.Group = &dto.GroupInfo{
			ID:             selectedGroup.ID,
			Name:           selectedGroup.Name,
			RateMultiplier: rateMultiplier,
		}
		resp.Notice = "Prices are standard-tier estimates. Actual costs may vary with service tier and context length."
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *PricingHandler) buildPricingItem(modelID string, rateMultiplier float64, hasGroup bool) *dto.ModelPricingItem {
	pricing, err := h.billingService.GetModelPricing(modelID)
	if err != nil {
		// Model exists but no pricing data — still show it with nil prices
		return &dto.ModelPricingItem{
			ID:          modelID,
			DisplayName: modelID,
			OwnedBy:     categorizeModel(modelID),
			Category:    categorizeModel(modelID),
			Pricing: dto.PriceSet{
				InputPerMillion:  0,
				OutputPerMillion: 0,
			},
		}
	}

	inputPerM := pricing.InputPricePerToken * tokensPerMillion
	outputPerM := pricing.OutputPricePerToken * tokensPerMillion

	var cacheReadPerM *float64
	if pricing.CacheReadPricePerToken > 0 {
		v := pricing.CacheReadPricePerToken * tokensPerMillion
		cacheReadPerM = &v
	}
	var cacheCreatePerM *float64
	if pricing.CacheCreationPricePerToken > 0 {
		v := pricing.CacheCreationPricePerToken * tokensPerMillion
		cacheCreatePerM = &v
	}

	item := &dto.ModelPricingItem{
		ID:          modelID,
		DisplayName: modelID,
		OwnedBy:     categorizeModel(modelID),
		Category:    categorizeModel(modelID),
		Pricing: dto.PriceSet{
			InputPerMillion:         inputPerM,
			OutputPerMillion:        outputPerM,
			CacheReadPerMillion:     cacheReadPerM,
			CacheCreationPerMillion: cacheCreatePerM,
		},
	}

	if hasGroup {
		effInput := inputPerM * rateMultiplier
		effOutput := outputPerM * rateMultiplier
		eff := &dto.EffectivePriceSet{
			InputPerMillion:  effInput,
			OutputPerMillion: effOutput,
			RateMultiplier:   rateMultiplier,
		}
		if cacheReadPerM != nil {
			v := *cacheReadPerM * rateMultiplier
			eff.CacheReadPerMillion = &v
		}
		if cacheCreatePerM != nil {
			v := *cacheCreatePerM * rateMultiplier
			eff.CacheCreationPerMillion = &v
		}
		item.EffectivePricing = eff
	}

	return item
}

// addDefaultModels adds platform default models when no model_mapping is configured.
// This matches the fallback behavior of /v1/models endpoint.
func (h *PricingHandler) addDefaultModels(modelSet map[string]struct{}) {
	// Add common default models that the gateway supports without explicit mapping
	defaults := []string{
		"claude-opus-4-6", "claude-sonnet-4-6", "claude-haiku-4-5",
		"gpt-5.4", "gpt-5.4-mini", "gpt-5.2", "gpt-5.2-mini",
		"gemini-2.5-pro", "gemini-2.5-flash",
	}
	for _, m := range defaults {
		modelSet[m] = struct{}{}
	}
}

// categorizeModel returns the provider/category for a model ID.
func categorizeModel(modelID string) string {
	lower := strings.ToLower(modelID)
	switch {
	case strings.HasPrefix(lower, "claude"):
		return "anthropic"
	case strings.HasPrefix(lower, "gpt") || strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3") || strings.HasPrefix(lower, "o4"):
		return "openai"
	case strings.HasPrefix(lower, "gemini"):
		return "google"
	default:
		return "other"
	}
}
```

Note: The `response.Unauthorized`, `response.BadRequest`, `response.InternalError` helpers and `middleware2.GetAuthSubjectFromContext` are existing patterns used throughout the codebase. The `service.Group` struct already has `ID`, `Name`, `RateMultiplier` fields (verified from `dto/mappers.go:162-165`).

**Important:** The `apiKeyService.GetUserGroupRate` method may not exist as a public method. During implementation, check if `UserGroupRateRepository.GetByUserAndGroup` is accessible through `APIKeyService` or if you need to inject the repository directly. If not available, fall back to using just `selectedGroup.RateMultiplier` for P1 and add user-specific rate resolution later.

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/handler/`
Expected: compilation errors related to wiring (handler not yet added to Handlers struct) — that's expected, we'll fix in Task 4.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/pricing_handler.go
git commit -m "feat(pricing): add PricingHandler for model pricing API"
```

---

### Task 3: Backend Handler Tests

**Files:**
- Create: `backend/internal/handler/pricing_handler_test.go`

- [ ] **Step 1: Write basic handler test**

```go
package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPricingHandler_CategorizeModel(t *testing.T) {
	tests := []struct {
		modelID  string
		expected string
	}{
		{"claude-sonnet-4-6", "anthropic"},
		{"claude-opus-4-6", "anthropic"},
		{"gpt-5.4", "openai"},
		{"gpt-5.2-mini", "openai"},
		{"o4-mini", "openai"},
		{"gemini-2.5-pro", "google"},
		{"gemini-2.5-flash", "google"},
		{"custom-model", "other"},
	}
	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			result := categorizeModel(tt.modelID)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestPricingHandler_ListModelPricing_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &PricingHandler{}
	r.GET("/api/v1/pricing/models", h.ListModelPricing)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pricing/models", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}
```

- [ ] **Step 2: Run tests**

Run: `cd backend && go test ./internal/handler/ -run TestPricing -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/pricing_handler_test.go
git commit -m "test(pricing): add handler unit tests"
```

---

### Task 4: Wire Up Backend (Handler → Route → DI)

**Files:**
- Modify: `backend/internal/handler/handler.go:38-55`
- Modify: `backend/internal/handler/wire.go:80-99`
- Modify: `backend/internal/server/routes/user.go:17-97`
- Modify: `backend/cmd/server/wire_gen.go`

- [ ] **Step 1: Add Pricing field to Handlers struct**

In `backend/internal/handler/handler.go`, add after line 54 (`PaymentCallback *PaymentCallbackHandler`):

```go
	Pricing         *PricingHandler
```

- [ ] **Step 2: Add pricingHandler to ProvideHandlers**

In `backend/internal/handler/wire.go`, add `pricingHandler *PricingHandler` parameter to `ProvideHandlers` function signature (after `paymentCallbackHandler`), and add `Pricing: pricingHandler,` to the returned `&Handlers{}` struct.

- [ ] **Step 3: Add pricing routes**

In `backend/internal/server/routes/user.go`, add after the groups block (after line 56):

```go
		// 模型定价
		pricing := authenticated.Group("/pricing")
		{
			pricing.GET("/models", h.Pricing.ListModelPricing)
		}
```

- [ ] **Step 4: Wire up in wire_gen.go**

In `backend/cmd/server/wire_gen.go`, add after line 82 (`apiKeyHandler`):

```go
	pricingHandler := handler.NewPricingHandler(gatewayService, billingService, apiKeyService)
```

And update the `handler.ProvideHandlers(...)` call (line 243) to include `pricingHandler` as a parameter.

- [ ] **Step 5: Verify compilation**

Run: `cd backend && go build ./cmd/server/`
Expected: no errors

- [ ] **Step 6: Verify tests pass**

Run: `cd backend && go test ./internal/handler/ -run TestPricing -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/handler.go backend/internal/handler/wire.go backend/internal/server/routes/user.go backend/cmd/server/wire_gen.go
git commit -m "feat(pricing): wire up pricing handler to routes and DI"
```

---

### Task 5: Frontend Types and API Client

**Files:**
- Create: `frontend/src/api/pricing.ts`
- Modify: `frontend/src/types/index.ts`

- [ ] **Step 1: Add TypeScript types**

Add to `frontend/src/types/index.ts`:

```typescript
// Model Pricing
export interface PriceSet {
  input_per_million: number
  output_per_million: number
  cache_read_per_million: number | null
  cache_creation_per_million: number | null
}

export interface EffectivePriceSet extends PriceSet {
  rate_multiplier: number
}

export interface ModelPricingItem {
  id: string
  display_name: string
  owned_by: string
  category: string
  pricing: PriceSet
  effective_pricing: EffectivePriceSet | null
}

export interface PricingGroupInfo {
  id: number
  name: string
  rate_multiplier: number
}

export interface ModelPricingResponse {
  models: ModelPricingItem[]
  group: PricingGroupInfo | null
  notice?: string
}
```

- [ ] **Step 2: Create pricing API client**

```typescript
// frontend/src/api/pricing.ts
import { apiClient } from './client'
import type { ModelPricingResponse } from '@/types'

export async function getModelPricing(groupId?: number): Promise<ModelPricingResponse> {
  const params: Record<string, string> = {}
  if (groupId !== undefined) {
    params.group_id = String(groupId)
  }
  const { data } = await apiClient.get<ModelPricingResponse>('/pricing/models', { params })
  return data
}

export const pricingAPI = {
  getModelPricing,
}

export default pricingAPI
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/types/index.ts frontend/src/api/pricing.ts
git commit -m "feat(pricing): add frontend types and API client"
```

---

### Task 6: Frontend i18n

**Files:**
- Modify: `frontend/src/i18n/locales/en.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`

- [ ] **Step 1: Add English translations**

Add `pricing` section and `nav.pricing` to `en.ts`:

```typescript
// In nav section:
pricing: 'Model Pricing',

// New top-level section:
pricing: {
  title: 'Model Pricing',
  description: 'View input/output prices for available AI models',
  searchPlaceholder: 'Search models...',
  groupFilter: 'Group',
  groupAll: 'All (Base Price)',
  currentRate: 'Current rate',
  unitPerMillion: '$/M tokens',
  unitPerThousand: '$/1K tokens',
  modelName: 'Model',
  input: 'Input',
  output: 'Output',
  cacheRead: 'Cache Read',
  cacheCreation: 'Cache Write',
  effectivePrice: 'Effective',
  basePrice: 'Base',
  details: 'Details',
  provider: 'Provider',
  noModels: 'No models available',
  noPricing: 'Pricing not available',
  notice: 'Prices are standard-tier estimates. Actual costs may vary with service tier and context length.',
  expand: 'Show details',
  collapse: 'Hide details',
},
```

- [ ] **Step 2: Add Chinese translations**

Add `pricing` section and `nav.pricing` to `zh.ts`:

```typescript
// In nav section:
pricing: '模型定价',

// New top-level section:
pricing: {
  title: '模型定价',
  description: '查看各 AI 模型的输入输出价格',
  searchPlaceholder: '搜索模型...',
  groupFilter: '分组',
  groupAll: '全部（基础价）',
  currentRate: '当前倍率',
  unitPerMillion: '$/百万 tokens',
  unitPerThousand: '$/千 tokens',
  modelName: '模型',
  input: '输入',
  output: '输出',
  cacheRead: '缓存读取',
  cacheCreation: '缓存写入',
  effectivePrice: '有效价格',
  basePrice: '基础价',
  details: '详情',
  provider: '提供商',
  noModels: '暂无可用模型',
  noPricing: '暂无价格信息',
  notice: '价格为标准层级估算，实际费用可能因 service tier 和上下文长度有所不同。',
  expand: '展开详情',
  collapse: '收起详情',
},
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts
git commit -m "feat(pricing): add i18n translations for pricing page"
```

---

### Task 7: Frontend Route and Sidebar

**Files:**
- Modify: `frontend/src/router/index.ts:153`
- Modify: `frontend/src/components/layout/AppSidebar.vue:489`

- [ ] **Step 1: Add route**

In `frontend/src/router/index.ts`, add after the `/keys` route block (after line 153):

```typescript
  {
    path: '/pricing',
    name: 'Pricing',
    component: () => import('@/views/user/PricingView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: false,
      title: 'Model Pricing',
      titleKey: 'pricing.title',
      descriptionKey: 'pricing.description'
    }
  },
```

- [ ] **Step 2: Add sidebar nav item**

In `frontend/src/components/layout/AppSidebar.vue`, find `userNavItems` computed (around line 486-490). Add pricing between keys and usage:

```typescript
    { path: '/dashboard', label: t('nav.dashboard'), icon: DashboardIcon },
    { path: '/keys', label: t('nav.apiKeys'), icon: KeyIcon },
    { path: '/pricing', label: t('nav.pricing'), icon: DollarIcon, hideInSimpleMode: true },
    { path: '/usage', label: t('nav.usage'), icon: ChartIcon, hideInSimpleMode: true },
```

Note: Check what icons are already imported in AppSidebar.vue. Use an existing dollar/money icon if available, or the same icon pattern used by the payment nav item. If `DollarIcon` doesn't exist, use `CurrencyDollarIcon` from heroicons or the existing `Icon` component with `name="dollar"`.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/router/index.ts frontend/src/components/layout/AppSidebar.vue
git commit -m "feat(pricing): add route and sidebar navigation"
```

---

### Task 8: Frontend PricingView Page

**Files:**
- Create: `frontend/src/views/user/PricingView.vue`

This is the largest task. The page should follow existing patterns from `UsageView.vue` and `PaymentView.vue`.

- [ ] **Step 1: Create PricingView.vue**

Create `frontend/src/views/user/PricingView.vue` with:

1. **Template structure:**
   - `<AppLayout>` wrapper
   - Page title + description
   - Toolbar: group filter dropdown, search input, unit toggle ($/M vs $/1K)
   - Amber badge showing current rate when group selected
   - Table with columns: Model | Input | Output | Cache Read | expand arrow
   - Models grouped by category (anthropic, openai, google, other) with section headers
   - Expandable rows showing: cache_creation price, provider info
   - When group selected: prices switch to effective values, base prices as gray sublabels
   - Mobile: card layout for sm breakpoint
   - Loading skeleton, empty state
   - Notice text when group selected

2. **Script setup:**
   - `import { ref, computed, onMounted } from 'vue'`
   - `import { useI18n } from 'vue-i18n'`
   - `import pricingAPI from '@/api/pricing'`
   - `import { userGroupsAPI } from '@/api/groups'`
   - `import AppLayout from '@/components/layout/AppLayout.vue'`
   - State: `models`, `groups`, `selectedGroupId`, `searchQuery`, `unit` (million/thousand), `expandedModels`, `loading`
   - Computed: `filteredModels` (search filter), `groupedModels` (by category)
   - Methods: `loadPricing()`, `loadGroups()`, `formatPrice()`, `toggleExpand()`
   - `onMounted`: load groups + load pricing

3. **Key implementation details:**
   - `formatPrice(value, unit)`: if `unit === 'million'` show as-is, if `'thousand'` divide by 1000
   - Null cache prices → show "—" with `text-gray-300 dark:text-gray-600`
   - Group selector: calls `loadPricing(groupId)` on change
   - Search: filters by `model.id` and `model.display_name` (case-insensitive)
   - All Tailwind classes include `dark:` variants
   - Responsive: `hidden lg:table-cell` for cache read column on small screens

The engineer implementing this should reference `UsageView.vue` for the exact styling patterns (`.card`, stats cards, table wrapper, etc.) and `PaymentView.vue` for the dropdown/tab patterns.

- [ ] **Step 2: Verify dev server loads the page**

Run: `pnpm --dir frontend dev`
Navigate to: `http://localhost:5173/pricing`
Expected: Page loads without console errors. If backend is not running, API calls will fail but the page structure should render.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/user/PricingView.vue
git commit -m "feat(pricing): add PricingView page with table, filters, and responsive layout"
```

---

### Task 9: Integration Test

**Files:**
- No new files — manual verification

- [ ] **Step 1: Start backend**

Run: `cd backend && go run ./cmd/server`
Expected: Server starts without errors

- [ ] **Step 2: Start frontend dev server**

Run: `pnpm --dir frontend dev`

- [ ] **Step 3: Verify end-to-end**

1. Login as a user
2. Navigate to `/pricing` via sidebar
3. Verify: table shows available models with prices
4. Select a group from dropdown → prices switch to effective values with sublabels
5. Search for a model → table filters correctly
6. Toggle $/M ↔ $/1K → prices recalculate
7. Click expand on a Claude model → shows cache_creation price
8. Switch to dark mode → verify all elements render correctly
9. Resize to mobile → verify card layout

- [ ] **Step 4: Run backend tests**

Run: `cd backend && go test ./internal/handler/ -run TestPricing -v`
Expected: all PASS

- [ ] **Step 5: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix(pricing): integration fixes"
```

---

## Implementation Notes

1. **Wire gen is manual** — `wire` tool doesn't work with Go 1.26.1, edit `wire_gen.go` by hand (per CLAUDE.md).

2. **`GetUserGroupRate` availability** — If `APIKeyService` doesn't expose this method, either:
   - Inject `UserGroupRateRepository` directly into `PricingHandler`, or
   - For P1, skip user-specific rates and just use `group.RateMultiplier`

3. **Default models fallback** — The `addDefaultModels` list should match what `/v1/models` returns as defaults. Check `gateway_handler.go:861-870` for the exact fallback logic and sync the list.

4. **Model alias pricing** — If `BillingService.GetModelPricing(alias)` returns nil, the handler still shows the model with zero prices. P2 can add model_mapping target resolution.

5. **No DB migration needed** — All data comes from existing services.
