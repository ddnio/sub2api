package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadStudioAuthDefaultsDisabled(t *testing.T) {
	resetViperWithJWTSecret(t)

	cfg, err := Load()
	require.NoError(t, err)
	require.False(t, cfg.StudioAuth.Enabled)
	require.Equal(t, 60, cfg.StudioAuth.MaxClockSkewSeconds)
	require.Equal(t, 120, cfg.StudioAuth.NonceTTLSeconds)
	require.Equal(t, int64(1<<20), cfg.StudioAuth.MaxBodyBytes)
}

func TestLoadStudioAuthFromEnvironment(t *testing.T) {
	resetViperWithJWTSecret(t)
	t.Setenv("STUDIO_AUTH_ENABLED", "true")
	t.Setenv("STUDIO_AUTH_CURRENT_KEY_ID", "studio-2026-08")
	t.Setenv("STUDIO_AUTH_CURRENT_SECRET", "studio-current-secret-that-is-at-least-32-bytes")
	t.Setenv("STUDIO_AUTH_PREVIOUS_KEY_ID", "studio-2026-07")
	t.Setenv("STUDIO_AUTH_PREVIOUS_SECRET", "studio-previous-secret-that-is-at-least-32-bytes")

	cfg, err := Load()
	require.NoError(t, err)
	require.True(t, cfg.StudioAuth.Enabled)
	require.Equal(t, "studio-2026-08", cfg.StudioAuth.CurrentKeyID)
	require.Equal(t, "studio-current-secret-that-is-at-least-32-bytes", cfg.StudioAuth.CurrentSecret)
	require.Equal(t, "studio-2026-07", cfg.StudioAuth.PreviousKeyID)
	require.Equal(t, "studio-previous-secret-that-is-at-least-32-bytes", cfg.StudioAuth.PreviousSecret)
}

func TestLoadStudioAuthRejectsIncompleteSigningKeys(t *testing.T) {
	resetViperWithJWTSecret(t)
	t.Setenv("STUDIO_AUTH_ENABLED", "true")
	t.Setenv("STUDIO_AUTH_CURRENT_KEY_ID", "studio-2026-08")

	_, err := Load()
	require.ErrorContains(t, err, "studio_auth.current_secret")
}
