# Upstream Release Coverage Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Finish upstream alignment through sequential release gates. A release/tag is complete only when every upstream change in that release has a final recorded outcome and the fork's release/tag marker is updated before moving to the next release.

**Architecture:** Do not cherry-pick a whole upstream release. Use the release/tag as the planning and closeout boundary, and use upstream PR/merge commits or smaller reviewed hunks as implementation units. Product, schema, migration, auth, channel, affiliate, image, Vertex, Fast/Flex, and payment semantics must receive explicit release-gate decisions instead of being silently carried as unresolved `HOLD`.

**Tech Stack:** Go 1.26.2, Vue 3, TypeScript, Vitest, PostgreSQL migrations, Git worktrees, GitHub PRs in `ddnio/sub2api`, Kimi review through `codex-buddy`.

---

## Current Baseline

- Local base: `origin/main` at `2acdfd66 docs(upstream-sync): add release coverage ledger`.
- Upstream published-tag scope: latest local upstream tag `v0.1.121` at `9d801595 test: 更新管理员设置契约字段`.
- Upstream main observed locally: `b2bdba78 stabilize image request handling`; this is outside the current published-tag scope.
- Tracking doc: `docs/engineering/upstream-release-coverage-2026-05.md`.
- Historical docs: `docs/engineering/upstream-sync-2026-05-phase2.md` and `docs/plans/2026-05-02-upstream-sync-phase2.md`.
- CI baseline fix PR: #36, merged at `fed065e6`; GitHub CI passed after lint, unit, integration, and security gates.
- Known root checkout noise: untracked `.pnpm-store/`; do not add it.

If `upstream/main` moves, do not expand scope. Only expand after a new upstream tag is fetched and the release interval is reviewed.

2026-05-03 refresh note: `git fetch upstream --tags --prune` and `git fetch origin --prune` hit GitHub transport timeouts. Use existing local refs for the current release gate. For push/PR creation, first retry with `git -c http.version=HTTP/1.1 push`; if HTTPS still fails, use the existing GitHub Git Data API fallback documented in `docs/engineering/upstream-sync-2026-05-phase2.md`.

## Task 1: CI Baseline PR - Completed

**Files:**
- Modify only the CI-drift files needed to restore tests/lint.

**Steps:**
1. Work from `origin/main` in `.claude/worktrees/ci-baseline-green-2026-05`.
2. Fix the `NewAuthService` test call signature drift in middleware tests.
3. Apply only formatting/lint cleanups already proven by current CI drift.
4. Run:
   - `cd backend && make test-unit`
   - `cd backend && GOCACHE="$PWD/../../.cache/go-build" go test -tags=unit ./internal/server/middleware ./internal/service`
   - `pnpm --dir frontend install --frozen-lockfile`
   - `pnpm --dir frontend typecheck`
   - `pnpm --dir frontend build`
   - `git diff --check`
5. Use Kimi pre-commit review on the diff and verification evidence.
6. Push to `origin` and open a PR with `--repo ddnio/sub2api`.
7. Let GitHub Actions run the official `golangci/golangci-lint-action@v9` with `version: v2.9`.
8. Merge only after PR CI and PR-level Kimi review have no blockers.

**Status:** Completed via PR #36. Kimi pre-commit review returned `NO BLOCKERS` in prose; PR-level Kimi attempts timed out without content, so GitHub CI and local self-review were the final merge gates.

## Release Gate Rule

Process upstream in strict tag order:

1. Start from the earliest open gate.
2. For every upstream change in that release, decide one final state:
   - merge it into the fork
   - adapt it through a fork-specific implementation
   - prove it is already present locally
   - reject/skip it as not applicable to this fork
   - freeze it as a later dedicated project with owner/reason
3. Run the required verification/review for accepted changes.
4. Update the fork release marker and create the corresponding fork tag only after the release gate closes.
5. Move to the next upstream tag.

`HOLD` is an intermediate state, not completion. A release containing unresolved `HOLD` items remains open.

## Current Start Point

The 2026-04 slice-based sync was useful but did not satisfy the sequential release gate rule. Do not process its HOLD items as a separate global queue. The earliest known open interval is `v0.1.110..v0.1.111`, not `v0.1.117`.

Current next task:

1. Finish review of every item in `v0.1.110..v0.1.111`.
2. If review confirms no unresolved item remains, update the fork release marker from `0.1.110` to `0.1.111` in a small marker PR.
3. After the marker PR lands, create and push the fork tag `v0.1.111` on the merged fork commit.
4. Then start `v0.1.111..v0.1.112`.

## Task 2: Release Coverage Ledger PR

**Files:**
- Create: `docs/engineering/upstream-release-coverage-2026-05.md`
- Create: `docs/plans/2026-05-03-upstream-release-coverage.md`

**Steps:**
1. Work from `origin/main` in `.claude/worktrees/upstream-sync-final-smoke`.
2. Generate release intervals with:
   - `git log --first-parent --reverse v0.1.110..v0.1.111`
   - next releases only after the current gate closes.
3. Map already merged fork PRs #19-#35 to upstream candidates.
4. Mark each upstream candidate as `MERGED`, `ADAPTED`, `PRESENT`, `PARTIAL`, `PORT`, `REJECTED`, `FROZEN`, or `SKIP`.
5. Treat `HOLD` as unresolved unless it has an explicit owner/reason and is converted to accepted-later, rejected, or long-term-frozen.
6. Default high-risk families to "needs explicit decision", not completed: schema/migration/Ent, auth identity, affiliate, channel monitor, image, Vertex, Fast/Flex, payment semantics, license/CLA, and sponsor/readme churn.
7. Run `git diff --check`.
8. Self-review the ledger for count drift, unresolved HOLD items, missing decision owners/reasons, and accidental code scope.
9. Send the ledger diff and raw command evidence to Kimi before commit.
10. Open a docs-only PR with `--repo ddnio/sub2api`; no deploy. "Docs-only" describes this fork PR, not the upstream commits listed inside the ledger.

## Task 3: Safe Port PRs

Only start after Task 1 is merged or CI baseline is otherwise green. This gate is now satisfied by PR #36.

**Steps:**
1. Choose the next item from the earliest open release gate. Do not pick from a later release out of order unless it is an emergency production fix.
2. Inspect the upstream PR/commit and local code before editing.
3. Prove local behavior is missing with a focused test or code-path evidence.
4. Port the smallest upstream-compatible hunk.
5. Update the tracking doc with impact, verification, rollback, and deploy notes.
6. Run targeted tests, `git diff --check`, self-review, Kimi pre-commit review, PR-level Kimi review, and GitHub CI.
7. Deploy only for runtime code changes after a recorded deployment gate.

## Verification Defaults

- Docs-only: `git diff --check` plus Kimi review.
- Backend runtime: targeted package tests plus `make test-unit` when CI baseline is relevant.
- Frontend: `pnpm --dir frontend typecheck` and `pnpm --dir frontend build` when frontend is touched.
- Scheduler, billing, account selection, or payment-adjacent changes: include payment regression tests.
- Any migration/schema/config change: stop and create a dedicated plan before implementation.

## Assumptions

- The release coverage effort is currently bounded to upstream `48912014` / `v0.1.121`.
- The release ledger is the source of truth for gate order; old phase2 docs are historical evidence when deciding earlier release intervals.
- Whole-release cherry-picks are not allowed.
- Ambiguous upstream candidates are held only temporarily; they must be resolved before the release gate closes.
- `v0.1.110..v0.1.111` is decision-complete only after self-review and Kimi review confirm the 17-item matrix has no unresolved item.
- A release gate is not fully closed until both the fork release marker and corresponding fork tag exist.
- PRs must target `ddnio/sub2api` explicitly.
