-- 093a_payment_audit_logs_unique_constraint.sql
-- Fork patch: add partial unique constraint to payment_audit_logs for affiliate rebate claim.
-- Required by payment_fulfillment.go tryClaimAffiliateRebateAudit which uses:
--   ON CONFLICT (order_id, action) DO NOTHING
--
-- IMPORTANT: Full UNIQUE(order_id, action) would break audit logs that intentionally
-- write the same action multiple times (e.g., AFFILIATE_REBATE_FAILED, RECHARGE_RETRY).
-- Solution: partial unique index covering only the two affiliate claim actions.
--
-- Idempotent: checks pg_class/pg_index before adding.
-- Note: ADD CONSTRAINT IF NOT EXISTS is NOT supported in PostgreSQL.

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE c.relname = 'uq_payment_audit_logs_affiliate_claim'
      AND n.nspname = 'public'
  ) THEN
    EXECUTE $idx$
      CREATE UNIQUE INDEX uq_payment_audit_logs_affiliate_claim
      ON payment_audit_logs (order_id, action)
      WHERE action IN ('AFFILIATE_REBATE_APPLIED', 'AFFILIATE_REBATE_SKIPPED')
    $idx$;
  END IF;
END $$;
