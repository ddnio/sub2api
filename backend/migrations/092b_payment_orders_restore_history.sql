-- 092b_payment_orders_restore_history.sql
-- Fork patch: migrate historical payment_orders data from backup table (077 schema)
-- into the new upstream v2 schema created by 092_payment_orders.sql.
--
-- Fresh-install safe: entire migration is inside a DO block; skipped if backup does not exist.
-- Idempotent: WHERE NOT EXISTS prevents duplicate inserts on retry.
--
-- NOTE: out_trade_no is NOT included here — column is added by 102, which runs after.
--       See 102a_backfill_out_trade_no_from_backup.sql for the backfill.
--
-- Data preserved: 56 real orders (24 user + 6 admin test refunds + 26 others).
-- Data not migrated (kept in backup only): credit_amount, currency, callback_raw,
--   admin_note, refund_no, refunded_at — these have no upstream v2 equivalent.

DO $$ DECLARE
  new_cnt BIGINT;
  bak_cnt BIGINT;
BEGIN
  -- Skip entirely if backup table does not exist (fresh install / no migration history)
  IF NOT EXISTS (
    SELECT 1 FROM pg_tables
    WHERE schemaname = 'public' AND tablename = 'payment_orders_v1_backup'
  ) THEN
    RAISE NOTICE '092b: payment_orders_v1_backup not found, skipping (fresh install)';
    RETURN;
  END IF;

  -- Preflight: amount precision must be <= 2 decimal places for all rows.
  IF EXISTS (
    SELECT 1 FROM payment_orders_v1_backup
    WHERE amount != ROUND(amount::numeric, 2)
    LIMIT 1
  ) THEN
    RAISE EXCEPTION 'PREFLIGHT FAILED: amount has > 2 decimal places. Aborting migration.';
  END IF;

  -- Restore historical data with field mapping.
  -- Uses LEFT JOIN so orphan orders (user deleted outside FK) are still restored with empty email/name.
  -- NOTE: out_trade_no excluded — added by 102 with DEFAULT '', backfilled by 102a.
  INSERT INTO payment_orders (
    id,
    user_id,
    user_email,
    user_name,
    amount,
    pay_amount,
    payment_type,
    payment_trade_no,
    order_type,
    plan_id,
    status,
    expires_at,
    paid_at,
    completed_at,
    client_ip,
    src_host,
    created_at,
    updated_at
  )
  SELECT
    bak.id,                                                              -- preserve original id
    bak.user_id,
    COALESCE(u.email, ''),                                               -- LEFT JOIN: orphan orders get ''
    COALESCE(u.username, ''),                                            -- LEFT JOIN: orphan orders get ''
    ROUND(bak.amount::numeric, 2),                                       -- precision: 8→2
    ROUND(bak.amount::numeric, 2),                                       -- pay_amount = amount (no fee)
    COALESCE(bak.provider, ''),                                          -- provider → payment_type
    COALESCE(bak.provider_order_no, ''),                                 -- → payment_trade_no
    CASE bak.type WHEN 'plan' THEN 'subscription' ELSE 'balance' END,   -- type → order_type
    bak.plan_id,
    UPPER(bak.status),                                                   -- lowercase → UPPERCASE
    bak.expired_at,                                                      -- expired_at → expires_at
    bak.paid_at,
    bak.completed_at,
    '',                                                                  -- client_ip: not recorded in v1
    '',                                                                  -- src_host: not recorded in v1
    bak.created_at,
    bak.updated_at
  FROM payment_orders_v1_backup bak
  LEFT JOIN users u ON bak.user_id = u.id                               -- LEFT JOIN preserves orphan orders
  WHERE NOT EXISTS (
    SELECT 1 FROM payment_orders WHERE id = bak.id                      -- idempotent guard
  );

  -- Reset BIGSERIAL sequence after explicit id inserts.
  -- Using COALESCE(..., 0) so next id = 1 on empty table (not 2).
  PERFORM setval(
    'payment_orders_id_seq',
    COALESCE((SELECT MAX(id) FROM payment_orders), 0)
  );

  -- Postcheck: all backup rows must be restored (including orphans via LEFT JOIN).
  SELECT COUNT(*) INTO new_cnt FROM payment_orders;
  SELECT COUNT(*) INTO bak_cnt FROM payment_orders_v1_backup;
  IF new_cnt != bak_cnt THEN
    RAISE EXCEPTION 'POSTCHECK FAILED: payment_orders=% but backup=%', new_cnt, bak_cnt;
  END IF;
END $$;
