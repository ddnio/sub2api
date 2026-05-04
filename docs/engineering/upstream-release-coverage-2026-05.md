# Upstream Release Coverage 2026-05

This document replaces ad-hoc continuation notes with a release-bounded coverage ledger. Releases are sequential gates: do not mark a release complete, update the fork release/tag marker, or advance to the next release until every upstream change in that release has a final recorded outcome.

The final outcome for a release item is one of:

- merged into the fork
- adapted into the fork through a deliberate local implementation
- already present locally with evidence
- intentionally skipped because it does not apply to the fork
- explicitly rejected or long-term frozen with a decision owner and reason

Implementation still moves by upstream PR/merge commit or smaller reviewed hunks. The release is the planning and closeout boundary, not a license to cherry-pick an entire tag.

## Baseline

- Local base: `origin/main` at `eb968255 docs(upstream-sync): reset release gate plan (#46)`.
- Upstream published-tag scope: latest local upstream tag `v0.1.121` at `9d801595 test: 更新管理员设置契约字段`.
- Upstream main observed locally: `b2bdba78 stabilize image request handling`; this is not in the published tag scope yet.
- Latest upstream tag in scope: `v0.1.121`.
- Work branch: `sync/release-v0.1.111-completion`.
- Worktree: `.claude/worktrees/release-gate-recheck-v0.1.111-v0.1.114`.
- Plan: `docs/plans/2026-05-03-upstream-release-sync-reset.md`.

If `upstream/main` advances, do not expand this ledger automatically. Only expand after a new upstream tag is fetched and the release interval is reviewed.

Fetch note: 2026-05-03 attempts to refresh `upstream` and `origin` failed with GitHub transport errors (`HTTP2 framing layer` / `Operation timed out`). Current decisions use the already present local refs above and the local upstream tags through `v0.1.121`. Push/fetch fallback order for this environment is: retry with `git -c http.version=HTTP/1.1`; if HTTPS still fails or hangs, use the explicit SSH URL (`git@github.com:ddnio/sub2api.git` for fork operations) or the GitHub Git Data API branch-creation fallback already documented in `docs/engineering/upstream-sync-2026-05-phase2.md`.

## Repeated-Issue Log

- GitHub HTTPS transport is unreliable in this environment. Do not spend multiple cycles retrying the same HTTPS fetch/push after `HTTP2 framing layer`, `Empty reply from server`, or timeout errors; switch to HTTP/1.1 once, then SSH or API fallback.
- Fork release marker tags use `fork/vX.Y.Z`, not upstream tag names. Upstream `vX.Y.Z` tags already exist and point at upstream commits.
- When inspecting upstream commits, copy full hashes from `git log --oneline --first-parent`; an abbreviated typo such as dropping one hex digit makes `git show` fail and wastes a cycle.
- In shell commands, avoid placeholder strings with angle brackets such as `<NULL>` unless they are safely quoted for the local shell; zsh can interpret them as redirection before SSH runs.
- Adding a lower-numbered migration after later migrations already exist is acceptable only when the runner sorts by filename and skips already-applied files by filename. In this repo, `applyMigrationsFS` sorts embedded `*.sql` names lexicographically and records each filename in `schema_migrations`, so a missing historical migration such as `097_*` can still be safely added before `098_*`.
- PR #40 CI note: pull-request integration run `25276045188` failed once in `TestOpsSystemLogSink_StartStopAndFlushSuccess` because the test signaled `done` before the sink goroutine incremented `writtenCount`; the same commit's push CI `25276036809` passed. Fix test races by waiting on the observed health condition, not by blindly rerunning.
- If a release item starts producing a broad hand-written diff, a large conflicted cherry-pick, or visible branch/worktree confusion, stop implementation and diagnose the process first. The correct next step is an import audit for the upstream commit/PR, not more manual code.

## Rules

- Do not merge or cherry-pick an entire release.
- Code merge unit is upstream first-parent commit / upstream PR merge commit. A smaller manually ported hunk is allowed only after the commit/PR has been tried or audited and the exact conflict/fork-divergence reason is recorded.
- Process releases in tag order. A later release must not start until the previous release gate is closed.
- Inside a release, use upstream first-parent commits only as the mainline entry index. For every upstream merge PR entry, also expand the second-parent branch/internal commits before claiming the release is complete. Do not jump from one reopened item to a later release, and do not start a broad hand-written implementation before the corresponding upstream commit/PR and its internal commits have a direct-import or already-present audit.
- Default fork PR packaging is one PR per upstream release gate. Keep all compatible items for the same release in one release worktree and one fork PR, with an itemized ledger and targeted tests. Split into separate PRs only for schema/migration changes, payment/auth/security/data-risk changes, very large or conflicted imports, or when a smaller PR is needed to unblock CI safely.
- After all items inside a release appear closed, run a release-level closeout review before updating the fork release marker/tag or starting the next release. The closeout must re-run both the upstream first-parent list and the full commit list, expand every merge PR's internal commits, confirm every commit/PR has a final outcome with evidence, confirm runtime tests and CI for changed code, decide whether deployment is required, and check that the ledger has no unresolved `HOLD`, `REOPENED`, `PORT`, or `PARTIAL` entries for that release.
- For each upstream commit/PR, use this order:
  1. Confirm whether the upstream commit is already an ancestor of the fork or already landed through a mapped fork PR.
  2. If not present, attempt or preview the direct upstream patch/cherry-pick in an isolated worktree.
  3. If the patch is clean and does not overwrite fork-specific product/data semantics, merge it as the upstream commit/PR unit.
  4. If it conflicts or crosses fork-specific architecture, split only the conflicted behavior into the smallest necessary subitems, with file-level evidence.
  5. Only then hand-port a subitem, reject it, or mark it adapted/present with tests or executable evidence.
- Large upstream PRs must not be converted directly into broad hand-written changes. First produce a subitem ledger showing: direct-portable files, fork-already-present behavior, conflict areas, data/schema impact, product-semantics impact, and required tests.
- A direct import attempt is not automatically a merge decision. If the patch applies but drags unrelated frontend/backend churn or overwrites fork-specific behavior, record that as a reason to split the PR into upstream-behavior subitems.
- `HOLD` is not a completed state by itself. A release with unresolved `HOLD` items is not complete. A `HOLD` item closes only when it is explicitly accepted for a later dedicated project, rejected for this fork, or long-term frozen with a documented owner/reason.
- Already merged fork PRs #19-#35 are treated as current implementation baseline.
- Product, schema/migration/Ent, auth identity, pending OAuth, affiliate, channel monitor/insights, Vertex, Fast/Flex, OpenAI image refactors, payment semantics, license/CLA, and sponsor/readme churn require explicit release-gate decisions.
- Docs-only PRs require `git diff --check`, self-review, Kimi review, and PR-level review; they do not deploy.
- Runtime work-package PRs require targeted tests, self-review, Kimi pre-commit review, PR-level Kimi review, and GitHub CI. Routine test/prod deployment is deferred until the whole release gate is complete, then performed once for that release. Exceptions are security hotfixes, migrations/schema changes, payment/auth/data-risk changes, or urgent production fixes.

## Ledger Semantics

This ledger PR is documentation-only in the fork. The upstream commits listed below are mostly code changes. Listing a code commit in this ledger does not authorize merging it; the `Action` column is the decision boundary.

- `MERGED`: the relevant behavior has already landed in this fork through a reviewed fork PR.
- `ADAPTED`: the upstream behavior or feature family was intentionally implemented through a fork-specific architecture instead of preserving the upstream commit ancestry.
- `PRESENT`: current fork code already provides the behavior; no new PR is planned.
- `PARTIAL`: some behavior exists or was ported, but more upstream code is not automatically accepted.
- `PORT`: a future small PR may port a proven low-risk missing behavior.
- `HOLD`: not complete; must be converted to accepted-later, rejected, or long-term-frozen before the release gate closes.
- `REJECTED`: explicitly not adopted for this fork/release, with reason.
- `FROZEN`: accepted only as a later dedicated project with owner, implementation plan, and an explicit decision not to block the current release marker. It is not a shortcut for risky items.
- `REOPENED`: previously marked closed under weaker rules, but now requires itemized audit or implementation before the fork can claim final alignment for that release.
- `SKIP`: intentionally ignore for this fork cycle, usually chore/version/churn or no proven local gap.

## Sequential Gate Status

The previous 2026-04 sync closed a slice-based Phase 1, not a release-gate sequence. Do not process the historical HOLD set as a separate global queue. Instead, start from the earliest release interval that was not closed under the release-gate rule, then decide each item inside that release before advancing.

Current gate:

| Gate | Status | Required next action |
| --- | --- | --- |
| `v0.1.110..v0.1.111` | Final | Historical marker `fork/v0.1.111` exists. PR #1538 and PR #1545 were reprocessed under the stricter rule through PR #48, #49, #50, and #51; no release-local unresolved item remains. |
| `v0.1.111..v0.1.112` | Final | Historical marker `fork/v0.1.112` exists. PR #40 closed the release-local migration gap and PR #41 marked fork coverage; re-confirmed after `v0.1.111` became final. |
| `v0.1.112..v0.1.113` | Local closeout pending review/deploy | Historical marker `fork/v0.1.113` exists. Branch `sync/v0.1.112-release-closeout` has reprocessed PR #1637, PR #1655, and PR #1666 under the stricter rule with local tests passing; final gate still requires independent review, release PR CI, merge to `main`, fork marker/tag, and release-level deployment verification. |
| `v0.1.113..v0.1.114` | Provisional | PR #44 was merged and deployed, and marker `fork/v0.1.114` exists. Reconfirm only after the reopened earlier gates are closed. |
| `v0.1.114..v0.1.115` | Parked | A partial PR #1752 worktree exists with uncommitted changes. Do not rebase, recreate, or mark this gate complete until the parked work is preserved and earlier reopened gates are closed. |
| `v0.1.115..v0.1.116` | Unstarted | Eleven upstream first-parent commits exist in this interval. Do not start it until `v0.1.114..v0.1.115` is closed and reviewed. |
| `v0.1.116` and later | Blocked | Do not advance to later releases until earlier gates are closed in order. |

Existing `fork/v0.1.111` through `fork/v0.1.114` tags must not be moved or deleted. They are historical fork sync markers. Any correctness gap found during recheck is fixed forward on latest `main`.

Release closeout review checklist:

1. Re-run `git log --oneline --first-parent --reverse <previous-upstream-tag>..<current-upstream-tag>` and compare every mainline entry against the release ledger.
2. Re-run `git log --oneline --reverse <previous-upstream-tag>..<current-upstream-tag>` and record the full commit count for the release.
3. For every merge PR in the first-parent list, expand its internal commits with the merge parents or PR branch history and confirm each internal commit maps to the same final release decision.
4. Re-run any PR/internal-commit subitem list used inside broad upstream PRs and confirm each subitem has a final state.
5. Confirm there are no release-local `HOLD`, `REOPENED`, `PORT`, or `PARTIAL` entries left unless they have been converted to an explicit accepted-later/rejected/frozen decision that does not block the marker.
6. Confirm local tests, GitHub CI, and Kimi review exist for every runtime PR in the release.
7. If any release item changed runtime behavior, deploy once for the completed release after merge/tag. Follow the repository deployment checklist and explicitly decide/apply all required environments, including ToC test, ToC production, and ToB production when the change affects shared runtime code.
8. Confirm deployment notes and log checks exist for the release-level deployment when deployment is required.
9. Only after this closeout review passes, update the fork release marker and create/push the `fork/vX.Y.Z` tag.

Release-level deployment checklist:

1. Deploy only after the whole release gate is closed, reviewed, merged to `main`, and tagged with the corresponding `fork/vX.Y.Z` marker. Do not deploy every small adapted commit during release work unless it is a security hotfix, data-risk migration, payment/auth change, or urgent production fix.
2. Before deployment, confirm the merged `main` commit includes every accepted item for the release and that the ledger has no unresolved release-local `HOLD`, `REOPENED`, `PORT`, or `PARTIAL` entries.
3. If the release added migrations, back up the test database first and verify the migration runner will apply only the intended new filenames. For data migrations, inspect real test database shape before production. For `v0.1.112..v0.1.113`, migration `127_add_balance_notify_fields.sql` uses `ADD COLUMN IF NOT EXISTS`, so deployment preflight must verify existing `users` columns are not present with incompatible defaults/nullability before applying test/prod:

```sql
SELECT
  column_name,
  is_nullable,
  data_type,
  column_default
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name = 'users'
  AND column_name IN (
    'balance_notify_enabled',
    'balance_notify_threshold',
    'balance_notify_extra_emails',
    'balance_notify_threshold_type',
    'total_recharged'
  )
ORDER BY column_name;
```

Expected compatible shape after migration: `balance_notify_enabled BOOLEAN NOT NULL DEFAULT true`, `balance_notify_threshold NUMERIC/DECIMAL nullable`, `balance_notify_extra_emails TEXT NOT NULL DEFAULT '[]'`, `balance_notify_threshold_type VARCHAR/TEXT NOT NULL DEFAULT 'fixed'`, and `total_recharged NUMERIC/DECIMAL NOT NULL DEFAULT 0`. If any column exists with a different default or nullability, stop deployment and write a corrective migration instead of relying on `IF NOT EXISTS`.
4. Deploy test from the merged `main` commit:

```bash
ssh nio@108.160.133.141 "cd /data/service/sub2api && git fetch origin && git checkout main && git pull --ff-only origin main && bash deploy/deploy-server.sh test"
```

5. Verify test:

```bash
ssh nio@108.160.133.141 "curl -s http://127.0.0.1:8081/health"
ssh nio@108.160.133.141 "curl -s -o /tmp/sub2api-test-models.out -w '%{http_code}' http://127.0.0.1:8081/v1/models"
ssh nio@108.160.133.141 "docker ps --filter name=sub2api-test"
ssh nio@108.160.133.141 "docker logs --since 20m sub2api-test 2>&1 | egrep -i 'panic|fatal|error|migration|failed|traceback|异常' || true"
```

Expected smoke results: `/health` returns `{"status":"ok"}`, unauthenticated `/v1/models` returns `401`, the container is healthy, and the severe-log scan has no unexplained runtime/migration errors.

6. For frontend-visible releases, perform a real browser/UI check on test for the changed screens before production. For payment/auth/data-risk releases, follow the domain-specific test plan in addition to the generic smoke checks.
7. Deploy production only after test passes:

```bash
ssh nio@108.160.133.141 "cd /data/service/sub2api && git checkout main && git pull --ff-only origin main && bash deploy/deploy-server.sh prod"
```

8. Verify production:

```bash
ssh nio@108.160.133.141 "curl -s http://127.0.0.1:8080/health"
ssh nio@108.160.133.141 "curl -s -o /tmp/sub2api-prod-models.out -w '%{http_code}' http://127.0.0.1:8080/v1/models"
ssh nio@108.160.133.141 "docker ps --filter name=sub2api-prod"
ssh nio@108.160.133.141 "docker logs --since 20m sub2api-prod 2>&1 | egrep -i 'panic|fatal|error|migration|failed|traceback|异常' || true"
```

9. Record the deployment evidence in this ledger or a release-specific deploy log: release gate, fork tag, merged commit, test/prod deploy time, local verification commands, migration/backup notes, health results, `/v1/models` 401 results, UI/manual checks, severe-log scan result, and rollback note.
10. Check the general deployment policy in `docs/engineering/deployment.md` before release rollout. If the release affects shared runtime/API/frontend behavior, deploy and verify ToB production (`fx.nanafox.com` on `43.106.8.109`) after ToC test and ToC production unless a release-specific note proves the change is ToC-only. Record the ToB decision and evidence together with the ToC deploy log.

## CI Baseline Closeout

CI drift independent of release coverage was fixed in PR #36 and merged to `origin/main` at `fed065e6`.

Local verification included:

```bash
cd backend && make test-unit
# passed

cd backend && GOCACHE="$PWD/../../.cache/go-build" go test -tags=unit ./internal/server/middleware ./internal/service
# passed

pnpm --dir frontend install --frozen-lockfile
pnpm --dir frontend typecheck
pnpm --dir frontend build
# passed; build emitted existing Vite chunk/dynamic import warnings only

git diff --check
# passed
```

Kimi review:

- First evidence pack timed out in wire transport after partial streaming.
- Smaller evidence pack returned `NO BLOCKERS`; runtime parser marked the prose result `INCONCLUSIVE`.
- PR-level Kimi retry attempts timed out without content; no blocker text was returned.

GitHub verification:

- PR #36 passed `test`, `golangci-lint`, `backend-security`, and `frontend-security`.
- The CI baseline gate is closed for this release-coverage PR.

## Fork PRs Already Mapped Into This Ledger

| Fork PR | Commit | Coverage meaning |
| --- | --- | --- |
| #19 | `fe210978` | Runtime safety: compressed request decoding and scheduler snapshot activation from v0.1.120 vicinity. |
| #22 | `14241fe4` | OpenAI/Codex request normalization covering parts of PR #2005, #2058, and #2068. |
| #24 | `3bfd5fb7` | Chat Completions to Responses tool output names from `f6fcafa9`. |
| #26 | `244c3f15` | Anthropic Claude Code prompt-cache preservation; related v0.1.119/v0.1.120 Anthropic fixes audited. |
| #28/#29 | `86a76164` / `6d943c6b` | Admin/frontend low-risk batch; follow-up hardening. |
| #30 | `7cfaf250` | Sticky-session false reject fix for scheduler snapshot accounts. |
| #31/#32 | `9308f805` / `520e0677` | Continuation review flow and ledger refresh. |
| #33 | `2c36c421` | Scheduler metadata keeps slim group membership snapshots from `733627cf`. |
| #34 | `682cee12` | OpenAI WS item-reference guard rails from `094e1171`. |
| #35 | `59b9cf34` | Deployed-slice closeout documentation. |
| #36 | `fed065e6` | CI baseline restored before release coverage closeout. |
| #37 | `2acdfd66` | Initial release coverage ledger; later corrected to start from `v0.1.110..v0.1.111`. |
| #39 | `d2a3e5a9` | Fork marker bumped to `0.1.111`; sync tag `fork/v0.1.111` points at this merged commit. |
| #40 | `fbaa1fdd` | `v0.1.112` migration gate closed by adding migration `097_fix_settings_updated_at_default.sql` and regression coverage. |
| #41 | `1d436745` | Fork marker bumped to `0.1.112`; sync tag `fork/v0.1.112` points at this merged commit. |

## Release Coverage Matrix

### v0.1.111

Range: `v0.1.110..v0.1.111`.

Source command:

```bash
git log --oneline --first-parent --reverse v0.1.110..v0.1.111
```

| Upstream source | Area | Local state | Outcome | Evidence / decision |
| --- | --- | --- | --- | --- |
| `f54e9d0b` | README churn | Commit is ancestor of local `HEAD`. | SKIP | Documentation/churn only; no release-gate behavior. |
| `0d69c0cd` | Version sync to `0.1.110` | Commit is ancestor of local `HEAD`; this was the historical upstream version stamp for the interval. | PRESENT | Historical note only. The fork has since advanced through `fork/v0.1.114`; do not read this row as the current checkout VERSION. |
| `155d3474` | Sponsors churn | Commit is ancestor of local `HEAD`. | SKIP | Sponsor/readme churn does not affect fork runtime or product behavior. |
| `1b79f6a7` / PR #1522 | Redis scheduler snapshot metadata and large MGET chunking | Commit is ancestor of local `HEAD`. | MERGED | Local `scheduler_cache.go` contains chunked `MGet` and preserves metadata fields such as `LoadFactor`. |
| `74302f60` / PR #1010 | OIDC login | Commit is ancestor of local `HEAD`. | MERGED | Current fork already includes OIDC config/public settings plumbing from this upstream family. |
| `9a72025a` / PR #1523 | Include `home_content` URL in CSP `frame-src` | Commit is ancestor of local `HEAD`. | MERGED | `SettingService.GetFrameSrcOrigins` adds `settings.HomeContent` before purchase/custom-menu origins. |
| `760cc7d6` / PR #1481 | Increase stored error-log body limit | Commit is ancestor of local `HEAD`. | MERGED | Local ops service has upstream-equivalent error-body/request-body sanitization; no further port required in this release. |
| `bbc79796` / PR #1529 | Group `/v1/messages` dispatch redo | Commit is ancestor of local `HEAD`. | MERGED | Local code has `OpenAIMessagesDispatchModelConfig`, group UI controls, migration `091_add_group_messages_dispatch_model_config.sql`, and dispatch resolution tests. |
| `00c08c57` / PR #1539 | Sync `load_factor` into scheduler cache | Commit is ancestor of local `HEAD`. | MERGED | `buildSchedulerMetadataAccount` copies `LoadFactor` into scheduler metadata snapshots. |
| `1ef3782d` / PR #1538 | Broad admin/repository/frontend bug-cleanup batch | Upstream merge commit is not an ancestor; direct whole-PR import was unsafe because it conflicted with fork settings/table/navigation surfaces. The PR was reprocessed by internal subitems A-I. | ADAPTED | Closed through PR #48, #49, and #50. Covered table scrollbar/account filters/settings/export filters, backend/frontend sort/search, axios security presence, messages-dispatch auth hydration, sidebar SVG color preservation, and applicable tests. Subitem ledger below has final states for A-I; no unresolved PR #1538 item remains. |
| `97f14b7a` / PR #1572 | Payment system v2 | Upstream merge commit is not an ancestor; fork intentionally replaced/adapted it through `623dda62` and the payment-b2 sequence through production hotfix `6518510b`. | ADAPTED | Payment-b2 audit and deploy logs show fork-specific migrations, provider instances, checkout/result flows, Stripe/Alipay/Wxpay providers, webhook/refund/resume tests, and test/prod deployment. Do not cherry-pick upstream payment v2 over the fork adaptation. |
| `54490cf6` / PR #1576 | Payment docs | Upstream merge commit is not an ancestor; upstream docs are superseded by fork payment-b2 operational docs. | ADAPTED | Current docs include `payment-b2-upstream-audit.md`, `payment-b2-deploy.md`, and `payment-b2-deploy-log.md`, which document the fork-specific payment architecture and deployment evidence. |
| `9b7b3755` / PR #1543 | Messages-dispatch i18n | Upstream merge commit is not an ancestor, but fork PR #9 imported the relevant i18n slice in `d80a3827`. | MERGED | `git log --all --grep 1543` maps PR #1543 to fork slice #9; local i18n keys for messages dispatch are present. |
| `16126a2c` / PR #1545 | Smooth sidebar collapse | Upstream merge commit is not an ancestor and the fork sidebar lacks equivalent smooth-collapse behavior. A direct commit import is too broad because upstream's hunk is based on a grouped sidebar/payment navigation structure that this fork does not use. | ADAPTED | Branch `sync/v0.1.111-pr1545-sidebar` ports only the compatible behavior: keep brand/link/button text DOM mounted, hide collapsed labels with `sidebar-*-collapsed` classes, animate sidebar width/header padding/link label width, and preserve fork-specific payment/referral/simple-mode navigation. Verification: `AppSidebar.spec.ts`, `app.spec.ts`, frontend typecheck, and `git diff --check` before merge. |
| `82b840c1` / PR #1587 | Anthropic 400 credit-balance handling | Upstream merge commit is not an ancestor, but fork PR #10 imported equivalent Anthropic handling in `a53527fa`. | MERGED | `ratelimit_service.go` disables Anthropic accounts on 400 bodies containing `credit balance`; fork slice #10 covered this family. |
| `a1a28368` | Sponsors churn | Not an ancestor after fork slices. | SKIP | Sponsor/readme churn; no fork behavior. |
| `9648c432` | Frontend TS2352 cast fix in API client | Upstream merge commit is not an ancestor, but equivalent code is present. | PRESENT | `frontend/src/api/client.ts` uses `apiResponse as unknown as Record<string, unknown>` and preserves `reason`/`metadata` for payment errors. |

Gate status: final. PR #39 bumped the fork release marker from `0.1.110` to `0.1.111` and merged at `d2a3e5a9`; annotated tag `fork/v0.1.111` points at that merged fork commit. That tag is retained as a historical marker. Release-level closeout was re-run after PR #51: the first-parent list for `v0.1.110..v0.1.111` has 17 entries and every row has a final outcome; PR #1538 subitems A-I are closed; PR #1545 is adapted; there are no release-local unresolved `HOLD`, `REOPENED`, `PORT`, or `PARTIAL` entries. Runtime changes in this release were already deployed during earlier PR #49 before the release-level deployment rule changed, and PR #50/#51 are covered by CI and targeted tests; no extra per-work-package deployment is recorded here.

Tag namespace note: do not create a fork tag named exactly `v0.1.111`. That tag name already exists for the upstream release and points at upstream commit `9648c432`; using the same tag name for a different fork commit would create a tag collision across remotes. Fork coverage tags use the `fork/vX.Y.Z` namespace.

Marker closeout: PR #39 bumped `backend/cmd/server/VERSION` to `0.1.111` and merged at `d2a3e5a9`; annotated tag `fork/v0.1.111` points at that merged fork commit.

#### PR #1538 import audit

Source: `1ef3782dd401d7efc0babee4b25ce00e5afcd6f2` / `Merge pull request #1538 from IanShaw027/fix/bug-cleanup-main`.

Upstream scope:

- `git diff --stat 1ef3782d^1 1ef3782d` reports 117 files, 3961 insertions, and 870 deletions.
- Backend impact is 66 files with 2196 insertions and 141 deletions.
- Frontend `src` impact is 49 files with 1226 insertions and 251 deletions.
- Largest touched groups by file count are repository query code, admin handlers, admin service/tests, admin views/API clients, frontend table components, settings/public config, and app/sidebar styling.
- The PR branch itself has useful internal commits and should be processed in that order after the merge commit proves unsafe as a whole:
  `fe211fc5` table scrollbar UI,
  `d8fa38d5` account status filter,
  `ad80606a` table page-size settings,
  `66e15a54` export filters,
  `5f8e60a1` backend table sort/search,
  lint/test repair commits,
  `13124059` public table setting fields,
  `b6bc0423` axios security upgrade,
  `67a05dfc` table defaults and dispatch preservation,
  `7dc7ff22` dispatch hydration preservation,
  and `f480e573` table default/sidebar color follow-up.
- Do not process `2b70d1d3` (`merge upstream main into fix/bug-cleanup-main`) as a portable feature subitem; it is branch maintenance noise inside the upstream PR branch.

Direct import evidence:

- A scratch cherry-pick into `.claude/worktrees/release-v0.1.111-audit` did not prove a safe direct import. The worktree has unresolved conflicts in:
  `backend/internal/handler/admin/setting_handler.go`,
  `backend/internal/server/api_contract_test.go`,
  `backend/internal/service/domain_constants.go`,
  `backend/internal/service/setting_service.go`,
  `backend/internal/service/setting_service_public_test.go`,
  `backend/internal/service/setting_service_update_test.go`,
  `frontend/src/components/common/Pagination.vue`,
  `frontend/src/composables/usePersistedPageSize.ts`,
  `frontend/src/i18n/locales/en.ts`,
  `frontend/src/i18n/locales/zh.ts`,
  `frontend/src/stores/app.ts`,
  `frontend/src/views/admin/SettingsView.vue`,
  plus add/add conflicts for `frontend/src/composables/__tests__/usePersistedPageSize.spec.ts`, `frontend/src/utils/tablePreferences.ts`, and `frontend/src/utils/__tests__/tablePreferences.spec.ts`.
- These conflicts are concentrated in fork-customized settings injection, table preference, API contract, and frontend settings areas. Those are product/config surfaces, so resolving the full cherry-pick as one PR would be too broad.
- A broad manual implementation attempt was exported to `/private/tmp/sub2api-v0.1.111-manual-sort-wip.patch` only as an investigation artifact, then removed from the worktree. Do not use that patch as implementation without subitem approval.

Already-present or adapted behavior in the current fork:

- Table page-size configuration exists through `SettingKeyTableDefaultPageSize`, `SettingKeyTablePageSizeOptions`, `SystemSettings`, `PublicSettings`, app-config injection, admin settings UI, `frontend/src/utils/tablePreferences.ts`, and tests. The fork defaults currently use `[10,20,50]` in public contract/tests while `InitializeDefaultSettings` still contains `[10,20,50,100]`, so the table page-size subitem must decide whether to keep fork behavior or normalize this mismatch.
- Pagination UI already reads configured page-size options and persists page-size selection through `Pagination.vue` and `usePersistedPageSize.ts`.
- Usage request-type filters and labels already exist in both admin and user usage views through the fork request-type work, so PR #1538 request-type behavior must be audited for deltas rather than blindly reimported.

Missing or unresolved behavior candidates:

- Generic backend `pagination.PaginationParams` still only has `Page` and `PageSize`; upstream PR #1538 added `SortBy`, `SortOrder`, `SortOrderAsc`, `SortOrderDesc`, `NormalizeSortOrder`, and raised `Limit()` from 100 to 1000.
- Upstream PR #1538 added `repository.paginateSlice`; the current fork does not have that helper.
- Upstream PR #1538 changed many repository list methods to accept or normalize sort/search/filter fields. Each repository must be checked against current fork behavior before code changes: accounts, announcements, API keys, channels, groups, promo codes, proxies, redeem codes, usage logs, and users.
- Frontend table loader/API query wiring changed across admin accounts, announcements, channels, groups, promo, proxies, redeem, usage, users, user keys, and user usage. Some fork views already use `getPersistedPageSize`; sort/search parity still needs itemized checks.
- Upstream PR #1538 includes sidebar and style edits, but PR #1545 separately owns smooth sidebar collapse. Sidebar behavior must be decided under PR #1545 to avoid mixing release items.

Decision for this gate:

- PR #1538 is not safe as a single upstream commit import.
- PR #1538 is also not allowed to remain a blanket `FROZEN` item.
- Close it through the PR branch's internal commits in order, using the planned subitems below. Low-risk adjacent subitems from the same upstream PR may share one fork PR when they touch the same product surface and can be reviewed/tested together; do not create one fork PR per upstream internal commit.
- Each subitem must end as `MERGED`, `ADAPTED`, `PRESENT`, or `REJECTED` with file/test evidence before `v0.1.111` can be final.

Implementation playbook for PR #1538:

1. Start from latest `origin/main` in a new runtime worktree. Do not continue from the conflicted scratch worktree.
2. For PR #1538, prefer the upstream PR branch's internal commit order over a hand-made split. For each internal commit, first try:
   `git cherry-pick -x <internal-commit>` in an isolated runtime worktree.
3. If the internal commit conflicts, narrow it to that commit's file group with:
   `git diff <internal-commit>^ <internal-commit> -- <file group>`,
   then run `git apply --3way --check`. If it is clean and preserves fork behavior, apply it and keep upstream code shape. If it conflicts, record the conflicted files and adapt only that internal-commit behavior.
4. Port upstream tests for the same internal commit before or with the behavior. If upstream tests are incompatible with fork architecture, write fork-equivalent tests and record why the upstream test was not copied.
5. Run the targeted tests for the touched group. Batch multiple internal commits into one fork PR when the behavior is adjacent and low-risk; keep high-risk or broad behavior such as backend sort/search, schema, auth, payment, and scheduling in separate PRs.
6. After merge and CI, record runtime impact for the release closeout. Do not deploy a routine work-package PR by itself; deploy test/prod once after the full release gate closes. Security hotfixes, migrations/schema changes, payment/auth/data-risk changes, and urgent production fixes remain exceptions.

Planned PR #1538 subitems:

| Subitem | Upstream internal commit / file group | Current fork read | Implementation rule | Required verification | Close condition |
| --- | --- | --- | --- | --- | --- |
| A. Table scrollbar UI | `fe211fc5`: `frontend/src/components/common/DataTable.vue`, `frontend/src/style.css` | Direct cherry-pick applied cleanly in branch `sync/v0.1.111-pr1538-scrollbar`; fork-specific cleanup only removed noisy comments/formatting and added a DataTable class regression test. | Batched with the first PR #1538 low-risk settings/export PR. This closes only the scrollbar behavior; it does not close PR #1538 or the release gate. | `pnpm --dir frontend exec vitest run src/components/common/__tests__/DataTable.spec.ts src/components/admin/account/__tests__/AccountTableFilters.spec.ts src/utils/__tests__/tablePreferences.spec.ts src/stores/__tests__/app.spec.ts` passed; `pnpm --dir frontend typecheck` passed. | `ADAPTED` after PR merge and CI, because upstream behavior is preserved with fork-quality cleanup and test coverage. |
| B. Account status filter | `d8fa38d5`: `frontend/src/components/admin/account/AccountTableFilters.vue` and its test | Direct cherry-pick applied cleanly in branch `sync/v0.1.111-pr1538-scrollbar`, but upstream deleted the existing fork test. The fork keeps and extends that test to cover the added `unschedulable` status option and existing privacy-mode behavior. | Batched with the first PR #1538 low-risk settings/export PR. This closes only the status filter behavior; it does not close PR #1538 or the release gate. | `AccountTableFilters.spec.ts` is included in the targeted frontend vitest above; `pnpm --dir frontend typecheck` passed. | `ADAPTED` after PR merge and CI, because upstream behavior is preserved without dropping fork test coverage. |
| C. Table page-size settings | `ad80606a`, `13124059`, `67a05dfc`, `f480e573` table-default portions: settings service/handler, API contract, app config, admin settings UI, table preference utils/tests. | Runtime branch normalizes the fork mismatch by aligning frontend fallback and API contract with backend initialization `[10,20,50,100]`; max page-size remains 1000. | Batched with the first PR #1538 low-risk settings/export PR because this is the same admin table settings surface and not schema/payment/auth behavior. | `GOCACHE=/Users/nio/project/nanafox/sub2api/.cache/go-build go test ./internal/server` passed; table preference/app store vitest passed; `pnpm --dir frontend typecheck` passed. | `ADAPTED` after PR merge and CI, because upstream's page-size option policy is aligned without replacing fork settings plumbing. |
| D. Export filters | `66e15a54`: account/proxy/redeem export handlers, tests, redeem API client | Runtime branch ports the current-fork-compatible filter gaps: account export now carries `group` and `privacy_mode`; redeem CSV export now carries `search`. Proxy export already respected current protocol/status/search filters. Sort propagation remains owned by Subitem E because current fork service/repository signatures do not yet support table sort. | Batched with the first PR #1538 low-risk settings/export PR. Do not import upstream sort params here; keep sort/search backend changes in Subitem E. | `GOCACHE=/Users/nio/project/nanafox/sub2api/.cache/go-build go test ./internal/handler/admin` passed; `pnpm --dir frontend typecheck` passed. | `ADAPTED` after PR merge and CI for current supported filters; sort parity remains open under Subitem E. |
| E. Backend/frontend table sort and search | `5f8e60a1` plus direct fixups `62962c05`, `b6946e78`, `269c7a06`: pagination primitives, repository sort/search, handlers, admin/frontend query wiring, tests. | Runtime branch `sync/v0.1.111-pr1538-backend-search-sort` ports the upstream sort/search contract across pagination, repository list queries, admin/user handlers, frontend API clients, and table views. The direct cherry-pick had conflicts in fork-modified stubs, API contract, i18n, and account/group/redeem views; those were resolved by preserving fork filters/settings while adding upstream sort/search parameters. | This branch keeps PR #1538 Subitem E separate from sidebar, dependency, dispatch, payment, auth, schema, and settings UI work. It also carries upstream lint/test fixups `62962c05`, `b6946e78`, and `269c7a06`. | Passed: `go test -tags=unit ./internal/pkg/pagination`; `go test -tags=unit ./internal/handler/admin`; `go test -tags=unit ./internal/handler -run 'Test.*Usage|Test.*APIKey'`; targeted `./internal/repository`; targeted `./internal/service`; `go test -tags=unit ./internal/handler/admin -run 'Test.*Export|Test.*AccountData|Test.*Redeem|Test.*ProxyData'`; `pnpm --dir frontend typecheck`; `pnpm --dir frontend exec vitest run src/components/admin/announcements/__tests__/AnnouncementReadStatusDialog.spec.ts src/views/user/__tests__/UsageView.spec.ts`; `git diff --check`. Expanded `go test -tags=unit ./internal/handler ./internal/service` still hits sandbox `httptest` bind denial in unrelated OIDC/proxy-quality tests. | `ADAPTED` after PR merge and CI. Runtime deploy was already performed before the release-level deployment rule changed; future routine runtime batches wait for full release closeout. |
| F. Frontend dependency security | `b6bc0423`: `frontend/package.json`, `frontend/pnpm-lock.yaml` axios upgrade | Current fork already has `axios` `^1.13.5` in `frontend/package.json`, and `frontend/pnpm-lock.yaml` resolves `axios@1.13.5`. | No code change. Keep dependency evidence in this closeout PR; do not create dependency-only churn. | `pnpm --dir frontend install --frozen-lockfile` is only needed if dependency files change; they do not in this PR. | `PRESENT`; axios security bump is already covered locally. |
| G. Messages dispatch hydration preservation | `7dc7ff22`: API key repository hydration tests/helpers | Fork messages dispatch was previously ported, but `GetByKeyForAuth` did not select `group.FieldMessagesDispatchModelConfig`, and `groupEntityToService` did not copy `MessagesDispatchModelConfig`. | This remainder branch ports only the missing hydration preservation behavior and upstream-equivalent repository tests. It does not touch dispatch routing, table sort, auth identity, or payment. | `GOCACHE=/Users/nio/project/nanafox/sub2api/.cache/go-build go test -tags=unit ./internal/repository -run 'TestGroupEntityToService_PreservesMessagesDispatchModelConfig|TestAPIKeyRepository_GetByKeyForAuth_PreservesMessagesDispatchModelConfig_SQLite'`. | `ADAPTED` after this PR merges and CI passes. |
| H. Sidebar SVG color follow-up | `f480e573` sidebar-color portions only | Current fork sidebar still had `.sidebar-svg-icon :deep(svg)` forcing `stroke: currentColor` and `fill: none`, which could strip uploaded SVG colors. PR #1545 still owns smooth collapse behavior. | This remainder branch ports only the SVG color-preservation hunk: keep inherited `color`, size/display constraints, and remove fill/stroke overrides. Collapse behavior remains PR #1545. | `pnpm --dir frontend exec vitest run src/components/layout/__tests__/AppSidebar.spec.ts`; frontend typecheck if broader frontend changes require it. | `ADAPTED` after this PR merges and CI passes. |
| I. Applicable upstream tests | All PR #1538 added tests. | Applicable tests have been handled with their owning subitems: DataTable/account filters/table preferences/app store tests in Subitems A-C, handler/repository/frontend tests in Subitems D-E, repository hydration tests in Subitem G, and AppSidebar style test in Subitem H. Broad upstream tests that assert replaced settings/sidebar structure are not copied as one dump. | Keep tests attached to the behavior they verify. Do not add a separate large upstream test import. | Same as owning subitems plus `git diff --check`. | `ADAPTED` after Subitems F-H are merged and the PR #1538 subitem ledger has no open rows. |

PR #1538 closeout:

- Subitem E (`5f8e60a1` plus direct sort/search fixups) has merged through PR #49 and was deployed before the release-level deployment rule changed.
- Subitems F-I merged through PR #50: F is present, G/H have narrow code/test fixes, I is handled by behavior-owned tests.
- PR #1538 is closed under the strict release rule.

### v0.1.112

Range: `v0.1.111..v0.1.112`.

Mainline source command:

```bash
git log --oneline --first-parent --reverse v0.1.111..v0.1.112
```

Full release audit commands:

```bash
git rev-list --count --first-parent v0.1.111..v0.1.112
git rev-list --count v0.1.111..v0.1.112
git log --oneline --reverse v0.1.111..v0.1.112
```

Audit result: 9 first-parent mainline entries and 17 total commits. The first-parent list is only the release entry index; the non-mainline commits inside upstream merge PRs are expanded below before the gate is treated as closed.

| Upstream source | Area | Local state | Outcome | Evidence / decision |
| --- | --- | --- | --- | --- |
| `ad64190b` | Version sync to `0.1.111` | Historical fork marker `fork/v0.1.111` exists. | PRESENT | PR #39 set `backend/cmd/server/VERSION` to `0.1.111` for the historical marker. The current checkout has advanced; do not treat this as current state. |
| `e70812f0` / PR #1623 | Anthropic buffered empty terminal output | Equivalent behavior is present through earlier fork Anthropic slice. | MERGED | `openai_gateway_messages.go` uses `apicompat.NewBufferedResponseAccumulator`, handles `response.done`, and calls `SupplementResponseOutput`; fork slice #10 includes `9d40fcaa` / PR #1623 mapping. |
| `7d80b5ad` / PR #1610 | Alipay/Wxpay base payment type mapping | Fork payment-b2 implements the mapping through provider instances and visible-method source selection. | ADAPTED | `DefaultLoadBalancer` can select across providers for base `alipay`/`wxpay`, `InstanceSelection` carries `ProviderKey`, and payment resume/source tests cover official/easypay routing. Do not cherry-pick upstream payment service code over fork payment-b2. |
| `75908800` / PR #1612 | QR code density | Equivalent frontend behavior is already present. | PRESENT | `PaymentQRDialog.vue` and `PaymentQRCodeView.vue` use `M` error correction with logos and `L` without logos; `PaymentStatusPanel.vue` uses `M`. |
| `d949acb1` / PR #1603 | DataTable mobile double render | Already landed through fork frontend slice. | MERGED | Fork commit `a845041a` maps PR #1603 and touched `DataTable.vue` plus `AccountUsageCell.vue`. |
| `ad6c3281` / PR #1575 | Cursor responses body compatibility | Already landed through fork Codex/Cursor slice. | MERGED | Fork commit `60f10e5b` includes `openai_codex_transform.go`, `openai_gateway_chat_completions.go`, and Cursor warmup tests for this family; `git log --all --grep 1575` also maps upstream PR #1575. |
| `66bea2b5` / PR #1624 | Version dropdown clipping | Fork applied a minimal sidebar-compatible fix instead of upstream sidebar churn. | ADAPTED | Fork commit `58c0f576` updates `AppSidebar.vue` and its spec for the expanded brand/version dropdown. This keeps the fork sidebar structure intact. |
| `92f4a6bb` | README/partner logo churn | Not product/runtime relevant for this fork gate. | SKIP | Documentation/logo sponsor churn; no local behavior. |
| `f9f57e95` | Restore `settings.updated_at` SQL default | Ported by PR #40. | MERGED | PR #40 added upstream `backend/migrations/097_fix_settings_updated_at_default.sql` and an integration assertion that final schema keeps `settings.updated_at DEFAULT now()`. Test/prod read-only checks showed both current databases already have `DEFAULT now()`, `is_nullable=NO`, `updated_at NULL count=0`, and already applied `098`/`111`; `097` is still absent there, so this is a compatibility/backfill marker for code and older instances rather than a current prod rescue. |

Expanded upstream PR/internal commit audit:

| Mainline entry | Internal commits | Local coverage |
| --- | --- | --- |
| `e70812f0` / PR #1623 | `a1e299a3` | Covered by fork Anthropic slice `a53527fa`; current `openai_gateway_messages.go` uses the buffered accumulator and supplements empty final output from prior deltas. |
| `7d80b5ad` / PR #1610 | `f498eb8f` | Covered through fork payment-b2 adaptation. Current payment load balancer/provider selection supports base `alipay`/`wxpay` across provider instances without replacing fork payment architecture. |
| `75908800` / PR #1612 | `24f0eebc` | Covered in current payment QR components: QR correction level is `M` with logos and `L` without logos, reducing QR density while preserving branded QR behavior. |
| `d949acb1` / PR #1603 | `3a113481`, `abe42675` | Covered by fork frontend slice `a845041a`; current `DataTable.vue` avoids hidden mobile table mounting and `AccountUsageCell.vue` lazy-loads mobile usage cells. |
| `ad6c3281` / PR #1575 | `b7edc3ed`, `422e25c9` | Covered by fork Codex/Cursor slice `60f10e5b`; current chat-completions and Codex transform code accepts Cursor Responses-shaped bodies and strips unsupported Responses parameters on the raw-body path. |
| `66bea2b5` / PR #1624 | `b9b52e74` | Covered by fork sidebar fix `58c0f576`; current expanded brand/version dropdown behavior is tested without importing unrelated upstream sidebar churn. |

Gate status: final. PR #40 merged the migration gate at `fbaa1fdd` after CI passed. PR #41 bumped `backend/cmd/server/VERSION` from `0.1.111` to `0.1.112` and merged at `1d436745`; annotated tag `fork/v0.1.112` points at that merged fork commit. Rechecked after `v0.1.110..v0.1.111` became final: the first-parent list for `v0.1.111..v0.1.112` has 9 mainline entries, the full release log has 17 commits, every merge PR's internal commits are mapped above, and every release row has a final outcome. There are no release-local unresolved `HOLD`, `REOPENED`, `PORT`, or `PARTIAL` entries.

Runtime/deploy note for `097`: this release gate contains a database migration file. Before any deployment of the merged PR, take the normal database backup. Current test/prod evidence indicates the migration should no-op on the live databases because the target default already exists, but it will still be recorded in `schema_migrations` on startup.

### v0.1.113

Range: `v0.1.112..v0.1.113`.

Source command:

```bash
git log --oneline --first-parent --reverse v0.1.112..v0.1.113
```

| Upstream source | Area | Local state | Outcome | Evidence / decision |
| --- | --- | --- | --- | --- |
| `e534e9ba` | Version sync to `0.1.112` | Historical fork marker `fork/v0.1.112` exists. | PRESENT | PR #41 set `backend/cmd/server/VERSION` to `0.1.112` for the historical marker. The current checkout has advanced; do not treat this as current state. |
| `d402e722` / PR #1637 | Websearch, balance notification, account pricing, and broad billing/settings changes | Not safely portable as a single raw cherry-pick, but fully reprocessed by internal commits/domains on branch `sync/v0.1.112-release-closeout`. | ADAPTED | Split decisions are closed locally: websearch emulation, balance/quota notification, account stats pricing, billing/settings API, payment residuals, frontend views, migrations, and mixed general-improvement commits. Verification: release-width backend targeted tests under non-sandbox permissions, frontend typecheck, payment/usage vitest batch, and `git diff --check`; still requires independent review and PR CI before merge. |
| `7c671b53` / PR #1635 | Version dropdown clipping | Current fork sidebar does not apply the upstream clipping style to the version badge wrapper. | PRESENT | `AppSidebar.vue` keeps the `VersionBadge` inside a plain `flex flex-col` wrapper, and `style.css` `.sidebar-header` has no `overflow-hidden` utility. The fork also uses a different sidebar DOM structure than upstream's `.sidebar-brand` path. The upstream commit `58c0f576` is not a fork ancestor, so do not cite it as fork evidence. |
| `9bf079b7` / PR #1655 | Payment fee multiplier | Upstream merge commit is not an ancestor, but fork payment-b2 already adapted the fee/multiplier semantics through the fork provider-instance/order/refund model. | ADAPTED | Current fork has `balance_recharge_multiplier`, `recharge_fee_rate`, `fee_rate`, `pay_amount`, proportional gateway refund calculation, refund/user-refund provider flags, order/result amount breakdown UI, and payment-flow tests. Do not cherry-pick PR #1655 over payment-b2; keep verifying with payment regression before release closeout. |
| `8fd29082` / PR #1663 | Abort account test stream when dialog closes | Already landed through fork slice. | MERGED | Fork commit `d80a3827` maps PR #1663 and updates both user and admin `AccountTestModal.vue` stream-close handling. |
| `1db32d69` / PR #1666 | Account cost display in usage/dashboard tables | Upstream merge commit is not an ancestor and cannot be copied literally because fork already has different payment/accounting history. The behavior is now adapted with upstream-compatible `account_stats_cost` override semantics: `COALESCE(account_stats_cost, total_cost) * COALESCE(account_rate_multiplier, 1)`. | ADAPTED | Local branch adds `account_cost` to dashboard aggregate tables, dashboard totals, model/group/user breakdown DTOs and UI, plus repository/service regression tests. Migrations `107_add_account_cost_to_dashboard_tables.sql` and `108_add_account_stats_pricing.sql` are schema-only; they intentionally do not scan/backfill `usage_logs` during startup. Release deployment must trigger `/api/v1/admin/dashboard/aggregation/backfill` or the existing recompute path for the desired historical window after the new columns exist. |
| `70d0569f` / PR #1668 | OpenAI rate-limit and usage scheduling fix | Already landed through fork OpenAI core slice. | MERGED | Fork commit `2ce67ca4` maps PR #1668 and updates account usage/rate-limit paths plus `openai_ws_ratelimit_signal_test.go`. |

Gate status: local closeout pending review/deploy. PR #42 completed the earlier decision matrix; PR #43 bumped `backend/cmd/server/VERSION` from `0.1.112` to `0.1.113` and merged at `32787ca4`; annotated tag `fork/v0.1.113` was pushed for the merged fork commit. That tag is retained as a historical marker. Current branch `sync/v0.1.112-release-closeout` closes the stricter PR #1637 reprocessing locally and revalidates adapted PR #1655/#1666 surfaces. The release is not production-final until this branch passes independent review, PR CI, merge to `main`, fork marker/tag handling, and release-level ToC/ToB deployment verification.

#### PR #1637 import audit

Source: `d402e722cf86e25a5fd1e7304bdc0752ccf41431` / `Merge pull request #1637 from touwaeriol/feat/websearch-notify-pricing`.

Direct import evidence:

- `git diff --stat d402e722^1 d402e722` reports 177 files, 13,643 insertions, and 1,201 deletions.
- A scratch whole-PR cherry-pick in `.claude/worktrees/audit-v0.1.113-import` produced broad conflicts across Ent/schema, settings, payment, routes, frontend views, and generated wire files. The scratch worktree is audit-only and must not be used as implementation.
- Therefore PR #1637 must be processed by internal commits / feature domains. This does not reduce coverage: every internal commit below remains tracked.

Internal commits in `d402e722^1..d402e722^2`:

| Domain | Internal commits | Current decision |
| --- | --- | --- |
| Branch maintenance / CI / conflict repair | `1cd033e5`, `37c23ecc`, `3d4d960d`, `49915987`, `1e6912ea`, `d6965b06`, `24e16b7f`, `b42f34c3`, `6a08efee`, `e8ee400a`, `3d202722` | REGISTERED | These commits are formatting, generated-wire repair, upstream branch maintenance, or CI repair. They do not become standalone fork features, but any touched behavior must still be covered by the functional domains below. `3d202722` is specifically adapted in current generated wiring: `wire_gen.go` calls `repository.ProvideSchedulerCache(redisClient, configConfig)`, so scheduler cache config is not silently ignored. |
| Channel restriction / channel model pricing base | `3de77130`, `2dce4306`, `160903fc`, `e3748741`, `794e8172`, `1c63ea14`, `58677dd5` | ADAPTED | Fork has channel model pricing and scheduler-stage restriction checks in `channel_repo_pricing.go`, `ChannelsView.vue`, `gateway_service.go`, and `openai_gateway_service.go`. The cache TTL/rebuild/error-cache behavior and restriction logging are present in `channel_service.go`, `gateway_service.go`, and `openai_gateway_service.go`. Branch `sync/v0.1.112-release-closeout` adapts channel websearch through `features_config` instead of upstream's incompatible `channels.features` column; `ChannelsView.vue` uses splice updates, writes false/deletes stale `web_search_emulation`, and the gateway reads channel `features_config` when account mode is default. Verification: `go test -tags=unit ./internal/service ./internal/handler/admin -run 'TestChannel|Test.*Channel|Test.*WebSearch|Test.*Restriction|TestResolveWebSearchProviderProxyURL|TestWebSearchProxyIDValue|TestResetWebSearchUsage'`. |
| Payment residuals and provider refund controls | `5bae3b05`, `3c884f8e`, `56e4a9a9`, `c738cfec`, `4aa0070e`, `f1297a36`, `c14d7393`, `5240b444`, `63f539b3` | ADAPTED | Fork payment-b2 already has provider instances, `allow_user_refund`, method limits, fee/pay amount separation, refund provider lookup, proportional gateway refund, Stripe payment type matching, inline/mobile payment flow, renewal modal behavior, and payment regression tests/deploy evidence. Keep as adapted unless a targeted diff finds a missing residual. |
| Websearch backend and settings | `1b53ffca`, `7fad9f60`, `fda61b06`, `60b0fa81`, `d0674e0f`, `889b5b4f`, `5df73099`, `49281bbe`, `1262654d`, `74f8a30f`, `9c09bd19`, `0a4ece5f`, `9e0d12d3`, `7c729293`, `9028d208` | ADAPTED | Branch `sync/v0.1.112-release-closeout` adapts the websearch base plus the settings/quota UI work: `backend/internal/pkg/websearch`, gateway emulation service, global provider config API, channel `features_config`, account/channel UI toggles, provider-level proxy tracking, proxy failover, proxy-error classification, subscribed-at monthly quota TTL, admin test endpoint with 15s timeout, quota usage display, collapsible provider cards, API-key show/copy controls for currently entered keys, hidden account/channel websearch controls when the global switch has no usable provider, and a modal admin test UI. The `1262654d` websearch tri-state subset is adapted: account `extra.web_search_emulation` supports `default` / `enabled` / `disabled`, legacy bool true still forces enabled, bool false/missing follows channel default, account disabled overrides channel enabled, create/edit account UI uses a tri-state select, and migration `126_migrate_websearch_emulation_to_tristate.sql` converts legacy bool data without touching existing string values. The `7c729293` quota/reset subset is adapted: admin config API accepts `quota_limit: null` as unlimited while still accepting legacy `0`, manager keeps internal `0` unlimited semantics, admin reset endpoint deletes provider quota keys, and Settings UI saves empty quota as `null`, shows `∞`, and exposes reset for nonzero usage. Later audit fixes add enabled-provider API-key validation and SSE write-error propagation. Verification: frontend typecheck, websearch targeted Go tests, release-width backend tests, and `git diff --check`. |
| Account stats pricing | `7535e312`, `80fa4844`, `11c46068`, `98c9d517`, `1262654d`, `a68df457`, `b7fb2e43`, `ca673f98`, `ed8a9d97`, `9d319cfa`, `594f0d17`, `9c09bd19`, `9028d208` | ADAPTED | Added fork migration `108_add_account_stats_pricing.sql`, `usage_logs.account_stats_cost`, channel account-stats pricing rules/intervals, service resolver priority, OpenAI/Anthropic usage-log write integration, account-cost aggregate formulas, admin channel API fields, and a basic UI entry. Verification so far: targeted service/repository tests and frontend typecheck passed. This closes the account-stats-pricing domain, but commits also touching websearch/notify remain open under those domains until implemented and release-level review confirms no shared leftovers. |
| Balance and quota notification | `b32d1a2c`, `c3812ce1`, `30b926ad`, `f694afbb`, `9e33d0c4`, `cef22c70`, `eba289a7`, `4e96a6fa`, `79d154ed`, `81287e96`, `42280751`, `61aa197b`, `915b7a4a`, `31550a2c`, `95f9b27e`, `48e8efe3`, `42f8ef33`, `2066c478`, `ac554432`, `7141dcee`, `216bda58`, `e27335ac`, `c1eb79e4`, `48b6c481`, `f571d8ff`, `6e9146e7`, `a43da622`, `ca673f98`, `ed8a9d97`, `9d319cfa`, `594f0d17`, `1b7c2951`, `74f8a30f`, `a9880ee7`, `0a4ece5f`, `b402c367`, `7c729293`, `9028d208` | ADAPTED | Current branch adapts the final upstream notification domain to the fork: additive migration `127_add_balance_notify_fields.sql`, Ent user fields, `NotifyEmailEntry`, `BalanceNotifyService`, saved-email verification cache methods, profile/admin/account settings APIs, billing/quota notify hooks, profile balance-notify UI, account quota notify controls, admin global settings, i18n, and regression tests. The implementation preserves fork payment/billing semantics: `total_recharged` is stored for schema parity but not wired into recharge totals in this release, and user-facing balance threshold editing remains fixed-value only. Verification so far: frontend typecheck, targeted notify/service/gateway tests, broader service/handler/server tests under normal permissions, and diff-check; rerun before final commit because the worktree is still dirty. |
| Frontend account/settings composition and usage queue | `a56151fe`, `3fa5b8bc`, `6ac8ccde`, `245f47ce`, `2066c478`, `ac554432`, `7141dcee`, `1b7c2951` | ADAPTED | Websearch/notify UI splits are adapted, `a56151fe` CapacityBadge extraction is ported while preserving fork quota badges, `3fa5b8bc` usage-load queue serializes Anthropic OAuth/setup-token usage checks grouped by proxy exit, and mixed frontend cleanups from `6ac8ccde`/`63f539b3` are either adapted or documented as not applicable to the fork's asset/deploy layout. Verification: frontend typecheck, usage queue vitest, payment frontend vitest batch, and release-width backend tests. |
| Messages routing compatibility | `8548a130` | PRESENT | Fork already includes the messages routing/subscription group behavior from earlier OpenAI/messages slices; confirm again during release closeout but do not reimport from PR #1637 wholesale. |

Notify import audit for PR #1637 final state:

- Final upstream state is the merge result at `d402e722`, not each intermediate notify commit in isolation. Intermediate `cef22c70` removed one earlier percentage-threshold shape, while later `eba289a7` reintroduced global visibility/percentage controls differently. Implementation must follow the final state and not blindly replay reverted intermediate UI/API shapes.
- Database impact is additive on `users`: `balance_notify_enabled BOOLEAN NOT NULL DEFAULT true`, `balance_notify_threshold DECIMAL(20,8) NULL`, `balance_notify_extra_emails TEXT NOT NULL DEFAULT '[]'`, `balance_notify_threshold_type VARCHAR(10) NOT NULL DEFAULT 'fixed'`, and `total_recharged DECIMAL(20,8) NOT NULL DEFAULT 0`. These columns do not rewrite existing balances or API keys, but they are runtime schema changes and require real DB precheck/backups before release deployment.
- Fork-specific schema note: local `users` already has referral edges/fields not present in upstream's notify-era schema. Ent schema/generated code must be regenerated from local schema with added notify fields, not copied wholesale from upstream generated files.
- Final setting keys to add are `balance_low_notify_enabled`, `balance_low_notify_threshold`, `balance_low_notify_recharge_url`, `account_quota_notify_enabled`, and `account_quota_notify_emails`. Existing ops email notification settings are separate and must not be reused as the product/user balance-notify config.
- Final backend service surface includes `BalanceNotifyService`, `NotifyEmailEntry`, parsing/marshalling verified-email lists, user notify-email verification cache methods, profile update fields, and quota notification checks after account quota increments. User-facing balance notifications must send only to verified, non-disabled user extra emails; admin account-quota alerts must use verified, non-disabled setting emails.
- Fork-specific billing note: local billing has both direct post-usage updates and `UsageBillingRepository.Apply` transaction paths. Notify hooks must attach to the actual quota/balance update path used by each gateway path, and must not double-send when both cache/deferred billing and DB transaction billing run.
- Frontend final state needs profile balance-notify controls, account quota notify toggles, admin global notify settings, recharge URL behavior, duplicate-email UX, and the email verification hint from `7c729293`. These should be adapted to the fork's existing profile/settings/account modal structure rather than copied as upstream file paths.
- Current branch implementation evidence: `backend/migrations/127_add_balance_notify_fields.sql` contains the final user columns and legacy email-list conversion guarded by `NOTICE`; `backend/internal/service/balance_notify_service.go` implements balance/quota crossing checks, verified/disabled recipient filtering, recharge URL email content, per-recipient timeout, panic recovery, and quota remaining-threshold semantics; `backend/internal/service/notify_email_entry.go` implements structured notify-email parsing/marshalling; `backend/internal/service/user_service.go`, `backend/internal/handler/user_handler.go`, and `backend/internal/server/routes/user.go` expose saved notify-email send/verify/toggle/remove flows; `frontend/src/components/user/profile/ProfileBalanceNotifyCard.vue`, `frontend/src/components/account/QuotaNotifyToggle.vue`, `frontend/src/components/account/QuotaDimensionRow.vue`, `frontend/src/composables/useQuotaNotifyState.ts`, and `frontend/src/views/admin/SettingsView.vue` adapt the profile/account/admin UI.
- Deployment rule: because this domain adds user columns and runtime email side effects, the completed release must deploy test then prod only after the full `v0.1.113` gate closes, with migration precheck on real test DB, DB backup before production, generic smoke checks, and manual UI/email-path checks where feasible.

v0.1.113 release deployment plan:

1. Do not deploy from this feature branch. Deployment starts only after PR #1637, PR #1655, and PR #1666 are all closed with final evidence, the release closeout checklist passes, the release PR is merged to `main`, and the fork release marker is tagged.
2. Before test deployment, inspect the real test database for the intended migration state: current `schema_migrations`, existing `users` notify columns, `settings.updated_at` default state, and any already-applied dashboard/account-stats migrations. Back up the test database before applying the merged release.
3. Deploy ToC test from merged `main`, then verify `/health`, unauthenticated `/v1/models` returns `401`, container health, severe logs, and the new runtime surfaces: profile balance-notify settings, notify-email send/verify/remove/toggle path, admin global notify settings, account quota notify controls, and websearch/account-stats regression screens.
4. Before ToC production deployment, take a production database backup and repeat the migration precheck against production so the deploy applies only the intended new release migrations. If test showed migration drift or email-side-effect errors, stop before production.
5. Deploy ToC production only after test passes, then repeat health, `401`, container, severe-log, and changed-screen smoke checks.
6. Because this release changes shared backend/frontend runtime behavior, run the ToB production applicability check from `docs/engineering/deployment.md`. If the ToB service uses this same code path, deploy `main` there after ToC production and repeat health/container/severe-log checks plus a minimal changed-screen/API smoke check. If not deployed, record the concrete reason.
7. Record backup path, deployed commit, fork tag, migration filenames applied, ToC test/prod evidence, ToB deploy-or-skip evidence, manual UI/email checks, and rollback note in this ledger or a release-specific deploy log.

Next implementation order for PR #1637:

1. Finish websearch follow-up commits after the base import: proxy failover/timeout/quota hardening, admin test timeout, tri-state behavior, API key visibility/copy controls, and quota cache fixes.
2. Implement balance/quota notification backend, then frontend profile/account settings.
3. Finish payment residual proof with tests, then keep it closed as `ADAPTED` unless a targeted residual is found.
4. Re-run the full PR #1637 internal commit coverage audit, especially commits that touched multiple domains (`1262654d`, `a68df457`, `b7fb2e43`, `ed8a9d97`, `9d319cfa`, `74f8a30f`, `0a4ece5f`, `9028d208`).
5. Local code closeout for these domains is complete. Before production-final status, complete independent review, PR CI, merge/tag, and release-level deployment verification.

#### PR #1637 internal commit coverage snapshot

This snapshot was refreshed while closing the `d402e722` PR #1637 reprocessing on branch `sync/v0.1.112-release-closeout`. It is a local release closeout snapshot: rows below are final local outcomes unless explicitly marked otherwise. Production-final status still requires independent review, PR CI, merge, tag, and deployment verification.

| Commit | Area | Current state | Evidence / remaining action |
| --- | --- | --- | --- |
| `1cd033e5` | Formatting | ADAPTED | Final touched Go files are gofmt'd and `git diff --check` passes. |
| `3de77130` | Channel pricing UI/debug | ADAPTED | `ChannelsView.vue` already updates model-pricing entries with `splice`, preserving Vue reactivity. The upstream console debug line is intentionally not kept as production UI behavior. Verification covered by frontend typecheck and channel targeted tests. |
| `2dce4306` | Channel restriction scheduling | COVERED | Fork already has scheduler-stage restriction behavior in gateway scheduling paths. |
| `160903fc` | Channel restriction review fixes | COVERED | Covered with the restriction scheduling behavior above. |
| `e3748741` | Channel cache/logging | ADAPTED | `channel_service.go` uses a 10-minute cache TTL, short error-cache TTL, `context.WithoutCancel` for cache rebuilds, and active rebuild on invalidation; gateway paths log channel pricing restriction blocks. Verification: targeted channel/restriction tests passed. |
| `37c23ecc` | Formatting | ADAPTED | Final touched Go files are gofmt'd and `git diff --check` passes. |
| `794e8172` | Channel feature model | ADAPTED | Fork uses `features_config` only; upstream `channels.features` is intentionally not imported because this fork has no matching schema contract. |
| `3d4d960d` | Formatting | ADAPTED | Final touched Go files are gofmt'd and `git diff --check` passes. |
| `1c63ea14` | Channel feature query | ADAPTED | Queries now include `features_config`; no `features` column is queried. |
| `5bae3b05` | Payment audit fixes | COVERED | Payment-b2 already covers provider/method semantics. |
| `3c884f8e` | Payment tests | COVERED | Payment-b2 regression coverage exists; rerun payment regression during release closeout. |
| `56e4a9a9` | Payment/frontend audit fixes | COVERED | Payment-b2 and frontend payment flow already adapted. |
| `c738cfec` | Payment security/idempotency | COVERED | Payment-b2 includes provider lookup, amount breakdown, refund, and idempotency behavior. |
| `1b53ffca` | Websearch base | ADAPTED | Current branch imports/adapts websearch package, gateway emulation, config API/UI, account/channel toggles, and tests. |
| `7fad9f60` | API contract field | ADAPTED | API contract and settings DTO include `web_search_emulation_enabled`. |
| `7535e312` | Account stats pricing | ADAPTED | Closed by account-stats-pricing branch work and migration `108_add_account_stats_pricing.sql`. |
| `fda61b06` | Websearch failover/quota | ADAPTED | Websearch manager uses quota-weighted provider ordering, provider/account proxy resolution, proxy unavailable tracking, quota rollback, invalid-proxy erroring without direct fallback, and resettable quota usage. Fork keeps subscribed-at monthly TTL instead of upstream refresh-interval strings. Verification: `go test -tags=unit ./internal/pkg/websearch ./internal/service -run 'Test.*WebSearch|TestIsProxyError|TestSelectByQuotaWeight|TestNewHTTPClient|TestResolveProxyID|TestIsProviderAvailable|TestIsProxyAvailable|TestManager_ResetUsage'`. |
| `49915987` | Websearch formatting | ADAPTED | Websearch code is gofmt'd and covered by targeted websearch tests plus `git diff --check`. |
| `60b0fa81` | Websearch proxy error tests | ADAPTED | `isProxyError` treats network, SOCKS, TLS handshake/certificate, DNS, timeout, and connection-refused failures as proxy/network errors while preserving API-rate-limit errors as provider failures; manager tests cover proxy ID, proxy availability, provider availability, quota weighting, and HTTP client proxy validation. |
| `b32d1a2c` | Balance/quota notify base | ADAPTED | Added fork migration `127_add_balance_notify_fields.sql`, Ent user fields/generated code, `BalanceNotifyService`, settings/user DTOs, profile/admin/account UI, and notify tests while preserving fork billing and payment semantics. |
| `c3812ce1` | Notify review fixes | ADAPTED | Notify recipient parsing/dedup and quota formula behavior are adapted in `NotifyEmailEntry` and `BalanceNotifyService`; account-stats formula fixes are covered by the account-stats-pricing domain. |
| `30b926ad` | Notify email timeout/removal | ADAPTED | `BalanceNotifyService` sends per-recipient emails with timeout/panic recovery, and `UserService.RemoveNotifyEmail` returns the updated user profile state. |
| `d0674e0f` | Websearch settings/quota UI | ADAPTED | Replaced provider priority/refresh interval with subscribed-at monthly quota TTL, added quota usage sanitization, admin test endpoint, collapsible provider cards, and entered-key show/copy UI. Fork uses `SettingService.RebuildWebSearchManager` instead of upstream `server/http.go` builder. Verification: frontend typecheck, targeted websearch Go tests, and diff-check. |
| `f694afbb` | Notify percentage threshold | ADAPTED | Final schema keeps `balance_notify_threshold_type` and backend can resolve percentage thresholds for compatibility, but the fork UI only exposes fixed threshold editing in this release because `total_recharged` is not yet wired to recharge/payment totals. |
| `9e33d0c4` | Websearch/notify audit | ADAPTED | Websearch manager/config audit fixes and notify service/repository defaults are adapted. `SaveWebSearchEmulationConfig` now merges saved API keys before rejecting enabled providers without a configured key, new registrations explicitly preserve balance-notify defaults, and default settings include notify keys without enabling email side effects. Verification: targeted websearch/notify service tests and backend release-width tests. |
| `cef22c70` | Notify threshold rollback | ADAPTED | The fork follows the final upstream state rather than blindly replaying the reverted intermediate UI/API shape; global/admin percentage threshold editing is not exposed. |
| `889b5b4f` | Websearch UI disabled state | ADAPTED | Settings UI uses compact provider controls and modal search testing. Account create/edit and channel form hide websearch controls unless the global config is enabled with at least one provider. This is a fork adaptation of the upstream UI diff, not a raw cherry-pick. Verification: frontend typecheck and diff-check. |
| `eba289a7` | Notify global toggles | ADAPTED | Admin settings now carry `balance_low_notify_enabled`, `balance_low_notify_threshold`, `balance_low_notify_recharge_url`, `account_quota_notify_enabled`, and `account_quota_notify_emails`; public settings expose the user-visible balance/quota notify flags without leaking admin recipient list. |
| `4e96a6fa` | Notify/websearch/security audit | ADAPTED | Notify recipient filtering, saved-email verification, per-recipient send hardening, WebSearch config validation, and Stripe CSP coverage are adapted in fork code. Verification: notify/websearch/security-header targeted tests and frontend typecheck. |
| `80fa4844` | Account-stats UI placement | COVERED | Account-stats pricing UI is integrated into channel platform tabs. |
| `5df73099` | Websearch admin test timeout | ADAPTED | `service.TestWebSearch` now wraps admin test searches in a 15s context timeout while continuing to bypass quota. Verification covered by targeted websearch Go tests. |
| `49281bbe` | Websearch API key visibility | ADAPTED | Hide show/copy controls when the API key input is empty. Later `9e0d12d3` restores visibility controls for saved providers whose API key is configured but sanitized out of the form. |
| `79d154ed` | Notify public settings fields | ADAPTED | Public/admin settings DTOs and mappers expose balance/quota notify flags and recharge URL; admin recipient list remains admin-only. |
| `81287e96` | Balance notify UX | ADAPTED | `ProfileBalanceNotifyCard.vue` adds enable/threshold/email verification controls adapted to the fork profile page. |
| `42280751` | Notify duplicate email UX | ADAPTED | Profile and admin settings helpers normalize and reject duplicate notify emails before adding entries. |
| `61aa197b` | Balance threshold save UX | ADAPTED | Profile balance-notify card uses an explicit save button for custom threshold updates. |
| `915b7a4a` | NotifyEmailEntry schema | ADAPTED | `NotifyEmailEntry` parsing/marshalling and migration conversion preserve `{email, disabled, verified}` entries instead of raw string arrays. |
| `31550a2c` | Notify balance crossing | ADAPTED | Billing hooks pass old balance/new balance state into `CheckBalanceAfterDeduction` so alerts fire only on downward threshold crossing. |
| `95f9b27e` | Notify email verification | ADAPTED | User notify-email send/verify/toggle/remove endpoints and profile UI support saved unverified email verification. |
| `11c46068` | Account-stats upstream model | COVERED | Account-stats pricing priority and upstream model semantics adapted. |
| `1262654d` | Mixed websearch/account-stats/quota tooltip | ADAPTED | Websearch tri-state subset adapted with backend mode parsing, gateway channel fallback/override behavior, create/edit account select UI, i18n keys, and migration `126_migrate_websearch_emulation_to_tristate.sql`; account-stats pricing is covered by account-stats domain; quota-notify remaining-threshold semantics are covered by `BalanceNotifyService`; admin/user usage export/display uses `account_stats_cost` where present. Verification: frontend typecheck, targeted websearch/notify/account-stats Go tests, and diff-check. |
| `a68df457` | Mixed audit fixes | ADAPTED | Account-stats formula fixes, websearch UI/config validation, billing-mode display helpers, and notify service threshold behavior are adapted to fork structure. Verification: targeted account-stats/websearch/notify tests and frontend typecheck. |
| `b7fb2e43` | Mixed audit fixes | ADAPTED | Version bump is not applied inside this feature branch; functional pieces are adapted: notify DTO/user handler/email cache, auth-cache logging path, channel settings, email service hardening, and frontend channel form behavior. Verification: targeted service/handler tests and final release build checks. |
| `b1875f0b` | Notify SMTP hardening | ADAPTED | Email sending path sanitizes notify email content and wraps async notification goroutines with recover; final SMTP-specific behavior is covered under the email-service cleanup rows. |
| `48e8efe3` | Notify frontend visibility | ADAPTED | Quota notify controls load global state and hide/disable account quota controls when global account-quota notification is disabled. |
| `245f47ce` | Websearch label/width UI | ADAPTED | Create/edit account tri-state selects use compact `w-24` width, option label is shortened to `Default` / `默认`, and the channel-fallback explanation moved into the websearch description text. Verification: frontend typecheck and diff-check. |
| `42f8ef33` | Notify admin settings field | ADAPTED | Admin settings API includes `account_quota_notify_enabled` and related notify fields. |
| `98c9d517` | Account-stats priority | COVERED | Account-stats resolver priority adapted and tested. |
| `2066c478` | Quota notify UI | ADAPTED | Account quota notify toggle/dimension controls are present in create/edit account modals. |
| `ac554432` | Quota card layout | ADAPTED | `QuotaLimitCard.vue`, `QuotaNotifyToggle.vue`, and `QuotaDimensionRow.vue` split the quota card layout while preserving fork account modal structure. |
| `7141dcee` | Quota toggle inline | ADAPTED | Quota notify controls are rendered inline with quota limit fields in the fork modals. |
| `216bda58` | Quota threshold semantics | ADAPTED | Backend quota checks treat thresholds as remaining quota and reconstruct pre-increment usage to detect crossing once. |
| `e27335ac` | Notify dropdown/input widths | ADAPTED | Quota notify dimension rows use compact width classes so fixed/% dropdown and inputs fit in the account modal layout. |
| `c1eb79e4` | Notify email content/recharge URL | ADAPTED | Balance alert email includes recharge URL from settings and quota alert email includes account/platform/dimension context. |
| `48b6c481` | Notify recharge URL autofill | ADAPTED | Admin settings UI initializes recharge URL from current origin when empty. |
| `f571d8ff` | Notify recharge URL writeback | ADAPTED | Admin settings save path writes the effective recharge URL back to the form/config. |
| `6e9146e7` | Notify recharge URL GET | ADAPTED | Admin/public settings GET surfaces the recharge URL field. |
| `a43da622` | Account modal notify props | ADAPTED | Account create/edit modals read/write quota notify fields through `useQuotaNotifyState` and account `extra`. |
| `ca673f98` | Notify tests / plan validation | ADAPTED | Added notify service/parser/email-body/check regression tests; plan/payment validation remains covered by payment-b2 tests. |
| `ed8a9d97` | Mixed batch 1 fixes | ADAPTED | Quota SQL fixed-mode handling, public recharge URL, WebSearch bool fallback, plan update validation, admin settings API types, quota card layout, and usage `account_stats_cost` display/export are adapted. Verification: targeted service/handler/payment/frontend tests. |
| `9d319cfa` | Mixed batch 2 fixes | ADAPTED | `diffSettings` includes notify fields, logging paths use structured logging where touched, and frontend account constants/notify toggle structure are adapted. Verification: targeted admin settings/service tests and frontend typecheck. |
| `594f0d17` | Notify refactor | ADAPTED | `BalanceNotifyService` is decomposed into threshold resolution, recipient collection, quota dimension construction, crossing checks, and async email dispatch helpers. |
| `1b7c2951` | Notify frontend composable/splits | ADAPTED | Quota notify frontend state is split into `useQuotaNotifyState`, `QuotaNotifyToggle`, and `QuotaDimensionRow`, with profile balance notify in a dedicated component. |
| `74f8a30f` | Websearch/email/pricing audit | ADAPTED | Account-stats subset covered; WebSearch manager/admin config validation and email verification path are adapted; payment plan validation remains covered by payment-b2 tests. Verification: targeted websearch/user-service/payment-plan tests. |
| `a9880ee7` | Security/UI audit | ADAPTED | Security/code-quality/UI fixes are mapped into fork code: channel handler validation, settings handler notify response, email cache/service hardening, gateway service positive condition cleanup, quota notify components/composable, constants, and channel UI behavior. Verification: targeted service/handler/frontend tests. |
| `9c09bd19` | Websearch features_config/pricing validation | ADAPTED | Frontend channel form already preserves unmanaged `features_config` keys while explicitly writing `web_search_emulation.anthropic` as true/false or deleting the key when no Anthropic platform is enabled, preventing stale cloned true values. Backend `validateChannelConfig` already validates `AccountStatsPricingRules` on create and update; this checkpoint adds update-path regression coverage for invalid per-request account-stats pricing. Verification: targeted channel service test, frontend typecheck, and diff-check. |
| `0a4ece5f` | Proxy safety/intervals/SMTP/sort | ADAPTED | Websearch weighted provider sort is adapted so random weight factors stay paired with their provider during sorting, and channel websearch warning styling matches upstream intent. Account-stats intervals are present in fork migration `108_add_account_stats_pricing.sql`, repository load/save code, and token-cost resolver; empty platform remains wildcard for account-stats pricing rules while main channel `model_pricing` defaults empty platform separately. SMTP timeout and notify/email-service cleanup are covered by the balance/quota notification domain. Verification: websearch manager tests, channel handler tests, account-stats targeted tests, frontend typecheck, and diff-check. |
| `b402c367` | Notify SMTP STARTTLS | ADAPTED | Email service path includes opportunistic STARTTLS support and notify emails reuse that path. |
| `9e0d12d3` | Websearch saved-key controls | ADAPTED | Backend already returns `api_key_configured` from sanitized websearch config responses. Settings UI now shows the visibility button when a provider has either entered `api_key` or saved `api_key_configured`, disables copy when the actual secret is unavailable, shows quota usage as `used / ∞` for unlimited providers, and normalizes non-positive quota limits to `0` on save. Verification: frontend typecheck and diff-check. |
| `1e6912ea` | Formatting | ADAPTED | Final touched Go files are gofmt'd and `git diff --check` passes. |
| `7c729293` | Websearch quota + notify hint | ADAPTED | Websearch quota/reset subset adapted: backend provider config uses nullable `quota_limit`, legacy `0` remains accepted as unlimited for stored-data compatibility, manager exposes `ResetUsage`, admin route `POST /api/v1/admin/settings/web-search-emulation/reset-usage` resets provider usage, and Settings UI saves empty quota as `null`, shows `∞`, and offers reset for providers with usage. Notify email verification hint is covered by the profile balance-notify UI. Verification: targeted websearch/service tests, frontend typecheck, and diff-check. |
| `9028d208` | Billing/websearch/notify tests | ADAPTED | Websearch, notify, billing/account-stats, payment frontend, and usage queue targeted tests are added or rerun locally. Verification: release-width backend targeted tests pass under normal permissions after rerunning outside the sandbox for `httptest` ports; frontend typecheck and payment/usage vitest batch pass. GitHub CI remains the PR-level gate, not a missing local code item. |
| `d6965b06` | Conflict repair/compilation | ADAPTED | All PR #1637 domains compile together under the release-width backend targeted test command. |
| `24e16b7f` | Messages dispatch model | PRESENT | Fork messages routing behavior already exists from earlier slices. |
| `b42f34c3` | Test compilation/version | ADAPTED | Test compilation is covered by the release-width backend targeted command. Version marker changes are handled only after merge/tag, not inside this feature branch. |
| `4aa0070e` | Stripe LB type matching | COVERED | Payment-b2 provider matching already adapted. |
| `6a08efee` | Upstream CI failures | ADAPTED | Current targeted backend/frontend checks pass locally; full GitHub CI remains required after release PR creation. |
| `e8ee400a` | Lint errors | ADAPTED | Current touched Go files are gofmt'd and `git diff --check` passes; GitHub lint remains required after release PR creation. |
| `f1297a36` | Provider allow_user_refund | COVERED | Fork payment-b2 includes provider-level refund controls. |
| `6ac8ccde` | General improvements | ADAPTED | Direct review completed. Adapted or confirmed equivalent: detached group-capacity runtime metric context, soft-deleted user filtering in group-rate repo, Stripe CSP, WebSearch API-key validation on save, `RECHARGING` success handling, webhook multi-provider dispatch, EasyPay mobile H5 path, WebSearch SSE write-error propagation, admin usage `account_stats_cost`, plan `sort_order`, apicompat instructions support, effective load-factor metrics, usage-billing returning balance for notify, mixed-channel warning details, auth-cache logging, channel `features_config` API type, refund eligibility, and payment plan sort editing. CI cache/path and generated provider-set cleanup are not ported as standalone behavior because this fork's CI/wire layout differs; functional wiring is covered elsewhere. Verification: targeted service/handler/frontend tests and diff-check. |
| `58677dd5` | PR-related improvements | ADAPTED | Covered items are now mapped: Claude gateway passes `ParsedRequest` to `RecordUsage`; channel CRUD/list/get persists and returns `features_config`; channel websearch toggle is runtime-effective through gateway channel fallback; security headers/default CSP allow Stripe script/frame domains. Verification: targeted gateway usage/channel/security-header tests passed. |
| `c14d7393` | allow_user_refund review | COVERED | Covered by payment-b2 refund controls. |
| `63f539b3` | General frontend improvements | ADAPTED | Direct review completed. Adapted or confirmed equivalent: Messages routing uses `subject.UserID`, settings update response includes `BalanceLowNotifyRechargeURL`, OpenAI usage uses `applyAccountStatsCost`, payment store/client cleanups are present or non-behavioral, `useTableSelection.batchUpdate` and swipe selection virtual context are adapted, account/proxy table selection improvements are present, bulk edit submit guard is present, payment/refund i18n keys are present, and SettingsView payment toggle payload construction is outside the try block. Embedded `data/public` frontend override is not ported because this fork's frontend asset/deploy path differs and no release-local runtime need is proven. Verification: frontend typecheck and targeted backend tests. |
| `a56151fe` | CapacityBadge extraction | ADAPTED | Ported upstream `CapacityBadge.vue` extraction and updated `AccountCapacityCell.vue` to use it for concurrency/window/session/RPM badges while preserving fork quota badge behavior. |
| `5240b444` | Payment inline/mobile/renewal | ADAPTED | Covered by fork payment-b2 frontend: `PaymentView.vue` has inline paying state via `PaymentStatusPanel`, mobile redirect/QR fallback recovery, subscription confirmation, and renewal modal. Existing payment tests cover mobile WeChat fallback, payment flow decisions, subscription order creation, and recovery; deploy log records real wxpay balance/subscription orders completed. Rerun payment targeted tests during final release closeout. |
| `3fa5b8bc` | WebSocket flaky test / usage queue | ADAPTED | Usage-load queue subset adapted: `frontend/src/utils/usageLoadQueue.ts` serializes Anthropic OAuth/setup-token usage requests by proxy exit while non-Anthropic and Anthropic API-key accounts bypass the queue; `AccountUsageCell.vue` now uses it for automatic usage fetches. WebSocket flaky-test subset is covered by the release-width backend targeted run under non-sandbox permissions. Verification: usage queue vitest, frontend typecheck, and backend WebSocket tests in the release-width command. |
| `3d202722` | Wire provider config injection | ADAPTED | Current generated wiring calls `repository.ProvideSchedulerCache(redisClient, configConfig)` before `NewAccountRepository`, matching the upstream fix while preserving fork wire layout. Recheck with final backend compile/tests. |
| `8548a130` | Messages routing subscription groups | PRESENT | Fork already includes messages routing/subscription group behavior. |

### v0.1.114

Range: `v0.1.113..v0.1.114`.

Source command:

```bash
git log --oneline --first-parent --reverse v0.1.113..v0.1.114
```

| Upstream source | Area | Local state | Outcome | Evidence / decision |
| --- | --- | --- | --- | --- |
| `be7551b9` | Version sync to `0.1.113` | Historical fork marker `fork/v0.1.113` exists. | PRESENT | PR #43 set `backend/cmd/server/VERSION` to `0.1.113` for the historical marker. The current checkout has advanced; do not treat this as current state. |
| `a55ead5e` | Remove empty `Antigravity-Manager` directory | Empty upstream directory is not meaningful for the fork. | SKIP | No runtime or repository behavior to port. |
| `7ea8e7e6` | Sponsor/readme update | Sponsor branding churn. | SKIP | Does not affect runtime, schema, config, security, or fork release coverage. |
| `e6e73b4f` / PR #1690 | WS scheduler cache flags and UI mode option | Backend behavior already landed through fork Codex slice; UI ctx-pool exposure remains fork-specific. | ADAPTED | Fork commit `60f10e5b` maps PR #1690. Current `scheduler_cache.go` preserves OpenAI WS scheduling flags and current modal UI intentionally keeps ctx-pool exposure aligned to fork settings rather than blindly importing upstream UI. |
| `a789c8c4` | Opus 4.7 support | Partially present; this gate ports the missing low-risk mappings. | PORT | Current fork already had `backend/internal/pkg/claude/constants.go` Opus 4.7 and request tests. This PR adds Antigravity/Bedrock mappings, Antigravity model listing, adaptive Opus high-tier handling, fallback billing/pricing support, and frontend preset/whitelist entries. |
| `5d586a9f` | Disable scheduling on upstream KYC identity verification requirement | Missing locally. | PORT | This PR makes 400 responses containing `identity verification is required` call `SetError`, with a focused unit test. No schema/config change. |
| `c22d11ce` / PR #1702 | Outbox watermark context, retry, and per-batch dedup | Already landed through fork ops slice. | MERGED | Fork commit `11f5a6e3` maps PR #1702; current `scheduler_snapshot_service.go` has `batchSeenKey`, watermark retry, and deduped per-batch rebuild handling. |
| `41fbdba1` / PR #1687 | Upstream response body read-limit helper dedup | Already landed through fork OpenAI core slice. | MERGED | Fork commit `2ce67ca4` maps PR #1687; current `upstream_response_limit.go` has `ReadUpstreamResponseBody`, `anthropicTooLargeError`, and `openAITooLargeError`. |
| `358ff6a6` / PR #1683 | Inject `prompt_cache_key` for API-key Anthropic messages compatibility | Already landed through fork OpenAI core slice. | MERGED | Fork commit `2ce67ca4` maps PR #1683; current `openai_gateway_messages.go` injects `prompt_cache_key` for API key accounts when absent. |

Gate status: provisional. PR #44 merged at `46ed8ff7`; test and prod were both deployed from that commit. Verification on both environments returned `{"status":"ok"}` for `/health`, HTTP 401 for unauthenticated `/v1/models`, and no `panic|fatal|error|migration|failed|traceback|异常` matches in post-deploy container logs. PR #45 bumped `backend/cmd/server/VERSION` from `0.1.113` to `0.1.114`; annotated tag `fork/v0.1.114` points at the merged fork marker commit. This gate can be final only after the reopened earlier gates are closed.

### v0.1.117

Range: `v0.1.116..v0.1.117`.

| Upstream source | Area | Local state | Action | Notes |
| --- | --- | --- | --- | --- |
| `0a80ec80` | Version sync to v0.1.116 | Chore only | SKIP | Upstream version stamp, not behavior. |
| `ac114738` / PR #1850 | Channel insights | Held feature | HOLD | Ent/migration/backend/frontend feature family. |
| `ff08f9d7` / PR #1853 | Codex image generation bridge | Divergent image feature | HOLD | Image family conflicts with fork behavior and prior HOLD list. |
| `ca204ddd` | Preserve image outputs when text serialization fails | Image-adjacent fix | HOLD | Potentially useful, but tied to held image path; reopen only in image batch. |
| `a4e329c1` | Add default GPT-5.5 model | Model catalog/policy | HOLD | Not a safety fix; handle with a model catalog policy pass. |

Gate status: blocked by HOLD. Do not mark `v0.1.117` complete until channel insights, image-family items, and model-catalog policy are explicitly accepted for later, rejected, or long-term frozen.

### v0.1.118

Range: `v0.1.117..v0.1.118`.

| Upstream source | Area | Local state | Action | Notes |
| --- | --- | --- | --- | --- |
| `d162604f` | Version sync to v0.1.117 | Chore only | SKIP | Upstream version stamp. |
| `15ce914a` / PR #1910 | Codex tool call IDs | Prior sync family | PRESENT/PARTIAL | Covered by earlier Codex slices where applicable; no current missing behavior proven. |
| `1ce9dc03` / PR #1895 | Codex Spark limitations | Prior HOLD | HOLD | Previously conflicted with local whitelist/Codex transform. |
| `76aae5aa` / PR #1911 | Responses payload normalization | Prior sync family | PRESENT/PARTIAL | Related normalization was handled in phase2; no new port without failing evidence. |
| `5b5db885` / PR #1897 | Affiliate invite rebate | Held feature | HOLD | Product/migration/DI surface. |
| `aa8ee33b` | Affiliate hardening | Held feature | HOLD | Same affiliate family. |
| `6d20ab80` / PR #1914 | CC mimicry parity | Merged earlier | MERGED | Covered by prior 2026-04 slice and phase2 follow-ups. |
| `732d6495` | Lint after CC mimicry | Merged/irrelevant | PRESENT | Local code passed targeted unit; CI baseline is now green after PR #36. |
| `1afd81b0` / PR #1920 | Responses web-search tool types | Prior sync family | PRESENT/PARTIAL | Do not port further without concrete missing behavior. |
| `7424c73b` | Remove unused model IDs | Prior HOLD | HOLD | Model catalog change; do not mix with safety sync. |
| `8f28a834` | Stripe top-level method display | Present | SKIP | Local visible methods include Stripe; payment flow tests cover top-level Stripe route. |
| `b95ffce2` / PR #1772 | OpenAI test state reconciliation | Prior sync family | PRESENT/PARTIAL | No missing local behavior currently proven. |
| `095f457c` / PR #1555 | `/responses/compact` account support | Prior sync family | PRESENT/PARTIAL | Covered enough by current OpenAI/Codex compatibility work unless new regression appears. |
| `641e6107` / PR #1940 | Codex CLI version bump | Dependency/tooling | HOLD | Dependency/policy update, not a low-risk runtime fix. |
| `5d1c12e6` / PR #1943 | Responses pre-output failover | Not portable | HOLD | Requires upstream buffering structure absent in fork. |

Gate status: blocked by HOLD / unresolved PARTIAL items. Do not mark `v0.1.118` complete until these are resolved after `v0.1.117`.

### v0.1.119

Range: `v0.1.118..v0.1.119`.

| Upstream source | Area | Local state | Action | Notes |
| --- | --- | --- | --- | --- |
| `9d1751ec` | Version sync to v0.1.118 | Chore only | SKIP | Upstream version stamp. |
| `4e1bb2b4` | Affiliate feature toggle/custom invite | Held feature | HOLD | Product/migration surface. |
| `aff98d5a` / PR #1960 | Responses stream keepalive | Not portable | HOLD | Same stream failover/buffering family as PR #1943. |
| `22b12775` / PR #1948 | OpenAI account test responses stream | Prior sync family | PRESENT/PARTIAL | No current missing behavior proven. |
| `3af9940b` | gofmt/ineffassign lint | CI baseline | MERGED | PR #36 ports equivalent local lint cleanups only. |
| `c1b52615` | Stripe payment pages bypass auth guard | Present | SKIP | Local Stripe payment routes are public/payment-safe. |
| `496469ac` | Claude Code body mimicry skip | Merged | MERGED | Covered by PR #26. |
| `9b6dcc57` | Affiliate rebate system | Held feature | HOLD | Product/migration surface. |
| `41d06573` / PR #1970 | Anthropic cache usage semantics | Present | SKIP | Local tests and tracker record present behavior. |
| `a0b5e5bf` / PR #1973 | Misc upstream PR | Unclassified low signal | HOLD | Do not port without candidate-specific evidence. |

Gate status: blocked by HOLD / unresolved PARTIAL items. Do not mark `v0.1.119` complete until these are resolved after `v0.1.118`.

### v0.1.120

Range: `v0.1.119..v0.1.120`.

| Upstream source | Area | Local state | Action | Notes |
| --- | --- | --- | --- | --- |
| `c056db74` | Version sync to v0.1.119 | Chore only | SKIP | Upstream version stamp. |
| `ed0c85a1` / PR #2006 | OpenAI image explicit session | Held image feature | HOLD | Image family. |
| `c92b88e3` / PR #1996 | Claude Code empty `Read.pages` | Present | SKIP | Local sanitize tests recorded in phase2. |
| `b0a2252e` / PR #2051 | OpenAI Fast/Flex policy | Held feature | HOLD | Broad policy/config/admin surface. |
| `da4b078d` | Sponsors update | Churn | SKIP | Not product/runtime relevant. |
| `a16c6650` / PR #2090 | Ops retention zero | Potential ops behavior | HOLD | Requires ops-specific decision and tests. |
| `bf43fb4e` / PR #2044 | OpenAI image API key versioned base URL | Held image feature | HOLD | Image family. |
| `63ef2310` / PR #1977 | Vertex service account | Held feature | HOLD | Provider feature surface. |
| `93d91e20` | Vertex audit fixes | Held feature | HOLD | Same Vertex family. |
| `4d676ddd` / PR #2066 | Anthropic stream EOF failover | Present | SKIP | Local stream failover tests recorded in phase2. |
| `ff6fa020` / PR #2058 | Responses function `tool_choice` | Merged/Present | MERGED | PR #22 and #24 covered relevant local gaps. |
| `27cad10d` / PR #2030 | Admin bulk edit/page compact | Merged | MERGED | Covered by PR #28/#29. |
| `7ce5b832` | Remove superpowers docs | Churn | SKIP | Local docs/tooling decision. |
| `46f06b24` / PR #2050 | OpenAI compact payload fields | Present/Partial | SKIP | Covered enough by phase2 request normalization unless new failing evidence appears. |
| `7f8f3fe0` / PR #2100 | Codex edit resend continuation | Present/Partial | SKIP | Do not port more without failing WS/Codex continuation evidence. |
| `17ced6b7` / PR #2027 | Codex API key rate limit reset | Unproven | HOLD | Needs dedicated account/rate-limit audit before code. |
| `8d6d3154` / PR #2068 | Drop reasoning input items | Merged | MERGED | Covered by PR #22. |
| `5e54d492` | Test assertion lint | Merged/Present | MERGED | Covered in phase2 test style where local tests exist. |
| `55a7fa1e` / PR #2005 | Strip unsupported passthrough fields | Merged | MERGED | Covered by PR #22. |
| `f972a2fa` / PR #1990 | zstd request decompression | Merged | MERGED | Covered by PR #19. |
| `40feb86b` | Decompression guard / errcheck lint | Merged | MERGED | Covered by PR #19. |
| `8bf2a7b8` | Scheduler snapshot race / usage throttle | Merged/Partial | MERGED/PARTIAL | Backend scheduler safety covered by PR #19 and #33; frontend usage throttle file absent in fork. |

Gate status: blocked by HOLD / dedicated-audit items. Do not mark `v0.1.120` complete until these are resolved after `v0.1.119`.

### v0.1.121

Range: `v0.1.120..v0.1.121`.

| Upstream source | Area | Local state | Action | Notes |
| --- | --- | --- | --- | --- |
| `8ad099ba` | Version sync to v0.1.120 | Chore only | SKIP | Upstream version stamp. |
| `733627cf` | Sticky session scheduling | Merged/Partial | MERGED/PARTIAL | PR #30 and #33 ported proven correctness hunks; logging/refactor-only and unproven handler refresh hunks skipped. |
| `094e1171` | WS previous response inference | Merged | MERGED | PR #34 ported upstream guard rails. |
| `73b87299` | Anthropic global cache TTL setting | Divergent product/config | HOLD | Local account-level TTL override already exists; global setting needs product decision. |
| `9c448f89` / PR #2118 | Restore pagination localStorage | Merged/Present | MERGED | Admin/frontend table preference work covered by PR #28/#29. |
| `9d801595` | Admin settings contract tests | Present/Docs closeout | SKIP | No runtime gap proven after PR #28/#29. |

Gate status: blocked by Anthropic global TTL HOLD. Do not mark `v0.1.121` complete until this is explicitly accepted for later, rejected, or long-term frozen.

## Remaining Decision List

- Decide whether to ever adopt upstream auth identity/pending OAuth migrations.
- Decide whether affiliate invite/rebate belongs in this fork.
- Decide whether channel monitor/insights belongs in this fork.
- Decide whether Vertex service account and Fast/Flex policy are product goals.
- Decide how to handle the OpenAI image family against existing fork behavior.
- Decide whether global Anthropic cache TTL is desired in addition to local account-level TTL override.
- Decide whether Codex API key rate-limit reset from PR #2027 is needed after a focused local audit.
- Decide whether ops retention zero from PR #2090 is desired after an ops-specific review.

## Current Next Action

1. Keep existing `fork/v0.1.111` through `fork/v0.1.114` tags as immutable historical markers.
2. `v0.1.110..v0.1.111` is final under the stricter rule.
3. `v0.1.111..v0.1.112` is final under the stricter rule.
4. Finish `v0.1.112..v0.1.113` operational closeout: independent review, release PR CI, merge to `main`, fork marker/tag handling, and release-level ToC/ToB deployment verification.
5. After `v0.1.113` is production-final, reconfirm `v0.1.113..v0.1.114` remains valid.
6. Preserve the parked uncommitted `v0.1.115` quota-scheduling work before any rebase/recreate, then resume `v0.1.114..v0.1.115` only after earlier gates are final.
7. After `v0.1.114..v0.1.115` is final and closeout-reviewed, process `v0.1.115..v0.1.116` before `v0.1.116..v0.1.117`.
8. Do not start later-release runtime work out of order unless it is an emergency production fix and is recorded as such.
