package handler

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
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
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	info, err := h.referralService.GetReferralInfo(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get referral info"})
		return
	}

	c.JSON(http.StatusOK, info)
}

// ListReferrals 获取当前用户的邀请列表（分页）
func (h *ReferralHandler) ListReferrals(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
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

	records, paginationResult, err := h.referralService.ListReferrals(c.Request.Context(), userID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list referrals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       records,
		"pagination": paginationResult,
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

	// 获取被邀请关系
	referral, err := h.referralService.GetReferralByInvitee(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get referral info"})
		return
	}

	// 获取邀请人数
	inviteCount, err := h.referralService.GetInviteCount(c.Request.Context(), userID)
	if err != nil {
		inviteCount = 0
	}

	// 获取用户的推荐码
	info, err := h.referralService.GetReferralInfo(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get referral info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"referral_code": info.ReferralCode,
		"invite_count":  inviteCount,
		"invited_by":    referral, // nil if not invited by anyone
	})
}
