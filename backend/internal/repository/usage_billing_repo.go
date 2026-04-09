package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type usageBillingRepository struct {
	db *sql.DB
}

func NewUsageBillingRepository(_ *dbent.Client, sqlDB *sql.DB) service.UsageBillingRepository {
	return &usageBillingRepository{db: sqlDB}
}

func (r *usageBillingRepository) Apply(ctx context.Context, cmd *service.UsageBillingCommand) (_ *service.UsageBillingApplyResult, err error) {
	if cmd == nil {
		return &service.UsageBillingApplyResult{}, nil
	}
	if r == nil || r.db == nil {
		return nil, errors.New("usage billing repository db is nil")
	}

	cmd.Normalize()
	if cmd.RequestID == "" {
		return nil, service.ErrUsageBillingRequestIDRequired
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	applied, err := r.claimUsageBillingKey(ctx, tx, cmd)
	if err != nil {
		return nil, err
	}
	if !applied {
		return &service.UsageBillingApplyResult{Applied: false}, nil
	}

	result := &service.UsageBillingApplyResult{Applied: true}
	if err := r.applyUsageBillingEffects(ctx, tx, cmd, result); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil
	return result, nil
}

func (r *usageBillingRepository) claimUsageBillingKey(ctx context.Context, tx *sql.Tx, cmd *service.UsageBillingCommand) (bool, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO usage_billing_dedup (request_id, api_key_id, request_fingerprint)
		VALUES ($1, $2, $3)
		ON CONFLICT (request_id, api_key_id) DO NOTHING
		RETURNING id
	`, cmd.RequestID, cmd.APIKeyID, cmd.RequestFingerprint).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		var existingFingerprint string
		if err := tx.QueryRowContext(ctx, `
			SELECT request_fingerprint
			FROM usage_billing_dedup
			WHERE request_id = $1 AND api_key_id = $2
		`, cmd.RequestID, cmd.APIKeyID).Scan(&existingFingerprint); err != nil {
			return false, err
		}
		if strings.TrimSpace(existingFingerprint) != strings.TrimSpace(cmd.RequestFingerprint) {
			return false, service.ErrUsageBillingRequestConflict
		}
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var archivedFingerprint string
	err = tx.QueryRowContext(ctx, `
		SELECT request_fingerprint
		FROM usage_billing_dedup_archive
		WHERE request_id = $1 AND api_key_id = $2
	`, cmd.RequestID, cmd.APIKeyID).Scan(&archivedFingerprint)
	if err == nil {
		if strings.TrimSpace(archivedFingerprint) != strings.TrimSpace(cmd.RequestFingerprint) {
			return false, service.ErrUsageBillingRequestConflict
		}
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return true, nil
}

func (r *usageBillingRepository) applyUsageBillingEffects(ctx context.Context, tx *sql.Tx, cmd *service.UsageBillingCommand, result *service.UsageBillingApplyResult) error {
	if cmd.SubscriptionCost > 0 && cmd.SubscriptionID != nil {
		if err := incrementUsageBillingSubscription(ctx, tx, *cmd.SubscriptionID, cmd.SubscriptionCost); err != nil {
			return err
		}
	}

	if cmd.BalanceCost > 0 {
		if err := deductUsageBillingBalance(ctx, tx, cmd.UserID, cmd.BalanceCost); err != nil {
			return err
		}
		// 首次余额扣费触发邀请奖励（原子：同一事务内；订阅扣费不触发）
		if err := grantReferralRewardOnFirstConsumption(ctx, tx, cmd.UserID, result); err != nil {
			return err
		}
	}

	if cmd.APIKeyQuotaCost > 0 {
		exhausted, err := incrementUsageBillingAPIKeyQuota(ctx, tx, cmd.APIKeyID, cmd.APIKeyQuotaCost)
		if err != nil {
			return err
		}
		result.APIKeyQuotaExhausted = exhausted
	}

	if cmd.APIKeyRateLimitCost > 0 {
		if err := incrementUsageBillingAPIKeyRateLimit(ctx, tx, cmd.APIKeyID, cmd.APIKeyRateLimitCost); err != nil {
			return err
		}
	}

	if cmd.AccountQuotaCost > 0 && (strings.EqualFold(cmd.AccountType, service.AccountTypeAPIKey) || strings.EqualFold(cmd.AccountType, service.AccountTypeBedrock)) {
		if err := incrementUsageBillingAccountQuota(ctx, tx, cmd.AccountID, cmd.AccountQuotaCost); err != nil {
			return err
		}
	}

	return nil
}

func incrementUsageBillingSubscription(ctx context.Context, tx *sql.Tx, subscriptionID int64, costUSD float64) error {
	const updateSQL = `
		UPDATE user_subscriptions us
		SET
			daily_usage_usd = us.daily_usage_usd + $1,
			weekly_usage_usd = us.weekly_usage_usd + $1,
			monthly_usage_usd = us.monthly_usage_usd + $1,
			updated_at = NOW()
		FROM groups g
		WHERE us.id = $2
			AND us.deleted_at IS NULL
			AND us.group_id = g.id
			AND g.deleted_at IS NULL
	`
	res, err := tx.ExecContext(ctx, updateSQL, costUSD, subscriptionID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	return service.ErrSubscriptionNotFound
}

func deductUsageBillingBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE users
		SET balance = balance - $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`, amount, userID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	return service.ErrUserNotFound
}

func incrementUsageBillingAPIKeyQuota(ctx context.Context, tx *sql.Tx, apiKeyID int64, amount float64) (bool, error) {
	var exhausted bool
	err := tx.QueryRowContext(ctx, `
		UPDATE api_keys
		SET quota_used = quota_used + $1,
			status = CASE
				WHEN quota > 0
					AND status = $3
					AND quota_used < quota
					AND quota_used + $1 >= quota
				THEN $4
				ELSE status
			END,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING quota > 0 AND quota_used >= quota AND quota_used - $1 < quota
	`, amount, apiKeyID, service.StatusAPIKeyActive, service.StatusAPIKeyQuotaExhausted).Scan(&exhausted)
	if errors.Is(err, sql.ErrNoRows) {
		return false, service.ErrAPIKeyNotFound
	}
	if err != nil {
		return false, err
	}
	return exhausted, nil
}

func incrementUsageBillingAPIKeyRateLimit(ctx context.Context, tx *sql.Tx, apiKeyID int64, cost float64) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE api_keys SET
			usage_5h = CASE WHEN window_5h_start IS NOT NULL AND window_5h_start + INTERVAL '5 hours' <= NOW() THEN $1 ELSE usage_5h + $1 END,
			usage_1d = CASE WHEN window_1d_start IS NOT NULL AND window_1d_start + INTERVAL '24 hours' <= NOW() THEN $1 ELSE usage_1d + $1 END,
			usage_7d = CASE WHEN window_7d_start IS NOT NULL AND window_7d_start + INTERVAL '7 days' <= NOW() THEN $1 ELSE usage_7d + $1 END,
			window_5h_start = CASE WHEN window_5h_start IS NULL OR window_5h_start + INTERVAL '5 hours' <= NOW() THEN NOW() ELSE window_5h_start END,
			window_1d_start = CASE WHEN window_1d_start IS NULL OR window_1d_start + INTERVAL '24 hours' <= NOW() THEN date_trunc('day', NOW()) ELSE window_1d_start END,
			window_7d_start = CASE WHEN window_7d_start IS NULL OR window_7d_start + INTERVAL '7 days' <= NOW() THEN date_trunc('day', NOW()) ELSE window_7d_start END,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`, cost, apiKeyID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrAPIKeyNotFound
	}
	return nil
}

func incrementUsageBillingAccountQuota(ctx context.Context, tx *sql.Tx, accountID int64, amount float64) error {
	rows, err := tx.QueryContext(ctx,
		`UPDATE accounts SET extra = (
			COALESCE(extra, '{}'::jsonb)
			|| jsonb_build_object('quota_used', COALESCE((extra->>'quota_used')::numeric, 0) + $1)
			|| CASE WHEN COALESCE((extra->>'quota_daily_limit')::numeric, 0) > 0 THEN
				jsonb_build_object(
					'quota_daily_used',
					CASE WHEN COALESCE((extra->>'quota_daily_start')::timestamptz, '1970-01-01'::timestamptz)
						+ '24 hours'::interval <= NOW()
					THEN $1
					ELSE COALESCE((extra->>'quota_daily_used')::numeric, 0) + $1 END,
					'quota_daily_start',
					CASE WHEN COALESCE((extra->>'quota_daily_start')::timestamptz, '1970-01-01'::timestamptz)
						+ '24 hours'::interval <= NOW()
					THEN `+nowUTC+`
					ELSE COALESCE(extra->>'quota_daily_start', `+nowUTC+`) END
				)
			ELSE '{}'::jsonb END
			|| CASE WHEN COALESCE((extra->>'quota_weekly_limit')::numeric, 0) > 0 THEN
				jsonb_build_object(
					'quota_weekly_used',
					CASE WHEN COALESCE((extra->>'quota_weekly_start')::timestamptz, '1970-01-01'::timestamptz)
						+ '168 hours'::interval <= NOW()
					THEN $1
					ELSE COALESCE((extra->>'quota_weekly_used')::numeric, 0) + $1 END,
					'quota_weekly_start',
					CASE WHEN COALESCE((extra->>'quota_weekly_start')::timestamptz, '1970-01-01'::timestamptz)
						+ '168 hours'::interval <= NOW()
					THEN `+nowUTC+`
					ELSE COALESCE(extra->>'quota_weekly_start', `+nowUTC+`) END
				)
			ELSE '{}'::jsonb END
		), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING
			COALESCE((extra->>'quota_used')::numeric, 0),
			COALESCE((extra->>'quota_limit')::numeric, 0)`,
		amount, accountID)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	var newUsed, limit float64
	if rows.Next() {
		if err := rows.Scan(&newUsed, &limit); err != nil {
			return err
		}
	} else {
		if err := rows.Err(); err != nil {
			return err
		}
		return service.ErrAccountNotFound
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if limit > 0 && newUsed >= limit && (newUsed-amount) < limit {
		if err := enqueueSchedulerOutbox(ctx, tx, service.SchedulerOutboxEventAccountChanged, &accountID, nil, nil); err != nil {
			logger.LegacyPrintf("repository.usage_billing", "[SchedulerOutbox] enqueue quota exceeded failed: account=%d err=%v", accountID, err)
			return err
		}
	}
	return nil
}

// grantReferralRewardOnFirstConsumption 首次余额消费时原子发放邀请奖励。
// 幂等：UPDATE ... WHERE reward_granted_at IS NULL 保证只触发一次。
// 无邀请关系的用户：UPDATE 0 rows，无额外开销。
func grantReferralRewardOnFirstConsumption(ctx context.Context, tx *sql.Tx, userID int64, result *service.UsageBillingApplyResult) error {
	// 原子标记 + 读取奖励信息
	var inviterID int64
	var inviterReward, inviteeReward float64
	err := tx.QueryRowContext(ctx, `
		UPDATE user_referrals
		SET reward_granted_at = NOW(),
			inviter_rewarded = inviter_reward_snapshot,
			invitee_rewarded = invitee_reward_snapshot
		WHERE invitee_id = $1 AND reward_granted_at IS NULL
		RETURNING inviter_id, inviter_reward_snapshot, invitee_reward_snapshot
	`, userID).Scan(&inviterID, &inviterReward, &inviteeReward)

	if errors.Is(err, sql.ErrNoRows) {
		return nil // 非被邀请用户或奖励已发放
	}
	if err != nil {
		return fmt.Errorf("grant referral reward: claim: %w", err)
	}

	// 给邀请人加余额（检查 RowsAffected 防止已删除用户导致账实不符）
	if inviterReward > 0 {
		res, err := tx.ExecContext(ctx, `
			UPDATE users SET balance = balance + $1, updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL
		`, inviterReward, inviterID)
		if err != nil {
			return fmt.Errorf("grant referral reward: inviter balance: %w", err)
		}
		if affected, _ := res.RowsAffected(); affected > 0 {
			if err := insertReferralRedeemCode(ctx, tx, "ref_inviter", inviterReward, inviterID,
				fmt.Sprintf("邀请用户注册奖励 (用户ID: %d)", userID)); err != nil {
				return fmt.Errorf("grant referral reward: inviter redeem: %w", err)
			}
		}
	}

	// 给被邀请人加余额
	if inviteeReward > 0 {
		res, err := tx.ExecContext(ctx, `
			UPDATE users SET balance = balance + $1, updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL
		`, inviteeReward, userID)
		if err != nil {
			return fmt.Errorf("grant referral reward: invitee balance: %w", err)
		}
		if affected, _ := res.RowsAffected(); affected > 0 {
			if err := insertReferralRedeemCode(ctx, tx, "ref_invitee", inviteeReward, userID,
				fmt.Sprintf("通过邀请注册奖励 (邀请人ID: %d)", inviterID)); err != nil {
				return fmt.Errorf("grant referral reward: invitee redeem: %w", err)
			}
		}
	}

	// 标记结果，供 service 层做缓存失效
	result.ReferralRewardGranted = true
	result.ReferralInviterID = inviterID
	result.ReferralInviteeID = userID
	result.ReferralInviterAmount = inviterReward
	result.ReferralInviteeAmount = inviteeReward

	return nil
}

// insertReferralRedeemCode 写入邀请奖励的 redeem_code 审计记录（128-bit 随机码）。
func insertReferralRedeemCode(ctx context.Context, tx *sql.Tx, redeemType string, value float64, usedBy int64, notes string) error {
	b := make([]byte, 16) // 16 bytes = 32 hex chars = MaxLen(32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("generate redeem code: %w", err)
	}
	code := hex.EncodeToString(b)

	_, err := tx.ExecContext(ctx, `
		INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, notes, created_at, validity_days)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6, NOW(), 0)
	`, code, redeemType, value, "used", usedBy, notes)
	return err
}
