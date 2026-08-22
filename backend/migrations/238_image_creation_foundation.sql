CREATE TABLE IF NOT EXISTS image_creation_assets (
    id VARCHAR(64) PRIMARY KEY,
    content BYTEA NOT NULL,
    content_type VARCHAR(32) NOT NULL CHECK (content_type IN ('image/png', 'image/jpeg', 'image/webp')),
    byte_size INTEGER NOT NULL CHECK (byte_size BETWEEN 1 AND 8388608),
    width INTEGER NOT NULL CHECK (width BETWEEN 1 AND 8192),
    height INTEGER NOT NULL CHECK (height BETWEEN 1 AND 8192),
    source_type VARCHAR(16) NOT NULL CHECK (source_type IN ('generated', 'uploaded', 'imported')),
    source_provider VARCHAR(64),
    source_model VARCHAR(120),
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT image_creation_assets_id_sha256 CHECK (id ~ '^[0-9a-f]{64}$'),
    CONSTRAINT image_creation_assets_content_size CHECK (octet_length(content) = byte_size)
);

CREATE INDEX IF NOT EXISTS image_creation_assets_creator_created_idx
    ON image_creation_assets(created_by, created_at DESC);

CREATE TABLE IF NOT EXISTS image_creation_templates (
    id BIGSERIAL PRIMARY KEY,
    state VARCHAR(16) NOT NULL DEFAULT 'draft' CHECK (state IN ('draft', 'published', 'archived')),
    draft_data JSONB NOT NULL CHECK (jsonb_typeof(draft_data) = 'object'),
    published_data JSONB CHECK (published_data IS NULL OR jsonb_typeof(published_data) = 'object'),
    revision INTEGER NOT NULL DEFAULT 1 CHECK (revision >= 1),
    published_version INTEGER NOT NULL DEFAULT 0 CHECK (published_version >= 0),
    draft_cover_asset_id VARCHAR(64) REFERENCES image_creation_assets(id) ON DELETE RESTRICT,
    published_cover_asset_id VARCHAR(64) REFERENCES image_creation_assets(id) ON DELETE RESTRICT,
    home_position SMALLINT CHECK (home_position BETWEEN 1 AND 6),
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    CONSTRAINT image_creation_templates_published_data_required
        CHECK (state <> 'published' OR published_data IS NOT NULL),
    CONSTRAINT image_creation_templates_home_published
        CHECK (home_position IS NULL OR state = 'published')
);

CREATE INDEX IF NOT EXISTS image_creation_templates_state_published_idx
    ON image_creation_templates(state, published_at DESC);
CREATE INDEX IF NOT EXISTS image_creation_templates_updated_idx
    ON image_creation_templates(updated_at DESC);
CREATE INDEX IF NOT EXISTS image_creation_templates_published_data_gin
    ON image_creation_templates USING GIN (published_data jsonb_path_ops)
    WHERE published_data IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS image_creation_templates_home_position_unique
    ON image_creation_templates(home_position)
    WHERE home_position IS NOT NULL;

CREATE TABLE IF NOT EXISTS image_creation_user_template_states (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id BIGINT NOT NULL REFERENCES image_creation_templates(id) ON DELETE CASCADE,
    favorited_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, template_id),
    CONSTRAINT image_creation_user_template_states_nonempty
        CHECK (favorited_at IS NOT NULL OR last_used_at IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS image_creation_user_template_states_favorite_idx
    ON image_creation_user_template_states(user_id, favorited_at DESC)
    WHERE favorited_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS image_creation_user_template_states_recent_idx
    ON image_creation_user_template_states(user_id, last_used_at DESC)
    WHERE last_used_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS image_creation_change_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action VARCHAR(32) NOT NULL CHECK (action IN ('create', 'update', 'publish', 'archive', 'restore', 'home_update', 'asset_create')),
    target_type VARCHAR(16) NOT NULL CHECK (target_type IN ('template', 'home', 'asset')),
    target_id VARCHAR(64) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(metadata) = 'object'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS image_creation_change_logs_actor_created_idx
    ON image_creation_change_logs(actor_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS image_creation_change_logs_target_created_idx
    ON image_creation_change_logs(target_type, target_id, created_at DESC);
