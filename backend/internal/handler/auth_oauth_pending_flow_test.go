package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
