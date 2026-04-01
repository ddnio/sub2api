package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCategorizeModel(t *testing.T) {
	tests := []struct {
		modelID  string
		expected string
	}{
		{"claude-sonnet-4-6", "anthropic"},
		{"claude-opus-4-6", "anthropic"},
		{"claude-haiku-4-5", "anthropic"},
		{"gpt-5.4", "openai"},
		{"gpt-5.2-mini", "openai"},
		{"o4-mini", "openai"},
		{"o3-mini", "openai"},
		{"o1-preview", "openai"},
		{"gemini-2.5-pro", "google"},
		{"gemini-2.5-flash", "google"},
		{"custom-model", "other"},
		{"llama-3", "other"},
	}
	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			result := categorizeModel(tt.modelID)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetModelPricing_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &PricingHandler{}
	r.GET("/api/v1/pricing/models", h.GetModelPricing)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pricing/models", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without auth middleware setting the subject, handler should return 401.
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
