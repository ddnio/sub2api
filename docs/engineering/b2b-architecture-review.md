# Sub2API B端架构方案审查报告

> 审查人：B端SaaS架构专家
> 日期：2026-04-16
> 项目：Sub2API (Go + Ent + PostgreSQL)
> 状态：拟定方案尚未落地

---

## 一、当前架构基线与拟定方案摘要

### 1.1 当前C端架构（已落地）

| 实体 | 核心职责 | 关键字段 |
|------|---------|---------|
| `users` | 唯一身份与计费主体 | `balance`, `concurrency`, `role` |
| `api_keys` | 归属个人用户 | `user_id` (FK), `quota`, `rate_limit_*` |
| `usage_logs` | 只追加调用记录 | `user_id`, `api_key_id`, `account_id` |
| `user_subscriptions` | 订阅配额控制 | `user_id`, `group_id`, `daily_usage_usd` |
| `payment_orders` | 充值订单 | `user_id`, `amount`, `status` |
| `groups` | **上游账号分组**（非团队概念） | `platform`, `rate_multiplier`, `model_routing` |

**核心计费路径（高度耦合User）：**
```
Gateway → UsageBillingCommand{UserID} → usage_billing_repo.Apply()
  → UPDATE users SET balance = balance - $1 WHERE id = $2
```

### 1.2 拟定B端方案

```
Organization (治理层, 充值入口)
    └── Team (一等资源所有者: api_keys, usage_logs, 计费扣 team.balance)
        └── TeamMember (owner/admin/member)
    └── OrganizationMember (owner/admin/member)

C端兼容: 每个 user 自动生成隐形的 Personal Organization + Personal Team
```

---

## 二、与行业主流架构对比评估

### 2.1 OpenAI Organization / Project 模型

| 维度 | OpenAI | 拟定方案 | 偏差分析 |
|------|--------|---------|---------|
| 资源所有者 | **Project** 是一等API Key所有者 | **Team** 是一等资源所有者 | ✅ 基本对齐 |
| 计费主体 | Organization统一billing；Project按tag分账 | Org充值 → 分配budget给Team | ⚠️ 方向相反。OpenAI是**事后分账**，拟定方案是**事前配额** |
| 权限粒度 | Owner / Reader / Billing (跨Project可配置) | Org owner/admin/member + Team owner/admin/member | ❌ 缺少跨Team的细粒度权限（如只读 analyst） |
| API Key隔离 | Project级别完全隔离 | 未明确 | ⚠️ 风险：Team间Key泄露的隔离策略未定义 |

**关键差异**：OpenAI的Project本身**不持有balance**，所有费用统一计到Organization，再通过cost center/tag做内部分账。拟定方案让Team直接扣费，这在B端Enterprise场景中会导致**预算超支控制困难**（Org想控总预算，但Team是实际扣费点）。

### 2.2 Anthropic Organization / Workspace 模型

| 维度 | Anthropic | 拟定方案 | 偏差分析 |
|------|-----------|---------|---------|
|  Workspace | 与 Team 概念等价 | Team | ✅ 对齐 |
| 权限 | 更强调 IAM + SAML Group Mapping | 简单角色 | ❌ 预留空间不足 |
| Usage Dashboard | Workspace 级别 + Org 级别聚合 | 未提及 | ⚠️ 查询架构需要提前设计 |

Anthropic 近期向 Enterprise Plan 过渡时，最大的架构改动就是**将 Workspace 的计费从独立账单合并到 Org 统一账单**，这与拟定方案"Org分配budget给Team"的方向再次形成对比。

### 2.3 LiteLLM Organization / Team / Budget 模型

| 维度 | LiteLLM | 拟定方案 | 偏差分析 |
|------|---------|---------|---------|
| 预算管理 | **独立的 `Budget` 表**，可绑定到 Team/Org/Key/User | budget 是 team/org 的字段 | ⚠️ 严重差异 |
| 灵活性 | 支持 soft budget / hard budget / time-window budget | 未体现 | ❌ 扩展性不足 |
| 模型限制 | Budget 可限制模型白名单 | 未体现 | ⚠️ 需考虑 |

LiteLLM 的架构成功之处就在于**将"资金/预算"与"组织实体"解耦**。拟定方案把 balance 放在 org/team 实体上，是C端思维的自然延伸，但B端场景下会迅速成为瓶颈。

---

## 三、3个最大风险（按严重性排序）

### 风险1：C端兼容策略（隐形 Personal Organization）将产生长期架构债务 [严重性: 高]

**问题描述：**
为每个现有 user 自动创建"不可见的 Personal Organization + Personal Team"，并将其所有数据迁移到该 Team。这意味着：
1. 未来**每一个查询**都需要经过 `users → organization_members → teams → team_members` 的 JOIN 链，或至少携带 `team_id`。
2. 现有代码中大量基于 `user_id` 的索引、权限检查、缓存键（如 `billingCacheService.GetUserBalance(userID)`）需要全量重写。
3. 这种模型混淆了"自然人"和"组织"的法律主体概念。当用户以后真的被邀请加入一个B端Org时，会同时存在"Personal Org"和"Enterprise Org"两个平行身份，导致数据归属和导出义务（如GDPR数据可携带权）极其复杂。

**与主流对比：**
- OpenAI 不存在"Personal Organization"：个人账户就是个人账户，切换到 Organization 是一个明确的 context switch。
- LiteLLM 的 User 可以独立存在，也可以被分配到 Team，但不会强制给每个 User 造一个 1:1 的 Org。

### 风险2："财权"与"事权"分离导致B端计费规则冲突 [严重性: 高]

**问题描述：**
拟定方案中，Organization 是充值入口（有钱），Team 是一等资源所有者且直接扣 `team.balance`（花钱）。这种分离在B端实际运营中会产生大量边界未定义问题：

1. **超支处理**：当 Team 的 budget 耗尽，但 Org 账户还有余额，API Key 应该被暂停（hard limit）还是允许透支（soft limit / shared pool）？
2. **退款归属**：如果某 Team 的 usage 发生退款，钱是退到 Team 还是 Org？跨 Team 的余额转移是否允许？
3. **发票主体**：B端客户要求发票抬头是 Organization 名称，但消费明细按 Team 拆分。如果 Team 是扣费主体，财务对账会变得困难。

**当前代码风险：**
现有 `usage_billing_repo.go` 中的扣费是原子性 `UPDATE users SET balance = balance - $1`。如果直接复制为 `UPDATE teams SET balance = balance - $1`，则 Org 级别的总预算控制将完全失去原子性，需要引入分布式事务或补偿机制，复杂度陡增。

### 风险3：权限模型缺少B端关键能力预留，未来演进成本极高 [严重性: 中高]

**问题描述：**
拟定方案只有简单的 `owner/admin/member` 双层级角色，缺少B端Enterprise SaaS的以下核心能力预留：

| 能力 | 拟定方案状态 | 行业必要性 |
|------|-------------|-----------|
| **SSO / SAML** | 无映射对象 | 必需。需要 `organization.sso_provider`, `organization_members.sso_external_id` |
| **SCIM 用户同步** | 无 external_id | 必需。需要 `organization_members.scim_external_id`, `team_members.scim_group_mapping` |
| **审计日志 (Audit Log)** | 未提及 | 必需。需要 `actor_type` (user/api_key/system) 和 `actor_organization_id` |
| **邀请机制 (Invite)** | 未提及 | 必需。B端通常由 Admin 发邮件邀请，需要独立的 `invitations` 表 |
| **细粒度权限 (RBAC/ABAC)** | 仅 owner/admin/member | 高阶需求。如"只读 Analyst"、"Billing Manager 但不能看 API Key" |
| **API Key 的服务账号 (Service Account)** | 未提及 | 常见需求。CI/CD 用的 Key 不应绑定到某个离职员工的 User |

**最紧迫的问题：** 如果现在不预留 `actor_organization_id` / `actor_team_id` 等审计字段，未来补审计日志时，需要对 `usage_logs`、`payment_orders`、`api_keys` 等核心表做破坏性加字段迁移，成本极高。

---

## 四、3个具体改进建议（可操作）

### 建议1：将 User 升级为 "Account" 概念，取消"隐形 Personal Organization" [优先级: P0]

**具体方案：**
不要给每个 User 创建隐形 Org，而是让 `users` 表本身具备**独立账户能力**。引入一个顶层抽象（可以是逻辑上的，不一定需要新表）：

```
Account (抽象)
  ├── PersonalAccount (就是现有的 user，直接拥有 balance/api_keys)
  └── OrganizationAccount (新的 Organization 实体)
       ├── Team (Project/Workspace)
       └── Member (User 的关联)
```

**操作步骤：**
1. 在 `api_keys`、`usage_logs`、`payment_orders` 等表中，新增 `owner_type` (`user` | `team`) 和 `owner_id` 字段。
2. C端用户的数据 `owner_type = 'user'`, `owner_id = user.id`，**不需要任何迁移到 Team 的操作**。
3. B端场景下，`owner_type = 'team'`, `owner_id = team.id`。
4. 所有查询统一基于 `(owner_type, owner_id)` 的复合索引，而不是 `user_id` 或 `team_id` 单字段。

**收益：**
- 彻底避免隐形 Org 带来的概念混淆。
- C端代码几乎零迁移（保留 `user_id` 作为自然人关联，但资源所有权由 `owner_*` 表达）。
- 未来一个 User 可以同时拥有 Personal Account 和属于多个 Organization，完全清晰。

### 建议2：引入独立的 `wallets` 和 `budgets` 表，将资金与组织实体解耦 [优先级: P0]

**Schema 建议：**

```go
// Wallet 表：真实的资金池
type Wallet struct {
    ent.Schema
}
func (Wallet) Fields() []ent.Field {
    return []ent.Field{
        field.String("owner_type").MaxLen(20), // "organization" | "team" | "user"
        field.Int64("owner_id"),
        field.Float("balance").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
        field.Float("credit_limit").Default(0), // 后付信用额度
        field.String("currency").MaxLen(3).Default("USD"),
        field.String("status").MaxLen(20).Default("active"),
    }
}

// Budget 表：预算控制规则（可绑定到 Team / API Key / User）
type Budget struct {
    ent.Schema
}
func (Budget) Fields() []ent.Field {
    return []ent.Field{
        field.String("scope_type").MaxLen(20), // "organization" | "team" | "api_key" | "user"
        field.Int64("scope_id"),
        field.Float("limit_amount"),           // 0 = unlimited
        field.String("limit_period").MaxLen(20), // "monthly" | "daily" | "total"
        field.Time("window_start"),
        field.Float("spent_amount").Default(0),
        field.Bool("is_hard_limit").Default(true), // true=超支即停，false=告警
    }
}
```

**计费路径重构：**
```
Gateway → UsageBillingCommand
  → 找到该请求对应的 Wallet (通过 API Key → Team → Wallet)
  → 原子扣减 Wallet.balance
  → 同步更新相关 Budget.spent_amount
  → 如果 Budget 是 hard limit 且超限，返回特定错误码（如 429 OUT_OF_BUDGET）
```

**收益：**
- 与 LiteLLM 对齐，具备极强的灵活性。
- Organization 可以拥有主 Wallet，Team 可以共享 Org Wallet 或拥有独立 Wallet。
- 支持未来"Credit Grant"、"Promotional Balance"、"多币种 Wallet"等高级功能。

### 建议3：在核心表中预埋审计与B端演进字段 [优先级: P1]

**即使第一期不实现完整功能，也要在 Schema 中预留字段：**

#### 3.1 `organizations` 表预留
```go
field.String("sso_provider").Optional().Nillable(),       // e.g., "saml", "oidc"ield.String("sso_domain").Optional().Nillable(),         // e.g., "nanafox.com"ield.JSON("sso_config", map[string]interface{}{}).Optional(),
field.String("scim_provider").Optional().Nillable(),
field.JSON("settings", map[string]interface{}{}).Optional(), // 如 mfa_required, invite_only
```

#### 3.2 `organization_members` / `team_members` 表预留
```go
field.String("role").MaxLen(20),                           // 现有
field.String("sso_external_id").Optional().Nillable(),     // SAML/OIDC sub
field.String("scim_external_id").Optional().Nillable(),    // SCIM 同步ID
field.JSON("permissions", []string{}).Optional(),          // 未来细粒度权限
field.Time("joined_at").Default(time.Now),
field.Time("last_active_at").Optional().Nillable(),
```

#### 3.3 `usage_logs` 和 `api_keys` 表预留审计字段
```go
// usage_logs
field.Int64("organization_id").Optional().Nillable(),
field.Int64("team_id").Optional().Nillable(),
// 保留 user_id 作为"操作人"身份，owner_type/owner_id 作为"资源所有者"

// api_keys
field.Int64("organization_id").Optional().Nillable(),
field.Int64("team_id").Optional().Nillable(),
field.String("created_by").Optional().Nillable(), // 创建者身份追踪
field.Bool("is_service_account").Default(false),  // 服务账号标记
```

#### 3.4 新增 `audit_logs` 表（第一期可只建表不写入）
```go
type AuditLog struct {
    ent.Schema
}
func (AuditLog) Fields() []ent.Field {
    return []ent.Field{
        field.Time("created_at").Default(time.Now),
        field.String("action").MaxLen(50).NotEmpty(),         // e.g., "api_key.created"
        field.String("actor_type").MaxLen(20),                // "user" | "api_key" | "system"
        field.Int64("actor_id"),
        field.Int64("actor_organization_id").Optional().Nillable(),
        field.Int64("actor_team_id").Optional().Nillable(),
        field.String("resource_type").MaxLen(50),             // "api_key" | "team" | "wallet"
        field.Int64("resource_id"),
        field.JSON("before", map[string]interface{}{}).Optional(),
        field.JSON("after", map[string]interface{}{}).Optional(),
        field.String("ip_address").Optional().Nillable(),
        field.String("user_agent").Optional().Nillable(),
    }
}
```

**收益：**
- 避免未来做破坏性 Schema 迁移。
- 向 waiting 的B端客户展示长期架构的完备性，增强信任。

---

## 五、是否应该引入独立的 Budget/Quota 表？

### 结论：**强烈建议引入**。

### 5.1 利弊分析

#### ✅ 利（Pros）

| 维度 | 说明 |
|------|------|
| **与行业对齐** | LiteLLM 的独立 Budget 表已经被市场验证，是B端AI Gateway的事实标准 |
| **资金与组织解耦** | Organization 可以专心做治理和身份管理，Wallet/Budget 做资金和配额管理。一个 Org 可以有多个 Wallet（如不同币种、不同部门），一个 Team 也可以共享或独占 Wallet |
| **复杂预算策略** | 支持 monthly/daily/total 预算、hard/soft limit、模型级别预算（如"GPT-4 每月限额 $1000"）、API Key 级别预算 |
| **退款和对账清晰** | 退款操作在 Wallet 层面统一处理，Budget 只是消费计划，不影响真实资金归属 |
| **兼容C端** | C端 User 直接关联一个 Personal Wallet，无需引入 Organization 概念即可支持余额和quota |

#### ❌ 弊（Cons）

| 维度 | 说明 |
|------|------|
| **Schema 复杂度增加** | 需要新增 `wallets`、`wallet_transactions`、`budgets`、`budget_spends` 等表 |
| **计费路径重构工作量大** | 现有 `usage_billing_repo.go` 中的原子扣费逻辑需要改为 Wallet 扣费 + Budget 更新，需要更仔细的事务设计 |
| **Gateway 侵入性** | 需要在 API Key 认证阶段就解析出对应的 Wallet 和 Budget，增加了热路径的查询成本（可通过缓存缓解） |

### 5.2 最小可行方案（MVP）

如果担心第一期工作量过大，可以采用**最小可行版本**：

1. **只建 `wallets` 表，不建 `budgets` 表（第一期）**：
   - `users` 关联 `personal_wallet_id`
   - `teams` 关联 `team_wallet_id`
   - `organizations` 关联 `org_wallet_id`
   - 扣费时统一走 `UPDATE wallets SET balance = balance - $1`

2. **Budget 逻辑用现有 `api_keys.quota` 字段过渡**：
   - API Key 级别的预算控制继续用现有 `api_keys.quota` 字段。
   - Team 级别的预算控制第一期通过后台 Admin 手动监控 `usage_logs` 聚合实现。

3. **第二期再引入 `budgets` 表**，将 `api_keys.quota` 迁移到 `budgets` 中。

---

## 六、总体评估与建议路径

### 6.1 拟定方案评分

| 维度 | 评分 | 说明 |
|------|------|------|
| 与行业主流对齐 | 6/10 | Team=Project 对齐，但 Org-Team 财权事权分离与主流相反 |
| 权限模型完整性 | 4/10 | 缺少 SSO/SCIM/审计日志/细粒度权限预留 |
| 多租户隔离严谨性 | 5/10 | Team 作为一等所有者合理，但跨 Team 数据查询的 RACL 未定义 |
| 未来功能演进空间 | 5/10 | 缺少关键字段预留，演进将伴随高成本 Schema 迁移 |

### 6.2 推荐实施路径

**Phase 1: Schema 重设计（4-6 周）**
1. 废弃"隐形 Personal Organization"方案。
2. 引入 `owner_type` / `owner_id` 资源所有权模型。
3. 新建 `organizations`, `teams`, `organization_members`, `team_members` 表（带B端预留字段）。
4. 新建 `wallets` 表，将 `users.balance` 迁移到 `wallets`。
5. 重写 `usage_billing_repo.go`，扣费指向 Wallet。

**Phase 2: Gateway 适配（3-4 周）**
1. 认证中间件从 API Key 解析出 `owner_type/owner_id` 和 `wallet_id`。
2. 重构 `BillingContext`，屏蔽 User/Team 差异，统一面向 Wallet。
3. 所有 `usage_logs` 写入同时记录 `organization_id` 和 `team_id`。

**Phase 3: B端功能上线（2-3 周）**
1. Admin 控制台：Org/Team/Member CRUD。
2. 用户控制台：Organization Switcher（类似 OpenAI 的左上角下拉框）。
3. 邀请机制：邮件邀请 + 链接加入。

**Phase 4: 企业级功能（按需）**
1. SSO/SAML。
2. SCIM 同步。
3. Audit Log 查询面板。
4. 独立 `budgets` 表与高级配额策略。

---

*End of Review*
