package routes

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

func TestRegisterPaymentRoutesExposesUpstreamPaymentSurface(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	allow := func(c *gin.Context) { c.Next() }

	handlers := &handler.Handlers{
		Payment:        &handler.PaymentHandler{},
		PaymentWebhook: &handler.PaymentWebhookHandler{},
		Admin: &handler.AdminHandlers{
			Payment: &adminhandler.PaymentHandler{},
		},
	}

	RegisterPaymentRoutes(
		v1,
		handlers,
		middleware.JWTAuthMiddleware(allow),
		middleware.AdminAuthMiddleware(allow),
		nil,
	)

	routes := map[string]bool{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	expected := []string{
		http.MethodGet + " /api/v1/payment/config",
		http.MethodGet + " /api/v1/payment/checkout-info",
		http.MethodGet + " /api/v1/payment/plans",
		http.MethodGet + " /api/v1/payment/channels",
		http.MethodGet + " /api/v1/payment/limits",
		http.MethodPost + " /api/v1/payment/orders",
		http.MethodPost + " /api/v1/payment/orders/verify",
		http.MethodGet + " /api/v1/payment/orders/my",
		http.MethodGet + " /api/v1/payment/orders/:id",
		http.MethodPost + " /api/v1/payment/orders/:id/cancel",
		http.MethodPost + " /api/v1/payment/orders/:id/refund-request",
		http.MethodGet + " /api/v1/payment/orders/refund-eligible-providers",
		http.MethodPost + " /api/v1/payment/public/orders/verify",
		http.MethodPost + " /api/v1/payment/public/orders/resolve",
		http.MethodGet + " /api/v1/payment/webhook/easypay",
		http.MethodPost + " /api/v1/payment/webhook/easypay",
		http.MethodPost + " /api/v1/payment/webhook/alipay",
		http.MethodPost + " /api/v1/payment/webhook/wxpay",
		http.MethodPost + " /api/v1/payment/webhook/stripe",
		http.MethodGet + " /api/v1/admin/payment/dashboard",
		http.MethodGet + " /api/v1/admin/payment/config",
		http.MethodPut + " /api/v1/admin/payment/config",
		http.MethodGet + " /api/v1/admin/payment/orders",
		http.MethodGet + " /api/v1/admin/payment/orders/:id",
		http.MethodPost + " /api/v1/admin/payment/orders/:id/cancel",
		http.MethodPost + " /api/v1/admin/payment/orders/:id/retry",
		http.MethodPost + " /api/v1/admin/payment/orders/:id/refund",
		http.MethodGet + " /api/v1/admin/payment/plans",
		http.MethodPost + " /api/v1/admin/payment/plans",
		http.MethodPut + " /api/v1/admin/payment/plans/:id",
		http.MethodDelete + " /api/v1/admin/payment/plans/:id",
		http.MethodGet + " /api/v1/admin/payment/providers",
		http.MethodPost + " /api/v1/admin/payment/providers",
		http.MethodPut + " /api/v1/admin/payment/providers/:id",
		http.MethodDelete + " /api/v1/admin/payment/providers/:id",
	}

	for _, want := range expected {
		if !routes[want] {
			t.Fatalf("missing payment route %s", want)
		}
	}
}
