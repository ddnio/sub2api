package handler

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/authidentity"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/oauth"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	oauthPendingBrowserCookiePath = "/api/v1/auth/oauth"
	oauthPendingBrowserCookieName = "oauth_pending_browser_session"
	oauthPendingSessionCookiePath = "/api/v1/auth/oauth/pending"
	oauthPendingSessionCookieName = "oauth_pending_session"
	oauthPendingCookieMaxAgeSec   = 10 * 60

	oauthCompletionResponseKey = "completion_response"
)

type oauthPendingSessionPayload struct {
	Intent                 string
	Identity               service.PendingAuthIdentityKey
	TargetUserID           *int64
	ResolvedEmail          string
	RedirectTo             string
	BrowserSessionKey      string
	UpstreamIdentityClaims map[string]any
	CompletionResponse     map[string]any
}

type oauthAdoptionDecisionRequest struct {
	AdoptDisplayName *bool `json:"adopt_display_name,omitempty"`
	AdoptAvatar      *bool `json:"adopt_avatar,omitempty"`
}

type createPendingOAuthAccountRequest struct {
	Email            string `json:"email" binding:"required,email"`
	VerifyCode       string `json:"verify_code,omitempty"`
	Password         string `json:"password" binding:"required,min=6"`
	InvitationCode   string `json:"invitation_code,omitempty"`
	AdoptDisplayName *bool  `json:"adopt_display_name,omitempty"`
	AdoptAvatar      *bool  `json:"adopt_avatar,omitempty"`
}

type bindPendingOAuthLoginRequest struct {
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required"`
	AdoptDisplayName *bool  `json:"adopt_display_name,omitempty"`
	AdoptAvatar      *bool  `json:"adopt_avatar,omitempty"`
}

func (r createPendingOAuthAccountRequest) adoptionDecision() oauthAdoptionDecisionRequest {
	return oauthAdoptionDecisionRequest{AdoptDisplayName: r.AdoptDisplayName, AdoptAvatar: r.AdoptAvatar}
}

func (r bindPendingOAuthLoginRequest) adoptionDecision() oauthAdoptionDecisionRequest {
	return oauthAdoptionDecisionRequest{AdoptDisplayName: r.AdoptDisplayName, AdoptAvatar: r.AdoptAvatar}
}

func (h *AuthHandler) pendingIdentityService() (*service.AuthPendingIdentityService, error) {
	if h == nil || h.authService == nil || h.authService.EntClient() == nil {
		return nil, infraerrors.ServiceUnavailable("PENDING_AUTH_NOT_READY", "pending auth service is not ready")
	}
	return service.NewAuthPendingIdentityService(h.authService.EntClient()), nil
}

func generateOAuthPendingBrowserSession() (string, error) {
	return oauth.GenerateState()
}

func setOAuthPendingBrowserCookie(c *gin.Context, sessionKey string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthPendingBrowserCookieName,
		Value:    encodeCookieValue(sessionKey),
		Path:     oauthPendingBrowserCookiePath,
		MaxAge:   oauthPendingCookieMaxAgeSec,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearOAuthPendingBrowserCookie(c *gin.Context, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthPendingBrowserCookieName,
		Value:    "",
		Path:     oauthPendingBrowserCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func readOAuthPendingBrowserCookie(c *gin.Context) (string, error) {
	return readCookieDecoded(c, oauthPendingBrowserCookieName)
}

func setOAuthPendingSessionCookie(c *gin.Context, sessionToken string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthPendingSessionCookieName,
		Value:    encodeCookieValue(sessionToken),
		Path:     oauthPendingSessionCookiePath,
		MaxAge:   oauthPendingCookieMaxAgeSec,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearOAuthPendingSessionCookie(c *gin.Context, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthPendingSessionCookieName,
		Value:    "",
		Path:     oauthPendingSessionCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func readOAuthPendingSessionCookie(c *gin.Context) (string, error) {
	return readCookieDecoded(c, oauthPendingSessionCookieName)
}

func redirectToFrontendCallback(c *gin.Context, frontendCallback string) {
	u, err := url.Parse(frontendCallback)
	if err != nil {
		c.Redirect(http.StatusFound, linuxDoOAuthDefaultRedirectTo)
		return
	}
	if u.Scheme != "" && !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		c.Redirect(http.StatusFound, linuxDoOAuthDefaultRedirectTo)
		return
	}
	u.Fragment = ""
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Redirect(http.StatusFound, u.String())
}

func (h *AuthHandler) createOAuthPendingSession(c *gin.Context, payload oauthPendingSessionPayload) error {
	svc, err := h.pendingIdentityService()
	if err != nil {
		return err
	}

	session, err := svc.CreatePendingSession(c.Request.Context(), service.CreatePendingAuthSessionInput{
		Intent:                 strings.TrimSpace(payload.Intent),
		Identity:               payload.Identity,
		TargetUserID:           payload.TargetUserID,
		ResolvedEmail:          strings.TrimSpace(payload.ResolvedEmail),
		RedirectTo:             strings.TrimSpace(payload.RedirectTo),
		BrowserSessionKey:      strings.TrimSpace(payload.BrowserSessionKey),
		UpstreamIdentityClaims: payload.UpstreamIdentityClaims,
		LocalFlowState: map[string]any{
			oauthCompletionResponseKey: payload.CompletionResponse,
		},
	})
	if err != nil {
		return infraerrors.InternalServer("PENDING_AUTH_SESSION_CREATE_FAILED", "failed to create pending auth session").WithCause(err)
	}

	setOAuthPendingSessionCookie(c, session.SessionToken, isRequestHTTPS(c))
	return nil
}

func linuxDoInvitationPendingPayload(email, username, subject, redirectTo, browserSessionKey string) oauthPendingSessionPayload {
	return oauthPendingSessionPayload{
		Intent: "login",
		Identity: service.PendingAuthIdentityKey{
			ProviderType:    "linuxdo",
			ProviderKey:     "linuxdo",
			ProviderSubject: strings.TrimSpace(subject),
		},
		ResolvedEmail:     strings.TrimSpace(email),
		RedirectTo:        strings.TrimSpace(redirectTo),
		BrowserSessionKey: strings.TrimSpace(browserSessionKey),
		UpstreamIdentityClaims: map[string]any{
			"email":    strings.TrimSpace(email),
			"username": strings.TrimSpace(username),
			"subject":  strings.TrimSpace(subject),
		},
		CompletionResponse: map[string]any{
			"error":    "invitation_required",
			"redirect": strings.TrimSpace(redirectTo),
		},
	}
}

func oidcInvitationPendingPayload(email, username, issuer, subject string, emailVerified bool, redirectTo, browserSessionKey string) oauthPendingSessionPayload {
	return oauthPendingSessionPayload{
		Intent: "login",
		Identity: service.PendingAuthIdentityKey{
			ProviderType:    "oidc",
			ProviderKey:     strings.TrimSpace(issuer),
			ProviderSubject: strings.TrimSpace(subject),
		},
		ResolvedEmail:     strings.TrimSpace(email),
		RedirectTo:        strings.TrimSpace(redirectTo),
		BrowserSessionKey: strings.TrimSpace(browserSessionKey),
		UpstreamIdentityClaims: map[string]any{
			"email":          strings.TrimSpace(email),
			"username":       strings.TrimSpace(username),
			"subject":        strings.TrimSpace(subject),
			"issuer":         strings.TrimSpace(issuer),
			"email_verified": emailVerified,
		},
		CompletionResponse: map[string]any{
			"error":    "invitation_required",
			"redirect": strings.TrimSpace(redirectTo),
		},
	}
}

func readCompletionResponse(session map[string]any) (map[string]any, bool) {
	if len(session) == 0 {
		return nil, false
	}
	value, ok := session[oauthCompletionResponseKey]
	if !ok {
		return nil, false
	}
	result, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	return result, true
}

func pendingSessionStringValue(values map[string]any, key string) string {
	if len(values) == 0 {
		return ""
	}
	raw, ok := values[key]
	if !ok {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func pendingSessionWantsInvitation(payload map[string]any) bool {
	return strings.EqualFold(strings.TrimSpace(pendingSessionStringValue(payload, "error")), "invitation_required")
}

func clonePendingCompletionMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func mergePendingCompletionResponse(session *dbent.PendingAuthSession, overrides map[string]any) map[string]any {
	payload, _ := readCompletionResponse(session.LocalFlowState)
	merged := clonePendingCompletionMap(payload)
	if strings.TrimSpace(session.RedirectTo) != "" {
		if _, exists := merged["redirect"]; !exists {
			merged["redirect"] = session.RedirectTo
		}
	}
	for key, value := range overrides {
		if value == nil {
			delete(merged, key)
			continue
		}
		merged[key] = value
	}
	applySuggestedProfileToCompletionResponse(merged, session.UpstreamIdentityClaims)
	return merged
}

func applySuggestedProfileToCompletionResponse(payload map[string]any, upstream map[string]any) {
	if len(payload) == 0 || len(upstream) == 0 {
		return
	}

	displayName := pendingSessionStringValue(upstream, "suggested_display_name")
	avatarURL := pendingSessionStringValue(upstream, "suggested_avatar_url")

	if displayName != "" {
		if _, exists := payload["suggested_display_name"]; !exists {
			payload["suggested_display_name"] = displayName
		}
	}
	if avatarURL != "" {
		if _, exists := payload["suggested_avatar_url"]; !exists {
			payload["suggested_avatar_url"] = avatarURL
		}
	}
	if displayName != "" || avatarURL != "" {
		payload["adoption_required"] = true
	}
}

func readPendingOAuthBrowserSession(c *gin.Context, h *AuthHandler) (*service.AuthPendingIdentityService, *dbent.PendingAuthSession, func(), error) {
	secureCookie := isRequestHTTPS(c)
	clearCookies := func() {
		clearOAuthPendingSessionCookie(c, secureCookie)
		clearOAuthPendingBrowserCookie(c, secureCookie)
	}

	sessionToken, err := readOAuthPendingSessionCookie(c)
	if err != nil || strings.TrimSpace(sessionToken) == "" {
		clearCookies()
		return nil, nil, clearCookies, service.ErrPendingAuthSessionNotFound
	}
	browserSessionKey, err := readOAuthPendingBrowserCookie(c)
	if err != nil || strings.TrimSpace(browserSessionKey) == "" {
		clearCookies()
		return nil, nil, clearCookies, service.ErrPendingAuthBrowserMismatch
	}

	svc, err := h.pendingIdentityService()
	if err != nil {
		clearCookies()
		return nil, nil, clearCookies, err
	}
	session, err := svc.GetBrowserSession(c.Request.Context(), sessionToken, browserSessionKey)
	if err != nil {
		clearCookies()
		return nil, nil, clearCookies, err
	}
	return svc, session, clearCookies, nil
}

func (r oauthAdoptionDecisionRequest) hasDecision() bool {
	return r.AdoptDisplayName != nil || r.AdoptAvatar != nil
}

func (r oauthAdoptionDecisionRequest) toServiceInput(sessionID int64) service.PendingIdentityAdoptionDecisionInput {
	input := service.PendingIdentityAdoptionDecisionInput{PendingAuthSessionID: sessionID}
	if r.AdoptDisplayName != nil {
		input.AdoptDisplayName = *r.AdoptDisplayName
	}
	if r.AdoptAvatar != nil {
		input.AdoptAvatar = *r.AdoptAvatar
	}
	return input
}

func (h *AuthHandler) upsertPendingOAuthAdoptionDecision(c *gin.Context, sessionID int64, req oauthAdoptionDecisionRequest) error {
	if !req.hasDecision() {
		return nil
	}
	svc, err := h.pendingIdentityService()
	if err != nil {
		return err
	}
	if _, err := svc.UpsertAdoptionDecision(c.Request.Context(), req.toServiceInput(sessionID)); err != nil {
		return infraerrors.InternalServer("PENDING_AUTH_ADOPTION_SAVE_FAILED", "failed to save oauth profile adoption decision").WithCause(err)
	}
	return nil
}

func updatePendingOAuthSessionProgress(ctx context.Context, client *dbent.Client, session *dbent.PendingAuthSession, intent string, resolvedEmail string, targetUserID *int64, completionResponse map[string]any) (*dbent.PendingAuthSession, error) {
	if client == nil || session == nil {
		return nil, infraerrors.BadRequest("PENDING_AUTH_SESSION_INVALID", "pending auth session is invalid")
	}
	localFlowState := clonePendingCompletionMap(session.LocalFlowState)
	localFlowState[oauthCompletionResponseKey] = clonePendingCompletionMap(completionResponse)
	update := client.PendingAuthSession.UpdateOneID(session.ID).
		SetIntent(strings.TrimSpace(intent)).
		SetResolvedEmail(strings.TrimSpace(resolvedEmail)).
		SetLocalFlowState(localFlowState)
	if targetUserID != nil && *targetUserID > 0 {
		update = update.SetTargetUserID(*targetUserID)
	} else {
		update = update.ClearTargetUserID()
	}
	return update.Save(ctx)
}

func ensurePendingOAuthIdentityForUser(ctx context.Context, client *dbent.Client, session *dbent.PendingAuthSession, userID int64) error {
	if client == nil || session == nil || userID <= 0 {
		return infraerrors.BadRequest("PENDING_AUTH_SESSION_INVALID", "pending auth session is invalid")
	}
	providerType := strings.TrimSpace(session.ProviderType)
	providerKey := strings.TrimSpace(session.ProviderKey)
	providerSubject := strings.TrimSpace(session.ProviderSubject)
	identity, err := client.AuthIdentity.Query().
		Where(
			authidentity.ProviderTypeEQ(providerType),
			authidentity.ProviderKeyEQ(providerKey),
			authidentity.ProviderSubjectEQ(providerSubject),
		).
		Only(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return err
	}
	if identity != nil {
		if identity.UserID != userID {
			return infraerrors.Conflict("AUTH_IDENTITY_OWNERSHIP_CONFLICT", "auth identity already belongs to another user")
		}
		_, err = client.AuthIdentity.UpdateOneID(identity.ID).
			SetMetadata(clonePendingCompletionMap(session.UpstreamIdentityClaims)).
			Save(ctx)
		return err
	}
	_, err = client.AuthIdentity.Create().
		SetUserID(userID).
		SetProviderType(providerType).
		SetProviderKey(providerKey).
		SetProviderSubject(providerSubject).
		SetMetadata(clonePendingCompletionMap(session.UpstreamIdentityClaims)).
		Save(ctx)
	return err
}

func rejectPendingOAuthIdentityOwnedByAnotherUser(ctx context.Context, client *dbent.Client, session *dbent.PendingAuthSession) error {
	if client == nil || session == nil {
		return infraerrors.BadRequest("PENDING_AUTH_SESSION_INVALID", "pending auth session is invalid")
	}
	identity, err := client.AuthIdentity.Query().
		Where(
			authidentity.ProviderTypeEQ(strings.TrimSpace(session.ProviderType)),
			authidentity.ProviderKeyEQ(strings.TrimSpace(session.ProviderKey)),
			authidentity.ProviderSubjectEQ(strings.TrimSpace(session.ProviderSubject)),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil
		}
		return err
	}
	if identity.UserID > 0 {
		return infraerrors.Conflict("AUTH_IDENTITY_OWNERSHIP_CONFLICT", "auth identity already belongs to another user")
	}
	return nil
}

func buildPendingOAuthSessionStatusPayload(session *dbent.PendingAuthSession) gin.H {
	payload := gin.H{
		"auth_result": "pending_session",
		"provider":    strings.TrimSpace(session.ProviderType),
		"intent":      strings.TrimSpace(session.Intent),
	}
	for key, value := range mergePendingCompletionResponse(session, nil) {
		payload[key] = value
	}
	if email := strings.TrimSpace(session.ResolvedEmail); email != "" {
		payload["email"] = email
	}
	return payload
}

func writeOAuthTokenPairResponse(c *gin.Context, tokenPair *service.TokenPair) {
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
		"token_type":    "Bearer",
	})
}

// CreatePendingOAuthAccount completes a DB-backed pending OAuth session by
// creating a local email account, or asks the frontend to bind-login when the
// email already belongs to an existing user.
func (h *AuthHandler) CreatePendingOAuthAccount(c *gin.Context) {
	var req createPendingOAuthAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	pendingSvc, session, clearCookies, err := readPendingOAuthBrowserSession(c, h)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	client := h.authService.EntClient()
	if client == nil {
		response.ErrorFrom(c, infraerrors.ServiceUnavailable("PENDING_AUTH_NOT_READY", "pending auth service is not ready"))
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	existingUser, err := client.User.Query().Where(dbuser.EmailEQ(email)).Only(c.Request.Context())
	if err != nil && !dbent.IsNotFound(err) {
		response.ErrorFrom(c, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "service temporarily unavailable"))
		return
	}
	if existingUser != nil {
		completionResponse := mergePendingCompletionResponse(session, map[string]any{
			"step":  "bind_login_required",
			"email": email,
		})
		session, err = updatePendingOAuthSessionProgress(c.Request.Context(), client, session, "adopt_existing_user_by_email", email, &existingUser.ID, completionResponse)
		if err != nil {
			response.ErrorFrom(c, infraerrors.InternalServer("PENDING_AUTH_SESSION_UPDATE_FAILED", "failed to update pending oauth session").WithCause(err))
			return
		}
		if err := h.upsertPendingOAuthAdoptionDecision(c, session.ID, req.adoptionDecision()); err != nil {
			response.ErrorFrom(c, err)
			return
		}
		c.JSON(http.StatusOK, buildPendingOAuthSessionStatusPayload(session))
		return
	}
	if err := rejectPendingOAuthIdentityOwnedByAnotherUser(c.Request.Context(), client, session); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	tokenPair, user, err := h.authService.RegisterOAuthEmailAccount(c.Request.Context(), email, req.Password, strings.TrimSpace(req.VerifyCode), strings.TrimSpace(req.InvitationCode), strings.TrimSpace(session.ProviderType))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.upsertPendingOAuthAdoptionDecision(c, session.ID, req.adoptionDecision()); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := ensurePendingOAuthIdentityForUser(c.Request.Context(), client, session, user.ID); err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("PENDING_AUTH_BIND_APPLY_FAILED", "failed to bind pending oauth identity").WithCause(err))
		return
	}
	if _, err := pendingSvc.ConsumeBrowserSession(c.Request.Context(), session.SessionToken, session.BrowserSessionKey); err != nil {
		clearCookies()
		response.ErrorFrom(c, err)
		return
	}

	clearCookies()
	writeOAuthTokenPairResponse(c, tokenPair)
}

// BindPendingOAuthLogin completes a DB-backed pending OAuth session by binding
// it to an existing local account after password verification.
func (h *AuthHandler) BindPendingOAuthLogin(c *gin.Context) {
	var req bindPendingOAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	pendingSvc, session, clearCookies, err := readPendingOAuthBrowserSession(c, h)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	user, err := h.authService.ValidatePasswordCredentials(c.Request.Context(), strings.TrimSpace(req.Email), req.Password)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if session.TargetUserID != nil && *session.TargetUserID > 0 && user.ID != *session.TargetUserID {
		response.ErrorFrom(c, infraerrors.Conflict("PENDING_AUTH_TARGET_USER_MISMATCH", "pending oauth session must be completed by the targeted user"))
		return
	}
	if err := h.upsertPendingOAuthAdoptionDecision(c, session.ID, req.adoptionDecision()); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := ensurePendingOAuthIdentityForUser(c.Request.Context(), h.authService.EntClient(), session, user.ID); err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("PENDING_AUTH_BIND_APPLY_FAILED", "failed to bind pending oauth identity").WithCause(err))
		return
	}
	h.authService.RecordSuccessfulLogin(c.Request.Context(), user.ID)
	tokenPair, err := h.authService.GenerateTokenPair(c.Request.Context(), user, "")
	if err != nil {
		response.InternalError(c, "Failed to generate token pair")
		return
	}
	if _, err := pendingSvc.ConsumeBrowserSession(c.Request.Context(), session.SessionToken, session.BrowserSessionKey); err != nil {
		clearCookies()
		response.ErrorFrom(c, err)
		return
	}

	clearCookies()
	writeOAuthTokenPairResponse(c, tokenPair)
}

// ExchangePendingOAuthCompletion redeems a pending OAuth browser session into a frontend-safe payload.
// POST /api/v1/auth/oauth/pending/exchange
func (h *AuthHandler) ExchangePendingOAuthCompletion(c *gin.Context) {
	secureCookie := isRequestHTTPS(c)
	clearCookies := func() {
		clearOAuthPendingSessionCookie(c, secureCookie)
		clearOAuthPendingBrowserCookie(c, secureCookie)
	}

	sessionToken, err := readOAuthPendingSessionCookie(c)
	if err != nil || strings.TrimSpace(sessionToken) == "" {
		clearCookies()
		response.ErrorFrom(c, service.ErrPendingAuthSessionNotFound)
		return
	}
	browserSessionKey, err := readOAuthPendingBrowserCookie(c)
	if err != nil || strings.TrimSpace(browserSessionKey) == "" {
		clearCookies()
		response.ErrorFrom(c, service.ErrPendingAuthBrowserMismatch)
		return
	}

	svc, err := h.pendingIdentityService()
	if err != nil {
		clearCookies()
		response.ErrorFrom(c, err)
		return
	}

	session, err := svc.GetBrowserSession(c.Request.Context(), sessionToken, browserSessionKey)
	if err != nil {
		clearCookies()
		response.ErrorFrom(c, err)
		return
	}

	payload, ok := readCompletionResponse(session.LocalFlowState)
	if !ok {
		clearCookies()
		response.ErrorFrom(c, infraerrors.InternalServer("PENDING_AUTH_COMPLETION_INVALID", "pending auth completion payload is invalid"))
		return
	}
	if strings.TrimSpace(session.RedirectTo) != "" {
		if _, exists := payload["redirect"]; !exists {
			payload["redirect"] = session.RedirectTo
		}
	}
	applySuggestedProfileToCompletionResponse(payload, session.UpstreamIdentityClaims)

	if pendingSessionWantsInvitation(payload) {
		response.Success(c, payload)
		return
	}

	if _, err := svc.ConsumeBrowserSession(c.Request.Context(), sessionToken, browserSessionKey); err != nil {
		clearCookies()
		response.ErrorFrom(c, err)
		return
	}

	clearCookies()
	response.Success(c, payload)
}
