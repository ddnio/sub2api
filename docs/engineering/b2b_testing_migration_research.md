# B2B 组织化改造：Repository 集成测试、并发一致性测试与数据迁移策略研究

> 目标项目栈：Go + Ent + PostgreSQL  
> 适用场景：Sub2API 从 C2C 升级到 B2B，需将 User 级资源（balance/concurrency/api_keys）迁移到 Team 级，并建立 Personal Organization / Personal Team 模型。

---

## 目录

1. [Testcontainers-go 使用模式](#1-testcontainers-go-使用模式)
2. [Fixture Factory 设计](#2-fixture-factory-设计)
3. [Repository 层集成测试模板](#3-repository-层集成测试模板)
4. [并发扣费一致性测试](#4-并发扣费一致性测试)
5. [SUM(balance) 守恒验证](#5-sumbalance-守恒验证)
6. [数据迁移测试（Migration Test）](#6-数据迁移测试migration-test)
7. [C2C → B2B 低停机迁移策略](#7-c2c--b2b-低停机迁移策略)
8. [Personal Organization / Personal Team 迁移脚本示例](#8-personal-organization--personal-team-迁移脚本示例)
9. [落地 Checklist](#9-落地-checklist)

---

## 1. Testcontainers-go 使用模式

### 1.1 核心依赖

```go
import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)
```

### 1.2 Suite 基类设计

推荐每个需要真实 PG 的 Repository 测试继承 `DBTestSuite`。生命周期：

- `SetupSuite`：启动 container、连接数据库、执行 schema 迁移。
- `SetupTest`：按需求清空表或依赖 factory 的隔离数据。
- `TearDownSuite`：终止 container、关闭连接。

```go
type DBTestSuite struct {
    suite.Suite
    container testcontainers.Container
    client    *ent.Client
    db        *sql.DB
    ctx       context.Context
}

func (s *DBTestSuite) SetupSuite() {
    s.ctx = context.Background()
    // 本地禁用 Ryuk 可显著加速（仅限本地开发）
    if os.Getenv("CI") == "" {
        _ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
    }

    ctx, cancel := context.WithTimeout(s.ctx, 120*time.Second)
    defer cancel()

    pgContainer, err := postgres.Run(ctx,
        "postgres:15-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    s.Require().NoError(err)
    s.container = pgContainer

    host, _ := pgContainer.Host(ctx)
    port, _ := pgContainer.MappedPort(ctx, "5432")
    dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

    db, err := sql.Open("postgres", dsn)
    s.Require().NoError(err)
    s.db = db

    drv := entsql.OpenDB(dialect.Postgres, db)
    s.client = ent.NewClient(ent.Driver(drv))

    // 自动迁移
    s.migrateSchema()
}

func (s *DBTestSuite) migrateSchema() {
    ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
    defer cancel()
    err := s.client.Schema.Create(ctx,
        schema.WithDropIndex(true),
        schema.WithDropColumn(true),
    )
    s.Require().NoError(err)
}

func (s *DBTestSuite) TearDownSuite() {
    if s.client != nil { _ = s.client.Close() }
    if s.db != nil     { _ = s.db.Close() }
    if s.container != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        _ = s.container.Terminate(ctx)
    }
}
```

### 1.3 快速创建函数（非 Suite 场景）

对于仅需一次性的并发测试，提供 `NewTestDB(t)`  helper：

```go
func NewTestDB(t *testing.T) (context.Context, *ent.Client, func()) {
    ctx := context.Background()
    pgContainer, err := postgres.Run(ctx, "postgres:15-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    require.NoError(t, err)

    host, _ := pgContainer.Host(ctx)
    port, _ := pgContainer.MappedPort(ctx, "5432")
    dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)

    drv := entsql.OpenDB(dialect.Postgres, db)
    client := ent.NewClient(ent.Driver(drv))
    require.NoError(t, client.Schema.Create(ctx))

    cleanup := func() {
        _ = client.Close()
        _ = db.Close()
        _ = pgContainer.Terminate(ctx)
    }
    return ctx, client, cleanup
}
```

### 1.4 性能优化建议

| 优化点 | 做法 | 效果 |
|--------|------|------|
| 禁用 Ryuk | `TESTCONTAINERS_RYUK_DISABLED=true` | 本地提速 30-50% |
| 复用 Container | 一个 package 内共享一个 Suite | 避免每测启动 PG |
| 使用 `TRUNCATE` 而非删 container | `SetupTest` 里快速清表 | 秒级重置 |
| 迁移缓存 | Atlas `hash` 校验跳过无变化迁移 | CI 更稳定 |

---

## 2. Fixture Factory 设计

### 2.1 设计原则

1. **Option 模式**：每个实体提供 `XxxOption`，灵活覆盖默认值。
2. **一键创建复杂拓扑**：`NewPersonalTeam(ctx)` 同时创建 User + Org + Team + TeamMember。
3. **唯一性保证**：内部使用 `atomic.Uint64` 生成序列号，避免并发冲突。
4. **失败即 Panic**：测试环境里 `Save` 失败直接 panic，快速定位 fixture 问题。

### 2.2 Factory 结构

```go
type Factory struct {
    client *ent.Client
    seq    uint64
}

func NewFactory(client *ent.Client) *Factory {
    return &Factory{client: client}
}

func (f *Factory) nextSeq() uint64 {
    return atomic.AddUint64(&f.seq, 1)
}
```

### 2.3 关键 Fixture 方法

```go
// User
type UserOption func(*ent.UserCreate)
func (f *Factory) NewUser(ctx context.Context, opts ...UserOption) *ent.User

// Organization
type OrgOption func(*ent.OrganizationCreate)
func (f *Factory) NewOrganization(ctx context.Context, opts ...OrgOption) *ent.Organization

// Team
type TeamOption func(*ent.TeamCreate)
func (f *Factory) NewTeam(ctx context.Context, opts ...TeamOption) *ent.Team

// Personal Team 一键创建
func (f *Factory) NewPersonalTeam(ctx context.Context, userOpts ...UserOption) (*ent.User, *ent.Organization, *ent.Team)

// Balance
type BalanceOption func(*ent.TeamBalanceCreate)
func (f *Factory) NewTeamBalance(ctx context.Context, opts ...BalanceOption) *ent.TeamBalance

// API Key
type APIKeyOption func(*ent.APIKeyCreate)
func (f *Factory) NewAPIKey(ctx context.Context, opts ...APIKeyOption) *ent.APIKey
```

### 2.4 辅助查询方法（用于断言）

```go
func (f *Factory) TotalTeamBalance(ctx context.Context) decimal.Decimal
func (f *Factory) GetTeamBalance(ctx context.Context, teamID uuid.UUID) decimal.Decimal
func (f *Factory) CountTeamMembers(ctx context.Context, teamID uuid.UUID) int
```

---

## 3. Repository 层集成测试模板

### 3.1 被测 Repository 示例

```go
type BalanceRepo interface {
    Deduct(ctx context.Context, teamID uuid.UUID, amount decimal.Decimal) error
    GetBalance(ctx context.Context, teamID uuid.UUID) (decimal.Decimal, error)
    ListByUserID(ctx context.Context, userID int) ([]*ent.TeamBalance, error)
}

type balanceRepo struct{ client *ent.Client }

func NewBalanceRepo(client *ent.Client) BalanceRepo {
    return &balanceRepo{client: client}
}

var ErrInsufficientBalance = fmt.Errorf("insufficient balance")

func (r *balanceRepo) Deduct(ctx context.Context, teamID uuid.UUID, amount decimal.Decimal) error {
    if amount.LessThanOrEqual(decimal.Zero) {
        return fmt.Errorf("amount must be positive")
    }
    affected, err := r.client.TeamBalance.Update().
        Where(teambalance.TeamID(teamID), teambalance.AmountGTE(amount)).
        AddAmount(amount.Neg()).
        Save(ctx)
    if err != nil {
        return fmt.Errorf("deduct failed: %w", err)
    }
    if affected == 0 {
        return ErrInsufficientBalance
    }
    return nil
}
```

### 3.2 Suite 集成测试

```go
type BalanceRepoTestSuite struct {
    DBTestSuite
    repo    BalanceRepo
    factory *Factory
}

func TestBalanceRepoTestSuite(t *testing.T) {
    suite.Run(t, new(BalanceRepoTestSuite))
}

func (s *BalanceRepoTestSuite) SetupSuite() {
    s.DBTestSuite.SetupSuite()
    s.repo = NewBalanceRepo(s.client)
    s.factory = NewFactory(s.client)
}

func (s *BalanceRepoTestSuite) SetupTest() {
    ctx := s.ctx
    _, _ = s.client.TeamBalance.Delete().Exec(ctx)
    _, _ = s.client.TeamMember.Delete().Exec(ctx)
    _, _ = s.client.Team.Delete().Exec(ctx)
    _, _ = s.client.Organization.Delete().Exec(ctx)
    _, _ = s.client.User.Delete().Exec(ctx)
}

func (s *BalanceRepoTestSuite) TestDeduct_Success() {
    ctx := s.ctx
    _, _, te := s.factory.NewPersonalTeam(ctx)
    s.factory.NewTeamBalance(ctx,
        WithTeamBalanceTeamID(te.ID),
        WithTeamBalanceAmount(decimal.RequireFromString("100.00")),
    )
    err := s.repo.Deduct(ctx, te.ID, decimal.RequireFromString("30.50"))
    s.Require().NoError(err)
    balance, err := s.repo.GetBalance(ctx, te.ID)
    s.Require().NoError(err)
    s.True(balance.Equal(decimal.RequireFromString("69.50")))
}

func (s *BalanceRepoTestSuite) TestDeduct_InsufficientBalance() {
    ctx := s.ctx
    _, _, te := s.factory.NewPersonalTeam(ctx)
    s.factory.NewTeamBalance(ctx,
        WithTeamBalanceTeamID(te.ID),
        WithTeamBalanceAmount(decimal.RequireFromString("10.00")),
    )
    err := s.repo.Deduct(ctx, te.ID, decimal.RequireFromString("20.00"))
    s.Require().ErrorIs(err, ErrInsufficientBalance)
}

func (s *BalanceRepoTestSuite) TestListByUserID_Isolation() {
    ctx := s.ctx
    uA, _, teA := s.factory.NewPersonalTeam(ctx)
    s.factory.NewTeamBalance(ctx, WithTeamBalanceTeamID(teA.ID), WithTeamBalanceAmount(decimal.RequireFromString("100.00")))
    _, _, teB := s.factory.NewPersonalTeam(ctx)
    s.factory.NewTeamBalance(ctx, WithTeamBalanceTeamID(teB.ID), WithTeamBalanceAmount(decimal.RequireFromString("200.00")))

    listA, err := s.repo.ListByUserID(ctx, uA.ID)
    s.Require().NoError(err)
    s.Len(listA, 1)
    s.True(listA[0].Amount.Equal(decimal.RequireFromString("100.00")))
}
```

---

## 4. 并发扣费一致性测试

### 4.1 测试目标

- **原子性**：同一 team 的并发扣费不能出现超扣（overdraft）。
- **最终一致性**：所有成功扣费后的余额 = 初始余额 - 成功扣费总和。
- **死锁免疫**：并发事务不应因死锁导致大量失败。

### 4.2 并发测试写法（推荐 `errgroup`）

```go
import (
    "golang.org/x/sync/errgroup"
    "github.com/shopspring/decimal"
)

func TestBalanceRepo_Deduct_Concurrent(t *testing.T) {
    ctx, client, cleanup := NewTestDB(t)
    defer cleanup()

    factory := NewFactory(client)
    repo := NewBalanceRepo(client)

    _, _, te := factory.NewPersonalTeam(ctx)
    initial := decimal.RequireFromString("100.00")
    factory.NewTeamBalance(ctx,
        WithTeamBalanceTeamID(te.ID),
        WithTeamBalanceAmount(initial),
    )

    // 并发 10 个请求，每次扣 15.00
    workers := 10
    deductAmount := decimal.RequireFromString("15.00")
    var successCount int64

    g, gctx := errgroup.WithContext(ctx)
    for i := 0; i < workers; i++ {
        g.Go(func() error {
            err := repo.Deduct(gctx, te.ID, deductAmount)
            if err == nil {
                atomic.AddInt64(&successCount, 1)
                return nil
            }
            if errors.Is(err, ErrInsufficientBalance) {
                return nil // 预期内的失败，不计入成功
            }
            return err // 非预期失败
        })
    }
    require.NoError(t, g.Wait())

    // 断言
    expectedBalance := initial.Sub(deductAmount.Mul(decimal.NewFromInt(successCount)))
    actualBalance, err := repo.GetBalance(ctx, te.ID)
    require.NoError(t, err)
    require.True(t, expectedBalance.Equal(actualBalance),
        "expected %s, got %s, success=%d", expectedBalance, actualBalance, successCount)

    // 最多成功 6 次 (100 / 15 = 6.66)
    require.LessOrEqual(t, successCount, int64(6))
}
```

### 4.3 超扣防御的 Repository 写法

Ent 的 `Update().Where(...).Save()` 在 PG 层面会生成：

```sql
UPDATE team_balances
SET amount = amount - $1
WHERE team_id = $2 AND amount >= $1;
```

返回 `affected == 0` 即表示余额不足或行锁冲突。这是**乐观锁 + 条件更新**的最佳实践，天然避免超扣。

> 若扣费逻辑更复杂（涉及多表），建议在 Service 层使用 `SELECT FOR UPDATE` 或在 Repository 层使用 `client.Tx()` 事务包裹。

---

## 5. SUM(balance) 守恒验证

### 5.1 为什么需要守恒验证？

在 B2B 迁移过程中，原有的 `users.balance` 要迁移到 `team_balances.amount`。必须保证：

```
SUM(users.balance) before migration == SUM(team_balances.amount) after migration
```

### 5.2 测试代码模板

```go
func TestMigration_BalanceConservation(t *testing.T) {
    ctx, client, cleanup := NewTestDB(t)
    defer cleanup()
    factory := NewFactory(client)

    // 1. 创建旧世界数据：3 个用户，各自有 balance
    u1 := factory.NewUser(ctx, WithBalance(decimal.RequireFromString("100.00")))
    u2 := factory.NewUser(ctx, WithBalance(decimal.RequireFromString("250.50")))
    u3 := factory.NewUser(ctx, WithBalance(decimal.RequireFromString("0.00")))

    // 模拟旧表余额（假设 users 表还有 balance 字段）
    _, err := client.User.UpdateOneID(u1.ID).SetBalance(decimal.RequireFromString("100.00")).Save(ctx)
    require.NoError(t, err)
    _, err = client.User.UpdateOneID(u2.ID).SetBalance(decimal.RequireFromString("250.50")).Save(ctx)
    require.NoError(t, err)
    _, err = client.User.UpdateOneID(u3.ID).SetBalance(decimal.Zero).Save(ctx)
    require.NoError(t, err)

    // 2. 计算迁移前总和
    var beforeSum decimal.Decimal
    rows, err := client.User.Query().Select(user.FieldBalance).All(ctx)
    require.NoError(t, err)
    for _, r := range rows {
        beforeSum = beforeSum.Add(r.Balance)
    }

    // 3. 执行迁移脚本
    err = RunMigrateUsersToPersonalTeams(ctx, client)
    require.NoError(t, err)

    // 4. 计算迁移后总和
    afterSum := factory.TotalTeamBalance(ctx)

    // 5. 断言守恒
    require.True(t, beforeSum.Equal(afterSum),
        "balance not conserved: before=%s after=%s", beforeSum, afterSum)
}
```

### 5.3 多币种守恒

若存在多币种，需按 `currency` 分组求和验证：

```go
type BalanceByCurrency struct {
    Currency string
    Sum      decimal.Decimal
}

func assertBalanceConservedByCurrency(t *testing.T, before, after []BalanceByCurrency) {
    beforeMap := make(map[string]decimal.Decimal)
    for _, b := range before { beforeMap[b.Currency] = b.Sum }
    afterMap := make(map[string]decimal.Decimal)
    for _, a := range after { afterMap[a.Currency] = a.Sum }
    require.Equal(t, beforeMap, afterMap)
}
```

---

## 6. 数据迁移测试（Migration Test）

### 6.1 测试策略

迁移测试不同于单元测试，它需要：

1. **真实数据库**：使用 testcontainers 启动旧 schema 的 PG。
2. **历史 schema 快照**：保留一个 `ent/migrate/snapshots/v1.sql` 作为 C2C 时代的 schema。
3. **种子数据**：插入旧 schema 的代表性数据（边缘 case：balance 为 0、负数、超大值、软删除用户）。
4. **执行迁移**：运行 Go 编写的迁移脚本或 Atlas 迁移。
5. **断言新 schema 状态**：检查外键、约束、默认值、守恒量。

### 6.2 迁移测试目录结构

```
test/migration/
├── snapshot_v1.sql          # C2C 旧 schema
├── migration_test.go        # Go 测试文件
├── seed.go                  # 旧数据种子生成器
└── assertions.go            # 迁移后断言辅助
```

### 6.3 迁移测试代码骨架

```go
func TestC2CToB2BMigration(t *testing.T) {
    ctx := context.Background()

    // 1. 启动 container
    pgContainer, err := postgres.Run(ctx, "postgres:15-alpine",
        postgres.WithDatabase("uc"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    require.NoError(t, err)
    defer func() { _ = pgContainer.Terminate(ctx) }()

    host, _ := pgContainer.Host(ctx)
    port, _ := pgContainer.MappedPort(ctx, "5432")
    dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/uc?sslmode=disable", host, port.Port())
    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)
    defer db.Close()

    // 2. 初始化旧 schema
    oldSchema, err := os.ReadFile("snapshot_v1.sql")
    require.NoError(t, err)
    _, err = db.Exec(string(oldSchema))
    require.NoError(t, err)

    // 3. 插入种子数据
    seedOldData(db)

    // 4. 执行迁移（Go 脚本）
    drv := entsql.OpenDB(dialect.Postgres, db)
    client := ent.NewClient(ent.Driver(drv))
    defer client.Close()

    // 运行 Atlas 迁移到最新 schema
    err = client.Schema.Create(ctx,
        schema.WithDropIndex(true),
        schema.WithDropColumn(true),
    )
    require.NoError(t, err)

    err = migration.MigrateC2CToB2B(ctx, client)
    require.NoError(t, err)

    // 5. 断言
    assertAllUsersHavePersonalTeam(t, ctx, client)
    assertBalanceConserved(t, ctx, client, db)
    assertApiKeysIsolatedByTeam(t, ctx, client)
}
```

### 6.4 边缘 Case 清单

| Case | 说明 |
|------|------|
| 余额为 0 | 必须创建 team_balance 0 记录或允许缺失 |
| 已删除/禁用用户 | 根据策略决定是否迁移 |
| 大量 API Keys | 验证 key 与 team 关联正确 |
| 并发额度 (concurrency) | 迁移后应落到 team 级配置 |
| 特殊字符用户名/组织名 | 验证 PG 插入无转义问题 |

---

## 7. C2C → B2B 低停机迁移策略

### 7.1 可选策略对比

| 策略 | 停机时间 | 复杂度 | 风险 | 适用规模 |
|------|---------|--------|------|---------|
| **Maintenance Window** | 分钟级 | 低 | 低 | 小-中 |
| **Blue-Green** | 秒级 | 中 | 中 | 中-大 |
| **Online (双写+回溯)** | 零停机 | 高 | 高 | 大 |

Sub2API 用户量不大，接受 **10 分钟维护窗口**，因此推荐 **Maintenance Window + 预演** 方案，兼顾简单与安全。

### 7.2 Maintenance Window 方案（推荐）

```
阶段 1: 代码双读兼容（提前 1-2 周上线）
    - 新代码同时支持 user_id 和 team_id 查询
    - 写操作仍落在旧表

阶段 2: 维护窗口前准备
    - 备份生产数据库
    - 在 staging 完整跑通迁移脚本+验证
    - 准备回滚脚本

阶段 3: 维护窗口执行（10 分钟内）
    1. 切流量到维护页 (00:00)
    2. 执行 schema 迁移 (01:00)
    3. 执行数据迁移脚本 (02:00)
    4. 运行守恒验证脚本 (06:00)
    5. 启动新版本应用 (07:00)
    6. 冒烟测试 (09:00)
    7. 开放流量 (10:00)
```

### 7.3 Blue-Green 方案（备选）

若未来用户量增长或无法接受任何停机：

1. **Green 环境**：部署新版本 + 新 schema。
2. **数据同步**：使用 logical replication 或 CDC 将旧库实时同步到 Green。
3. **迁移脚本在 Green 跑**：不影响 Blue 生产。
4. **流量切换**：DNS/LoadBalancer 秒级切到 Green。
5. **回滚**：发现问题立即切回 Blue。

> 当前阶段不建议实施 Blue-Green，因为需要额外基础设施（CDC、双集群）。

### 7.4 回滚策略

- **schema 回滚**：维护 Atlas `down` 迁移脚本。
- **数据回滚**：迁移前做 `pg_dump` 逻辑备份。
- **应用回滚**：保留旧版本 Docker 镜像，5 分钟内可重新部署。

---

## 8. Personal Organization / Personal Team 迁移脚本示例

### 8.1 迁移逻辑说明

每个已有用户需生成：

1. `Organization` (type = personal, owner_user_id = user.id)
2. `Team` (type = personal, organization_id = org.id)
3. `TeamMember` (user_id = user.id, team_id = team.id, role = owner)
4. `TeamBalance` (team_id = team.id, amount = user.balance, currency = 'CNY')
5. `APIKey` 更新 team_id = team.id（若之前有 user_id）

### 8.2 Go 迁移脚本

```go
package migration

import (
    "context"
    "fmt"
    ""github.com/google/uuid"
    "uc/ent"
    "uc/ent/organization"
    "uc/ent/team"
    "uc/ent/user"
)

// MigrateUsersToPersonalTeams 将现有用户迁移到 Personal Org/Team
func MigrateUsersToPersonalTeams(ctx context.Context, client *ent.Client) error {
    users, err := client.User.Query().
        Where(user.Or(
            user.DeletedAtIsNil(),
            // 若逻辑删除也迁移，去掉这行条件
        )).
        All(ctx)
    if err != nil {
        return fmt.Errorf("query users failed: %w", err)
    }

    for _, u := range users {
        if err := migrateSingleUser(ctx, client, u); err != nil {
            return fmt.Errorf("migrate user %d failed: %w", u.ID, err)
        }
    }
    return nil
}

func migrateSingleUser(ctx context.Context, client *ent.Client, u *ent.User) error {
    // 幂等：若该用户已有 personal org，跳过
    exists, err := client.Organization.Query().
        Where(organization.OwnerUserID(u.ID), organization.Type(organization.TypePersonal)).
        Exist(ctx)
    if err != nil {
        return err
    }
    if exists {
        return nil
    }

    orgID := uuid.New()
    teamID := uuid.New()

    // 1. 创建 Organization
    org, err := client.Organization.Create().
        SetID(orgID).
        SetName(u.Nickname + "'s Org").
        SetType(organization.TypePersonal).
        SetOwnerUserID(u.ID).
        Save(ctx)
    if err != nil {
        return fmt.Errorf("create org failed: %w", err)
    }

    // 2. 创建 Team
    te, err := client.Team.Create().
        SetID(teamID).
        SetName(u.Nickname + "'s Team").
        SetType(team.TypePersonal).
        SetOrganizationID(org.ID).
        Save(ctx)
    if err != nil {
        return fmt.Errorf("create team failed: %w", err)
    }

    // 3. 创建 TeamMember
    _, err = client.TeamMember.Create().
        SetTeamID(te.ID).
        SetUserID(u.ID).
        SetRole("owner").
        Save(ctx)
    if err != nil {
        return fmt.Errorf("create team member failed: %w", err)
    }

    // 4. 迁移 Balance
    if u.Balance.GreaterThan(decimal.Zero) || true { // true 表示即使 0 也创建记录
        _, err = client.TeamBalance.Create().
            SetTeamID(te.ID).
            SetAmount(u.Balance).
            SetCurrency("CNY").
            Save(ctx)
        if err != nil {
            return fmt.Errorf("create team balance failed: %w", err)
        }
    }

    // 5. 迁移 API Keys（假设旧表 api_keys 有 user_id 字段）
    _, err = client.APIKey.Update().
        Where(apikey.UserID(u.ID)).
        SetTeamID(te.ID).
        Save(ctx)
    if err != nil {
        return fmt.Errorf("update api keys failed: %w", err)
    }

    return nil
}
```

### 8.3 SQL 一次性迁移脚本（备用）

如果数据量较小（<10万用户），也可以用单个 SQL 事务执行，速度更快：

```sql
BEGIN;

-- 1. 插入 personal organizations
INSERT INTO organizations (id, name, type, owner_user_id, created_at)
SELECT gen_random_uuid(), u.nickname || '''s Org', 'personal', u.id, NOW()
FROM users u
LEFT JOIN organizations o ON o.owner_user_id = u.id AND o.type = 'personal'
WHERE o.id IS NULL;

-- 2. 插入 personal teams
INSERT INTO teams (id, name, type, organization_id, created_at)
SELECT gen_random_uuid(), u.nickname || '''s Team', 'personal', o.id, NOW()
FROM users u
JOIN organizations o ON o.owner_user_id = u.id AND o.type = 'personal'
LEFT JOIN teams t ON t.organization_id = o.id AND t.type = 'personal'
WHERE t.id IS NULL;

-- 3. 插入 team_members
INSERT INTO team_members (team_id, user_id, role, created_at)
SELECT t.id, u.id, 'owner', NOW()
FROM users u
JOIN organizations o ON o.owner_user_id = u.id AND o.type = 'personal'
JOIN teams t ON t.organization_id = o.id AND t.type = 'personal'
LEFT JOIN team_members tm ON tm.team_id = t.id AND tm.user_id = u.id
WHERE tm.team_id IS NULL;

-- 4. 迁移 balances
INSERT INTO team_balances (team_id, amount, currency, updated_at)
SELECT t.id, u.balance, 'CNY', NOW()
FROM users u
JOIN organizations o ON o.owner_user_id = u.id AND o.type = 'personal'
JOIN teams t ON t.organization_id = o.id AND t.type = 'personal'
LEFT JOIN team_balances tb ON tb.team_id = t.id
WHERE tb.team_id IS NULL;

-- 5. 迁移 api_keys
UPDATE api_keys k
SET team_id = t.id
FROM users u
JOIN organizations o ON o.owner_user_id = u.id AND o.type = 'personal'
JOIN teams t ON t.organization_id = o.id AND t.type = 'personal'
WHERE k.user_id = u.id AND k.team_id IS NULL;

COMMIT;
```

### 8.4 迁移后验证 SQL

```sql
-- 每个用户都有且只有一个 personal org
SELECT u.id
FROM users u
LEFT JOIN organizations o ON o.owner_user_id = u.id AND o.type = 'personal'
GROUP BY u.id
HAVING COUNT(o.id) != 1;

-- balance 守恒
SELECT (
    (SELECT COALESCE(SUM(balance), 0) FROM users)
    -
    (SELECT COALESCE(SUM(amount), 0) FROM team_balances)
) AS diff;

-- 没有孤立的 api_keys
SELECT COUNT(*) FROM api_keys WHERE team_id IS NULL;
```

---

## 9. 落地 Checklist

### 测试基础设施

- [ ] 引入 `testcontainers-go/modules/postgres`
- [ ] 创建 `test/suite/db_test_suite.go` 作为所有 Repository 测试基类
- [ ] 创建 `test/fixture/factory.go` 实现 User/Org/Team/Balance/APIKey 的 fixture
- [ ] CI 中确保 Docker 可用（testcontainers 需要 Docker daemon）

### Repository 集成测试

- [ ] 为每个 B2B 新增 Repository 编写 Suite 测试（CRUD + 边界条件）
- [ ] 组织隔离查询测试（用户 A 不能查到用户 B 的数据）

### 并发一致性测试

- [ ] 余额扣费并发测试（10+ goroutine 同时扣费）
- [ ] SUM(balance) 守恒断言嵌入每次迁移相关测试
- [ ] 额度并发使用测试（若存在 concurrency limit）

### 迁移测试

- [ ] 保存 C2C schema 快照 SQL 到 `test/migration/snapshot_v1.sql`
- [ ] 编写 `test/migration/c2c_to_b2b_test.go`
- [ ] 覆盖边缘 case：余额为 0、已删除用户、大量 keys

### 生产迁移

- [ ] 维护窗口计划文档化（精确到分钟）
- [ ] 准备 `pg_dump` 备份命令
- [ ] 准备回滚脚本（应用 + schema + 数据）
- [ ] Staging 环境完整预演至少 2 次
- [ ] 迁移后执行验证 SQL，diff 必须为 0

---

## 参考与延伸阅读

1. **testcontainers-go**  
   https://golang.testcontainers.org/

2. **Ent 集成测试最佳实践**  
   https://entgo.io/docs/testing

3. **Atlas 迁移**  
   https://atlasgo.io/guides/migration-golang

4. **Stripe 组织化迁移经验（C2C → B2B）**  
   关键词："Stripe account migration to organization"，核心思路：
   - 提前双读兼容
   - maintenance window 执行一次性数据迁移
   - 严格守恒校验 + 灰度开放

5. **OpenAI / Anthropic 组织模型**  
   - Personal Account 自动属于 Default Organization
   - 所有 billing/usage/key 挂在 Organization / Project 下
   - 现有用户升级时透明迁移
