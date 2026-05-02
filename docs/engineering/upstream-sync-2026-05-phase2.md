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
| Task 0 | Baseline verification | Complete | Not required; command evidence only | Passed with drift findings recorded | None |
| Task 1 | Plan + tracking docs | Ready to commit | Passed after revisions | Doc review only | None |
| Task 2 | Runtime safety: request decoding + scheduler | Request decoding sub-slice deployed to test; scheduler pending | Kimi no blockers | httputil, handler, payment tests, test smoke passed | Test image rebuilt; no config or DB change |
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

Accepted deployment-impacting changes:

### Runtime request body decoding

- Slice and commit: Task 2 request decoding sub-slice, `0f7a2042 sync(runtime): decode compressed request bodies`
- What changed: request body reader now decodes `Content-Encoding: zstd`, `gzip`, `x-gzip`, and `deflate`; unsupported or malformed encodings return errors; decompressed bodies above 64 MiB return `http.MaxBytesError`.
- Diff-size note: slightly above the soft 300-line gate because the helper change includes focused unit coverage; production code is limited to one helper file and the rest is tests/docs.
- Migration files added or changed: none
- Ent schema or generated-code impact: none
- New config keys, setting names, or env vars: none
- New frontend `localStorage` keys: none
- External API / customer-facing behavior change: clients may now send compressed request bodies to existing gateway endpoints; oversized decompressed bodies receive existing 413 handling.
- Fresh install affected: no
- Existing DB upgrade affected: no
- Required backup command: not required for this sub-slice
- Docker image rebuild required: yes
- Safe for rolling deploy: yes
- Expected downtime window: normal rolling restart only
- Monitoring/alerting impact: watch 400/413 rates on gateway routes after deploy
- Exact test environment verification: deployed to `sub2api-test` on 2026-05-02; local health and public health checks returned `{"status":"ok"}`; gzip-compressed `/v1/responses` request returned HTTP 200; gzip body decompressing above 64 MiB returned HTTP 413 with `Request body too large, limit is 64MB`; post-deploy log-level check found no `ERROR`, `FATAL`, `PANIC`, `POSTCHECK FAILED`, or `PREFLIGHT FAILED`.
- Customer-facing changelog/API note required: optional; useful if announcing compressed request-body support
- Rollback notes: revert the request decoding commit and redeploy; no DB rollback

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
| 2026-05-02 | Kimi | Request body decoding code diff | No blockers; suggested direct dependency, typed limit error, extra gzip tests, comment update | Addressed suggestions and re-ran tests |
| 2026-05-02 | Kimi | Final request body decoding diff | No blockers | Commit after recording deployment note |

## Test Deployment Log

### 2026-05-02 Runtime Request Body Decoding

- Host: `108.160.133.141`
- Environment: `test`
- Branch: `feature/upstream-sync-2026-05-phase2`
- Commit deployed: `0f7a2042 sync(runtime): decode compressed request bodies`
- Backup: `/home/nio/backups/sub2api_test_pre_runtime_decode_20260502-072241.sql` (58M)
- Deploy command:

```bash
cd /data/service/sub2api
git fetch origin
git checkout feature/upstream-sync-2026-05-phase2
git pull --ff-only origin feature/upstream-sync-2026-05-phase2
bash deploy/deploy-server.sh test
```

- Container result: `sub2api-test` healthy on `127.0.0.1:8081->8080/tcp`
- Production impact: `sub2api-prod` remained running and healthy; production was not redeployed.
- Health checks:

```bash
curl -fsS http://127.0.0.1:8081/health
curl -fsS https://router-test.nanafox.com/health
# {"status":"ok"}
```

- Compressed request smoke:
  - gzip `/v1/responses` with a test API key returned HTTP 200.
  - gzip request that decompresses above 64 MiB returned HTTP 413 and `Request body too large, limit is 64MB`.
- Log check:

```bash
docker logs --since 5m sub2api-test 2>&1 | egrep "\t(ERROR|FATAL|PANIC)\t|panic|POSTCHECK FAILED|PREFLIGHT FAILED" || true
# no output
```

## Task 0 Baseline Results

Recorded on 2026-05-02 from `feature/upstream-sync-2026-05-phase2`.

Refs:

- `HEAD`: `43ed49c9`
- `origin/main`: `010a662e`
- `upstream/main`: `48912014`

Dependency drift:

- `backend/go.mod`
- `backend/go.sum`
- `frontend/package.json`
- `frontend/pnpm-lock.yaml`

Ent and migration drift:

- Upstream has broad Ent generated-code drift, including auth identity, pending auth session, channel monitor, group/user changes, and payment plan removals.
- Upstream migration drift is broad and risky: auth identity, channel monitor, affiliate, account stats, notify settings, and deletions of several fork payment-b2 migrations appear in the diff.
- Any slice that touches these areas must classify migrations before code is ported. Direct upstream merge remains blocked.

Baseline verification:

```bash
cd backend
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment
# ok github.com/Wei-Shaw/sub2api/internal/payment 2.552s

GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/service -run 'Test.*Payment|Test.*Order|Test.*Provider|Test.*Refund|Test.*Fulfillment|Test.*Config'
# ok github.com/Wei-Shaw/sub2api/internal/service 1.516s

cd ../frontend
pnpm install --frozen-lockfile
pnpm exec vitest run src/__tests__/buttonClasses.spec.ts src/components/payment/__tests__/paymentFlow.spec.ts src/views/user/__tests__/PaymentView.spec.ts
# 3 files passed, 27 tests passed
```

Frontend bootstrap note:

- The new worktree had no `node_modules`; `pnpm install --frozen-lockfile` reused the existing store and downloaded 0 packages.
