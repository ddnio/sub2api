# Upstream Sync Phase 2 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Safely continue upstream sync after payment-b2 by porting high-value runtime and protocol fixes in small, reviewable batches.

**Architecture:** Do not merge `upstream/main` wholesale. Work from `origin/main` in an isolated worktree, classify upstream changes by risk, then port one narrow slice at a time. Payment-b2 is treated as a protected baseline: only payment residual fixes with clear evidence should touch payment code.

**Tech Stack:** Go, Vue 3, TypeScript, Vitest, PostgreSQL migrations, Git worktrees, Kimi review through `codex-buddy`.

---

## Current Baseline

- Base branch: `origin/main` at `010a662e docs(payment-b2): record order zero amount hotfix`
- Work branch: `feature/upstream-sync-2026-05-phase2`
- Worktree: `.claude/worktrees/upstream-sync-2026-05-phase2`
- Upstream ref inspected: `upstream/main` at `48912014 chore: sync VERSION to 0.1.121 [skip ci]`
- Known main-worktree noise: untracked `.pnpm-store/`; do not add it.
- Upstream SHAs are pinned to the inspected ref. If `upstream/main` moves, do not silently advance scope; amend this plan and review it again.

## Guardrails

- Do not merge `upstream/main` directly.
- Do not deploy production during this phase unless a separate deployment gate is approved.
- Do not rewrite payment-b2 unless the upstream fix is proven missing and low-risk.
- Any produced plan, code change, or deployment-impacting doc change must be reviewed by Kimi before merge.
- Any deployment-impacting change must update `docs/engineering/upstream-sync-2026-05-phase2.md`.
- Before porting any upstream fix, prove it is missing locally by code inspection, a failing/absent test, or a reproducible behavior gap.
- Any slice exceeding about 300 changed lines or touching more than 5 files must be split or explicitly justified in the tracking doc before Kimi review.
- Every slice must include a config audit: new `config.yaml` keys, env vars, setting names, and frontend `localStorage` keys.

## Task 0: Baseline Verification

**Files:**
- No code changes.

**Step 1: Confirm branch and refs**

Run:

```bash
git status --short --branch
git rev-parse --short HEAD
git rev-parse --short origin/main
git rev-parse --short upstream/main
git worktree list
```

Expected:
- Current branch is `feature/upstream-sync-2026-05-phase2`.
- `HEAD` equals `origin/main`.
- Worktree is isolated from the root checkout.

**Step 2: Check dependency drift**

Run:

```bash
git diff --name-status origin/main..upstream/main -- backend/go.mod backend/go.sum frontend/package.json frontend/pnpm-lock.yaml
git diff -- backend/go.mod backend/go.sum frontend/package.json frontend/pnpm-lock.yaml
```

Expected:
- Any dependency required by a candidate slice is recorded in the tracking doc before code is ported.
- Dependency updates are not accepted as incidental churn; they must be tied to a specific upstream fix.

**Step 3: Check Ent and migration drift**

Run:

```bash
git diff --name-status origin/main..upstream/main -- 'backend/ent/**' 'backend/migrations/**'
git log --oneline --reverse --cherry-pick origin/main...upstream/main -- 'backend/ent/**' 'backend/migrations/**'
```

Expected:
- Any migration or Ent schema change blocks the relevant slice until classified as port / hold / skip.
- Migration changes require fresh-install and existing-DB upgrade verification before deployment can be considered.

**Step 4: Run baseline tests before new code**

Run:

```bash
cd backend
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/service -run 'Test.*Payment|Test.*Order|Test.*Provider|Test.*Refund|Test.*Fulfillment|Test.*Config'
cd ../frontend
pnpm exec vitest run src/__tests__/buttonClasses.spec.ts src/components/payment/__tests__/paymentFlow.spec.ts src/views/user/__tests__/PaymentView.spec.ts
```

Expected:
- Baseline passes or any existing failure is recorded before new changes.

## Task 1: Create Phase 2 Tracking Doc

**Files:**
- Create: `docs/engineering/upstream-sync-2026-05-phase2.md`

**Step 1: Write the tracking doc**

Include:
- upstream baseline and local baseline
- slice list
- review status
- verification status
- deployment notes
- hold list

**Step 2: Kimi review the plan and tracking doc**

Run:

```bash
node /Users/nio/.codex/skills/codex-buddy/scripts/buddy-runtime.mjs --action preflight --buddy-model kimi
```

Then send the plan and tracking doc as evidence to Kimi. Expected: Kimi either approves the slice ordering or flags missing risks. Address review feedback before Task 2.

## Task 2: Runtime Safety Slice

**Files:**
- Inspect before editing:
  - `backend/internal/httputil/**`
  - request body decoding call sites
  - scheduler snapshot/cache files touched by upstream commits
- Update tracking doc if deployment risk is found:
  - `docs/engineering/upstream-sync-2026-05-phase2.md`

**Candidate upstream commits:**
- `40feb86b fix(httputil): add decompression bomb guard and fix errcheck lint`
- PR #1990 zstd request decompression
- `8bf2a7b8 fix(scheduler): resolve SetSnapshot race conditions and remove usage throttle`
- `733627cf fix: improve sticky session scheduling`

**Step 1: Inspect candidate patches**

Run:

```bash
git show --stat --name-only 40feb86b
git show --stat --name-only 8bf2a7b8
git show --stat --name-only 733627cf
git show --color=never --unified=80 40feb86b
git show --color=never --unified=80 8bf2a7b8
git show --color=never --unified=80 733627cf
```

Expected:
- Identify whether each patch is isolated or entangled with upstream-only files.
- Read the upstream commit/PR rationale for scheduler changes. If throttle removal lacks a compensating limiter or clear safety proof, hold it.
- Verify the bug or missing guard exists locally before porting.

**Step 2: Port only isolated safety fixes**

Prefer `git cherry-pick -n <sha>` only when the patch applies cleanly and does not drag unrelated features. Otherwise manually port the minimal changed hunks.

**Step 3: Verify**

Run targeted Go tests for changed packages and scheduler tests. If request decoding changed, add or port tests covering:
- compressed body under the limit decodes correctly
- compressed body over the limit returns the expected error without high memory use
- malformed or truncated compressed body returns an error
- unsupported content encoding falls through safely

If scheduler files changed, run the relevant package under `-race`, for example:

```bash
cd backend
GOCACHE="$PWD/../../.cache/go-build" go test -race -count=1 ./internal/... -run 'Test.*Scheduler|Test.*Sticky|Test.*Snapshot'
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/service -run 'Test.*Payment|Test.*Order|Test.*Provider|Test.*Refund|Test.*Fulfillment|Test.*Config'
```

**Step 4: Kimi review**

Send the diff and test output to Kimi. Do not commit until review feedback is addressed.

**Step 5: Commit**

Commit message pattern:

```bash
git commit -m "sync(runtime): port upstream request and scheduler safety fixes"
```

## Task 3: OpenAI Responses / Codex Compatibility Slice

**Files:**
- Inspect before editing:
  - `backend/internal/apicompat/**`
  - `backend/internal/gateway/**`
  - OpenAI account scheduler and Responses transform tests
  - affected frontend settings only if an upstream fix requires UI fields
- Update tracking doc if deployment behavior changes.

**Candidate upstream commits:**
- `094e1171 fix(openai): infer previous response for item references`
- `55a7fa1e` / upstream PR #2005 passthrough strip fields
- `5e54d492 fix(lint): check type assertion error in codex transform test`
- upstream PR #2068 drop reasoning items from input
- upstream PR #2100 Codex CLI edit resend continuation
- upstream PR #2058 function tool_choice format
- upstream PR #1948 account test responses stream
- upstream PR #1960 stream keepalive downstream idle
- upstream PR #1943 responses pre-output failover
- upstream PR #1772 OpenAI test state reconciliation

**Step 1: Build a missing-fix matrix**

For each candidate, record:
- upstream SHA / PR
- touched files
- already present locally: yes/no/partial
- local evidence: code inspection / failing test / missing test
- depends on: upstream SHA or local prerequisite
- dependency risk
- API contract drift: DTO/request/response changes and frontend consumers
- config/localStorage drift
- planned action: port / hold / skip

**Step 2: Port one compatible sub-batch**

Keep each commit focused on one behavior group:
- payload normalization
- stream failover/keepalive
- tool choice / tool call ID compatibility
- account scheduler state reconciliation

**Step 3: Verify**

Run targeted tests for changed transforms/scheduler paths. If no local test exists, port the upstream test with minimal dependency changes.

If a transform or API response shape changes, compare backend DTOs/transforms against frontend `src/api/`, `src/types/`, and affected tests. If externally visible behavior changes, record a customer-facing changelog/API note requirement in the tracking doc.

**Step 4: Kimi review and commit**

Review the diff and test output with Kimi before each commit.

## Task 4: Anthropic / Claude Compatibility Slice

**Files:**
- Inspect before editing:
  - Anthropic request/response conversion
  - gateway mimicry code
  - cache usage and SSE error handling tests
- Update tracking doc if default behavior or config changes.

**Candidate upstream commits:**
- `73b87299 feat: 添加 Anthropic 缓存 TTL 注入开关`
- upstream PR #2066 Anthropic stream EOF failover
- upstream PR #1996 Claude Code empty Read.pages
- upstream PR #1970 cache token usage semantics
- `496469ac fix(gateway): skip body mimicry for real Claude Code clients to restore prompt caching`

**Steps:**
1. Inspect candidate patches.
2. Verify the bug or missing behavior exists locally before porting.
3. Classify config changes separately from pure bug fixes.
4. Audit new config keys, env vars, setting names, and error/log patterns.
5. Port pure compatibility fixes first.
6. Verify targeted Go tests and any upstream parity tests.
7. If a response shape or SSE behavior changes, record API/consumer impact in the tracking doc.
8. Kimi review before commit.

## Task 5: Admin / Frontend Low-Risk UX Slice

**Candidate upstream commits:**
- PR #2118 table pagination localStorage persistence
- account bulk edit scope / compact settings
- settings contract field tests
- table defaults and admin filtering fixes

**Rules:**
- Do not mix with backend protocol fixes.
- If a backend protocol slice changes a response shape consumed by admin/frontend tables, settings, or payment UI, do the frontend adjustment after that backend slice or explicitly include it in the same reviewed slice.
- Do not accept broad profile/auth redesign changes as part of this slice.
- Run targeted Vitest specs, `vue-tsc --noEmit`, and `pnpm build` if shared types, routing, or bundling-relevant imports changed.

## Task 6: Payment Residual Audit Only

**Candidate upstream commits:**
- `c1b52615 fix(payment): allow Stripe payment pages to bypass router auth guard`
- `8f28a834 fix(payment): 同时启用易支付和 Stripe 时显示 Stripe 按钮`
- PR #1764 wxpay pubkey hardening
- PR #1655 payment fee multiplier

**Rules:**
- First prove whether payment-b2 already contains equivalent behavior.
- Do not port fee/multiplier semantics unless product behavior is explicitly approved.
- Any migration or payment table change requires fresh-install and existing-DB verification.
- Any deploy note goes into `docs/engineering/upstream-sync-2026-05-phase2.md` and, if payment-specific, `docs/engineering/payment-b2-deploy.md`.
- For any payment behavior change, follow the backup and preflight SQL checklist format from `docs/engineering/payment-b2-deploy.md`.

## Task 7: Integration Smoke Gate

**Files:**
- Update: `docs/engineering/upstream-sync-2026-05-phase2.md`

After all accepted slices are committed and before PR merge:

1. Run the final targeted backend and frontend test set.
2. Start the server against a safe local or test config.
3. Hit affected routes with curl/API smoke tests.
4. If migrations changed, verify both fresh install and existing DB upgrade on a non-production database copy.
5. Record smoke-test commands and results in the tracking doc.
6. Kimi review the final diff, verification output, and deployment notes.

## Task 8: Hold List For Later Phases

Hold unless separately approved:
- auth identity foundation / pending OAuth flow
- WeChat OAuth auth flow outside payment
- channel monitor / channel insights
- affiliate invite rebate system
- Vertex service account
- OpenAI Fast/Flex policy
- license / CLA workflow changes
- sponsor/readme churn

## Final Gate For Each Slice

Before saying a slice is complete:
- `git diff --stat` reviewed
- targeted tests run and recorded
- Kimi review result recorded
- deployment-impacting notes documented
- rollback notes documented when the slice affects migrations, config, service startup, routing, or external API behavior
- one focused commit created
- no unrelated files added, especially `.pnpm-store/`
