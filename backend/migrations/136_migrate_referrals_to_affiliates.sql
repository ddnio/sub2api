-- Backfill legacy fork referral relationships into upstream affiliate tables.
--
-- This migration intentionally copies attribution only. It must not mutate
-- affiliate ledger/quota, user balances, redeem codes, or legacy referral rows.

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

-- Preflight guard: duplicate legacy referral codes would make aff_code backfill ambiguous.
-- Only users participating in legacy referral relationships need backfilled rows.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM (
            SELECT u.id, upper(btrim(u.referral_code)) AS code
            FROM (
                SELECT inviter_id AS user_id FROM user_referrals
                UNION
                SELECT invitee_id AS user_id FROM user_referrals
            ) participants
            JOIN users u ON u.id = participants.user_id
            LEFT JOIN user_affiliates ua ON ua.user_id = u.id
            WHERE u.deleted_at IS NULL
              AND ua.user_id IS NULL
              AND u.referral_code IS NOT NULL
              AND btrim(u.referral_code) <> ''
        ) legacy_codes
        GROUP BY code
        HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION 'legacy referral migration blocked: duplicate referral_code values found';
    END IF;
END $$;

-- Preflight guard: legacy codes must fit upstream affiliate code constraints.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM (
            SELECT inviter_id AS user_id FROM user_referrals
            UNION
            SELECT invitee_id AS user_id FROM user_referrals
        ) participants
        JOIN users u ON u.id = participants.user_id
        LEFT JOIN user_affiliates ua ON ua.user_id = u.id
        WHERE u.deleted_at IS NULL
          AND ua.user_id IS NULL
          AND u.referral_code IS NOT NULL
          AND btrim(u.referral_code) <> ''
          AND (
            length(btrim(u.referral_code)) > 32
            OR btrim(u.referral_code) !~ '^[A-Za-z0-9_-]+$'
          )
    ) THEN
        RAISE EXCEPTION 'legacy referral migration blocked: invalid referral_code values found';
    END IF;
END $$;

-- Preflight guard: do not let another user's affiliate code collide with a legacy code.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM (
            SELECT inviter_id AS user_id FROM user_referrals
            UNION
            SELECT invitee_id AS user_id FROM user_referrals
        ) participants
        JOIN users u ON u.id = participants.user_id
        LEFT JOIN user_affiliates own_affiliate ON own_affiliate.user_id = u.id
        JOIN user_affiliates other_affiliate
          ON other_affiliate.aff_code = upper(btrim(u.referral_code))
         AND other_affiliate.user_id <> u.id
        WHERE u.deleted_at IS NULL
          AND own_affiliate.user_id IS NULL
          AND u.referral_code IS NOT NULL
          AND btrim(u.referral_code) <> ''
    ) THEN
        RAISE EXCEPTION 'legacy referral migration blocked: referral_code conflicts with existing aff_code';
    END IF;
END $$;

-- Preflight guard: do not overwrite a different upstream affiliate inviter.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM user_referrals r
        JOIN user_affiliates ua ON ua.user_id = r.invitee_id
        WHERE ua.inviter_id IS NOT NULL
          AND ua.inviter_id <> r.inviter_id
    ) THEN
        RAISE EXCEPTION 'legacy referral migration blocked: existing affiliate inviter conflicts with legacy referral';
    END IF;
END $$;

-- Preflight guard: this relationship-only migration uses legacy referral_code
-- for newly inserted affiliate rows. Only legacy referral participants must be
-- present here; unrelated users can still get affiliate rows lazily upstream.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM (
            SELECT inviter_id AS user_id FROM user_referrals
            UNION
            SELECT invitee_id AS user_id FROM user_referrals
        ) participants
        JOIN users u ON u.id = participants.user_id
        LEFT JOIN user_affiliates ua ON ua.user_id = u.id
        WHERE u.deleted_at IS NULL
          AND ua.user_id IS NULL
          AND (
            u.referral_code IS NULL
            OR btrim(u.referral_code) = ''
          )
    ) THEN
        RAISE EXCEPTION 'legacy referral migration blocked: referral participants without affiliate rows or referral_code found';
    END IF;
END $$;

-- Ensure affiliate rows exist for active users participating in legacy referrals.
INSERT INTO user_affiliates (user_id, aff_code, created_at, updated_at)
SELECT u.id,
       upper(btrim(u.referral_code)),
       NOW(),
       NOW()
FROM (
    SELECT inviter_id AS user_id FROM user_referrals
    UNION
    SELECT invitee_id AS user_id FROM user_referrals
) participants
JOIN users u ON u.id = participants.user_id
WHERE u.deleted_at IS NULL
  AND u.referral_code IS NOT NULL
  AND btrim(u.referral_code) <> ''
ON CONFLICT (user_id) DO NOTHING;

-- Copy legacy inviter relationships where no upstream inviter is already set.
UPDATE user_affiliates ua
SET inviter_id = r.inviter_id,
    updated_at = NOW()
FROM user_referrals r
WHERE ua.user_id = r.invitee_id
  AND ua.inviter_id IS NULL;

-- Recompute invite counts from the final inviter graph.
WITH counts AS (
    SELECT inviter_id AS user_id, count(*)::integer AS cnt
    FROM user_affiliates
    WHERE inviter_id IS NOT NULL
    GROUP BY inviter_id
)
UPDATE user_affiliates ua
SET aff_count = COALESCE(counts.cnt, 0),
    updated_at = NOW()
FROM users u
LEFT JOIN counts ON counts.user_id = u.id
WHERE ua.user_id = u.id
  AND ua.aff_count <> COALESCE(counts.cnt, 0);
