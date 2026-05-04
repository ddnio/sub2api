-- Balance and quota notification user preferences.
-- Upstream introduced these as 101/102/104. The fork already has later
-- migration numbers, so keep the final schema in one forward migration.
ALTER TABLE users ADD COLUMN IF NOT EXISTS balance_notify_enabled BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS balance_notify_threshold DECIMAL(20,8) DEFAULT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS balance_notify_extra_emails TEXT NOT NULL DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS balance_notify_threshold_type VARCHAR(10) NOT NULL DEFAULT 'fixed';
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_recharged DECIMAL(20,8) NOT NULL DEFAULT 0;

DO $$
DECLARE
  user_row RECORD;
  parsed jsonb;
  converted text;
BEGIN
  FOR user_row IN
    SELECT id, balance_notify_extra_emails
    FROM users
    WHERE balance_notify_extra_emails IS NOT NULL
      AND balance_notify_extra_emails <> '[]'
      AND balance_notify_extra_emails <> ''
  LOOP
    BEGIN
      parsed := user_row.balance_notify_extra_emails::jsonb;
      IF jsonb_typeof(parsed) = 'array'
        AND jsonb_array_length(parsed) > 0
        AND jsonb_typeof(parsed -> 0) = 'string'
      THEN
        SELECT COALESCE(
          jsonb_agg(jsonb_build_object('email', elem::text, 'disabled', false, 'verified', false)),
          '[]'::jsonb
        )::text
        INTO converted
        FROM jsonb_array_elements_text(parsed) AS elem;

        UPDATE users
        SET balance_notify_extra_emails = converted
        WHERE id = user_row.id;
      END IF;
    EXCEPTION WHEN others THEN
      RAISE NOTICE 'skip legacy users.balance_notify_extra_emails conversion for user id %: %', user_row.id, SQLERRM;
    END;
  END LOOP;
END $$;

DO $$
DECLARE
  setting_row RECORD;
  parsed jsonb;
  converted text;
BEGIN
  FOR setting_row IN
    SELECT key, value
    FROM settings
    WHERE key = 'account_quota_notify_emails'
      AND value IS NOT NULL
      AND value <> '[]'
      AND value <> ''
  LOOP
    BEGIN
      parsed := setting_row.value::jsonb;
      IF jsonb_typeof(parsed) = 'array'
        AND jsonb_array_length(parsed) > 0
        AND jsonb_typeof(parsed -> 0) = 'string'
      THEN
        SELECT COALESCE(
          jsonb_agg(jsonb_build_object('email', elem::text, 'disabled', false, 'verified', false)),
          '[]'::jsonb
        )::text
        INTO converted
        FROM jsonb_array_elements_text(parsed) AS elem;

        UPDATE settings
        SET value = converted
        WHERE key = setting_row.key;
      END IF;
    EXCEPTION WHEN others THEN
      RAISE NOTICE 'skip legacy settings.account_quota_notify_emails conversion for key %: %', setting_row.key, SQLERRM;
    END;
  END LOOP;
END $$;
