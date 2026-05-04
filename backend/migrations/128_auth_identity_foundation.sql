ALTER TABLE users
ADD COLUMN IF NOT EXISTS signup_source VARCHAR(20) NOT NULL DEFAULT 'email',
ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ NULL,
ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMPTZ NULL;

UPDATE users
SET signup_source = 'email'
WHERE signup_source IS NULL OR signup_source = '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'users_signup_source_check'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT users_signup_source_check
            CHECK (signup_source IN ('email', 'linuxdo', 'wechat', 'oidc'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS auth_identities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_type VARCHAR(20) NOT NULL,
    provider_key TEXT NOT NULL,
    provider_subject TEXT NOT NULL,
    verified_at TIMESTAMPTZ NULL,
    issuer TEXT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT auth_identities_provider_type_check
        CHECK (provider_type IN ('email', 'linuxdo', 'wechat', 'oidc'))
);

CREATE UNIQUE INDEX IF NOT EXISTS auth_identities_provider_subject_key
    ON auth_identities (provider_type, provider_key, provider_subject);

CREATE INDEX IF NOT EXISTS auth_identities_user_id_idx
    ON auth_identities (user_id);

CREATE INDEX IF NOT EXISTS auth_identities_user_provider_idx
    ON auth_identities (user_id, provider_type);

CREATE TABLE IF NOT EXISTS auth_identity_channels (
    id BIGSERIAL PRIMARY KEY,
    identity_id BIGINT NOT NULL REFERENCES auth_identities(id) ON DELETE CASCADE,
    provider_type VARCHAR(20) NOT NULL,
    provider_key TEXT NOT NULL,
    channel VARCHAR(20) NOT NULL,
    channel_app_id TEXT NOT NULL,
    channel_subject TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT auth_identity_channels_provider_type_check
        CHECK (provider_type IN ('email', 'linuxdo', 'wechat', 'oidc'))
);

CREATE UNIQUE INDEX IF NOT EXISTS auth_identity_channels_channel_key
    ON auth_identity_channels (provider_type, provider_key, channel, channel_app_id, channel_subject);

CREATE INDEX IF NOT EXISTS auth_identity_channels_identity_id_idx
    ON auth_identity_channels (identity_id);

CREATE TABLE IF NOT EXISTS pending_auth_sessions (
    id BIGSERIAL PRIMARY KEY,
    session_token VARCHAR(255) NOT NULL,
    intent VARCHAR(40) NOT NULL,
    provider_type VARCHAR(20) NOT NULL,
    provider_key TEXT NOT NULL,
    provider_subject TEXT NOT NULL,
    target_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    redirect_to TEXT NOT NULL DEFAULT '',
    resolved_email TEXT NOT NULL DEFAULT '',
    registration_password_hash TEXT NOT NULL DEFAULT '',
    upstream_identity_claims JSONB NOT NULL DEFAULT '{}'::jsonb,
    local_flow_state JSONB NOT NULL DEFAULT '{}'::jsonb,
    browser_session_key TEXT NOT NULL DEFAULT '',
    completion_code_hash TEXT NOT NULL DEFAULT '',
    completion_code_expires_at TIMESTAMPTZ NULL,
    email_verified_at TIMESTAMPTZ NULL,
    password_verified_at TIMESTAMPTZ NULL,
    totp_verified_at TIMESTAMPTZ NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pending_auth_sessions_intent_check
        CHECK (intent IN ('login', 'bind_current_user', 'adopt_existing_user_by_email')),
    CONSTRAINT pending_auth_sessions_provider_type_check
        CHECK (provider_type IN ('email', 'linuxdo', 'wechat', 'oidc'))
);

CREATE UNIQUE INDEX IF NOT EXISTS pending_auth_sessions_session_token_key
    ON pending_auth_sessions (session_token);

CREATE INDEX IF NOT EXISTS pending_auth_sessions_target_user_id_idx
    ON pending_auth_sessions (target_user_id);

CREATE INDEX IF NOT EXISTS pending_auth_sessions_expires_at_idx
    ON pending_auth_sessions (expires_at);

CREATE INDEX IF NOT EXISTS pending_auth_sessions_provider_idx
    ON pending_auth_sessions (provider_type, provider_key, provider_subject);

CREATE INDEX IF NOT EXISTS pending_auth_sessions_completion_code_idx
    ON pending_auth_sessions (completion_code_hash);

CREATE TABLE IF NOT EXISTS identity_adoption_decisions (
    id BIGSERIAL PRIMARY KEY,
    pending_auth_session_id BIGINT NOT NULL REFERENCES pending_auth_sessions(id) ON DELETE CASCADE,
    identity_id BIGINT NULL REFERENCES auth_identities(id) ON DELETE SET NULL,
    adopt_display_name BOOLEAN NOT NULL DEFAULT FALSE,
    adopt_avatar BOOLEAN NOT NULL DEFAULT FALSE,
    decided_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS identity_adoption_decisions_pending_auth_session_id_key
    ON identity_adoption_decisions (pending_auth_session_id);

CREATE INDEX IF NOT EXISTS identity_adoption_decisions_identity_id_idx
    ON identity_adoption_decisions (identity_id);

CREATE TABLE IF NOT EXISTS auth_identity_migration_reports (
    id BIGSERIAL PRIMARY KEY,
    report_type VARCHAR(80) NOT NULL,
    report_key TEXT NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    resolved_at TIMESTAMPTZ NULL,
    resolved_by_user_id BIGINT NULL,
    resolution_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS auth_identity_migration_reports_type_idx
    ON auth_identity_migration_reports (report_type);

CREATE INDEX IF NOT EXISTS idx_auth_identity_migration_reports_resolved_at
    ON auth_identity_migration_reports (resolved_at);

CREATE UNIQUE INDEX IF NOT EXISTS auth_identity_migration_reports_type_key
    ON auth_identity_migration_reports (report_type, report_key);

INSERT INTO auth_identities (
    user_id,
    provider_type,
    provider_key,
    provider_subject,
    verified_at,
    metadata
)
SELECT
    u.id,
    'email',
    'email',
    LOWER(BTRIM(u.email)),
    COALESCE(u.updated_at, u.created_at, NOW()),
    jsonb_build_object(
        'backfill_source', 'users.email',
        'migration', '128_auth_identity_foundation'
    )
FROM users AS u
WHERE u.deleted_at IS NULL
  AND BTRIM(COALESCE(u.email, '')) <> ''
  AND RIGHT(LOWER(BTRIM(u.email)), LENGTH('@linuxdo-connect.invalid')) <> '@linuxdo-connect.invalid'
  AND RIGHT(LOWER(BTRIM(u.email)), LENGTH('@oidc-connect.invalid')) <> '@oidc-connect.invalid'
  AND RIGHT(LOWER(BTRIM(u.email)), LENGTH('@wechat-connect.invalid')) <> '@wechat-connect.invalid'
ON CONFLICT (provider_type, provider_key, provider_subject) DO NOTHING;

INSERT INTO auth_identities (
    user_id,
    provider_type,
    provider_key,
    provider_subject,
    verified_at,
    metadata
)
SELECT
    u.id,
    'linuxdo',
    'linuxdo',
    SUBSTRING(BTRIM(u.email) FROM '(?i)^linuxdo-(.+)@linuxdo-connect\.invalid$'),
    COALESCE(u.updated_at, u.created_at, NOW()),
    jsonb_build_object(
        'backfill_source', 'synthetic_email',
        'legacy_email', BTRIM(u.email),
        'migration', '128_auth_identity_foundation'
    )
FROM users AS u
WHERE u.deleted_at IS NULL
  AND LOWER(BTRIM(u.email)) ~ '^linuxdo-.+@linuxdo-connect\.invalid$'
ON CONFLICT (provider_type, provider_key, provider_subject) DO NOTHING;

INSERT INTO auth_identities (
    user_id,
    provider_type,
    provider_key,
    provider_subject,
    verified_at,
    metadata
)
SELECT
    u.id,
    'wechat',
    'wechat',
    SUBSTRING(BTRIM(u.email) FROM '(?i)^wechat-(.+)@wechat-connect\.invalid$'),
    COALESCE(u.updated_at, u.created_at, NOW()),
    jsonb_build_object(
        'backfill_source', 'synthetic_email',
        'legacy_email', BTRIM(u.email),
        'migration', '128_auth_identity_foundation'
    )
FROM users AS u
WHERE u.deleted_at IS NULL
  AND LOWER(BTRIM(u.email)) ~ '^wechat-.+@wechat-connect\.invalid$'
ON CONFLICT (provider_type, provider_key, provider_subject) DO NOTHING;

UPDATE users
SET signup_source = 'linuxdo'
WHERE deleted_at IS NULL
  AND LOWER(BTRIM(COALESCE(email, ''))) ~ '^linuxdo-.+@linuxdo-connect\.invalid$';

UPDATE users
SET signup_source = 'wechat'
WHERE deleted_at IS NULL
  AND LOWER(BTRIM(COALESCE(email, ''))) ~ '^wechat-.+@wechat-connect\.invalid$';

UPDATE users
SET signup_source = 'oidc'
WHERE deleted_at IS NULL
  AND LOWER(BTRIM(COALESCE(email, ''))) ~ '^oidc-.+@oidc-connect\.invalid$';

INSERT INTO auth_identity_migration_reports (report_type, report_key, details)
SELECT
    'oidc_synthetic_email_requires_manual_recovery',
    CAST(u.id AS TEXT),
    jsonb_build_object(
        'user_id', u.id,
        'email', LOWER(BTRIM(u.email)),
        'reason', 'cannot recover issuer_plus_sub deterministically from synthetic email alone',
        'migration', '128_auth_identity_foundation'
    )
FROM users AS u
WHERE u.deleted_at IS NULL
  AND LOWER(BTRIM(u.email)) ~ '^oidc-.+@oidc-connect\.invalid$'
ON CONFLICT (report_type, report_key) DO NOTHING;

INSERT INTO auth_identity_migration_reports (report_type, report_key, details)
SELECT
    'wechat_openid_only_requires_remediation',
    CAST(u.id AS TEXT),
    jsonb_build_object(
        'user_id', u.id,
        'email', LOWER(BTRIM(u.email)),
        'reason', 'legacy wechat synthetic identity requires explicit unionid remediation if channel-only data exists',
        'migration', '128_auth_identity_foundation'
    )
FROM users AS u
WHERE u.deleted_at IS NULL
  AND LOWER(BTRIM(u.email)) ~ '^wechat-.+@wechat-connect\.invalid$'
  AND NOT EXISTS (
      SELECT 1
      FROM auth_identities ai
      WHERE ai.user_id = u.id
        AND ai.provider_type = 'wechat'
        AND ai.provider_key = 'wechat'
  )
ON CONFLICT (report_type, report_key) DO NOTHING;
