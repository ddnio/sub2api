package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/studioauth"
	"github.com/gin-gonic/gin"
)

func RegisterStudioAuthRoutes(r *gin.Engine, handlers *handler.Handlers, verifier *studioauth.Verifier, maxBodyBytes int64) {
	if r == nil || handlers == nil || handlers.StudioAuth == nil || verifier == nil {
		return
	}

	studio := r.Group("/internal/v1/studio-auth")
	studio.Use(servermiddleware.StudioAuth(verifier, maxBodyBytes))
	studio.POST("/send-verify-code", handlers.StudioAuth.SendVerifyCode)
	studio.POST("/register", handlers.StudioAuth.Register)
	studio.POST("/login", handlers.StudioAuth.Login)
	studio.POST("/login/2fa", handlers.StudioAuth.Login2FA)
	studio.POST("/resolve", handlers.StudioAuth.Resolve)
}
