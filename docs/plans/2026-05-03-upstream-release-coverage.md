# Upstream Release Coverage Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Finish the current upstream alignment by using upstream releases as audit boundaries and upstream PR/merge commits as the only code merge units.

**Architecture:** Do not cherry-pick a whole upstream release. First map already merged fork PRs back to upstream release intervals, then port only proven low-risk gaps. Product, schema, migration, auth, channel, affiliate, image, Vertex, Fast/Flex, and payment semantics remain `HOLD` unless a later task explicitly reopens them.

**Tech Stack:** Go 1.26.2, Vue 3, TypeScript, Vitest, PostgreSQL migrations, Git worktrees, GitHub PRs in `ddnio/sub2api`, Kimi review through `codex-buddy`.

---

## Current Baseline

- Local base: `origin/main` at `59b9cf34 docs(upstream-sync): close out deployed slices`.
- Upstream scope: `upstream/main` at `48912014 chore: sync VERSION to 0.1.121 [skip ci]`.
- Tracking doc: `docs/engineering/upstream-release-coverage-2026-05.md`.
- Historical docs: `docs/engineering/upstream-sync-2026-05-phase2.md` and `docs/plans/2026-05-02-upstream-sync-phase2.md`.
- CI baseline fix branch: `fix/ci-baseline-green-2026-05`, local commit `2e6b0cef`; push is blocked by GitHub HTTPS connectivity errors and must be retried before opening the PR.
- Known root checkout noise: untracked `.pnpm-store/`; do not add it.

If `upstream/main` moves, do not expand scope. Pin the new SHA in the tracking doc and review the plan again.

## Task 1: CI Baseline PR

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

## Task 2: Release Coverage Ledger PR

**Files:**
- Create: `docs/engineering/upstream-release-coverage-2026-05.md`
- Create: `docs/plans/2026-05-03-upstream-release-coverage.md`

**Steps:**
1. Work from `origin/main` in `.claude/worktrees/upstream-release-coverage-2026-05`.
2. Generate release intervals with:
   - `git log --first-parent --reverse v0.1.116..v0.1.121`
   - per-release checks such as `v0.1.116..v0.1.117`, `v0.1.117..v0.1.118`, etc.
3. Map already merged fork PRs #19-#35 to upstream candidates.
4. Mark each upstream candidate as `MERGED`, `PRESENT`, `PARTIAL`, `PORT`, `HOLD`, or `SKIP`.
5. Default to `HOLD` for schema/migration/Ent, auth identity, affiliate, channel monitor, image, Vertex, Fast/Flex, payment semantics, license/CLA, and sponsor/readme churn.
6. Run `git diff --check`.
7. Self-review the ledger for count drift, missing HOLD reasons, and accidental code scope.
8. Send the ledger diff and raw command evidence to Kimi before commit.
9. Open a docs-only PR with `--repo ddnio/sub2api`; no deploy. "Docs-only" describes this fork PR, not the upstream commits listed inside the ledger.

## Task 3: Safe Port PRs

Only start after Task 1 is merged or CI baseline is otherwise green.

**Steps:**
1. Choose one `PORT` item from the ledger.
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

- The release coverage effort is bounded to upstream `48912014` / `v0.1.121`.
- The release ledger is the source of truth for what remains; old phase2 docs are historical evidence.
- Whole-release cherry-picks are not allowed.
- Ambiguous upstream candidates are held, not forced.
- PRs must target `ddnio/sub2api` explicitly.
