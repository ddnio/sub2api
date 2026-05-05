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
