# RFC: Sub2API B2B 组织化改造 (Org-Team-User 3-Tier)

**状态**: Draft  
**日期**: 2025-04-17  
**目标**: 为 Sub2API 引入 B2B (Organization/Team/User) 支持，在保留 C2C 能力的前提下，建立可服务企业客户的平台架构。

---

## 1. 背景与问题

当前 Sub2API 的用户模型是扁平的：
- `users` 表直接承载 `balance`、`concurrency`、`role`
- `api_keys` 直接归属 `user_id`
- 计费路径：`api_key_auth` → `AuthSubject{UserID}` → `userRepo.UpdateBalance(userID, -cost)`

这导致无法支持以下 B2B 场景：
1. 企业客户需要多个员工共享同一组 API Keys 和预算
2. 企业管理员需要邀请/移除成员、查看团队总用量
3. 不同部门需要独立的预算隔离和成本归因
4. 平台侧需要区分 "个人用户" 与 "企业租户" 进行运营和定价

---

## 2. 设计目标

| 目标 | 说明 |
|------|------|
| **企业就绪** | 支持 Organization 创建、成员邀请、角色权限、Team 级资源隔离 |
| **C2C 兼容** | 现有用户无感知迁移，原有 API Keys 和余额不受影响 |
| **零额外查询 Gateway** | API Key 认证热路径不得增加新的 DB 查询 |
| **Schema 一步到位** | MVP 只暴露 Organization，但 Schema 预留多 Team 扩展能力 |
| **快速迁移** | 利用低用户量优势，通过 10 分钟维护窗口完成全量迁移 |

---

## 3. 上游对标研究

### 3.1 对比总览

| 平台 | 层级深度 | Key 归属 | 计费单位 | 角色模型 |
|------|----------|----------|----------|----------|
| **OpenAI Platform** | Organization > Project > API Key | **Project** | Project | Org: owner/reader; Project: owner/member |
| **Anthropic Admin API** | Organization > Workspace > Member | **Workspace** | Org (聚合), Workspace (追踪) | Org-level + Workspace-level 双轨 |
| **LiteLLM Proxy** | Organization > Team (optional) > Key | **Team / Key** | BudgetTable (复用于 Org/Team/Key/User) | admins[], members[], members_with_roles JSON |
| **OpenRouter** | Organization (扁平) | Organization | Organization | org:admin |
| **Stripe** | Account > Member (扁平) | Account | Account | Owner/Admin/Dev/Analyst |
| **Clerk** | Organization > Member (扁平) | Organization | Organization | 预定义 + 自定义 Role + Permission |

### 3.2 关键洞察

1. **OpenAI 和 Anthropic 都采用 "Org 治理 + 子单元执行" 模式**：
   - OpenAI 的 **Project** = 我们的 **Team**
   - Anthropic 的 **Workspace** = 我们的 **Team**
   - 两者都将 **API Key 的归属放到子单元**，而不是直接挂在 Org 下

2. **LiteLLM 的 BudgetTable 是亮点，但对 MVP 过度设计**：
   - 它将预算/额度抽象为独立表，可被 Org/Team/Key/User 复用
   - 优点是支持极其复杂的代理计费（per-end-user, per-tag）
   - 缺点是引入了一层抽象，对当前 "Team 直接有 balance" 的模型是负担

3. **Clerk 明确采用 Personal Organization 模式**：
   - 每个自然人用户默认拥有一个 Personal Org
   - 这是 C2C → B2B 无感知迁移的行业最佳实践

4. **OpenRouter 的 Organization 是扁平的**：
   - 没有 Project/Team 子层
   - 这意味着它只能服务"小公司统一预算"场景，无法满足部门隔离需求
   - 我们预留 Team 层比 OpenRouter 更具扩展性

---

## 4. 架构决策

### 4.1 核心模型：Org-Team-User 3-Tier

```
Organization (治理 & 充值聚合)
    │
    ├── Team (资源所有者：API Keys, Usage Logs, 实际计费扣除)
    │       │
    │       └── TeamMember (user_id + role)
    │
    └── OrganizationMember (user_id + org_role)
```

**Division of Labor**：
- **Organization**: 公司法律实体。负责成员邀请、充值、总账单查看、企业级设置
- **Team**: 部门/项目。负责 API Key 生命周期、实际用量计费、并发控制
- **User**: 自然人。通过 TeamMember / OrganizationMember 获得在两个层级上的角色

### 4.2 为什么 Team 必须是一级资源所有者

被拒绝的替代方案：
- **"Organization 直接拥有 Key"**：未来客户要求"研发部和运维部预算分开"时，必须再次重构
- **"User 拥有 Key，但 billing 打到 Org"**：造成所有权和计费权分离，离职员工的 Key 归属混乱，审计困难

**Team 作为 Key 所有者**与 OpenAI Project、Anthropic Workspace、LiteLLM Team 对齐，避免二次重构。

### 4.3 MVP 约束：单 Default Team

虽然 Schema 支持多 Team，但 MVP 阶段：
- 每个 Organization 创建时自动生成一个 `default` Team (`is_default = true`)
- 所有 API Keys 归属该 Team
- 前端路由不暴露 Team 概念，用户只感知 Organization

这为未来多 Team 扩展预留了空间，而不增加 MVP 前端复杂度。

### 4.4 现有用户迁移策略：Personal Organization + Personal Team

所有现有用户将自动获得：
- 一个 `Personal Organization`（名称为用户昵称或 email 前缀）
- 一个 `Personal Team`（`is_default = true`）
- `OrganizationMember` 记录（role = owner）
- `TeamMember` 记录
- `api_keys` 的 `team_id` 迁移到 Personal Team
- `users.balance` 迁移到 `teams.balance`

迁移后：
- 老用户从个人中心看到的余额 = Personal Team 的余额
- 老用户的 API Key = Personal Team 的 Key
- 对用户体验完全无感知

### 4.5 Gateway 热路径：零额外查询

当前 API Key 认证查询已经通过 JOIN 获取用户信息。改造后：
- 在同一查询中 LEFT JOIN `teams` 表
- 将 `team.id`, `team.balance`, `team.concurrency` 一并取出
- `AuthSubject` 新增 `TeamID` 和 `OrgID` 字段
- 但**不修改 `UserID` 的语义**：`UserID` 仍然代表当前操作者（自然人）
- 计费、并发控制、Rate Limit Key 等按 Team 隔离的逻辑，显式使用 `TeamID`，不将 `UserID` 替换为 `TeamID`

**不得**在认证后再发一次 `SELECT * FROM teams WHERE id = ?`。

改造前必须全量 `grep -r "AuthSubject" --include="*.go"`，评估所有下游使用点。

---

## 5. Schema 设计 (Ent)

### 5.1 新增实体

```go
// Organization: 顶层租户
func (Organization) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.String("name").NotEmpty(),
        field.String("slug").Unique().NotEmpty(),
        field.Bool("is_personal").Default(false), // true = Personal Org（C2C 迁移壳）
        field.Int64("balance").Default(0),        // 聚合余额（充值入口）
        field.Enum("status").Values("active", "suspended").Default("active"),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

// Team: 资源执行单元
func (Team) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.String("name").NotEmpty(),
        field.UUID("organization_id", uuid.UUID{}),
        field.Bool("is_default").Default(false),
        field.Int64("balance").Default(0),        // 实际扣费从这里扣
        field.Int64("budget_limit").Default(0),   // MVP 不启用，为未来 wallets 预留
        field.Int("concurrency").Default(0),
        field.Enum("status").Values("active", "suspended").Default("active"),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

// OrganizationMember: 用户在组织内的身份
func (OrganizationMember) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("organization_id", uuid.UUID{}),
        field.UUID("user_id", uuid.UUID{}),
        field.Enum("role").Values("owner", "admin", "member").Default("member"),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}
func (OrganizationMember) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("organization_id", "user_id").Unique(),
    }
}

// TeamMember: 用户在 Team 内的身份（MVP 阶段所有 OrgMember 自动加入 Default Team）
func (TeamMember) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("team_id", uuid.UUID{}),
        field.UUID("user_id", uuid.UUID{}),
        field.Enum("role").Values("owner", "admin", "member").Default("member"),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}
func (TeamMember) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("team_id", "user_id").Unique(),
    }
}
```

### 5.2 改造现有实体

```go
// APIKey: 增加 team_id，保留 user_id（审计/创建者追踪）
func (APIKey) Fields() []ent.Field {
    return append(existingFields,
        field.UUID("team_id", uuid.UUID{}).Optional(),
    )
}

// UsageLog: 增加 team_id 和 organization_id（ nullable for historical logs ）
func (UsageLog) Fields() []ent.Field {
    return append(existingFields,
        field.UUID("team_id", uuid.UUID{}).Optional(),
        field.UUID("organization_id", uuid.UUID{}).Optional(),
    )
}

// User: 保留 balance 和 concurrency 到迁移完成，之后逐步废弃
// （或迁移后立即设为冗余字段， Phase 2 删除）
```

---

## 6. 计费与权限模型

### 6.1 BillingAccount 接口

为了统一 C2C 和 B2B 的计费路径，引入一个抽象：

```go
type BillingAccount interface {
    ID() int64
    Balance() int64
    Concurrency() int
    Deduct(ctx context.Context, amount int64) error
}

type UserBillingAccount struct { user *ent.User }
type TeamBillingAccount struct { team *ent.Team }

func ResolveBillingAccount(apiKey *ent.APIKey) (BillingAccount, error) {
    if apiKey.TeamID != nil {
        return NewTeamBillingAccount(apiKey.Edges.Team)
    }
    return NewUserBillingAccount(apiKey.Edges.User), nil
}
```

**MVP 过渡期**：双轨并存（`team_id == nil` 走 User 路径，`team_id != nil` 走 Team 路径）。
**迁移完成后**：所有 `team_id != nil`，可删除 User 路径。

### 6.2 权限模型：MVP 阶段强制绑定 Org 与 Team 角色

采用**双轨角色表 + 绑定一致性**策略：

| 层级 | 角色 | 权限范围 |
|------|------|----------|
| **Organization** | owner | 删除 Org、管理充值、管理所有成员 |
| **Organization** | admin | 邀请成员、修改成员 Org 角色、查看总账单 |
| **Organization** | member | 仅查看自己所属信息 |
| **Team** | owner | 删除 Team、管理 Team 成员、创建/删除 API Keys |
| **Team** | admin | 创建/删除 API Keys、查看 Team 用量 |
| **Team** | member | 使用 Keys（只读） |

**MVP 约束**：
- 每个 Organization 只有一个 Default Team。
- **Org 角色与 Default Team 角色必须一致**：Org admin 对应 Team admin，Org member 对应 Team member。
- 在 Default Team 上的权限检查，**只查 `OrganizationMember`，不查 `TeamMember`**，避免双轨角色冲突。
- 所有权限检查统一走 `Authorize(userID, orgID, action)`，而不是分散在两张表上判断。

未来支持多 Team 后，再引入独立的 `TeamMember.role` 判断逻辑。

### 6.3 数据隔离策略

1. **应用层过滤**：所有 Repository 查询增加 `.Where(organization_id.EQ(...))` 或 `.Where(team_id.EQ(...))`
2. **JWT/Context 注入**：认证中间件将 `organization_id` 和 `team_id` 注入 `context`
3. **PostgreSQL RLS (Phase 2)**：作为安全兜底，防止 BI/分析师直连 DB 时绕过应用层权限

---

## 7. 迁移计划

### 7.1 执行时间线（Maintenance Window）

**窗口时间 = Staging 全量复制压测实测时间 × 2**。如果 staging 测出来 8 分钟，生产窗口预留 16 分钟。

| 时间 | 步骤 | 操作 |
|------|------|------|
| T-1h | 停服预告 | 发送维护通知，停止新注册（可选） |
| T+0 | 停服 | 部署 `OrganizationCreationEnabled=false`，拒绝新创建 Org |
| T+1m | Schema 升级 | 运行 Ent migration：创建新表，为 `api_keys`/`usage_logs` 添加列 |
| T+3m | 数据 Backfill | 执行 `MigrateUsersToPersonalTeams`：为每个 user 创建 Personal Org + Team，迁移 balance 和 keys |
| T+3m | **其中 `api_keys` 的 `team_id` 更新必须分批执行** | batch size 5000，避免大事务锁表 |
| T+6m | 校验 | 运行守恒检查：`SUM(users.balance) == SUM(teams.balance)`，`api_keys.team_id IS NULL` 计数为 0 |
| T+8m | 切流量 | 部署新版 Gateway 和 API 服务（已按 Team 计费） |
| T+9m | 开服 | `OrganizationCreationEnabled=true`，恢复服务 |
| T+10m | 监控 | 观察扣费、认证、用量日志是否正常 |

> **Plan B**：如果窗口超时，切换到只读模式（停止新的扣费请求，但允许查询），而不是完全停服。

### 7.2 回滚策略

如果校验失败或发现异常：
1. 立即将 `OrganizationCreationEnabled` 切回 `false`
2. 回滚到旧版本 Deployment（仍按 User 计费）
3. 由于 `users.balance` 在迁移前已被完整备份（或迁移脚本采用"复制"而非"清零"策略），可快速恢复

> **建议**：迁移脚本先执行 `users.balance` 的复制到 `teams.balance`，在验证通过后再执行 `users.balance = 0`（或在一个独立事务中完成）。

### 7.3 成员邀请：MVP 只支持邀请已有用户

为控制复杂度，MVP 阶段不支持"邀请注册"：
1. Admin 输入被邀请人 email
2. 系统检查该 email 是否已注册
3. 未注册 → 返回错误"该用户尚未注册，请先让对方注册账号"
4. 已注册 → 发送邀请链接，被邀请人点击后成为 Member

"邀请注册"推到 Phase 2。

---

## 8. 测试策略

详见 `b2b_testing_migration_research.md`。核心测试清单：

| 优先级 | 测试项 | 验证目标 |
|--------|--------|----------|
| **P0** | `TeamDeductBalance_Concurrent` | 并发扣费无超扣、无负余额 |
| **P0** | `BackfillPersonalTeams` | 迁移后 `SUM(balance)` 守恒，零孤儿 Key |
| **P0** | `APIKeyAuth_EmbedsTeamInfo` | Gateway 热路径无额外 DB 查询 |
| **P1** | `OrgAdminLifecycle` (HTTP) | 创建 Org → Invite → Key → Chat → Billing 端到端 |
| **P1** | `OrgIsolation_Breaches` | 跨组织数据访问被严格拦截 |
| **P2** | `FeatureFlag` 测试 | 热切换能正确启停 B2B 路由 |

---

## 9. 风险与待决策项

### 9.1 已识别风险

1. **并发扣费原子性**：`teamRepo.UpdateBalance` 必须保持与 `userRepo` 相同的 `WHERE balance >= cost` 条件更新模式，并且**必须在同一事务中写入 `usage_log`**
2. **迁移脚本幂等性**：必须保证多次执行不会重复创建 Personal Org/Team
3. **JWT AuthSubject 改造范围**：所有依赖 `AuthSubject.UserID` 的下游代码都需要 review，但不得修改 `UserID` 语义
4. **生产并发上限**：MVP 阶段必须给出单 Team 硬性 QPS 上限（如 100 QPS），超限返回 429

### 9.2 待决策项

1. **是否引入 `wallets` 表？** 当前方案：MVP 不引入，直接在 `teams` 存 `balance`。Phase 2 若需多币种或多钱包再抽象。
2. **是否引入 Redis 预扣？** 当前方案：MVP 不引入，继续使用 PostgreSQL 原子更新。Phase 2 若并发量激增再引入。
3. **API 路由设计**：
   - ~~Option A: `/api/v1/org/:org_id/...`~~ 已否决。虽然 MVP 简单，但未来加 Team 时 API 契约会变，债务更大。
   - **确定方案**：API 路由现在就采用 `/api/v1/org/:org_id/teams/:team_id/...`，前端在 MVP 阶段自动填充 `team_id = default`。页面路由仍保持简洁形式。

---

## 10. 结论

Sub2API 的 B2B 改造应采用 **Org-Team-User 3-Tier 架构**，Team 作为 API Key 和实际计费的第一级资源所有者。该方案与 OpenAI Project、Anthropic Workspace、LiteLLM Team 等上游标杆对齐，兼顾了 MVP 的交付速度和未来的扩展性。

迁移策略上，利用低用户量优势，采用 **Personal Organization + 10 分钟 Maintenance Window** 完成无感知迁移，是成本最低、风险可控的路径。

---

## 附录：上游参考链接

- OpenAI Platform Docs: https://platform.openai.com/docs/api-reference/organizations
- OpenAI Projects API: https://platform.openai.com/docs/api-reference/projects
- Anthropic Admin API: https://docs.anthropic.com/en/api/admin
- LiteLLM schema.prisma: https://github.com/BerriAI/litellm/blob/main/litellm/proxy/schema.prisma
- Clerk Organizations: https://clerk.com/docs/organizations/overview
- Stripe Accounts: https://stripe.com/docs/connect/accounts
