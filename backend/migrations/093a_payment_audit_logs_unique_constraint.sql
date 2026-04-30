-- 093a_payment_audit_logs_unique_constraint.sql
-- Fork patch: add UNIQUE(order_id, action) constraint to payment_audit_logs.
-- Required by payment_fulfillment.go tryClaimAffiliateRebateAudit which uses:
--   ON CONFLICT (order_id, action) DO NOTHING
-- Without this constraint, ON CONFLICT raises:
--   ERROR: there is no unique or exclusion constraint matching the ON CONFLICT specification
-- Idempotent: checks pg_constraint before adding.
-- Note: ADD CONSTRAINT IF NOT EXISTS is NOT supported in PostgreSQL; use pg_constraint check.

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'uq_payment_audit_logs_order_action'
      AND conrelid = 'payment_audit_logs'::regclass
  ) THEN
    ALTER TABLE payment_audit_logs
      ADD CONSTRAINT uq_payment_audit_logs_order_action
      UNIQUE (order_id, action);
  END IF;
END $$;
