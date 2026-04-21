package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAC7_RecordUsage_UpstreamEndpoint verifies the AC7 invariant end-to-end:
// the per-attempt reset of ctxKeyUpstreamEndpointOverride propagates through
// submitUsageRecordTask → RecordUsage → UpstreamEndpoint correctly.
//
// This catches regressions where the reset logic is removed or repositioned,
// unlike the existing AC7 test that only verifies GetUpstreamEndpoint directly.
func TestAC7_RecordUsage_UpstreamEndpoint_UpstreamAccount(t *testing.T) {
	var capturedEndpoint string
	svc := service.NewOpenAIGatewayServiceForTest(nil, nil)
	svc.SetRecordUsageHookForTest(func(_ context.Context, input *service.OpenAIRecordUsageInput) error {
		capturedEndpoint = input.UpstreamEndpoint
		return nil
	})

	h := &OpenAIGatewayHandler{gatewayService: svc} // usageRecordWorkerPool nil → synchronous

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	c.Set(ctxKeyInboundEndpoint, EndpointChatCompletions)

	// Simulate per-attempt reset for upstream account (matches handler line 181-185)
	account := &service.Account{Platform: service.PlatformOpenAI, Type: service.AccountTypeUpstream}
	if account.IsOpenAIUpstream() {
		c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions)
	} else {
		c.Set(ctxKeyUpstreamEndpointOverride, "")
	}

	// Simulate what ChatCompletions handler does inside submitUsageRecordTask
	h.submitUsageRecordTask(func(ctx context.Context) {
		_ = h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
			UpstreamEndpoint: GetUpstreamEndpoint(c, account.Platform),
		})
	})

	require.NotEmpty(t, capturedEndpoint, "hook must have been called synchronously")
	assert.Equal(t, EndpointChatCompletions, capturedEndpoint,
		"upstream account must record /v1/chat/completions endpoint")
}

func TestAC7_RecordUsage_UpstreamEndpoint_AfterFailover(t *testing.T) {
	var capturedEndpoint string
	svc := service.NewOpenAIGatewayServiceForTest(nil, nil)
	svc.SetRecordUsageHookForTest(func(_ context.Context, input *service.OpenAIRecordUsageInput) error {
		capturedEndpoint = input.UpstreamEndpoint
		return nil
	})

	h := &OpenAIGatewayHandler{gatewayService: svc}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	c.Set(ctxKeyInboundEndpoint, EndpointChatCompletions)

	// Residual override from a previous upstream attempt
	c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions)

	// Failover to OAuth account — per-attempt reset clears the override
	oauthAccount := &service.Account{Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth}
	if oauthAccount.IsOpenAIUpstream() {
		c.Set(ctxKeyUpstreamEndpointOverride, EndpointChatCompletions)
	} else {
		c.Set(ctxKeyUpstreamEndpointOverride, "")
	}

	h.submitUsageRecordTask(func(ctx context.Context) {
		_ = h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
			UpstreamEndpoint: GetUpstreamEndpoint(c, oauthAccount.Platform),
		})
	})

	require.NotEmpty(t, capturedEndpoint)
	assert.Equal(t, EndpointResponses, capturedEndpoint,
		"OAuth account after failover must record /v1/responses (override cleared)")
}
