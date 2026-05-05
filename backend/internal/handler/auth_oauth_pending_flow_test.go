package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/authidentity"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/pendingauthsession"
	dbsetting "github.com/Wei-Shaw/sub2api/ent/setting"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
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
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS user_avatars (
	user_id INTEGER PRIMARY KEY,
	storage_provider TEXT NOT NULL DEFAULT 'database',
	storage_key TEXT NOT NULL DEFAULT '',
	url TEXT NOT NULL DEFAULT '',
	content_type TEXT NOT NULL DEFAULT '',
	byte_size INTEGER NOT NULL DEFAULT 0,
	sha256 TEXT NOT NULL DEFAULT '',
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`)
	require.NoError(t, err)
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS user_provider_default_grants (
	user_id INTEGER NOT NULL,
	provider_type TEXT NOT NULL,
	grant_reason TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (user_id, provider_type, grant_reason)
)`)
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

type oauthPendingSettingRepoStub struct {
	client *dbent.Client
}

func (r *oauthPendingSettingRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	row, err := r.client.Setting.Query().Where(dbsetting.KeyEQ(key)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrSettingNotFound
		}
		return nil, err
	}
	return &service.Setting{ID: row.ID, Key: row.Key, Value: row.Value, UpdatedAt: row.UpdatedAt}, nil
}

func (r *oauthPendingSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	row, err := r.client.Setting.Query().Where(dbsetting.KeyEQ(key)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return "", service.ErrSettingNotFound
		}
		return "", err
	}
	return row.Value, nil
}

func (r *oauthPendingSettingRepoStub) Set(ctx context.Context, key, value string) error {
	row, err := r.client.Setting.Query().Where(dbsetting.KeyEQ(key)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			_, createErr := r.client.Setting.Create().SetKey(key).SetValue(value).Save(ctx)
			return createErr
		}
		return err
	}
	return r.client.Setting.UpdateOneID(row.ID).SetValue(value).Exec(ctx)
}

func (r *oauthPendingSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		value, err := r.GetValue(ctx, key)
		if err == nil {
			result[key] = value
			continue
		}
		if !errors.Is(err, service.ErrSettingNotFound) {
			return nil, err
		}
	}
	return result, nil
}

func (r *oauthPendingSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	for key, value := range settings {
		if err := r.Set(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (r *oauthPendingSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.client.Setting.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(rows))
	for _, row := range rows {
		result[row.Key] = row.Value
	}
	return result, nil
}

func (r *oauthPendingSettingRepoStub) Delete(ctx context.Context, key string) error {
	_, err := r.client.Setting.Delete().Where(dbsetting.KeyEQ(key)).Exec(ctx)
	return err
}

type oauthPendingUserRepoStub struct {
	client *dbent.Client
}

func (r *oauthPendingUserRepoStub) Create(ctx context.Context, user *service.User) error {
	create := r.client.User.Create().
		SetEmail(user.Email).
		SetUsername(user.Username).
		SetNotes(user.Notes).
		SetPasswordHash(user.PasswordHash).
		SetRole(user.Role).
		SetBalance(user.Balance).
		SetBalanceNotifyEnabled(user.BalanceNotifyEnabled).
		SetBalanceNotifyExtraEmails(service.MarshalNotifyEmails(user.BalanceNotifyExtraEmails)).
		SetBalanceNotifyThresholdType(user.BalanceNotifyThresholdType).
		SetTotalRecharged(user.TotalRecharged).
		SetConcurrency(user.Concurrency).
		SetStatus(user.Status)
	if user.BalanceNotifyThreshold != nil {
		create.SetBalanceNotifyThreshold(*user.BalanceNotifyThreshold)
	}
	created, err := create.Save(ctx)
	if err != nil {
		return service.ErrEmailExists.WithCause(err)
	}
	applyEntUserToServiceUser(user, created)
	return r.ensureEmailAuthIdentity(ctx, created.ID, created.Email)
}

func (r *oauthPendingUserRepoStub) GetByID(ctx context.Context, id int64) (*service.User, error) {
	row, err := oauthPendingClientFromContext(ctx, r.client).User.Query().Where(dbuser.IDEQ(id)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}
	return entUserToServiceUser(row), nil
}

func (r *oauthPendingUserRepoStub) GetByEmail(ctx context.Context, email string) (*service.User, error) {
	row, err := r.client.User.Query().Where(func(s *entsql.Selector) {
		s.Where(entsql.ExprP("LOWER(TRIM(email)) = LOWER(TRIM(?))", email))
	}).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}
	return entUserToServiceUser(row), nil
}

func (r *oauthPendingUserRepoStub) GetFirstAdmin(ctx context.Context) (*service.User, error) {
	panic("unexpected GetFirstAdmin call")
}

func (r *oauthPendingUserRepoStub) Update(ctx context.Context, user *service.User) error {
	_, err := r.client.User.UpdateOneID(user.ID).
		SetEmail(user.Email).
		SetUsername(user.Username).
		SetNotes(user.Notes).
		SetPasswordHash(user.PasswordHash).
		SetRole(user.Role).
		SetBalance(user.Balance).
		SetBalanceNotifyEnabled(user.BalanceNotifyEnabled).
		SetBalanceNotifyExtraEmails(service.MarshalNotifyEmails(user.BalanceNotifyExtraEmails)).
		SetBalanceNotifyThresholdType(user.BalanceNotifyThresholdType).
		SetTotalRecharged(user.TotalRecharged).
		SetConcurrency(user.Concurrency).
		SetStatus(user.Status).
		Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *oauthPendingUserRepoStub) Delete(ctx context.Context, id int64) error {
	return r.client.User.DeleteOneID(id).Exec(ctx)
}

func (r *oauthPendingUserRepoStub) GetUserAvatar(ctx context.Context, userID int64) (*service.UserAvatar, error) {
	return nil, nil
}

func (r *oauthPendingUserRepoStub) UpsertUserAvatar(ctx context.Context, userID int64, input service.UpsertUserAvatarInput) (*service.UserAvatar, error) {
	return nil, nil
}

func (r *oauthPendingUserRepoStub) DeleteUserAvatar(ctx context.Context, userID int64) error {
	return nil
}

func (r *oauthPendingUserRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (r *oauthPendingUserRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters service.UserListFilters) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (r *oauthPendingUserRepoStub) GetLatestUsedAtByUserIDs(ctx context.Context, userIDs []int64) (map[int64]*time.Time, error) {
	return map[int64]*time.Time{}, nil
}

func (r *oauthPendingUserRepoStub) GetLatestUsedAtByUserID(ctx context.Context, userID int64) (*time.Time, error) {
	return nil, nil
}

func (r *oauthPendingUserRepoStub) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	return oauthPendingClientFromContext(ctx, r.client).User.UpdateOneID(id).AddBalance(amount).Exec(ctx)
}

func (r *oauthPendingUserRepoStub) DeductBalance(ctx context.Context, id int64, amount float64) error {
	return oauthPendingClientFromContext(ctx, r.client).User.UpdateOneID(id).AddBalance(-amount).Exec(ctx)
}

func (r *oauthPendingUserRepoStub) UpdateConcurrency(ctx context.Context, id int64, amount int) error {
	return oauthPendingClientFromContext(ctx, r.client).User.UpdateOneID(id).AddConcurrency(amount).Exec(ctx)
}

func (r *oauthPendingUserRepoStub) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.client.User.Query().Where(func(s *entsql.Selector) {
		s.Where(entsql.ExprP("LOWER(TRIM(email)) = LOWER(TRIM(?))", email))
	}).Exist(ctx)
}

func (r *oauthPendingUserRepoStub) RemoveGroupFromAllowedGroups(ctx context.Context, groupID int64) (int64, error) {
	return 0, nil
}

func (r *oauthPendingUserRepoStub) RemoveGroupFromUserAllowedGroups(ctx context.Context, userID int64, groupID int64) error {
	return nil
}

func (r *oauthPendingUserRepoStub) AddGroupToAllowedGroups(ctx context.Context, userID int64, groupID int64) error {
	return nil
}

func (r *oauthPendingUserRepoStub) ListUserAuthIdentities(ctx context.Context, userID int64) ([]service.UserAuthIdentityRecord, error) {
	rows, err := r.client.AuthIdentity.Query().Where(authidentity.UserIDEQ(userID)).All(ctx)
	if err != nil {
		return nil, err
	}
	records := make([]service.UserAuthIdentityRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, service.UserAuthIdentityRecord{
			ProviderType:    row.ProviderType,
			ProviderKey:     row.ProviderKey,
			ProviderSubject: row.ProviderSubject,
			VerifiedAt:      row.VerifiedAt,
			Issuer:          row.Issuer,
			Metadata:        row.Metadata,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return records, nil
}

func (r *oauthPendingUserRepoStub) UnbindUserAuthProvider(ctx context.Context, userID int64, provider string) error {
	_, err := r.client.AuthIdentity.Delete().
		Where(authidentity.UserIDEQ(userID), authidentity.ProviderTypeEQ(provider)).
		Exec(ctx)
	return err
}

func (r *oauthPendingUserRepoStub) UpdateTotpSecret(ctx context.Context, userID int64, encryptedSecret *string) error {
	update := r.client.User.UpdateOneID(userID)
	if encryptedSecret == nil {
		update.ClearTotpSecretEncrypted()
	} else {
		update.SetTotpSecretEncrypted(*encryptedSecret)
	}
	return update.Exec(ctx)
}

func (r *oauthPendingUserRepoStub) EnableTotp(ctx context.Context, userID int64) error {
	return r.client.User.UpdateOneID(userID).SetTotpEnabled(true).SetTotpEnabledAt(time.Now()).Exec(ctx)
}

func (r *oauthPendingUserRepoStub) DisableTotp(ctx context.Context, userID int64) error {
	return r.client.User.UpdateOneID(userID).SetTotpEnabled(false).ClearTotpEnabledAt().ClearTotpSecretEncrypted().Exec(ctx)
}

func (r *oauthPendingUserRepoStub) EnsureEmailAuthIdentity(ctx context.Context, userID int64, email string) error {
	return r.ensureEmailAuthIdentity(ctx, userID, email)
}

func (r *oauthPendingUserRepoStub) ReplaceEmailAuthIdentity(ctx context.Context, userID int64, oldEmail, newEmail string) error {
	_, _ = r.client.AuthIdentity.Delete().
		Where(authidentity.ProviderTypeEQ("email"), authidentity.ProviderKeyEQ("email"), authidentity.ProviderSubjectEQ(strings.TrimSpace(strings.ToLower(oldEmail)))).
		Exec(ctx)
	return r.ensureEmailAuthIdentity(ctx, userID, newEmail)
}

func (r *oauthPendingUserRepoStub) ensureEmailAuthIdentity(ctx context.Context, userID int64, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil
	}
	return oauthPendingClientFromContext(ctx, r.client).AuthIdentity.Create().
		SetUserID(userID).
		SetProviderType("email").
		SetProviderKey("email").
		SetProviderSubject(email).
		SetVerifiedAt(time.Now().UTC()).
		SetMetadata(map[string]any{"source": "handler_test"}).
		OnConflictColumns(
			authidentity.FieldProviderType,
			authidentity.FieldProviderKey,
			authidentity.FieldProviderSubject,
		).
		DoNothing().
		Exec(ctx)
}

func oauthPendingClientFromContext(ctx context.Context, fallback *dbent.Client) *dbent.Client {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	return fallback
}

func entUserToServiceUser(row *dbent.User) *service.User {
	if row == nil {
		return nil
	}
	out := &service.User{}
	applyEntUserToServiceUser(out, row)
	return out
}

func applyEntUserToServiceUser(out *service.User, row *dbent.User) {
	out.ID = row.ID
	out.Email = row.Email
	out.Username = row.Username
	out.Notes = row.Notes
	out.PasswordHash = row.PasswordHash
	out.Role = row.Role
	out.Balance = row.Balance
	out.BalanceNotifyEnabled = row.BalanceNotifyEnabled
	out.BalanceNotifyThreshold = row.BalanceNotifyThreshold
	out.BalanceNotifyThresholdType = row.BalanceNotifyThresholdType
	out.BalanceNotifyExtraEmails = service.ParseNotifyEmails(row.BalanceNotifyExtraEmails)
	out.TotalRecharged = row.TotalRecharged
	out.Concurrency = row.Concurrency
	out.Status = row.Status
	out.ReferralCode = row.ReferralCode
	out.TotpSecretEncrypted = row.TotpSecretEncrypted
	out.TotpEnabled = row.TotpEnabled
	out.TotpEnabledAt = row.TotpEnabledAt
	out.CreatedAt = row.CreatedAt
	out.UpdatedAt = row.UpdatedAt
}

func newOAuthPendingSettingService(client *dbent.Client, cfg *config.Config) *service.SettingService {
	return service.NewSettingService(&oauthPendingSettingRepoStub{client: client}, cfg)
}

func newOAuthPendingUserRepo(client *dbent.Client) service.UserRepository {
	return &oauthPendingUserRepoStub{client: client}
}

func newOAuthPendingAuthService(
	client *dbent.Client,
	userRepo service.UserRepository,
	settingSvc *service.SettingService,
	assigner service.DefaultSubscriptionAssigner,
) *service.AuthService {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "test-secret", ExpireHour: 1, RefreshTokenExpireDays: 30}}
	return service.NewAuthService(
		client,
		userRepo,
		nil,
		&oauthPendingRefreshTokenCacheStub{},
		cfg,
		settingSvc,
		service.NewEmailService(&oauthPendingSettingRepoStub{client: client}, &oauthPendingEmailCacheStub{}),
		nil,
		nil,
		nil,
		nil,
		assigner,
	)
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

type oauthPendingEmailCacheStub struct{}

func (s *oauthPendingEmailCacheStub) GetVerificationCode(context.Context, string) (*service.VerificationCodeData, error) {
	return &service.VerificationCodeData{Code: "123456"}, nil
}

func (s *oauthPendingEmailCacheStub) SetVerificationCode(context.Context, string, *service.VerificationCodeData, time.Duration) error {
	return nil
}

func (s *oauthPendingEmailCacheStub) DeleteVerificationCode(context.Context, string) error {
	return nil
}

func (s *oauthPendingEmailCacheStub) GetNotifyVerifyCode(context.Context, string) (*service.VerificationCodeData, error) {
	return nil, service.ErrInvalidVerifyCode
}

func (s *oauthPendingEmailCacheStub) SetNotifyVerifyCode(context.Context, string, *service.VerificationCodeData, time.Duration) error {
	return nil
}

func (s *oauthPendingEmailCacheStub) DeleteNotifyVerifyCode(context.Context, string) error {
	return nil
}

func (s *oauthPendingEmailCacheStub) GetPasswordResetToken(context.Context, string) (*service.PasswordResetTokenData, error) {
	return nil, service.ErrInvalidResetToken
}

func (s *oauthPendingEmailCacheStub) SetPasswordResetToken(context.Context, string, *service.PasswordResetTokenData, time.Duration) error {
	return nil
}

func (s *oauthPendingEmailCacheStub) DeletePasswordResetToken(context.Context, string) error {
	return nil
}

func (s *oauthPendingEmailCacheStub) IsPasswordResetEmailInCooldown(context.Context, string) bool {
	return false
}

func (s *oauthPendingEmailCacheStub) SetPasswordResetEmailCooldown(context.Context, string, time.Duration) error {
	return nil
}

func (s *oauthPendingEmailCacheStub) IncrNotifyCodeUserRate(context.Context, int64, time.Duration) (int64, error) {
	return 0, nil
}

func (s *oauthPendingEmailCacheStub) GetNotifyCodeUserRate(context.Context, int64) (int64, error) {
	return 0, nil
}

type oauthPendingTotpLoginSessionCacheStub struct {
	loginSessions map[string]*service.TotpLoginSession
}

type oauthPendingDefaultSubscriptionAssignerStub struct {
	calls []service.AssignSubscriptionInput
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

func (s *oauthPendingDefaultSubscriptionAssignerStub) AssignOrExtendSubscription(_ context.Context, input *service.AssignSubscriptionInput) (*service.UserSubscription, bool, error) {
	if input != nil {
		s.calls = append(s.calls, *input)
	}
	return &service.UserSubscription{UserID: input.UserID, GroupID: input.GroupID}, false, nil
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

func TestFindUserByNormalizedEmailMatchesLegacySpacingAndCase(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()

	user, err := client.User.Create().
		SetEmail(" Owner@Example.com ").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	got, err := findUserByNormalizedEmail(ctx, client, "owner@example.com")
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
}

func TestFindUserByNormalizedEmailRejectsDuplicates(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()

	_, err := client.User.Create().
		SetEmail(" Owner@Example.com ").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.User.Create().
		SetEmail("owner@example.com").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = findUserByNormalizedEmail(ctx, client, " owner@example.com ")
	require.Error(t, err)
	require.Equal(t, "USER_EMAIL_CONFLICT", serviceErrorReason(t, err))
}

func TestCreatePendingOAuthAccountPreservesNormalizedDuplicateConflict(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	_, err := client.User.Create().
		SetEmail(" Owner@Example.com ").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.User.Create().
		SetEmail("owner@example.com").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	pendingSvc := service.NewAuthPendingIdentityService(client)
	pendingSession, err := pendingSvc.CreatePendingSession(ctx, service.CreatePendingAuthSessionInput{
		Intent: "login",
		Identity: service.PendingAuthIdentityKey{
			ProviderType:    "oidc",
			ProviderKey:     "https://issuer.example",
			ProviderSubject: "subject-duplicate-email",
		},
		BrowserSessionKey:      "browser-duplicate-email",
		UpstreamIdentityClaims: map[string]any{"email": "oidc-duplicate@oidc-connect.invalid"},
	})
	require.NoError(t, err)

	userRepo := newOAuthPendingUserRepo(client)
	authSvc := newOAuthPendingAuthService(client, userRepo, nil, nil)
	handler := NewAuthHandler(&config.Config{}, authSvc, service.NewUserService(userRepo, nil, nil), nil, nil, nil, nil)

	body, err := json.Marshal(createPendingOAuthAccountRequest{
		Email:      "owner@example.com",
		Password:   "password",
		VerifyCode: "123456",
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/pending/create-account", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: oauthPendingSessionCookieName, Value: encodeCookieValue(pendingSession.SessionToken)})
	req.AddCookie(&http.Cookie{Name: oauthPendingBrowserCookieName, Value: encodeCookieValue(pendingSession.BrowserSessionKey)})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler.CreatePendingOAuthAccount(c)

	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "USER_EMAIL_CONFLICT")
}

func TestBindPendingOAuthLoginRequires2FAWithoutBindingOrConsumingSession(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	settingRepo := &oauthPendingSettingRepoStub{client: client}
	require.NoError(t, settingRepo.Set(ctx, service.SettingKeyTotpEnabled, "true"))

	userRepo := newOAuthPendingUserRepo(client)
	settingSvc := newOAuthPendingSettingService(client, &config.Config{})
	authSvc := newOAuthPendingAuthService(client, userRepo, settingSvc, nil)
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

	authSvc := newOAuthPendingAuthService(client, newOAuthPendingUserRepo(client), newOAuthPendingSettingService(client, &config.Config{}), nil)

	err = completePendingOAuthBindSession(ctx, client, authSvc, session)
	require.NoError(t, err)

	identity, err := client.AuthIdentity.Query().Only(ctx)
	require.NoError(t, err)
	require.Equal(t, user.ID, identity.UserID)
	require.Equal(t, "linuxdo", identity.ProviderType)
	require.Equal(t, "subject-bind", identity.ProviderSubject)
}

func TestBindPendingOAuthLoginAppliesFirstBindGrantOnce(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	settingRepo := &oauthPendingSettingRepoStub{client: client}
	require.NoError(t, settingRepo.Set(ctx, service.SettingKeyAuthSourceDefaultLinuxDoBalance, "21.75"))
	require.NoError(t, settingRepo.Set(ctx, service.SettingKeyAuthSourceDefaultLinuxDoConcurrency, "9"))
	require.NoError(t, settingRepo.Set(ctx, service.SettingKeyAuthSourceDefaultLinuxDoSubscriptions, `[{"group_id":22,"validity_days":14}]`))
	require.NoError(t, settingRepo.Set(ctx, service.SettingKeyAuthSourceDefaultLinuxDoGrantOnFirstBind, "true"))
	settingSvc := newOAuthPendingSettingService(client, &config.Config{})
	userRepo := newOAuthPendingUserRepo(client)
	assigner := &oauthPendingDefaultSubscriptionAssignerStub{}
	authSvc := newOAuthPendingAuthService(client, userRepo, settingSvc, assigner)
	hash, err := authSvc.HashPassword("password")
	require.NoError(t, err)
	user, err := client.User.Create().
		SetEmail("first-bind@example.com").
		SetPasswordHash(hash).
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		SetConcurrency(1).
		Save(ctx)
	require.NoError(t, err)

	pendingSvc := service.NewAuthPendingIdentityService(client)
	pendingSession, err := pendingSvc.CreatePendingSession(ctx, service.CreatePendingAuthSessionInput{
		Intent: oauthIntentBindCurrentUser,
		Identity: service.PendingAuthIdentityKey{
			ProviderType:    "linuxdo",
			ProviderKey:     "linuxdo",
			ProviderSubject: "subject-first-bind",
		},
		TargetUserID:           &user.ID,
		BrowserSessionKey:      "browser-first-bind",
		UpstreamIdentityClaims: map[string]any{"email": "first-bind@linuxdo-connect.invalid"},
	})
	require.NoError(t, err)

	handler := NewAuthHandler(&config.Config{}, authSvc, service.NewUserService(userRepo, nil, nil), settingSvc, nil, nil, nil)
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
	updated, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 21.75, updated.Balance)
	require.Equal(t, 10, updated.Concurrency)
	require.Len(t, assigner.calls, 1)
	require.Equal(t, int64(22), assigner.calls[0].GroupID)

	secondSession, err := pendingSvc.CreatePendingSession(ctx, service.CreatePendingAuthSessionInput{
		Intent: oauthIntentBindCurrentUser,
		Identity: service.PendingAuthIdentityKey{
			ProviderType:    "linuxdo",
			ProviderKey:     "linuxdo",
			ProviderSubject: "subject-first-bind-second",
		},
		TargetUserID:           &user.ID,
		BrowserSessionKey:      "browser-first-bind-second",
		UpstreamIdentityClaims: map[string]any{"email": "first-bind-2@linuxdo-connect.invalid"},
	})
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/pending/bind-login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: oauthPendingSessionCookieName, Value: encodeCookieValue(secondSession.SessionToken)})
	req.AddCookie(&http.Cookie{Name: oauthPendingBrowserCookieName, Value: encodeCookieValue(secondSession.BrowserSessionKey)})
	rec = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(rec)
	c.Request = req

	handler.BindPendingOAuthLogin(c)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	updated, err = client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 21.75, updated.Balance)
	require.Equal(t, 10, updated.Concurrency)
	require.Len(t, assigner.calls, 1)
}

func TestLogin2FACompletesPendingOAuthBindSession(t *testing.T) {
	client := newOAuthPendingHandlerTestClient(t)
	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	settingRepo := &oauthPendingSettingRepoStub{client: client}
	require.NoError(t, settingRepo.Set(ctx, service.SettingKeyTotpEnabled, "true"))

	userRepo := newOAuthPendingUserRepo(client)
	settingSvc := newOAuthPendingSettingService(client, &config.Config{})
	authSvc := newOAuthPendingAuthService(client, userRepo, settingSvc, nil)
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

	identity, err := client.AuthIdentity.Query().
		Where(
			authidentity.ProviderTypeEQ("linuxdo"),
			authidentity.ProviderKeyEQ("linuxdo"),
			authidentity.ProviderSubjectEQ("subject-2fa"),
		).
		Only(ctx)
	require.NoError(t, err)
	require.Equal(t, user.ID, identity.UserID)

	emailIdentity, err := client.AuthIdentity.Query().
		Where(
			authidentity.ProviderTypeEQ("email"),
			authidentity.ProviderKeyEQ("email"),
			authidentity.ProviderSubjectEQ(user.Email),
		).
		Only(ctx)
	require.NoError(t, err)
	require.Equal(t, user.ID, emailIdentity.UserID)

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
