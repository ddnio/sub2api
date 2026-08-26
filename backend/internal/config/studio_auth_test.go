package config

import (
	"strings"
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

func TestValidateStudioAuthRejectsUnsafeConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*StudioAuthConfig)
		wantErr string
	}{
		{name: "missing key id", mutate: func(c *StudioAuthConfig) { c.CurrentKeyID = "" }, wantErr: "current_key_id"},
		{name: "short secret", mutate: func(c *StudioAuthConfig) { c.CurrentSecret = "short" }, wantErr: "current_secret"},
		{name: "partial previous key", mutate: func(c *StudioAuthConfig) { c.PreviousKeyID = "previous" }, wantErr: "previous_key_id"},
		{name: "reused key id", mutate: func(c *StudioAuthConfig) {
			c.PreviousKeyID = c.CurrentKeyID
			c.PreviousSecret = strings.Repeat("p", 32)
		}, wantErr: "must differ"},
		{name: "clock skew too small", mutate: func(c *StudioAuthConfig) { c.MaxClockSkewSeconds = 0 }, wantErr: "max_clock_skew_seconds"},
		{name: "clock skew too large", mutate: func(c *StudioAuthConfig) { c.MaxClockSkewSeconds = 301 }, wantErr: "max_clock_skew_seconds"},
		{name: "nonce ttl too short", mutate: func(c *StudioAuthConfig) { c.NonceTTLSeconds = 119 }, wantErr: "nonce_ttl_seconds"},
		{name: "nonce ttl too long", mutate: func(c *StudioAuthConfig) { c.NonceTTLSeconds = 3601 }, wantErr: "nonce_ttl_seconds"},
		{name: "body too small", mutate: func(c *StudioAuthConfig) { c.MaxBodyBytes = 1023 }, wantErr: "max_body_bytes"},
		{name: "body too large", mutate: func(c *StudioAuthConfig) { c.MaxBodyBytes = 1<<20 + 1 }, wantErr: "max_body_bytes"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resetViperWithJWTSecret(t)
			cfg, err := Load()
			require.NoError(t, err)
			cfg.StudioAuth = StudioAuthConfig{
				Enabled:             true,
				CurrentKeyID:        "studio-current",
				CurrentSecret:       strings.Repeat("c", 32),
				MaxClockSkewSeconds: 60,
				NonceTTLSeconds:     120,
				MaxBodyBytes:        1 << 20,
			}
			test.mutate(&cfg.StudioAuth)
			require.ErrorContains(t, cfg.Validate(), test.wantErr)
		})
	}
}
