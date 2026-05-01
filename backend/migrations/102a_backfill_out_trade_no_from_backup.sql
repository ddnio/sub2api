-- 102a_backfill_out_trade_no_from_backup.sql
-- Fork patch: backfill out_trade_no for historical orders migrated in 092b.
-- Runs AFTER 102_add_out_trade_no_to_payment_orders.sql adds the column.
--
-- 092b could not include out_trade_no because the column didn't exist yet at 092.
-- This migration fills in the value from the backup table. If the legacy backup
-- row has no order_no, fall back to the historical sub2_<id> format so late
-- legacy callbacks can still be matched by out_trade_no.
-- Fresh-install safe: skipped if backup table does not exist.
-- Idempotent: WHERE out_trade_no = '' is safe to re-run.

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_tables
    WHERE schemaname = 'public' AND tablename = 'payment_orders_v1_backup'
  ) THEN
    RAISE NOTICE '102a: payment_orders_v1_backup not found, skipping (fresh install)';
    RETURN;
  END IF;

  UPDATE payment_orders po
  SET out_trade_no = COALESCE(NULLIF(bak.order_no, ''), 'sub2_' || po.id::text)
  FROM payment_orders_v1_backup bak
  WHERE po.id = bak.id
    AND po.out_trade_no = '';
END $$;
