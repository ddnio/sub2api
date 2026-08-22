package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterImageCreationSessionRoutes(
	v1 *gin.RouterGroup,
	sessions *handler.ImageCreationSessionHandler,
	content *handler.ImageCreationHandler,
	jwtAuth middleware.JWTAuthMiddleware,
	adminAuth middleware.AdminAuthMiddleware,
	panelRateLimiter *middleware.PanelRateLimiter,
) {
	imageCreation := v1.Group("/image-creation")
	imageCreation.POST("/sessions", panelRateLimiter.PublicIP(), sessions.Exchange)
	imageCreation.POST("/embed-tickets", gin.HandlerFunc(jwtAuth), panelRateLimiter.Global(), sessions.IssueUserTicket)
	imageCreation.GET("/assets/:id/content", panelRateLimiter.PublicIP(), content.AssetContent)

	imageCreation.Use(sessions.ScopedAuth(service.ImageCreationScopeUser), panelRateLimiter.Global())
	imageCreation.GET("/templates", content.List)
	imageCreation.GET("/templates/:id", content.Get)
	imageCreation.PUT("/templates/:id/favorite", content.Favorite)
	imageCreation.DELETE("/templates/:id/favorite", content.Unfavorite)
	imageCreation.POST("/templates/:id/apply", content.Apply)

	admin := v1.Group("/admin/image-creation")
	admin.POST("/embed-tickets", gin.HandlerFunc(adminAuth), panelRateLimiter.Global(), sessions.IssueAdminTicket)
	admin.Use(sessions.ScopedAuth(service.ImageCreationScopeAdmin), panelRateLimiter.Global())
	admin.GET("/templates", content.AdminList)
	admin.POST("/templates", content.AdminCreate)
	admin.GET("/templates/:id", content.AdminGet)
	admin.PUT("/templates/:id/draft", content.AdminUpdateDraft)
	admin.POST("/templates/:id/publish", content.AdminPublish)
	admin.POST("/templates/:id/archive", content.AdminArchive)
	admin.POST("/templates/:id/restore", content.AdminRestore)
	admin.POST("/assets", content.AdminUploadAsset)
	admin.GET("/home-featured", content.AdminGetHome)
	admin.PUT("/home-featured", content.AdminReplaceHome)
}
