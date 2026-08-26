package handler

import (
	"bytes"
	"context"
	"encoding/json"
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
	registrationArgs []string
	recordedLogins   []int64
}

func (s *studioAuthServiceStub) RegisterWithoutTokenWithVerification(_ context.Context, email, password, verifyCode, promoCode, invitationCode, affiliateCode string) (*service.User, error) {
	s.registrationArgs = []string{email, password, verifyCode, promoCode, invitationCode, affiliateCode}
	return s.registered, nil
}

func (s *studioAuthServiceStub) SendVerifyCodeAsync(context.Context, string, ...string) (*service.SendVerifyCodeResult, error) {
	return &service.SendVerifyCodeResult{Countdown: 60}, nil
}

func (s *studioAuthServiceStub) AuthenticatePassword(context.Context, string, string) (*service.User, error) {
	return s.authenticated, nil
}

func (s *studioAuthServiceStub) RecordSuccessfulLogin(_ context.Context, userID int64) {
	s.recordedLogins = append(s.recordedLogins, userID)
}

type studioTotpServiceStub struct {
	sessions        map[string]*service.TotpLoginSession
	createdAudience string
	verifiedUserIDs []int64
	deletedTokens   []string
}

func (s *studioTotpServiceStub) CreateLoginSessionForAudience(_ context.Context, userID int64, email, audience string) (string, error) {
	s.createdAudience = audience
	return "studio-challenge", nil
}

func (s *studioTotpServiceStub) GetLoginSession(_ context.Context, tempToken string) (*service.TotpLoginSession, error) {
	return s.sessions[tempToken], nil
}

func (s *studioTotpServiceStub) VerifyCode(_ context.Context, userID int64, _ string) error {
	s.verifiedUserIDs = append(s.verifiedUserIDs, userID)
	return nil
}

func (s *studioTotpServiceStub) DeleteLoginSession(_ context.Context, tempToken string) error {
	s.deletedTokens = append(s.deletedTokens, tempToken)
	delete(s.sessions, tempToken)
	return nil
}

type studioUserServiceStub struct {
	user *service.User
}

func (s *studioUserServiceStub) GetByID(context.Context, int64) (*service.User, error) {
	return s.user, nil
}

type studioSettingServiceStub struct {
	totpEnabled bool
}

func (s *studioSettingServiceStub) IsTotpEnabled(context.Context) bool {
	return s.totpEnabled
}

func TestStudioAuthRegisterReturnsIdentityWithoutRouterTokens(t *testing.T) {
	auth := &studioAuthServiceStub{registered: studioUser(42, false)}
	handler := NewStudioAuthHandler(auth, &studioUserServiceStub{}, &studioSettingServiceStub{}, &studioTotpServiceStub{})

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
	user := data["user"].(map[string]any)
	require.Equal(t, "019c0000-0000-7000-8000-000000000042", user["subject"])
	require.Equal(t, "studio@example.com", user["email"])
}

func TestStudioAuthLoginCreatesAudienceBound2FAChallenge(t *testing.T) {
	auth := &studioAuthServiceStub{authenticated: studioUser(42, true)}
	totp := &studioTotpServiceStub{}
	handler := NewStudioAuthHandler(auth, &studioUserServiceStub{}, &studioSettingServiceStub{totpEnabled: true}, totp)

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

func TestStudioAuthLogin2FAConsumesOnlyStudioChallenges(t *testing.T) {
	user := studioUser(42, true)

	t.Run("reject Router challenge", func(t *testing.T) {
		auth := &studioAuthServiceStub{}
		totp := &studioTotpServiceStub{sessions: map[string]*service.TotpLoginSession{
			"router-challenge": {UserID: user.ID, Email: user.Email},
		}}
		handler := NewStudioAuthHandler(auth, &studioUserServiceStub{user: user}, &studioSettingServiceStub{}, totp)

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
		handler := NewStudioAuthHandler(auth, &studioUserServiceStub{user: user}, &studioSettingServiceStub{}, totp)

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
