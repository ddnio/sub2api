package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ImageCreationSessionHandler struct {
	sessions *service.ImageCreationSessionService
	apiKeys  *service.APIKeyService
}

type imageCreationTicketRequest struct {
	Surface string `json:"surface" binding:"required,oneof=user admin"`
}

type imageCreationSessionRequest struct {
	Ticket string `json:"ticket" binding:"required"`
}

type imageCreationAPIKey struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

func isEligibleImageCreationAPIKey(key *service.APIKey) bool {
	return key.IsActive() &&
		!key.IsExpired() &&
		!key.IsQuotaExhausted() &&
		service.GroupAllowsImageGeneration(key.Group)
}

func NewImageCreationSessionHandler(sessions *service.ImageCreationSessionService, apiKeys *service.APIKeyService) *ImageCreationSessionHandler {
	return &ImageCreationSessionHandler{sessions: sessions, apiKeys: apiKeys}
}

func (h *ImageCreationSessionHandler) IssueUserTicket(c *gin.Context) {
	h.issueTicket(c, service.ImageCreationScopeUser)
}

func (h *ImageCreationSessionHandler) IssueAdminTicket(c *gin.Context) {
	h.issueTicket(c, service.ImageCreationScopeAdmin)
}

func (h *ImageCreationSessionHandler) issueTicket(c *gin.Context, scope string) {
	var req imageCreationTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Surface != scope {
		response.BadRequest(c, "Invalid image creation surface")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	ticket, err := h.sessions.IssueTicket(c.Request.Context(), subject.UserID, scope)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, gin.H{"ticket": ticket, "expires_in": int(service.ImageCreationTicketTTL.Seconds())})
}

func (h *ImageCreationSessionHandler) Exchange(c *gin.Context) {
	var req imageCreationSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid image creation ticket")
		return
	}
	token, viewer, err := h.sessions.ExchangeTicket(c.Request.Context(), req.Ticket)
	if err != nil {
		h.writeError(c, err)
		return
	}
	keys, _, err := h.apiKeys.List(c.Request.Context(), viewer.UserID, pagination.PaginationParams{
		Page: 1, PageSize: 1000, SortBy: "created_at", SortOrder: pagination.SortOrderDesc,
	}, service.APIKeyListFilters{Status: service.StatusAPIKeyActive})
	if err != nil {
		response.InternalError(c, "Failed to load image creation API keys")
		return
	}
	eligible := make([]imageCreationAPIKey, 0, len(keys))
	for i := range keys {
		if !isEligibleImageCreationAPIKey(&keys[i]) {
			continue
		}
		eligible = append(eligible, imageCreationAPIKey{ID: keys[i].ID, Name: keys[i].Name, Key: keys[i].Key})
	}
	response.Success(c, gin.H{
		"session_token": token,
		"expires_in":    int(service.ImageCreationSessionTTL.Seconds()),
		"viewer":        viewer,
		"api_keys":      eligible,
	})
}

func (h *ImageCreationSessionHandler) ScopedAuth(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.SplitN(strings.TrimSpace(c.GetHeader("Authorization")), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			h.writeError(c, service.ErrImageCreationSessionInvalid)
			c.Abort()
			return
		}
		viewer, err := h.sessions.Authenticate(c.Request.Context(), strings.TrimSpace(parts[1]), requiredScope)
		if err != nil {
			h.writeError(c, err)
			c.Abort()
			return
		}
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: viewer.UserID})
		c.Set(string(middleware2.ContextKeyUserRole), viewer.Role)
		c.Set("image_creation_scope", viewer.Scope)
		c.Next()
	}
}

func (h *ImageCreationSessionHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrImageCreationSessionUnavailable):
		response.ErrorWithDetails(c, http.StatusServiceUnavailable, "Image creation session unavailable", "IMAGE_CREATION_SESSION_UNAVAILABLE", nil)
	case errors.Is(err, service.ErrImageCreationAdminRequired):
		response.ErrorWithDetails(c, http.StatusForbidden, "Image creation admin access required", "IMAGE_CREATION_ADMIN_REQUIRED", nil)
	case errors.Is(err, service.ErrImageCreationTicketInvalid), errors.Is(err, service.ErrImageCreationSessionInvalid):
		response.ErrorWithDetails(c, http.StatusUnauthorized, "Invalid image creation session", "IMAGE_CREATION_SESSION_INVALID", nil)
	default:
		response.InternalError(c, "Image creation session failed")
	}
}
