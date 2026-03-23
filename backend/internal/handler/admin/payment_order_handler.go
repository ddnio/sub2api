package admin

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentOrderHandler handles admin payment order management
type PaymentOrderHandler struct {
	paymentService *service.PaymentService
}

// NewPaymentOrderHandler creates a new PaymentOrderHandler
func NewPaymentOrderHandler(paymentService *service.PaymentService) *PaymentOrderHandler {
	return &PaymentOrderHandler{paymentService: paymentService}
}

// List GET /api/v1/admin/payment/orders
func (h *PaymentOrderHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	filter := service.OrderFilter{
		Status: c.Query("status"),
		Type:   c.Query("type"),
	}
	orders, pageResult, err := h.paymentService.ListAllOrders(c.Request.Context(), filter, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.AdminPaymentOrderDTO, len(orders))
	for i := range orders {
		out[i] = dto.AdminPaymentOrderFromService(&orders[i])
	}
	response.Paginated(c, out, pageResult.Total, page, pageSize)
}

// GetByID GET /api/v1/admin/payment/orders/:id
func (h *PaymentOrderHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	order, err := h.paymentService.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminPaymentOrderFromService(order))
}

// Complete POST /api/v1/admin/payment/orders/:id/complete
func (h *PaymentOrderHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	var req struct {
		AdminNote string `json:"admin_note"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.paymentService.AdminCompleteOrder(c.Request.Context(), id, req.AdminNote); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Refund POST /api/v1/admin/payment/orders/:id/refund
func (h *PaymentOrderHandler) Refund(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	var req struct {
		AdminNote string `json:"admin_note"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.paymentService.AdminRefundOrder(c.Request.Context(), id, req.AdminNote); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Stats GET /api/v1/admin/payment/orders/stats
func (h *PaymentOrderHandler) Stats(c *gin.Context) {
	filter := service.StatsFilter{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		GroupBy:   c.Query("group_by"),
	}
	stats, err := h.paymentService.GetOrderStats(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, stats)
}
