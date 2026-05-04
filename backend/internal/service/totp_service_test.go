//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type totpLoginSessionCacheStub struct {
	loginSessions map[string]*TotpLoginSession
}

func (s *totpLoginSessionCacheStub) GetSetupSession(context.Context, int64) (*TotpSetupSession, error) {
	return nil, nil
}

func (s *totpLoginSessionCacheStub) SetSetupSession(context.Context, int64, *TotpSetupSession, time.Duration) error {
	return nil
}

func (s *totpLoginSessionCacheStub) DeleteSetupSession(context.Context, int64) error { return nil }

func (s *totpLoginSessionCacheStub) GetLoginSession(_ context.Context, tempToken string) (*TotpLoginSession, error) {
	return s.loginSessions[tempToken], nil
}

func (s *totpLoginSessionCacheStub) SetLoginSession(_ context.Context, tempToken string, session *TotpLoginSession, _ time.Duration) error {
	if s.loginSessions == nil {
		s.loginSessions = map[string]*TotpLoginSession{}
	}
	s.loginSessions[tempToken] = session
	return nil
}

func (s *totpLoginSessionCacheStub) DeleteLoginSession(_ context.Context, tempToken string) error {
	delete(s.loginSessions, tempToken)
	return nil
}

func (s *totpLoginSessionCacheStub) IncrementVerifyAttempts(context.Context, int64) (int, error) {
	return 0, nil
}

func (s *totpLoginSessionCacheStub) GetVerifyAttempts(context.Context, int64) (int, error) {
	return 0, nil
}

func (s *totpLoginSessionCacheStub) ClearVerifyAttempts(context.Context, int64) error { return nil }

func TestCreatePendingOAuthBindLoginSessionStoresBrowserBoundSession(t *testing.T) {
	cache := &totpLoginSessionCacheStub{}
	svc := &TotpService{cache: cache}

	tempToken, err := svc.CreatePendingOAuthBindLoginSession(context.Background(), 42, "user@example.com", "pending-token", "browser-key")
	require.NoError(t, err)
	require.NotEmpty(t, tempToken)

	session, err := svc.GetLoginSession(context.Background(), tempToken)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, int64(42), session.UserID)
	require.Equal(t, "user@example.com", session.Email)
	require.NotNil(t, session.PendingOAuthBind)
	require.Equal(t, "pending-token", session.PendingOAuthBind.PendingSessionToken)
	require.Equal(t, "browser-key", session.PendingOAuthBind.BrowserSessionKey)
}
