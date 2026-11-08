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

func TestReleaseV0220UpgradePreservesV0217Data(t *testing.T) {
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
		if name >= "270_" {
			continue
		}
		data, readErr := migrations.FS.ReadFile(name)
		require.NoError(t, readErr)
		baseline[name] = &fstest.MapFile{Data: data}
	}
	require.Len(t, baseline, 324)
	require.Len(t, names, 328)
	require.NoError(t, applyMigrationsFS(ctx, db, baseline))
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, balance) VALUES (1, 'upgrade@example.test', 'test-only', 123.45);
		INSERT INTO groups (id, name, platform, allow_image_generation) VALUES
		(100001, 'gemini off', 'gemini', false), (100002, 'antigravity off', 'antigravity', false),
		(100003, 'composite off', 'composite', false), (100004, 'openai off', 'openai', false),
		(100005, 'gemini on', 'gemini', true);
		INSERT INTO payment_plans (id, name, group_id, duration_days, price)
		VALUES (1, 'legacy plan', 100001, 30, 100);
		INSERT INTO payment_orders (id, user_id, amount, pay_amount, bonus_amount, expires_at, out_trade_no, plan_id)
		VALUES (1, 1, 110, 100, 10, NOW() + INTERVAL '1 day', 'upgrade_0220', 1);
		INSERT INTO user_platform_quotas (user_id, platform, daily_limit_usd) VALUES (1, 'openai', 25);
	`)
	require.NoError(t, err)
	snapshot := func() string {
		var value string
		require.NoError(t, db.QueryRowContext(ctx, `SELECT jsonb_build_object(
			'users', (SELECT jsonb_agg(to_jsonb(u) ORDER BY id) FROM users u),
			'orders', (SELECT jsonb_agg(to_jsonb(o) ORDER BY id) FROM payment_orders o),
			'plans', (SELECT jsonb_agg(to_jsonb(p) ORDER BY id) FROM payment_plans p),
			'quotas', (SELECT jsonb_agg(to_jsonb(q) ORDER BY id) FROM user_platform_quotas q),
			'checksums', (SELECT jsonb_agg(to_jsonb(m) ORDER BY filename) FROM schema_migrations m WHERE filename < '270_')
		)::text`).Scan(&value))
		return value
	}
	before := snapshot()
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))
	require.JSONEq(t, before, snapshot())
	var applied int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&applied))
	require.Equal(t, 328, applied)
	for id, allowed := range map[int]bool{100001: true, 100002: true, 100003: false, 100004: false, 100005: true} {
		var imageGeneration, imageCountBilling bool
		require.NoError(t, db.QueryRowContext(ctx, "SELECT allow_image_generation, image_count_billing FROM groups WHERE id=$1", id).Scan(&imageGeneration, &imageCountBilling))
		require.Equal(t, allowed, imageGeneration, "group %d", id)
		require.False(t, imageCountBilling, "group %d", id)
	}
	for _, table := range []string{"scoped_admin_api_keys", "mcp_oauth_grants", "mcp_oauth_codes", "mcp_oauth_credentials"} {
		var count int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count))
		require.Zero(t, count, table)
	}
	// The old migration set can reopen the upgraded database without down SQL.
	// This checks schema/checksum compatibility, not new authorization semantics.
	require.NoError(t, applyMigrationsFS(ctx, db, baseline))
	require.JSONEq(t, before, snapshot())
	t.Logf("upgraded %d -> %d migrations; legacy financial data and checksums preserved", len(baseline), applied)
}
