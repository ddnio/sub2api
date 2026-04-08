package handler

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ReferralHandler 推荐码 handler
type ReferralHandler struct {
	referralService *service.ReferralService
}

// NewReferralHandler 创建推荐码 handler
func NewReferralHandler(referralService *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{referralService: referralService}
}

// GetReferralInfo 获取当前用户的推荐码信息
func (h *ReferralHandler) GetReferralInfo(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	info, err := h.referralService.GetReferralInfo(c.Request.Context(), subject.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get referral info"})
		return
	}

	c.JSON(http.StatusOK, info)
}

// ListReferrals 获取当前用户的邀请列表（分页）
func (h *ReferralHandler) ListReferrals(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	params := pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	records, paginationResult, err := h.referralService.ListReferrals(c.Request.Context(), subject.UserID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list referrals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": records,
		"pagination": gin.H{
			"total":     paginationResult.Total,
			"page":      paginationResult.Page,
			"page_size": paginationResult.PageSize,
			"pages":     paginationResult.Pages,
		},
	})
}

// GetUserReferralInfo 管理员查看用户邀请信息
func (h *ReferralHandler) GetUserReferralInfo(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	referral, err := h.referralService.GetReferralByInvitee(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get referral info"})
		return
	}

	inviteCount, err := h.referralService.GetInviteCount(c.Request.Context(), userID)
	if err != nil {
		inviteCount = 0
	}

	info, err := h.referralService.GetReferralInfo(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get referral info"})
		return
	}

	// 获取邀请列表
	records, _, err := h.referralService.ListReferrals(c.Request.Context(), userID, pagination.PaginationParams{Page: 1, PageSize: 100})
	if err != nil {
		records = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"referral_code":   info.ReferralCode,
		"invite_count":    inviteCount,
		"total_rewarded":  info.TotalRewarded,
		"invited_by":      referral,
		"invite_records":  records,
	})
}
