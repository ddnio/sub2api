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
func (h *PaymentCallbackHandler) Handle(c *gin.Context) {
	err := h.paymentService.ProcessCallback(c.Request.Context(), c.Request)
	if err != nil {
		c.String(http.StatusBadRequest, "FAIL")
		return
	}
	c.String(http.StatusOK, "SUCCESS")
}
