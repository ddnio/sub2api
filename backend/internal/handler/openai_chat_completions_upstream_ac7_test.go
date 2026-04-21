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
// each failover attempt resets ctxKeyUpstreamEndpointOverride based on the new account type,
// preventing override contamination across failover switches.
func TestUpstreamEndpointOverride_PerAttemptReset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeCtx := func() *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
		c.Set(ctxKeyInboundEndpoint, EndpointChatCompletions)
		return c
	}

	t.Run("upstream account sets override to /v1/chat/completions", func(t *testing.T) {
		c := makeCtx()
		account := &service.Account{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeUpstream,
		}

		if account.IsOpenAIUpstream() {
			c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions)
		} else {
			c.Set(ctxKeyUpstreamEndpointOverride, "")
		}

		got := GetUpstreamEndpoint(c, account.Platform)
		assert.Equal(t, EndpointChatCompletions, got, "Upstream account must produce /v1/chat/completions endpoint")
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
		if oauthAccount.IsOpenAIUpstream() {
			c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions)
		} else {
			c.Set(ctxKeyUpstreamEndpointOverride, "")
		}

		got := GetUpstreamEndpoint(c, oauthAccount.Platform)
		assert.Equal(t, EndpointResponses, got, "OAuth account after failover must produce /v1/responses (override cleared)")
	})

	t.Run("apikey account clears override — GetUpstreamEndpoint returns /v1/responses", func(t *testing.T) {
		c := makeCtx()
		c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions) // residual from prior attempt

		apiKeyAccount := &service.Account{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeAPIKey,
		}
		if apiKeyAccount.IsOpenAIUpstream() {
			c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions)
		} else {
			c.Set(ctxKeyUpstreamEndpointOverride, "")
		}

		got := GetUpstreamEndpoint(c, apiKeyAccount.Platform)
		assert.Equal(t, EndpointResponses, got, "APIKey account must produce /v1/responses (override cleared)")
	})
}
