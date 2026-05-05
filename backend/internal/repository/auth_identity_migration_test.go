package repository

import (
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestAuthIdentityFoundationMigrationContainsExpectedBackfills(t *testing.T) {
	sqlBytes, err := migrations.FS.ReadFile("128_auth_identity_foundation.sql")
	require.NoError(t, err)

	sql := string(sqlBytes)
	requiredFragments := []string{
		"ADD COLUMN IF NOT EXISTS signup_source",
		"ADD COLUMN IF NOT EXISTS last_login_at",
		"ADD COLUMN IF NOT EXISTS last_active_at",
		"CREATE TABLE IF NOT EXISTS auth_identities",
		"CREATE TABLE IF NOT EXISTS auth_identity_channels",
		"CREATE TABLE IF NOT EXISTS pending_auth_sessions",
		"CREATE TABLE IF NOT EXISTS identity_adoption_decisions",
		"CREATE TABLE IF NOT EXISTS auth_identity_migration_reports",
		"'email'",
		"'linuxdo'",
		"'wechat'",
		"oidc_synthetic_email_requires_manual_recovery",
		"wechat_openid_only_requires_remediation",
		"ON CONFLICT (provider_type, provider_key, provider_subject) DO NOTHING",
	}

	for _, fragment := range requiredFragments {
		require.Contains(t, sql, fragment)
	}

	require.Equal(t, 3, strings.Count(sql, "INSERT INTO auth_identities"), "expected email, linuxdo, and wechat identity backfills")
}
