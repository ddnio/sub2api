//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthServiceRegisterWithoutTokenUsesExistingRegistrationRules(t *testing.T) {
	repo := &userRepoStub{nextID: 123}
	svc := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil, nil)

	// 该入口不能依赖 Router JWT 配置，更不能在注册成功后偷偷签发 Router token。
	svc.cfg = nil
	user, err := svc.RegisterWithoutTokenWithVerification(
		context.Background(),
		"studio-user@example.com",
		"password",
		"",
		"",
		"",
		"",
	)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, int64(123), user.ID)
	require.Equal(t, "studio-user@example.com", user.Email)
	require.Len(t, repo.created, 1)
}

func TestAuthServiceRegisterWithoutTokenPreservesRegistrationDisabled(t *testing.T) {
	svc := newAuthService(&userRepoStub{}, map[string]string{
		SettingKeyRegistrationEnabled: "false",
	}, nil, nil)

	user, err := svc.RegisterWithoutTokenWithVerification(
		context.Background(),
		"studio-user@example.com",
		"password",
		"",
		"",
		"",
		"",
	)
	require.Nil(t, user)
	require.ErrorIs(t, err, ErrRegDisabled)
}
