package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/studioauth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDisabledMCPPreservesSignedStudioRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := mcpRouteHandlers()
	h.MCPOAuth = handler.NewMCPOAuthHandler(nil, nil, nil, nil, &config.Config{})
	RegisterMCPRoutes(router.Group("/api/v1"), h,
		middleware.AuditLogMiddleware(mcpRouteNext), middleware.StepUpAuthMiddleware(mcpRouteNext),
		nil, middleware.NewPanelRateLimiter(nil, nil), nil, nil)
	RegisterMCPOAuthRoutes(router, router.Group("/api/v1"), h, nil)
	now := time.Unix(1_800_000_000, 0)
	verifier, err := studioauth.NewVerifier(studioauth.VerifierConfig{
		ClientID: studioauth.ClientID,
		Current:  studioauth.SigningKey{ID: "studio-test-key", Secret: "studio-test-secret-that-is-at-least-32-bytes"},
		Clock:    func() time.Time { return now },
	}, &routeStudioNonceStore{claims: map[string]struct{}{}})
	require.NoError(t, err)
	h.StudioAuth = handler.NewStudioAuthHandler(nil, nil, nil, nil)
	RegisterStudioAuthRoutes(router, h, verifier, 1024)
	for _, path := range []string{
		"/.well-known/oauth-protected-resource", "/.well-known/oauth-protected-resource/api/v1/mcp",
		"/.well-known/oauth-authorization-server", "/api/v1/mcp",
	} {
		method := http.MethodGet
		if path == "/api/v1/mcp" {
			method = http.MethodPost
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
		require.Equal(t, http.StatusServiceUnavailable, w.Code, w.Body.String())
		if path == "/api/v1/mcp" {
			require.Contains(t, w.Body.String(), "MCP_OAUTH_NOT_CONFIGURED")
		} else {
			require.JSONEq(t, `{"error":"server_error","error_description":"MCP is disabled; set MCP_PUBLIC_URL to enable it"}`, w.Body.String())
		}
	}
	path := "/internal/v1/studio-auth/login"
	unsigned := httptest.NewRecorder()
	router.ServeHTTP(unsigned, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{}`)))
	require.Equal(t, http.StatusUnauthorized, unsigned.Code)
	for i, status := range []int{http.StatusBadRequest, http.StatusUnauthorized} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, signedStudioRouteRequest(now, path, "00112233445566778899aabbccddeeff", []byte(`{}`)))
		require.Equal(t, status, w.Code, "signed request/replay %d", i)
	}
}
