# Upstream Release Continuation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Continue upstream alignment from the current verified fork state without overwriting fork-specific product behavior or risking online data.

**Architecture:** Use upstream releases as the planning, review, deployment, and fork-marker boundary. Inside each release, process upstream mainline entries in order, expand merge PR internal commits, and import or adapt only the behavior that is compatible with the fork. Keep public fork marker tags immutable and fix gaps forward on latest `main`.

**Tech Stack:** Git worktrees, GitHub PRs against `ddnio/sub2api`, Go/Ent/PostgreSQL migrations, Vue/TypeScript frontend, Kimi or local review agents, test/prod deployment smoke checks.

---

## Current Verified State

- Root checkout: `/Users/nio/project/nanafox/sub2api`.
- Current `origin/main`: `f33ca0b5 docs(upstream): tighten release gate sync workflow (#56)`.
- Root checkout is clean except untracked `.pnpm-store/`.
- Release gates final through `v0.1.113..v0.1.114`.
- Existing fork marker tags through `fork/v0.1.114` are historical markers and must not be moved.
- Current active gate: `v0.1.114..v0.1.115`.
- Current continuation branch: `.claude/worktrees/release-v0.1.115-closeout`, branch `sync/v0.1.115-release-closeout`, ahead of `origin/main` by 26 commits when checked on 2026-05-04. It is not clean: `backend/internal/service/account_test_service_openai_test.go` and `docs/engineering/upstream-release-coverage-2026-05.md` have uncommitted v0.1.115 closeout fixes.
- Latest merged release/process PRs:
  - PR #54: `sync(v0.1.113): close release alignment`, merged as `6140ec11`.
  - PR #55: `fix: restore balance notify card rendering`, merged as `252b3f1f`.
  - PR #56: `docs(upstream): tighten release gate sync workflow`, merged as `f33ca0b5`.

Do not restart earlier releases. Treat `v0.1.111`, `v0.1.112`, `v0.1.113`, and `v0.1.114` as closed unless new evidence proves a concrete gap. New gaps are fixed forward on latest `main`; old fork tags are not rewritten.

## Non-Negotiable Rules Going Forward

1. Process upstream releases in tag order.
2. Do not start `v0.1.115..v0.1.116` until `v0.1.114..v0.1.115` is closed, reviewed, merged, deployed when required, and tagged as `fork/v0.1.115`.
3. One completed upstream release should normally produce one fork PR. Split only for migrations/schema, auth/payment/security/data-risk, very large conflict areas, or CI unblock work.
4. Do not create one PR per commit unless the commit is high-risk or blocks the whole release.
5. For every upstream merge PR, first-parent is only the release index. Expand and check the internal commits before claiming coverage.
6. A release item cannot close as `HOLD`, `PORT`, `PARTIAL`, or `REOPENED`.
7. Hard items cannot be skipped. Product, schema/migration, auth, payment, security, and data-risk items must become `MERGED`, `ADAPTED`, `PRESENT`, `REJECTED`, or `FROZEN` with evidence. Reserve `SKIP` for non-applicable churn such as sponsor/readme/no-op repository maintenance.
8. Routine deployment happens once after the whole release closes. Deploy earlier only for security hotfixes, migrations/schema, payment/auth/data-risk, or urgent production fixes.
9. For low-risk non-runtime or obviously contained code, skip Kimi if it is slowing the release down; record local self-review plus agent/code review evidence instead. Use stronger review for auth, payment, migration, schema, and data-risk work.
10. If a cherry-pick or implementation starts producing a broad hand-written diff, conflicts across fork-specific payment/auth/migration/UI surfaces, or branch-state confusion, stop and write an import audit before coding more.
11. Do not replace fork invitation/referral data models with upstream auth-source default grants. The fork has real online data in `redeem_codes(type=invitation)`, `users.referral_code`, and `user_referrals`; upstream auth-source grants may be added as a compatible provider-default feature only after a data audit proves they do not overwrite or reinterpret those records.

## Data Compatibility Gate

Any upstream item that touches auth identity, invitation/referral/affiliate, payment, subscriptions, user balance, provider configuration, Ent schema, or migrations must pass a data compatibility gate before implementation and again before deployment.

The gate has three parts:

1. **Current production shape:** run read-only SQL against ToC test, ToC production, and ToB production where the shared runtime is affected. Record table existence, column shape, setting values, row counts, and pending/processed state. If SSH/network access fails, do not guess; record the failure and retry through the documented SSH/API fallback before making a merge decision.
2. **Semantic mapping:** write down how every existing fork concept maps to the upstream concept. If there is no one-to-one mapping, the item must be `ADAPTED`, `REJECTED`, or `FROZEN`; it must not be hard-replaced by upstream code.
3. **No-loss migration plan:** for any accepted schema/data change, define a forward migration, backfill behavior, idempotency guard, rollback/restore path, and tests. Existing rows must remain readable and must not be re-awarded, re-charged, deleted, or silently detached.

Minimum evidence to record in the release ledger:

- Read-only SQL used and summarized results.
- Tables/columns/settings affected.
- Whether the change is additive, bridge/backfill, semantic replacement, or rejection.
- Targeted tests and deployment preflight needed.
- Manual checks required after deployment.

### Invitation, Referral, and Affiliate Rule

The fork currently has two separate concepts that upstream affiliate must not collapse:

- `redeem_codes(type='invitation')` is a registration admission code.
- `users.referral_code` plus `user_referrals` is referral attribution and first-recharge reward state.

Production ToC read-only check on 2026-05-04 showed:

- `invitation_code_enabled=false`.
- `referral_enabled=true`.
- `redeem_codes(type='invitation')`: 9 used, 1 unused.
- active users with `referral_code`: 60/60.
- `user_referrals`: 10 total, 6 rewarded, 4 pending first-recharge reward, inviter rewards sum 30, invitee rewards sum 30.
- referral reward audit rows exist through `redeem_codes(type='ref_inviter')` and `redeem_codes(type='ref_invitee')`.

Production ToB read-only check on 2026-05-04 showed:

- `invitation_code_enabled=false`.
- `referral_enabled=false`.
- active users with `referral_code`: 2/3.
- `user_referrals`: 0.

Therefore, when upstream affiliate commits are reached in later releases:

1. Do not replace `redeem_codes(type='invitation')`; it remains the fork's registration gate.
2. Do not discard or regenerate existing `users.referral_code` values.
3. Do not delete `user_referrals` or lose `reward_granted_at`, `inviter_rewarded`, `invitee_rewarded`, or reward snapshot semantics.
4. If adopting upstream `user_affiliates`, backfill it from existing `users.referral_code` and `user_referrals` so old links keep working.
5. Already rewarded rows must be marked processed in any upstream ledger/affiliate bridge so they cannot be paid twice.
6. Pending rows must keep their current first-recharge reward behavior unless the release explicitly adapts that behavior with a tested migration.
7. Upstream affiliate quota/freeze/transfer behavior is a product change. It can be adopted only through an `ADAPTED` implementation that preserves existing fork data, or rejected/frozen with an owner and reason.

## v0.1.115 Release Scope

Range:

```bash
git log --oneline --first-parent --reverse v0.1.114..v0.1.115
git log --oneline --reverse v0.1.114..v0.1.115
```

Mainline entries to close:

- `6cfdf4ec` version sync to `0.1.114`.
- `6c73b621` / PR #1734 Kyren payment docs.
- `51af8df3` / PR #1731 rate billing autofill response limit.
- `061fd48d` / PR #1749 xhigh reasoning effort.
- `e8be4344` / PR #1752 quota-exceeded scheduling.
- `f5ee9379` / PR #1753 orphaned scheduled tests.
- `23def40b` license MIT to LGPL v3.
- `a8854947` / PR #1764 wxpay pubkey hardening.
- `ffc9c387` / PR #1766 codex drop removed models.
- `960b2bb8` CLA workflow.
- `78f691d2` sponsor churn.
- `8eb3f9e7` / PR #1785 auth identity foundation.
- `32107b4f` / PR #1795 OpenAI image API sync.
- `4d0483f5` GPT image test feature.
- `ddf80f5e` / PR #1799 auth identity follow-up.
- `45065c23` auth migration order test.
- `c6d25f69` restore payment docs/files.
- `1da4bd72` / PR #1802 profile auth bindings i18n.
- `755c7d50` README revert.

The current `sync/v0.1.115-release-closeout` branch has already started this gate and includes commits for PR #1731/#1752/#1764/#1766/#1785 partial work. Continue from that branch after rebasing or recreating it safely on latest `origin/main`.

## v0.1.115 Execution Order

### Task 1: Preserve and Rebase the Active v0.1.115 Work

**Files:**
- Existing worktree: `.claude/worktrees/release-v0.1.115-closeout`
- Branch: `sync/v0.1.115-release-closeout`

**Steps:**
1. Run `git -C .claude/worktrees/release-v0.1.115-closeout status --short --branch`.
2. The current known dirty files are `backend/internal/service/account_test_service_openai_test.go` and `docs/engineering/upstream-release-coverage-2026-05.md`. Inspect them first; if they are still only the verified image-test signature fix and ledger closeout update, commit them before any rebase.
3. Confirm the branch contains only v0.1.115 work by reviewing `git log --oneline origin/main..HEAD`.
4. Rebase onto latest `origin/main` only after the worktree is clean.
5. If rebase conflicts become broad, stop and create a fresh worktree from `origin/main`, then cherry-pick the existing local v0.1.115 commits one at a time.

**Verify:**

```bash
git -C .claude/worktrees/release-v0.1.115-closeout status --short --branch
git -C .claude/worktrees/release-v0.1.115-closeout log --oneline origin/main..HEAD
```

### Task 2: Close Low-Risk Already-Started Items

**Scope:**
- Version marker source row.
- Payment docs rows.
- License/CLA/sponsor/readme rows.
- PR #1731, PR #1749, PR #1752, PR #1753, PR #1764, PR #1766 if existing local commits already cover them.

**Steps:**
1. For each item, compare upstream commit diff with the current branch.
2. If the behavior is already present, mark `PRESENT` or `ADAPTED` with file/test evidence.
3. If missing and low-risk, cherry-pick or port directly in the release branch.
4. Keep these in the same v0.1.115 PR unless a data-risk item appears.

**Verify:**
- Backend targeted tests for touched service/handler/payment/model-policy areas.
- `pnpm --dir frontend typecheck` if frontend files changed.
- `git diff --check`.

### Task 3: Finish PR #1785 Auth Identity Work

**Current branch evidence:**
- Auth identity schema core is partially adapted.
- Pending auth session service exists.
- Pending OAuth exchange endpoint exists.
- LinuxDo/OIDC invitation-required paths bridge into DB-backed pending sessions while preserving legacy fragment compatibility.
- Frontend pending OAuth exchange API exists.
- Backend generic pending OAuth create-account/bind-login endpoints now exist. They align the upstream API shape while preserving this fork's current invitation-code redemption and referral data semantics; do not replace existing `redeem_codes` invitation data or `user_referrals` history with upstream auth-source-default grants without a dedicated data migration plan.

**Remaining areas to close:**
- Frontend callback views use pending exchange while preserving legacy `pending_oauth_token` fallback.
- Profile identity binding backend/API/UI.
- Frontend email verification and pending OAuth onboarding completion UI.
- 2FA pending bind paths where upstream requires them.
- WeChat OAuth/bind capability handling.
- Legacy migration/report closeout.
- User activity/profile support.
- Internal PR #1785 commits and follow-up PR #1799/#1802 rows mapped to final outcomes.

**Rules:**
1. Do not copy upstream auth files wholesale over fork auth/payment/profile behavior.
2. Keep legacy callback fragments compatible until all current frontend flows are migrated.
3. Migration filenames must avoid already-applied fork numbers; use fork-safe new filenames.
4. Run real DB read-only preflight before deploying auth identity migrations.
5. Preserve existing invitation-code and referral semantics. `redeem_codes(type=invitation)` remains the registration gate source when invitation codes are enabled; `users.referral_code` and `user_referrals` remain the attribution/reward source when referral is enabled. Upstream `user_provider_default_grants` is not a replacement for either model.

**Verify:**
- Auth/session/user targeted Go tests.
- Frontend auth/profile targeted Vitest where available.
- `pnpm --dir frontend typecheck`.
- `git diff --check`.
- Strong review required before merge because this touches auth, migrations, and online user identity data.

### Task 4: Close PR #1795 and Image Test Items

**Scope:**
- `32107b4f` OpenAI image API sync.
- `4d0483f5` GPT image test feature.

**Steps:**
1. Compare upstream image API changes against local OpenAI gateway/image code.
2. Preserve fork scheduler, billing, routing, and model whitelist behavior.
3. Port only missing image API compatibility and test support.
4. Confirm image billing/account selection is not overwritten by upstream defaults.

**Verify:**
- Targeted OpenAI/image service tests.
- Payment/billing regression if image billing paths changed.
- Frontend typecheck if UI changed.

### Task 5: Release-Level Closeout Review

**Steps:**
1. Re-run first-parent list for `v0.1.114..v0.1.115`.
2. Re-run full commit list for `v0.1.114..v0.1.115`.
3. Expand every merge PR internal commit and map it to a final release outcome.
4. Confirm the ledger has no release-local `HOLD`, `PORT`, `PARTIAL`, or `REOPENED`.
5. Run release-width local tests based on touched areas.
6. Run independent code review. Use Kimi or a local code-review agent; use Kimi only where it adds value or risk demands it.
7. Open one v0.1.115 PR against `ddnio/sub2api` unless auth/migration/image work needs a safety split.
8. Wait for GitHub CI.
9. Merge only when CI and review are clean.

**Required release evidence:**
- Local test commands and results.
- Review result.
- GitHub CI result.
- Deployment decision.
- Manual/browser QA note for changed auth/profile/payment/image UI paths.

### Task 6: Mark, Deploy, and Tag v0.1.115

Because current v0.1.115 work includes auth identity migrations and runtime auth/profile/payment/image behavior, deployment is required after merge.

**Steps:**
1. After the release PR merges and CI is green on `main`, update `backend/cmd/server/VERSION` to `0.1.115` in a marker PR or marker commit.
2. Merge the marker to `main`.
3. Create and push annotated tag `fork/v0.1.115` pointing at the merged fork marker commit.
4. Back up ToC test DB, ToC prod DB, and ToB prod DB if migrations are included.
5. Run real DB read-only preflight for new migration filenames and target columns/tables.
6. Deploy ToC test from the tagged marker commit on `main`.
7. Verify test `/health`, unauthenticated `/v1/models` returns 401, container health, migration status, and severe logs.
8. Run browser/manual checks on test for auth/profile/payment/image paths touched by the release.
9. Deploy ToC production from the same tagged marker commit.
10. Verify production with the same smoke checks.
11. Deploy ToB production from the same tagged marker commit if shared runtime/frontend behavior changed.
12. Verify ToB production.

## Branch and Worktree Cleanup

Do not use old branch graph screenshots to judge release status. The only status source is latest `origin/main` plus the release ledger.

After v0.1.115 merges:

1. Fetch `origin`.
2. Confirm the PR merge commit is on `origin/main`.
3. Confirm each old worktree has no uncommitted changes.
4. Remove only merged/clean worktrees.
5. Delete only branches that Git confirms are merged, or ask before force deleting.

Keep these until their content is explicitly resolved:

- `.claude/worktrees/release-v0.1.115-closeout`
- `.claude/worktrees/release-gate-v0.1.115`
- `.claude/worktrees/payment-b2-phase1-plan`
- Any worktree with local uncommitted or unpushed changes.

## Speed Optimizations That Are Safe

- Batch low-risk entries inside one release PR.
- Use one release-level deployment instead of deploying every small commit.
- Skip Kimi for docs-only or low-risk mechanical changes when a local code-review agent and tests cover the risk.
- Do not re-audit final gates `v0.1.111` through `v0.1.114` unless new concrete evidence appears.
- Do not clean branches during active implementation unless branch confusion blocks the release.
- Do not expand beyond upstream `v0.1.115` until the current gate is fully closed.

## Slow Work That Must Not Be Skipped

- Auth identity migrations and legacy-user migration safety.
- Payment/provider/wxpay compatibility and online order safety.
- Fork-specific migration numbering and real DB shape checks.
- OpenAI/image billing and scheduler routing behavior.
- Release closeout review that checks both first-parent and full commit lists.
- Test/prod deployment verification after a completed runtime release.

## Completion Definition

The upstream continuation plan is successful when:

1. `v0.1.115` has every upstream mainline entry and internal PR commit mapped to a final outcome.
2. All accepted v0.1.115 behavior is merged to `main`.
3. Required tests and CI are green.
4. Required test/prod/ToB deployment checks are recorded.
5. `backend/cmd/server/VERSION` is marked to `0.1.115`.
6. `fork/v0.1.115` is created and pushed.
7. Only then does work start on `v0.1.115..v0.1.116`.
