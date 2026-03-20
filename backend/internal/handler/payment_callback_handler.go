package handler

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentCallbackHandler handles payment provider callbacks
type PaymentCallbackHandler struct {
	paymentService *service.PaymentService
}

// NewPaymentCallbackHandler creates a new PaymentCallbackHandler
func NewPaymentCallbackHandler(paymentService *service.PaymentService) *PaymentCallbackHandler {
	return &PaymentCallbackHandler{paymentService: paymentService}
}

// Handle POST /api/v1/payment/callback/:provider
// wxpay v3 requires JSON response: {"code":"SUCCESS","message":"ok"} / {"code":"FAIL","message":"..."}
// easypay expects plain text: "SUCCESS" / "FAIL"
func (h *PaymentCallbackHandler) Handle(c *gin.Context) {
	provider := c.Param("provider")
	err := h.paymentService.ProcessCallback(c.Request.Context(), c.Request)
	if err != nil {
		if provider == "wxpay" {
			c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": err.Error()})
		} else {
			c.String(http.StatusBadRequest, "FAIL")
		}
		return
	}
	if provider == "wxpay" {
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "ok"})
	} else {
		c.String(http.StatusOK, "SUCCESS")
	}
}
