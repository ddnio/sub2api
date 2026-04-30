-- 096a_payment_provider_instances_unique.sql
-- Fork patch: upstream 096 does not add a unique constraint on (provider_key, name).
-- We need it so ON CONFLICT (provider_key, name) works in config migration scripts.
-- Runs after 096_payment_provider_instances.sql creates the table.

ALTER TABLE payment_provider_instances
  ADD CONSTRAINT uq_payment_provider_instances_provider_name
  UNIQUE (provider_key, name);
