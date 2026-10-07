//go:build integration

package repository

import (
	"context"
	"database/sql"
	"io/fs"
	"net/url"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestReleaseV0214UpgradePreservesV028Data(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pg, err := tcpostgres.Run(ctx, selectDockerImage(ctx, postgresImageTag),
		tcpostgres.WithDatabase("release_upgrade"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies())
	require.NoError(t, err)
	t.Cleanup(func() { _ = pg.Terminate(context.Background()) })
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	u, err := url.Parse(dsn)
	require.NoError(t, err)
	query := u.Query()
	query.Set("options", "-c lock_timeout=5000")
	u.RawQuery = query.Encode()
	db, err := sql.Open("postgres", u.String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// Existing migrations are unchanged; omitting only the new pair reproduces
	// the v0.2.8 migration set without importing production data.
	baseline := fstest.MapFS{}
	names, err := fs.Glob(migrations.FS, "*.sql")
	require.NoError(t, err)
	for _, name := range names {
		if name == "263_add_payment_order_bonus_amount.sql" || name == "264_add_typesafe_platform.sql" {
			continue
		}
		content, readErr := migrations.FS.ReadFile(name)
		require.NoError(t, readErr)
		baseline[name] = &fstest.MapFile{Data: content}
	}
	require.Len(t, baseline, 317)
	require.NoError(t, applyMigrationsFS(ctx, db, baseline))
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash) VALUES (1, 'upgrade@example.test', 'test-only');
		INSERT INTO groups (id, name) VALUES (100001, 'upgrade');
		INSERT INTO payment_plans (id, name, group_id, duration_days, price)
		VALUES (1, 'legacy plan', 100001, 30, 100);
		INSERT INTO payment_orders (id, user_id, amount, pay_amount, expires_at, out_trade_no, plan_id)
		VALUES (1, 1, 100, 100, NOW() + INTERVAL '1 day', 'upgrade_legacy', 1);
		INSERT INTO user_platform_quotas (user_id, platform, daily_limit_usd)
		VALUES (1, 'openai', 25);
		INSERT INTO composite_model_routes (group_id, public_model, target_platform)
		VALUES (100001, 'legacy-model', 'openai');
	`)
	require.NoError(t, err)
	var foreignKeysBefore int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pg_constraint
		WHERE conrelid = 'payment_orders'::regclass AND contype = 'f'`).Scan(&foreignKeysBefore))

	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))
	var applied int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&applied))
	require.Equal(t, 319, applied)
	var bonus, amount, limit float64
	var planID int64
	require.NoError(t, db.QueryRowContext(ctx,
		"SELECT bonus_amount, amount, plan_id FROM payment_orders WHERE id = 1").Scan(&bonus, &amount, &planID))
	require.Zero(t, bonus)
	require.Equal(t, 100.0, amount)
	require.EqualValues(t, 1, planID)
	require.NoError(t, db.QueryRowContext(ctx,
		"SELECT daily_limit_usd FROM user_platform_quotas WHERE user_id = 1 AND platform = 'openai'").Scan(&limit))
	require.Equal(t, 25.0, limit)
	var foreignKeysAfter int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pg_constraint
		WHERE conrelid = 'payment_orders'::regclass AND contype = 'f'`).Scan(&foreignKeysAfter))
	require.Equal(t, foreignKeysBefore, foreignKeysAfter)
	_, err = db.ExecContext(ctx, `
		INSERT INTO user_platform_quotas (user_id, platform, daily_limit_usd)
		VALUES (1, 'typesafe', 10);
		INSERT INTO composite_model_routes (group_id, public_model, target_platform)
		VALUES (100001, 'typesafe-model', 'typesafe');
		INSERT INTO payment_orders (id, user_id, amount, pay_amount, bonus_amount, expires_at, out_trade_no)
		VALUES (2, 1, 110, 100, 10, NOW() + INTERVAL '1 day', 'upgrade_bonus');
	`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO user_platform_quotas (user_id, platform) VALUES (1, 'invalid-platform')")
	require.Error(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO composite_model_routes (group_id, public_model, target_platform)
		VALUES (100001, 'invalid-model', 'invalid-platform')`)
	require.Error(t, err)
	t.Logf("upgraded %d baseline migrations to %d; existing order, plan, quota and foreign keys preserved", len(baseline), applied)
}
