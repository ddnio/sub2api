//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/websearch"
	"github.com/stretchr/testify/require"
)

// --- isOnlyWebSearchToolInBody ---

func TestIsOnlyWebSearchToolInBody_WebSearchType(t *testing.T) {
	require.True(t, isOnlyWebSearchToolInBody([]byte(`{"tools":[{"type":"web_search"}]}`)))
}

func TestIsOnlyWebSearchToolInBody_WebSearch2025Type(t *testing.T) {
	require.True(t, isOnlyWebSearchToolInBody([]byte(`{"tools":[{"type":"web_search_20250305"}]}`)))
}

func TestIsOnlyWebSearchToolInBody_GoogleSearchType(t *testing.T) {
	require.True(t, isOnlyWebSearchToolInBody([]byte(`{"tools":[{"type":"google_search"}]}`)))
}

func TestIsOnlyWebSearchToolInBody_NameWebSearch(t *testing.T) {
	require.True(t, isOnlyWebSearchToolInBody([]byte(`{"tools":[{"name":"web_search"}]}`)))
}

func TestIsOnlyWebSearchToolInBody_NameWebSearch2025(t *testing.T) {
	require.True(t, isOnlyWebSearchToolInBody([]byte(`{"tools":[{"name":"web_search_20250305"}]}`)))
}

func TestIsOnlyWebSearchToolInBody_NameGoogleSearch(t *testing.T) {
	require.True(t, isOnlyWebSearchToolInBody([]byte(`{"tools":[{"name":"google_search"}]}`)))
}

func TestIsOnlyWebSearchToolInBody_MultipleTools(t *testing.T) {
	require.False(t, isOnlyWebSearchToolInBody(
		[]byte(`{"tools":[{"type":"web_search"},{"type":"text_editor"}]}`)))
}

func TestIsOnlyWebSearchToolInBody_NoTools(t *testing.T) {
	require.False(t, isOnlyWebSearchToolInBody([]byte(`{"model":"claude-3"}`)))
}

func TestIsOnlyWebSearchToolInBody_EmptyToolsArray(t *testing.T) {
	require.False(t, isOnlyWebSearchToolInBody([]byte(`{"tools":[]}`)))
}

func TestIsOnlyWebSearchToolInBody_NonWebSearchTool(t *testing.T) {
	require.False(t, isOnlyWebSearchToolInBody([]byte(`{"tools":[{"type":"text_editor"}]}`)))
}

func TestIsOnlyWebSearchToolInBody_ToolsNotArray(t *testing.T) {
	require.False(t, isOnlyWebSearchToolInBody([]byte(`{"tools":"web_search"}`)))
}

func TestIsWebSearchEmulationEnabledByChannel_UsesAccountGroup(t *testing.T) {
	repo := makeStandardRepo(Channel{
		ID:       1,
		Status:   StatusActive,
		GroupIDs: []int64{10},
		FeaturesConfig: map[string]any{
			featureKeyWebSearchEmulation: map[string]any{"anthropic": true},
		},
	}, map[int64]string{10: PlatformAnthropic})
	svc := &GatewayService{channelService: newTestChannelService(repo)}
	account := &Account{
		Platform:      PlatformAnthropic,
		Type:          AccountTypeAPIKey,
		AccountGroups: []AccountGroup{{GroupID: 10}},
	}

	require.True(t, svc.isWebSearchEmulationEnabledByChannel(context.Background(), account))
}

func TestIsWebSearchEmulationEnabledByChannel_Disabled(t *testing.T) {
	repo := makeStandardRepo(Channel{
		ID:       1,
		Status:   StatusActive,
		GroupIDs: []int64{10},
		FeaturesConfig: map[string]any{
			featureKeyWebSearchEmulation: map[string]any{"anthropic": false},
		},
	}, map[int64]string{10: PlatformAnthropic})
	svc := &GatewayService{channelService: newTestChannelService(repo)}
	account := &Account{
		Platform:      PlatformAnthropic,
		Type:          AccountTypeAPIKey,
		AccountGroups: []AccountGroup{{GroupID: 10}},
	}

	require.False(t, svc.isWebSearchEmulationEnabledByChannel(context.Background(), account))
}

func TestShouldEmulateWebSearch_AccountDisabledOverridesChannelEnabled(t *testing.T) {
	svc := newGatewayServiceWithWebSearchChannel(t, true)
	account := &Account{
		Platform:      PlatformAnthropic,
		Type:          AccountTypeAPIKey,
		Extra:         map[string]any{featureKeyWebSearchEmulation: WebSearchModeDisabled},
		AccountGroups: []AccountGroup{{GroupID: 10}},
	}

	require.False(t, svc.shouldEmulateWebSearch(context.Background(), account, []byte(`{"tools":[{"type":"web_search"}]}`)))
}

func TestShouldEmulateWebSearch_DefaultFollowsChannel(t *testing.T) {
	svc := newGatewayServiceWithWebSearchChannel(t, true)
	account := &Account{
		Platform:      PlatformAnthropic,
		Type:          AccountTypeAPIKey,
		Extra:         map[string]any{featureKeyWebSearchEmulation: WebSearchModeDefault},
		AccountGroups: []AccountGroup{{GroupID: 10}},
	}

	require.True(t, svc.shouldEmulateWebSearch(context.Background(), account, []byte(`{"tools":[{"type":"web_search"}]}`)))
}

func TestShouldEmulateWebSearch_AccountEnabledForcesOn(t *testing.T) {
	svc := newGatewayServiceWithWebSearchChannel(t, false)
	account := &Account{
		Platform:      PlatformAnthropic,
		Type:          AccountTypeAPIKey,
		Extra:         map[string]any{featureKeyWebSearchEmulation: WebSearchModeEnabled},
		AccountGroups: []AccountGroup{{GroupID: 10}},
	}

	require.True(t, svc.shouldEmulateWebSearch(context.Background(), account, []byte(`{"tools":[{"type":"web_search"}]}`)))
}

func newGatewayServiceWithWebSearchChannel(t *testing.T, channelEnabled bool) *GatewayService {
	t.Helper()
	SetWebSearchManager(websearch.NewManager([]websearch.ProviderConfig{{Type: "brave", APIKey: "test-key"}}, nil))
	webSearchEmulationSF.Forget(sfKeyWebSearchConfig)
	webSearchEmulationCache.Store(&cachedWebSearchEmulationConfig{
		config: &WebSearchEmulationConfig{
			Enabled:   true,
			Providers: []WebSearchProviderConfig{{Type: "brave", APIKey: "test-key"}},
		},
		expiresAt: time.Now().Add(time.Minute).UnixNano(),
	})
	t.Cleanup(func() {
		SetWebSearchManager(nil)
		webSearchEmulationSF.Forget(sfKeyWebSearchConfig)
		webSearchEmulationCache.Store(&cachedWebSearchEmulationConfig{
			config:    &WebSearchEmulationConfig{},
			expiresAt: 0,
		})
	})

	repo := makeStandardRepo(Channel{
		ID:       1,
		Status:   StatusActive,
		GroupIDs: []int64{10},
		FeaturesConfig: map[string]any{
			featureKeyWebSearchEmulation: map[string]any{"anthropic": channelEnabled},
		},
	}, map[int64]string{10: PlatformAnthropic})
	return &GatewayService{
		channelService: newTestChannelService(repo),
		settingService: NewSettingService(&webSearchSettingRepoStub{values: map[string]string{
			SettingKeyWebSearchEmulationConfig: `{"enabled":true,"providers":[{"type":"brave","api_key":"test-key"}]}`,
		}}, nil),
	}
}

type webSearchSettingRepoStub struct {
	values map[string]string
}

func (s *webSearchSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *webSearchSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *webSearchSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *webSearchSettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *webSearchSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *webSearchSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *webSearchSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

// --- extractSearchQueryFromBody ---

func TestExtractSearchQueryFromBody_StringContent(t *testing.T) {
	body := `{"messages":[{"role":"user","content":"what is golang"}]}`
	require.Equal(t, "what is golang", extractSearchQueryFromBody([]byte(body)))
}

func TestExtractSearchQueryFromBody_ArrayContent(t *testing.T) {
	body := `{"messages":[{"role":"user","content":[{"type":"text","text":"search this"}]}]}`
	require.Equal(t, "search this", extractSearchQueryFromBody([]byte(body)))
}

func TestExtractSearchQueryFromBody_MultipleMessages(t *testing.T) {
	body := `{"messages":[{"role":"user","content":"first"},{"role":"assistant","content":"ok"},{"role":"user","content":"second"}]}`
	require.Equal(t, "second", extractSearchQueryFromBody([]byte(body)))
}

func TestExtractSearchQueryFromBody_LastMessageNotUser(t *testing.T) {
	body := `{"messages":[{"role":"user","content":"q"},{"role":"assistant","content":"a"}]}`
	require.Equal(t, "", extractSearchQueryFromBody([]byte(body)))
}

func TestExtractSearchQueryFromBody_EmptyMessages(t *testing.T) {
	require.Equal(t, "", extractSearchQueryFromBody([]byte(`{"messages":[]}`)))
}

func TestExtractSearchQueryFromBody_NoMessages(t *testing.T) {
	require.Equal(t, "", extractSearchQueryFromBody([]byte(`{"model":"claude-3"}`)))
}

func TestExtractSearchQueryFromBody_ArrayContentSkipsEmptyText(t *testing.T) {
	body := `{"messages":[{"role":"user","content":[{"type":"image"},{"type":"text","text":""},{"type":"text","text":"real query"}]}]}`
	require.Equal(t, "real query", extractSearchQueryFromBody([]byte(body)))
}

func TestExtractSearchQueryFromBody_ArrayContentNoTextBlock(t *testing.T) {
	body := `{"messages":[{"role":"user","content":[{"type":"image","source":{}}]}]}`
	require.Equal(t, "", extractSearchQueryFromBody([]byte(body)))
}

// --- buildSearchResultBlocks ---

func TestBuildSearchResultBlocks_WithResults(t *testing.T) {
	results := []websearch.SearchResult{
		{URL: "https://a.com", Title: "A", Snippet: "snippet a", PageAge: "2 days"},
		{URL: "https://b.com", Title: "B", Snippet: "snippet b"},
	}
	blocks := buildSearchResultBlocks(results)
	require.Len(t, blocks, 2)
	require.Equal(t, "web_search_result", blocks[0]["type"])
	require.Equal(t, "https://a.com", blocks[0]["url"])
	require.Equal(t, "snippet a", blocks[0]["page_content"])
	require.Equal(t, "2 days", blocks[0]["page_age"])
	// Second result has no PageAge
	require.Equal(t, "https://b.com", blocks[1]["url"])
	_, hasPageAge := blocks[1]["page_age"]
	require.False(t, hasPageAge)
}

func TestBuildSearchResultBlocks_Empty(t *testing.T) {
	blocks := buildSearchResultBlocks(nil)
	require.Empty(t, blocks)
}

func TestBuildSearchResultBlocks_SnippetEmpty(t *testing.T) {
	blocks := buildSearchResultBlocks([]websearch.SearchResult{{URL: "https://x.com", Title: "X", Snippet: ""}})
	_, hasContent := blocks[0]["page_content"]
	require.False(t, hasContent)
}

// --- buildTextSummary ---

func TestBuildTextSummary_WithResults(t *testing.T) {
	results := []websearch.SearchResult{
		{URL: "https://a.com", Title: "A", Snippet: "desc a"},
	}
	summary := buildTextSummary("test query", results)
	require.Contains(t, summary, "test query")
	require.Contains(t, summary, "1. **A**")
	require.Contains(t, summary, "https://a.com")
}

func TestBuildTextSummary_NoResults(t *testing.T) {
	summary := buildTextSummary("test", nil)
	require.Contains(t, summary, "No search results found for: test")
}
