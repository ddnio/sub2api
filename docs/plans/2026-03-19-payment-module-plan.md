# 支付模块实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 Sub2API 添加支付功能，支持用户购买订阅套餐和充值余额（支付宝/微信扫码）。

**Architecture:** 在现有 Go 后端新增 payment 领域模块，遵循 service/repository/handler 分层。PaymentProvider 抽象接口隔离第三方支付渠道。Ent ORM 新增 payment_plans 和 payment_orders 两张表。

**Tech Stack:** Go 1.26+, Gin, Ent ORM, PostgreSQL 16, Redis, Wire DI

**Spec:** `docs/plans/2026-03-19-payment-module-design.md`

---

## 文件结构

### 新建文件

| 文件 | 职责 |
|------|------|
| `backend/ent/schema/payment_plan.go` | Ent schema：套餐表 |
| `backend/ent/schema/payment_order.go` | Ent schema：订单表 |
| `backend/internal/domain/payment_constants.go` | 支付相关常量 |
| `backend/internal/service/payment_service.go` | 领域类型、接口、错误变量、PaymentService |
| `backend/internal/service/payment_expiry_service.go` | 过期订单清理 worker |
| `backend/internal/repository/payment_order_repo.go` | PaymentOrderRepository 实现 |
| `backend/internal/repository/payment_plan_repo.go` | PaymentPlanRepository 实现 |
| `backend/internal/repository/payment_cache.go` | PaymentCache Redis 实现 |
| `backend/internal/repository/easypay_provider.go` | PaymentProvider EasyPay 实现 |
| `backend/internal/handler/payment_handler.go` | 用户端 handler |
| `backend/internal/handler/payment_callback_handler.go` | 支付回调 handler |
| `backend/internal/handler/admin/payment_plan_handler.go` | 管理端套餐 handler |
| `backend/internal/handler/admin/payment_order_handler.go` | 管理端订单 handler |

### 修改文件

| 文件 | 修改内容 |
|------|---------|
| `backend/internal/config/config.go` | 新增 `PaymentConfig` |
| `backend/internal/handler/handler.go` | Handlers/AdminHandlers 加字段 |
| `backend/internal/handler/wire.go` | ProvideHandlers/ProvideAdminHandlers 加参数，ProviderSet 加条目 |
| `backend/internal/handler/dto/types.go` | 新增 Payment DTO 结构体 |
| `backend/internal/handler/dto/mappers.go` | 新增 DTO 转换函数 |
| `backend/internal/service/wire.go` | 注册 ProvidePaymentService, ProvidePaymentExpiryService |
| `backend/internal/repository/wire.go` | 注册 NewPaymentOrderRepository, NewPaymentPlanRepository, NewPaymentCache, ProvideEasyPayProvider |
| `backend/internal/server/routes/user.go` | 注册用户支付路由 |
| `backend/internal/server/routes/admin.go` | 注册管理端支付路由 |
| `backend/internal/server/routes/router.go` | 注册回调路由（v1 顶层，不走 JWT） |
| `backend/cmd/server/wire.go` | provideCleanup 加 PaymentExpiryService 参数 |

---

## Task 1: 支付常量和配置

**Files:**
- Create: `backend/internal/domain/payment_constants.go`
- Modify: `backend/internal/config/config.go`

- [ ] **Step 1: 创建支付常量文件**

```go
// backend/internal/domain/payment_constants.go
package domain

// Payment order types
const (
	PaymentOrderTypePlan  = "plan"
	PaymentOrderTypeTopup = "topup"
)

// Payment order status
const (
	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusExpired   = "expired"
	PaymentStatusRefunded  = "refunded"
)

// Payment providers
const (
	PaymentProviderAlipay = "alipay"
	PaymentProviderWechat = "wechat"
)

// Default currency
const (
	PaymentCurrencyCNY = "CNY"
)
```

- [ ] **Step 2: 在 config.go 新增 PaymentConfig**

在 `backend/internal/config/config.go` 的 Config 结构体中，紧接已有字段添加：

```go
Payment PaymentConfig `mapstructure:"payment"`
```

在同一文件中添加 PaymentConfig 类型定义：

```go
type PaymentConfig struct {
	Provider        string  `mapstructure:"provider"`           // easypay
	EasyPayBaseURL  string  `mapstructure:"easypay_base_url"`
	EasyPayAppID    string  `mapstructure:"easypay_app_id"`
	EasyPaySignKey  string  `mapstructure:"easypay_sign_key"`
	CallbackBaseURL string  `mapstructure:"callback_base_url"`
	OrderExpirySec  int     `mapstructure:"order_expiry_sec"`   // default 900
	ExpiryTickSec   int     `mapstructure:"expiry_tick_sec"`    // default 60
	MinTopupAmount  float64 `mapstructure:"min_topup_amount"`   // default 1.00
	MaxTopupAmount  float64 `mapstructure:"max_topup_amount"`   // default 10000.00
}
```

- [ ] **Step 3: 验证编译**

Run: `cd backend && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/payment_constants.go backend/internal/config/config.go
git commit -m "feat(payment): add payment constants and config"
```

---

## Task 2: Ent Schema — payment_plans

**Files:**
- Create: `backend/ent/schema/payment_plan.go`

- [ ] **Step 1: 创建 payment_plan schema**

参考 `backend/ent/schema/redeem_code.go` 和 `backend/ent/schema/user_subscription.go` 的模式：

```go
// backend/ent/schema/payment_plan.go
package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PaymentPlan holds the schema definition for the PaymentPlan entity.
type PaymentPlan struct {
	ent.Schema
}

func (PaymentPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_plans"},
	}
}

func (PaymentPlan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (PaymentPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Unique(),
		field.String("description").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),
		field.String("badge").
			MaxLen(20).
			Optional().
			Nillable(),
		field.Int64("group_id"),
		field.Int("duration_days").
			Positive(),
		field.Float("price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("original_price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Optional().
			Nillable(),
		field.Int("sort_order").
			Default(0).
			Min(0),
		field.Bool("is_active").
			Default(true),
	}
}

func (PaymentPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("payment_plans").
			Field("group_id").
			Required().
			Unique(),
	}
}

func (PaymentPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("is_active", "sort_order"),
	}
}
```

- [ ] **Step 2: 在 Group schema 添加反向 edge**

在 `backend/ent/schema/group.go` 的 `Edges()` 方法中添加：

```go
edge.To("payment_plans", PaymentPlan.Type),
```

- [ ] **Step 3: 生成 Ent 代码**

Run: `cd backend && go generate ./ent`
Expected: 生成 `ent/paymentplan/` 目录和相关文件

- [ ] **Step 4: 验证编译**

Run: `cd backend && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 5: Commit**

```bash
git add backend/ent/
git commit -m "feat(payment): add payment_plans ent schema"
```

---

## Task 3: Ent Schema — payment_orders

**Files:**
- Create: `backend/ent/schema/payment_order.go`

- [ ] **Step 1: 创建 payment_order schema**

参考 `backend/ent/schema/redeem_code.go`（硬删除模式，手动 timestamp）：

```go
// backend/ent/schema/payment_order.go
package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PaymentOrder holds the schema definition for the PaymentOrder entity.
type PaymentOrder struct {
	ent.Schema
}

func (PaymentOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_orders"},
	}
}

func (PaymentOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").
			MaxLen(32).
			NotEmpty().
			Unique(),
		field.Int64("user_id"),
		field.String("type").
			MaxLen(20).
			NotEmpty(),
		field.Int64("plan_id").
			Optional().
			Nillable(),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("credit_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Optional().
			Nillable(),
		field.String("currency").
			MaxLen(3).
			Default(domain.PaymentCurrencyCNY),
		field.String("status").
			MaxLen(20).
			Default(domain.PaymentStatusPending),
		field.String("provider").
			MaxLen(20).
			Optional().
			Nillable(),
		field.String("provider_order_no").
			MaxLen(64).
			Optional().
			Nillable().
			Unique(),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("completed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("refunded_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expired_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("callback_raw").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		field.String("admin_note").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (PaymentOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("payment_orders").
			Field("user_id").
			Required().
			Unique(),
		edge.From("plan", PaymentPlan.Type).
			Ref("orders").
			Field("plan_id").
			Unique(),
	}
}

func (PaymentOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("plan_id"),
		index.Fields("status"),
		index.Fields("status", "expired_at"),
	}
}
```

- [ ] **Step 2: 在 User schema 添加反向 edge**

在 `backend/ent/schema/user.go` 的 `Edges()` 方法中添加：

```go
edge.To("payment_orders", PaymentOrder.Type),
```

- [ ] **Step 3: 在 PaymentPlan schema 添加 orders edge**

在 `backend/ent/schema/payment_plan.go` 的 `Edges()` 方法中添加：

```go
edge.To("orders", PaymentOrder.Type),
```

- [ ] **Step 4: 生成 Ent 代码**

Run: `cd backend && go generate ./ent`
Expected: 生成 `ent/paymentorder/` 目录和相关文件

- [ ] **Step 5: 验证编译**

Run: `cd backend && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 6: Commit**

```bash
git add backend/ent/
git commit -m "feat(payment): add payment_orders ent schema"
```

---

## Task 4: Service 层 — 领域类型和接口

**Files:**
- Create: `backend/internal/service/payment_service.go`

- [ ] **Step 1: 创建 payment_service.go — 错误变量、常量、领域类型、接口**

参考 `redeem_service.go:1-100` 的模式：

```go
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
	ErrPaymentOrderNotFound   = infraerrors.NotFound("PAYMENT_ORDER_NOT_FOUND", "payment order not found")
	ErrPaymentPlanNotFound    = infraerrors.NotFound("PAYMENT_PLAN_NOT_FOUND", "payment plan not found")
	ErrPaymentPlanInactive    = infraerrors.BadRequest("PAYMENT_PLAN_INACTIVE", "payment plan is not active")
	ErrPaymentOrderExpired    = infraerrors.BadRequest("PAYMENT_ORDER_EXPIRED", "payment order has expired")
	ErrPaymentAmountMismatch  = infraerrors.BadRequest("PAYMENT_AMOUNT_MISMATCH", "callback amount does not match order amount")
	ErrPaymentOrderProcessed  = infraerrors.Conflict("PAYMENT_ORDER_PROCESSED", "payment order already processed")
	ErrPaymentRateLimited     = infraerrors.TooManyRequests("PAYMENT_RATE_LIMITED", "too many payment orders, please try again later")
	ErrPaymentAmountInvalid   = infraerrors.BadRequest("PAYMENT_AMOUNT_INVALID", "payment amount out of allowed range")
	ErrPaymentProviderError   = infraerrors.Internal("PAYMENT_PROVIDER_ERROR", "payment provider error")
	ErrPaymentDeliveryFailed  = infraerrors.Internal("PAYMENT_DELIVERY_FAILED", "failed to deliver payment benefits")
	ErrPaymentInvalidStatus   = infraerrors.BadRequest("PAYMENT_INVALID_STATUS", "invalid order status for this operation")
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
	User *UserShallow
}

type CreateOrderInput struct {
	UserID   int64
	Type     string   // plan / topup
	PlanID   *int64   // for plan orders
	Amount   float64  // for topup orders
	Provider string   // alipay / wechat
}

type OrderFilter struct {
	Status   string
	Type     string
	UserID   *int64
	Search   string
}

type StatsFilter struct {
	StartDate string
	EndDate   string
	GroupBy   string // day / month
}

type OrderStats struct {
	TotalAmount float64          `json:"total_amount"`
	TotalOrders int              `json:"total_orders"`
	Breakdown   []StatsBreakdown `json:"breakdown"`
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
	Provider string // alipay / wechat
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

// NOTE: Wire 不能直接注入 primitive 参数（int, float64）。
// 在 service/wire.go 中需要创建 ProvidePaymentService 包装函数：
//
// func ProvidePaymentService(
//     orderRepo PaymentOrderRepository,
//     planRepo PaymentPlanRepository,
//     provider PaymentProvider,
//     cache PaymentCache,
//     userService *UserService,
//     subscriptionService *SubscriptionService,
//     billingCacheService *BillingCacheService,
//     entClient *dbent.Client,
//     cfg *config.Config,
// ) *PaymentService {
//     return NewPaymentService(
//         orderRepo, planRepo, provider, cache,
//         userService, subscriptionService, billingCacheService, entClient,
//         cfg.Payment.OrderExpirySec, cfg.Payment.MinTopupAmount, cfg.Payment.MaxTopupAmount,
//     )
// }
//
// 同样，在 repository/wire.go 中需要：
//
// func ProvideEasyPayProvider(cfg *config.Config) service.PaymentProvider {
//     return NewEasyPayProvider(cfg.Payment)
// }
//
// 注册到 ProviderSet 的是 ProvideXxx，而不是 NewXxx。

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
			"callback_raw":     result.Raw,
		},
	)
	if err != nil {
		return err
	}
	if affected == 0 {
		return nil // Already processed, idempotent success
	}

	// 6. Deliver benefits in transaction
	deliverErr := s.deliverBenefits(ctx, order, result)
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

func (s *PaymentService) deliverBenefits(ctx context.Context, order *PaymentOrder, result *CallbackResult) error {
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
			domain.PaymentStatusPaid, // same status, just updating fields
			map[string]any{"credit_amount": creditAmount},
		)
		// Invalidate balance cache
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
		// 注意：AssignOrExtendSubscription 接受 *AssignSubscriptionInput 指针，返回 3 个值
		_, _, err = s.subscriptionService.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
			UserID:     order.UserID,
			GroupID:    plan.GroupID,
			Days:       plan.DurationDays,
			AssignedBy: 0, // System
			Notes:      fmt.Sprintf("Payment order: %s", order.OrderNo),
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
			s.billingCacheService.InvalidateUserBalance(ctx, userID)
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
	if order.Status != domain.PaymentStatusPaid && order.Status != domain.PaymentStatusFailed {
		return ErrPaymentInvalidStatus
	}

	deliverErr := s.deliverBenefits(ctx, order, nil)
	if deliverErr != nil {
		return ErrPaymentDeliveryFailed
	}

	completedAt := time.Now()
	_, err = s.orderRepo.UpdateStatusAtomically(ctx, order.OrderNo,
		[]string{domain.PaymentStatusPaid, domain.PaymentStatusFailed},
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
	if order.Status != domain.PaymentStatusCompleted {
		return ErrPaymentInvalidStatus
	}

	refundedAt := time.Now()
	_, err = s.orderRepo.UpdateStatusAtomically(ctx, order.OrderNo,
		[]string{domain.PaymentStatusCompleted},
		domain.PaymentStatusRefunded,
		map[string]any{
			"refunded_at": refundedAt,
			"admin_note":  adminNote,
		},
	)
	return err
}
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build ./...`
Expected: 可能因为 `AssignSubscriptionInput` 或 `UpdateBalance` 签名不匹配而有编译错误，需要对照实际代码调整

- [ ] **Step 3: 修复编译错误并验证**

检查 `AssignOrExtendSubscription` 实际签名，调整 `deliverBenefits` 中的调用参数。
检查 `UserService.UpdateBalance` 签名，确认接受 `float64`。
检查 `BillingCacheService.InvalidateUserBalance` 签名。

Run: `cd backend && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/service/payment_service.go
git commit -m "feat(payment): add payment service with domain types and business logic"
```

---

## Task 5: 过期订单清理 Worker

**Files:**
- Create: `backend/internal/service/payment_expiry_service.go`

- [ ] **Step 1: 创建过期清理 worker**

参考 `backend/internal/service/account_expiry_service.go` 的 Start/Stop 模式：

```go
// backend/internal/service/payment_expiry_service.go
package service

import (
	"context"
	"log"
	"sync"
	"time"
)

type PaymentExpiryService struct {
	orderRepo PaymentOrderRepository
	interval  time.Duration
	stopCh    chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

func NewPaymentExpiryService(orderRepo PaymentOrderRepository, interval time.Duration) *PaymentExpiryService {
	return &PaymentExpiryService{
		orderRepo: orderRepo,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

func (s *PaymentExpiryService) Start() {
	if s == nil || s.orderRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *PaymentExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *PaymentExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expired, err := s.orderRepo.ExpirePendingOrders(ctx)
	if err != nil {
		log.Printf("[PaymentExpiry] Failed to expire pending orders: %v", err)
		return
	}
	if expired > 0 {
		log.Printf("[PaymentExpiry] Expired %d pending orders", expired)
	}
}
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/payment_expiry_service.go
git commit -m "feat(payment): add payment order expiry background worker"
```

---

## Task 6: Repository 层实现

**Files:**
- Create: `backend/internal/repository/payment_plan_repo.go`
- Create: `backend/internal/repository/payment_order_repo.go`
- Create: `backend/internal/repository/payment_cache.go`

- [ ] **Step 1: 创建 payment_plan_repo.go**

参考 `backend/internal/repository/redeem_code_repo.go` 的未导出 struct + 返回接口模式：

```go
// backend/internal/repository/payment_plan_repo.go
package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentplan"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentPlanRepository struct {
	client *dbent.Client
}

func NewPaymentPlanRepository(client *dbent.Client) service.PaymentPlanRepository {
	return &paymentPlanRepository{client: client}
}
```

实现所有接口方法：`Create`, `Update`, `GetByID`（使用 `mixins.SkipSoftDelete(ctx)` 绕过软删除过滤）, `GetByIDActive`, `ListActive`（加 `WithGroup` edge 获取 Group 限额）, `ListAll`, `SoftDelete`。

- [ ] **Step 2: 创建 payment_order_repo.go**

```go
// backend/internal/repository/payment_order_repo.go
package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentOrderRepository struct {
	client *dbent.Client
}

func NewPaymentOrderRepository(client *dbent.Client) service.PaymentOrderRepository {
	return &paymentOrderRepository{client: client}
}
```

关键方法实现要点：
- `UpdateStatusAtomically`: `WHERE order_no = ? AND status IN (fromStatuses...)` + 返回 affected rows
- `ExpirePendingOrders`: `UPDATE payment_orders SET status='expired' WHERE status='pending' AND expired_at < NOW()`
- `Stats`: 按 `group_by` (day/month) 做 SQL 聚合查询
- `ListByUser` / `ListAll`: 支持 `OrderFilter` 的可选筛选条件

- [ ] **Step 3: 创建 payment_cache.go**

参考 redeem cache 的 Redis 锁模式：

```go
// backend/internal/repository/payment_cache.go
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

type paymentCache struct {
	rdb *redis.Client
}

func NewPaymentCache(rdb *redis.Client) service.PaymentCache {
	return &paymentCache{rdb: rdb}
}

func (c *paymentCache) AcquireCallbackLock(ctx context.Context, orderNo string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("payment:lock:%s", orderNo)
	return c.rdb.SetNX(ctx, key, "1", ttl).Result()
}

func (c *paymentCache) ReleaseCallbackLock(ctx context.Context, orderNo string) error {
	key := fmt.Sprintf("payment:lock:%s", orderNo)
	return c.rdb.Del(ctx, key).Err()
}

func (c *paymentCache) GetOrderCreateCount(ctx context.Context, userID int64) (int, error) {
	key := fmt.Sprintf("payment:rate:%d", userID)
	val, err := c.rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *paymentCache) IncrementOrderCreateCount(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("payment:rate:%d", userID)
	pipe := c.rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}
```

- [ ] **Step 4: 验证编译**

Run: `cd backend && go build ./...`
Expected: BUILD SUCCESS（可能需要调整 Ent 生成的类型名称）

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/payment_plan_repo.go backend/internal/repository/payment_order_repo.go backend/internal/repository/payment_cache.go
git commit -m "feat(payment): add payment repositories and cache implementation"
```

---

## Task 7: EasyPay Provider（桩实现）

**Files:**
- Create: `backend/internal/repository/easypay_provider.go`

- [ ] **Step 1: 创建 EasyPay provider 桩实现**

一期先实现接口骨架，具体对接等确定服务商后填充：

```go
// backend/internal/repository/easypay_provider.go
package repository

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type easypayProvider struct {
	cfg config.PaymentConfig
}

func NewEasyPayProvider(cfg config.PaymentConfig) service.PaymentProvider {
	return &easypayProvider{cfg: cfg}
}

func (p *easypayProvider) CreatePayment(ctx context.Context, req service.PaymentRequest) (*service.PaymentResult, error) {
	// TODO: 对接实际支付 API
	// 1. 构造签名参数
	// 2. 发送 HTTP 请求到 EasyPay API
	// 3. 解析返回的支付二维码 URL
	return nil, fmt.Errorf("easypay provider not implemented yet")
}

func (p *easypayProvider) ParseCallback(r *http.Request) (*service.CallbackResult, error) {
	// TODO: 对接实际支付回调
	// 1. 读取回调参数
	// 2. 验证签名
	// 3. 解析订单号和金额
	return nil, fmt.Errorf("easypay callback parser not implemented yet")
}
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add backend/internal/repository/easypay_provider.go
git commit -m "feat(payment): add easypay provider stub implementation"
```

---

## Task 8: DTO 和 Handler 层

**Files:**
- Modify: `backend/internal/handler/dto/mappers.go`
- Create: `backend/internal/handler/payment_handler.go`
- Create: `backend/internal/handler/payment_callback_handler.go`
- Create: `backend/internal/handler/admin/payment_plan_handler.go`
- Create: `backend/internal/handler/admin/payment_order_handler.go`

- [ ] **Step 1: 在 dto/types.go 添加 Payment DTO 结构体**

注意：DTO 结构体定义在 `dto/types.go`，转换函数在 `dto/mappers.go`，遵循现有代码分离惯例。

在 `backend/internal/handler/dto/types.go` 末尾添加：

```go
// --- Payment DTOs ---

type PaymentPlanDTO struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Badge           *string  `json:"badge,omitempty"`
	GroupName       string   `json:"group_name"`
	DurationDays    int      `json:"duration_days"`
	Price           float64  `json:"price"`
	OriginalPrice   *float64 `json:"original_price,omitempty"`
	DailyLimitUSD   float64  `json:"daily_limit_usd"`
	WeeklyLimitUSD  float64  `json:"weekly_limit_usd"`
	MonthlyLimitUSD float64  `json:"monthly_limit_usd"`
}

type AdminPaymentPlanDTO struct {
	PaymentPlanDTO
	GroupID   int64  `json:"group_id"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

type PaymentOrderDTO struct {
	ID          int64    `json:"id"`
	OrderNo     string   `json:"order_no"`
	Type        string   `json:"type"`
	Amount      float64  `json:"amount"`
	Currency    string   `json:"currency"`
	Status      string   `json:"status"`
	Provider    *string  `json:"provider,omitempty"`
	CreatedAt   string   `json:"created_at"`
	PaidAt      *string  `json:"paid_at,omitempty"`
	CompletedAt *string  `json:"completed_at,omitempty"`
	PlanName    *string  `json:"plan_name,omitempty"`
}

type AdminPaymentOrderDTO struct {
	PaymentOrderDTO
	UserID          int64   `json:"user_id"`
	PlanID          *int64  `json:"plan_id,omitempty"`
	CreditAmount    *float64 `json:"credit_amount,omitempty"`
	ProviderOrderNo *string `json:"provider_order_no,omitempty"`
	RefundedAt      *string `json:"refunded_at,omitempty"`
	AdminNote       *string `json:"admin_note,omitempty"`
}

```

在 `backend/internal/handler/dto/mappers.go` 末尾添加转换函数：

```go
func PaymentPlanFromService(p *service.PaymentPlan) PaymentPlanDTO {
	return PaymentPlanDTO{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		Badge:           p.Badge,
		GroupName:       p.GroupName,
		DurationDays:    p.DurationDays,
		Price:           p.Price,
		OriginalPrice:   p.OriginalPrice,
		DailyLimitUSD:   p.DailyLimitUSD,
		WeeklyLimitUSD:  p.WeeklyLimitUSD,
		MonthlyLimitUSD: p.MonthlyLimitUSD,
	}
}

func AdminPaymentPlanFromService(p *service.PaymentPlan) AdminPaymentPlanDTO {
	return AdminPaymentPlanDTO{
		PaymentPlanDTO: PaymentPlanFromService(p),
		GroupID:        p.GroupID,
		SortOrder:      p.SortOrder,
		IsActive:       p.IsActive,
	}
}

func PaymentOrderFromService(o *service.PaymentOrder) PaymentOrderDTO {
	dto := PaymentOrderDTO{
		ID:        o.ID,
		OrderNo:   o.OrderNo,
		Type:      o.Type,
		Amount:    o.Amount,
		Currency:  o.Currency,
		Status:    o.Status,
		Provider:  o.Provider,
		CreatedAt: o.CreatedAt.Format(time.RFC3339),
	}
	if o.PaidAt != nil {
		s := o.PaidAt.Format(time.RFC3339)
		dto.PaidAt = &s
	}
	if o.CompletedAt != nil {
		s := o.CompletedAt.Format(time.RFC3339)
		dto.CompletedAt = &s
	}
	if o.Plan != nil {
		dto.PlanName = &o.Plan.Name
	}
	return dto
}

func AdminPaymentOrderFromService(o *service.PaymentOrder) AdminPaymentOrderDTO {
	dto := AdminPaymentOrderDTO{
		PaymentOrderDTO: PaymentOrderFromService(o),
		UserID:          o.UserID,
		PlanID:          o.PlanID,
		CreditAmount:    o.CreditAmount,
		ProviderOrderNo: o.ProviderOrderNo,
		AdminNote:       o.AdminNote,
	}
	if o.RefundedAt != nil {
		s := o.RefundedAt.Format(time.RFC3339)
		dto.RefundedAt = &s
	}
	return dto
}
```

- [ ] **Step 2: 创建用户端 payment_handler.go**

参考 `backend/internal/handler/redeem_handler.go` 模式：

实现 4 个 handler 方法：
- `ListPlans(c *gin.Context)` — GET /payment/plans
- `CreateOrder(c *gin.Context)` — POST /payment/orders
- `ListOrders(c *gin.Context)` — GET /payment/orders
- `GetOrderStatus(c *gin.Context)` — GET /payment/orders/:id/status

- [ ] **Step 3: 创建回调 payment_callback_handler.go**

```go
type PaymentCallbackHandler struct {
	paymentService *service.PaymentService
}

func (h *PaymentCallbackHandler) Handle(c *gin.Context) {
	err := h.paymentService.ProcessCallback(c.Request.Context(), c.Request)
	if err != nil {
		c.String(http.StatusBadRequest, "FAIL")
		return
	}
	c.String(http.StatusOK, "SUCCESS")
}
```

- [ ] **Step 4: 创建管理端 payment_plan_handler.go**

参考 `admin/redeem_handler.go`，所有写操作使用 `executeAdminIdempotentJSON`。

- [ ] **Step 5: 创建管理端 payment_order_handler.go**

实现：`List`, `GetByID`, `Complete`, `Refund`, `Stats`。

- [ ] **Step 6: 验证编译**

Run: `cd backend && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/dto/mappers.go backend/internal/handler/payment_handler.go backend/internal/handler/payment_callback_handler.go backend/internal/handler/admin/payment_plan_handler.go backend/internal/handler/admin/payment_order_handler.go
git commit -m "feat(payment): add payment handlers and DTOs"
```

---

## Task 9: Wire 注册和路由接线

**Files:**
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/service/wire.go`
- Modify: `backend/internal/server/routes/user.go`
- Modify: `backend/internal/server/routes/admin.go`

- [ ] **Step 1: 在 handler.go 添加 handler 字段**

`Handlers` 结构体添加：
```go
Payment         *PaymentHandler
PaymentCallback *PaymentCallbackHandler
```

`AdminHandlers` 结构体添加：
```go
PaymentPlan  *admin.PaymentPlanHandler
PaymentOrder *admin.PaymentOrderHandler
```

- [ ] **Step 1b: 更新 handler/wire.go 的 ProvideHandlers 和 ProvideAdminHandlers**

在 `ProvideAdminHandlers` 函数中：
1. 添加参数 `paymentPlanHandler *admin.PaymentPlanHandler, paymentOrderHandler *admin.PaymentOrderHandler`
2. 在返回的 struct 中赋值 `PaymentPlan: paymentPlanHandler, PaymentOrder: paymentOrderHandler`

在 `ProvideHandlers` 函数中：
1. 添加参数 `paymentHandler *PaymentHandler, paymentCallbackHandler *PaymentCallbackHandler`
2. 在返回的 struct 中赋值 `Payment: paymentHandler, PaymentCallback: paymentCallbackHandler`

- [ ] **Step 1c: 更新 cmd/server/wire.go 的 provideCleanup**

在 `provideCleanup` 函数中：
1. 添加参数 `paymentExpiry *service.PaymentExpiryService`
2. 在 `parallelSteps` 中添加：`{name: "PaymentExpiry", fn: func() error { paymentExpiry.Stop(); return nil }}`

- [ ] **Step 2: 在三个 wire.go 中分别注册 Provider**

**2a. `backend/internal/service/wire.go`** — 添加 Provide 函数和 ProviderSet 条目：

```go
// 在文件中添加 Provide 函数
func ProvidePaymentService(
	orderRepo PaymentOrderRepository,
	planRepo PaymentPlanRepository,
	provider PaymentProvider,
	cache PaymentCache,
	userService *UserService,
	subscriptionService *SubscriptionService,
	billingCacheService *BillingCacheService,
	entClient *dbent.Client,
	cfg *config.Config,
) *PaymentService {
	return NewPaymentService(
		orderRepo, planRepo, provider, cache,
		userService, subscriptionService, billingCacheService, entClient,
		cfg.Payment.OrderExpirySec, cfg.Payment.MinTopupAmount, cfg.Payment.MaxTopupAmount,
	)
}

func ProvidePaymentExpiryService(orderRepo PaymentOrderRepository, cfg *config.Config) *PaymentExpiryService {
	interval := time.Duration(cfg.Payment.ExpiryTickSec) * time.Second
	if interval <= 0 {
		interval = time.Minute
	}
	svc := NewPaymentExpiryService(orderRepo, interval)
	svc.Start()
	return svc
}
```

在 `ProviderSet` 中添加：
```go
ProvidePaymentService,
ProvidePaymentExpiryService,
```

**2b. `backend/internal/repository/wire.go`** — 添加 Provide 函数和 ProviderSet 条目：

```go
func ProvideEasyPayProvider(cfg *config.Config) service.PaymentProvider {
	return NewEasyPayProvider(cfg.Payment)
}
```

在 `ProviderSet` 中添加：
```go
NewPaymentOrderRepository,
NewPaymentPlanRepository,
NewPaymentCache,
ProvideEasyPayProvider,
```

**2c. `backend/internal/handler/wire.go`** — 添加到 ProviderSet：

```go
// Top-level handlers 组中添加
NewPaymentHandler,
NewPaymentCallbackHandler,

// Admin handlers 组中添加
admin.NewPaymentPlanHandler,
admin.NewPaymentOrderHandler,
```

- [ ] **Step 3: 在 user.go 注册用户支付路由**

```go
func registerPaymentRoutes(authenticated *gin.RouterGroup, h *handler.Handlers) {
	payment := authenticated.Group("/payment")
	{
		payment.GET("/plans", h.Payment.ListPlans)
		payment.POST("/orders", h.Payment.CreateOrder)
		payment.GET("/orders", h.Payment.ListOrders)
		payment.GET("/orders/:id/status", h.Payment.GetOrderStatus)
	}
}
```

在 `RegisterUserRoutes` 中调用 `registerPaymentRoutes(authenticated, h)`。

- [ ] **Step 4: 在 admin.go 注册管理端支付路由**

```go
func registerAdminPaymentRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	plans := admin.Group("/payment/plans")
	{
		plans.GET("", h.Admin.PaymentPlan.List)
		plans.POST("", h.Admin.PaymentPlan.Create)
		plans.PUT("/:id", h.Admin.PaymentPlan.Update)
		plans.DELETE("/:id", h.Admin.PaymentPlan.Delete)
	}
	orders := admin.Group("/payment/orders")
	{
		orders.GET("", h.Admin.PaymentOrder.List)
		orders.GET("/stats", h.Admin.PaymentOrder.Stats) // 必须在 :id 之前注册
		orders.GET("/:id", h.Admin.PaymentOrder.GetByID)
		orders.POST("/:id/complete", h.Admin.PaymentOrder.Complete)
		orders.POST("/:id/refund", h.Admin.PaymentOrder.Refund)
	}
}
```

注意：`/stats` 必须在 `/:id` 之前注册，否则 Gin 会把 `stats` 当作 `:id`。

- [ ] **Step 5: 在 routes/router.go 注册回调路由（v1 顶层，不走 JWT 鉴权）**

查找 `backend/internal/server/routes/` 中注册 v1 group 的文件（通常是 `router.go`），在 JWT 鉴权 group 之外添加：

```go
func registerPaymentCallbackRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	// 支付回调 - 不走 JWT 鉴权，签名验证在 handler 内部完成
	paymentCallback := v1.Group("/payment")
	{
		paymentCallback.POST("/callback/:provider", h.PaymentCallback.Handle)
	}
}
```

在路由注册入口函数中调用 `registerPaymentCallbackRoutes(v1, h)`，确保在 JWT middleware group 之外。

- [ ] **Step 6: 运行 Wire 代码生成**

Run: `cd backend && wire ./cmd/server/`
Expected: 更新 `cmd/server/wire_gen.go`

如果 Wire 报错，检查：
1. 所有 Provide 函数的参数类型是否都有 provider
2. 所有 interface 类型（PaymentProvider, PaymentCache, PaymentOrderRepository, PaymentPlanRepository）是否都有返回对应类型的 provider
3. 不要直接注册 `NewPaymentService` 或 `NewEasyPayProvider`，要用 `ProvidePaymentService` 和 `ProvideEasyPayProvider` 包装函数

- [ ] **Step 7: 验证编译**

Run: `cd backend && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 8: Commit**

```bash
git add backend/internal/handler/handler.go backend/internal/service/wire.go backend/internal/server/routes/user.go backend/internal/server/routes/admin.go
git commit -m "feat(payment): wire up payment module routes and DI"
```

---

## Task 10: 数据库迁移

**Files:**
- 取决于项目的迁移方式（Ent auto-migration 或手动 SQL）

- [ ] **Step 1: 确认项目的数据库迁移策略**

检查项目是否使用 `ent.Schema.Create` 自动迁移，还是使用 atlas/goose 等手动迁移工具。

- [ ] **Step 2: 执行迁移**

如果是自动迁移：启动服务时 Ent 会自动创建表。
如果是手动迁移：生成 SQL 并创建迁移文件。

Run: 启动后端服务，确认 `payment_plans` 和 `payment_orders` 表已创建。

- [ ] **Step 3: 验证表结构**

连接数据库确认：
```sql
\d payment_plans
\d payment_orders
```

- [ ] **Step 4: Commit（如有迁移文件）**

---

## Task 11: 端到端冒烟测试

- [ ] **Step 1: 启动后端服务**

Run: `cd backend && go run ./cmd/server`
Expected: 服务启动成功，无 panic

- [ ] **Step 2: 测试管理端创建套餐**

```bash
curl -X POST http://localhost:8080/api/v1/admin/payment/plans \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"基础版月度","description":"日额度 $20","group_id":1,"duration_days":30,"price":9.98}'
```

Expected: 201 Created

- [ ] **Step 3: 测试用户端获取套餐列表**

```bash
curl http://localhost:8080/api/v1/payment/plans \
  -H "Authorization: Bearer <user_token>"
```

Expected: 返回套餐列表 JSON

- [ ] **Step 4: 测试创建订单**

```bash
curl -X POST http://localhost:8080/api/v1/payment/orders \
  -H "Authorization: Bearer <user_token>" \
  -H "Content-Type: application/json" \
  -d '{"type":"plan","plan_id":1,"provider":"alipay"}'
```

Expected: 返回 order_no（由于 EasyPay 未实现，会返回 provider error，这是预期的）

- [ ] **Step 5: Commit 冒烟测试通过**

```bash
git commit --allow-empty -m "test(payment): smoke test passed - payment module endpoints functional"
```

---

## 实现顺序依赖

```
Task 1 (常量+配置)
  ↓
Task 2 (Plan schema) → Task 3 (Order schema)
  ↓
Task 4 (Service 层)
  ↓
Task 5 (Expiry worker)
  ↓
Task 6 (Repository 层)
  ↓
Task 7 (EasyPay 桩)
  ↓
Task 8 (Handler + DTO)
  ↓
Task 9 (Wire + 路由)
  ↓
Task 10 (数据库迁移)
  ↓
Task 11 (冒烟测试)
```
