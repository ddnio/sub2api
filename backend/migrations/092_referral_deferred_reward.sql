-- Referral deferred reward: add reward snapshot + grant tracking fields.
-- Rewards are no longer granted at registration; they are granted on first balance charge (BalanceCost > 0).
-- Note: subscription-only consumption does not trigger referral reward.

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

-- 1. Add reward snapshot fields (amount captured at registration time)
ALTER TABLE user_referrals ADD COLUMN IF NOT EXISTS inviter_reward_snapshot DECIMAL(20,8) NOT NULL DEFAULT 0;
ALTER TABLE user_referrals ADD COLUMN IF NOT EXISTS invitee_reward_snapshot DECIMAL(20,8) NOT NULL DEFAULT 0;

-- 2. Add reward grant tracking (NULL = pending, non-NULL = granted)
ALTER TABLE user_referrals ADD COLUMN IF NOT EXISTS reward_granted_at TIMESTAMPTZ NULL;

-- 3. Backfill existing records: they were granted at registration time
UPDATE user_referrals
SET reward_granted_at = created_at,
    inviter_reward_snapshot = inviter_rewarded,
    invitee_reward_snapshot = invitee_rewarded
WHERE reward_granted_at IS NULL;
