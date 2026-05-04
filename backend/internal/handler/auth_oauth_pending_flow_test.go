package handler

import (
	"context"
	"database/sql"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
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
