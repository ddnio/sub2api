package handler

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler handles user-facing payment requests
type PaymentHandler struct {
	paymentService *service.PaymentService
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// ListPlans GET /api/v1/payment/plans
func (h *PaymentHandler) ListPlans(c *gin.Context) {
	plans, err := h.paymentService.ListActivePlans(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.PaymentPlanDTO, len(plans))
	for i := range plans {
		out[i] = dto.PaymentPlanFromService(&plans[i])
	}
	response.Success(c, out)
}

// CreateOrder POST /api/v1/payment/orders
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Type     string  `json:"type" binding:"required"`
		PlanID   *int64  `json:"plan_id"`
		Amount   float64 `json:"amount"`
		Provider string  `json:"provider" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	input := service.CreateOrderInput{
		UserID:   subject.UserID,
		Type:     req.Type,
		PlanID:   req.PlanID,
		Amount:   req.Amount,
		Provider: req.Provider,
	}
	order, result, err := h.paymentService.CreateOrder(c.Request.Context(), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"order":       dto.PaymentOrderFromService(order),
		"qr_code_url": result.QRCodeURL,
	})
}

// ListOrders GET /api/v1/payment/orders
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	filter := service.OrderFilter{
		Status: c.Query("status"),
		Type:   c.Query("type"),
	}
	orders, pageResult, err := h.paymentService.ListUserOrders(c.Request.Context(), subject.UserID, filter, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.PaymentOrderDTO, len(orders))
	for i := range orders {
		out[i] = dto.PaymentOrderFromService(&orders[i])
	}
	response.Paginated(c, out, pageResult.Total, page, pageSize)
}

// GetOrderStatus GET /api/v1/payment/orders/:id/status
func (h *PaymentHandler) GetOrderStatus(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	status, err := h.paymentService.GetOrderStatus(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"status": status})
}
