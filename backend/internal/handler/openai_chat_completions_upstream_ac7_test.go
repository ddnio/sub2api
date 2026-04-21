package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestUpstreamEndpointOverride_PerAttemptReset verifies the D2 invariant:
// each failover attempt resets ctxKeyUpstreamEndpointOverride based on the new account's
// ShouldUseDirectChatCompletionsUpstream() verdict, preventing override contamination across
// failover switches.
//
// 每个子 case 都直接调用 account.ShouldUseDirectChatCompletionsUpstream()，
// 与 handler 内的判定保持一致，避免测试与实现漂移。
func TestUpstreamEndpointOverride_PerAttemptReset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeCtx := func() *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
		c.Set(ctxKeyInboundEndpoint, EndpointChatCompletions)
		return c
	}

	// applyOverride 复制 handler 的判定-设值逻辑到测试，验证两者一致。
	applyOverride := func(c *gin.Context, a *service.Account) {
		if a.ShouldUseDirectChatCompletionsUpstream() {
			c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions)
		} else {
			c.Set(ctxKeyUpstreamEndpointOverride, "")
		}
	}

	t.Run("upstream account sets override to /v1/chat/completions", func(t *testing.T) {
		c := makeCtx()
		account := &service.Account{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeUpstream,
		}
		applyOverride(c, account)
		got := GetUpstreamEndpoint(c, account.Platform)
		assert.Equal(t, EndpointChatCompletions, got, "upstream account must produce /v1/chat/completions endpoint")
	})

	t.Run("apikey+passthrough+custom base_url sets override to /v1/chat/completions", func(t *testing.T) {
		c := makeCtx()
		account := &service.Account{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeAPIKey,
			Credentials: map[string]any{
				"api_key":  "sk-test",
				"base_url": "https://api.kimi.com/coding/v1",
			},
			Extra: map[string]any{"openai_passthrough": true},
		}
		applyOverride(c, account)
		got := GetUpstreamEndpoint(c, account.Platform)
		assert.Equal(t, EndpointChatCompletions, got,
			"apikey + passthrough + custom base_url must route to /v1/chat/completions (new branch)")
	})

	t.Run("apikey+passthrough+official base_url does NOT override (stays /v1/responses)", func(t *testing.T) {
		c := makeCtx()
		account := &service.Account{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeAPIKey,
			Credentials: map[string]any{
				"api_key":  "sk-test",
				"base_url": "https://api.openai.com/v1",
			},
			Extra: map[string]any{"openai_passthrough": true},
		}
		applyOverride(c, account)
		got := GetUpstreamEndpoint(c, account.Platform)
		assert.Equal(t, EndpointResponses, got,
			"official openai.com base_url must remain on Responses passthrough (no regression)")
	})

	t.Run("oauth account clears override — GetUpstreamEndpoint returns /v1/responses", func(t *testing.T) {
		c := makeCtx()

		// First attempt: Upstream account sets override
		c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions)

		// Failover: OAuth account clears override (per-attempt reset)
		oauthAccount := &service.Account{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeOAuth,
		}
		applyOverride(c, oauthAccount)
		got := GetUpstreamEndpoint(c, oauthAccount.Platform)
		assert.Equal(t, EndpointResponses, got, "OAuth account after failover must produce /v1/responses (override cleared)")
	})

	t.Run("apikey account without passthrough clears override", func(t *testing.T) {
		c := makeCtx()
		c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions) // residual from prior attempt

		apiKeyAccount := &service.Account{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeAPIKey,
		}
		applyOverride(c, apiKeyAccount)
		got := GetUpstreamEndpoint(c, apiKeyAccount.Platform)
		assert.Equal(t, EndpointResponses, got, "APIKey account (no passthrough) must produce /v1/responses (override cleared)")
	})

	t.Run("failover: chat-upstream → plain apikey clears override", func(t *testing.T) {
		c := makeCtx()

		// Attempt 1: apikey+passthrough+kimi → sets override
		first := &service.Account{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeAPIKey,
			Credentials: map[string]any{
				"api_key":  "sk-a",
				"base_url": "https://api.kimi.com/coding/v1",
			},
			Extra: map[string]any{"openai_passthrough": true},
		}
		applyOverride(c, first)
		assert.Equal(t, EndpointChatCompletions, GetUpstreamEndpoint(c, first.Platform))

		// Attempt 2: plain apikey (no passthrough) → must clear override
		second := &service.Account{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeAPIKey,
		}
		applyOverride(c, second)
		assert.Equal(t, EndpointResponses, GetUpstreamEndpoint(c, second.Platform),
			"plain apikey after chat-upstream must reset override to empty (so responses is used)")
	})
}
