# Referral Feature Implementation Plan

**Date:** 2026-04-08
**Status:** Draft v2 (Revised)
**Risk Level:** HIGH (production online iteration, database migration, registration flow modification)

---

## Principles (Non-Negotiable)

1. **Zero-downtime migration**: All database changes must be additive (ADD COLUMN, CREATE TABLE). No ALTER that locks tables or breaks existing rows. Migration must be safe to run on a live PostgreSQL with concurrent traffic.
2. **Registration flow must never break**: The referral code logic is strictly additive to the existing registration path. A bad referral code must never prevent registration. All referral failures are non-critical errors (log and continue).
3. **Feature-flag everything**: The entire referral system is gated behind `referral_enabled` setting. When disabled, zero code paths are affected -- the system behaves exactly as today.
4. **Manual wire_gen.go discipline**: Every new service/handler dependency must be manually wired in `wire_gen.go`. No `wire gen` command.
5. **Soft-delete awareness**: Users table uses soft delete (`deleted_at`). The `referral_code` UNIQUE constraint must be a partial unique index (`WHERE deleted_at IS NULL`) to match the existing pattern.

---

## Decision Drivers (Top 3)

1. **Production safety**: This is a live system. Migration safety, rollback capability, and graceful degradation are the highest priority.
2. **Minimal blast radius**: Changes to the registration flow (`auth_service.go`) must be surgical. The referral logic should be encapsulated in a new `ReferralService` rather than bloating `AuthService`.
3. **Consistency with existing patterns**: Follow the established patterns for settings (SettingService), services (RedeemService/PromoService), handlers (AuthHandler), and frontend nav (AppSidebar computed items).

---

## Options Analysis

### Option A: New ReferralService (Recommended)

Create a dedicated `ReferralService` that owns all referral logic. `AuthService` calls `ReferralService.ProcessRegistrationReferral()` after user creation, mirroring how it already calls `PromoService.ApplyPromoCode()`.

**Pros:**
- Single Responsibility: referral logic is self-contained
- Easy to test independently
- Matches existing pattern (PromoService is separate from AuthService)
- Easy to disable/remove later

**Cons:**
- One more service to wire in wire_gen.go
- Slightly more files to create

### Option B: Inline in AuthService

Add referral code handling directly inside `RegisterWithVerification()` and add referral query methods to `UserService`.

**Pros:**
- Fewer new files
- All registration logic in one place

**Cons:**
- AuthService is already large (~650 lines)
- Violates SRP; harder to test referral logic in isolation
- Harder to remove if feature is deprecated
- Mixes concerns: auth + referral rewards

**Decision: Option A** -- New ReferralService. Matches the existing PromoService pattern, keeps AuthService focused, and is easier to test and maintain.

---

## ADR: Referral Feature Architecture

- **Decision:** Dedicated ReferralService + UserReferral ent schema + raw SQL migration
- **Drivers:** Production safety, minimal blast radius, pattern consistency
- **Alternatives considered:** Inline in AuthService (rejected: SRP violation, AuthService already large)
- **Why chosen:** Follows PromoService pattern, encapsulated, testable, easy to feature-flag
- **Consequences:** One additional service in wire_gen.go; one new ent schema + migration file
- **Follow-ups:** Admin UI for referral stats (future iteration); rate limiting on referral rewards (future)

---

## Task Flow (6 Steps)

### Step 1: Database Migration + Ent Schema

**Files to create/modify:**
- `backend/migrations/091_add_referral_system.sql` (NEW)
- `backend/ent/schema/user.go` (MODIFY -- add `referral_code` field)
- `backend/ent/schema/user_referral.go` (NEW)

> **Migration numbering note:** Upstream has migrations up to `090_drop_sora.sql`. Use 091 here. At implementation time, confirm the actual next available number by checking `max(local, upstream) + 1` against both the local `backend/migrations/` directory and any pending upstream merges.

**Migration SQL (`091_add_referral_system.sql`):**
```sql
-- Add referral_code to users (nullable, partial unique for soft delete)
ALTER TABLE users ADD COLUMN IF NOT EXISTS referral_code VARCHAR(16);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_referral_code_active
  ON users (referral_code) WHERE deleted_at IS NULL AND referral_code IS NOT NULL;

-- Create user_referrals table
CREATE TABLE IF NOT EXISTS user_referrals (
  id                BIGSERIAL PRIMARY KEY,
  inviter_id        BIGINT NOT NULL REFERENCES users(id),
  invitee_id        BIGINT NOT NULL REFERENCES users(id),
  code              VARCHAR(16) NOT NULL,
  inviter_rewarded  DECIMAL(20,8) DEFAULT 0,
  invitee_rewarded  DECIMAL(20,8) DEFAULT 0,
  created_at        TIMESTAMPTZ DEFAULT now(),
  UNIQUE(invitee_id),
  CONSTRAINT no_self_referral CHECK (inviter_id != invitee_id)
);
CREATE INDEX IF NOT EXISTS idx_user_referrals_inviter_id ON user_referrals(inviter_id);
```

**Ent schema changes:**

`user.go` -- add field:
```go
field.String("referral_code").
    MaxLen(16).
    Optional().
    Nillable(),
```
Add edge: `edge.To("referrals_as_inviter", UserReferral.Type)`

`user_referral.go` (new schema):
- Fields: `inviter_id`, `invitee_id`, `code`, `inviter_rewarded`, `invitee_rewarded`, `created_at`
- Edges: `From("inviter", User.Type)`, `From("invitee", User.Type)`
- Table annotation: `user_referrals`

After schema changes, run: `cd backend && go generate ./ent/...`

**Acceptance Criteria:**
- [ ] Migration applies cleanly on a fresh DB and on a DB with existing users (idempotent with `IF NOT EXISTS`)
- [ ] `referral_code` column exists on `users` with partial unique index
- [ ] `user_referrals` table exists with correct constraints
- [ ] `go generate ./ent/...` runs without errors
- [ ] Existing users have `referral_code = NULL` (populated lazily or via backfill)

---

### Step 2: Backend Service Layer (ReferralService + Settings)

**Files to create/modify:**
- `backend/internal/service/referral_service.go` (NEW)
- `backend/internal/service/setting_service.go` (MODIFY -- add referral settings)
- `backend/internal/service/settings_view.go` (MODIFY -- add referral fields to SystemSettings)
- `backend/internal/service/auth_service.go` (MODIFY -- call ReferralService after registration)

**ReferralService struct and dependencies:**

Following the PromoService pattern (see `promo_service.go`), ReferralService must include `billingCacheService` and `authCacheInvalidator` to invalidate caches after balance changes:

```go
type ReferralService struct {
    entClient            *dbent.Client
    userRepo             UserRepository
    settingService       *SettingService
    billingCacheService  *BillingCacheService
    authCacheInvalidator APIKeyAuthCacheInvalidator
}

func NewReferralService(
    entClient *dbent.Client,
    userRepo UserRepository,
    settingService *SettingService,
    billingCacheService *BillingCacheService,
    authCacheInvalidator APIKeyAuthCacheInvalidator,
) *ReferralService
```

**ReferralService methods:**

```go
// GenerateReferralCode generates a unique 16-char code for a user.
// Called unconditionally right after userRepo.Create succeeds (separate from ProcessRegistrationReferral).
func (s *ReferralService) GenerateReferralCode(ctx context.Context, userID int64) (string, error)

// ProcessRegistrationReferral handles referral logic after a new user registers.
// Uses a transaction to atomically: INSERT user_referrals + UPDATE inviter balance + UPDATE invitee balance.
// Post-commit: invalidates billingCacheService and authCacheInvalidator for both users.
// Outer caller treats errors as non-critical (log and continue).
func (s *ReferralService) ProcessRegistrationReferral(ctx context.Context, inviteeID int64, referralCode string) error

// GetReferralInfo returns the user's referral code and stats
func (s *ReferralService) GetReferralInfo(ctx context.Context, userID int64) (*ReferralInfo, error)

// ListReferrals returns paginated list of users invited by this user
func (s *ReferralService) ListReferrals(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ReferralRecord, *pagination.PaginationResult, error)

// GetReferralByInvitee returns the referral record for an invitee (for admin view)
func (s *ReferralService) GetReferralByInvitee(ctx context.Context, inviteeID int64) (*ReferralRecord, error)
```

**ProcessRegistrationReferral transaction pattern** (mirrors `PromoService.ApplyPromoCode`):
```go
func (s *ReferralService) ProcessRegistrationReferral(ctx context.Context, inviteeID int64, referralCode string) error {
    // ... validation, lookup inviter ...

    // Transaction: atomic reward distribution
    tx, err := s.entClient.Tx(ctx)
    if err != nil { return err }
    defer func() { _ = tx.Rollback() }()
    txCtx := dbent.NewTxContext(ctx, tx)

    // INSERT user_referrals record
    // UPDATE inviter balance (if inviter_amount > 0)
    // UPDATE invitee balance (if invitee_amount > 0)

    if err := tx.Commit(); err != nil { return err }

    // Post-commit cache invalidation (non-transactional)
    if s.billingCacheService != nil {
        go func() {
            cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            _ = s.billingCacheService.InvalidateUserBalance(cacheCtx, inviterID)
            _ = s.billingCacheService.InvalidateUserBalance(cacheCtx, inviteeID)
        }()
    }
    if s.authCacheInvalidator != nil {
        s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, inviterID)
        s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, inviteeID)
    }

    return nil
}
```

**GenerateReferralCode call location:**
- Called unconditionally in `RegisterWithVerification`, immediately after `userRepo.Create` succeeds.
- Separate from `ProcessRegistrationReferral` -- every new user gets a code regardless of whether they used a referral code to register.

**SettingService additions:**
- New constants: `SettingKeyReferralEnabled`, `SettingKeyReferralInviterAmount`, `SettingKeyReferralInviteeAmount`
- New methods: `IsReferralEnabled(ctx) bool`, `GetReferralInviterAmount(ctx) float64`, `GetReferralInviteeAmount(ctx) float64`
- Add `referral_enabled` to `PublicSettings` struct and `GetPublicSettings()` key list
- Add `referral_enabled`, `referral_inviter_amount`, `referral_invitee_amount` to `SystemSettings` struct and admin `GetAllSettings()`/`UpdateSettings()`

**AuthService modification -- RegisterWithVerification signature unchanged:**

Do NOT add a new `referralCode` parameter. Instead, route the existing `invitationCode` parameter based on settings inside AuthService:

```go
// In RegisterWithVerification, after user creation:

// 1. Generate referral code for the new user (unconditional)
if s.referralService != nil {
    if _, err := s.referralService.GenerateReferralCode(ctx, user.ID); err != nil {
        logger.LegacyPrintf("service.auth", "[Auth] Failed to generate referral code for user %d: %v", user.ID, err)
    }
}

// 2. Handle invitation code vs referral code routing
if s.settingService.IsInvitationCodeEnabled(ctx) {
    // Existing invitation code logic (mandatory, redeem-type)
    // ... (already present, no change)
} else if s.settingService.IsReferralEnabled(ctx) && invitationCode != "" {
    // invitationCode is treated as a referral code
    if err := s.referralService.ProcessRegistrationReferral(ctx, user.ID, invitationCode); err != nil {
        logger.LegacyPrintf("service.auth", "[Auth] Failed to process referral for user %d: %v", user.ID, err)
    } else {
        // Refresh user to get updated balance
        if updatedUser, err := s.userRepo.GetByID(ctx, user.ID); err == nil {
            user = updatedUser
        }
    }
}
```

This avoids any signature change and keeps the routing logic inside AuthService where it has access to `settingService`.

**Acceptance Criteria:**
- [ ] `ReferralService` created with all methods, including `billingCacheService` and `authCacheInvalidator` dependencies
- [ ] `ProcessRegistrationReferral` uses a transaction wrapping INSERT + both balance UPDATEs
- [ ] Post-commit cache invalidation calls both `InvalidateUserBalance` and `InvalidateAuthCacheByUserID`
- [ ] `GenerateReferralCode` called unconditionally after `userRepo.Create`, separate from referral processing
- [ ] Settings keys defined and wired into PublicSettings/SystemSettings
- [ ] `referral_enabled=false` by default (no behavior change on deploy)
- [ ] `RegisterWithVerification` signature unchanged -- routing via `settingService.IsInvitationCodeEnabled()` internally
- [ ] When `invitation_code_enabled=true`: existing invitation code logic, no referral processing
- [ ] When `invitation_code_enabled=false` + `referral_enabled=true`: `invitationCode` treated as referral code
- [ ] Registration with a valid referral code creates a `user_referrals` record and awards balance atomically
- [ ] Registration with invalid/missing referral code proceeds normally (no error)
- [ ] Self-referral (inviter == invitee) silently ignored
- [ ] New user gets a `referral_code` generated automatically on registration

---

### Step 3: Backend Handler + Routes + Wire

**Files to create/modify:**
- `backend/internal/handler/referral_handler.go` (NEW)
- `backend/internal/handler/auth_handler.go` (MODIFY -- pass referral code)
- `backend/internal/handler/handler.go` or `wire.go` (MODIFY -- add ReferralHandler to Handlers)
- `backend/internal/server/routes/user.go` (MODIFY -- add referral routes)
- `backend/cmd/server/wire_gen.go` (MODIFY -- wire ReferralService + ReferralHandler)
- `backend/internal/handler/dto/` (MODIFY -- add referral DTOs)

**Frontend RegisterRequest:**
Add a separate `referral_code` field to `RegisterRequest` DTO (do NOT reuse `invitation_code`):
```go
type RegisterRequest struct {
    Email          string `json:"email"`
    Password       string `json:"password"`
    VerifyCode     string `json:"verify_code"`
    PromoCode      string `json:"promo_code"`
    InvitationCode string `json:"invitation_code"`
    ReferralCode   string `json:"referral_code"`   // NEW: separate field
}
```

The handler maps `req.ReferralCode` (or falls back to `req.InvitationCode` when referral mode is active) before calling `RegisterWithVerification`. Since the AuthService signature is unchanged and routes via settings internally, the handler simply passes whichever value is present as `invitationCode`.

**ReferralHandler endpoints:**
```
GET  /api/v1/user/referral          -> GetReferralInfo (code + stats summary)
GET  /api/v1/user/referral/list     -> ListReferrals (paginated invitee list)
```

**Admin endpoint (optional, add to admin routes):**
```
GET  /api/v1/admin/users/:id/referral  -> GetUserReferralInfo (admin view for a specific user)
```

**Handler struct:**
```go
type ReferralHandler struct {
    referralService *service.ReferralService
}
```

**wire_gen.go changes:**
1. Create `ReferralService` instance with dependencies: `entClient`, `userRepo`, `settingService`, `billingCacheService`, `authCacheInvalidator`
2. Create `ReferralHandler` instance
3. Add `ReferralHandler` to `Handlers` struct
4. Inject `ReferralService` into `AuthService` (add to constructor)

**Acceptance Criteria:**
- [ ] `GET /api/v1/user/referral` returns referral code + stats for authenticated user
- [ ] `GET /api/v1/user/referral/list` returns paginated invitee list
- [ ] `RegisterRequest` has a separate `referral_code` JSON field
- [ ] `wire_gen.go` compiles without errors, includes `billingCacheService` and `authCacheInvalidator` in ReferralService constructor
- [ ] All existing tests still pass

---

### Step 4: Frontend -- Registration Flow + API

**Files to create/modify:**
- `frontend/src/views/auth/RegisterView.vue` (MODIFY)
- `frontend/src/api/referral.ts` (NEW)
- `frontend/src/types/index.ts` (MODIFY -- add `referral_enabled` to PublicSettings type, add `referral_code` to RegisterRequest)
- `frontend/src/stores/app.ts` (MODIFY -- add `referral_enabled` default)

**RegisterView.vue changes:**

The existing `RegisterView.vue` already has conditional rendering:
- `v-if="invitationCodeEnabled"` shows the invitation code input (mandatory)
- `v-if="promoCodeEnabled"` shows the promo code input (optional)

Add a third condition for referral mode:
```vue
<!-- Referral Code Input (Optional, when referral enabled and invitation code disabled) -->
<div v-if="referralEnabled && !invitationCodeEnabled">
  <label for="referral_code" class="input-label">
    {{ t('auth.referralCodeLabel') }}
    <span class="ml-1 text-xs font-normal text-gray-400">({{ t('common.optional') }})</span>
  </label>
  <!-- Similar input as invitation code, but optional, no real-time validation needed -->
</div>
```

**Frontend RegisterRequest:** Add an independent `referral_code` field. Do NOT reuse `invitation_code`. The form binds to `formData.referral_code` when referral mode is active:
```typescript
interface RegisterRequest {
  email: string
  password: string
  verify_code: string
  promo_code?: string
  invitation_code?: string
  referral_code?: string  // NEW: separate field
}
```

**URL pre-fill support:**
```typescript
// In RegisterView.vue setup
const route = useRoute()
onMounted(() => {
  const ref = route.query.ref as string
  if (ref) {
    formData.referral_code = ref
  }
})
```

**API module (`frontend/src/api/referral.ts`):**
```typescript
export function getReferralInfo() { return api.get('/user/referral') }
export function getReferralList(params) { return api.get('/user/referral/list', { params }) }
```

**Acceptance Criteria:**
- [ ] When `referral_enabled=true` and `invitation_code_enabled=false`, referral code input appears on register page
- [ ] When `invitation_code_enabled=true`, only invitation code input appears (existing behavior unchanged)
- [ ] Frontend sends `referral_code` as a separate field in RegisterRequest (not reusing `invitation_code`)
- [ ] URL `?ref=ABC123` pre-fills the referral code
- [ ] Registration with/without referral code works correctly
- [ ] `PublicSettings` type includes `referral_enabled`

---

### Step 5: Frontend -- ReferralView + Sidebar + Admin

**Files to create/modify:**
- `frontend/src/views/user/ReferralView.vue` (NEW)
- `frontend/src/router/index.ts` (MODIFY -- add `/referral` route)
- `frontend/src/components/layout/AppSidebar.vue` (MODIFY -- add referral nav item)
- `frontend/src/i18n/locales/zh.ts` (MODIFY -- add referral translations)
- `frontend/src/i18n/locales/en.ts` (MODIFY -- add referral translations)
- `frontend/src/views/admin/SettingsView.vue` (MODIFY -- add referral settings toggle + amounts)
- `frontend/src/api/admin/settings.ts` (MODIFY -- add referral settings fields)
- Admin `UsersView.vue` edit dialog (MODIFY -- add read-only referral info block)

**ReferralView.vue layout:**
1. **Referral Code Card** (top, prominent): Show the user's referral code with copy button + share link
2. **Reward Summary Card**: Total earned, inviter reward per referral, invitee reward
3. **Invitee List Table**: Paginated table with columns: Email (masked), Registration Date, Inviter Reward, Invitee Reward

**Sidebar nav item** (in both `userNavItems` and `personalNavItems`):
```typescript
// Insert after '/redeem' entry, conditional on referral_enabled
...(appStore.cachedPublicSettings?.referral_enabled
  ? [{ path: '/referral', label: t('nav.referral'), icon: UsersIcon, hideInSimpleMode: true }]
  : []),
```

**Admin SettingsView.vue:**
Add a "Referral" section (near Invitation Code section) with:
- Toggle: `referral_enabled`
- Number input: `referral_inviter_amount` (inviter reward amount)
- Number input: `referral_invitee_amount` (invitee reward amount)

**Admin UsersView.vue:**
In the user edit modal, add a read-only section:
- Referral code (if exists)
- Invited by: [inviter email] on [date] (if this user was invited)
- Invited count: N users

**Acceptance Criteria:**
- [ ] `/referral` route accessible and renders ReferralView
- [ ] Sidebar shows "Referral" link only when `referral_enabled=true`
- [ ] ReferralView shows referral code, copy/share, reward summary, invitee list
- [ ] Admin settings page has referral configuration section
- [ ] Admin user edit dialog shows read-only referral information
- [ ] All i18n keys added for zh and en

---

### Step 6: Testing + Verification

**Scope:**
- Backend unit tests for ReferralService
- Integration verification on test environment
- Manual e2e walkthrough

**Test cases (backend):**
1. `TestReferralService_GenerateCode` -- generates unique 16-char codes
2. `TestReferralService_ProcessReferral_Success` -- creates record + awards balance atomically in transaction
3. `TestReferralService_ProcessReferral_InvalidCode` -- returns nil (no error)
4. `TestReferralService_ProcessReferral_SelfReferral` -- silently ignored
5. `TestReferralService_ProcessReferral_Disabled` -- no-op when setting is off
6. `TestReferralService_ProcessReferral_DuplicateInvitee` -- unique constraint prevents double referral
7. `TestReferralService_ProcessReferral_TransactionFailure` -- verifies atomicity: if any step in the transaction fails, no partial state is committed (no orphan referral record without balance update, no balance update without referral record)
8. `TestReferralService_ProcessReferral_CacheInvalidation` -- verifies `InvalidateUserBalance` and `InvalidateAuthCacheByUserID` called for both inviter and invitee after successful commit
9. `TestRegistration_WithReferralCode` -- full registration flow with referral
10. `TestRegistration_WithoutReferralCode` -- existing flow unaffected

**Integration verification (test environment):**
1. Deploy to `router-test.nanafox.com`
2. Enable `referral_enabled` via admin settings
3. Register User A -> verify referral_code generated
4. Register User B with User A's referral code -> verify:
   - user_referrals record created
   - Both users' balances updated (if reward > 0)
5. User A visits /referral -> sees User B in invitee list
6. Disable `referral_enabled` -> verify referral input disappears from register page
7. Verify existing registration flow (invitation code mode) still works

**Observability:**
- All referral operations logged with `[Referral]` prefix via `logger.LegacyPrintf`
- Failed reward distribution logged at WARN level
- No new metrics needed for v1 (future: referral conversion rate)

**Acceptance Criteria:**
- [ ] All new unit tests pass (including transaction failure and cache invalidation tests)
- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/service/...` passes (existing + new)
- [ ] Frontend builds without errors (`pnpm build`)
- [ ] Test environment deployment successful
- [ ] Manual e2e walkthrough passes all 7 scenarios above

---

## Pre-Mortem: 3 Failure Scenarios

### Scenario 1: Migration locks users table under load
**What happens:** `ALTER TABLE users ADD COLUMN` takes a lock on the users table. On PostgreSQL with concurrent transactions, this can queue behind long-running queries, causing a cascading lock wait.
**Prevention:**
- The `ADD COLUMN ... DEFAULT NULL` (nullable, no default) in PostgreSQL is a metadata-only operation -- it does NOT rewrite the table. This is safe.
- Use `IF NOT EXISTS` for idempotency.
- Deploy during low-traffic window as extra precaution.
- **Rollback:** `ALTER TABLE users DROP COLUMN IF EXISTS referral_code;` is also fast.

### Scenario 2: Referral reward double-spend via race condition
**What happens:** Two concurrent registrations with the same invitee_id (shouldn't happen, but...) or a bug causes double reward distribution.
**Prevention:**
- `UNIQUE(invitee_id)` constraint on `user_referrals` prevents duplicate referral records at DB level.
- Balance updates use `UpdateBalance` which does atomic `UPDATE ... SET balance = balance + $amount`.
- The referral record INSERT and both balance UPDATEs are wrapped in a single database transaction, guaranteeing atomicity. If any step fails, the entire operation rolls back -- no partial state.
- **Monitoring:** Log every reward distribution with amounts for audit.

### Scenario 3: Invitation code mode breaks when referral is also enabled
**What happens:** Both `invitation_code_enabled=true` and `referral_enabled=true`. The single input field tries to serve two purposes, causing confusion.
**Prevention:**
- Frontend logic: when `invitation_code_enabled=true`, always show invitation code mode (mandatory input, existing behavior). Referral input only shows when invitation code is OFF.
- Backend logic: `AuthService.RegisterWithVerification` routes internally via `settingService.IsInvitationCodeEnabled()`. When invitation code mode is active, the `invitationCode` parameter is consumed by the existing invitation code validation. Referral processing only runs when invitation code mode is off.
- **Validation:** Add a note in admin settings UI that invitation code and referral code are mutually exclusive for the input field (both can be enabled, but invitation code takes priority on the registration form).

---

## Test Plan (Expanded - Deliberate Mode)

### Unit Tests
| Test | File | Description |
|------|------|-------------|
| GenerateCode uniqueness | `referral_service_test.go` | Generate 1000 codes, verify no duplicates |
| ProcessReferral happy path | `referral_service_test.go` | Valid code -> record created, balance updated atomically in transaction |
| ProcessReferral invalid code | `referral_service_test.go` | Unknown code -> nil error, no side effects |
| ProcessReferral self-referral | `referral_service_test.go` | inviter==invitee -> nil error |
| ProcessReferral disabled | `referral_service_test.go` | setting off -> no-op |
| ProcessReferral zero rewards | `referral_service_test.go` | reward amounts = 0 -> record created, no balance change |
| ProcessReferral tx failure | `referral_service_test.go` | Simulated commit failure -> no partial state (no orphan records, no balance changes) |
| ProcessReferral cache invalidation | `referral_service_test.go` | After successful commit -> `InvalidateUserBalance` + `InvalidateAuthCacheByUserID` called for both inviter and invitee |
| RegisterWithVerification + referral | `auth_service_register_test.go` | Full registration with referral code (invitation_code_enabled=false, referral_enabled=true) |
| Settings public exposure | `setting_service_public_test.go` | `referral_enabled` in public settings |

### Integration Tests
| Test | Description |
|------|-------------|
| Full registration flow | Register -> referral_code generated -> register with code -> rewards distributed |
| Migration idempotency | Run migration twice, no errors |
| Concurrent registrations | Two users register simultaneously with same inviter code |

### E2E (Manual)
| Test | Description |
|------|-------------|
| Register with referral URL | `/register?ref=CODE` pre-fills and works |
| ReferralView displays correctly | Code, copy, list all render |
| Admin settings toggle | Enable/disable referral, verify frontend reacts |
| Invitation code priority | When both enabled, invitation code UI shown |

### Observability
- Log format: `[Referral] action=process inviter=%d invitee=%d code=%s reward_inviter=%.2f reward_invitee=%.2f`
- Error logging: `[Referral] action=process_failed invitee=%d code=%s error=%v`

---

## Production Safety Checklist

1. **Migration safety**: `ADD COLUMN NULL` is metadata-only in PostgreSQL (no table rewrite). `CREATE TABLE IF NOT EXISTS` is idempotent. Safe for online execution.
2. **Rollback plan**:
   - Settings: Set `referral_enabled=false` via admin UI (instant, no deploy needed)
   - Code: Revert to previous binary (referral_code column stays but is unused)
   - Data: `DROP TABLE IF EXISTS user_referrals; ALTER TABLE users DROP COLUMN IF EXISTS referral_code;` (only if needed, data loss)
3. **Deploy order**: Backend first (migration runs on startup), then frontend (or single binary).
4. **Feature flag**: `referral_enabled` defaults to `false`. Admin must explicitly enable after verifying on test environment.
5. **Backward compatibility**: No existing API contracts change. `RegisterRequest` gets a new optional `referral_code` field; old clients omit it.

---

## File Change Summary

| Layer | New Files | Modified Files |
|-------|-----------|----------------|
| Migration | `091_add_referral_system.sql` | -- |
| Ent Schema | `ent/schema/user_referral.go` | `ent/schema/user.go` |
| Ent Generated | -- | `ent/` (auto-generated via `go generate`) |
| Service | `service/referral_service.go`, `service/referral_service_test.go` | `service/auth_service.go`, `service/setting_service.go`, `service/settings_view.go` |
| Handler | `handler/referral_handler.go` | `handler/auth_handler.go`, `handler/wire.go` |
| Handler DTO | -- | `handler/dto/` (add referral DTOs + `referral_code` to RegisterRequest) |
| Routes | -- | `server/routes/user.go` |
| Wire | -- | `cmd/server/wire_gen.go` |
| Frontend API | `api/referral.ts` | `api/admin/settings.ts`, `types/index.ts` |
| Frontend Views | `views/user/ReferralView.vue` | `views/auth/RegisterView.vue`, `views/admin/SettingsView.vue`, `views/admin/UsersView.vue` |
| Frontend Layout | -- | `components/layout/AppSidebar.vue`, `router/index.ts` |
| Frontend i18n | -- | `i18n/locales/zh.ts`, `i18n/locales/en.ts` |
| Frontend Store | -- | `stores/app.ts` |

**Total: ~8 new files, ~16 modified files**
**Estimated complexity: MEDIUM-HIGH**

---

## Revision Log

### v2 (2026-04-08) -- Architect/Critic Review Integration

1. **CRITICAL: Migration numbering 081 -> 091** -- Upstream has migrations through `090_drop_sora.sql`. Changed all references from `081_add_referral_system.sql` to `091_add_referral_system.sql`. Added note to verify `max(local, upstream) + 1` at implementation time.

2. **CRITICAL: Transaction wrapping for reward distribution** -- `ProcessRegistrationReferral` now uses a database transaction (following `PromoService.ApplyPromoCode` pattern) to atomically wrap: INSERT user_referrals + UPDATE inviter balance + UPDATE invitee balance. Removed "idempotent on retry" claim from Pre-mortem Scenario 2; replaced with "transaction guarantees atomicity." Added transaction failure test case.

3. **MAJOR: RegisterWithVerification signature unchanged** -- Removed the suggestion to add a new `referralCode` parameter. Instead, `AuthService` routes internally: `settingService.IsInvitationCodeEnabled()` true -> existing invitation code logic; false + `referral_enabled` true -> `invitationCode` param treated as referral code and passed to `ReferralService`.

4. **MAJOR: ReferralService cache dependencies** -- Added `billingCacheService *BillingCacheService` and `authCacheInvalidator APIKeyAuthCacheInvalidator` to `ReferralService` struct. `ProcessRegistrationReferral` calls `InvalidateUserBalance` and `InvalidateAuthCacheByUserID` for both inviter and invitee post-commit, matching the PromoService pattern. Updated wire_gen.go notes to include these dependencies.

5. **MINOR: GenerateReferralCode call location clarified** -- Called unconditionally after `userRepo.Create`, separate from `ProcessRegistrationReferral`. Every new user gets a referral code regardless of how they registered.

6. **MINOR: Frontend RegisterRequest uses independent `referral_code` field** -- Added a separate `referral_code` field to both frontend `RegisterRequest` type and backend DTO. Does not reuse `invitation_code` on the frontend side.

7. **Test plan expanded** -- Added: `TestReferralService_ProcessReferral_TransactionFailure` (atomicity verification) and `TestReferralService_ProcessReferral_CacheInvalidation` (post-commit cache invalidation for both users).
