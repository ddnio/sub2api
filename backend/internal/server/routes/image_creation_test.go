package routes

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRegisterImageCreationSessionRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	next := func(c *gin.Context) { c.Next() }

	RegisterImageCreationSessionRoutes(
		router.Group("/api/v1"),
		handler.NewImageCreationSessionHandler(nil, nil),
		middleware.JWTAuthMiddleware(next),
		middleware.AdminAuthMiddleware(next),
		nil,
	)

	routes := make(map[string]bool)
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	require.True(t, routes["POST /api/v1/image-creation/sessions"])
	require.True(t, routes["POST /api/v1/image-creation/embed-tickets"])
	require.True(t, routes["POST /api/v1/admin/image-creation/embed-tickets"])
}
