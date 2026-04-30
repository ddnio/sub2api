package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterPaymentRoutes registers the upstream payment v2 HTTP surface.
func RegisterPaymentRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	adminAuth middleware.AdminAuthMiddleware,
	settingService *service.SettingService,
) {
	authenticated := v1.Group("/payment")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(middleware.BackendModeUserGuard(settingService))
	{
		authenticated.GET("/config", h.Payment.GetPaymentConfig)
		authenticated.GET("/checkout-info", h.Payment.GetCheckoutInfo)
		authenticated.GET("/plans", h.Payment.GetPlans)
		authenticated.GET("/channels", h.Payment.GetChannels)
		authenticated.GET("/limits", h.Payment.GetLimits)

		orders := authenticated.Group("/orders")
		{
			orders.POST("", h.Payment.CreateOrder)
			orders.POST("/verify", h.Payment.VerifyOrder)
			orders.GET("/my", h.Payment.GetMyOrders)
			orders.GET("/refund-eligible-providers", h.Payment.GetRefundEligibleProviders)
			orders.GET("/:id", h.Payment.GetOrder)
			orders.POST("/:id/cancel", h.Payment.CancelOrder)
			orders.POST("/:id/refund-request", h.Payment.RequestRefund)
		}
	}

	public := v1.Group("/payment/public")
	{
		public.POST("/orders/verify", h.Payment.VerifyOrderPublic)
		public.POST("/orders/resolve", h.Payment.ResolveOrderPublicByResumeToken)
	}

	webhook := v1.Group("/payment/webhook")
	{
		webhook.GET("/easypay", h.PaymentWebhook.EasyPayNotify)
		webhook.POST("/easypay", h.PaymentWebhook.EasyPayNotify)
		webhook.POST("/alipay", h.PaymentWebhook.AlipayNotify)
		webhook.POST("/wxpay", h.PaymentWebhook.WxpayNotify)
		webhook.POST("/stripe", h.PaymentWebhook.StripeWebhook)
	}

	adminGroup := v1.Group("/admin/payment")
	adminGroup.Use(gin.HandlerFunc(adminAuth))
	{
		adminGroup.GET("/dashboard", h.Admin.Payment.GetDashboard)
		adminGroup.GET("/config", h.Admin.Payment.GetConfig)
		adminGroup.PUT("/config", h.Admin.Payment.UpdateConfig)

		adminOrders := adminGroup.Group("/orders")
		{
			adminOrders.GET("", h.Admin.Payment.ListOrders)
			adminOrders.GET("/:id", h.Admin.Payment.GetOrderDetail)
			adminOrders.POST("/:id/cancel", h.Admin.Payment.CancelOrder)
			adminOrders.POST("/:id/retry", h.Admin.Payment.RetryFulfillment)
			adminOrders.POST("/:id/refund", h.Admin.Payment.ProcessRefund)
		}

		plans := adminGroup.Group("/plans")
		{
			plans.GET("", h.Admin.Payment.ListPlans)
			plans.POST("", h.Admin.Payment.CreatePlan)
			plans.PUT("/:id", h.Admin.Payment.UpdatePlan)
			plans.DELETE("/:id", h.Admin.Payment.DeletePlan)
		}

		providers := adminGroup.Group("/providers")
		{
			providers.GET("", h.Admin.Payment.ListProviders)
			providers.POST("", h.Admin.Payment.CreateProvider)
			providers.PUT("/:id", h.Admin.Payment.UpdateProvider)
			providers.DELETE("/:id", h.Admin.Payment.DeleteProvider)
		}
	}
}
