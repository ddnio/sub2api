//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

// --- validateWebSearchConfig ---

func TestValidateWebSearchConfig_Nil(t *testing.T) {
	require.NoError(t, validateWebSearchConfig(nil))
}

func TestValidateWebSearchConfig_Valid(t *testing.T) {
	quotaLimit := int64(1000)
	quotaLimit2 := int64(500)
	cfg := &WebSearchEmulationConfig{
		Enabled: true,
		Providers: []WebSearchProviderConfig{
			{Type: "brave", QuotaLimit: &quotaLimit},
			{Type: "tavily", QuotaLimit: &quotaLimit2},
		},
	}
	require.NoError(t, validateWebSearchConfig(cfg))
}

func TestValidateWebSearchConfig_TooManyProviders(t *testing.T) {
	cfg := &WebSearchEmulationConfig{Providers: make([]WebSearchProviderConfig, 11)}
	for i := range cfg.Providers {
		cfg.Providers[i] = WebSearchProviderConfig{Type: "brave"}
	}
	err := validateWebSearchConfig(cfg)
	require.ErrorContains(t, err, "too many providers")
}

func TestValidateWebSearchConfig_InvalidType(t *testing.T) {
	cfg := &WebSearchEmulationConfig{
		Providers: []WebSearchProviderConfig{{Type: "bing"}},
	}
	require.ErrorContains(t, validateWebSearchConfig(cfg), "invalid type")
}

func TestValidateWebSearchConfig_NegativeQuotaLimit(t *testing.T) {
	quotaLimit := int64(-1)
	cfg := &WebSearchEmulationConfig{
		Providers: []WebSearchProviderConfig{{Type: "brave", QuotaLimit: &quotaLimit}},
	}
	require.ErrorContains(t, validateWebSearchConfig(cfg), "quota_limit must be >= 0 or null")
}

func TestValidateWebSearchConfig_DuplicateType(t *testing.T) {
	cfg := &WebSearchEmulationConfig{
		Providers: []WebSearchProviderConfig{
			{Type: "brave"},
			{Type: "brave"},
		},
	}
	require.ErrorContains(t, validateWebSearchConfig(cfg), "duplicate type")
}

func TestValidateWebSearchConfig_ZeroQuotaLimitAllowedForLegacyUnlimited(t *testing.T) {
	quotaLimit := int64(0)
	cfg := &WebSearchEmulationConfig{
		Providers: []WebSearchProviderConfig{{Type: "brave", QuotaLimit: &quotaLimit}},
	}
	require.NoError(t, validateWebSearchConfig(cfg))
}

func TestValidateWebSearchConfig_NilQuotaLimitUnlimited(t *testing.T) {
	cfg := &WebSearchEmulationConfig{
		Providers: []WebSearchProviderConfig{{Type: "brave", QuotaLimit: nil}},
	}
	require.NoError(t, validateWebSearchConfig(cfg))
}

// --- parseWebSearchConfigJSON ---

func TestParseWebSearchConfigJSON_ValidJSON(t *testing.T) {
	raw := `{"enabled":true,"providers":[{"type":"brave","api_key":"sk-xxx"}]}`
	cfg := parseWebSearchConfigJSON(raw)
	require.True(t, cfg.Enabled)
	require.Len(t, cfg.Providers, 1)
	require.Equal(t, "brave", cfg.Providers[0].Type)
}

func TestParseWebSearchConfigJSON_EmptyString(t *testing.T) {
	cfg := parseWebSearchConfigJSON("")
	require.False(t, cfg.Enabled)
	require.Empty(t, cfg.Providers)
}

func TestParseWebSearchConfigJSON_InvalidJSON(t *testing.T) {
	cfg := parseWebSearchConfigJSON("not{json")
	require.False(t, cfg.Enabled)
	require.Empty(t, cfg.Providers)
}

func TestParseWebSearchConfigJSON_BackwardCompatibility(t *testing.T) {
	raw := `{"enabled":true,"providers":[{"type":"brave","priority":1,"quota_refresh_interval":"monthly","quota_limit":1000}]}`
	cfg := parseWebSearchConfigJSON(raw)
	require.True(t, cfg.Enabled)
	require.Len(t, cfg.Providers, 1)
	require.NotNil(t, cfg.Providers[0].QuotaLimit)
	require.Equal(t, int64(1000), *cfg.Providers[0].QuotaLimit)
}

func TestParseWebSearchConfigJSON_NullQuotaLimit(t *testing.T) {
	raw := `{"enabled":true,"providers":[{"type":"brave","quota_limit":null}]}`
	cfg := parseWebSearchConfigJSON(raw)
	require.True(t, cfg.Enabled)
	require.Len(t, cfg.Providers, 1)
	require.Nil(t, cfg.Providers[0].QuotaLimit)
}

func TestWebSearchProviderConfig_RejectsStringQuotaLimit(t *testing.T) {
	var cfg WebSearchEmulationConfig
	err := json.Unmarshal([]byte(`{"providers":[{"type":"brave","quota_limit":""}]}`), &cfg)
	require.Error(t, err)
}

// --- SanitizeWebSearchConfig ---

func TestSanitizeWebSearchConfig_MaskAPIKey(t *testing.T) {
	cfg := &WebSearchEmulationConfig{
		Enabled: true,
		Providers: []WebSearchProviderConfig{
			{Type: "brave", APIKey: "sk-secret-xxx"},
		},
	}
	out := SanitizeWebSearchConfig(context.Background(), cfg)
	require.Equal(t, "", out.Providers[0].APIKey)
	require.True(t, out.Providers[0].APIKeyConfigured)
}

func TestSanitizeWebSearchConfig_NoAPIKey(t *testing.T) {
	cfg := &WebSearchEmulationConfig{
		Providers: []WebSearchProviderConfig{{Type: "brave", APIKey: ""}},
	}
	out := SanitizeWebSearchConfig(context.Background(), cfg)
	require.Equal(t, "", out.Providers[0].APIKey)
	require.False(t, out.Providers[0].APIKeyConfigured)
}

func TestSanitizeWebSearchConfig_Nil(t *testing.T) {
	require.Nil(t, SanitizeWebSearchConfig(context.Background(), nil))
}

func TestSanitizeWebSearchConfig_PreservesOtherFields(t *testing.T) {
	quotaLimit := int64(1000)
	cfg := &WebSearchEmulationConfig{
		Enabled: true,
		Providers: []WebSearchProviderConfig{
			{Type: "brave", APIKey: "secret", QuotaLimit: &quotaLimit},
		},
	}
	out := SanitizeWebSearchConfig(context.Background(), cfg)
	require.True(t, out.Enabled)
	require.NotNil(t, out.Providers[0].QuotaLimit)
	require.Equal(t, int64(1000), *out.Providers[0].QuotaLimit)
}

func TestSanitizeWebSearchConfig_DoesNotMutateOriginal(t *testing.T) {
	cfg := &WebSearchEmulationConfig{
		Providers: []WebSearchProviderConfig{{Type: "brave", APIKey: "secret"}},
	}
	_ = SanitizeWebSearchConfig(context.Background(), cfg)
	require.Equal(t, "secret", cfg.Providers[0].APIKey)
}

func TestResolveWebSearchProviderProxyURL_ActiveProxy(t *testing.T) {
	proxyID := int64(10)
	svc := &SettingService{webSearchProxyRepo: &webSearchProxyRepoStub{
		proxies: map[int64]*Proxy{
			proxyID: {
				ID:       proxyID,
				Protocol: "http",
				Host:     "proxy.example.com",
				Port:     8080,
				Username: "user",
				Password: "pass",
				Status:   StatusActive,
			},
		},
	}}

	require.Equal(t, "http://user:pass@proxy.example.com:8080", svc.resolveWebSearchProviderProxyURL(context.Background(), &proxyID))
}

func TestResolveWebSearchProviderProxyURL_DisabledProxy(t *testing.T) {
	proxyID := int64(10)
	svc := &SettingService{webSearchProxyRepo: &webSearchProxyRepoStub{
		proxies: map[int64]*Proxy{
			proxyID: {ID: proxyID, Protocol: "http", Host: "proxy.example.com", Port: 8080, Status: StatusDisabled},
		},
	}}

	require.Equal(t, "", svc.resolveWebSearchProviderProxyURL(context.Background(), &proxyID))
}

func TestResolveWebSearchProviderProxyURL_NoRepo(t *testing.T) {
	proxyID := int64(10)
	svc := &SettingService{}

	require.Equal(t, "", svc.resolveWebSearchProviderProxyURL(context.Background(), &proxyID))
}

func TestWebSearchProxyIDValue(t *testing.T) {
	proxyID := int64(10)
	require.Equal(t, int64(10), webSearchProxyIDValue(&proxyID))
	require.Equal(t, int64(0), webSearchProxyIDValue(nil))
}

func TestWebSearchQuotaLimitValue(t *testing.T) {
	quotaLimit := int64(10)
	require.Equal(t, int64(10), webSearchQuotaLimitValue(&quotaLimit))
	require.Equal(t, int64(0), webSearchQuotaLimitValue(nil))
}

func TestResetWebSearchUsage_NoManager(t *testing.T) {
	SetWebSearchManager(nil)
	require.ErrorContains(t, ResetWebSearchUsage(context.Background(), "brave"), "manager not initialized")
}

func TestSaveWebSearchEmulationConfig_RejectsEnabledProviderWithoutAPIKey(t *testing.T) {
	repo := newMockSettingRepo()
	svc := NewSettingService(repo, nil)

	err := svc.SaveWebSearchEmulationConfig(context.Background(), &WebSearchEmulationConfig{
		Enabled:   true,
		Providers: []WebSearchProviderConfig{{Type: "brave"}},
	})

	require.Error(t, err)
	require.Equal(t, "MISSING_API_KEY", infraerrors.Reason(err))
}

func TestSaveWebSearchEmulationConfig_MergesExistingAPIKeyBeforeValidation(t *testing.T) {
	repo := newMockSettingRepo()
	existing := `{"enabled":true,"providers":[{"type":"brave","api_key":"saved-key"}]}`
	require.NoError(t, repo.Set(context.Background(), SettingKeyWebSearchEmulationConfig, existing))
	svc := NewSettingService(repo, nil)

	err := svc.SaveWebSearchEmulationConfig(context.Background(), &WebSearchEmulationConfig{
		Enabled:   true,
		Providers: []WebSearchProviderConfig{{Type: "brave"}},
	})

	require.NoError(t, err)
	saved, err := repo.GetValue(context.Background(), SettingKeyWebSearchEmulationConfig)
	require.NoError(t, err)
	var cfg WebSearchEmulationConfig
	require.NoError(t, json.Unmarshal([]byte(saved), &cfg))
	require.Equal(t, "saved-key", cfg.Providers[0].APIKey)
}

type webSearchProxyRepoStub struct {
	proxies map[int64]*Proxy
}

func (s *webSearchProxyRepoStub) Create(ctx context.Context, proxy *Proxy) error {
	return nil
}

func (s *webSearchProxyRepoStub) GetByID(ctx context.Context, id int64) (*Proxy, error) {
	if proxy, ok := s.proxies[id]; ok {
		return proxy, nil
	}
	return nil, ErrProxyNotFound
}

func (s *webSearchProxyRepoStub) ListByIDs(ctx context.Context, ids []int64) ([]Proxy, error) {
	return nil, nil
}

func (s *webSearchProxyRepoStub) Update(ctx context.Context, proxy *Proxy) error {
	return nil
}

func (s *webSearchProxyRepoStub) Delete(ctx context.Context, id int64) error {
	return nil
}

func (s *webSearchProxyRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]Proxy, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (s *webSearchProxyRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]Proxy, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (s *webSearchProxyRepoStub) ListWithFiltersAndAccountCount(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]ProxyWithAccountCount, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (s *webSearchProxyRepoStub) ListActive(ctx context.Context) ([]Proxy, error) {
	return nil, nil
}

func (s *webSearchProxyRepoStub) ListActiveWithAccountCount(ctx context.Context) ([]ProxyWithAccountCount, error) {
	return nil, nil
}

func (s *webSearchProxyRepoStub) ExistsByHostPortAuth(ctx context.Context, host string, port int, username, password string) (bool, error) {
	return false, nil
}

func (s *webSearchProxyRepoStub) CountAccountsByProxyID(ctx context.Context, proxyID int64) (int64, error) {
	return 0, nil
}

func (s *webSearchProxyRepoStub) ListAccountSummariesByProxyID(ctx context.Context, proxyID int64) ([]ProxyAccountSummary, error) {
	return nil, nil
}
