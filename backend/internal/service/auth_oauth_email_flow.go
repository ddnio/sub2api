package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// VerifyOAuthEmailCode verifies the locally entered email verification code for
// third-party signup and binding flows. This is intentionally independent from
// the global registration email verification toggle.
func (s *AuthService) VerifyOAuthEmailCode(ctx context.Context, email, verifyCode string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	verifyCode = strings.TrimSpace(verifyCode)
	if email == "" || verifyCode == "" {
		return ErrEmailVerifyRequired
	}
	if s == nil || s.emailService == nil {
		return ErrServiceUnavailable
	}
	return s.emailService.VerifyCode(ctx, email, verifyCode)
}

// RegisterOAuthEmailAccount creates a local account from a third-party first
// login after the user has verified a local email address. It follows this
// fork's existing registration rules so historical invitation/redeem/referral
// data keeps its current meaning while the public flow aligns with upstream.
func (s *AuthService) RegisterOAuthEmailAccount(
	ctx context.Context,
	email string,
	password string,
	verifyCode string,
	invitationCode string,
	signupSource string,
) (*TokenPair, *User, error) {
	if s == nil {
		return nil, nil, ErrServiceUnavailable
	}
	if s.settingService == nil || !s.settingService.IsRegistrationEnabled(ctx) {
		return nil, nil, ErrRegDisabled
	}
	if s.refreshTokenCache == nil {
		return nil, nil, ErrServiceUnavailable
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if isReservedEmail(email) {
		return nil, nil, ErrEmailReserved
	}
	if err := s.validateRegistrationEmailPolicy(ctx, email); err != nil {
		return nil, nil, err
	}
	if err := s.VerifyOAuthEmailCode(ctx, email, verifyCode); err != nil {
		return nil, nil, err
	}

	var invitationRedeemCode *RedeemCode
	if s.settingService.IsInvitationCodeEnabled(ctx) {
		if strings.TrimSpace(invitationCode) == "" {
			return nil, nil, ErrInvitationCodeRequired
		}
		redeemCode, err := s.redeemRepo.GetByCode(ctx, strings.TrimSpace(invitationCode))
		if err != nil {
			logger.LegacyPrintf("service.auth", "[Auth] Invalid oauth email invitation code: %s, error: %v", invitationCode, err)
			return nil, nil, ErrInvitationCodeInvalid
		}
		if redeemCode.Type != RedeemTypeInvitation || redeemCode.Status != StatusUnused {
			logger.LegacyPrintf("service.auth", "[Auth] OAuth email invitation code invalid: type=%s, status=%s", redeemCode.Type, redeemCode.Status)
			return nil, nil, ErrInvitationCodeInvalid
		}
		invitationRedeemCode = redeemCode
	}

	existsEmail, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Database error checking oauth email exists: %v", err)
		return nil, nil, ErrServiceUnavailable
	}
	if existsEmail {
		return nil, nil, ErrEmailExists
	}

	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	defaultBalance := s.cfg.Default.UserBalance
	defaultConcurrency := s.cfg.Default.UserConcurrency
	if s.settingService != nil {
		defaultBalance = s.settingService.GetDefaultBalance(ctx)
		defaultConcurrency = s.settingService.GetDefaultConcurrency(ctx)
	}

	user := &User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         RoleUser,
		Balance:      defaultBalance,
		Concurrency:  defaultConcurrency,
		Status:       StatusActive,
	}
	applyUserNotifyDefaults(user)

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, ErrEmailExists) {
			return nil, nil, ErrEmailExists
		}
		logger.LegacyPrintf("service.auth", "[Auth] Database error creating oauth email user: %v", err)
		return nil, nil, ErrServiceUnavailable
	}

	s.assignDefaultSubscriptions(ctx, user.ID)

	if invitationRedeemCode != nil {
		if err := s.redeemRepo.Use(ctx, invitationRedeemCode.ID, user.ID); err != nil {
			logger.LegacyPrintf("service.auth", "[Auth] Failed to mark oauth email invitation code as used for user %d: %v", user.ID, err)
		}
	}

	if s.referralService != nil {
		if _, err := s.referralService.GenerateReferralCode(ctx, user.ID); err != nil {
			logger.LegacyPrintf("service.auth", "[Auth] Failed to generate referral code for oauth email user %d: %v", user.ID, err)
		}
	}

	if s.referralService != nil && s.settingService != nil &&
		!s.settingService.IsInvitationCodeEnabled(ctx) && s.settingService.IsReferralEnabled(ctx) &&
		strings.TrimSpace(invitationCode) != "" {
		if err := s.referralService.ProcessRegistrationReferral(ctx, user.ID, invitationCode); err != nil {
			logger.LegacyPrintf("service.auth", "[Auth] Failed to process oauth email referral for user %d: %v", user.ID, err)
		}
	}

	s.updateUserSignupSource(ctx, user.ID, normalizeOAuthSignupSource(signupSource))
	s.touchUserLogin(ctx, user.ID)

	tokenPair, err := s.GenerateTokenPair(ctx, user, "")
	if err != nil {
		return nil, nil, fmt.Errorf("generate token pair: %w", err)
	}
	return tokenPair, user, nil
}

// ValidatePasswordCredentials checks the local password without completing the
// login flow. Pending third-party account adoption calls this before binding
// the external identity.
func (s *AuthService) ValidatePasswordCredentials(ctx context.Context, email, password string) (*User, error) {
	if s == nil {
		return nil, ErrServiceUnavailable
	}
	user, err := s.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, ErrServiceUnavailable
	}
	if !user.IsActive() {
		return nil, ErrUserNotActive
	}
	if !s.CheckPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// RecordSuccessfulLogin updates login activity after a non-standard login flow
// finishes with a real session.
func (s *AuthService) RecordSuccessfulLogin(ctx context.Context, userID int64) {
	s.touchUserLogin(ctx, userID)
}

func (s *AuthService) updateUserSignupSource(ctx context.Context, userID int64, signupSource string) {
	if s == nil || s.entClient == nil || userID <= 0 {
		return
	}
	signupSource = normalizeOAuthSignupSource(signupSource)
	if err := s.entClient.User.UpdateOneID(userID).
		SetSignupSource(signupSource).
		Exec(ctx); err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Failed to update signup source: user_id=%d source=%s err=%v", userID, signupSource, err)
	}
}

func (s *AuthService) touchUserLogin(ctx context.Context, userID int64) {
	if s == nil || s.entClient == nil || userID <= 0 {
		return
	}
	now := time.Now().UTC()
	if err := s.entClient.User.UpdateOneID(userID).
		SetLastLoginAt(now).
		SetLastActiveAt(now).
		Exec(ctx); err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Failed to touch login timestamps: user_id=%d err=%v", userID, err)
	}
}

func normalizeOAuthSignupSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "linuxdo", "wechat", "oidc":
		return strings.ToLower(strings.TrimSpace(source))
	default:
		return "email"
	}
}
