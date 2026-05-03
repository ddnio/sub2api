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

- Local base: `origin/main` at `fbaa1fdd sync(upstream): close v0.1.112 release gate (#40)`.
- Upstream published-tag scope: latest local upstream tag `v0.1.121` at `9d801595 test: 更新管理员设置契约字段`.
- Upstream main observed locally: `b2bdba78 stabilize image request handling`; this is not in the published tag scope yet.
- Latest upstream tag in scope: `v0.1.121`.
- Work branch: `docs/upstream-sync-final-smoke`.
- Worktree: `.claude/worktrees/upstream-sync-final-smoke`.
- Plan: `docs/plans/2026-05-03-upstream-release-coverage.md`.

If `upstream/main` advances, do not expand this ledger automatically. Only expand after a new upstream tag is fetched and the release interval is reviewed.

Fetch note: 2026-05-03 attempts to refresh `upstream` and `origin` failed with GitHub transport errors (`HTTP2 framing layer` / `Operation timed out`). Current decisions use the already present local refs above and the local upstream tags through `v0.1.121`. Push/fetch fallback order for this environment is: retry with `git -c http.version=HTTP/1.1`; if HTTPS still fails or hangs, use the explicit SSH URL (`git@github.com:ddnio/sub2api.git` for fork operations) or the GitHub Git Data API branch-creation fallback already documented in `docs/engineering/upstream-sync-2026-05-phase2.md`.

## Repeated-Issue Log

- GitHub HTTPS transport is unreliable in this environment. Do not spend multiple cycles retrying the same HTTPS fetch/push after `HTTP2 framing layer`, `Empty reply from server`, or timeout errors; switch to HTTP/1.1 once, then SSH or API fallback.
- Fork release marker tags use `fork/vX.Y.Z`, not upstream tag names. Upstream `vX.Y.Z` tags already exist and point at upstream commits.
- When inspecting upstream commits, copy full hashes from `git log --oneline --first-parent`; an abbreviated typo such as dropping one hex digit makes `git show` fail and wastes a cycle.
- In shell commands, avoid placeholder strings with angle brackets such as `<NULL>` unless they are safely quoted for the local shell; zsh can interpret them as redirection before SSH runs.
- Adding a lower-numbered migration after later migrations already exist is acceptable only when the runner sorts by filename and skips already-applied files by filename. In this repo, `applyMigrationsFS` sorts embedded `*.sql` names lexicographically and records each filename in `schema_migrations`, so a missing historical migration such as `097_*` can still be safely added before `098_*`.
- PR #40 CI note: pull-request integration run `25276045188` failed once in `TestOpsSystemLogSink_StartStopAndFlushSuccess` because the test signaled `done` before the sink goroutine incremented `writtenCount`; the same commit's push CI `25276036809` passed. Fix test races by waiting on the observed health condition, not by blindly rerunning.

## Rules

- Do not merge or cherry-pick an entire release.
- Code merge unit is upstream PR/merge commit or a smaller manually ported hunk.
- Process releases in tag order. A later release must not start until the previous release gate is closed.
- `HOLD` is not a completed state by itself. A release with unresolved `HOLD` items is not complete. A `HOLD` item closes only when it is explicitly accepted for a later dedicated project, rejected for this fork, or long-term frozen with a documented owner/reason.
- Already merged fork PRs #19-#35 are treated as current implementation baseline.
- Product, schema/migration/Ent, auth identity, pending OAuth, affiliate, channel monitor/insights, Vertex, Fast/Flex, OpenAI image refactors, payment semantics, license/CLA, and sponsor/readme churn require explicit release-gate decisions.
- Docs-only PRs require `git diff --check`, self-review, Kimi review, and PR-level review; they do not deploy.
- Runtime PRs require targeted tests, self-review, Kimi pre-commit review, PR-level Kimi review, GitHub CI, and a recorded deployment gate if deployed.

## Ledger Semantics

This ledger PR is documentation-only in the fork. The upstream commits listed below are mostly code changes. Listing a code commit in this ledger does not authorize merging it; the `Action` column is the decision boundary.

- `MERGED`: the relevant behavior has already landed in this fork through a reviewed fork PR.
- `ADAPTED`: the upstream behavior or feature family was intentionally implemented through a fork-specific architecture instead of preserving the upstream commit ancestry.
- `PRESENT`: current fork code already provides the behavior; no new PR is planned.
- `PARTIAL`: some behavior exists or was ported, but more upstream code is not automatically accepted.
- `PORT`: a future small PR may port a proven low-risk missing behavior.
- `HOLD`: not complete; must be converted to accepted-later, rejected, or long-term-frozen before the release gate closes.
- `REJECTED`: explicitly not adopted for this fork/release, with reason.
- `FROZEN`: accepted only as a later dedicated project, with owner/reason; it is closed for the current release gate.
- `SKIP`: intentionally ignore for this fork cycle, usually chore/version/churn or no proven local gap.

## Sequential Gate Status

The previous 2026-04 sync closed a slice-based Phase 1, not a release-gate sequence. Do not process the historical HOLD set as a separate global queue. Instead, start from the earliest release interval that was not closed under the release-gate rule, then decide each item inside that release before advancing.

Current gate:

| Gate | Status | Required next action |
| --- | --- | --- |
| `v0.1.110..v0.1.111` | Closed | Decision matrix completed, fork marker bumped to `0.1.111`, and sync tag `fork/v0.1.111` pushed. |
| `v0.1.111..v0.1.112` | Marker in progress | Decision matrix completed and PR #40 merged at `fbaa1fdd`; bump the fork marker to `0.1.112`, then push sync tag `fork/v0.1.112`. |
| `v0.1.112..v0.1.113` | Blocked | Start only after the `0.1.112` marker PR lands and sync tag `fork/v0.1.112` is pushed. |
| `v0.1.117` and later | Blocked | The earlier ledger that started at `v0.1.117` is not the active start point anymore. Do not advance here until the earlier gates are closed in order. |

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
| `0d69c0cd` | Version sync to `0.1.110` | Commit is ancestor of local `HEAD`; current `backend/cmd/server/VERSION` is `0.1.110`. | PRESENT | Current fork release marker is still `0.1.110`; bumping to `0.1.111` is the marker step after this gate closes. |
| `155d3474` | Sponsors churn | Commit is ancestor of local `HEAD`. | SKIP | Sponsor/readme churn does not affect fork runtime or product behavior. |
| `1b79f6a7` / PR #1522 | Redis scheduler snapshot metadata and large MGET chunking | Commit is ancestor of local `HEAD`. | MERGED | Local `scheduler_cache.go` contains chunked `MGet` and preserves metadata fields such as `LoadFactor`. |
| `74302f60` / PR #1010 | OIDC login | Commit is ancestor of local `HEAD`. | MERGED | Current fork already includes OIDC config/public settings plumbing from this upstream family. |
| `9a72025a` / PR #1523 | Include `home_content` URL in CSP `frame-src` | Commit is ancestor of local `HEAD`. | MERGED | `SettingService.GetFrameSrcOrigins` adds `settings.HomeContent` before purchase/custom-menu origins. |
| `760cc7d6` / PR #1481 | Increase stored error-log body limit | Commit is ancestor of local `HEAD`. | MERGED | Local ops service has upstream-equivalent error-body/request-body sanitization; no further port required in this release. |
| `bbc79796` / PR #1529 | Group `/v1/messages` dispatch redo | Commit is ancestor of local `HEAD`. | MERGED | Local code has `OpenAIMessagesDispatchModelConfig`, group UI controls, migration `091_add_group_messages_dispatch_model_config.sql`, and dispatch resolution tests. |
| `00c08c57` / PR #1539 | Sync `load_factor` into scheduler cache | Commit is ancestor of local `HEAD`. | MERGED | `buildSchedulerMetadataAccount` copies `LoadFactor` into scheduler metadata snapshots. |
| `1ef3782d` / PR #1538 | Broad admin/repository/frontend bug-cleanup batch | Upstream merge commit is not an ancestor; selected behavior has landed through later fork/admin slices. | FROZEN | Owner: upstream-sync maintainer. Reason: 117-file mixed admin/repository/frontend cleanup overlaps later fork work and is not safe as a whole-release import. Keep as a later item-by-item audit source; no current missing behavior is proven. |
| `97f14b7a` / PR #1572 | Payment system v2 | Upstream merge commit is not an ancestor; fork intentionally replaced/adapted it through `623dda62` and the payment-b2 sequence through production hotfix `6518510b`. | ADAPTED | Payment-b2 audit and deploy logs show fork-specific migrations, provider instances, checkout/result flows, Stripe/Alipay/Wxpay providers, webhook/refund/resume tests, and test/prod deployment. Do not cherry-pick upstream payment v2 over the fork adaptation. |
| `54490cf6` / PR #1576 | Payment docs | Upstream merge commit is not an ancestor; upstream docs are superseded by fork payment-b2 operational docs. | ADAPTED | Current docs include `payment-b2-upstream-audit.md`, `payment-b2-deploy.md`, and `payment-b2-deploy-log.md`, which document the fork-specific payment architecture and deployment evidence. |
| `9b7b3755` / PR #1543 | Messages-dispatch i18n | Upstream merge commit is not an ancestor, but fork PR #9 imported the relevant i18n slice in `d80a3827`. | MERGED | `git log --all --grep 1543` maps PR #1543 to fork slice #9; local i18n keys for messages dispatch are present. |
| `16126a2c` / PR #1545 | Smooth sidebar collapse | Upstream merge commit is not an ancestor and conflicts with fork sidebar. | REJECTED | Owner: upstream-sync maintainer. Reason: fork `AppSidebar.vue` has diverged through local navigation/payment/contact changes; previous slice-7 explicitly skipped #1545 because the upstream sidebar rewrite is not safely applicable. Reopen only if a frontend sidebar redesign is requested. |
| `82b840c1` / PR #1587 | Anthropic 400 credit-balance handling | Upstream merge commit is not an ancestor, but fork PR #10 imported equivalent Anthropic handling in `a53527fa`. | MERGED | `ratelimit_service.go` disables Anthropic accounts on 400 bodies containing `credit balance`; fork slice #10 covered this family. |
| `a1a28368` | Sponsors churn | Not an ancestor after fork slices. | SKIP | Sponsor/readme churn; no fork behavior. |
| `9648c432` | Frontend TS2352 cast fix in API client | Upstream merge commit is not an ancestor, but equivalent code is present. | PRESENT | `frontend/src/api/client.ts` uses `apiResponse as unknown as Record<string, unknown>` and preserves `reason`/`metadata` for payment errors. |

Gate status: decision-complete, pending release-marker update. Next code step is to bump the fork release marker from `0.1.110` to `0.1.111` in a small marker PR if self-review and Kimi review confirm no unresolved item remains. After that marker PR lands, create the fork sync tag `fork/v0.1.111` on the merged fork commit before starting `v0.1.111..v0.1.112`.

Tag namespace note: do not create a fork tag named exactly `v0.1.111`. That tag name already exists for the upstream release and points at upstream commit `9648c432`; using the same tag name for a different fork commit would create a tag collision across remotes. Fork coverage tags use the `fork/vX.Y.Z` namespace.

Marker closeout: PR #39 bumped `backend/cmd/server/VERSION` to `0.1.111` and merged at `d2a3e5a9`; annotated tag `fork/v0.1.111` points at that merged fork commit.

### v0.1.112

Range: `v0.1.111..v0.1.112`.

Source command:

```bash
git log --oneline --first-parent --reverse v0.1.111..v0.1.112
```

| Upstream source | Area | Local state | Outcome | Evidence / decision |
| --- | --- | --- | --- | --- |
| `ad64190b` | Version sync to `0.1.111` | Current fork marker is already `0.1.111`. | PRESENT | PR #39 set `backend/cmd/server/VERSION` to `0.1.111` before this gate started. |
| `e70812f0` / PR #1623 | Anthropic buffered empty terminal output | Equivalent behavior is present through earlier fork Anthropic slice. | MERGED | `openai_gateway_messages.go` uses `apicompat.NewBufferedResponseAccumulator`, handles `response.done`, and calls `SupplementResponseOutput`; fork slice #10 includes `9d40fcaa` / PR #1623 mapping. |
| `7d80b5ad` / PR #1610 | Alipay/Wxpay base payment type mapping | Fork payment-b2 implements the mapping through provider instances and visible-method source selection. | ADAPTED | `DefaultLoadBalancer` can select across providers for base `alipay`/`wxpay`, `InstanceSelection` carries `ProviderKey`, and payment resume/source tests cover official/easypay routing. Do not cherry-pick upstream payment service code over fork payment-b2. |
| `75908800` / PR #1612 | QR code density | Equivalent frontend behavior is already present. | PRESENT | `PaymentQRDialog.vue` and `PaymentQRCodeView.vue` use `M` error correction with logos and `L` without logos; `PaymentStatusPanel.vue` uses `M`. |
| `d949acb1` / PR #1603 | DataTable mobile double render | Already landed through fork frontend slice. | MERGED | Fork commit `a845041a` maps PR #1603 and touched `DataTable.vue` plus `AccountUsageCell.vue`. |
| `ad6c3281` / PR #1575 | Cursor responses body compatibility | Already landed through fork Codex/Cursor slice. | MERGED | Fork commit `60f10e5b` includes `openai_codex_transform.go`, `openai_gateway_chat_completions.go`, and Cursor warmup tests for this family; `git log --all --grep 1575` also maps upstream PR #1575. |
| `66bea2b5` / PR #1624 | Version dropdown clipping | Fork applied a minimal sidebar-compatible fix instead of upstream sidebar churn. | ADAPTED | Fork commit `58c0f576` updates `AppSidebar.vue` and its spec for the expanded brand/version dropdown. This keeps the fork sidebar structure intact. |
| `92f4a6bb` | README/partner logo churn | Not product/runtime relevant for this fork gate. | SKIP | Documentation/logo sponsor churn; no local behavior. |
| `f9f57e95` | Restore `settings.updated_at` SQL default | Missing locally; this PR ports the migration. | PORT | Added upstream `backend/migrations/097_fix_settings_updated_at_default.sql` and an integration assertion that final schema keeps `settings.updated_at DEFAULT now()`. Test/prod read-only checks showed both current databases already have `DEFAULT now()`, `is_nullable=NO`, `updated_at NULL count=0`, and already applied `098`/`111`; `097` is still absent there, so this is a compatibility/backfill marker for code and older instances rather than a current prod rescue. |

Gate status: decision-complete, pending release-marker update. PR #40 merged the migration gate at `fbaa1fdd` after CI passed. This marker PR bumps `backend/cmd/server/VERSION` from `0.1.111` to `0.1.112`; after it lands, create annotated fork sync tag `fork/v0.1.112`.

Runtime/deploy note for `097`: this release gate contains a database migration file. Before any deployment of the merged PR, take the normal database backup. Current test/prod evidence indicates the migration should no-op on the live databases because the target default already exists, but it will still be recorded in `schema_migrations` on startup.

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

1. Merge the small release-marker PR that bumps `backend/cmd/server/VERSION` from `0.1.111` to `0.1.112`.
2. After the marker PR lands, create and push fork sync tag `fork/v0.1.112` on the merged fork commit.
3. Only after the tag exists in `ddnio/sub2api`, start the next gate: `v0.1.112..v0.1.113`.
4. Do not start later-release runtime `PORT` work out of order unless it is an emergency production fix and is recorded as such.
