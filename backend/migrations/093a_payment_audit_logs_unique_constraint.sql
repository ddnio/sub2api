-- 093a_payment_audit_logs_unique_constraint.sql
-- Fork patch: add UNIQUE(order_id, action) constraint to payment_audit_logs.
-- Required by payment_fulfillment.go tryClaimAffiliateRebateAudit which uses:
--   ON CONFLICT (order_id, action) DO NOTHING
-- Without this constraint, ON CONFLICT would raise:
--   ERROR: there is no unique or exclusion constraint matching the ON CONFLICT specification
-- Idempotent: IF NOT EXISTS guard.

ALTER TABLE payment_audit_logs
  ADD CONSTRAINT IF NOT EXISTS uq_payment_audit_logs_order_action
  UNIQUE (order_id, action);
