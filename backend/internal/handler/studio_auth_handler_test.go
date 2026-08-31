package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type studioAuthServiceStub struct {
	registered       *service.User
	authenticated    *service.User
	registerErr      error
	sendErr          error
	authenticateErr  error
	forgotErr        error
	resetErr         error
	registrationArgs []string
	forgotArgs       []string
	resetArgs        []string
	recordedLogins   []int64
}

func (s *studioAuthServiceStub) RegisterWithoutTokenWithVerification(_ context.Context, email, password, verifyCode, promoCode, invitationCode, affiliateCode string) (*service.User, error) {
	s.registrationArgs = []string{email, password, verifyCode, promoCode, invitationCode, affiliateCode}
	return s.registered, s.registerErr
}

func (s *studioAuthServiceStub) SendVerifyCodeAsync(context.Context, string, ...string) (*service.SendVerifyCodeResult, error) {
	return &service.SendVerifyCodeResult{Countdown: 60}, s.sendErr
}

func (s *studioAuthServiceStub) AuthenticatePassword(context.Context, string, string) (*service.User, error) {
	return s.authenticated, s.authenticateErr
}

func (s *studioAuthServiceStub) RequestPasswordResetAsync(_ context.Context, email, frontendBaseURL string, locale ...string) error {
	s.forgotArgs = []string{email, frontendBaseURL}
	s.forgotArgs = append(s.forgotArgs, locale...)
	return s.forgotErr
}

func (s *studioAuthServiceStub) ResetPassword(_ context.Context, email, token, newPassword string) error {
	s.resetArgs = []string{email, token, newPassword}
	return s.resetErr
}

func (s *studioAuthServiceStub) RecordSuccessfulLogin(_ context.Context, userID int64) {
	s.recordedLogins = append(s.recordedLogins, userID)
}

type studioTotpServiceStub struct {
	sessions        map[string]*service.TotpLoginSession
	createdAudience string
	verifiedUserIDs []int64
	deletedTokens   []string
	createErr       error
	getErr          error
	verifyErr       error
	deleteErr       error
}

func (s *studioTotpServiceStub) CreateLoginSessionForAudience(_ context.Context, userID int64, email, audience string) (string, error) {
	s.createdAudience = audience
	return "studio-challenge", s.createErr
}

func (s *studioTotpServiceStub) GetLoginSession(_ context.Context, tempToken string) (*service.TotpLoginSession, error) {
	return s.sessions[tempToken], s.getErr
}

func (s *studioTotpServiceStub) VerifyCode(_ context.Context, userID int64, _ string) error {
	s.verifiedUserIDs = append(s.verifiedUserIDs, userID)
	return s.verifyErr
}

func (s *studioTotpServiceStub) DeleteLoginSession(_ context.Context, tempToken string) error {
	s.deletedTokens = append(s.deletedTokens, tempToken)
	if s.deleteErr == nil {
		delete(s.sessions, tempToken)
	}
	return s.deleteErr
}

type studioUserServiceStub struct {
	user *service.User
	err  error
}

func (s *studioUserServiceStub) GetByID(context.Context, int64) (*service.User, error) {
	return s.user, s.err
}

func (s *studioUserServiceStub) GetByEmail(context.Context, string) (*service.User, error) {
	return s.user, s.err
}

type studioSettingServiceStub struct {
	totpEnabled bool
}

func (s *studioSettingServiceStub) IsTotpEnabled(context.Context) bool {
	return s.totpEnabled
}

func TestStudioAuthRegisterReturnsIdentityWithoutRouterTokens(t *testing.T) {
	auth := &studioAuthServiceStub{registered: studioUser(42, false)}
	handler := newStudioAuthHandler(auth, &studioUserServiceStub{}, &studioSettingServiceStub{}, &studioTotpServiceStub{})

	recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/register", `{
		"email":"studio@example.com",
		"password":"Password123!",
		"verify_code":"246810",
		"promo_code":"PROMO"
	}`, handler.Register)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "access_token")
	require.NotContains(t, recorder.Body.String(), "refresh_token")
	require.Equal(t, []string{"studio@example.com", "Password123!", "246810", "PROMO", "", ""}, auth.registrationArgs)
	data := studioResponseData(t, recorder)
	user, ok := data["user"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "019c0000-0000-7000-8000-000000000042", user["subject"])
	require.Equal(t, "studio@example.com", user["email"])
}

func TestNewStudioAuthHandlerWiresProductionServices(t *testing.T) {
	require.NotNil(t, NewStudioAuthHandler(nil, nil, nil, nil))
}

func TestStudioAuthRegisterRejectsInvalidOrMissingIdentity(t *testing.T) {
	handler := newStudioAuthHandler(&studioAuthServiceStub{}, &studioUserServiceStub{}, &studioSettingServiceStub{}, &studioTotpServiceStub{})

	invalid := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/register", `{"email":"invalid"}`, handler.Register)
	require.Equal(t, http.StatusBadRequest, invalid.Code)

	missingIdentity := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/register", `{"email":"studio@example.com","password":"Password123!"}`, handler.Register)
	require.Equal(t, http.StatusInternalServerError, missingIdentity.Code)
}

func TestStudioAuthSendVerifyCodeAndValidation(t *testing.T) {
	handler := newStudioAuthHandler(&studioAuthServiceStub{}, &studioUserServiceStub{}, &studioSettingServiceStub{}, &studioTotpServiceStub{})

	success := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/send-verify-code", `{"email":"studio@example.com"}`, handler.SendVerifyCode)
	require.Equal(t, http.StatusOK, success.Code)
	require.Equal(t, float64(60), studioResponseData(t, success)["countdown"])

	invalid := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/send-verify-code", `{"email":"invalid"}`, handler.SendVerifyCode)
	require.Equal(t, http.StatusBadRequest, invalid.Code)
}

func TestStudioAuthPasswordRecoveryUsesTheSignedStudioCallback(t *testing.T) {
	auth := &studioAuthServiceStub{}
	handler := newStudioAuthHandler(auth, &studioUserServiceStub{}, &studioSettingServiceStub{}, &studioTotpServiceStub{})

	forgot := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/forgot-password", `{
		"email":"studio@example.com",
		"frontend_base_url":"https://studio.nanafox.com/tools/image-studio"
	}`, handler.ForgotPassword)
	require.Equal(t, http.StatusOK, forgot.Code)
	require.Equal(t, []string{"studio@example.com", "https://studio.nanafox.com/tools/image-studio", ""}, auth.forgotArgs)

	unsafeURL := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/forgot-password", `{
		"email":"studio@example.com",
		"frontend_base_url":"http://evil.example/reset"
	}`, handler.ForgotPassword)
	require.Equal(t, http.StatusBadRequest, unsafeURL.Code)

	reset := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/reset-password", `{
		"email":"studio@example.com",
		"token":"single-use-token",
		"new_password":"NewPassword123!"
	}`, handler.ResetPassword)
	require.Equal(t, http.StatusOK, reset.Code)
	require.Equal(t, []string{"studio@example.com", "single-use-token", "NewPassword123!"}, auth.resetArgs)
}

func TestStudioAuthServiceErrorsAreReturned(t *testing.T) {
	t.Run("send verify code", func(t *testing.T) {
		handler := newStudioAuthHandler(&studioAuthServiceStub{sendErr: errors.New("send failed")}, &studioUserServiceStub{}, &studioSettingServiceStub{}, &studioTotpServiceStub{})
		recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/send-verify-code", `{"email":"studio@example.com"}`, handler.SendVerifyCode)
		require.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("register", func(t *testing.T) {
		handler := newStudioAuthHandler(&studioAuthServiceStub{registerErr: service.ErrRegDisabled}, &studioUserServiceStub{}, &studioSettingServiceStub{}, &studioTotpServiceStub{})
		recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/register", `{"email":"studio@example.com","password":"Password123!"}`, handler.Register)
		require.Equal(t, http.StatusForbidden, recorder.Code)
	})

	t.Run("login", func(t *testing.T) {
		handler := newStudioAuthHandler(&studioAuthServiceStub{authenticateErr: service.ErrInvalidCredentials}, &studioUserServiceStub{}, &studioSettingServiceStub{}, &studioTotpServiceStub{})
		recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/login", `{"email":"studio@example.com","password":"wrong"}`, handler.Login)
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestStudioAuthLoginCreatesAudienceBound2FAChallenge(t *testing.T) {
	auth := &studioAuthServiceStub{authenticated: studioUser(42, true)}
	totp := &studioTotpServiceStub{}
	handler := newStudioAuthHandler(auth, &studioUserServiceStub{}, &studioSettingServiceStub{totpEnabled: true}, totp)

	recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/login", `{
		"email":"studio@example.com",
		"password":"Password123!"
	}`, handler.Login)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, service.TotpLoginAudienceStudio, totp.createdAudience)
	require.Empty(t, auth.recordedLogins)
	data := studioResponseData(t, recorder)
	require.Equal(t, true, data["requires_2fa"])
	require.Equal(t, "studio-challenge", data["temp_token"])
}

func TestStudioAuthLoginReturnsIdentityWithout2FA(t *testing.T) {
	user := studioUser(42, false)
	user.Role = service.RoleAdmin
	auth := &studioAuthServiceStub{authenticated: user}
	handler := newStudioAuthHandler(auth, &studioUserServiceStub{}, &studioSettingServiceStub{totpEnabled: true}, &studioTotpServiceStub{})

	recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/login", `{"email":"studio@example.com","password":"Password123!"}`, handler.Login)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, []int64{42}, auth.recordedLogins)
	identity, ok := studioResponseData(t, recorder)["user"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, auth.authenticated.PublicID, identity["subject"])
	require.Equal(t, service.RoleAdmin, identity["role"])
}

func TestStudioAuthResolveReturnsCurrentRouterRole(t *testing.T) {
	user := studioUser(42, false)
	user.Role = service.RoleAdmin
	handler := newStudioAuthHandler(&studioAuthServiceStub{}, &studioUserServiceStub{user: user}, &studioSettingServiceStub{}, &studioTotpServiceStub{})

	recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/resolve", `{
		"subject":"019c0000-0000-7000-8000-000000000042",
		"email":"studio@example.com"
	}`, handler.Resolve)

	require.Equal(t, http.StatusOK, recorder.Code)
	identity, ok := studioResponseData(t, recorder)["user"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, service.RoleAdmin, identity["role"])
}

func TestStudioAuthResolveRejectsMismatchedOrInactiveIdentity(t *testing.T) {
	user := studioUser(42, false)
	handler := newStudioAuthHandler(&studioAuthServiceStub{}, &studioUserServiceStub{user: user}, &studioSettingServiceStub{}, &studioTotpServiceStub{})

	mismatch := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/resolve", `{
		"subject":"019c0000-0000-7000-8000-000000000099",
		"email":"studio@example.com"
	}`, handler.Resolve)
	require.Equal(t, http.StatusForbidden, mismatch.Code)

	user.Status = service.StatusDisabled
	inactive := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/resolve", `{
		"subject":"019c0000-0000-7000-8000-000000000042",
		"email":"studio@example.com"
	}`, handler.Resolve)
	require.Equal(t, http.StatusForbidden, inactive.Code)
}

func TestStudioAuthLoginFailsClosedWhen2FAChallengeCannotBeStored(t *testing.T) {
	auth := &studioAuthServiceStub{authenticated: studioUser(42, true)}
	handler := newStudioAuthHandler(auth, &studioUserServiceStub{}, &studioSettingServiceStub{totpEnabled: true}, &studioTotpServiceStub{createErr: errors.New("cache unavailable")})

	recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/login", `{"email":"studio@example.com","password":"Password123!"}`, handler.Login)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Empty(t, auth.recordedLogins)
}

func TestStudioAuthLogin2FAConsumesOnlyStudioChallenges(t *testing.T) {
	user := studioUser(42, true)

	t.Run("reject Router challenge", func(t *testing.T) {
		auth := &studioAuthServiceStub{}
		totp := &studioTotpServiceStub{sessions: map[string]*service.TotpLoginSession{
			"router-challenge": {UserID: user.ID, Email: user.Email},
		}}
		handler := newStudioAuthHandler(auth, &studioUserServiceStub{user: user}, &studioSettingServiceStub{}, totp)

		recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/login/2fa", `{
			"temp_token":"router-challenge",
			"totp_code":"123456"
		}`, handler.Login2FA)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Empty(t, totp.verifiedUserIDs)
		require.Empty(t, totp.deletedTokens)
	})

	t.Run("consume Studio challenge", func(t *testing.T) {
		auth := &studioAuthServiceStub{}
		totp := &studioTotpServiceStub{sessions: map[string]*service.TotpLoginSession{
			"studio-challenge": {
				UserID:   user.ID,
				Email:    user.Email,
				Audience: service.TotpLoginAudienceStudio,
			},
		}}
		handler := newStudioAuthHandler(auth, &studioUserServiceStub{user: user}, &studioSettingServiceStub{}, totp)

		recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/login/2fa", `{
			"temp_token":"studio-challenge",
			"totp_code":"123456"
		}`, handler.Login2FA)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, []int64{42}, totp.verifiedUserIDs)
		require.Equal(t, []string{"studio-challenge"}, totp.deletedTokens)
		require.Equal(t, []int64{42}, auth.recordedLogins)
		require.NotContains(t, recorder.Body.String(), "access_token")
		data := studioResponseData(t, recorder)
		identity := data["user"].(map[string]any)
		require.Equal(t, user.PublicID, identity["subject"])
	})
}

func TestStudioAuthLogin2FAFailurePathsDoNotCreateAStudioLogin(t *testing.T) {
	user := studioUser(42, true)
	tests := []struct {
		name       string
		totp       *studioTotpServiceStub
		users      *studioUserServiceStub
		wantStatus int
	}{
		{
			name: "challenge lookup error",
			totp: &studioTotpServiceStub{
				getErr: errors.New("cache unavailable"),
			},
			users:      &studioUserServiceStub{user: user},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid code",
			totp: &studioTotpServiceStub{
				sessions:  map[string]*service.TotpLoginSession{"challenge": studioSession(user)},
				verifyErr: service.ErrTotpInvalidCode,
			},
			users:      &studioUserServiceStub{user: user},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user lookup error",
			totp: &studioTotpServiceStub{
				sessions: map[string]*service.TotpLoginSession{"challenge": studioSession(user)},
			},
			users:      &studioUserServiceStub{err: service.ErrUserNotFound},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "inactive user",
			totp: &studioTotpServiceStub{
				sessions: map[string]*service.TotpLoginSession{"challenge": studioSession(user)},
			},
			users:      &studioUserServiceStub{user: &service.User{ID: user.ID, PublicID: user.PublicID, Status: "disabled"}},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "challenge delete error",
			totp: &studioTotpServiceStub{
				sessions:  map[string]*service.TotpLoginSession{"challenge": studioSession(user)},
				deleteErr: errors.New("cache unavailable"),
			},
			users:      &studioUserServiceStub{user: user},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			auth := &studioAuthServiceStub{}
			handler := newStudioAuthHandler(auth, test.users, &studioSettingServiceStub{}, test.totp)
			recorder := serveStudioAuthRequest(t, http.MethodPost, "/internal/v1/studio-auth/login/2fa", `{"temp_token":"challenge","totp_code":"123456"}`, handler.Login2FA)
			require.Equal(t, test.wantStatus, recorder.Code)
			require.Empty(t, auth.recordedLogins)
		})
	}
}

func studioUser(id int64, totpEnabled bool) *service.User {
	return &service.User{
		ID:          id,
		PublicID:    "019c0000-0000-7000-8000-000000000042",
		Email:       "studio@example.com",
		Username:    "Studio User",
		Role:        service.RoleUser,
		Status:      service.StatusActive,
		TotpEnabled: totpEnabled,
	}
}

func studioSession(user *service.User) *service.TotpLoginSession {
	return &service.TotpLoginSession{
		UserID:   user.ID,
		Email:    user.Email,
		Audience: service.TotpLoginAudienceStudio,
	}
}

func serveStudioAuthRequest(t *testing.T, method, path, body string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler(ctx)
	return recorder
}

func studioResponseData(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.NotNil(t, payload.Data)
	return payload.Data
}
