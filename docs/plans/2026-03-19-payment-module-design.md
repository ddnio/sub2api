# 支付模块设计文档

> 日期：2026-03-19
> 状态：草案
> 分支：feature/payment-module

## 概述

为 Sub2API 平台增加自助支付功能，支持用户在线购买订阅套餐和充值余额。一期支持国内支付宝和微信扫码支付，通过 PaymentProvider 抽象层为后续接入更多支付渠道留出扩展空间。

## 需求

### 功能需求

- 管理员在后台创建/编辑/上下架订阅套餐（Plan），每个套餐关联一个 Group
- 用户在 Dashboard 浏览套餐列表，选择后扫码支付
- 用户可自由输入金额充值余额
- 支付成功后自动发放权益（订阅 or 余额）
- 管理员可查看订单列表、筛选、统计收入、手动处理异常订单
- 用户可查看自己的订单历史

### 非功能需求

- 回调幂等：同一笔支付回调多次只发放一次权益
- 支付渠道可插拔：更换服务商只需实现 PaymentProvider 接口
- 与现有代码风格完全一致（service/repository/handler 分层、Wire DI、Ent ORM）
- 手动续费，无自动扣款

## 商业模型

### 混合模式

| 模式 | 说明 |
|------|------|
| 订阅套餐 | 按 Plan 购买，获得关联 Group 的订阅（日/周/月限额由 Group 定义） |
| 余额充值 | 自由输入金额，充入 `users.balance`，按实际 token 消耗扣费 |

### Plan 与 Group 的关系

- Plan = Group + 价格 + 时长 的商业包装
- 一个 Group 可对应多个 Plan（月度/季度/年度不同定价）
- Plan 不重复存储限额，限额变更只需改 Group
- 短期加油包也是 Plan（`duration_days=1`）

## 数据模型

### payment_plans（套餐定义表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, auto-increment | |
| name | string(100) | NOT NULL, UNIQUE | 套餐名称 |
| description | text | NOT NULL | 前端展示描述 |
| badge | string(20) | nullable | 角标："推荐"、"热卖" |
| group_id | bigint | FK → groups, ON DELETE RESTRICT | 关联 Group |
| duration_days | int | NOT NULL, > 0 | 订阅时长（天） |
| price | decimal(20,8) | NOT NULL, >= 0 | 价格（元） |
| original_price | decimal(20,8) | nullable, >= 0 | 划线价 |
| sort_order | int | NOT NULL, default 0, >= 0 | 排序权重 |
| is_active | bool | NOT NULL, default true | 上下架控制 |
| + TimeMixin | | | created_at, updated_at |
| + SoftDeleteMixin | | | deleted_at |

**索引：**
- `group_id`
- `(is_active, sort_order)`

**Mixin 顺序：** `TimeMixin{}` → `SoftDeleteMixin{}`

### payment_orders（订单表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, auto-increment | |
| order_no | string(32) | NOT NULL, UNIQUE | 时间戳前缀 + crypto/rand |
| user_id | bigint | FK → users, ON DELETE RESTRICT | 下单用户 |
| type | string(20) | NOT NULL | `plan` / `topup` |
| plan_id | bigint | nullable, FK → payment_plans, ON DELETE RESTRICT | 套餐订单关联 |
| amount | decimal(20,8) | NOT NULL, >= 0 | 实付金额（元） |
| credit_amount | decimal(20,8) | nullable, >= 0 | 实际发放金额 |
| status | string(20) | NOT NULL, default 'pending' | 订单状态 |
| provider | string(20) | nullable | `alipay` / `wechat` |
| provider_order_no | string(64) | nullable, UNIQUE | 第三方流水号 |
| paid_at | timestamptz | nullable | 支付时间 |
| completed_at | timestamptz | nullable | 权益发放时间 |
| refunded_at | timestamptz | nullable | 退款时间 |
| expired_at | timestamptz | NOT NULL | 订单过期时间 |
| callback_raw | text | nullable | 回调原始数据 |
| admin_note | text | nullable | 管理员备注 |
| created_at | timestamptz | NOT NULL, immutable | |
| updated_at | timestamptz | NOT NULL | |

**索引：**
- `user_id`
- `plan_id`
- `status`
- `(status, expired_at)` — 过期清理查询
- `provider_order_no` — 回调查找（UNIQUE 已隐含索引）

**不使用软删除** — 财务记录只做状态流转，不允许删除。

### 订单状态流转

```
pending → paid      回调验签通过，支付确认
paid → completed    权益发放成功
paid → failed       权益发放异常（需人工介入）
pending → expired   超时未支付（定时清理）
completed → refunded  管理员手动标记退款
```

**特殊情况：** 用户在订单过期后支付，回调时 `status='expired'`，仍然处理（钱已扣），流转为 `paid → completed`。

## 系统架构

### 组件依赖图

```
┌──────────────────────────────────────────────────┐
│  Handler 层                                       │
│  ├── PaymentHandler (用户)                        │
│  ├── PaymentCallbackHandler (回调，独立路由组)     │
│  ├── admin/PaymentPlanHandler                     │
│  └── admin/PaymentOrderHandler                    │
└──────────┬───────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────┐
│  Service 层                                       │
│  PaymentService                                   │
│  ├── PaymentOrderRepository (interface)           │
│  ├── PaymentPlanRepository (interface)            │
│  ├── PaymentProvider (interface)                  │
│  ├── PaymentCache (interface, Redis 锁)           │
│  ├── *SubscriptionService (已有，发放订阅)        │
│  ├── *UserService (已有，更新余额)                │
│  ├── *BillingCacheService (已有，缓存失效)        │
│  └── *dbent.Client (事务)                         │
└──────────┬───────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────┐
│  Repository 层                                    │
│  ├── payment_order_repo.go → PaymentOrderRepository │
│  ├── payment_plan_repo.go → PaymentPlanRepository   │
│  ├── payment_cache.go → PaymentCache (Redis)        │
│  └── easypay_provider.go → PaymentProvider          │
└──────────────────────────────────────────────────┘
```

### 文件结构

```
backend/internal/
├── service/
│   ├── payment_service.go            # 领域类型、接口定义、错误变量、PaymentService
│   └── payment_expiry_service.go     # 过期订单清理 worker（带 Stop() 生命周期）
├── repository/
│   ├── payment_order_repo.go         # 实现 service.PaymentOrderRepository
│   ├── payment_plan_repo.go          # 实现 service.PaymentPlanRepository
│   ├── payment_cache.go              # 实现 service.PaymentCache
│   └── easypay_provider.go           # 实现 service.PaymentProvider
├── handler/
│   ├── payment_handler.go            # 用户端
│   └── payment_callback_handler.go   # 回调端
├── handler/admin/
│   ├── payment_plan_handler.go       # 套餐管理
│   └── payment_order_handler.go      # 订单管理
├── handler/dto/
│   └── (mappers.go 中新增转换函数)
├── ent/schema/
│   ├── payment_plan.go
│   └── payment_order.go
├── config/
│   └── (config.go 中新增 PaymentConfig)
├── server/routes/
│   ├── (user.go 中新增 registerPaymentRoutes)
│   └── (admin.go 中新增 registerAdminPaymentRoutes)
└── (handler.go 新增字段，wire.go 注册)
```

### 配置

在 `config.go` 的 `Config` 结构体中新增：

```go
Payment PaymentConfig `mapstructure:"payment"`
```

```go
type PaymentConfig struct {
    Provider        string `mapstructure:"provider"`          // easypay, alipay, wechat
    EasyPayBaseURL  string `mapstructure:"easypay_base_url"`
    EasyPayAppID    string `mapstructure:"easypay_app_id"`
    EasyPaySignKey  string `mapstructure:"easypay_sign_key"`
    CallbackBaseURL string `mapstructure:"callback_base_url"` // 回调地址前缀
    OrderExpirySec  int    `mapstructure:"order_expiry_sec"`  // 订单过期时间，默认 900（15分钟）
}
```

## API 设计

### 用户端

#### GET /api/v1/payment/plans

获取可用套餐列表。

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "name": "进阶版月度",
      "description": "日额度 $50，适合高频使用",
      "badge": "推荐",
      "group_name": "Pro",
      "duration_days": 30,
      "price": 19.98,
      "original_price": 29.98,
      "daily_limit_usd": 50.0,
      "weekly_limit_usd": 200.0,
      "monthly_limit_usd": 500.0
    }
  ]
}
```

**说明：** `group_name` 和限额信息通过 JOIN Group 获取，前端无需额外请求。

#### POST /api/v1/payment/orders

创建订单。

**Request（套餐）：**
```json
{ "type": "plan", "plan_id": 1, "provider": "alipay" }
```

**Request（充值）：**
```json
{ "type": "topup", "amount": 50.00, "provider": "wechat" }
```

**Response:**
```json
{
  "data": {
    "order_no": "20260319143052abcdef123456",
    "qr_code_url": "https://qr.alipay.com/xxx",
    "amount": 19.98,
    "expired_at": "2026-03-19T14:45:52Z"
  }
}
```

**限流：** 每用户每小时最多创建 10 个订单。

#### GET /api/v1/payment/orders

我的订单列表（分页）。

**Query params:** `page`, `page_size`, `status`（可选筛选）

#### GET /api/v1/payment/orders/:id/status

轮询订单状态（前端支付弹窗使用）。

**Response:**
```json
{ "data": { "status": "completed" } }
```

### 回调端

#### POST /api/v1/payment/callback/:provider

支付平台异步通知，独立路由组，使用签名验证中间件，不走 JWT 鉴权。

**处理流程：**
1. `Provider.ParseCallback(r)` — 验签 + 解析订单号和金额
2. 校验回调金额 == 订单金额
3. `AcquireCallbackLock(orderNo)` — 防并发（Redis 不可用时降级跳过）
4. `UPDATE payment_orders SET status='paid' WHERE order_no=? AND status IN ('pending','expired')` — 乐观锁
5. 如果 affected rows = 0，说明已处理，返回成功（幂等）
6. 在 DB 事务内发放权益：
   - `topup` → `UserService.UpdateBalance(userID, creditAmount)`
   - `plan` → `SubscriptionService.AssignOrExtendSubscription(userID, groupID, days)`
7. 更新 status = 'completed'，写入 `paid_at`、`completed_at`、`provider_order_no`、`callback_raw`
8. 事务提交后异步失效缓存（5s 超时 goroutine）
9. `ReleaseCallbackLock(orderNo)`
10. 返回支付平台要求的成功响应

**异常：** 步骤 6 权益发放失败 → 订单停在 `paid`，管理员可通过 `/admin/payment/orders/:id/complete` 手动补发。

### 管理端

#### 套餐管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/payment/plans | 列表（含已下架） |
| POST | /api/v1/admin/payment/plans | 创建（幂等） |
| PUT | /api/v1/admin/payment/plans/:id | 编辑 |
| DELETE | /api/v1/admin/payment/plans/:id | 下架（软删除） |

所有写操作使用 `executeAdminIdempotentJSON` 包装。

#### 订单管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/payment/orders | 列表（分页、按状态/时间/用户/类型筛选） |
| GET | /api/v1/admin/payment/orders/:id | 详情 |
| POST | /api/v1/admin/payment/orders/:id/complete | 手动补发权益（paid → completed） |
| POST | /api/v1/admin/payment/orders/:id/refund | 标记退款（completed → refunded） |
| GET | /api/v1/admin/payment/orders/stats | 收入统计 |

#### 收入统计 GET /api/v1/admin/payment/orders/stats

**Query params:** `start_date`, `end_date`, `group_by`（day/month）

**Response:**
```json
{
  "data": {
    "total_amount": 12580.00,
    "total_orders": 342,
    "breakdown": [
      { "date": "2026-03-18", "amount": 1580.00, "count": 42 },
      { "date": "2026-03-19", "amount": 980.00, "count": 28 }
    ]
  }
}
```

## 接口定义

### PaymentProvider

```go
// 定义在 service/payment_service.go

type PaymentProvider interface {
    // CreatePayment 创建支付订单，返回二维码 URL
    CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResult, error)
    // ParseCallback 解析并验证支付回调
    ParseCallback(r *http.Request) (*CallbackResult, error)
}

type PaymentRequest struct {
    OrderNo  string          // 商户订单号
    Amount   decimal.Decimal // 支付金额（元）
    Provider string          // alipay / wechat
    Subject  string          // 商品描述
}

type PaymentResult struct {
    QRCodeURL string // 支付二维码 URL
}

type CallbackResult struct {
    OrderNo         string          // 商户订单号
    ProviderOrderNo string          // 第三方流水号
    Amount          decimal.Decimal // 实付金额
    Raw             string          // 原始数据
}
```

### PaymentCache

```go
// 定义在 service/payment_service.go

type PaymentCache interface {
    AcquireCallbackLock(ctx context.Context, orderNo string, ttl time.Duration) (bool, error)
    ReleaseCallbackLock(ctx context.Context, orderNo string) error
}
```

### PaymentOrderRepository / PaymentPlanRepository

```go
// 定义在 service/payment_service.go

type PaymentOrderRepository interface {
    Create(ctx context.Context, order *PaymentOrder) (*PaymentOrder, error)
    GetByOrderNo(ctx context.Context, orderNo string) (*PaymentOrder, error)
    GetByID(ctx context.Context, id int64) (*PaymentOrder, error)
    UpdateStatus(ctx context.Context, orderNo string, fromStatus, toStatus string, updates map[string]any) (int, error) // 返回 affected rows
    ListByUser(ctx context.Context, userID int64, filter OrderFilter, pagination Pagination) ([]*PaymentOrder, int, error)
    ListAll(ctx context.Context, filter OrderFilter, pagination Pagination) ([]*PaymentOrder, int, error)
    ExpirePendingOrders(ctx context.Context) (int, error) // 批量过期
    Stats(ctx context.Context, filter StatsFilter) (*OrderStats, error)
}

type PaymentPlanRepository interface {
    Create(ctx context.Context, plan *PaymentPlan) (*PaymentPlan, error)
    Update(ctx context.Context, id int64, updates map[string]any) (*PaymentPlan, error)
    GetByID(ctx context.Context, id int64) (*PaymentPlan, error)
    ListActive(ctx context.Context) ([]*PaymentPlan, error) // 用户端，只看激活的
    ListAll(ctx context.Context, pagination Pagination) ([]*PaymentPlan, int, error) // 管理端
    SoftDelete(ctx context.Context, id int64) error
}
```

## 故障处理

| 场景 | 处理方式 |
|------|---------|
| 回调到达但权益发放失败 | 订单停在 `paid`，管理员可手动 `/complete` 补发 |
| 回调重复到达 | `WHERE status IN ('pending','expired')` 乐观锁，已处理返回成功 |
| 用户在订单过期后支付 | 仍然处理（钱已扣），`expired` → `paid` → `completed` |
| 支付平台不可用 | `CreatePayment` 返回错误，前端提示"支付暂不可用" |
| Redis 不可用 | 跳过分布式锁，依赖 DB 乐观锁保证正确性 |
| 余额更新竞争 | 复用已有模式：`UPDATE users SET balance = balance + ? WHERE id = ?` |

## 安全

| 措施 | 说明 |
|------|------|
| 回调验签 | Provider.ParseCallback 内部验证签名，失败返回 400 |
| 回调金额校验 | 回调金额 vs 订单金额，不一致拒绝处理 |
| 订单创建限流 | 每用户每小时最多 10 个订单 |
| 回调路由隔离 | 独立路由组 + 签名验证中间件，不走 JWT |
| 订单号不可预测 | 时间戳前缀 + crypto/rand 随机后缀 |
| 管理操作幂等 | 所有 admin 写操作使用 `executeAdminIdempotentJSON` |

## 前端页面

### 用户端

1. **套餐选购页** — Dashboard → "套餐"
   - 卡片式布局展示套餐列表
   - 显示套餐名、描述、价格（划线价）、角标
   - 显示关联 Group 的限额信息
   - 点击 → 选择支付方式（支付宝/微信）→ 弹窗显示二维码
   - 前端轮询 `/orders/:id/status`，完成后自动关闭

2. **余额充值** — Dashboard → "充值"
   - 预设档位（¥10/¥50/¥100）+ 自定义金额输入
   - 同样的支付弹窗流程

3. **订单历史** — Dashboard → "订单"
   - 订单列表：时间、类型、金额、状态
   - 分页

### 管理端

4. **套餐管理** — Admin → "套餐管理"
   - 套餐列表（含已下架）
   - 创建/编辑/下架操作

5. **订单管理** — Admin → "订单管理"
   - 订单列表 + 筛选（状态、时间、用户、类型）
   - 异常处理（手动补发、标记退款）
   - 收入统计（按日/月聚合）

## 与现有代码的一致性要点

| 项 | 惯例 | 本模块遵循方式 |
|----|------|---------------|
| Service 构造 | `NewXxxService(deps...) *XxxService`，具体类型注入 | `NewPaymentService(orderRepo, planRepo, provider, cache, userSvc, subSvc, billingCache, entClient, authInvalidator)` |
| Repository 接口 | 定义在 service 包，repo 包提供未导出实现 | 同上 |
| 事务 | `s.entClient.Tx(ctx)` + defer Rollback | 回调处理流程使用同一模式 |
| 分布式锁 | service 定义 Cache interface，repo 实现，失败降级 | PaymentCache interface + 降级到 DB 乐观锁 |
| 缓存失效 | 事务后 goroutine + 5s timeout | 权益发放后异步失效 |
| Handler | 持有 `*service.XxxService`，DTO mapper 转换 | 同上 |
| Admin 幂等 | `executeAdminIdempotentJSON` | 所有写操作使用 |
| 路由注册 | `registerXxxRoutes` 私有函数 | `registerPaymentRoutes` + `registerAdminPaymentRoutes` |
| Wire | 注册到 ProviderSet，加入 Handlers/AdminHandlers | 同上 |
| Ent schema | TimeMixin → SoftDeleteMixin，string 存 enum | 同上 |
| 金额类型 | decimal(20,8) | 同上 |
| 生命周期 | background worker 暴露 Stop() | PaymentExpiryService.Stop() |

## 不在一期范围

- 自动续费
- Stripe 国际支付
- 优惠券/促销码与支付联动
- 邮件/站内通知
- 退款自动化（一期只做标记）
- 独立 Pricing Page（登录后在 Dashboard 内完成）
