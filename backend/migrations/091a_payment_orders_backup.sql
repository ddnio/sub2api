-- 091a_payment_orders_backup.sql
-- Fork patch: backup existing payment_orders (077 schema) before upstream v2 rebuild.
-- Must run BEFORE 092_payment_orders.sql so the new CREATE TABLE gets a clean namespace.
--
-- Safety: CREATE TABLE ... AS SELECT does NOT copy constraints/indexes/sequences,
-- so old constraint names (payment_orders_pkey, order_no_key, etc.) are freed after DROP.
-- This allows 092_payment_orders.sql to create payment_orders_pkey without conflict.
--
-- Idempotent: safe to re-run if migration was interrupted.

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_tables
    WHERE schemaname = 'public' AND tablename = 'payment_orders_v1_backup'
  ) THEN
    EXECUTE 'CREATE TABLE payment_orders_v1_backup AS SELECT * FROM payment_orders';
  END IF;
END $$;

-- DROP releases all constraint/index names (pkey, unique, fk) associated with the table.
DROP TABLE IF EXISTS payment_orders;
