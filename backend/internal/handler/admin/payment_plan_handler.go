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

// PaymentPlanHandler handles admin payment plan management
type PaymentPlanHandler struct {
	paymentService *service.PaymentService
}

// NewPaymentPlanHandler creates a new PaymentPlanHandler
func NewPaymentPlanHandler(paymentService *service.PaymentService) *PaymentPlanHandler {
	return &PaymentPlanHandler{paymentService: paymentService}
}

// List GET /api/v1/admin/payment/plans
func (h *PaymentPlanHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	plans, pageResult, err := h.paymentService.ListAllPlans(c.Request.Context(), params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.AdminPaymentPlanDTO, len(plans))
	for i := range plans {
		out[i] = dto.AdminPaymentPlanFromService(&plans[i])
	}
	response.Paginated(c, out, pageResult.Total, page, pageSize)
}

// Create POST /api/v1/admin/payment/plans
func (h *PaymentPlanHandler) Create(c *gin.Context) {
	var req struct {
		Name          string   `json:"name" binding:"required"`
		Description   string   `json:"description"`
		Badge         *string  `json:"badge"`
		GroupID       int64    `json:"group_id" binding:"required"`
		DurationDays  int      `json:"duration_days" binding:"required,min=1"`
		Price         float64  `json:"price" binding:"min=0"`
		OriginalPrice *float64 `json:"original_price"`
		SortOrder     int      `json:"sort_order"`
		IsActive      *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	plan := &service.PaymentPlan{
		Name:          req.Name,
		Description:   req.Description,
		Badge:         req.Badge,
		GroupID:       req.GroupID,
		DurationDays:  req.DurationDays,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		SortOrder:     req.SortOrder,
		IsActive:      isActive,
	}
	if err := h.paymentService.CreatePlan(c.Request.Context(), plan); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.AdminPaymentPlanFromService(plan))
}

// Update PUT /api/v1/admin/payment/plans/:id
func (h *PaymentPlanHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	updated, err := h.paymentService.UpdatePlan(c.Request.Context(), id, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminPaymentPlanFromService(updated))
}

// Delete DELETE /api/v1/admin/payment/plans/:id
func (h *PaymentPlanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}
	if err := h.paymentService.DeletePlan(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
