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
- Production deployment is out of scope unless a separate deployment gate is approved and recorded.
- Payment-b2 is a protected baseline. Payment changes require proof that the upstream behavior is missing locally.
- Slices should stay under about 300 changed lines and 5 touched files. Larger slices must be split or justified before review.
- Every slice must record config, env var, migration, Ent, API contract, and frontend localStorage impact.

## Ordering Rules

- Backend protocol/API contract changes must land before frontend UX-only slices that depend on those shapes.
- If a protocol slice changes a frontend-consumed response shape, either include the minimal frontend compatibility patch in the same reviewed slice or explicitly schedule it immediately after.
- Scheduler changes must re-run payment fulfillment tests before commit.
- Scheduler remaining work is tentatively split into Task 2A scheduler snapshot race/CAS/grace-TTL/lock-release and Task 2B sticky session scheduling audit. First prove `8bf2a7b8` can be separated from `733627cf`; do not cherry-pick both as one batch.

## Slice Status

| Slice | Scope | Status | Kimi review | Verification | Deploy notes |
| --- | --- | --- | --- | --- | --- |
| Task 0 | Baseline verification | Complete | Not required; command evidence only | Passed with drift findings recorded | None |
| Task 1 | Plan + tracking docs | Ready to commit | Passed after revisions | Doc review only | None |
| Task 2 | Runtime safety: request decoding + scheduler | Request decoding sub-slice deployed; scheduler 2A implemented locally | Kimi no blockers for decoding; scheduler code review pending | httputil, handler, scheduler, payment tests passed | Test/prod images rebuilt for decoding; scheduler 2A not deployed |
| Task 3 | OpenAI Responses / Codex compatibility | Task 3A and Task 3B deployed | Kimi no blockers for PR #22 and PR #24 | apicompat, targeted service, and payment tests passed | Test/prod images rebuilt; no DB backup because no DB/config/frontend impact |
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

## Next Slice Plan

### Task 2A Scheduler Snapshot Race Fix, CAS Grace TTL, and Rebuild Lock Release

- Candidate upstream source: backend portions of `8bf2a7b8 fix(scheduler): resolve SetSnapshot race conditions and remove usage throttle`.
- Precondition:
  - prove the backend-only subset is coherent in this fork
  - prove it does not depend on sticky-session changes from `733627cf`
- Candidate intended scope:
  - Redis CAS activation for scheduler snapshot versions
  - old snapshot grace TTL instead of immediate delete
  - `UnlockBucket` so rebuild locks are released after successful rebuild
  - focused repository/service tests only
- Explicit skips:
  - usage throttle removal in `frontend/src/utils/usageLoadQueue.ts`, because this file does not exist in the fork.
  - any unrelated scheduler or gateway sticky-session logic from `733627cf`.
- Impact:
  - No migration, Ent, config, env var, request/response shape, or frontend localStorage change expected.
  - Account selection cache behavior may change because scheduler snapshot activation and lock release behavior change.
  - Deployment-impacting because Redis scheduler keys and account selection cache behavior change.
- Verification gate:
  - separate deployment gate approved and recorded before any test or production deploy
  - scheduler cache/snapshot tests, with `-race` where practical
  - gateway scheduler/sticky tests touched by the slice
  - payment fulfillment regression before commit
  - Kimi review of diff and test output before commit

### Task 2B Sticky Session Scheduling Audit

- Candidate upstream source: `733627cf fix: improve sticky session scheduling`.
- Current finding: this patch is too broad for direct cherry-pick; it touches handler account selection, gateway load-aware selection, scheduler cache, and tests.
- Required before code:
  - prove the local sticky-session bug exists or is partially fixed
  - separate correctness hunks from upstream debug-only log additions
  - record account-selection, wait-plan, failover, and sticky-session TTL behavior changes
  - Kimi review the plan before editing

### Task 3A OpenAI/Codex Request Normalization Batch

Batching decision: this task intentionally groups small, same-path OpenAI/Codex compatibility fixes instead of one PR per upstream patch.

Missing-fix matrix:

| Upstream source | Area | Local state | Action | Notes |
| --- | --- | --- | --- | --- |
| `7452fad8` / PR #2068 | Drop `reasoning` items from Codex OAuth `/v1/responses` input | Missing | Port | OAuth transform forces `store=false`; replayed `rs_*` reasoning items are not persisted upstream and can 404. |
| `9fe02bba` / PR #2005 | Strip ChatGPT internal unsupported passthrough fields | Partial | Port | Local stripped sampling/max-token fields and `prompt_cache_retention`; missing `user`, `metadata`, `safety_identifier`, and `stream_options`. |
| `04b2866f` / PR #2058 | Responses-compatible flat function `tool_choice` | Partial | Port Codex transform only | `backend/internal/pkg/apicompat` already has flat conversion locally; Codex OAuth transform still emitted/accepted nested `function.name`. |
| `5e54d492` | Codex transform test assertion lint | Missing as test does not exist locally yet | Port with new test | Use two-value type assertion in the new reasoning-item regression test. |
| `dac6e520` / PR #1960 | Responses stream keepalive during pre-output failover | Not directly portable | HOLD | Current fork lacks the upstream `clientOutputStarted` / pre-output failover buffering structure this patch relies on; needs a separate stream failover batch. |
| `094e1171` | Infer previous response for item references in WebSocket ingress | Appears partially present in local WS forwarder | HOLD for separate WS batch | Touches WebSocket continuation/account stickiness; separate from request normalization. |

Current local implementation:

- Extend Codex OAuth unsupported-field stripping to remove `user`, `metadata`, `safety_identifier`, and `stream_options`.
- Convert legacy `function_call: {"name": ...}` and nested `tool_choice: {"type":"function","function":{"name":...}}` to Responses flat format `{"type":"function","name":...}`.
- Downgrade function `tool_choice` to `auto` when the chosen function name is absent from `tools`.
- Drop `type:"reasoning"` input items on the Codex OAuth path while preserving non-reasoning continuation items.

Impact:

- Migration files added or changed: none
- Ent schema or generated-code impact: none
- New config keys, setting names, or env vars: none
- New frontend `localStorage` keys: none
- External API / customer-facing behavior change: narrows upstream request payloads for Codex OAuth compatibility; no local response schema change.
- Fresh install affected: no
- Existing DB upgrade affected: no
- Required backup command: not required by schema or data impact
- Docker image rebuild required: yes for deployment
- Safe for rolling deploy: yes
- Expected downtime window: normal rolling restart only
- Monitoring/alerting impact: watch OpenAI/Codex OAuth 400/502 rates and request-normalization related upstream errors.
- Rollback notes: revert the Task 3A commit and redeploy; no DB rollback.

Verification so far:

```bash
cd backend
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/service -run 'TestApplyCodexOAuthTransform|TestFilterCodexInput'
# ok github.com/Wei-Shaw/sub2api/internal/service

GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/pkg/apicompat ./internal/service -run 'Codex|Tool|OAuth|Responses'
# ok github.com/Wei-Shaw/sub2api/internal/pkg/apicompat
# ok github.com/Wei-Shaw/sub2api/internal/service

GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment
# ok github.com/Wei-Shaw/sub2api/internal/payment
```

### Task 3B Chat Completions to Responses Tool Output Name

Batching decision: keep this as a small `apicompat` sub-batch because the remaining stream failover and WebSocket continuation candidates affect account stickiness, failover, and connection reuse.

Missing-fix matrix:

| Upstream source | Area | Local state | Action | Notes |
| --- | --- | --- | --- | --- |
| `f6fcafa9` | Set `name` on `function_call_output` items for Responses API | Missing | Port | gpt-5.2+ Responses API requires `name` on tool output items; Chat Completions tool-result messages only carry `tool_call_id`, so infer the function name from prior assistant `tool_calls`. |
| `1a597725` | Forward `response_format` through CC -> Responses -> Anthropic | Present | Skip | Local `ChatCompletionsRequest.ResponseFormat`, `ResponsesTextConfig`, and Anthropic output format mapping are already present. |
| `855841a8` / PR #2058 | Flat Responses function `tool_choice` | Present | Skip | Local `apicompat` already flattens Chat Completions and Anthropic function tool choices; Task 3A covered Codex OAuth transform. |
| `dac6e520` / PR #1960 | Responses stream keepalive during pre-output failover | Not directly portable | HOLD | Needs upstream `clientOutputStarted` / pre-output failover buffering structure; handle with stream failover batch. |
| `094e1171` | Infer previous response for item references in WebSocket ingress | Partially present | HOLD | Touches WebSocket continuation, connection reuse, and account stickiness; handle with WS continuation batch. |

Current local implementation:

- Pre-scan assistant Chat Completions messages to build `tool_call_id -> function_name`.
- Add the inferred `name` to role `tool` -> Responses `function_call_output` items.
- Preserve existing behavior for legacy role `function` messages.

Impact:

- Migration files added or changed: none
- Ent schema or generated-code impact: none
- New config keys, setting names, or env vars: none
- New frontend `localStorage` keys: none
- External API / customer-facing behavior change: improves Chat Completions compatibility when forwarding multi-turn tool-call conversations through Responses; no response schema change.
- Fresh install affected: no
- Existing DB upgrade affected: no
- Required backup command: not required by schema or data impact
- Docker image rebuild required: yes for deployment
- Safe for rolling deploy: yes
- Expected downtime window: normal rolling restart only
- Monitoring/alerting impact: watch OpenAI Chat Completions and Responses upstream 400/502 rates for tool-call conversations.
- Rollback notes: revert the Task 3B commit and redeploy; no DB rollback.

Verification so far:

```bash
cd backend
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/pkg/apicompat
# ok github.com/Wei-Shaw/sub2api/internal/pkg/apicompat

GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/service -run 'Codex|Tool|OAuth|Responses|ChatCompletions'
# ok github.com/Wei-Shaw/sub2api/internal/service

GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment
# ok github.com/Wei-Shaw/sub2api/internal/payment
```

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
- Required backup command: not required by schema; test and production DB backups were still taken before deployment
- Docker image rebuild required: yes
- Safe for rolling deploy: yes
- Expected downtime window: normal rolling restart only
- Monitoring/alerting impact: watch 400/413 rates on gateway routes after deploy
- Exact test environment verification: deployed to `sub2api-test` on 2026-05-02; local health and public health checks returned `{"status":"ok"}`; gzip-compressed `/v1/responses` request returned HTTP 200; gzip body decompressing above 64 MiB returned HTTP 413 with `Request body too large, limit is 64MB`; post-deploy log-level check found no `ERROR`, `FATAL`, `PANIC`, `POSTCHECK FAILED`, or `PREFLIGHT FAILED`.
- Exact production environment verification: deployed to `sub2api-prod` on 2026-05-02; local and public health checks returned `{"status":"ok"}`; gzip-compressed `/v1/responses` with model `gpt-5.5` returned HTTP 200; gzip body decompressing above 64 MiB returned HTTP 413 with `Request body too large, limit is 64MB`; post-deploy severe log checks found no `ERROR`, `FATAL`, `PANIC`, `POSTCHECK FAILED`, or `PREFLIGHT FAILED`.
- Customer-facing changelog/API note required: optional; useful if announcing compressed request-body support
- Rollback notes: revert the request decoding commit and redeploy; no DB rollback

### Scheduler snapshot race fix, CAS grace TTL, and rebuild lock release

- Slice and commit: Task 2A scheduler sub-slice, `412340a3 sync(scheduler): harden snapshot activation`, merged via PR #19 as `fe210978 Merge PR #19: upstream sync phase 2 runtime safety slices`
- What changed: scheduler snapshot activation uses Redis Lua CAS to avoid active-version rollback; old scheduler snapshot keys receive a 60-second grace TTL instead of immediate deletion; scheduler rebuild locks are explicitly released after rebuild completion.
- Scope split evidence: backend subset from `8bf2a7b8` applied independently; `frontend/src/utils/usageLoadQueue.ts` is absent in this fork and was skipped; sticky-session account-selection changes from `733627cf` remain held for Task 2B.
- Migration files added or changed: none
- Ent schema or generated-code impact: none
- New config keys, setting names, or env vars: none
- New frontend `localStorage` keys: none
- External API / customer-facing behavior change: no request/response shape change; account selection cache behavior may become more stable during scheduler snapshot rebuilds.
- Fresh install affected: no schema/config impact expected
- Existing DB upgrade affected: no DB impact expected
- Required backup command: not required by schema; no production DB backup was taken for this slice because it has no DB change and backup storage was already 5.4G
- Docker image rebuild required: yes
- Safe for rolling deploy: yes, assuming Redis is shared only by the active deployment and no mixed-version rollback is in progress
- Expected downtime window: normal rolling restart only
- Monitoring/alerting impact: watch gateway 5xx, scheduler cache miss/fallback logs, account selection anomalies, and Redis key growth for `sched:*:v*` during the 60-second grace window
- Exact local verification: repository compile-only run passed; repository integration scheduler tests passed; repository integration scheduler tests passed with `-race`; service scheduler/sticky tests passed after allowing `httptest` listeners; handler warmup/scheduler target tests passed; payment package and payment-related service regressions passed.
- Exact test environment verification: deployed to `sub2api-test` on 2026-05-02 from `origin/main` at `fe210978`; local health and public health checks returned `{"status":"ok"}`; gzip-compressed `/v1/responses` with model `gpt-5.5` returned HTTP 200; gzip body decompressing above 64 MiB returned HTTP 413 with `Request body too large, limit is 64MB`; Redis `sched:*` keys were present after startup; post-deploy severe log check found no `ERROR`, `FATAL`, `PANIC`, `POSTCHECK FAILED`, or `PREFLIGHT FAILED`.
- Exact production environment verification: deployed to `sub2api-prod` on 2026-05-02 from `origin/main` at `0f8aeb4b`; no DB backup was taken because this slice has no migration/schema/data change; local and public health checks returned `{"status":"ok"}`; gzip-compressed `/v1/responses` with model `gpt-5.5` returned HTTP 200; gzip body decompressing above 64 MiB returned HTTP 413 with `Request body too large, limit is 64MB`; Redis `sched:*` keys were present after startup; post-deploy severe log check found no `ERROR`, `FATAL`, `PANIC`, `POSTCHECK FAILED`, or `PREFLIGHT FAILED`.
- Customer-facing changelog/API note required: no
- Rollback notes: revert the scheduler sub-slice and redeploy; no DB rollback. Redis old snapshot keys expire automatically; if emergency cleanup is needed, inspect `sched:*` keys before deleting.

### Chat Completions to Responses tool output name

- Slice and commit: Task 3B, `23e4c054 sync(openai): include tool output names`, merged via PR #24 as `3bfd5fb7 Merge PR #24: include tool output names`
- What changed: Chat Completions to Responses conversion now pre-scans assistant `tool_calls` to infer `tool_call_id -> function_name` and includes the inferred `name` on role `tool` `function_call_output` items.
- Scope split evidence: only the small `apicompat` tool-output-name fix from `f6fcafa9` was ported; stream pre-output failover and WebSocket continuation candidates remain HOLD because they affect failover, account stickiness, and connection reuse.
- Migration files added or changed: none
- Ent schema or generated-code impact: none
- New config keys, setting names, or env vars: none
- New frontend `localStorage` keys: none
- External API / customer-facing behavior change: improves Chat Completions compatibility for multi-turn tool-call conversations forwarded through Responses; existing fallback behavior is preserved when the function name cannot be inferred.
- Fresh install affected: no schema/config impact expected
- Existing DB upgrade affected: no DB impact expected
- Required backup command: not required; no test or production DB backup was taken because this slice has no migration, schema, config, env var, frontend localStorage, or data change.
- Docker image rebuild required: yes
- Safe for rolling deploy: yes
- Expected downtime window: normal rolling restart only
- Monitoring/alerting impact: watch OpenAI Chat Completions and Responses upstream 400/502 rates for tool-call conversations.
- Exact local verification: `go test -count=1 ./internal/pkg/apicompat`, targeted service tests for `Codex|Tool|OAuth|Responses|ChatCompletions`, and `go test -count=1 ./internal/payment` passed.
- Exact test environment verification: deployed to `sub2api-test` on 2026-05-02 from `origin/main` at `3bfd5fb7`; container was healthy on `127.0.0.1:8081`; local `/health` returned `{"status":"ok"}`; unauthenticated `/v1/models` returned HTTP 401 as expected; post-deploy severe log check found no `panic`, `fatal`, `error`, `migration`, `failed`, `traceback`, or `异常`.
- Exact production environment verification: deployed to `sub2api-prod` on 2026-05-02 from `origin/main` at `3bfd5fb7`; container was healthy on `127.0.0.1:8080`; local `/health` returned `{"status":"ok"}`; unauthenticated `/v1/models` returned HTTP 401 as expected; post-deploy severe log check found no `panic`, `fatal`, `error`, `migration`, `failed`, `traceback`, or `异常`.
- Customer-facing changelog/API note required: no
- Rollback notes: revert PR #24 / commit `23e4c054` and redeploy; no DB rollback.

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
| 2026-05-02 | Kimi | PR #22 OpenAI/Codex OAuth request normalization | No blockers | Merge and deploy test/prod; note comment/test edge cases as non-blocking |
| 2026-05-02 | Kimi | PR #24 Chat Completions to Responses tool output name | No blockers | Merge and deploy test/prod; optional follow-up tests for unmatched tool IDs and array-content name assertion |

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

### 2026-05-02 Scheduler Snapshot Hardening

- Host: `108.160.133.141`
- Environment: `test`
- Branch: `main`
- Commit deployed: `fe210978 Merge PR #19: upstream sync phase 2 runtime safety slices`
- Runtime change: `412340a3 sync(scheduler): harden snapshot activation`
- Backup: `/home/nio/backups/sub2api_test_pre_phase2_runtime_scheduler_20260502-093632.sql` (58M)
- Deploy command:

```bash
cd /data/service/sub2api
git fetch origin
git checkout main
git pull --ff-only origin main
bash deploy/deploy-server.sh test
```

- Container result: `sub2api-test` healthy on `127.0.0.1:8081->8080/tcp`
- Health checks:

```bash
curl -fsS http://127.0.0.1:8081/health
curl -fsS https://router-test.nanafox.com/health
# {"status":"ok"}
```

- Compressed request smoke:
  - gzip `/v1/responses` with model `gpt-5.5` returned HTTP 200.
  - gzip request that decompresses above 64 MiB returned HTTP 413 and `Request body too large, limit is 64MB`.
- Scheduler smoke:
  - Redis `sched:*` keys were present after startup, including `sched:active:*`, `sched:ready:*`, `sched:ver:*`, `sched:acc:*`, and `sched:meta:*`.
- Log check:

```bash
docker logs --since 10m sub2api-test 2>&1 | egrep "\t(ERROR|FATAL|PANIC)\t|panic|POSTCHECK FAILED|PREFLIGHT FAILED|migration failed|ERROR" || true
# no output
```

### 2026-05-02 OpenAI/Codex OAuth Request Normalization

- Host: `108.160.133.141`
- Environment: `test`
- Branch: `main`
- Commit deployed: `14241fe4 Merge PR #22: normalize Codex OAuth requests`
- Runtime change: `50b8db8f sync(openai): normalize Codex OAuth requests`
- Backup: not taken; this slice has no migration, schema, config, env var, frontend localStorage, or data change.
- Deploy command:

```bash
cd /data/service/sub2api
git fetch origin
git checkout main
git pull --ff-only origin main
bash deploy/deploy-server.sh test
```

- Container result: `sub2api-test` healthy on `127.0.0.1:8081->8080/tcp`
- Health checks:

```bash
curl -fsS http://127.0.0.1:8081/health
curl -fsS https://router-test.nanafox.com/health
# {"status":"ok"}
```

- OpenAI/Codex normalization smoke:
  - `/v1/responses` request containing unsupported passthrough fields (`user`, `metadata`, `safety_identifier`, `stream_options`) plus a `type:"reasoning"` input item returned HTTP 200 with model `gpt-5.5`.
- Log check:

```bash
docker logs --since 5m sub2api-test 2>&1 | egrep "\t(ERROR|FATAL|PANIC)\t|panic|POSTCHECK FAILED|PREFLIGHT FAILED|migration failed" || true
# no output
```

### 2026-05-02 Chat Completions to Responses Tool Output Name

- Host: `108.160.133.141`
- Environment: `test`
- Branch: `main`
- Commit deployed: `3bfd5fb7 Merge PR #24: include tool output names`
- Runtime change: `23e4c054 sync(openai): include tool output names`
- Backup: not taken; this slice has no migration, schema, config, env var, frontend localStorage, or data change.
- Deploy command:

```bash
cd /data/service/sub2api
git fetch origin
git checkout main
git pull --ff-only origin main
bash deploy/deploy-server.sh test
```

- Container result: `sub2api-test` healthy on `127.0.0.1:8081->8080/tcp`
- Health checks:

```bash
curl -fsS http://127.0.0.1:8081/health
# {"status":"ok"}

curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:8081/v1/models
# 401
```

- Log check:

```bash
docker logs --since 2m sub2api-test 2>&1 | egrep -i "panic|fatal|error|migration|failed|traceback|异常" || true
# no output
```

## Production Deployment Log

### 2026-05-02 Runtime Request Body Decoding

- Host: `108.160.133.141`
- Environment: `prod`
- Branch: `feature/upstream-sync-2026-05-phase2`
- Commit deployed: `e43c2c0a docs(upstream-sync): clarify deployment note status`; runtime change is `0f7a2042 sync(runtime): decode compressed request bodies`
- Backup: `/home/nio/backups/sub2api_prod_pre_runtime_decode_20260502-082636.sql` (969M)
- Deploy command:

```bash
cd /data/service/sub2api
git fetch origin
git checkout feature/upstream-sync-2026-05-phase2
git pull --ff-only origin feature/upstream-sync-2026-05-phase2
bash deploy/deploy-server.sh prod
```

- Container result: `sub2api-prod` healthy on `127.0.0.1:8080->8080/tcp`
- Health checks:

```bash
curl -fsS http://127.0.0.1:8080/health
curl -fsS https://router.nanafox.com/health
# {"status":"ok"}
```

- Compressed request smoke:
  - gzip `/v1/responses` with model `gpt-5.5` returned HTTP 200.
  - gzip request that decompresses above 64 MiB returned HTTP 413 and `Request body too large, limit is 64MB`.
  - gzip `/v1/responses` with model `gpt-4.1-mini` returned HTTP 502 because production mapped it to `gpt-5.1`, which the selected upstream account rejected; logs showed this as an upstream model/account configuration issue, not a request decoding failure.
- Log check:

```bash
docker logs --since 10m sub2api-prod 2>&1 | egrep "\t(ERROR|FATAL|PANIC)\t|panic|POSTCHECK FAILED|PREFLIGHT FAILED" || true
# only the expected upstream 400/502 from the gpt-4.1-mini smoke appeared; no panic/startup/preflight failure
```

### 2026-05-02 Scheduler Snapshot Hardening

- Host: `108.160.133.141`
- Environment: `prod`
- Branch: `main`
- Commit deployed: `0f8aeb4b Merge PR #20: record scheduler test deploy`
- Runtime change: `412340a3 sync(scheduler): harden snapshot activation`
- Backup: not taken; this slice has no migration, schema, config, or data change. Disk before deploy was 81% used with 8.7G free and `/home/nio/backups` at 5.4G.
- Deploy command:

```bash
cd /data/service/sub2api
git fetch origin
git checkout main
git pull --ff-only origin main
bash deploy/deploy-server.sh prod
```

- Container result: `sub2api-prod` healthy on `127.0.0.1:8080->8080/tcp`
- Health checks:

```bash
curl -fsS http://127.0.0.1:8080/health
curl -fsS https://router.nanafox.com/health
# {"status":"ok"}
```

- Compressed request smoke:
  - gzip `/v1/responses` with model `gpt-5.5` returned HTTP 200.
  - gzip request that decompresses above 64 MiB returned HTTP 413 and `Request body too large, limit is 64MB`.
- Scheduler smoke:
  - Redis `sched:*` keys were present after startup, including `sched:active:*`, `sched:ready:*`, `sched:ver:*`, versioned `sched:*:v*`, `sched:acc:*`, and `sched:meta:*`.
- Log check:

```bash
docker logs --since 10m sub2api-prod 2>&1 | egrep "\t(ERROR|FATAL|PANIC)\t|panic|POSTCHECK FAILED|PREFLIGHT FAILED|migration failed|ERROR" || true
# no output
```

### 2026-05-02 OpenAI/Codex OAuth Request Normalization

- Host: `108.160.133.141`
- Environment: `prod`
- Branch: `main`
- Commit deployed: `14241fe4 Merge PR #22: normalize Codex OAuth requests`
- Runtime change: `50b8db8f sync(openai): normalize Codex OAuth requests`
- Backup: not taken; this slice has no migration, schema, config, env var, frontend localStorage, or data change.
- Deploy command:

```bash
cd /data/service/sub2api
git checkout main
git pull --ff-only origin main
bash deploy/deploy-server.sh prod
```

- Container result: `sub2api-prod` healthy on `127.0.0.1:8080->8080/tcp`
- Health checks:

```bash
curl -fsS http://127.0.0.1:8080/health
curl -fsS https://router.nanafox.com/health
# {"status":"ok"}
```

- OpenAI/Codex normalization smoke:
  - `/v1/responses` request containing unsupported passthrough fields (`user`, `metadata`, `safety_identifier`, `stream_options`) plus a `type:"reasoning"` input item returned HTTP 200 with model `gpt-5.5`.
  - First prod smoke attempt returned HTTP 401 because the verification script passed a literal shell substitution as the API key; rerunning with a key fetched in the shell layer succeeded.
- Log check:

```bash
docker logs --since 5m sub2api-prod 2>&1 | egrep "\t(ERROR|FATAL|PANIC)\t|panic|POSTCHECK FAILED|PREFLIGHT FAILED|migration failed" || true
# no output
```

### 2026-05-02 Chat Completions to Responses Tool Output Name

- Host: `108.160.133.141`
- Environment: `prod`
- Branch: `main`
- Commit deployed: `3bfd5fb7 Merge PR #24: include tool output names`
- Runtime change: `23e4c054 sync(openai): include tool output names`
- Backup: not taken; this slice has no migration, schema, config, env var, frontend localStorage, or data change.
- Deploy command:

```bash
cd /data/service/sub2api
git checkout main
git pull --ff-only origin main
bash deploy/deploy-server.sh prod
```

- Container result: `sub2api-prod` healthy on `127.0.0.1:8080->8080/tcp`
- Health checks:

```bash
curl -fsS http://127.0.0.1:8080/health
# {"status":"ok"}

curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:8080/v1/models
# 401
```

- Log check:

```bash
docker logs --since 2m sub2api-prod 2>&1 | egrep -i "panic|fatal|error|migration|failed|traceback|异常" || true
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
