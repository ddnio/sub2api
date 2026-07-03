package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration112UsesIdempotentAddColumn(t *testing.T) {
	content, err := FS.ReadFile("112_add_payment_order_provider_key_snapshot.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS provider_key VARCHAR(30)")
	require.NotContains(t, sql, "ADD COLUMN provider_key VARCHAR(30);")
}

func TestMigration118DoesNotForceOverwriteAuthSourceGrantDefaults(t *testing.T) {
	content, err := FS.ReadFile("118_wechat_dual_mode_and_auth_source_defaults.sql")
	require.NoError(t, err)

	sql := string(content)
	require.NotContains(t, sql, "UPDATE settings")
	require.NotContains(t, sql, "SET value = 'false'")
	require.True(t, strings.Contains(sql, "ON CONFLICT (key) DO NOTHING"))
	require.Contains(t, sql, "THEN ''")
}

func TestAuthIdentityReportTypeWideningRunsBeforeLongReportWritersAndStillReconcilesAt121(t *testing.T) {
	preflightContent, err := FS.ReadFile("108a_widen_auth_identity_migration_report_type.sql")
	require.NoError(t, err)

	preflightSQL := string(preflightContent)
	require.Contains(t, preflightSQL, "ALTER TABLE auth_identity_migration_reports")
	require.Contains(t, preflightSQL, "ALTER COLUMN report_type TYPE VARCHAR(80)")

	content, err := FS.ReadFile("109_auth_identity_compat_backfill.sql")
	require.NoError(t, err)

	sql := string(content)
	require.NotContains(t, sql, "ALTER TABLE auth_identity_migration_reports")

	followupContent, err := FS.ReadFile("121_auth_identity_migration_report_type_widen.sql")
	require.NoError(t, err)

	followupSQL := string(followupContent)
	require.Contains(t, followupSQL, "ALTER TABLE auth_identity_migration_reports")
	require.Contains(t, followupSQL, "ALTER COLUMN report_type TYPE VARCHAR(80)")
}

func TestMigration119DefersPaymentIndexRolloutToOnlineFollowup(t *testing.T) {
	content, err := FS.ReadFile("119_enforce_payment_orders_out_trade_no_unique.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "120_enforce_payment_orders_out_trade_no_unique_notx.sql")
	require.Contains(t, sql, "NULL;")
	require.NotContains(t, sql, "CREATE UNIQUE INDEX")
	require.NotContains(t, sql, "DROP INDEX")

	followupContent, err := FS.ReadFile("120_enforce_payment_orders_out_trade_no_unique_notx.sql")
	require.NoError(t, err)

	followupSQL := string(followupContent)
	require.Contains(t, followupSQL, "explicit duplicate out_trade_no precheck")
	require.Contains(t, followupSQL, "stale invalid paymentorder_out_trade_no_unique index")
	require.Contains(t, followupSQL, "CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS paymentorder_out_trade_no_unique")
	require.NotContains(t, followupSQL, "DROP INDEX CONCURRENTLY IF EXISTS paymentorder_out_trade_no_unique")
	require.Contains(t, followupSQL, "DROP INDEX CONCURRENTLY IF EXISTS paymentorder_out_trade_no")
	require.Contains(t, followupSQL, "WHERE out_trade_no <> ''")

	alignmentContent, err := FS.ReadFile("120a_align_payment_orders_out_trade_no_index_name.sql")
	require.NoError(t, err)

	alignmentSQL := string(alignmentContent)
	require.Contains(t, alignmentSQL, "paymentorder_out_trade_no_unique")
	require.Contains(t, alignmentSQL, "RENAME TO paymentorder_out_trade_no")
}

func TestMigration136BackfillsReferralRelationshipsWithoutMoneyState(t *testing.T) {
	content, err := FS.ReadFile("136_migrate_referrals_to_affiliates.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "duplicate referral_code")
	require.Contains(t, sql, "invalid referral_code")
	require.Contains(t, sql, "conflicts with existing aff_code")
	require.Contains(t, sql, "existing affiliate inviter conflicts")
	require.Contains(t, sql, "referral participants without affiliate rows or referral_code")
	require.Contains(t, sql, "INSERT INTO user_affiliates")
	require.Contains(t, sql, "UPDATE user_affiliates ua")
	require.Contains(t, sql, "SET inviter_id = r.inviter_id")
	require.Contains(t, sql, "SET aff_count = COALESCE(counts.cnt, 0)")

	require.NotContains(t, sql, "INSERT INTO user_affiliate_ledger")
	require.NotContains(t, sql, "UPDATE user_affiliate_ledger")
	require.NotContains(t, sql, "aff_quota =")
	require.NotContains(t, sql, "aff_history_quota =")
	require.NotContains(t, sql, "aff_frozen_quota =")
	require.NotContains(t, sql, "UPDATE users")
	require.NotContains(t, sql, "UPDATE redeem_codes")
	require.NotContains(t, sql, "UPDATE user_referrals")
}

func TestMigration110SeedsAuthSourceSignupGrantsDisabledByDefault(t *testing.T) {
	content, err := FS.ReadFile("110_pending_auth_and_provider_default_grants.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "('auth_source_default_email_grant_on_signup', 'false')")
	require.Contains(t, sql, "('auth_source_default_linuxdo_grant_on_signup', 'false')")
	require.Contains(t, sql, "('auth_source_default_oidc_grant_on_signup', 'false')")
	require.Contains(t, sql, "('auth_source_default_wechat_grant_on_signup', 'false')")
	require.NotContains(t, sql, "('auth_source_default_email_grant_on_signup', 'true')")
}

func TestMigration122ScrubsPendingOAuthCompletionTokensAtRest(t *testing.T) {
	content, err := FS.ReadFile("122_pending_auth_completion_token_cleanup.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "UPDATE pending_auth_sessions")
	require.Contains(t, sql, "completion_response")
	require.Contains(t, sql, "access_token")
	require.Contains(t, sql, "refresh_token")
	require.Contains(t, sql, "expires_in")
	require.Contains(t, sql, "token_type")
}

func TestMigration123BackfillsLegacyAuthSourceGrantDefaultsSafely(t *testing.T) {
	content, err := FS.ReadFile("123_fix_legacy_auth_source_grant_on_signup_defaults.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "110_pending_auth_and_provider_default_grants.sql")
	require.Contains(t, sql, "schema_migrations")
	require.Contains(t, sql, "updated_at")
	require.Contains(t, sql, "'_grant_on_signup'")
	require.Contains(t, sql, "value = 'false'")
	require.Contains(t, sql, "auth_identity_migration_reports")
}

func TestMigration124BackfillsLegacyOIDCSecurityFlagsSafely(t *testing.T) {
	content, err := FS.ReadFile("124_backfill_legacy_oidc_security_flags.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "oidc_connect_use_pkce")
	require.Contains(t, sql, "oidc_connect_validate_id_token")
	require.Contains(t, sql, "ON CONFLICT (key) DO NOTHING")
	require.Contains(t, sql, "oidc_connect_enabled")
	require.Contains(t, sql, "'false'")
}

func TestMigration131KeepsPaymentAuditUniquenessScopedToAffiliateClaims(t *testing.T) {
	content, err := FS.ReadFile("131_affiliate_rebate_hardening.sql")
	require.NoError(t, err)

	sql := string(content)
	require.NotContains(t, sql, "idx_payment_audit_logs_order_action_uniq")
	require.Contains(t, sql, "CREATE UNIQUE INDEX IF NOT EXISTS uq_payment_audit_logs_affiliate_claim")
	require.Contains(t, sql, "WHERE action IN ('AFFILIATE_REBATE_APPLIED', 'AFFILIATE_REBATE_SKIPPED')")

	rankedStart := strings.Index(sql, "WITH ranked AS")
	deleteStart := strings.Index(sql, "DELETE FROM payment_audit_logs")
	require.NotEqual(t, -1, rankedStart)
	require.NotEqual(t, -1, deleteStart)
	require.Less(t, rankedStart, deleteStart)
	require.Contains(t, sql[rankedStart:deleteStart], "WHERE action IN ('AFFILIATE_REBATE_APPLIED', 'AFFILIATE_REBATE_SKIPPED')")
}

func TestMigration132UpgradesClaudeCodeMonitorTemplateSnapshots(t *testing.T) {
	content, err := FS.ReadFile("132_update_claude_code_monitor_template.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "UPDATE channel_monitor_request_templates")
	require.Contains(t, sql, "UPDATE channel_monitors")
	require.Contains(t, sql, "claude-cli/2.1.92 (external, cli)")
	require.Contains(t, sql, "oauth-2025-04-20")
	require.Contains(t, sql, "effort-2025-11-24")
	require.Contains(t, sql, "redact-thinking-2026-02-12")
	require.Contains(t, sql, "extended-cache-ttl-2025-04-11")
	require.Contains(t, sql, "advisor-tool-2026-03-01")
	require.Contains(t, sql, "jsonb_set")
	require.NotContains(t, sql, "SET body_override")
}

func TestMigration134AddsAffiliateLedgerAuditFieldsWithoutJSONCast(t *testing.T) {
	content, err := FS.ReadFile("134_affiliate_ledger_audit_snapshots.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS source_order_id BIGINT")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS balance_after DECIMAL(20,8)")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS aff_quota_after DECIMAL(20,8)")
	require.Contains(t, sql, "substring(")
	require.Contains(t, sql, `"rebateAmount"`)
	require.Contains(t, sql, "COUNT(*) OVER (PARTITION BY ra.order_id) AS order_match_count")
	require.Contains(t, sql, "COUNT(*) OVER (PARTITION BY ual.id) AS ledger_match_count")
	require.NotContains(t, sql, "detail::jsonb")
}

func TestMigration135AllowsGitHubAndGoogleAuthProviders(t *testing.T) {
	content, err := FS.ReadFile("135_allow_email_oauth_provider_types.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "users_signup_source_check")
	require.Contains(t, sql, "auth_identities_provider_type_check")
	require.Contains(t, sql, "auth_identity_channels_provider_type_check")
	require.Contains(t, sql, "pending_auth_sessions_provider_type_check")
	require.Contains(t, sql, "'github'")
	require.Contains(t, sql, "'google'")
}

func TestMigration154AddsAccountAutoPauseExpiryPartialIndex(t *testing.T) {
	content, err := FS.ReadFile("154_account_autopause_expiry_index_notx.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_autopause_expiry_due")
	require.Contains(t, sql, "ON accounts (expires_at)")
	require.Contains(t, sql, "WHERE deleted_at IS NULL")
	require.Contains(t, sql, "schedulable = TRUE")
	require.Contains(t, sql, "auto_pause_on_expired = TRUE")
	require.Contains(t, sql, "expires_at IS NOT NULL")
}

func TestMigration162AddsSparkShadowColumnsAndConstraintsWithoutHotIndexes(t *testing.T) {
	content, err := FS.ReadFile("162_account_spark_shadow.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS parent_account_id BIGINT")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS quota_dimension VARCHAR(20) NOT NULL DEFAULT 'global'")
	require.Contains(t, sql, "chk_accounts_parent_dimension")
	// 约束已放开为「影子 ⇒ 非 global 维度」（spark 不再写死进 parent 约束）
	require.Contains(t, sql, "parent_account_id IS NOT NULL AND quota_dimension <> 'global'")
	require.NotContains(t, sql, "parent_account_id IS NOT NULL AND quota_dimension = 'spark'")
	require.Contains(t, sql, "chk_accounts_parent_not_self")
	require.Contains(t, sql, "fk_accounts_parent_account_id")
	require.Contains(t, sql, "FOREIGN KEY (parent_account_id) REFERENCES accounts(id)")
	require.Contains(t, sql, "ON DELETE RESTRICT")
	require.Contains(t, sql, "NOT VALID")
	require.NotContains(t, sql, "CREATE INDEX")
	require.NotContains(t, sql, "CREATE UNIQUE INDEX")
	require.NotContains(t, sql, "CONCURRENTLY")
}

func TestMigration163AddsSparkShadowIndexesConcurrently(t *testing.T) {
	content, err := FS.ReadFile("163_account_spark_shadow_indexes_notx.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_parent_account_id")
	require.Contains(t, sql, "CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uq_accounts_spark_shadow_per_parent")
	require.Contains(t, sql, "ON accounts (parent_account_id)")
	require.Contains(t, sql, "WHERE parent_account_id IS NOT NULL")
	require.Contains(t, sql, "quota_dimension = 'spark'")
	require.Contains(t, sql, "deleted_at IS NULL")
}

func TestMigration164AddsGroupPeakRateMultiplierColumns(t *testing.T) {
	content, err := FS.ReadFile("164_add_group_peak_rate_multiplier.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS peak_rate_enabled")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS peak_start")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS peak_end")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS peak_rate_multiplier")
}

func TestMigration165BackfillsGrokMediaGenerationGroups(t *testing.T) {
	content, err := FS.ReadFile("165_enable_grok_media_generation_groups.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "UPDATE groups")
	require.Contains(t, sql, "SET allow_image_generation = true")
	require.Contains(t, sql, "WHERE platform = 'grok'")
	require.Contains(t, sql, "AND allow_image_generation = false")
}
