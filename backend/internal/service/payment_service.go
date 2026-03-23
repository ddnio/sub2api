// backend/internal/service/payment_service.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// Errors
var (
	ErrPaymentOrderNotFound  = infraerrors.NotFound("PAYMENT_ORDER_NOT_FOUND", "payment order not found")
	ErrPaymentPlanNotFound   = infraerrors.NotFound("PAYMENT_PLAN_NOT_FOUND", "payment plan not found")
	ErrPaymentPlanInactive   = infraerrors.BadRequest("PAYMENT_PLAN_INACTIVE", "payment plan is not active")
	ErrPaymentOrderExpired   = infraerrors.BadRequest("PAYMENT_ORDER_EXPIRED", "payment order has expired")
	ErrPaymentAmountMismatch = infraerrors.BadRequest("PAYMENT_AMOUNT_MISMATCH", "callback amount does not match order amount")
	ErrPaymentOrderProcessed = infraerrors.Conflict("PAYMENT_ORDER_PROCESSED", "payment order already processed")
	ErrPaymentRateLimited    = infraerrors.TooManyRequests("PAYMENT_RATE_LIMITED", "too many payment orders, please try again later")
	ErrPaymentAmountInvalid  = infraerrors.BadRequest("PAYMENT_AMOUNT_INVALID", "payment amount out of allowed range")
	ErrPaymentProviderError  = infraerrors.InternalServer("PAYMENT_PROVIDER_ERROR", "payment provider error")
	ErrPaymentDeliveryFailed = infraerrors.InternalServer("PAYMENT_DELIVERY_FAILED", "failed to deliver payment benefits")
	ErrPaymentInvalidStatus  = infraerrors.BadRequest("PAYMENT_INVALID_STATUS", "invalid order status for this operation")
)

const (
	paymentMaxOrdersPerHour = 10
	paymentLockDuration     = 30 * time.Second
)

// --- Domain Types ---

type PaymentPlan struct {
	ID            int64
	Name          string
	Description   string
	Badge         *string
	GroupID       int64
	GroupName     string
	DurationDays  int
	Price         float64
	OriginalPrice *float64
	SortOrder     int
	IsActive      bool
	// Group limits (joined)
	DailyLimitUSD   float64
	WeeklyLimitUSD  float64
	MonthlyLimitUSD float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PaymentOrder struct {
	ID              int64
	OrderNo         string
	UserID          int64
	Type            string
	PlanID          *int64
	Amount          float64
	CreditAmount    *float64
	Currency        string
	Status          string
	Provider        *string
	ProviderOrderNo *string
	PaidAt          *time.Time
	CompletedAt     *time.Time
	RefundedAt      *time.Time
	ExpiredAt       time.Time
	CallbackRaw     *string
	AdminNote       *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	// Joined
	Plan *PaymentPlan
	User *User
}

type CreateOrderInput struct {
	UserID   int64
	Type     string  // plan / topup
	PlanID   *int64  // for plan orders
	Amount   float64 // for topup orders
	Provider string  // alipay / wxpay
}

type OrderFilter struct {
	Status string
	Type   string
	UserID *int64
	Search string
}

type StatsFilter struct {
	StartDate string
	EndDate   string
	GroupBy   string // day / month
}

type OrderStats struct {
	TotalAmount     float64          `json:"total_amount"`
	TotalOrders     int              `json:"total_orders"`
	PaidAmount      float64          `json:"paid_amount"`
	PaidOrders      int              `json:"paid_orders"`
	CompletedAmount float64          `json:"completed_amount"`
	CompletedOrders int              `json:"completed_orders"`
	Breakdown       []StatsBreakdown `json:"breakdown"`
}

type StatsBreakdown struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

// --- Interfaces ---

type PaymentProvider interface {
	CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResult, error)
	ParseCallback(r *http.Request) (*CallbackResult, error)
}

type PaymentRequest struct {
	OrderNo  string
	Amount   float64
	Provider string // alipay / wxpay
	Subject  string
}

type PaymentResult struct {
	QRCodeURL string
}

type CallbackResult struct {
	OrderNo         string
	ProviderOrderNo string
	Amount          float64
	Raw             string
}

type PaymentCache interface {
	AcquireCallbackLock(ctx context.Context, orderNo string, ttl time.Duration) (bool, error)
	ReleaseCallbackLock(ctx context.Context, orderNo string) error
	GetOrderCreateCount(ctx context.Context, userID int64) (int, error)
	IncrementOrderCreateCount(ctx context.Context, userID int64) error
}

type PaymentOrderRepository interface {
	Create(ctx context.Context, order *PaymentOrder) error
	GetByOrderNo(ctx context.Context, orderNo string) (*PaymentOrder, error)
	GetByID(ctx context.Context, id int64) (*PaymentOrder, error)
	UpdateStatusAtomically(ctx context.Context, orderNo string, fromStatuses []string, toStatus string, updates map[string]any) (int, error)
	ListByUser(ctx context.Context, userID int64, filter OrderFilter, params pagination.PaginationParams) ([]PaymentOrder, *pagination.PaginationResult, error)
	ListAll(ctx context.Context, filter OrderFilter, params pagination.PaginationParams) ([]PaymentOrder, *pagination.PaginationResult, error)
	ExpirePendingOrders(ctx context.Context) (int, error)
	Stats(ctx context.Context, filter StatsFilter) (*OrderStats, error)
}

type PaymentPlanRepository interface {
	Create(ctx context.Context, plan *PaymentPlan) error
	Update(ctx context.Context, id int64, updates map[string]any) (*PaymentPlan, error)
	GetByID(ctx context.Context, id int64) (*PaymentPlan, error)
	GetByIDActive(ctx context.Context, id int64) (*PaymentPlan, error)
	ListActive(ctx context.Context) ([]PaymentPlan, error)
	ListAll(ctx context.Context, params pagination.PaginationParams) ([]PaymentPlan, *pagination.PaginationResult, error)
	SoftDelete(ctx context.Context, id int64) error
}

// --- Service ---

type PaymentService struct {
	orderRepo           PaymentOrderRepository
	planRepo            PaymentPlanRepository
	provider            PaymentProvider
	cache               PaymentCache
	userService         *UserService
	subscriptionService *SubscriptionService
	billingCacheService *BillingCacheService
	entClient           *dbent.Client
	orderExpirySec      int
	minTopupAmount      float64
	maxTopupAmount      float64
}

func NewPaymentService(
	orderRepo PaymentOrderRepository,
	planRepo PaymentPlanRepository,
	provider PaymentProvider,
	cache PaymentCache,
	userService *UserService,
	subscriptionService *SubscriptionService,
	billingCacheService *BillingCacheService,
	entClient *dbent.Client,
	orderExpirySec int,
	minTopupAmount float64,
	maxTopupAmount float64,
) *PaymentService {
	if orderExpirySec <= 0 {
		orderExpirySec = 900
	}
	if minTopupAmount <= 0 {
		minTopupAmount = 1.0
	}
	if maxTopupAmount <= 0 {
		maxTopupAmount = 10000.0
	}
	return &PaymentService{
		orderRepo:           orderRepo,
		planRepo:            planRepo,
		provider:            provider,
		cache:               cache,
		userService:         userService,
		subscriptionService: subscriptionService,
		billingCacheService: billingCacheService,
		entClient:           entClient,
		orderExpirySec:      orderExpirySec,
		minTopupAmount:      minTopupAmount,
		maxTopupAmount:      maxTopupAmount,
	}
}

// --- Order Number Generation ---

func generateOrderNo() string {
	ts := time.Now().Format("20060102150405")
	b := make([]byte, 9) // 18 hex chars
	_, _ = rand.Read(b)
	return ts + hex.EncodeToString(b) // 14 + 18 = 32 chars
}

// --- Plan Methods ---

func (s *PaymentService) ListActivePlans(ctx context.Context) ([]PaymentPlan, error) {
	return s.planRepo.ListActive(ctx)
}

func (s *PaymentService) GetPlan(ctx context.Context, id int64) (*PaymentPlan, error) {
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrPaymentPlanNotFound
	}
	return plan, nil
}

func (s *PaymentService) ListAllPlans(ctx context.Context, params pagination.PaginationParams) ([]PaymentPlan, *pagination.PaginationResult, error) {
	return s.planRepo.ListAll(ctx, params)
}

func (s *PaymentService) CreatePlan(ctx context.Context, plan *PaymentPlan) error {
	return s.planRepo.Create(ctx, plan)
}

func (s *PaymentService) UpdatePlan(ctx context.Context, id int64, updates map[string]any) (*PaymentPlan, error) {
	return s.planRepo.Update(ctx, id, updates)
}

func (s *PaymentService) DeletePlan(ctx context.Context, id int64) error {
	return s.planRepo.SoftDelete(ctx, id)
}

// --- Order Methods ---

func (s *PaymentService) CreateOrder(ctx context.Context, input CreateOrderInput) (*PaymentOrder, *PaymentResult, error) {
	// Rate limit check
	if s.cache != nil {
		count, err := s.cache.GetOrderCreateCount(ctx, input.UserID)
		if err == nil && count >= paymentMaxOrdersPerHour {
			return nil, nil, ErrPaymentRateLimited
		}
	}

	var amount float64
	var subject string
	var planID *int64

	// C1: Validate provider
	if input.Provider != "wxpay" && input.Provider != "alipay" {
		return nil, nil, infraerrors.BadRequest("PAYMENT_INVALID_PROVIDER", "invalid payment provider, must be wxpay or alipay")
	}

	switch input.Type {
	case domain.PaymentOrderTypePlan:
		if input.PlanID == nil {
			return nil, nil, infraerrors.BadRequest("PAYMENT_PLAN_REQUIRED", "plan_id is required for plan orders")
		}
		plan, err := s.planRepo.GetByIDActive(ctx, *input.PlanID)
		if err != nil {
			return nil, nil, err
		}
		if plan == nil {
			return nil, nil, ErrPaymentPlanNotFound
		}
		amount = plan.Price
		subject = fmt.Sprintf("订阅套餐: %s", plan.Name)
		planID = &plan.ID

	case domain.PaymentOrderTypeTopup:
		if input.Amount < s.minTopupAmount || input.Amount > s.maxTopupAmount {
			return nil, nil, ErrPaymentAmountInvalid
		}
		amount = input.Amount
		subject = fmt.Sprintf("余额充值: ¥%.2f", amount)

	default:
		return nil, nil, infraerrors.BadRequest("PAYMENT_INVALID_TYPE", "invalid order type")
	}

	orderNo := generateOrderNo()
	expiredAt := time.Now().Add(time.Duration(s.orderExpirySec) * time.Second)

	order := &PaymentOrder{
		OrderNo:   orderNo,
		UserID:    input.UserID,
		Type:      input.Type,
		PlanID:    planID,
		Amount:    amount,
		Currency:  domain.PaymentCurrencyCNY,
		Status:    domain.PaymentStatusPending,
		Provider:  &input.Provider,
		ExpiredAt: expiredAt,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, nil, err
	}

	// Call payment provider
	result, err := s.provider.CreatePayment(ctx, PaymentRequest{
		OrderNo:  orderNo,
		Amount:   amount,
		Provider: input.Provider,
		Subject:  subject,
	})
	if err != nil {
		// I7: Mark order as failed immediately so no orphan pending records accumulate
		_, _ = s.orderRepo.UpdateStatusAtomically(ctx, orderNo,
			[]string{domain.PaymentStatusPending},
			domain.PaymentStatusFailed,
			map[string]any{},
		)
		log.Printf("[Payment] Provider error for order %s: %v", orderNo, err)
		return nil, nil, ErrPaymentProviderError
	}

	// Increment rate limit counter
	if s.cache != nil {
		_ = s.cache.IncrementOrderCreateCount(ctx, input.UserID)
	}

	return order, result, nil
}

func (s *PaymentService) GetOrderStatus(ctx context.Context, userID int64, orderID int64) (string, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return "", err
	}
	if order == nil || order.UserID != userID {
		return "", ErrPaymentOrderNotFound
	}
	return order.Status, nil
}

func (s *PaymentService) ListUserOrders(ctx context.Context, userID int64, filter OrderFilter, params pagination.PaginationParams) ([]PaymentOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.ListByUser(ctx, userID, filter, params)
}

func (s *PaymentService) ListAllOrders(ctx context.Context, filter OrderFilter, params pagination.PaginationParams) ([]PaymentOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.ListAll(ctx, filter, params)
}

func (s *PaymentService) GetOrder(ctx context.Context, id int64) (*PaymentOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrPaymentOrderNotFound
	}
	return order, nil
}

func (s *PaymentService) GetOrderStats(ctx context.Context, filter StatsFilter) (*OrderStats, error) {
	return s.orderRepo.Stats(ctx, filter)
}

// --- Callback Processing ---

func (s *PaymentService) ProcessCallback(ctx context.Context, r *http.Request) error {
	// 1. Parse and verify callback
	result, err := s.provider.ParseCallback(r)
	if err != nil {
		return infraerrors.BadRequest("PAYMENT_CALLBACK_INVALID", "invalid payment callback: "+err.Error())
	}

	// 2. Lookup order
	order, err := s.orderRepo.GetByOrderNo(ctx, result.OrderNo)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrPaymentOrderNotFound
	}

	// 3. Verify amount
	if fmt.Sprintf("%.2f", result.Amount) != fmt.Sprintf("%.2f", order.Amount) {
		log.Printf("[Payment] Amount mismatch for order %s: callback=%.2f, order=%.2f", order.OrderNo, result.Amount, order.Amount)
		return ErrPaymentAmountMismatch
	}

	// 4. Acquire distributed lock (degrade if Redis unavailable)
	if s.cache != nil {
		locked, lockErr := s.cache.AcquireCallbackLock(ctx, order.OrderNo, paymentLockDuration)
		if lockErr == nil && !locked {
			return nil // Another worker is processing, return success
		}
		if lockErr == nil {
			defer s.cache.ReleaseCallbackLock(ctx, order.OrderNo)
		}
		// If lockErr != nil, degrade: proceed without lock, rely on DB optimistic lock
	}

	// 5. Atomically transition status (optimistic lock)
	now := time.Now()
	affected, err := s.orderRepo.UpdateStatusAtomically(ctx, order.OrderNo,
		[]string{domain.PaymentStatusPending, domain.PaymentStatusExpired},
		domain.PaymentStatusPaid,
		map[string]any{
			"paid_at":           now,
			"provider_order_no": result.ProviderOrderNo,
			"callback_raw":      result.Raw,
		},
	)
	if err != nil {
		return err
	}
	if affected == 0 {
		return nil // Already processed, idempotent success
	}

	// 6. Deliver benefits in transaction
	deliverErr := s.deliverBenefits(ctx, order)
	if deliverErr != nil {
		log.Printf("[Payment] Failed to deliver benefits for order %s: %v", order.OrderNo, deliverErr)
		// Order stays at 'paid', admin can manually complete
		_, _ = s.orderRepo.UpdateStatusAtomically(ctx, order.OrderNo,
			[]string{domain.PaymentStatusPaid},
			domain.PaymentStatusFailed,
			map[string]any{},
		)
		return nil // Return success to payment provider to stop retries
	}

	// 7. Mark completed
	completedAt := time.Now()
	_, _ = s.orderRepo.UpdateStatusAtomically(ctx, order.OrderNo,
		[]string{domain.PaymentStatusPaid},
		domain.PaymentStatusCompleted,
		map[string]any{"completed_at": completedAt},
	)

	return nil
}

func (s *PaymentService) deliverBenefits(ctx context.Context, order *PaymentOrder) error {
	switch order.Type {
	case domain.PaymentOrderTypeTopup:
		creditAmount := order.Amount // v1: credit_amount == amount
		err := s.userService.UpdateBalance(ctx, order.UserID, creditAmount)
		if err != nil {
			return fmt.Errorf("update balance: %w", err)
		}
		// Update credit_amount on order
		_, _ = s.orderRepo.UpdateStatusAtomically(ctx, order.OrderNo,
			[]string{domain.PaymentStatusPaid},
			domain.PaymentStatusPaid,
			map[string]any{"credit_amount": creditAmount},
		)
		s.asyncInvalidateCache(order.UserID)
		return nil

	case domain.PaymentOrderTypePlan:
		if order.PlanID == nil {
			return fmt.Errorf("plan order missing plan_id")
		}
		plan, err := s.planRepo.GetByID(ctx, *order.PlanID)
		if err != nil || plan == nil {
			return fmt.Errorf("get plan: %w", err)
		}
		_, _, err = s.subscriptionService.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
			UserID:       order.UserID,
			GroupID:      plan.GroupID,
			ValidityDays: plan.DurationDays,
			AssignedBy:   0,
			Notes:        fmt.Sprintf("Payment order: %s", order.OrderNo),
		})
		if err != nil {
			return fmt.Errorf("assign subscription: %w", err)
		}
		s.asyncInvalidateCache(order.UserID)
		return nil

	default:
		return fmt.Errorf("unknown order type: %s", order.Type)
	}
}

func (s *PaymentService) asyncInvalidateCache(userID int64) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if s.billingCacheService != nil {
			_ = s.billingCacheService.InvalidateUserBalance(ctx, userID)
		}
	}()
}

// --- Admin Operations ---

func (s *PaymentService) AdminCompleteOrder(ctx context.Context, orderID int64, adminNote string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrPaymentOrderNotFound
	}
	// Only allow completing "failed" orders. "paid" orders may have already had benefits
	// delivered (crash between deliverBenefits and status→completed), so re-running
	// deliverBenefits on them risks double-crediting. Failed orders have never had benefits
	// delivered, making them safe to complete manually.
	if order.Status != domain.PaymentStatusFailed {
		return ErrPaymentInvalidStatus
	}

	log.Printf("[Payment] AdminCompleteOrder: order=%s status=%s", order.OrderNo, order.Status)
	deliverErr := s.deliverBenefits(ctx, order)
	if deliverErr != nil {
		return ErrPaymentDeliveryFailed
	}

	completedAt := time.Now()
	_, err = s.orderRepo.UpdateStatusAtomically(ctx, order.OrderNo,
		[]string{domain.PaymentStatusFailed},
		domain.PaymentStatusCompleted,
		map[string]any{
			"completed_at": completedAt,
			"admin_note":   adminNote,
		},
	)
	return err
}

func (s *PaymentService) AdminRefundOrder(ctx context.Context, orderID int64, adminNote string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrPaymentOrderNotFound
	}
	if order.Status != domain.PaymentStatusCompleted && order.Status != domain.PaymentStatusPaid {
		return ErrPaymentInvalidStatus
	}

	// C2: Reverse benefits for completed topup orders
	if order.Status == domain.PaymentStatusCompleted && order.Type == domain.PaymentOrderTypeTopup {
		creditAmount := order.Amount
		if order.CreditAmount != nil {
			creditAmount = *order.CreditAmount
		}
		if err := s.userService.UpdateBalance(ctx, order.UserID, -creditAmount); err != nil {
			log.Printf("[Payment] Refund: failed to deduct balance for order %s: %v", order.OrderNo, err)
			return ErrPaymentDeliveryFailed
		}
		s.asyncInvalidateCache(order.UserID)
	} else if order.Status == domain.PaymentStatusCompleted && order.Type == domain.PaymentOrderTypePlan {
		log.Printf("[Payment] WARN: Refunding plan order %s - subscription benefits not reversed automatically, handle manually", order.OrderNo)
	}

	refundedAt := time.Now()
	_, err = s.orderRepo.UpdateStatusAtomically(ctx, order.OrderNo,
		[]string{domain.PaymentStatusCompleted, domain.PaymentStatusPaid},
		domain.PaymentStatusRefunded,
		map[string]any{
			"refunded_at": refundedAt,
			"admin_note":  adminNote,
		},
	)
	return err
}
