# Upstream Release Sync Reset Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Reset the upstream-sync process to a strict release-by-release workflow, clean up branch/worktree confusion, and prevent further blind merging before older release gates are genuinely complete.

**Architecture:** Treat upstream tags as the planning and closeout boundary, but implement by upstream PR/merge commit or smaller reviewed slices. Previously pushed fork marker tags are immutable historical snapshots; if a gate was closed under weak rules, fix forward with new ledger/code PRs on latest `main` instead of rewriting public history.

**Tech Stack:** Git worktrees, GitHub PRs against `ddnio/sub2api`, Go backend tests, Vue/TypeScript frontend tests, PostgreSQL migrations, Kimi/codex-buddy reviews, test/prod deployment checks.

---

## Current Verified State

- Root checkout: `/Users/nio/project/nanafox/sub2api`.
- `main = origin/main = cfb23ff58f73465906f0f5e24a6f32d88b4ddfc4`.
- Latest merged fork marker: PR #45, `chore(release): mark fork coverage v0.1.114`.
- Existing fork sync tags: `fork/v0.1.111`, `fork/v0.1.112`, `fork/v0.1.113`, `fork/v0.1.114`.
- Upstream tags available locally through `v0.1.121`.
- Root checkout is clean except untracked `.pnpm-store/`.
- Remote model is correct:
  - `origin = https://github.com/ddnio/sub2api.git`
  - `upstream = https://github.com/Wei-Shaw/sub2api.git`

## Problem Statement

The repo has two separate problems:

1. **Process problem:** Earlier release gates treated broad `FROZEN` / `REJECTED` / vague `PRESENT/PARTIAL` entries as closed. That conflicts with the updated rule: every upstream release item must be handled before moving to the next release.
2. **Branch visibility problem:** Many old worktrees and branches remain after squash-merged PRs. The actual merged commits are on `main`, but the Git graph looks fragmented and makes it hard to see the true state.

Do not continue with `v0.1.115` completion work until the process reset is merged and the old gates are re-opened or re-closed under the stricter rule. Until that happens, `fork/v0.1.112`, `fork/v0.1.113`, and `fork/v0.1.114` remain useful historical sync markers, but the corresponding release gates are provisional rather than final correctness claims.

## Updated Rules

1. Process upstream in tag order.
2. A release item cannot close as plain `HOLD`.
3. `FROZEN` is allowed only when it means: "accepted as a dedicated project with owner, implementation plan, and explicit decision not to block the current release marker." It is not a shortcut for risky items.
4. `REJECTED` is allowed only when the fork intentionally does not want the feature/behavior, with concrete product or technical reason.
5. `PRESENT` requires code evidence and, for behavior, test or executable-path evidence.
6. `ADAPTED` requires fork-specific implementation evidence and a regression plan.
7. Runtime changes require local verification, GitHub CI, deployment to test/prod after merge, and `/health`, unauthenticated `/v1/models`, and log-scan verification.
8. Existing public fork tags are not rewritten or deleted. Incorrectly closed gates are corrected forward by new PRs on latest `main`, and the ledger must clearly say those older markers are provisional until reopened items close.
9. After each PR merges:
   - fetch `origin/main`;
   - confirm the merge commit is on `main`;
   - remove the merged worktree/branch after checking it has no uncommitted changes;
   - create the next worktree from latest `origin/main`.

## Reopened Release Items

These items were closed too loosely and must be reprocessed before declaring the fork genuinely aligned through `v0.1.114`.

### v0.1.111

Range: `v0.1.110..v0.1.111`.

- `1ef3782d` / PR #1538, broad admin/repository/frontend cleanup.
  - Current status is not acceptable as a blanket `FROZEN`.
  - Split into itemized slices: pagination/sort/search behavior, settings/public fields, repository query changes, cache behavior, frontend table preference behavior, and tests.
  - Each slice must become `MERGED`, `ADAPTED`, `PRESENT`, or explicitly `REJECTED` with evidence.
- `16126a2c` / PR #1545, smooth sidebar collapse.
  - Current blanket `REJECTED` is too broad.
  - Re-audit against the fork sidebar. If upstream UX still applies, port/adapt only the compatible collapse behavior and tests. If rejected, document the exact fork UX reason.

### v0.1.113

Range: `v0.1.112..v0.1.113`.

- `d402e722` / PR #1637, websearch, balance notification, account pricing, broad billing/settings.
  - Current blanket `FROZEN` is too broad.
  - Split into dedicated decisions: websearch emulation, balance notification, account stats pricing, billing/settings API, payment residuals, frontend views, migrations.
  - Schema/migration subitems need a separate plan before code.
- `9bf079b7` / PR #1655, payment fee multiplier.
  - Re-audit against fork payment-b2.
  - Likely outcome should be `ADAPTED` or `PRESENT` if fork `BalanceRechargeMultiplier`, `RechargeFeeRate`, order amount, refund, and checkout display already cover upstream semantics. Otherwise implement missing parity.
- `1db32d69` / PR #1666, account cost display.
  - Re-audit local usage/dashboard/account-cost data path.
  - Either prove present/adapted, or plan the missing migration and UI/API changes.

### v0.1.114

Range: `v0.1.113..v0.1.114`.

- Reconfirm PR #44 runtime changes remain valid under the stricter rule.
- No new known reopened item yet, but the gate cannot be considered final until `v0.1.111` and `v0.1.113` rechecks are resolved.

### v0.1.115

Range: `v0.1.114..v0.1.115`.

Current branch `feature/release-gate-v0.1.115-quota-scheduling` contains a partial implementation for `e8be4344` / PR #1752 only. Keep it parked until older reopened gates are handled.

Do not mark `v0.1.115` complete until every item below is processed:

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

## Worktree And Branch Cleanup Plan

Do not delete anything until the reset PR is reviewed. Cleanup must be a separate no-code maintenance step.

### Safe-cleanup candidates after verification

These worktrees correspond to merged PRs or old docs branches and appear to have no uncommitted changes, but still require a final `git -C <path> status --short` before removal:

- `.claude/worktrees/admin-frontend-followup`
- `.claude/worktrees/ci-baseline-green-2026-05`
- `.claude/worktrees/feature+min-topup-10`
- `.claude/worktrees/payment-b2`
- `.claude/worktrees/release-gate-v0.1.112`
- `.claude/worktrees/release-gate-v0.1.113`
- `.claude/worktrees/release-gate-v0.1.114`
- `.claude/worktrees/release-marker-v0.1.111`
- `.claude/worktrees/release-marker-v0.1.112`
- `.claude/worktrees/release-marker-v0.1.113`
- `.claude/worktrees/release-marker-v0.1.114`
- `.claude/worktrees/upstream-release-coverage-2026-05`
- `.claude/worktrees/upstream-sync-2026-05-phase2`
- `.claude/worktrees/upstream-sync-admin-frontend`
- `.claude/worktrees/upstream-sync-anthropic-compat`
- `.claude/worktrees/upstream-sync-continuation-plan`
- `.claude/worktrees/upstream-sync-final-smoke`
- `.claude/worktrees/upstream-sync-ledger-refresh`
- `.claude/worktrees/upstream-sync-openai-codex`
- `.claude/worktrees/upstream-sync-openai-stream`
- `.claude/worktrees/upstream-sync-openai-ws-itemref`
- `.claude/worktrees/upstream-sync-sticky-audit`
- `.claude/worktrees/upstream-sync-sticky-session`

### Do not clean yet

- `.claude/worktrees/release-gate-v0.1.115`: contains uncommitted `v0.1.115` quota scheduling work.
- `.claude/worktrees/payment-b2-phase1-plan`: contains untracked payment-b2 handoff files.
- Any worktree not listed above or any worktree that shows uncommitted changes during final cleanup verification.

### Cleanup commands

For each verified clean, merged worktree:

```bash
git -C <worktree> status --short --branch
git worktree list --porcelain
git worktree remove <worktree>
git branch -d <branch>
```

Before running `git branch -d <branch>`, map `<worktree>` to its exact `branch refs/heads/<branch>` line from `git worktree list --porcelain`. Stop if the worktree branch is ambiguous, missing, marked `[gone]`, or different from the expected PR branch.

If `git branch -d` refuses because the branch is not fully merged, stop and inspect. Do not use `-D` unless the exact branch was confirmed merged through PR and the user approved force deletion.

## Implementation Tasks

### Task 1: Commit the reset plan only

**Files:**
- Create: `docs/plans/2026-05-03-upstream-release-sync-reset.md`

**Steps:**
1. Run `git status --short --branch` in root.
2. Run `git worktree list --porcelain`.
3. Run `gh pr list --repo ddnio/sub2api --state all --limit 80 --json number,state,headRefName,baseRefName,mergeCommit,title,mergedAt`.
4. Write this plan.
5. Run `git diff --check`.
6. Self-review the plan for branch safety and release-order consistency.
7. Send the plan to Kimi/codex-buddy for independent review.
8. If no blockers, open a docs-only PR against `ddnio/sub2api`.

### Task 2: Update the release coverage ledger to mark reopened gates

**Files:**
- Modify: `docs/engineering/upstream-release-coverage-2026-05.md`

**Steps:**
1. From latest `origin/main`, create `docs/release-gate-recheck-v0.1.111-v0.1.114`.
2. Change `v0.1.111`, `v0.1.113`, and dependent later gate status from closed to reopened/provisional.
3. Replace broad `FROZEN`/`REJECTED` rows with per-item action lists and evidence requirements.
4. Run `git diff --check`.
5. Kimi review the ledger correction.
6. Open docs-only PR. No deploy.

### Task 3: Reprocess `v0.1.111` reopened items

**Files:** determined by the upstream PR slices after audit.

**Steps:**
1. Create a worktree from latest `origin/main`.
2. Audit `1ef3782d` by upstream file group, not as one 117-file import.
3. For each missing low-risk behavior, write or port focused tests first.
4. Implement the smallest compatible fork slice.
5. Re-audit `16126a2c` against current fork sidebar.
6. Run backend/frontend verification based on touched areas.
7. Kimi review code and PR.
8. Merge only after CI passes.
9. Deploy test/prod if runtime behavior changed.
10. Clean the worktree after merge.

### Task 4: Reprocess `v0.1.113` reopened items

**Files:** determined by the upstream PR slices after audit.

**Steps:**
1. Create a worktree from latest `origin/main`.
2. Re-audit `9bf079b7` payment fee multiplier against payment-b2 first, because it may close as `ADAPTED/PRESENT`.
3. Re-audit `1db32d69` account cost display.
4. Split `d402e722` into product/schema-safe subprojects.
5. For schema/migration-affecting subitems, write a dedicated plan before code.
6. Run targeted tests, Kimi review, PR CI, and deployment gates for runtime changes.
7. Update the ledger with final outcomes.

### Task 5: Resume `v0.1.115`

**Files:** current parked quota-scheduling branch plus remaining upstream item files.

**Steps:**
1. Do not rebase, recreate, or clean `.claude/worktrees/release-gate-v0.1.115` while it has uncommitted changes.
2. First preserve the parked work with one explicit action:
   - create a temporary WIP commit on `feature/release-gate-v0.1.115-quota-scheduling`; or
   - export a patch with `git diff > .omc/state/release-gate-v0.1.115-quota-scheduling.patch` plus `git diff --cached` if anything is staged; or
   - copy the changed files into a new worktree and verify the old worktree remains untouched.
3. Record which preservation path was used in the PR description or ledger note.
4. Only after the parked work is preserved, rebase or recreate the `v0.1.115` quota branch from latest `origin/main`.
5. Keep PR #1752 as a narrow slice if earlier gates are clean.
6. Then process the remaining `v0.1.115` items in order.
7. License/CLA/auth/image/payment-doc restoration items require explicit decision evidence, not silent skip.
8. Only after every `v0.1.115` item closes, bump fork marker and create `fork/v0.1.115`.

## Verification Gates

- Docs-only PR: `git diff --check`, self-review, Kimi review, GitHub docs/CI checks.
- Backend runtime PR: targeted package tests plus relevant `make test-unit` or `go test -tags=unit ./...` if shared service behavior changes.
- Frontend PR: `pnpm --dir frontend install --frozen-lockfile`, `pnpm --dir frontend typecheck`, `pnpm --dir frontend build`, targeted Vitest where available.
- Payment PR: payment backend regression tests, frontend payment flow tests, and deployment smoke tests after merge.
- Migration PR: local migration integration test, live DB read-only precheck, backup plan, test deploy, prod deploy, post-deploy schema check.
- Deployment smoke:
  - test and prod `/health` return `{"status":"ok"}`;
  - unauthenticated `/v1/models` returns 401;
  - logs since deploy contain no `panic|fatal|error|migration|failed|traceback|异常`.

## Success Criteria

- The reset plan PR is merged.
- The ledger clearly distinguishes historical marker tags from current release-gate truth.
- `v0.1.111` and `v0.1.113` reopened items are either implemented/adapted/proven present/rejected with evidence.
- Old merged worktrees/branches are cleaned after explicit verification.
- No later release marker is created until all earlier release items are closed under the stricter rule.
