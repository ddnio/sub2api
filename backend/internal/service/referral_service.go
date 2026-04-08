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
	settingService       *SettingService
	billingCacheService  *BillingCacheService
	authCacheInvalidator APIKeyAuthCacheInvalidator
}

// NewReferralService 创建推荐码服务实例
func NewReferralService(
	entClient *dbent.Client,
	userRepo UserRepository,
	settingService *SettingService,
	billingCacheService *BillingCacheService,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
) *ReferralService {
	return &ReferralService{
		entClient:            entClient,
		userRepo:             userRepo,
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
	InviterRewardAmt float64 `json:"inviter_reward_amount"` // 当前配置的邀请人奖励
	InviteeRewardAmt float64 `json:"invitee_reward_amount"` // 当前配置的被邀请人奖励
}

// ReferralRecord 邀请记录
type ReferralRecord struct {
	ID              int64     `json:"id"`
	InviterID       int64     `json:"inviter_id"`
	InviteeID       int64     `json:"invitee_id"`
	InviteeEmail    string    `json:"invitee_email"`
	Code            string    `json:"code"`
	InviterRewarded float64   `json:"inviter_rewarded"`
	InviteeRewarded float64   `json:"invitee_rewarded"`
	CreatedAt       time.Time `json:"created_at"`
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

// ProcessRegistrationReferral 处理注册时的推荐码归因和奖励
// 此方法为非关键操作：失败不影响注册，只记录日志并返回 error
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

	// 获取奖励金额配置
	inviterAmount := s.settingService.GetReferralInviterAmount(ctx)
	inviteeAmount := s.settingService.GetReferralInviteeAmount(ctx)

	// 使用事务：INSERT referral + UPDATE 双方余额
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin referral transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	// 创建邀请关系记录
	_, err = tx.UserReferral.Create().
		SetInviterID(inviter.ID).
		SetInviteeID(inviteeID).
		SetCode(referralCode).
		SetInviterRewarded(inviterAmount).
		SetInviteeRewarded(inviteeAmount).
		Save(txCtx)
	if err != nil {
		// 可能是 UNIQUE(invitee_id) 冲突——已被邀请过
		logger.LegacyPrintf("service.referral", "[Referral] Failed to create referral record: inviter=%d invitee=%d error=%v", inviter.ID, inviteeID, err)
		return fmt.Errorf("create referral record: %w", err)
	}

	// 给邀请人加余额
	if inviterAmount > 0 {
		if err := s.userRepo.UpdateBalance(txCtx, inviter.ID, inviterAmount); err != nil {
			return fmt.Errorf("update inviter balance: %w", err)
		}
	}

	// 给被邀请人加额外余额
	if inviteeAmount > 0 {
		if err := s.userRepo.UpdateBalance(txCtx, inviteeID, inviteeAmount); err != nil {
			return fmt.Errorf("update invitee balance: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit referral transaction: %w", err)
	}

	// Post-commit: 失效缓存（非关键，失败不影响）
	s.invalidateReferralCaches(ctx, inviter.ID, inviteeID, inviterAmount, inviteeAmount)

	logger.LegacyPrintf("service.referral",
		"[Referral] Success: inviter=%d invitee=%d code=%s reward_inviter=%.2f reward_invitee=%.2f",
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

// GetReferralInfo 获取用户的推荐码信息和统计
func (s *ReferralService) GetReferralInfo(ctx context.Context, userID int64) (*ReferralInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 如果用户还没有推荐码，惰性生成
	referralCode := ""
	if user.ReferralCode != nil {
		referralCode = *user.ReferralCode
	}
	if referralCode == "" {
		code, err := s.GenerateReferralCode(ctx, userID)
		if err != nil {
			logger.LegacyPrintf("service.referral", "[Referral] Lazy generate code failed for user %d: %v", userID, err)
			// 不阻断，返回空码
		} else {
			referralCode = code
		}
	}

	// 统计邀请人数和奖励总额
	totalInvited, err := s.entClient.UserReferral.Query().
		Where(userreferral.InviterID(userID)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count referrals: %w", err)
	}

	// 汇总奖励金额
	var totalRewarded float64
	referrals, err := s.entClient.UserReferral.Query().
		Where(userreferral.InviterID(userID)).
		All(ctx)
	if err == nil {
		for _, r := range referrals {
			totalRewarded += r.InviterRewarded
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
		InviterRewardAmt: inviterAmt,
		InviteeRewardAmt: inviteeAmt,
	}, nil
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
			ID:              r.ID,
			InviterID:       r.InviterID,
			InviteeID:       r.InviteeID,
			InviteeEmail:    email,
			Code:            r.Code,
			InviterRewarded: r.InviterRewarded,
			InviteeRewarded: r.InviteeRewarded,
			CreatedAt:       r.CreatedAt,
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
		ID:              r.ID,
		InviterID:       r.InviterID,
		InviteeID:       r.InviteeID,
		InviteeEmail:    inviterEmail, // admin 看到的是邀请人邮箱
		Code:            r.Code,
		InviterRewarded: r.InviterRewarded,
		InviteeRewarded: r.InviteeRewarded,
		CreatedAt:       r.CreatedAt,
	}, nil
}

// GetInviteCount 获取邀请人数（admin 用）
func (s *ReferralService) GetInviteCount(ctx context.Context, userID int64) (int, error) {
	return s.entClient.UserReferral.Query().
		Where(userreferral.InviterID(userID)).
		Count(ctx)
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
