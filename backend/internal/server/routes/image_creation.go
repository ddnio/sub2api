package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterImageCreationSessionRoutes(
	v1 *gin.RouterGroup,
	h *handler.ImageCreationSessionHandler,
	jwtAuth middleware.JWTAuthMiddleware,
	adminAuth middleware.AdminAuthMiddleware,
	panelRateLimiter *middleware.PanelRateLimiter,
) {
	imageCreation := v1.Group("/image-creation")
	imageCreation.POST("/sessions", panelRateLimiter.PublicIP(), h.Exchange)
	imageCreation.POST("/embed-tickets", gin.HandlerFunc(jwtAuth), panelRateLimiter.Global(), h.IssueUserTicket)

	admin := v1.Group("/admin/image-creation")
	admin.POST("/embed-tickets", gin.HandlerFunc(adminAuth), panelRateLimiter.Global(), h.IssueAdminTicket)
}
