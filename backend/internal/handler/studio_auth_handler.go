package handler

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type studioAuthService interface {
	RegisterWithoutTokenWithVerification(ctx context.Context, email, password, verifyCode, promoCode, invitationCode, affiliateCode string) (*service.User, error)
	SendVerifyCodeAsync(ctx context.Context, email string, locale ...string) (*service.SendVerifyCodeResult, error)
	AuthenticatePassword(ctx context.Context, email, password string) (*service.User, error)
	RecordSuccessfulLogin(ctx context.Context, userID int64)
}

type studioUserService interface {
	GetByID(ctx context.Context, id int64) (*service.User, error)
	GetByEmail(ctx context.Context, email string) (*service.User, error)
}

type studioSettingService interface {
	IsTotpEnabled(ctx context.Context) bool
}

type studioTotpService interface {
	CreateLoginSessionForAudience(ctx context.Context, userID int64, email, audience string) (string, error)
	GetLoginSession(ctx context.Context, tempToken string) (*service.TotpLoginSession, error)
	VerifyCode(ctx context.Context, userID int64, code string) error
	DeleteLoginSession(ctx context.Context, tempToken string) error
}

type StudioAuthHandler struct {
	authService    studioAuthService
	userService    studioUserService
	settingService studioSettingService
	totpService    studioTotpService
}

type studioRegisterRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	VerifyCode     string `json:"verify_code"`
	PromoCode      string `json:"promo_code"`
	InvitationCode string `json:"invitation_code"`
	AffiliateCode  string `json:"aff_code"`
}

type studioLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type studioLogin2FARequest struct {
	TempToken string `json:"temp_token" binding:"required"`
	TotpCode  string `json:"totp_code" binding:"required,len=6"`
}

type studioSendVerifyCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type studioResolveRequest struct {
	Subject string `json:"subject" binding:"required,max=128"`
	Email   string `json:"email" binding:"required,email"`
}

type studioIdentity struct {
	Subject     string `json:"subject"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name,omitempty"`
	Role        string `json:"role"`
}

func NewStudioAuthHandler(authService *service.AuthService, userService *service.UserService, settingService *service.SettingService, totpService *service.TotpService) *StudioAuthHandler {
	return &StudioAuthHandler{
		authService:    authService,
		userService:    userService,
		settingService: settingService,
		totpService:    totpService,
	}
}

func newStudioAuthHandler(authService studioAuthService, userService studioUserService, settingService studioSettingService, totpService studioTotpService) *StudioAuthHandler {
	return &StudioAuthHandler{
		authService:    authService,
		userService:    userService,
		settingService: settingService,
		totpService:    totpService,
	}
}

func (h *StudioAuthHandler) SendVerifyCode(c *gin.Context) {
	var req studioSendVerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := h.authService.SendVerifyCodeAsync(c.Request.Context(), req.Email, c.GetHeader("Accept-Language"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"countdown": result.Countdown})
}

func (h *StudioAuthHandler) Register(c *gin.Context) {
	var req studioRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	user, err := h.authService.RegisterWithoutTokenWithVerification(
		c.Request.Context(),
		req.Email,
		req.Password,
		req.VerifyCode,
		req.PromoCode,
		req.InvitationCode,
		req.AffiliateCode,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	h.respondWithUser(c, user)
}

func (h *StudioAuthHandler) Login(c *gin.Context) {
	var req studioLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	user, err := h.authService.AuthenticatePassword(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if h.totpService != nil && h.settingService != nil && h.settingService.IsTotpEnabled(c.Request.Context()) && user.TotpEnabled {
		tempToken, err := h.totpService.CreateLoginSessionForAudience(
			c.Request.Context(),
			user.ID,
			user.Email,
			service.TotpLoginAudienceStudio,
		)
		if err != nil {
			response.InternalError(c, "Failed to create 2FA session")
			return
		}
		response.Success(c, gin.H{
			"requires_2fa":      true,
			"temp_token":        tempToken,
			"user_email_masked": service.MaskEmail(user.Email),
		})
		return
	}

	h.authService.RecordSuccessfulLogin(c.Request.Context(), user.ID)
	h.respondWithUser(c, user)
}

func (h *StudioAuthHandler) Login2FA(c *gin.Context) {
	var req studioLogin2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	session, err := h.totpService.GetLoginSession(c.Request.Context(), req.TempToken)
	if err != nil || session == nil || session.Audience != service.TotpLoginAudienceStudio {
		response.BadRequest(c, "Invalid or expired 2FA session")
		return
	}
	if err := h.totpService.VerifyCode(c.Request.Context(), session.UserID, req.TotpCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), session.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := ensureLoginUserActive(user); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.totpService.DeleteLoginSession(c.Request.Context(), req.TempToken); err != nil {
		response.InternalError(c, "Failed to complete 2FA session")
		return
	}

	h.authService.RecordSuccessfulLogin(c.Request.Context(), user.ID)
	h.respondWithUser(c, user)
}

func (h *StudioAuthHandler) Resolve(c *gin.Context) {
	var req studioResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	user, err := h.userService.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := ensureLoginUserActive(user); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if user.PublicID != req.Subject {
		response.Forbidden(c, "Identity does not match")
		return
	}
	h.respondWithUser(c, user)
}

func (h *StudioAuthHandler) respondWithUser(c *gin.Context, user *service.User) {
	if user == nil || user.PublicID == "" {
		response.InternalError(c, "User identity is unavailable")
		return
	}
	response.Success(c, gin.H{
		"user": studioIdentity{
			Subject:     user.PublicID,
			Email:       user.Email,
			DisplayName: user.Username,
			Role:        user.Role,
		},
	})
}
