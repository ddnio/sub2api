package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/pendingauthsession"
	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func newOAuthPendingHandlerTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:auth_oauth_pending_flow?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

type oauthPendingTestEncryptor struct {
	plaintext string
}

func (e oauthPendingTestEncryptor) Encrypt(string) (string, error) { return "encrypted", nil }
func (e oauthPendingTestEncryptor) Decrypt(string) (string, error) { return e.plaintext, nil }

type oauthPendingRefreshTokenCacheStub struct{}

func (s *oauthPendingRefreshTokenCacheStub) StoreRefreshToken(context.Context, string, *service.RefreshTokenData, time.Duration) error {
	return nil
}

func (s *oauthPendingRefreshTokenCacheStub) GetRefreshToken(context.Context, string) (*service.RefreshTokenData, error) {
	return nil, service.ErrRefreshTokenNotFound
}

func (s *oauthPendingRefreshTokenCacheStub) DeleteRefreshToken(context.Context, string) error {
	return nil
}

func (s *oauthPendingRefreshTokenCacheStub) DeleteUserRefreshTokens(context.Context, int64) error {
	return nil
}

func (s *oauthPendingRefreshTokenCacheStub) DeleteTokenFamily(context.Context, string) error {
	return nil
}

func (s *oauthPendingRefreshTokenCacheStub) AddToUserTokenSet(context.Context, int64, string, time.Duration) error {
	return nil
}

func (s *oauthPendingRefreshTokenCacheStub) AddToFamilyTokenSet(context.Context, string, string, time.Duration) error {
	return nil
}

func (s *oauthPendingRefreshTokenCacheStub) GetUserTokenHashes(context.Context, int64) ([]string, error) {
	return nil, nil
}

func (s *oauthPendingRefreshTokenCacheStub) GetFamilyTokenHashes(context.Context, string) ([]string, error) {
	return nil, nil
}

func (s *oauthPendingRefreshTokenCacheStub) IsTokenInFamily(context.Context, string, string) (bool, error) {
	return false, nil
}

type oauthPendingTotpLoginSessionCacheStub struct {
	loginSessions map[string]*service.TotpLoginSession
}

func (s *oauthPendingTotpLoginSessionCacheStub) GetSetupSession(context.Context, int64) (*service.TotpSetupSession, error) {
	return nil, nil
}

func (s *oauthPendingTotpLoginSessionCacheStub) SetSetupSession(context.Context, int64, *service.TotpSetupSession, time.Duration) error {
	return nil
}

func (s *oauthPendingTotpLoginSessionCacheStub) DeleteSetupSession(context.Context, int64) error {
	return nil
}

func (s *oauthPendingTotpLoginSessionCacheStub) GetLoginSession(_ context.Context, tempToken string) (*service.TotpLoginSession, error) {
	return s.loginSessions[tempToken], nil
}

func (s *oauthPendingTotpLoginSessionCacheStub) SetLoginSession(_ context.Context, tempToken string, session *service.TotpLoginSession, _ time.Duration) error {
	if s.loginSessions == nil {
		s.loginSessions = map[string]*service.TotpLoginSession{}
	}
	s.loginSessions[tempToken] = session
	return nil
}

func (s *oauthPendingTotpLoginSessionCacheStub) DeleteLoginSession(_ context.Context, tempToken string) error {
	delete(s.loginSessions, tempToken)
	return nil
}

func (s *oauthPendingTotpLoginSessionCacheStub) IncrementVerifyAttempts(context.Context, int64) (int, error) {
	return 0, nil
}

func (s *oauthPendingTotpLoginSessionCacheStub) GetVerifyAttempts(context.Context, int64) (int, error) {
	return 0, nil
}

func (s *oauthPendingTotpLoginSessionCacheStub) ClearVerifyAttempts(context.Context, int64) error {
	return nil
}

func TestApplySuggestedProfileToCompletionResponse(t *testing.T) {
	payload := map[string]any{
		"access_token": "token",
	}
	upstream := map[string]any{
		"suggested_display_name": "Alice",
		"suggested_avatar_url":   "https://cdn.example/avatar.png",
	}

	applySuggestedProfileToCompletionResponse(payload, upstream)

	require.Equal(t, "Alice", payload["suggested_display_name"])
	require.Equal(t, "https://cdn.example/avatar.png", payload["suggested_avatar_url"])
	require.Equal(t, true, payload["adoption_required"])
}

func TestApplySuggestedProfileToCompletionResponseKeepsExistingPayloadValues(t *testing.T) {
	payload := map[string]any{
		"suggested_display_name": "Existing",
		"adoption_required":      false,
	}
	upstream := map[string]any{
		"suggested_display_name": "Alice",
		"suggested_avatar_url":   "https://cdn.example/avatar.png",
	}

	applySuggestedProfileToCompletionResponse(payload, upstream)

	require.Equal(t, "Existing", payload["suggested_display_name"])
	require.Equal(t, "https://cdn.example/avatar.png", payload["suggested_avatar_url"])
	require.Equal(t, true, payload["adoption_required"])
}

func TestInvitationPendingPayloads(t *testing.T) {
	linuxDoPayload := linuxDoInvitationPendingPayload(" user@example.com ", " alice ", " subject-1 ", "/profile", " browser ")
	require.Equal(t, "login", linuxDoPayload.Intent)
	require.Equal(t, "linuxdo", linuxDoPayload.Identity.ProviderType)
	require.Equal(t, "linuxdo", linuxDoPayload.Identity.ProviderKey)
	require.Equal(t, "subject-1", linuxDoPayload.Identity.ProviderSubject)
	require.Equal(t, "user@example.com", linuxDoPayload.ResolvedEmail)
	require.Equal(t, "browser", linuxDoPayload.BrowserSessionKey)
	require.Equal(t, "invitation_required", linuxDoPayload.CompletionResponse["error"])

	oidcPayload := oidcInvitationPendingPayload(" oidc@example.com ", " bob ", " https://issuer.example ", " sub-1 ", true, "/dashboard", " browser-2 ")
	require.Equal(t, "oidc", oidcPayload.Identity.ProviderType)
	require.Equal(t, "https://issuer.example", oidcPayload.Identity.ProviderKey)
	require.Equal(t, "sub-1", oidcPayload.Identity.ProviderSubject)
	require.Equal(t, true, oidcPayload.UpstreamIdentityClaims["email_verified"])
	require.Equal(t, "/dashboard", oidcPayload.CompletionResponse["redirect"])
}

func TestMergePendingCompletionResponseMarksExistingAccountBindLogin(t *testing.T) {
	session := &dbent.PendingAuthSession{
		RedirectTo: "/profile",
		LocalFlowState: map[string]any{
			oauthCompletionResponseKey: map[string]any{
				"error": "invitation_required",
			},
		},
		UpstreamIdentityClaims: map[string]any{
			"suggested_display_name": "Alice",
		},
	}

	payload := mergePendingCompletionResponse(session, map[string]any{
		"step":  "bind_login_required",
		"email": "user@example.com",
	})

	require.Equal(t, "bind_login_required", payload["step"])
	require.Equal(t, "user@example.com", payload["email"])
	require.Equal(t, "/profile", payload["redirect"])
	require.Equal(t, "Alice", payload["suggested_display_name"])
	require.Equal(t, true, payload["adoption_required"])
}

func TestEnsurePendingOAuthIdentityForUserCreatesAndRejectsOwnershipConflict(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()

	user, err := client.User.Create().
		SetEmail("bind@example.com").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	other, err := client.User.Create().
		SetEmail("other@example.com").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	session := &dbent.PendingAuthSession{
		ProviderType:    "linuxdo",
		ProviderKey:     "linuxdo",
		ProviderSubject: "subject-1",
		UpstreamIdentityClaims: map[string]any{
			"email": "linuxdo-1@linuxdo-connect.invalid",
		},
	}

	err = ensurePendingOAuthIdentityForUser(ctx, client, session, user.ID)
	require.NoError(t, err)
	identity, err := client.AuthIdentity.Query().Only(ctx)
	require.NoError(t, err)
	require.Equal(t, user.ID, identity.UserID)

	err = ensurePendingOAuthIdentityForUser(ctx, client, session, other.ID)
	require.Error(t, err)
	require.Equal(t, "AUTH_IDENTITY_OWNERSHIP_CONFLICT", serviceErrorReason(t, err))
}

func TestPendingOAuthBindApplyErrorPreservesApplicationErrors(t *testing.T) {
	err := infraerrors.Conflict("AUTH_IDENTITY_OWNERSHIP_CONFLICT", "auth identity already belongs to another user")

	wrapped := pendingOAuthBindApplyError(err)

	require.Equal(t, "AUTH_IDENTITY_OWNERSHIP_CONFLICT", serviceErrorReason(t, wrapped))
}

func TestPendingOAuthBindApplyErrorWrapsUnexpectedErrors(t *testing.T) {
	wrapped := pendingOAuthBindApplyError(errors.New("db failed"))

	require.Equal(t, "PENDING_AUTH_BIND_APPLY_FAILED", serviceErrorReason(t, wrapped))
}

func TestBindPendingOAuthLoginRequires2FAWithoutBindingOrConsumingSession(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	settingRepo := repository.NewSettingRepository(client)
	require.NoError(t, settingRepo.Set(ctx, service.SettingKeyTotpEnabled, "true"))

	userRepo := repository.NewUserRepository(client, nil)
	settingSvc := service.NewSettingService(settingRepo, &config.Config{})
	authSvc := service.NewAuthService(client, userRepo, nil, &oauthPendingRefreshTokenCacheStub{}, &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", ExpireHour: 1, RefreshTokenExpireDays: 30},
	}, settingSvc, nil, nil, nil, nil, nil, nil)
	hash, err := authSvc.HashPassword("password")
	require.NoError(t, err)
	user, err := client.User.Create().
		SetEmail("totp-bind@example.com").
		SetPasswordHash(hash).
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		SetTotpEnabled(true).
		SetTotpSecretEncrypted("encrypted").
		SetTotpEnabledAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	pendingSvc := service.NewAuthPendingIdentityService(client)
	pendingSession, err := pendingSvc.CreatePendingSession(ctx, service.CreatePendingAuthSessionInput{
		Intent: "login",
		Identity: service.PendingAuthIdentityKey{
			ProviderType:    "linuxdo",
			ProviderKey:     "linuxdo",
			ProviderSubject: "subject-bind-login-2fa",
		},
		BrowserSessionKey:      "browser-2fa",
		UpstreamIdentityClaims: map[string]any{"email": "linuxdo-bind-login-2fa@linuxdo-connect.invalid"},
	})
	require.NoError(t, err)

	totpCache := &oauthPendingTotpLoginSessionCacheStub{}
	totpSvc := service.NewTotpService(userRepo, oauthPendingTestEncryptor{plaintext: "JBSWY3DPEHPK3PXP"}, totpCache, settingSvc, nil, nil)
	handler := NewAuthHandler(&config.Config{}, authSvc, service.NewUserService(userRepo, nil, nil), settingSvc, nil, nil, totpSvc)

	body, err := json.Marshal(bindPendingOAuthLoginRequest{Email: user.Email, Password: "password"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/pending/bind-login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: oauthPendingSessionCookieName, Value: encodeCookieValue(pendingSession.SessionToken)})
	req.AddCookie(&http.Cookie{Name: oauthPendingBrowserCookieName, Value: encodeCookieValue(pendingSession.BrowserSessionKey)})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler.BindPendingOAuthLogin(c)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var payload struct {
		Data struct {
			Requires2FA bool   `json:"requires_2fa"`
			TempToken   string `json:"temp_token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.True(t, payload.Data.Requires2FA)
	require.NotEmpty(t, payload.Data.TempToken)

	count, err := client.AuthIdentity.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
	stored, err := client.PendingAuthSession.Get(ctx, pendingSession.ID)
	require.NoError(t, err)
	require.Nil(t, stored.ConsumedAt)
	loginSession, err := totpSvc.GetLoginSession(ctx, payload.Data.TempToken)
	require.NoError(t, err)
	require.NotNil(t, loginSession)
	require.NotNil(t, loginSession.PendingOAuthBind)
}

func TestCompletePendingOAuthBindSessionCreatesIdentity(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()

	user, err := client.User.Create().
		SetEmail("bind-target@example.com").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	session := &dbent.PendingAuthSession{
		ProviderType:    "linuxdo",
		ProviderKey:     "linuxdo",
		ProviderSubject: "subject-bind",
		TargetUserID:    &user.ID,
		UpstreamIdentityClaims: map[string]any{
			"email": "linuxdo-bind@linuxdo-connect.invalid",
		},
	}

	err = completePendingOAuthBindSession(ctx, client, session)
	require.NoError(t, err)

	identity, err := client.AuthIdentity.Query().Only(ctx)
	require.NoError(t, err)
	require.Equal(t, user.ID, identity.UserID)
	require.Equal(t, "linuxdo", identity.ProviderType)
	require.Equal(t, "subject-bind", identity.ProviderSubject)
}

func TestLogin2FACompletesPendingOAuthBindSession(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	settingRepo := repository.NewSettingRepository(client)
	require.NoError(t, settingRepo.Set(ctx, service.SettingKeyTotpEnabled, "true"))

	userRepo := repository.NewUserRepository(client, nil)
	settingSvc := service.NewSettingService(settingRepo, &config.Config{})
	authSvc := service.NewAuthService(client, userRepo, nil, &oauthPendingRefreshTokenCacheStub{}, &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", ExpireHour: 1, RefreshTokenExpireDays: 30},
	}, settingSvc, nil, nil, nil, nil, nil, nil)
	hash, err := authSvc.HashPassword("password")
	require.NoError(t, err)
	totpSecret := "JBSWY3DPEHPK3PXP"
	totpCode, err := totp.GenerateCode(totpSecret, time.Now().UTC())
	require.NoError(t, err)
	user, err := client.User.Create().
		SetEmail("totp-bind@example.com").
		SetPasswordHash(hash).
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		SetTotpEnabled(true).
		SetTotpSecretEncrypted("encrypted").
		SetTotpEnabledAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	pendingSvc := service.NewAuthPendingIdentityService(client)
	pendingSession, err := pendingSvc.CreatePendingSession(ctx, service.CreatePendingAuthSessionInput{
		Intent: "login",
		Identity: service.PendingAuthIdentityKey{
			ProviderType:    "linuxdo",
			ProviderKey:     "linuxdo",
			ProviderSubject: "subject-2fa",
		},
		BrowserSessionKey:      "browser-2fa",
		UpstreamIdentityClaims: map[string]any{"email": "linuxdo-2fa@linuxdo-connect.invalid"},
	})
	require.NoError(t, err)

	totpCache := &oauthPendingTotpLoginSessionCacheStub{}
	totpSvc := service.NewTotpService(userRepo, oauthPendingTestEncryptor{plaintext: totpSecret}, totpCache, settingSvc, nil, nil)
	tempToken, err := totpSvc.CreatePendingOAuthBindLoginSession(ctx, user.ID, user.Email, pendingSession.SessionToken, pendingSession.BrowserSessionKey)
	require.NoError(t, err)
	handler := NewAuthHandler(&config.Config{}, authSvc, service.NewUserService(userRepo, nil, nil), settingSvc, nil, nil, totpSvc)

	body, err := json.Marshal(Login2FARequest{TempToken: tempToken, TotpCode: totpCode})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/2fa", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler.Login2FA(c)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var payload struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.NotEmpty(t, payload.Data.AccessToken)

	identity, err := client.AuthIdentity.Query().Only(ctx)
	require.NoError(t, err)
	require.Equal(t, user.ID, identity.UserID)
	require.Equal(t, "linuxdo", identity.ProviderType)
	require.Equal(t, "subject-2fa", identity.ProviderSubject)

	stored, err := client.PendingAuthSession.Query().
		Where(pendingauthsession.IDEQ(pendingSession.ID)).
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, stored.ConsumedAt)
	loginSession, err := totpSvc.GetLoginSession(ctx, tempToken)
	require.NoError(t, err)
	require.Nil(t, loginSession)
}

func TestRejectPendingOAuthIdentityOwnedByAnotherUser(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()

	user, err := client.User.Create().
		SetEmail("owner@example.com").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	session := &dbent.PendingAuthSession{
		ProviderType:    "oidc",
		ProviderKey:     "https://issuer.example",
		ProviderSubject: "subject-owned",
	}
	require.NoError(t, rejectPendingOAuthIdentityOwnedByAnotherUser(ctx, client, session))

	_, err = client.AuthIdentity.Create().
		SetUserID(user.ID).
		SetProviderType(session.ProviderType).
		SetProviderKey(session.ProviderKey).
		SetProviderSubject(session.ProviderSubject).
		SetMetadata(map[string]any{}).
		Save(ctx)
	require.NoError(t, err)

	err = rejectPendingOAuthIdentityOwnedByAnotherUser(ctx, client, session)
	require.Error(t, err)
	require.Equal(t, "AUTH_IDENTITY_OWNERSHIP_CONFLICT", serviceErrorReason(t, err))
}

func serviceErrorReason(t *testing.T, err error) string {
	t.Helper()
	appErr := infraerrors.FromError(err)
	if appErr == nil {
		return ""
	}
	return appErr.Reason
}
