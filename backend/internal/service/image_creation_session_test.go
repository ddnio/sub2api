package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type imageCreationSessionStoreStub struct {
	ticket       ImageCreationSessionClaims
	ticketUsed   bool
	session      ImageCreationSessionClaims
	storeErr     error
	sessionToken string
}

func (s *imageCreationSessionStoreStub) StoreTicket(_ context.Context, claims ImageCreationSessionClaims, _ time.Duration) (string, error) {
	if s.storeErr != nil {
		return "", s.storeErr
	}
	s.ticket = claims
	return "ticket-token", nil
}

func (s *imageCreationSessionStoreStub) ConsumeTicket(_ context.Context, token string) (ImageCreationSessionClaims, bool, error) {
	if token != "ticket-token" || s.ticketUsed {
		return ImageCreationSessionClaims{}, false, nil
	}
	s.ticketUsed = true
	return s.ticket, true, nil
}

func (s *imageCreationSessionStoreStub) StoreSession(_ context.Context, claims ImageCreationSessionClaims, _ time.Duration) (string, error) {
	s.session = claims
	if s.sessionToken == "" {
		s.sessionToken = "scoped-session"
	}
	return s.sessionToken, nil
}

func (s *imageCreationSessionStoreStub) GetSession(_ context.Context, token string) (ImageCreationSessionClaims, bool, error) {
	if token != s.sessionToken {
		return ImageCreationSessionClaims{}, false, nil
	}
	return s.session, true, nil
}

type imageCreationUserReaderStub struct {
	user *User
}

func (s imageCreationUserReaderStub) GetByID(context.Context, int64) (*User, error) {
	if s.user == nil {
		return nil, errors.New("not found")
	}
	copy := *s.user
	return &copy, nil
}

func TestImageCreationSessionTicketIsSingleUseAndScopeBound(t *testing.T) {
	store := &imageCreationSessionStoreStub{}
	svc := NewImageCreationSessionService(store, imageCreationUserReaderStub{user: &User{
		ID: 9, Role: RoleUser, Status: StatusActive, TokenVersion: 7, TokenVersionResolved: true,
	}})

	ticket, err := svc.IssueTicket(context.Background(), 9, ImageCreationScopeUser)
	require.NoError(t, err)
	require.Equal(t, "ticket-token", ticket)

	session, viewer, err := svc.ExchangeTicket(context.Background(), ticket)
	require.NoError(t, err)
	require.Equal(t, "scoped-session", session)
	require.Equal(t, int64(9), viewer.UserID)
	require.Equal(t, ImageCreationScopeUser, viewer.Scope)

	_, _, err = svc.ExchangeTicket(context.Background(), ticket)
	require.ErrorIs(t, err, ErrImageCreationTicketInvalid)

	_, err = svc.Authenticate(context.Background(), session, ImageCreationScopeAdmin)
	require.ErrorIs(t, err, ErrImageCreationAdminRequired)
}

func TestImageCreationSessionFailsClosedWhenStoreIsUnavailable(t *testing.T) {
	svc := NewImageCreationSessionService(&imageCreationSessionStoreStub{storeErr: errors.New("redis unavailable")}, imageCreationUserReaderStub{user: &User{
		ID: 9, Role: RoleUser, Status: StatusActive, TokenVersion: 7, TokenVersionResolved: true,
	}})

	_, err := svc.IssueTicket(context.Background(), 9, ImageCreationScopeUser)
	require.ErrorIs(t, err, ErrImageCreationSessionUnavailable)
}
