-- Add referral system: referral_code on users + user_referrals table.
-- Safe for online execution: ADD COLUMN NULL is metadata-only (no table rewrite).

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

-- 1. Add referral_code to users (nullable, partial unique for soft delete)
ALTER TABLE users ADD COLUMN IF NOT EXISTS referral_code VARCHAR(16);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_referral_code_active
  ON users (referral_code) WHERE deleted_at IS NULL AND referral_code IS NOT NULL;

-- 2. Create user_referrals table
CREATE TABLE IF NOT EXISTS user_referrals (
    id                BIGSERIAL PRIMARY KEY,
    inviter_id        BIGINT NOT NULL REFERENCES users(id),
    invitee_id        BIGINT NOT NULL REFERENCES users(id),
    code              VARCHAR(16) NOT NULL,
    inviter_rewarded  DECIMAL(20,8) DEFAULT 0,
    invitee_rewarded  DECIMAL(20,8) DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_referrals_invitee UNIQUE(invitee_id),
    CONSTRAINT no_self_referral CHECK (inviter_id != invitee_id)
);

CREATE INDEX IF NOT EXISTS idx_user_referrals_inviter_id ON user_referrals(inviter_id);
CREATE INDEX IF NOT EXISTS idx_user_referrals_code ON user_referrals(code);

-- 3. Backfill existing users with referral codes (id-seeded to guarantee uniqueness)
UPDATE users SET referral_code = UPPER(SUBSTR(MD5(id::text || now()::text || random()::text), 1, 8))
WHERE referral_code IS NULL AND deleted_at IS NULL;
