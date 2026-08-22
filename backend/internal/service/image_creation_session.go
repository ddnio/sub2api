package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ImageCreationScopeUser  = "user"
	ImageCreationScopeAdmin = "admin"

	ImageCreationTicketTTL  = time.Minute
	ImageCreationSessionTTL = 2 * time.Hour
)

var (
	ErrImageCreationTicketInvalid      = errors.New("image creation ticket is invalid")
	ErrImageCreationSessionInvalid     = errors.New("image creation session is invalid")
	ErrImageCreationSessionUnavailable = errors.New("image creation session store unavailable")
	ErrImageCreationAdminRequired      = errors.New("image creation admin access required")
)

type ImageCreationSessionClaims struct {
	UserID       int64  `json:"user_id"`
	TokenVersion int64  `json:"token_version"`
	Scope        string `json:"scope"`
}

type ImageCreationSessionViewer struct {
	UserID int64  `json:"id"`
	Role   string `json:"role"`
	Scope  string `json:"scope"`
}

type ImageCreationSessionStore interface {
	StoreTicket(ctx context.Context, claims ImageCreationSessionClaims, ttl time.Duration) (string, error)
	ConsumeTicket(ctx context.Context, token string) (ImageCreationSessionClaims, bool, error)
	StoreSession(ctx context.Context, claims ImageCreationSessionClaims, ttl time.Duration) (string, error)
	GetSession(ctx context.Context, token string) (ImageCreationSessionClaims, bool, error)
}

type imageCreationSessionUserReader interface {
	GetByID(ctx context.Context, id int64) (*User, error)
}

type ImageCreationSessionService struct {
	store ImageCreationSessionStore
	users imageCreationSessionUserReader
}

func NewImageCreationSessionService(store ImageCreationSessionStore, users *UserService) *ImageCreationSessionService {
	return newImageCreationSessionService(store, users)
}

func newImageCreationSessionService(store ImageCreationSessionStore, users imageCreationSessionUserReader) *ImageCreationSessionService {
	return &ImageCreationSessionService{store: store, users: users}
}

func (s *ImageCreationSessionService) IssueTicket(ctx context.Context, userID int64, scope string) (string, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil || !user.IsActive() {
		return "", ErrImageCreationSessionInvalid
	}
	if scope != ImageCreationScopeUser && scope != ImageCreationScopeAdmin {
		return "", ErrImageCreationSessionInvalid
	}
	if scope == ImageCreationScopeAdmin && !user.IsAdmin() {
		return "", ErrImageCreationAdminRequired
	}
	token, err := s.store.StoreTicket(ctx, ImageCreationSessionClaims{
		UserID: user.ID, TokenVersion: user.TokenVersion, Scope: scope,
	}, ImageCreationTicketTTL)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrImageCreationSessionUnavailable, err)
	}
	return token, nil
}

func (s *ImageCreationSessionService) ExchangeTicket(ctx context.Context, ticket string) (string, ImageCreationSessionViewer, error) {
	if strings.TrimSpace(ticket) == "" {
		return "", ImageCreationSessionViewer{}, ErrImageCreationTicketInvalid
	}
	claims, ok, err := s.store.ConsumeTicket(ctx, ticket)
	if err != nil {
		return "", ImageCreationSessionViewer{}, fmt.Errorf("%w: %v", ErrImageCreationSessionUnavailable, err)
	}
	if !ok {
		return "", ImageCreationSessionViewer{}, ErrImageCreationTicketInvalid
	}
	viewer, err := s.validateClaims(ctx, claims)
	if err != nil {
		return "", ImageCreationSessionViewer{}, err
	}
	token, err := s.store.StoreSession(ctx, claims, ImageCreationSessionTTL)
	if err != nil {
		return "", ImageCreationSessionViewer{}, fmt.Errorf("%w: %v", ErrImageCreationSessionUnavailable, err)
	}
	return token, viewer, nil
}

func (s *ImageCreationSessionService) Authenticate(ctx context.Context, token, requiredScope string) (ImageCreationSessionViewer, error) {
	if strings.TrimSpace(token) == "" {
		return ImageCreationSessionViewer{}, ErrImageCreationSessionInvalid
	}
	claims, ok, err := s.store.GetSession(ctx, token)
	if err != nil {
		return ImageCreationSessionViewer{}, fmt.Errorf("%w: %v", ErrImageCreationSessionUnavailable, err)
	}
	if !ok {
		return ImageCreationSessionViewer{}, ErrImageCreationSessionInvalid
	}
	viewer, err := s.validateClaims(ctx, claims)
	if err != nil {
		return ImageCreationSessionViewer{}, err
	}
	if requiredScope == ImageCreationScopeAdmin && viewer.Scope != ImageCreationScopeAdmin {
		return ImageCreationSessionViewer{}, ErrImageCreationAdminRequired
	}
	return viewer, nil
}

func (s *ImageCreationSessionService) validateClaims(ctx context.Context, claims ImageCreationSessionClaims) (ImageCreationSessionViewer, error) {
	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil || user == nil || !user.IsActive() || user.TokenVersion != claims.TokenVersion {
		return ImageCreationSessionViewer{}, ErrImageCreationSessionInvalid
	}
	if claims.Scope == ImageCreationScopeAdmin && !user.IsAdmin() {
		return ImageCreationSessionViewer{}, ErrImageCreationAdminRequired
	}
	if claims.Scope != ImageCreationScopeUser && claims.Scope != ImageCreationScopeAdmin {
		return ImageCreationSessionViewer{}, ErrImageCreationSessionInvalid
	}
	return ImageCreationSessionViewer{UserID: user.ID, Role: user.Role, Scope: claims.Scope}, nil
}
