//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthServiceAuthenticatePasswordDoesNotIssueRouterToken(t *testing.T) {
	user := &User{ID: 42, Email: "creator@example.com", Username: "creator", Role: RoleUser, Status: StatusActive}
	require.NoError(t, user.SetPassword("correct-password"))
	svc := newAuthService(&userRepoStub{user: user}, nil, nil, nil)

	authenticated, err := svc.AuthenticatePassword(context.Background(), user.Email, "correct-password")
	require.NoError(t, err)
	require.Same(t, user, authenticated)

	token, publicLoginUser, err := svc.Login(context.Background(), user.Email, "correct-password")
	require.NoError(t, err)
	require.NotEmpty(t, token, "the existing public Router login must still issue a token")
	require.Same(t, user, publicLoginUser)
}

func TestAuthServiceAuthenticatePasswordPreservesExistingLoginErrors(t *testing.T) {
	active := &User{ID: 42, Email: "creator@example.com", Role: RoleUser, Status: StatusActive}
	require.NoError(t, active.SetPassword("correct-password"))
	inactive := &User{ID: 43, Email: "inactive@example.com", Role: RoleUser, Status: "disabled"}
	require.NoError(t, inactive.SetPassword("correct-password"))

	tests := []struct {
		name     string
		repo     *userRepoStub
		email    string
		password string
		wantErr  error
	}{
		{name: "unknown email", repo: &userRepoStub{}, email: "missing@example.com", password: "correct-password", wantErr: ErrInvalidCredentials},
		{name: "wrong password", repo: &userRepoStub{user: active}, email: active.Email, password: "wrong-password", wantErr: ErrInvalidCredentials},
		{name: "inactive user", repo: &userRepoStub{user: inactive}, email: inactive.Email, password: "correct-password", wantErr: ErrUserNotActive},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := newAuthService(test.repo, nil, nil, nil)
			user, err := svc.AuthenticatePassword(context.Background(), test.email, test.password)
			require.ErrorIs(t, err, test.wantErr)
			require.Nil(t, user)
		})
	}
}
