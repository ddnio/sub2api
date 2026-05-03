package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/ent/userreferral"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// ReferralService 推荐码服务
type ReferralService struct {
	entClient            *dbent.Client
	userRepo             UserRepository
	redeemRepo           RedeemCodeRepository
	settingService       *SettingService
	billingCacheService  *BillingCacheService
	authCacheInvalidator APIKeyAuthCacheInvalidator
}

// NewReferralService 创建推荐码服务实例
func NewReferralService(
	entClient *dbent.Client,
	userRepo UserRepository,
	redeemRepo RedeemCodeRepository,
	settingService *SettingService,
	billingCacheService *BillingCacheService,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
) *ReferralService {
	return &ReferralService{
		entClient:            entClient,
		userRepo:             userRepo,
		redeemRepo:           redeemRepo,
		settingService:       settingService,
		billingCacheService:  billingCacheService,
		authCacheInvalidator: authCacheInvalidator,
	}
}

// ReferralInfo 推荐码信息
type ReferralInfo struct {
	ReferralCode     string  `json:"referral_code"`
	TotalInvited     int     `json:"total_invited"`
	TotalRewarded    float64 `json:"total_rewarded"`
	PendingCount     int     `json:"pending_count"`         // 待激活人数
	InviterRewardAmt float64 `json:"inviter_reward_amount"` // 当前配置的邀请人奖励
	InviteeRewardAmt float64 `json:"invitee_reward_amount"` // 当前配置的被邀请人奖励
}

// ReferralRecord 邀请记录
type ReferralRecord struct {
	ID                    int64      `json:"id"`
	InviterID             int64      `json:"inviter_id"`
	InviteeID             int64      `json:"invitee_id"`
	InviterEmail          string     `json:"inviter_email"`
	InviteeEmail          string     `json:"invitee_email"`
	Code                  string     `json:"code"`
	InviterRewarded       float64    `json:"inviter_rewarded"`
	InviteeRewarded       float64    `json:"invitee_rewarded"`
	InviterRewardSnapshot float64    `json:"inviter_reward_snapshot"`
	InviteeRewardSnapshot float64    `json:"invitee_reward_snapshot"`
	RewardGrantedAt       *time.Time `json:"reward_granted_at"`
	CreatedAt             time.Time  `json:"created_at"`
}

// GenerateReferralCode 为用户生成唯一推荐码（带重试）
func (s *ReferralService) GenerateReferralCode(ctx context.Context, userID int64) (string, error) {
	const maxRetries = 3

	for i := 0; i < maxRetries; i++ {
		code, err := generateRandomCode(4) // 4 bytes = 8 hex chars
		if err != nil {
			return "", fmt.Errorf("generate random code: %w", err)
		}
		code = strings.ToUpper(code) // 统一大写

		// 尝试更新用户的 referral_code
		err = s.entClient.User.UpdateOneID(userID).
			SetReferralCode(code).
			Exec(ctx)
		if err != nil {
			if dbent.IsConstraintError(err) && i < maxRetries-1 {
				continue // 唯一冲突，重试
			}
			return "", fmt.Errorf("set referral code: %w", err)
		}
		return code, nil
	}
	return "", fmt.Errorf("failed to generate unique referral code after %d retries", maxRetries)
}

// ProcessRegistrationReferral 处理注册时的推荐码归因（仅记录，不发奖励）。
// 奖励在被邀请人首次消费时由 usage_billing_repo 触发发放。
// 此方法为非关键操作：失败不影响注册，只记录日志并返回 error。
func (s *ReferralService) ProcessRegistrationReferral(ctx context.Context, inviteeID int64, referralCode string) error {
	referralCode = strings.TrimSpace(strings.ToUpper(referralCode))
	if referralCode == "" {
		return nil
	}

	// 检查功能是否启用
	if s.settingService == nil || !s.settingService.IsReferralEnabled(ctx) {
		return nil
	}

	// 查找邀请人
	inviter, err := s.entClient.User.Query().
		Where(
			dbuser.ReferralCodeEQ(referralCode),
			dbuser.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		logger.LegacyPrintf("service.referral", "[Referral] Referral code not found or invalid: %s, error: %v", referralCode, err)
		return nil // 无效码静默忽略
	}

	// 检查自邀
	if inviter.ID == inviteeID {
		logger.LegacyPrintf("service.referral", "[Referral] Self-referral ignored: user %d", inviteeID)
		return nil
	}

	// 快照当前奖励金额配置
	inviterAmount := s.settingService.GetReferralInviterAmount(ctx)
	inviteeAmount := s.settingService.GetReferralInviteeAmount(ctx)

	// 只创建归因记录，不发放奖励（reward_granted_at 为 NULL 表示待激活）
	_, err = s.entClient.UserReferral.Create().
		SetInviterID(inviter.ID).
		SetInviteeID(inviteeID).
		SetCode(referralCode).
		SetInviterRewardSnapshot(inviterAmount).
		SetInviteeRewardSnapshot(inviteeAmount).
		SetInviterRewarded(0).
		SetInviteeRewarded(0).
		Save(ctx)
	if err != nil {
		// 可能是 UNIQUE(invitee_id) 冲突——已被邀请过
		logger.LegacyPrintf("service.referral", "[Referral] Failed to create referral record: inviter=%d invitee=%d error=%v", inviter.ID, inviteeID, err)
		return fmt.Errorf("create referral record: %w", err)
	}

	logger.LegacyPrintf("service.referral",
		"[Referral] Attribution recorded: inviter=%d invitee=%d code=%s snapshot_inviter=%.2f snapshot_invitee=%.2f (pending first consumption)",
		inviter.ID, inviteeID, referralCode, inviterAmount, inviteeAmount)

	return nil
}

// invalidateReferralCaches 失效相关缓存
func (s *ReferralService) invalidateReferralCaches(ctx context.Context, inviterID, inviteeID int64, inviterAmount, inviteeAmount float64) {
	cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Auth cache
	if s.authCacheInvalidator != nil {
		if inviterAmount > 0 {
			s.authCacheInvalidator.InvalidateAuthCacheByUserID(cacheCtx, inviterID)
		}
		if inviteeAmount > 0 {
			s.authCacheInvalidator.InvalidateAuthCacheByUserID(cacheCtx, inviteeID)
		}
	}

	// Billing cache
	if s.billingCacheService != nil {
		if inviterAmount > 0 {
			go func() {
				c, cn := context.WithTimeout(context.Background(), 5*time.Second)
				defer cn()
				_ = s.billingCacheService.InvalidateUserBalance(c, inviterID)
			}()
		}
		if inviteeAmount > 0 {
			go func() {
				c, cn := context.WithTimeout(context.Background(), 5*time.Second)
				defer cn()
				_ = s.billingCacheService.InvalidateUserBalance(c, inviteeID)
			}()
		}
	}
}

// GetReferralInfo 获取用户的推荐码信息和统计（纯查询，无副作用）
func (s *ReferralService) GetReferralInfo(ctx context.Context, userID int64) (*ReferralInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	referralCode := ""
	if user.ReferralCode != nil {
		referralCode = *user.ReferralCode
	}

	// 聚合查询：拿到总邀请数、已发放奖励总额、待激活人数
	referrals, err := s.entClient.UserReferral.Query().
		Where(userreferral.InviterID(userID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query referrals: %w", err)
	}

	var totalInvited, pendingCount int
	var totalRewarded float64
	totalInvited = len(referrals)
	for _, r := range referrals {
		if r.RewardGrantedAt != nil {
			totalRewarded += r.InviterRewarded
		} else {
			pendingCount++
		}
	}

	inviterAmt := float64(0)
	inviteeAmt := float64(0)
	if s.settingService != nil {
		inviterAmt = s.settingService.GetReferralInviterAmount(ctx)
		inviteeAmt = s.settingService.GetReferralInviteeAmount(ctx)
	}

	return &ReferralInfo{
		ReferralCode:     referralCode,
		TotalInvited:     totalInvited,
		TotalRewarded:    totalRewarded,
		PendingCount:     pendingCount,
		InviterRewardAmt: inviterAmt,
		InviteeRewardAmt: inviteeAmt,
	}, nil
}

// EnsureReferralCode 确保用户有推荐码（惰性生成，仅用户主动访问邀请页时调用）
func (s *ReferralService) EnsureReferralCode(ctx context.Context, userID int64) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}
	if user.ReferralCode != nil && *user.ReferralCode != "" {
		return *user.ReferralCode, nil
	}
	return s.GenerateReferralCode(ctx, userID)
}

// ListReferrals 获取用户的邀请列表（分页）
func (s *ReferralService) ListReferrals(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ReferralRecord, *pagination.PaginationResult, error) {
	// 查总数
	total, err := s.entClient.UserReferral.Query().
		Where(userreferral.InviterID(userID)).
		Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("count referrals: %w", err)
	}

	pageSize := params.Limit()
	pages := total / pageSize
	if total%pageSize > 0 {
		pages++
	}
	paginationResult := &pagination.PaginationResult{
		Total:    int64(total),
		Page:     params.Page,
		PageSize: pageSize,
		Pages:    pages,
	}

	// 查列表
	referrals, err := s.entClient.UserReferral.Query().
		Where(userreferral.InviterID(userID)).
		WithInvitee().
		Order(dbent.Desc(userreferral.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list referrals: %w", err)
	}

	records := make([]ReferralRecord, 0, len(referrals))
	for _, r := range referrals {
		email := ""
		if r.Edges.Invitee != nil {
			email = maskEmail(r.Edges.Invitee.Email)
		}
		records = append(records, ReferralRecord{
			ID:                    r.ID,
			InviterID:             r.InviterID,
			InviteeID:             r.InviteeID,
			InviteeEmail:          email,
			Code:                  r.Code,
			InviterRewarded:       r.InviterRewarded,
			InviteeRewarded:       r.InviteeRewarded,
			InviterRewardSnapshot: r.InviterRewardSnapshot,
			InviteeRewardSnapshot: r.InviteeRewardSnapshot,
			RewardGrantedAt:       r.RewardGrantedAt,
			CreatedAt:             r.CreatedAt,
		})
	}

	return records, paginationResult, nil
}

// GetReferralByInvitee 查询某用户的被邀请关系（admin 用）
func (s *ReferralService) GetReferralByInvitee(ctx context.Context, inviteeID int64) (*ReferralRecord, error) {
	r, err := s.entClient.UserReferral.Query().
		Where(userreferral.InviteeID(inviteeID)).
		WithInviter().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil // 没有被邀请记录
		}
		return nil, fmt.Errorf("query referral by invitee: %w", err)
	}

	inviterEmail := ""
	if r.Edges.Inviter != nil {
		inviterEmail = r.Edges.Inviter.Email
	}

	return &ReferralRecord{
		ID:                    r.ID,
		InviterID:             r.InviterID,
		InviteeID:             r.InviteeID,
		InviterEmail:          inviterEmail,
		Code:                  r.Code,
		InviterRewarded:       r.InviterRewarded,
		InviteeRewarded:       r.InviteeRewarded,
		InviterRewardSnapshot: r.InviterRewardSnapshot,
		InviteeRewardSnapshot: r.InviteeRewardSnapshot,
		RewardGrantedAt:       r.RewardGrantedAt,
		CreatedAt:             r.CreatedAt,
	}, nil
}

// GetInviteCount 获取邀请人数（admin 用）
func (s *ReferralService) GetInviteCount(ctx context.Context, userID int64) (int, error) {
	return s.entClient.UserReferral.Query().
		Where(userreferral.InviterID(userID)).
		Count(ctx)
}

// GrantFirstRechargeReward 首次充值/付款时触发邀请奖励发放。
// 幂等：原子 UPDATE reward_granted_at IS NULL 保证只发一次。
// 始终自建事务，失败整体回滚，下次充值/付款时重试。
func (s *ReferralService) GrantFirstRechargeReward(ctx context.Context, inviteeID int64) error {
	// 快速检查：是否有待发放的邀请关系
	ref, err := s.entClient.UserReferral.Query().
		Where(
			userreferral.InviteeID(inviteeID),
			userreferral.RewardGrantedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil // 非被邀请用户或奖励已发放
		}
		return fmt.Errorf("query pending referral: %w", err)
	}

	inviterID := ref.InviterID
	inviterReward := ref.InviterRewardSnapshot
	inviteeReward := ref.InviteeRewardSnapshot

	// 自建事务：claim + 发余额 + 写审计记录，全部原子
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin referral reward tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	// 原子标记发放（WHERE reward_granted_at IS NULL 保证幂等）
	now := time.Now()
	affected, err := tx.UserReferral.Update().
		Where(
			userreferral.ID(ref.ID),
			userreferral.RewardGrantedAtIsNil(),
		).
		SetRewardGrantedAt(now).
		SetInviterRewarded(inviterReward).
		SetInviteeRewarded(inviteeReward).
		Save(txCtx)
	if err != nil {
		return fmt.Errorf("claim referral reward: %w", err)
	}
	if affected == 0 {
		return nil // 被其他并发请求抢先 claim 了
	}

	// 给邀请人加余额 + 写审计记录
	if inviterReward > 0 {
		if err := s.userRepo.UpdateBalance(txCtx, inviterID, inviterReward); err != nil {
			return fmt.Errorf("inviter balance: %w", err)
		}
		inviterCode, err := generateRandomCode(16) // 128-bit
		if err != nil {
			return fmt.Errorf("generate inviter redeem code: %w", err)
		}
		if _, err := tx.RedeemCode.Create().
			SetCode(inviterCode).
			SetType(AdjustmentTypeReferralInviter).
			SetValue(inviterReward).
			SetStatus(StatusUsed).
			SetUsedBy(inviterID).
			SetUsedAt(now).
			SetNotes(fmt.Sprintf("邀请用户充值奖励 (用户ID: %d)", inviteeID)).
			Save(txCtx); err != nil {
			return fmt.Errorf("create inviter redeem record: %w", err)
		}
	}

	// 给被邀请人加余额 + 写审计记录
	if inviteeReward > 0 {
		if err := s.userRepo.UpdateBalance(txCtx, inviteeID, inviteeReward); err != nil {
			return fmt.Errorf("invitee balance: %w", err)
		}
		inviteeCode, err := generateRandomCode(16) // 128-bit
		if err != nil {
			return fmt.Errorf("generate invitee redeem code: %w", err)
		}
		if _, err := tx.RedeemCode.Create().
			SetCode(inviteeCode).
			SetType(AdjustmentTypeReferralInvitee).
			SetValue(inviteeReward).
			SetStatus(StatusUsed).
			SetUsedBy(inviteeID).
			SetUsedAt(now).
			SetNotes(fmt.Sprintf("通过邀请充值奖励 (邀请人ID: %d)", inviterID)).
			Save(txCtx); err != nil {
			return fmt.Errorf("create invitee redeem record: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit referral reward: %w", err)
	}

	// Post-commit: 异步失效缓存
	go s.invalidateReferralCaches(context.Background(), inviterID, inviteeID, inviterReward, inviteeReward)

	logger.LegacyPrintf("service.referral",
		"[Referral] Reward granted on first recharge: inviter=%d invitee=%d reward_inviter=%.2f reward_invitee=%.2f",
		inviterID, inviteeID, inviterReward, inviteeReward)

	return nil
}

// generateRandomCode 生成随机 hex 码
func generateRandomCode(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// maskEmail 邮箱脱敏
func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return "***"
	}
	name := parts[0]
	if len(name) <= 2 {
		return name[:1] + "***@" + parts[1]
	}
	return name[:2] + "***@" + parts[1]
}
