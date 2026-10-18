//go:build integration

package repository

import (
	"context"
	"database/sql"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestReleaseV0217UpgradePreservesV0214Data(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pg, err := tcpostgres.Run(ctx, selectDockerImage(ctx, postgresImageTag),
		tcpostgres.WithDatabase("release_upgrade"), tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"), tcpostgres.BasicWaitStrategies())
	require.NoError(t, err)
	t.Cleanup(func() { _ = pg.Terminate(context.Background()) })
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable&options=-c%20lock_timeout%3D5000")
	require.NoError(t, err)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	baseline := fstest.MapFS{}
	names, err := fs.Glob(migrations.FS, "*.sql")
	require.NoError(t, err)
	for _, name := range names {
		if name >= "265_" {
			continue
		}
		content, readErr := migrations.FS.ReadFile(name)
		require.NoError(t, readErr)
		baseline[name] = &fstest.MapFile{Data: content}
	}
	require.Len(t, baseline, 319)
	require.NoError(t, applyMigrationsFS(ctx, db, baseline))
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash) VALUES (1, 'upgrade@example.test', 'test-only');
		INSERT INTO groups (id, name) VALUES (100001, 'legacy group');
		INSERT INTO payment_plans (id, name, group_id, duration_days, price)
		VALUES (1, 'legacy plan', 100001, 30, 100);
		INSERT INTO payment_orders (id, user_id, amount, pay_amount, bonus_amount, expires_at, out_trade_no, plan_id)
		VALUES (1, 1, 110, 100, 10, NOW() + INTERVAL '1 day', 'upgrade_0217', 1);
		INSERT INTO user_platform_quotas (user_id, platform, daily_limit_usd) VALUES (1, 'openai', 25);
		INSERT INTO composite_model_routes (group_id, public_model, target_platform)
		VALUES (100001, 'legacy-model', 'openai');
	`)
	require.NoError(t, err)
	checksum := func() string {
		var fingerprint string
		require.NoError(t, db.QueryRowContext(ctx, `SELECT md5(string_agg(filename || ':' || checksum, ',' ORDER BY filename))
			FROM schema_migrations WHERE filename < '265_'`).Scan(&fingerprint))
		return fingerprint
	}
	before := checksum()
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))
	require.Equal(t, before, checksum())
	var applied, constraints int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&applied))
	require.Equal(t, 324, applied)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pg_constraint
		WHERE conname IN ('user_platform_quotas_platform_check', 'composite_model_routes_target_platform_check')`).Scan(&constraints))
	require.Zero(t, constraints)
	var amount, payAmount, bonus, quota float64
	var planID int64
	require.NoError(t, db.QueryRowContext(ctx, `SELECT amount, pay_amount, bonus_amount, plan_id
		FROM payment_orders WHERE id = 1`).Scan(&amount, &payAmount, &bonus, &planID))
	require.Equal(t, 110.0, amount)
	require.Equal(t, 100.0, payAmount)
	require.Equal(t, 10.0, bonus)
	require.EqualValues(t, 1, planID)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT daily_limit_usd FROM user_platform_quotas
		WHERE user_id = 1 AND platform = 'openai'`).Scan(&quota))
	require.Equal(t, 25.0, quota)
	var i18n string
	require.NoError(t, db.QueryRowContext(ctx, "SELECT i18n::text FROM groups WHERE id = 100001").Scan(&i18n))
	require.JSONEq(t, `{}`, i18n)

	// New catalog platforms must remain writable after all historical migrations.
	_, err = db.ExecContext(ctx, `
		INSERT INTO user_platform_quotas (user_id, platform) VALUES (1, 'cline');
		INSERT INTO composite_model_routes (group_id, public_model, target_platform)
		VALUES (100001, 'command-model', 'command_code');
		INSERT INTO gemini_cached_contents
		(public_id, user_id, api_key_id, group_id, account_id, upstream_name, model, upstream_model, expire_time)
		VALUES ('test-cache', 1, 1, 100001, 1, 'cachedContents/test', 'gemini-test', 'gemini-test', NOW() + INTERVAL '1 hour');
	`)
	require.NoError(t, err)
	var channelUsage string
	var mutationToken sql.NullString
	require.NoError(t, db.QueryRowContext(ctx, `SELECT channel_usage::text, mutation_token
		FROM gemini_cached_contents WHERE public_id = 'test-cache'`).Scan(&channelUsage, &mutationToken))
	require.JSONEq(t, `{}`, channelUsage)
	require.False(t, mutationToken.Valid)
	t.Logf("upgraded %d baseline migrations to %d without changing historical checksums or payment/quota data", len(baseline), applied)
}
