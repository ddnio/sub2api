# Payment Upstream Backport Audit Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Finish the Payment B-2 upstream sync by auditing and selectively backporting only payment-related upstream changes that are still missing from `worktree-payment-b2`.

**Architecture:** Do not merge `upstream/main` wholesale. First generate a complete upstream-delta matrix from `HEAD..upstream/main`, classify every payment-looking commit by relevance, dependency, migration risk, business-semantics risk, and whether it is already implemented locally. Only after that matrix is reviewed should small, dependency-safe batches be backported and verified.

**Tech Stack:** Vue 3, TypeScript, Vitest, Go, PostgreSQL, Docker deployment scripts, `git cherry-pick` only for isolated commits, manual patching for mixed commits.

---

## Current Baseline

- Branch: `worktree-payment-b2`
- Latest pushed commits:
  - `2cafd049 fix(payment-b2): restore payment button styles`
  - `052cb44d docs(payment-b2): record test deploy verification`
- Test env currently verified:
  - `sub2api-test` healthy on `127.0.0.1:8081`
  - `/purchase` recharge and subscription both create wxpay QR orders
  - Test DB backup before last deploy: `/home/nio/backups/sub2api_test_pre_payment_button_styles_20260501-135458.sql`
- Production must not be touched during this plan.

---

## Kimi Review Incorporated

Kimi reviewed the first draft and flagged real blockers. This version incorporates those findings:

- The first candidate list was too narrow; upstream contains more payment-looking commits than the obvious 30.
- Some commits were listed but not assigned to a batch: `fd0c9a13`, `61a008f7`, `c3cb0280`.
- `f35e9675` was duplicated across batches.
- Auth-dependent commits such as `09351e94` and broad hardening commits such as `c229f33e` require dependency analysis before any patch.
- Migration, webhook, fee-rate, provider-config, and production rollback gates were too weak.

---

## Seed Candidate Commits

This seed list is only input for the audit matrix. It is not a backport order and is not complete.

```text
4aa0070e fix: Stripe payment type matching in load balancer
5bae3b05 fix(payment): audit fixes for alipay/wxpay/stripe payment providers
c738cfec fix(payment): critical audit fixes for security, idempotency and correctness
75e1b40f fix(payment): ack unknown-order webhooks with 2xx to stop provider retries
d5dac84e test(payment): cover ErrOrderNotFound sentinel contract
c1b52615 fix(payment): allow Stripe payment pages to bypass router auth guard
fd0c9a13 fix(payment): store provider config as plaintext JSON with legacy ciphertext fallback
61a008f7 chore(payment): mark legacy AES ciphertext fallback as deprecated
c3cb0280 fix(payment): alipay redirect-only flow, H5 detection and popup sizing
235f7108 feat(payment): redact provider secrets in admin config API
79192cf6 feat(payment): harden wxpay config validation with structured errors
561405ab feat(payment): add payment order provider snapshots
c0b24aef feat: snapshot payment provider keys on orders
bdcd3d87 fix: resolve unique legacy payment providers
b3098221 fix: tighten legacy payment provider resolution
6f00efa3 fix: support legacy payment method aliases
7c7924e9 fix: guard payment fulfillment provider mismatch
e3f69e02 fix: tighten webhook provider resolution
9742796e fix: retire public payment verify and backfill trade no
1ffebbb5 fix(migrations): keep auth identity and payment upgrades safe
60a4b931 feat(payment): balance recharge multiplier and refund amount separation
98140f6c feat(payment): add recharge fee rate setting and fix provider card UI
e761d38f fix(payment): integrate recharge fee rate in order flow and fix UI display
d149dbc9 fix(payment): enhance fee rate input validation and UI
3053c56c fix(payment): show full amount breakdown on payment result page
342dbd2e fix(payment): use original recharge amount in product name, not pay_amount
b51bc7ee feat: wire payment return url payloads
40d4e167 feat(payment): i18n payment error codes and label localization
9bebf1c1 feat: resolve payment results by resume token
7ef7fd19 fix: restore wechat payment oauth and jsapi flow
f83fd59d Refine payment UX for wallet flows
16be82b9 fix payment visible methods and resume recovery
55e8dd55 Tighten WeChat payment resume flow
c297d011 Keep pending payment results in processing state
a27a7add fix payment resume result consistency
09351e94 fix auth completion and payment resume hardening
07f23aaa fix wxpay config contract and h5 scene info
65d3bd72 frontend: normalize payment error presentation
29caf851 fix(frontend): stabilize wechat payment resume recovery
dd314c41 fix(payment): restore public resume and result flows
c229f33e fix(review): harden payment, oauth, and migration paths
906802ab Fix mobile payment launch detection
1aab084e fix(payment): restore upgrade-safe payment flows
f35e9675 fix payment qr fallback and admin guidance
8f28a834 fix(payment): 同时启用易支付和 Stripe 时显示 Stripe 按钮
```

Some may already be present through `623dda62` or later local B-2 patches. Some may be blocked by upstream auth identity, settings DTO, migration, or unrelated module changes.

---

## Non-Goals

- Do not merge unrelated upstream modules such as affiliate, channel monitor, profile redesign, available channels, notification, pricing, or gateway model changes.
- Do not redesign the `/purchase` UI in this pass.
- Do not deploy production in this pass.
- Task 9 is only a future production decision gate. This plan does not run production deployment.
- Do not change test/prod payment provider credentials except when explicitly asked and after backup.

---

## Task 0: Baseline And Scope Inventory

**Files:**
- Update later through Task 1: `docs/engineering/payment-b2-upstream-audit.md`

**Step 1: Confirm repo baseline**

Run:

```bash
git fetch upstream
git status --short
git log --oneline -5
git rev-parse --short HEAD
git rev-parse --short upstream/main
```

Expected: clean or only this plan doc modified; upstream ref recorded.

**Step 2: Verify current branch before backport**

Run:

```bash
cd backend
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/service -run 'Test.*Payment|Test.*Wechat|Test.*WeChat|Test.*Provider|Test.*Order|Test.*Refund|Test.*Fulfillment|Test.*Config'
cd ../frontend
pnpm exec vitest run src/__tests__/buttonClasses.spec.ts src/components/payment/__tests__/paymentFlow.spec.ts src/views/user/__tests__/PaymentView.spec.ts
pnpm exec vue-tsc --noEmit
```

Expected: current branch baseline passes before new changes.

**Step 3: Inventory local fork-specific payment patches**

Record in the audit doc:

- button style restoration and `btn-*` test from `2cafd049`
- sticky payment action bar and inline error display
- `formatAmount()` guard
- localStorage try/catch guards
- post-order scroll to QR
- fork-specific wxpay QR/native-pay behavior and any JSAPI/OAuth stubs or config additions already present
- local payment migrations already applied

These patches are preservation constraints. A backport may change them only when explicitly justified in the matrix.

**Step 4: Inventory migration delta**

Run:

```bash
git diff --name-status HEAD..upstream/main -- backend/internal/migrations backend/ent/schema backend/ent/migrate
git log --oneline --reverse --cherry-pick HEAD...upstream/main -- backend/internal/migrations backend/ent/schema backend/ent/migrate
```

Expected: every payment-adjacent migration is listed in Task 1. Unknown migration implications block production readiness.

---

## Task 1: Produce A Full Commit Classification Matrix

**Files:**
- Create or update: `docs/engineering/payment-b2-upstream-audit.md`

**Step 1: Generate raw candidate lists**

Run path-based scan:

```bash
git fetch upstream
git log --oneline --reverse --cherry-pick HEAD...upstream/main -- \
  backend/internal/migrations \
  backend/ent/schema \
  backend/ent/migrate \
  frontend/src/views/user/PaymentView.vue \
  frontend/src/views/user/PaymentQRCodeView.vue \
  frontend/src/views/user/PaymentResultView.vue \
  frontend/src/views/user/StripePaymentView.vue \
  frontend/src/views/user/StripePopupView.vue \
  frontend/src/views/user/__tests__ \
  frontend/src/components/payment \
  frontend/src/api/payment.ts \
  frontend/src/types/payment.ts \
  frontend/src/router \
  frontend/src/i18n \
  frontend/src/locales \
  backend/internal/payment \
  backend/internal/server/routes/payment.go \
  backend/internal/server/routes \
  backend/internal/handler \
  backend/internal/service \
  backend/internal/config \
  backend/internal/setup \
  backend/internal/domain \
  docs/engineering/payment-b2-deploy.md
```

Run subject-based scan:

```bash
git log --oneline --reverse --cherry-pick HEAD...upstream/main --grep='payment\|wxpay\|wechat\|stripe\|alipay\|provider\|webhook\|resume\|order'
```

Expected: broad candidate set, including mixed commits. This is raw input only.

**Step 2: Inspect each candidate commit**

For each candidate:

```bash
git show --stat --name-only <sha>
git show --color=never --unified=80 <sha> -- <payment-related-files>
```

Classify each as:

- `MUST`: payment correctness/security/compatibility fix with safe dependency profile.
- `SHOULD`: payment UX/error handling or operational clarity.
- `MAYBE`: payment-adjacent but dependency profile is unclear.
- `BLOCKED`: payment-related but requires upstream auth identity, profile, channel, affiliate, or large non-payment architecture sync.
- `HOLD`: payment-related feature line that changes business semantics and needs product decision first, for example recharge fee-rate or balance multiplier changes.
- `SKIP`: unrelated or already covered locally.
- `PARTIAL`: already implemented in this branch but needs verification.

**Step 3: Record dependency notes**

For every `MUST` / `SHOULD` / `MAYBE` / `BLOCKED` / `HOLD` commit, record:

- upstream files touched
- local files already changed
- dependency on migrations, auth identity, settings DTO, i18n, route changes, or ent schema
- exact predecessor/successor commits if part of a feature line
- whether the commit touches more than 30% non-payment files
- whether cherry-pick is safe, manual payment-only patch is required, or implementation is blocked

**Step 4: Define skip/backport rules**

Record these rules in the audit doc:

- Do not cherry-pick a commit that requires unrelated auth/profile/channel/affiliate infrastructure; mark `BLOCKED`.
- Do not introduce a new DB migration until schema impact, test DB status, production prerequisite status, and rollback notes are documented.
- Do not backport broad commits if a smaller manual payment-only patch is safer.
- Do not remove local fork-specific payment fixes unless upstream supersedes them and tests prove the replacement.
- If the upstream commit is already present through `623dda62` or later local B-2 patches, mark `PARTIAL` or `SKIP` with proof.

**Step 5: Ask Kimi to review the matrix**

Run Kimi with the matrix and raw command outputs. Ask:

```text
Review this upstream payment backport classification. Identify commits that are misclassified, hidden dependencies, and payment-related upstream commits missing from the matrix. Focus on avoiding unrelated module scope creep.
```

Expected: Kimi either confirms or flags missing/misclassified commits. Update the matrix before coding.

**Step 6: Commit the audit doc**

```bash
git add docs/engineering/payment-b2-upstream-audit.md
git commit -m "docs(payment-b2): audit upstream payment backports"
git push origin worktree-payment-b2
```

---

## Task 2: Backport Low-Dependency Backend Safety Batch

**Files:**
- Likely modify:
  - `backend/internal/payment/*`
  - `backend/internal/service/*payment*`
  - `backend/internal/server/routes/payment.go`
  - selected payment handler files
- Tests:
  - existing payment/service tests
  - add tests only when an upstream fix has no local coverage

**Scope candidates after Task 1 classification:**

```text
4aa0070e Stripe payment type matching in load balancer
5bae3b05 audit fixes for alipay/wxpay/stripe payment providers
c738cfec critical audit fixes for security, idempotency and correctness
75e1b40f ack unknown-order webhooks with 2xx to stop provider retries
d5dac84e cover ErrOrderNotFound sentinel contract
7c7924e9 guard payment fulfillment provider mismatch
e3f69e02 tighten webhook provider resolution
6f00efa3 support legacy payment method aliases
9742796e retire public payment verify and backfill trade no
```

Only include a candidate here if Task 1 proves it has no unresolved migration/auth dependency.

**Step 1: Apply one commit or one logical patch at a time**

Prefer `git cherry-pick -n <sha>` only for isolated payment commits. If the commit contains unrelated changes, use manual patching into exact payment files only. If a safe payment-only patch is not obvious, stop and mark the commit `BLOCKED` in the audit doc.

**Step 2: Resolve conflicts conservatively**

Keep fork-specific payment provider configuration and current tested wxpay flow unless upstream clearly fixes a correctness bug.

**Step 3: Run backend tests**

```bash
cd backend
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/service -run 'Test.*Payment|Test.*Wechat|Test.*WeChat|Test.*Provider|Test.*Order|Test.*Refund|Test.*Fulfillment|Test.*Config'
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/server/routes
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/handler ./internal/handler/admin
```

Expected: all selected tests pass. If sandbox blocks `httptest`, rerun with network permission.

**Step 4: Commit the batch**

```bash
git add <changed-files>
git commit -m "fix(payment-b2): backport upstream payment safety fixes"
git push origin worktree-payment-b2
```

---

## Task 3: Backport Provider Config And Migration-Sensitive Batch

**Files:**
- Likely modify:
  - `backend/internal/payment/*`
  - `backend/internal/service/*payment*`
  - `backend/internal/handler/*payment*`
  - `backend/internal/migrations/*`
  - `backend/ent/schema/*`
  - `docs/engineering/payment-b2-deploy.md`

**Scope candidates after Task 1 classification:**

```text
fd0c9a13 store provider config as plaintext JSON with legacy ciphertext fallback
61a008f7 mark legacy AES ciphertext fallback as deprecated
79192cf6 harden wxpay config validation with structured errors
235f7108 redact provider secrets in admin config API
c0b24aef snapshot payment provider keys on orders
561405ab add payment order provider snapshots
bdcd3d87 resolve unique legacy payment providers
b3098221 tighten legacy payment provider resolution
1ffebbb5 keep auth identity and payment upgrades safe
```

**Step 1: Migration gate**

Before applying any candidate in this batch:

```bash
git diff --name-status HEAD..upstream/main -- backend/internal/migrations backend/ent/schema backend/ent/migrate
```

For every migration involved, document:

- whether the test DB already has the equivalent schema
- whether production likely has the prerequisite migration
- whether rollback is data-safe
- whether the migration is payment-only or mixed with auth identity

If a migration is mixed with auth identity or unrelated modules, mark it `BLOCKED` unless a payment-only manual migration can be proven safe.

**Step 2: Apply safe patches only**

Manual patching is preferred. Do not blindly cherry-pick migration-sensitive commits.

**Step 3: Verify provider config compatibility**

Run or add tests for:

- existing encrypted provider config still readable if legacy fallback is retained
- plaintext provider config stores and reloads correctly
- provider secrets are redacted from admin responses
- existing test environment wxpay provider still works after deploy

**Step 4: Run tests**

Use Task 2 backend tests plus any provider-config tests added in this batch.

**Step 5: Commit**

```bash
git add <changed-files>
git commit -m "fix(payment-b2): align upstream provider config safety"
git push origin worktree-payment-b2
```

---

## Task 4: Decide Fee-Rate And Balance Multiplier Feature Line

**Files:**
- Audit first:
  - payment service/order files
  - admin settings/provider UI files
  - payment result frontend files
  - migrations/settings DTOs

**Scope candidates:**

```text
60a4b931 balance recharge multiplier and refund amount separation
98140f6c add recharge fee rate setting and fix provider card UI
e761d38f integrate recharge fee rate in order flow and fix UI display
d149dbc9 enhance fee rate input validation and UI
3053c56c show full amount breakdown on payment result page
342dbd2e use original recharge amount in product name, not pay_amount
```

**Decision gate:**

This batch changes business semantics around recharge amount, pay amount, refund amount, and product naming. Do not implement it as a routine bugfix.

Classify the whole feature line as:

- `MUST` only if current branch already includes partial fee-rate/multiplier behavior and upstream fixes correctness.
- `HOLD` if it is a new product/business behavior not required for production.
- `PARTIAL` if current branch already implements the needed pieces.

If marked `HOLD`, record the reason in `docs/engineering/payment-b2-upstream-audit.md` and skip implementation for this pass.

---

## Task 5: Backport Frontend Payment Recovery And Launch Batch

**Files:**
- Likely modify:
  - `frontend/src/components/payment/paymentFlow.ts`
  - `frontend/src/components/payment/__tests__/paymentFlow.spec.ts`
  - `frontend/src/views/user/PaymentView.vue`
  - `frontend/src/views/user/PaymentResultView.vue`
  - `frontend/src/views/user/PaymentQRCodeView.vue`
  - `frontend/src/views/user/StripePaymentView.vue`
  - `frontend/src/views/user/StripePopupView.vue`
  - `frontend/src/api/payment.ts`
  - `frontend/src/types/payment.ts`
  - `frontend/src/router/index.ts`

**Scope candidates after Task 1 classification:**

```text
b51bc7ee wire payment return url payloads
9bebf1c1 resolve payment results by resume token
7ef7fd19 restore wechat payment oauth and jsapi flow
16be82b9 payment visible methods and resume recovery
55e8dd55 Tighten WeChat payment resume flow
c297d011 Keep pending payment results in processing state
a27a7add payment resume result consistency
29caf851 stabilize wechat payment resume recovery
dd314c41 restore public resume and result flows
906802ab mobile payment launch detection
1aab084e restore upgrade-safe payment flows
8f28a834 show Stripe button when easypay and Stripe are both enabled
c1b52615 allow Stripe payment pages to bypass router auth guard
```

Candidates that may be `BLOCKED` unless Task 1 proves dependencies are isolated:

```text
09351e94 auth completion and payment resume hardening
c229f33e harden payment, oauth, and migration paths
```

**Step 1: Compare local implementation before patching**

For each candidate file:

```bash
git diff HEAD..upstream/main -- <file>
```

Do not overwrite local fixes from:

- `2cafd049` button class restoration
- sticky payment action bar
- inline error display inside sticky action area
- `formatAmount()`
- localStorage try/catch
- post-order scroll to QR

**Step 2: Apply minimal upstream-compatible patches**

Backport missing flow logic and tests. Preserve local UX fixes unless Kimi or tests show a conflict.

**Step 3: Run frontend payment tests**

```bash
cd frontend
pnpm exec vitest run \
  src/__tests__/buttonClasses.spec.ts \
  src/components/payment/__tests__/paymentFlow.spec.ts \
  src/components/payment/__tests__/PaymentStatusPanel.spec.ts \
  src/views/user/__tests__/PaymentView.spec.ts \
  src/views/user/__tests__/PaymentResultView.spec.ts \
  src/views/user/__tests__/paymentUx.spec.ts \
  src/views/user/__tests__/paymentWechatResume.spec.ts \
  src/utils/__tests__/device.spec.ts
pnpm exec vue-tsc --noEmit
pnpm build
```

Expected: tests, type check, and build pass. Existing Vite chunk warnings are acceptable.

**Step 4: Commit the batch**

```bash
git add <changed-files>
git commit -m "fix(payment-b2): backport upstream payment recovery flows"
git push origin worktree-payment-b2
```

---

## Task 6: Backport Payment Error And Admin Guidance Batch

**Files:**
- Likely modify:
  - `frontend/src/views/user/PaymentView.vue`
  - `frontend/src/views/user/PaymentResultView.vue`
  - `frontend/src/api/payment.ts`
  - i18n files, if upstream messages are missing locally
  - `docs/engineering/payment-b2-deploy.md`
  - selected admin payment views only if directly payment-provider related

**Scope candidates:**

```text
40d4e167 i18n payment error codes and label localization
65d3bd72 normalize payment error presentation
f35e9675 payment qr fallback and admin guidance
07f23aaa wxpay config contract and h5 scene info
c3cb0280 alipay redirect-only flow, H5 detection and popup sizing
```

**Step 1: Preserve local inline error behavior**

Upstream `65d3bd72` removed some inline error blocks. Our local sticky inline error display is intentional because it helps users understand why a click did not progress. Do not remove it without a better replacement.

**Step 2: Backport message mappings and docs**

Add missing i18n keys or API error mappings needed by upstream payment flows.

**Step 3: Verify docs**

Read `docs/engineering/payment-b2-deploy.md` after edits and confirm:

- test and production commands are separate
- provider config fields match current API
- no secrets or passwords are documented
- rollback/backup language is accurate

**Step 4: Run tests**

Use the frontend and backend payment test commands from Tasks 2 and 5.

**Step 5: Commit**

```bash
git add <changed-files>
git commit -m "fix(payment-b2): align upstream payment error handling"
git push origin worktree-payment-b2
```

---

## Task 7: Full Local Verification

**Step 1: Run targeted frontend verification**

```bash
cd frontend
pnpm exec vitest run \
  src/__tests__/buttonClasses.spec.ts \
  src/components/payment/__tests__/paymentFlow.spec.ts \
  src/components/payment/__tests__/PaymentStatusPanel.spec.ts \
  src/views/user/__tests__/PaymentView.spec.ts \
  src/views/user/__tests__/PaymentResultView.spec.ts \
  src/views/user/__tests__/paymentUx.spec.ts \
  src/views/user/__tests__/paymentWechatResume.spec.ts \
  src/utils/__tests__/device.spec.ts
pnpm exec vue-tsc --noEmit
pnpm build
```

**Step 2: Run backend targeted and full verification**

```bash
cd backend
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/service -run 'Test.*Payment|Test.*Wechat|Test.*WeChat|Test.*Provider|Test.*Order|Test.*Refund|Test.*Fulfillment|Test.*Config'
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/server/routes
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/handler ./internal/handler/admin
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./...
```

If full `go test ./...` fails for unrelated existing issues, document exact failures and rerun the payment-relevant subset. Do not hide failures.

---

## Task 8: Test Environment Deploy Verification

**Files:**
- Update: `docs/engineering/payment-b2-frontend-handoff.md`

**Step 1: Backup test DB**

```bash
ssh nio@108.160.133.141 'mkdir -p /home/nio/backups && docker exec sub2api-postgres pg_dump -U sub2api -d sub2api_test > /home/nio/backups/sub2api_test_pre_upstream_payment_backport_$(date +%Y%m%d-%H%M%S).sql && ls -lh /home/nio/backups/sub2api_test_pre_upstream_payment_backport_*.sql | tail -1'
```

Expected: backup path and size printed.

**Step 2: Deploy only test**

```bash
ssh nio@108.160.133.141 'cd /data/service/sub2api && git fetch origin && git checkout worktree-payment-b2 && git pull --ff-only origin worktree-payment-b2 && bash deploy/deploy-server.sh test'
```

**Step 3: Health check**

```bash
ssh nio@108.160.133.141 'docker ps --filter name=sub2api-test --format "{{.Names}} {{.Status}} {{.Ports}}"; curl -fsS http://127.0.0.1:8081/health'
ssh nio@108.160.133.141 'docker ps --filter name=sub2api-prod --format "{{.Names}} {{.Status}} {{.Ports}}"'
```

Expected:

- `sub2api-test` healthy on `127.0.0.1:8081`
- `sub2api-prod` remains healthy on `127.0.0.1:8080`

**Step 4: Browser verification**

Use `agent-browser` on `https://router-test.nanafox.com`.

Verify:

- `/purchase` loads latest `PaymentView-*.js`
- recharge tab:
  - amount 10
  - wxpay selected
  - confirm button styled and enabled
  - click creates QR canvas `220x220`
- subscription tab:
  - all enabled plans visible
  - select a plan
  - confirm button styled and enabled
  - click creates QR canvas `220x220`
- payment result/resume:
  - pending orders remain in processing state rather than false failure
  - cancelling or expired orders display expected result state
- `/orders` shows newly created orders
- admin orders page can see latest test orders

**Step 5: DB verification**

```bash
ssh nio@108.160.133.141 "docker exec sub2api-postgres psql -U sub2api -d sub2api_test -c \"SELECT id,user_id,order_type,amount,status,payment_type,out_trade_no,qr_code <> '' AS has_qr,created_at FROM payment_orders ORDER BY created_at DESC LIMIT 10;\""
```

Expected: balance and subscription test orders exist with `has_qr = true`.

**Step 6: Webhook/result compatibility check**

If a batch included webhook/provider-resolution changes, run the webhook tests locally and perform a test-environment API-level check that unknown/stale order handling does not produce provider retry-prone 5xx responses. If a real provider callback cannot be safely simulated, record the limitation and the unit/integration evidence used instead.

**Step 7: Record results**

Update `docs/engineering/payment-b2-frontend-handoff.md` with:

- commit deployed
- backup path
- health status
- browser verification results
- residual risks

Commit:

```bash
git add docs/engineering/payment-b2-frontend-handoff.md
git commit -m "docs(payment-b2): record upstream backport test deploy"
git push origin worktree-payment-b2
```

---

## Task 9: Kimi Final Review Before Production Decision

**Step 1: Prepare evidence**

Gather:

```bash
git log --oneline origin/main..HEAD
git diff --stat origin/main..HEAD
git diff --name-only origin/main..HEAD
```

Include test outputs, audit matrix, and test-environment browser verification notes.

**Step 2: Ask Kimi**

```text
Review the final Payment B-2 upstream backport branch before production decision.
Focus on:
1. payment-related upstream commits still missing,
2. accidental unrelated scope creep,
3. production deployment blockers,
4. tests or manual verification still missing.
Return findings by severity.
```

**Step 3: Resolve findings**

- P0/P1: fix before production.
- P2: decide case by case.
- Info: record in handoff if relevant.

---

## Task 10: Documentation Archive

**Files:**
- Update: `docs/engineering/payment-b2-upstream-audit.md`
- Update: `docs/engineering/payment-b2-frontend-handoff.md`

Record:

- upstream payment commits synced
- upstream payment commits skipped and why
- upstream payment commits blocked and dependency required
- `HOLD` feature lines and product decision needed
- fork-specific payment patches intentionally retained
- final tests run and their outputs
- residual risks before production

Commit:

```bash
git add docs/engineering/payment-b2-upstream-audit.md docs/engineering/payment-b2-frontend-handoff.md
git commit -m "docs(payment-b2): archive upstream payment backport decisions"
git push origin worktree-payment-b2
```

---

## Future Production Readiness Gate

Production deployment is allowed only if all conditions are true:

- Task 1 audit is committed and Kimi-reviewed.
- Tasks 2-6 have no open P0/P1 findings.
- Full selected verification passes locally, and any `go test ./...` failure is explicitly documented.
- Test environment deployed after final branch tip.
- Test environment browser verification passes for recharge, subscription, result/resume, and orders.
- Test DB backup exists for the final test deploy.
- Migration status and rollback notes are documented.
- Existing order data compatibility is checked against test DB.
- Production DB backup command and rollback command are prepared.
- A final change summary and residual risk list is prepared for the user.
- User explicitly approves production deployment after seeing final status.

If approved later, production deploy must start with:

```bash
ssh nio@108.160.133.141 'mkdir -p /home/nio/backups && docker exec sub2api-postgres pg_dump -U sub2api -d sub2api > /home/nio/backups/sub2api_prod_pre_payment_b2_$(date +%Y%m%d-%H%M%S).sql && ls -lh /home/nio/backups/sub2api_prod_pre_payment_b2_*.sql | tail -1'
```

Do not run production deployment as part of this plan without a separate explicit approval.
