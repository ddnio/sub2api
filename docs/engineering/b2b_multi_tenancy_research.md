# B2B SaaS 多租户模型最佳实践研究
## 目标：为 Sub2API (Go/Ent/PostgreSQL) 设计 org-team-user 三级架构

---

## 一、行业标杆方案深度分析

### 1. Stripe — Org + Member 模型
Stripe 的 Console/Account 体系是 B2B 多租户的经典标杆：
- **Account (Organization)**：所有资源（Customers、Subscriptions、Products、API Keys、Webhook Endpoints）都归属 Account。
- **Member**：一个 User 可以加入多个 Account，每个 Membership 有独立 Role。
- **Role 模型**：Owner / Admin / Developer / Analyst / Support（预定义），不支持自定义权限粒度，只支持预定义 Role 的分配。
- **数据隔离**：完全基于 Account ID。所有 API 请求携带 `Stripe-Account` header 或通过 OAuth 限定 Account 范围。DB 层所有表都有 `account_id` 字段（或内部等价物）。
- **Migration 模式**：Stripe 早期是单 Account 单 User，后来引入 Team/Organization 时，将每个已有 Account 的创建者自动设为 Owner，并生成一个 default Team（后演进为 Members）。

### 2. Clerk — Organization 组件模型
Clerk 作为身份认证基础设施，其 Organization 组件是专为 B2B SaaS 设计的：
- **层级深度**：Organization → (optional: Org Domain/Slug) → Member。没有内置 Team 层级，但支持通过 Metadata 或自定义扩展。
- **资源归属单位**：Organization 是资源的唯一归属单位。User 通过 OrganizationMembership 关联。
- **角色模型**：管理员 (admin) / 成员 (member) / 基本成员 (basic_member) / 访客 (guest_member)。支持自定义 Role 和 Permissions（基于 permission 字符串列表）。
- **数据隔离策略**：
  - Session 携带 `org_id` 和 `org_role`。
  - 通过 Active Organization 切换上下文。
  - API 层使用 `organization_id` 过滤。
- **常见 Migration 模式**：
  - 每个现有 User 自动创建一个 Personal Organization。
  - 后续邀请/迁移时，将用户数据从 Personal Org 迁移到目标 Org（或保持只读）。

### 3. Auth0 — RBAC 与 Organization 隔离
Auth0 的 Organizations 功能（相对较新）是为了弥补早期纯 RBAC 缺少 B2B 组织隔离的短板：
- **层级深度**：Organization → Member → (可选 Groups)。Groups 可视为 Team 的等价物。
- **资源归属单位**：Application/Connection 是全局的；Organization 分配 Member 和 Role。真正 SaaS 应用的资源隔离由应用层实现（Auth0 主要提供 Token 中的 `org_id`）。
- **角色模型**：Organization-specific Roles。每个 Role 是一组 Permissions 的集合。支持自定义 Role。
- **数据隔离策略**：
  - 访问令牌 (Access Token) 中注入 `org_id`。
  - 应用层用 `org_id` 做查询过滤。
  - 支持 Just-in-Time (JIT) Provisioning。
- **常见 Migration 模式**：
  - 从纯 App Metadata 中的 `company_id` 迁移到原生 Organization 对象。
  - Backfill：遍历所有 User，根据 `company_id` metadata 创建 Organization 和 Membership。

### 4. Supabase — Multi-tenancy 方案
Supabase 作为 PostgreSQL 托管平台，其多租户方案更偏向数据库层：
- **层级深度**：Project → (Schema/Database) → Table。应用层组织模型（如 Organization/Team）由开发者自己建表实现。
- **资源归属单位**：
  - 共享 Schema 模式：所有租户共享表，用 `tenant_id` 列隔离（推荐）。
  - Schema-per-tenant：每个租户一个 PostgreSQL Schema。
  - Database-per-tenant：每个租户一个数据库（Supabase 不推荐，成本高）。
- **角色模型**：依赖 PostgreSQL RLS (Row Level Security)。通过 RLS Policy 检查当前用户是否拥有该 `tenant_id` 的访问权。
- **数据隔离策略**：
  - **RLS + tenant_id**：最主流。在每个表上加 `tenant_id` 或 `org_id`，RLS policy 做自动过滤。
  - 应用层通过 `set_config('app.current_tenant', ...)` 注入当前租户，RLS 使用 `current_setting()` 读取。
- **常见 Migration 模式**：
  - 添加 `tenant_id` / `organization_id` 列（nullable → not null after backfill）。
  - 批量 backfill：根据已有 `user_id` 关联到默认 Organization 填充 `org_id`。
  - 逐步启用 RLS：先创建 policy 为 permissive，测试后再收紧。

---

## 二、对比总览表

| 维度 | Stripe | Clerk | Auth0 Organizations | Supabase |
|------|--------|-------|---------------------|----------|
| **层级深度** | Account → Member（扁平，无 Team） | Organization → Member（扁平，Team 需自定义扩展） | Organization → Member → Group（可选 Team 等价物） | 应用层自定（推荐 Org → Team → User） |
| **资源归属单位** | Account | Organization | 应用层自定，Auth0 提供 `org_id` | 共享表 + `tenant_id` / `org_id` |
| **角色模型** | 预定义 Role（Owner/Admin/Dev/Analyst） | 预定义 + 自定义 Role + Permission 字符串 | 自定义 Organization Role + Permission | 应用层 Role 表 + RLS Policy |
| **数据隔离策略** | Account ID 全表过滤 | `org_id` Session/API 过滤 | `org_id` Token 声明 + 应用层过滤 | PostgreSQL RLS + `tenant_id` 列 |
| **常见 Migration 模式** | 原 Account Owner → Org Owner，生成 default Team | User → Personal Org → 可选迁移 | User Metadata `company_id` → Organization + Membership Backfill | 添加 `org_id` 列 → Backfill → 启用 RLS |
| **API Key/凭证隔离** | Account-scoped Keys | Instance-scoped（应用层自行隔离） | Application-scoped | Project-scoped（应用层自行隔离） |
| **成员跨组织** | 支持（一个 User 多个 Account） | 支持（一个 User 多个 Organization） | 支持（一个 User 多个 Organization） | 应用层实现 |
| **邀请/审批机制** | 邮件邀请，Owner/Admin 可管理 | 内置 Invitation + 域名自动加入 | 内置 Invitation + JIT Provision | 应用层实现 |

---

## 三、对 Go/Ent + PostgreSQL 项目的 Schema 设计建议

### 3.1 核心设计原则
1. **Schema 支持三级，但 MVP 强制单 Default Team**：为未来预留扩展性，避免二次大迁移。
2. **所有业务表必须含 `organization_id`**：Team 级别的资源归属在 MVP 阶段通过 default Team 映射到 Organization，但 Schema 上要能支持 `team_id`。
3. **User 是全局身份，Member 是组织内身份**：区分全局 `users` 表与组织内 `organization_members` 表。
4. **角色使用 Permission-Based (RBAC)**：避免硬编码角色，通过 `permissions` 表 + 关联表实现细粒度控制。
5. **PostgreSQL RLS 作为安全网**：即使应用层已做过滤，RLS 可防止漏网之鱼（如直接连 DB 查询）。

### 3.2 推荐 Entity Schema (Ent)

```go
// User：全局用户身份
func (User) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.String("email").Unique().NotEmpty(),
        field.String("password_hash").Sensitive(),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

// Organization：顶层租户单位
func (Organization) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.String("name").NotEmpty(),
        field.String("slug").Unique().NotEmpty(), // 用于子域名或 URL
        field.UUID("owner_id", uuid.UUID{}),      // 创建者，冗余加速查询
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}
func (Organization) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("teams", Team.Type),
        edge.To("members", OrganizationMember.Type),
        edge.To("owner", User.Type).Field("owner_id").Unique().Required(),
    }
}

// Team：组织内的子单元。MVP 阶段每个 Organization 只有一个 default Team。
func (Team) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.String("name").NotEmpty(),
        field.UUID("organization_id", uuid.UUID{}),
        field.Bool("is_default").Default(false), // MVP 保障只有一个 true
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}
func (Team) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("organization", Organization.Type).Ref("teams").Field("organization_id").Unique().Required(),
        edge.To("members", TeamMember.Type),
    }
}

// OrganizationMember：用户在组织内的成员身份（与 Team 解耦）
func (OrganizationMember) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("organization_id", uuid.UUID{}),
        field.UUID("user_id", uuid.UUID{}),
        field.Enum("status").Values("active", "invited", "suspended").Default("active"),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}
func (OrganizationMember) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("organization", Organization.Type).Ref("members").Field("organization_id").Unique().Required(),
        edge.From("user", User.Type).Ref("organization_memberships").Field("user_id").Unique().Required(),
        edge.To("roles", OrganizationMemberRole.Type),
    }
}

// TeamMember：用户在特定 Team 中的归属。MVP 阶段所有 OrgMember 自动加入 Default Team。
func (TeamMember) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("team_id", uuid.UUID{}),
        field.UUID("organization_member_id", uuid.UUID{}),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

// Role：预定义或自定义角色
func (Role) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.String("name").NotEmpty(),
        field.String("key").Unique().NotEmpty(), // e.g. "owner", "admin", "developer"
        field.UUID("organization_id", uuid.UUID{}).Optional(), // null = 系统全局角色
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

// Permission：细粒度权限点
func (Permission) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.String("resource").NotEmpty(), // e.g. "api_key", "billing"
        field.String("action").NotEmpty(),   // e.g. "read", "write", "delete"
        field.String("key").Unique().NotEmpty(), // e.g. "api_key:read"
    }
}

// RolePermission：角色拥有哪些权限
// OrganizationMemberRole：成员在组织中拥有哪些角色

// API Key / Service Account 示例（归属 Organization，可选归属 Team）
func (APIKey) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("organization_id", uuid.UUID{}),
        field.UUID("team_id", uuid.UUID{}).Optional(), // MVP 阶段填 default team id
        field.String("name"),
        field.String("prefix"),     // 可读的 key 前缀
        field.String("hash").Sensitive(), // 存储 hash，不存明文
        field.Time("expires_at").Optional(),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}
```

### 3.3 业务资源表统一规范
所有业务资源（如 Gateway、Route、Quota、Log）建议增加以下字段：

```sql
ALTER TABLE your_resource_table ADD COLUMN organization_id UUID NOT NULL;
ALTER TABLE your_resource_table ADD COLUMN team_id UUID; -- MVP 可 NULL 或强制 default team
ALTER TABLE your_resource_table ADD COLUMN created_by UUID; -- user_id，审计用
```

### 3.4 PostgreSQL RLS Policy 建议

```sql
-- 1. 在应用连接时注入当前 organization_id（通过 set_config）
--    e.g. SELECT set_config('app.current_org_id', '...', false);

-- 2. 为所有业务表启用 RLS
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;

-- 3. 创建 Policy（示例）
CREATE POLICY org_isolation ON api_keys
    USING (organization_id = current_setting('app.current_org_id', true)::UUID);

-- 4. 对于 super admin 角色，可创建 bypass RLS 的 policy 或使用超级用户连接
```

> **注意**：Go/Ent 目前不原生生成 RLS，需要手写 migration SQL。建议在 `ent/migrate` 后追加自定义 SQL migration。

### 3.5 Migration / Backfill 方案

#### Step 1: Schema 变更（零停机）
1. 创建 `organizations`, `teams`, `organization_members`, `roles`, `permissions` 等新表。
2. 在现有业务表上添加 `organization_id` 和 `team_id` 列（先 `NULLABLE`）。

#### Step 2: 数据 Backfill
1. 为每个现有 `user` 创建默认 `organization`（名称为用户昵称或 email 前缀）。
2. 在该 organization 下创建 `default_team`（`is_default = true`）。
3. 创建 `organization_member` 记录，user → owner。
4. 将现有业务表的 `owner_id` 或 `user_id` 映射到对应的 `organization_id` 和 `team_id` 并更新。

#### Step 3: 约束收紧
1. 业务表 `organization_id` 改为 `NOT NULL`。
2. 添加 `organization_id` 上的索引（B-Tree，查询高频）。
3. 可选：添加复合索引 `(organization_id, team_id)`。

#### Step 4: 应用层改造
1. 所有 Repository/DAO 层查询增加 `Where(organization_id.EQ(...))`。
2. 认证中间件解析 JWT/Session 后，将 `organization_id` 注入 Context。
3. API 路由增加 `/orgs/{org_id}/...` 或 Header `X-Organization-ID` 支持。

---

## 四、决策建议摘要

| 决策项 | 推荐方案 | 理由 |
|--------|----------|------|
| **层级深度** | Org → Team → User | Stripe/Clerk 扁平模型在复杂 B2B 场景下扩展性弱；预留 Team 支持未来部门/项目隔离 |
| **资源归属** | 优先归 Organization，可选 Team | 与 Stripe 一致；API Key、Billing 等适合 Org 级；Gateway Route 等可归 Team |
| **角色模型** | RBAC (Role + Permission) | Auth0/Clerk 都支持自定义 Role + Permission，避免硬编码 |
| **数据隔离** | 应用层过滤 + PostgreSQL RLS 双保险 | Supabase 最佳实践；应用层负责性能，RLS 负责安全兜底 |
| **Migration** | 一 User 一 Default Org + Default Team | Clerk Personal Org 模式的变体；对用户无感知 |
| **跨组织用户** | 支持 | Clerk/Auth0/Stripe 均支持；通过独立 Membership 表实现 |

---

*文档生成时间: 2025-04-16*
*适用于: Sub2API (Go + Ent + PostgreSQL) B2B 组织化改造*
