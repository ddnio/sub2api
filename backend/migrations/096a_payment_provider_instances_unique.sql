-- 096a_payment_provider_instances_unique.sql
-- Fork patch: upstream 096 does not add a unique constraint on (provider_key, name).
-- We need it so ON CONFLICT (provider_key, name) works in config migration scripts.
-- Runs after 096_payment_provider_instances.sql creates the table.
-- Idempotent: pg_constraint check prevents duplicate constraint error on re-run.

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'uq_payment_provider_instances_provider_name'
      AND conrelid = 'payment_provider_instances'::regclass
  ) THEN
    ALTER TABLE payment_provider_instances
      ADD CONSTRAINT uq_payment_provider_instances_provider_name
      UNIQUE (provider_key, name);
  END IF;
END $$;
