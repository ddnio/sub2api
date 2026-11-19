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

func TestReleaseV0223UpgradePreservesV0220Data(t *testing.T) {
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
		if name >= "274_" {
			continue
		}
		data, readErr := migrations.FS.ReadFile(name)
		require.NoError(t, readErr)
		baseline[name] = &fstest.MapFile{Data: data}
	}
	require.Len(t, baseline, 328)
	require.Len(t, names, 333)
	require.NoError(t, applyMigrationsFS(ctx, db, baseline))
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, balance) VALUES (1, 'upgrade@example.test', 'test-only', 123.45);
		INSERT INTO groups (id, name, platform, force_openai_fast) VALUES
		(100001, 'openai fast', 'openai', true), (100002, 'composite', 'composite', false);
		INSERT INTO payment_plans (id, name, group_id, duration_days, price)
		VALUES (1, 'legacy plan', 100001, 30, 100);
		INSERT INTO payment_orders (id, user_id, amount, pay_amount, bonus_amount, expires_at, out_trade_no, plan_id)
		VALUES (1, 1, 110, 100, 10, NOW() + INTERVAL '1 day', 'upgrade_0223', 1);
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
			'groups', (SELECT jsonb_agg(jsonb_build_object('id', id, 'force_openai_fast', force_openai_fast) ORDER BY id) FROM groups),
			'checksums', (SELECT jsonb_agg(to_jsonb(m) ORDER BY filename) FROM schema_migrations m WHERE filename < '274_')
		)::text`).Scan(&value))
		return value
	}
	before := snapshot()
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))
	require.JSONEq(t, before, snapshot())
	var applied int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&applied))
	require.Equal(t, 333, applied)

	// New switches and prices default to off/unset; no existing group is moved to Ultra Fast.
	var ultrafastGroups int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM groups WHERE force_openai_ultrafast").Scan(&ultrafastGroups))
	require.Zero(t, ultrafastGroups)
	for _, column := range []struct{ table, name, defaultValue string }{
		{"usage_logs", "cache_creation_30m_tokens", "0"},
		{"usage_logs", "cache_creation_12h_tokens", "0"},
		{"groups", "force_openai_ultrafast", "false"},
	} {
		var nullable, columnDefault string
		require.NoError(t, db.QueryRowContext(ctx, `SELECT is_nullable, column_default FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2`, column.table, column.name).Scan(&nullable, &columnDefault))
		require.Equal(t, "NO", nullable, column.name)
		require.Equal(t, column.defaultValue, columnDefault, column.name)
	}
	for _, column := range []struct{ table, name string }{
		{"channel_model_pricing", "cache_write_12h_price"},
		{"channel_pricing_intervals", "cache_write_12h_price"},
		{"channel_account_stats_model_pricing", "cache_write_12h_price"},
		{"channel_account_stats_pricing_intervals", "cache_write_12h_price"},
		{"channel_model_pricing", "ultrafast_multiplier"},
	} {
		var nullable string
		var columnDefault sql.NullString
		require.NoError(t, db.QueryRowContext(ctx, `SELECT is_nullable, column_default FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2`, column.table, column.name).Scan(&nullable, &columnDefault))
		require.Equal(t, "YES", nullable, column.table+"."+column.name)
		require.False(t, columnDefault.Valid, column.table+"."+column.name)
	}
	// 274 builds CONCURRENTLY outside a transaction; the index must end up valid.
	var indexValid bool
	require.NoError(t, db.QueryRowContext(ctx, `SELECT i.indisvalid FROM pg_index i
		JOIN pg_class c ON c.oid = i.indexrelid WHERE c.relname = 'idx_usage_logs_api_key_created_cost'`).Scan(&indexValid))
	require.True(t, indexValid)
	// 277 extends the fork-renumbered 229/230 rollup state without starting a rebuild.
	var buildingUntil, buildDirtyFrom sql.NullTime
	require.NoError(t, db.QueryRowContext(ctx, `SELECT building_until, build_dirty_from FROM usage_group_rollup_state WHERE id = 1`).Scan(&buildingUntil, &buildDirtyFrom))
	require.False(t, buildingUntil.Valid)
	require.False(t, buildDirtyFrom.Valid)

	// The old migration set can reopen the upgraded database without down SQL.
	require.NoError(t, applyMigrationsFS(ctx, db, baseline))
	require.JSONEq(t, before, snapshot())
	t.Logf("upgraded %d -> %d migrations; legacy financial data and checksums preserved", len(baseline), applied)
}
