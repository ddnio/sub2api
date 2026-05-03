# Upstream Sync Continuation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Finish the post-payment-b2 upstream sync quickly while keeping upstream-first behavior, review gates, and no speculative local rewrites.

**Architecture:** Treat `upstream/main` as the source of truth for feature behavior, but choose batch size from evidence. Low-divergence maintenance and already-proven bugfix families can be grouped. Product, migration, auth, payment, and config families stay as separate slices or `HOLD`.

**Tech Stack:** Go, Vue 3, TypeScript, Vitest, PostgreSQL migrations, Git worktrees, GitHub PRs in `ddnio/sub2api`, Kimi review through `codex-buddy`.

---

## Current Baseline

- Local base: `origin/main` at `682cee12 fix(openai): guard ws previous response inference`.
- Upstream ref inspected: `upstream/main` at `48912014 chore: sync VERSION to 0.1.121 [skip ci]`.
- Current closeout branch: `docs/upstream-sync-closeout-20260503`.
- Current worktree: `.claude/worktrees/upstream-sync-openai-ws-itemref`.
- Existing Phase 2 tracker: `docs/engineering/upstream-sync-2026-05-phase2.md`.
- Existing Phase 2 plan: `docs/plans/2026-05-02-upstream-sync-phase2.md`.
- Root checkout noise: untracked `.pnpm-store/`; do not add it.

If `upstream/main` advances, pin the new SHA in the tracker before expanding scope. If `git fetch upstream` fails because of GitHub HTTPS/HTTP2 transport errors, do not expand scope from memory; retry later or use a verified remote/API fallback.

## Current Closeout Status

- PR #31: merged; continuation review flow documented.
- PR #32: merged; continuation ledger refreshed.
- PR #33: merged; scheduler metadata keeps slim group membership snapshots.
- PR #34: merged; OpenAI WS item-reference previous-response inference has upstream guard rails.
- Test deployment: `682cee12` deployed to `sub2api-test` on `108.160.133.141`; container healthy, local/public `/health` OK, unauthenticated `/v1/models` returned 401, severe log scan had no output.
- Production deployment: `682cee12` deployed to `sub2api-prod` on `108.160.133.141`; same verification passed.
- Known CI drift: GitHub CI remains red on `internal/server/middleware/*_test.go` `service.NewAuthService` argument count plus pre-existing lint issues outside the upstream-sync slices. Do not fix these inside unrelated upstream alignment PRs.

## Non-Negotiable Rules

- No direct merge from `upstream/main`.
- No direct edits on root `main`; use a task worktree from `origin/main`.
- Every plan update, code slice, and deployment-impacting doc update gets review before merge.
- Every plan or code change must get a self-review pass before commit. Record what was checked and any residual risk in the commit message body or PR description.
- Every PR must trigger a code review pass after it is opened, even when the PR is docs-only. Record the PR review result in the PR description or tracker before merge.
- Review target is the fork repo: `ddnio/sub2api`.
- Feature behavior should follow upstream unless it conflicts with protected local product decisions.
- Prefer cherry-picking or manually porting upstream hunks over inventing local alternatives.
- Only write local adapter code when required to preserve existing fork behavior or make an upstream hunk compile.
- If an upstream hunk conflicts because local behavior already diverged, stop and classify it as `HOLD` unless the missing behavior is proven.
- Payment-b2, auth identity, affiliate, channel insights, Vertex, Fast/Flex, license/CLA, and sponsor/readme churn stay held unless separately approved.
- Each slice must record migration, Ent, config/env, settings, frontend `localStorage`, API contract, deploy, monitoring, and rollback impact.

## Fast Alignment Mode

The fork has local product work, so "large merge" is safe only when the upstream area is low-divergence and has no schema/config/product contract impact.

Use a larger batch when all of these are true:

- The candidate files are absent from the hold list and do not touch migrations, Ent generated code/schema, payment, auth identity, affiliate, channel insights, Vertex, Fast/Flex, image API, settings/env/config, or frontend API contracts.
- `git diff --name-only origin/main..upstream/main -- <area>` shows a narrow, coherent file set.
- Local behavior is missing or clearly older, not divergent.
- The batch can be verified with one focused test family plus `git diff --check`.
- Review evidence can stay small enough for Kimi PR-level review.

Keep a small slice or `HOLD` when any of these are true:

- The upstream change requires a product decision, new UI contract, new config setting, data migration, or generated Ent drift.
- The fork already has an equivalent local implementation.
- The upstream diff mixes bugfixes with debug logging, unrelated refactors, or broad feature policy.
- The only proof is "upstream has it"; no missing local behavior has been shown.

Expected practical result: group docs/tests/low-risk compatibility cleanups, but keep payment/auth/migration/config/frontend-contract changes separate.

## Review Protocol For Every Update

1. Create or update a narrow branch/worktree from current `origin/main`.
2. Inspect candidate upstream commit(s) with `git show --stat --name-only` and focused diffs.
3. Record a missing-fix matrix before editing.
4. Port only one behavior group.
5. Run targeted tests and `git diff --check`.
6. Self-review the plan/code diff before commit.
   - For plan/docs, check upstream SHA accuracy, hold-list consistency, review gates, downstream task order, and whether the text would allow unintended local rewrites.
   - For code, inspect changed control flow, tests, config impact, rollback, and whether upstream behavior was preserved.
7. Send raw diff plus test output to Kimi before commit. This is the pre-commit independent review gate.
8. Address blockers; if Kimi flags product/architecture uncertainty, move the item to `HOLD`.
9. Commit the reviewed changes and push to the fork branch.
10. Open a fork PR with explicit `--repo ddnio/sub2api`.
11. Trigger a PR code review after the PR exists. This is separate from the pre-commit Kimi review and should evaluate the PR diff as it will merge.
12. Merge only after self-review, pre-commit Kimi review, PR code review, and verification evidence are recorded. For docs-only changes, verification evidence is `git diff --check` plus the recorded review results.
13. Deploy only when the slice requires it or the user approves the deployment gate.

## Task 0: Refresh Scope Ledger

**Files:**
- Modify: `docs/engineering/upstream-sync-2026-05-phase2.md`

**Step 1: Confirm refs**

Run:

```bash
git fetch origin
git fetch upstream
git status --short --branch
git rev-parse --short origin/main
git rev-parse --short upstream/main
git log --oneline --cherry-pick --right-only origin/main...upstream/main | head -120
```

Expected:
- `origin/main` is the base for all new work.
- `upstream/main` SHA is recorded before any new candidate is accepted.

**Step 2: Build a continuation matrix**

For each remaining candidate, record:
- upstream SHA / PR
- files touched upstream
- current local state: missing / present / partial / divergent
- proof command
- planned action: port / skip / hold
- review status
- verification status

**Step 3: Review the updated ledger**

Send only the ledger diff and raw command output to Kimi. Do not start code until review returns no blockers.

## Task 1: Sticky Session Scheduling Remainder Audit (Done)

**Upstream source:**
- `733627cf fix: improve sticky session scheduling`

**Files to inspect:**
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/repository/scheduler_cache.go`
- scheduler and sticky-session tests

**Current local evidence:**
- PR #30 already fixed the snapshot-account false reject in a minimal local slice.
- Current `gateway_service.go` already has `isAccountSchedulableForSelection` and sticky wait-plan paths.
- Upstream `733627cf` is broader than the already merged local fix.

**Steps:**
1. Compare current local code with `733627cf` hunk-by-hunk.
2. Separate correctness hunks from logging-only or refactor-only hunks.
3. Write or identify a failing test for any missing correctness behavior before porting.
4. If no failing behavior is proven, record `skip/present-enough` and do not edit.
5. If a real gap is proven, port the smallest upstream hunk and tests.
6. Run scheduler, sticky-session, gateway load-aware, and payment regression tests.
7. Kimi review before PR.

**Default action:** audit first; no code until a missing behavior is proven.

**Closeout:** Completed through PR #33. The accepted upstream hunk was limited to scheduler metadata `GroupIDs` / slim `AccountGroups` retention. Debug logging, service-layer false-reject logic already fixed by PR #30, and handler sticky-binding refresh were not ported because no independent missing behavior was proven. Deployed to test and prod as part of `682cee12`.

## Task 2: OpenAI WebSocket Continuation Audit (Done)

**Upstream source:**
- `094e1171 fix(openai): infer previous response for item references`

**Files to inspect:**
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`

**Current local evidence:**
- Current local tests already cover automatic `previous_response_id` fill for `function_call_output`.
- Current local code has extensive previous-response recovery and sticky connection handling.

**Steps:**
1. Compare `094e1171` against current local WS forwarder behavior.
2. Prove whether item-reference inference is already covered locally.
3. If present, record `skip/present` in the tracker.
4. If partial, port only the upstream test and missing hunk.
5. Run targeted WS ingress/session tests.
6. Kimi review the matrix and any diff before PR.

**Default action:** likely audit/skip unless the exact upstream regression is missing locally.

**Closeout:** Completed through PR #34. Local already covered the positive item-reference inference path, but upstream guard rails were missing. The accepted hunk skips inference when full `function_call` context is present or when `function_call_output` lacks `call_id`. Unrelated upstream Fast/Flex/compact nearby changes remained out of scope. Deployed to test and prod as part of `682cee12`.

## Task 3: Anthropic Cache TTL Setting Decision

**Upstream source:**
- `73b87299 feat: 添加 Anthropic 缓存 TTL 注入开关`

**Files to inspect:**
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_tool_rewrite.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/dto/settings.go`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/api/admin/settings.ts`
- i18n files and settings tests

**Current local evidence:**
- Current local code already has account-level `cache_ttl_override_enabled` / `cache_ttl_override_target`.
- Current local gateway tool/message rewrite already injects or preserves cache-control TTL in several paths.
- Upstream `73b87299` is a cross-stack settings feature, not a small compatibility bugfix.

**Steps:**
1. Decide whether upstream's system setting adds behavior not covered by the local account-level override.
2. If the behavior is a product choice, keep it in `HOLD`.
3. If the behavior is accepted, port upstream's setting contract, backend defaults, frontend controls, i18n, and tests as one reviewed feature slice.
4. Run admin settings handler tests, gateway body-order/cache tests, frontend settings tests, `vue-tsc`, and build.
5. Kimi review before PR.

**Default action:** HOLD unless the user explicitly wants upstream's global TTL injection control.

## Task 4: Payment Residual Equivalence Audit

**Upstream sources:**
- `c1b52615 fix(payment): allow Stripe payment pages to bypass router auth guard`
- `8f28a834 fix(payment): 同时启用易支付和 Stripe 时显示 Stripe 按钮`
- held payment PRs such as wxpay pubkey hardening and fee multiplier

**Files to inspect:**
- `frontend/src/router/index.ts`
- `frontend/src/components/payment/paymentFlow.ts`
- `frontend/src/views/user/PaymentView.vue`
- `frontend/src/components/payment/__tests__/paymentFlow.spec.ts`
- payment backend and docs only if a missing backend behavior is proven

**Current local evidence:**
- Stripe routes already have `requiresPayment: false`.
- `VISIBLE_METHOD_ALIASES` already includes `stripe`.
- Existing paymentFlow tests already cover Stripe as a top-level method.

**Steps:**
1. Mark the two Stripe frontend fixes as `present` if inspection remains true.
2. Do not port fee/multiplier semantics without product approval.
3. For wxpay hardening, compare upstream behavior against local payment-b2 provider code before deciding.
4. If any payment change is accepted, run payment frontend tests, backend payment tests, and payment fulfillment regression.
5. Follow `docs/engineering/payment-b2-deploy.md` for backup/preflight/deploy notes.
6. Kimi review before PR.

**Default action:** skip already-present Stripe fixes; hold payment semantics.

## Task 5: Final Integration Smoke Gate and Next Scope Refresh

**Files:**
- Modify: `docs/engineering/upstream-sync-2026-05-phase2.md`

**Steps:**
1. Confirm every candidate is classified as port / skip / hold.
2. Refresh `upstream/main` and record the new SHA. If fetch fails, keep scope pinned to `48912014` and do not expand.
3. Run the final targeted backend and frontend tests for accepted slices.
4. Run `git diff --check`.
5. If any slice was deployed, record exact test/prod health, log, rollback, and monitoring notes.
6. Ask Kimi to review the final ledger and verification evidence.
7. Close or supersede stale Phase 2 branches/worktrees only after their commits are on `origin/main`.

## Quick Finish Sequence

1. Documentation closeout PR: merge this deployment record and fast-alignment plan after self-review, Kimi review, and PR-level review.
2. Scope refresh PR: fetch `upstream/main`, update the continuation matrix, and classify every remaining item as `present`, `missing`, `partial`, `divergent`, or `unknown`.
3. Large low-risk PR batch: group only `present/skip` documentation, docs/test-only churn, or small compatibility bugfixes that do not touch protected areas.
4. Payment residual audit PR: mark Stripe route/button fixes as `present`; keep fee/multiplier and wxpay semantics held unless a missing local behavior is proven.
5. Final smoke/deploy gate: deploy only accepted runtime slices; record health, 401, and severe log scans.
6. HOLD handoff: leave product/migration/config families in the tracker with exact upstream SHAs and required decision owner.

## Expected Continuation Order

1. Documentation closeout PR for `682cee12` deploy evidence and this fast-alignment plan.
2. Scope refresh against the latest reachable `upstream/main`.
3. Payment residual equivalence audit, because the likely Stripe fixes are already present and can close fast.
4. Low-risk compatibility batch for candidates classified `missing` or `partial` with no protected-surface impact.
5. Anthropic global TTL setting only if explicitly accepted.
6. Final integration smoke gate and HOLD handoff.

This order prioritizes bugfix equivalence and avoids feature/config expansion until the remaining low-risk items are exhausted.

## Hold List

Keep these held unless separately approved:

- auth identity foundation / pending OAuth flow
- WeChat OAuth auth flow outside payment
- channel monitor / channel insights
- affiliate invite rebate system
- Vertex service account
- OpenAI Fast/Flex policy
- OpenAI image API family if it requires replacing local fork behavior
- payment fee/multiplier semantics
- Anthropic global cache TTL setting, unless accepted as product behavior
- license / CLA workflow changes
- sponsor/readme churn
