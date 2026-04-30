-- 092b_payment_orders_restore_history.sql
-- Fork patch: migrate historical payment_orders data from backup table (077 schema)
-- into the new upstream v2 schema created by 092_payment_orders.sql.
--
-- Runs AFTER 092_payment_orders.sql creates the fresh table.
-- Idempotent: WHERE NOT EXISTS prevents duplicate inserts on retry.
--
-- Data preserved: 56 real orders (24 user + 6 admin test refunds + 26 others).
-- Data not migrated (kept in backup only): credit_amount, currency, callback_raw,
--   admin_note, refund_no, refunded_at — these have no upstream v2 equivalent.

-- Preflight: amount precision must be <= 2 decimal places for all rows.
DO $$ BEGIN
  IF EXISTS (
    SELECT 1 FROM payment_orders_v1_backup
    WHERE amount != ROUND(amount::numeric, 2)
    LIMIT 1
  ) THEN
    RAISE EXCEPTION 'PREFLIGHT FAILED: amount has > 2 decimal places. Aborting migration.';
  END IF;
END $$;

-- Restore historical data with field mapping.
-- NOTE: out_trade_no is NOT included here — column is added by 102_add_out_trade_no_to_payment_orders.sql
-- which runs after this file. The DEFAULT '' in 102 is equivalent to COALESCE(bak.order_no, '').
-- To backfill out_trade_no from backup, see the UPDATE below (runs after 102).
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
  u.email,                                                             -- from users table
  u.username,                                                          -- from users table
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
JOIN users u ON bak.user_id = u.id
WHERE NOT EXISTS (
  SELECT 1 FROM payment_orders WHERE id = bak.id                      -- idempotent guard
);

-- Reset BIGSERIAL sequence after explicit id inserts.
-- Without this, next auto-generated id starts at 1 and conflicts with restored ids.
SELECT setval(
  'payment_orders_id_seq',
  COALESCE((SELECT MAX(id) FROM payment_orders), 1)
);

-- Postcheck: row counts must match.
DO $$ DECLARE
  new_cnt  BIGINT;
  bak_cnt  BIGINT;
BEGIN
  SELECT COUNT(*) INTO new_cnt FROM payment_orders;
  SELECT COUNT(*) INTO bak_cnt FROM payment_orders_v1_backup;
  IF new_cnt != bak_cnt THEN
    RAISE EXCEPTION 'POSTCHECK FAILED: payment_orders=% but backup=%', new_cnt, bak_cnt;
  END IF;
END $$;
