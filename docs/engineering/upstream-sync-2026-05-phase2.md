# Upstream Sync 2026-05 Phase 2

This document tracks the post-payment-b2 upstream sync. It is intentionally operational: update it whenever a slice is planned, reviewed, verified, or found to affect deployment.

## Baseline

- Local base: `origin/main` at `010a662e docs(payment-b2): record order zero amount hotfix`
- Work branch: `feature/upstream-sync-2026-05-phase2`
- Worktree: `.claude/worktrees/upstream-sync-2026-05-phase2`
- Upstream reference: `upstream/main` at `48912014 chore: sync VERSION to 0.1.121 [skip ci]`
- Plan: `docs/plans/2026-05-02-upstream-sync-phase2.md`

## Pinned Upstream Scope

All upstream commits in this phase are scoped to the inspected `upstream/main` ref at `48912014`.

If `upstream/main` advances, do not auto-expand this work. Amend the plan, re-run Kimi review, and record the new pinned SHA here before considering newer commits.

## Rules

- No direct merge from `upstream/main`.
- Port small slices only.
- Kimi review is required before committing each plan or code slice.
- Production deployment is out of scope unless a separate deployment gate is approved.
- Payment-b2 is a protected baseline. Payment changes require proof that the upstream behavior is missing locally.
- Slices should stay under about 300 changed lines and 5 touched files. Larger slices must be split or justified before review.
- Every slice must record config, env var, migration, Ent, API contract, and frontend localStorage impact.

## Ordering Rules

- Backend protocol/API contract changes must land before frontend UX-only slices that depend on those shapes.
- If a protocol slice changes a frontend-consumed response shape, either include the minimal frontend compatibility patch in the same reviewed slice or explicitly schedule it immediately after.
- Scheduler changes must re-run payment fulfillment tests before commit.

## Slice Status

| Slice | Scope | Status | Kimi review | Verification | Deploy notes |
| --- | --- | --- | --- | --- | --- |
| Task 0 | Baseline verification | Pending | Not started | Not started | None |
| Task 1 | Plan + tracking docs | Ready to commit | Passed after revisions | Doc review only | None |
| Task 2 | Runtime safety: request decoding + scheduler | Pending | Not started | Not started | TBD |
| Task 3 | OpenAI Responses / Codex compatibility | Pending | Not started | Not started | TBD |
| Task 4 | Anthropic / Claude compatibility | Pending | Not started | Not started | TBD |
| Task 5 | Admin/frontend low-risk UX | Pending | Not started | Not started | TBD |
| Task 6 | Payment residual audit only | Pending | Not started | Not started | TBD |
| Task 7 | Integration smoke gate | Pending | Not started | Not started | TBD |

## Hold List

These upstream areas are intentionally held for later phases unless separately approved:

- auth identity foundation / pending OAuth flow
- WeChat OAuth auth flow outside payment
- channel monitor / channel insights
- affiliate invite rebate system
- Vertex service account
- OpenAI Fast/Flex policy
- license / CLA workflow changes
- sponsor/readme churn

## Deployment Notes

No deployment-impacting change has been accepted in Phase 2 yet.

When a slice changes migrations, config defaults, service startup behavior, payment behavior, or externally visible routes, record:

- slice and commit
- what changed
- migration files added or changed
- Ent schema or generated-code impact
- new config keys, setting names, or env vars
- new frontend `localStorage` keys
- external API / customer-facing behavior change
- whether fresh install is affected and how it was verified
- whether existing DB upgrade is affected and how it was verified
- required backup command
- Docker image rebuild required: yes/no
- safe for rolling deploy: yes/no
- expected downtime window
- monitoring/alerting impact: logs, error rates, p99 latency, dashboards
- exact test environment verification
- customer-facing changelog/API note required: yes/no
- rollback notes

For payment-specific changes, also follow the backup and preflight SQL checklist format from `docs/engineering/payment-b2-deploy.md`.

## Rollback Template

For each deployment-impacting slice, record:

- Revert command or revert PR
- Rebuild command and expected image tag
- Redeploy command
- Migration rollback status:
  - no migration: safe revert
  - additive migration: revert app first, DB rollback usually not required
  - destructive or data migration: manual rollback required, describe exact backup restore path
- Data backup location before deployment
- Health checks after rollback

## Review Log

| Date | Reviewer | Scope | Result | Follow-up |
| --- | --- | --- | --- | --- |
| 2026-05-02 | Kimi | Plan + tracking docs | First review found missing dependency, Ent/migration, race, contract, deployment, and rollback gates | Addressed in plan and tracking doc |
| 2026-05-02 | Kimi | Revised plan + tracking docs | Do not block; ready to commit | Commit docs, then run Task 0 baseline verification |
