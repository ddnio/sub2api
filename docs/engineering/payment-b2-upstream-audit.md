# Payment B-2 Upstream Backport Audit

Date: 2026-05-01
Branch: `worktree-payment-b2`
Local HEAD: `8fdb3e8a`
Upstream baseline: `upstream/main` at `48912014`

## Goal And Scope

This audit is the gate for the remaining Payment B-2 work. The goal is to finish payment-related upstream sync without importing unrelated upstream modules.

In scope:

- payment provider selection and visible payment method routing
- payment order creation, fulfillment, refund, webhook, resume/result flows
- payment provider config validation and redaction
- user payment pages, Stripe/Alipay/Wxpay launch flows, public result pages
- payment migrations already required by this branch
- deployment and rollback notes for payment only

Out of scope unless explicitly separated into a payment-only patch:

- auth identity foundation and pending OAuth redesign
- profile binding pages
- channel monitor / available channels
- affiliate rebate system
- gateway/OpenAI/Codex model changes
- table defaults, notification, pricing, RPM, ops changes

## Baseline Verification

Baseline commands were run before opening this audit:

```text
backend: GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment
result: ok github.com/Wei-Shaw/sub2api/internal/payment

backend: GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/service -run 'Test.*Payment|Test.*Wechat|Test.*WeChat|Test.*Provider|Test.*Order|Test.*Refund|Test.*Fulfillment|Test.*Config'
result: ok github.com/Wei-Shaw/sub2api/internal/service

frontend: pnpm exec vitest run src/__tests__/buttonClasses.spec.ts src/components/payment/__tests__/paymentFlow.spec.ts src/views/user/__tests__/PaymentView.spec.ts
result: 27 tests passed

frontend: pnpm exec vue-tsc --noEmit
result: passed
```

These commands must be rerun after any code backport.

## Local Payment Fixes To Preserve

The following fork-local Payment B-2 fixes are intentional and must not be lost during upstream comparison:

- `623dda62`: squashed payment v2 architecture replacement from upstream, adapted to this fork's schema and routes.
- migration safety fixes: `b5455f76`, `6306e5e7`, `09320621`, `bc3a9312`, `40f07418`.
- admin payment routes/config completion: `395db392`, `4a8b7659`.
- provider admin UI and upstream field alignment: `a143b696`, `a57da228`, `6ee63fc1`.
- checkout limit loading and min/max validation: `4883feaf`, `90f1a8e1`.
- legacy order and plan recovery: `e7f58237`, `043415a0`.
- subscription plan group guard: `2f38cd34`, `b8aeb06a`.
- upstream page/limit alignment plus review fixes: `a15de9b9`, `e66560b1`, `68cf4d16`, `9b5911c2`.
- payment action reachability and inline errors: `af41e931`, `8b6f9d0e`.
- restored global payment button classes and guard test: `2cafd049`.
- test deployment record: `052cb44d`.

Code evidence currently present:

- Stripe pages are public: `/payment/result`, `/payment/stripe`, `/payment/stripe-popup` use `requiresAuth: false` in `frontend/src/router/index.ts`.
- payment button CSS is guarded by `frontend/src/__tests__/buttonClasses.spec.ts`.
- public result and resume APIs exist: `/payment/public/orders/verify` and `/payment/public/orders/resolve`.
- unknown-order webhook returns provider success through `service.ErrOrderNotFound`.
- provider snapshots exist on orders through `provider_snapshot`, `provider_key`, and tests in `payment_order_provider_snapshot_test.go`.
- provider secrets are masked on admin list through `decryptAndMaskConfig`.
- config read path accepts plaintext JSON with legacy ciphertext fallback. Initial audit found the write path was still AES-only; this was corrected after Kimi review.

## Migration Delta

`git diff --name-status HEAD..upstream/main -- backend/internal/migrations backend/ent/schema backend/ent/migrate` currently reports mostly non-payment schema drift:

```text
A backend/ent/migrate/auth_identity_fk_ondelete_test.go
M backend/ent/migrate/schema.go
A backend/ent/schema/auth_identity.go
A backend/ent/schema/auth_identity_channel.go
A backend/ent/schema/auth_identity_schema_test.go
A backend/ent/schema/channel_monitor.go
A backend/ent/schema/channel_monitor_daily_rollup.go
A backend/ent/schema/channel_monitor_history.go
A backend/ent/schema/channel_monitor_request_template.go
M backend/ent/schema/group.go
A backend/ent/schema/identity_adoption_decision.go
D backend/ent/schema/payment_plan.go
A backend/ent/schema/pending_auth_session.go
M backend/ent/schema/user.go
D backend/ent/schema/user_referral.go
```

`git diff --name-status HEAD..upstream/main -- backend/migrations` shows upstream has additional auth identity, monitor, RPM, affiliate, notification, pricing and gateway migrations. Payment-adjacent differences are:

```text
A upstream 113_normalize_legacy_wechat_provider_key.sql
A upstream 118_wechat_dual_mode_and_auth_source_defaults.sql
A upstream auth_identity_payment_migrations_regression_test.go
D local 120b_backfill_subscription_plans_from_payment_plans.sql
D local 121_validate_payment_orders_out_trade_no_index.sql
D local 122_backfill_empty_payment_order_out_trade_no.sql
D local 123_disable_invalid_subscription_plans.sql
D local 124_disable_resold_invalid_subscription_plans.sql
```

Decision:

- Do not add upstream auth identity migrations in this pass.
- Do not delete local payment repair migrations; they are protecting test/prod upgrade paths observed during Payment B-2.
- Upstream `113_normalize_legacy_wechat_provider_key.sql` remains blocked because this branch already removed the previous auth-module version after deployment experience (`ac575113` / `db98c21d`).
- Any future migration change requires a DB backup, dry-run on test, and explicit rollback note before deployment.

## Commit Classification Matrix

Legend:

- `MUST`: correctness/security fix needed and safe to apply.
- `SHOULD`: user/ops quality fix, safe if isolated.
- `PARTIAL`: behavior already present locally; verify with tests.
- `SKIP`: already covered or superseded locally.
- `BLOCKED`: requires non-payment upstream infrastructure.
- `HOLD`: changes product/business semantics; do not alter without explicit decision.

| Upstream commit | Classification | Reason / evidence | Action |
| --- | --- | --- | --- |
| `63d1860d` complete payment system | `PARTIAL` | Base architecture was squashed into `623dda62` with fork-specific migrations/routes. | Preserve local version; no cherry-pick. |
| `e1547d78`, `e3a000e0`, `75155903`, `a020fc52` early payment fixes | `PARTIAL` | Current branch has v2 files, H5/mobile fields, Stripe pages and expiry handling. | Covered by local tests. |
| `f498eb8f`, `4aa0070e` provider type / Stripe matching | `PARTIAL` | `load_balancer.go` special-cases `TypeStripe` and canonical visible method routing. | Verify load balancer tests. |
| `5bae3b05` audit fixes for providers | `PARTIAL` | Current code has provider mismatch guard, masked admin config, visible method and webhook updates. | Covered by service tests; no broad cherry-pick. |
| `3c884f8e`, `56e4a9a9` audit tests/constants | `SHOULD` | Some coverage exists locally, but exact missing tests may still be useful. | Inspect for payment-only tests after matrix review. |
| `c738cfec` critical payment fixes | `PARTIAL` | Current code checks provider amount validity, provider mismatch, route hardening and result behavior. | Verify targeted service tests. |
| `75e1b40f` unknown-order webhook 2xx | `PARTIAL` | `PaymentWebhookHandler` handles `errors.Is(err, service.ErrOrderNotFound)` and writes success. | Covered by existing tests. |
| `d5dac84e` ErrOrderNotFound contract tests | `PARTIAL` | `payment_fulfillment_order_not_found_test.go` and webhook handler tests exist. | Keep tests. |
| `c1b52615` Stripe pages bypass auth guard | `PARTIAL` | Router marks `/payment/stripe` and `/payment/stripe-popup` public. | Covered by router tests if available. |
| `fd0c9a13`, `61a008f7` plaintext config with legacy fallback | `MUST/APPLIED` | Initial audit found local write path still required AES. Applied payment-only fix: new provider configs are plaintext JSON; both service and load balancer keep legacy AES fallback. | Verify config tests and admin masking. |
| `c3cb0280` Alipay redirect/H5/popup sizing | `PARTIAL` | Alipay provider uses WAP/PagePay/Precreate; frontend has popup/redirect launch handling. | Verify paymentFlow + Alipay provider tests. |
| `235f7108` redact provider secrets | `PARTIAL` | Admin list masks sensitive provider config fields. | Verify provider config tests. |
| `79192cf6` wxpay structured config errors | `PARTIAL` | Save path validates enabled provider by constructor; wxpay tests exist. | Verify provider tests. |
| `60a4b931` balance multiplier/refund separation | `HOLD/PRESENT` | Business semantics already present locally: credited balance uses multiplier, `pay_amount` is separate. | Do not change semantics now; test current behavior. |
| `98140f6c`, `e761d38f`, `d149dbc9`, `3053c56c`, `342dbd2e` recharge fee-rate line | `HOLD/PRESENT` | Current checkout exposes fee rate and order stores `fee_rate`; user-facing amount logic was already debugged. | Preserve current product behavior; no new change unless user asks. |
| `b51bc7ee` return URL payload | `PARTIAL` | Frontend sends `return_url`; backend validates internal payment result URL. | Verify resume service tests. |
| `40d4e167` payment i18n/error normalization | `PARTIAL` | Local i18n and inline error display exist; exact upstream strings may differ. | Keep local if tests pass. |
| `9bebf1c1` result by resume token | `PARTIAL` | `/payment/public/orders/resolve` and `payment_resume_lookup.go` exist. | Covered by resume tests. |
| `7ef7fd19`, `55e8dd55`, `29caf851`, `dd314c41` WeChat resume/result flow | `PARTIAL` | WeChat payment callback, resume parser, signed resume token, and tests exist. | Verify handler/service/frontend tests. |
| `f83fd59d`, `65d3bd72`, `c297d011`, `a27a7add` payment UX/result consistency | `PARTIAL` | `paymentUx.ts`, `PaymentStatusPanel`, pending result state and inline errors are present. | Verify frontend tests. |
| `16be82b9`, `9d5e9bbc`, `b22d00e5`, `8f28a834` visible payment method source | `PARTIAL` | Current code normalizes visible methods and supports configured source provider. | Verify methods when EasyPay + Stripe are both enabled. |
| `07f23aaa` wxpay config contract/H5 scene | `PARTIAL` | Wxpay config and providerConfig tests exist; code includes JSAPI app ID filtering. | Verify wxpay/providerConfig tests. |
| `561405ab`, `c0b24aef`, `35aeeaa6`, `119f784d`, `0934f737` provider snapshots | `PARTIAL` | Local migrations `112` and `117`, order snapshot build/read paths, metadata validation tests exist. | Preserve; do not rework migrations. |
| `7c7924e9`, `e3f69e02`, `bdcd3d87`, `b3098221`, `6f00efa3`, `64e401e2`, `267844eb` webhook/refund provider resolution | `PARTIAL` | `payment_webhook_provider.go`, alias support, provider snapshot fallback and refund tests exist. | Run backend payment/service tests. |
| `9742796e` retire public verify and backfill trade no | `PARTIAL/FORK` | Upstream retired active public verify; local deliberately keeps a passive public legacy lookup for compatibility and resume recovery. | Preserve local fork behavior; ensure no upstream reconciliation from public endpoint. |
| `09351e94`, `c229f33e`, `1ffebbb5`, `06136af8`, `9de7a72c`, `1aab084e` auth/migration hardening bundles | `BLOCKED/PARTIAL` | Mixed auth identity, pending OAuth, migration and payment changes. Local has payment-only equivalents, but full commits require upstream auth identity schema or files absent in this fork, e.g. `auth_wechat_oauth.go`. | Do not cherry-pick. Extract only future payment-only tests if Kimi finds a gap. |
| `906802ab`, `a13ae5a0` mobile launch detection | `PARTIAL` | `paymentFlow.ts` mobile redirect logic and tests exist. | Verify frontend paymentFlow tests. |
| `f35e9675` QR fallback/admin guidance | `PARTIAL` | `HelpTooltip`, provider admin guidance, QR/payment status tests and button fixes exist. | Verify frontend tests. |
| `5240b444` inline payment flow/mobile support/renewal modal | `PARTIAL/FORK` | Current `/purchase` is locally adapted to fork UX and latest payment controls; upstream UI is not copied 1:1. | Do not force UI rewrite in this pass. |
| `794e8172` remove PaymentChannel / channel features | `BLOCKED` | Touches channel schema and route architecture. This branch already has payment provider instances without importing channel features. | Skip. |
| `ee3f158f`, `2cebb0dc`, `54dc1767`, `40f7e832` settings/wechat compatibility | `BLOCKED/PARTIAL` | Mixed with auth/channel-level WeChat config. Local payment config works for test. | No broad cherry-pick. |
| `faee59ee` API error reason/metadata propagation | `MUST/APPLIED` | Kimi found local API client dropped `reason` and `metadata`, breaking PaymentView-specific errors such as pending-order and cancel-limit messages. | Applied upstream-equivalent patch and added client test coverage. |
| `27cd2f8e` remove legacy purchase subscription settings | `SHOULD/APPLIED-FRONTEND` | Kimi found old iframe purchase settings still exposed in frontend admin settings. This conflicts with Payment v2 `/purchase`. | Removed frontend settings fields/UI/save payload; backend DTO remains for compatibility. |
| `58b2cc38` payment result resume tests | `PARTIAL` | Current `PaymentResultView.spec.ts` already contains stronger resume-token, stale snapshot, pending polling and legacy fallback coverage than the upstream initial test. | No extra backport needed. |
| `4ebdfcd1` admin visible method source tests | `BLOCKED/NOT-APPLICABLE` | Upstream test targets SettingsView visible-method source controls. This fork currently has backend visible-source support but no matching admin UI controls in `SettingsView.vue`; frontend test cannot be ported without adding UI scope. | Keep as future UI task, not Payment B-2 production blocker. |
| `147ed42a`, `1d8432b8`, `276ce052`, `d6a04bb7`, `f1297a36`, `c14d7393`, `1a0cabbf` | `PARTIAL/SKIP` | Kimi cross-check: equivalent return URL, resume/webhook routing, refund endpoint, source routing and allow-user-refund behavior already exists locally or is not applicable to current fork. | Preserve local implementation; verify tests. |

## Kimi Review Findings

Kimi reviewed this matrix on 2026-05-01 and found:

- P0: missing `faee59ee`; applied.
- P1: `purchase_subscription` frontend fields from `27cd2f8e`; frontend portion applied.
- P1: `1aab084e` needed explicit `BLOCKED`; recorded.
- P1: `58b2cc38` / `4ebdfcd1` test commits; `58b2cc38` is already covered locally, `4ebdfcd1` is not directly applicable without adding admin visible-source UI.
- P2: several additional payment-looking commits are already covered locally; recorded above.

Manual review additionally found the `fd0c9a13`/`61a008f7` provider config write path was not actually applied; this was fixed before further deployment.

## Raw Candidate Groups Considered

The path/subject scan also returned many unrelated commits. They are intentionally out of payment scope:

- websearch, notification, pricing, account stats, RPM and gateway commits
- channel monitor and available channels commits
- affiliate invite/rebate commits
- OpenAI/Codex/Vertex/gateway streaming and image commits
- profile/auth binding UI commits
- table default and ops cleanup commits

These must remain outside Payment B-2 unless the user opens a separate upstream-sync task.

## Backport Rules

- Do not cherry-pick mixed auth/profile/channel/affiliate commits.
- Do not introduce a new migration unless schema impact, test DB status, production prerequisite and rollback note are documented.
- Prefer local payment-only manual patches over broad upstream cherry-picks.
- Preserve fork-local payment repair migrations and deployed compatibility behavior.
- Treat `PARTIAL` as "must verify", not as "done by assumption".
- Treat `HOLD/PRESENT` fee/multiplier behavior as a product decision line: do not silently alter money semantics.

## Next Verification Batch

Verification run after applying the matrix review fixes:

```text
GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./...
result: pass when run with permission for httptest local listeners

GOCACHE="$PWD/../../.cache/go-build" go test -count=1 ./internal/payment ./internal/service -run 'Test.*Payment|Test.*Wechat|Test.*WeChat|Test.*Provider|Test.*Order|Test.*Refund|Test.*Fulfillment|Test.*Config|Test.*Resume|Test.*Visible|Test.*Webhook'
result: pass

GOCACHE="$PWD/../../.cache/go-build" go test -tags unit -count=1 ./internal/payment ./internal/handler -run 'TestDecryptConfig|TestApplyWeChatPaymentResumeClaims|TestVerifyOrderPublic|TestResolveOrderPublicByResumeToken|TestWriteSuccessResponse|TestUnknownOrderWebhookAcksWithSuccess|TestWebhookConstants|TestExtractOutTradeNo|TestVerifyNotificationWithProviders'
result: pass

pnpm exec vitest run src/api/__tests__/client.spec.ts src/api/__tests__/payment-contract.spec.ts src/__tests__/buttonClasses.spec.ts src/components/payment/__tests__ src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/paymentWechatResume.spec.ts src/views/user/__tests__/paymentUx.spec.ts src/router/__tests__/guards.spec.ts
result: 12 files, 99 tests passed

pnpm exec vue-tsc --noEmit
result: pass

git diff --check
result: pass
```

After local verification, deploy to test only after backing up test DB. Production remains out of scope for this plan.

## Current Change List And Impact

- Backend payment provider config persistence:
  - New writes are plaintext JSON.
  - Existing AES ciphertext remains readable when a valid legacy key is configured.
  - Unreadable legacy config is treated as empty so admins can re-enter it.
  - Admin provider list still masks sensitive fields.
  - Impact: fixes provider config loss/restart risk; requires keeping DB backups before deployment because stored format changes on next save.
- Frontend API error propagation:
  - `reason` and `metadata` now survive API interceptors.
  - Impact: payment page can show specific errors for pending-order limits, cancel limits and gateway-specific payment guidance.
- Frontend admin settings:
  - Removed old purchase subscription iframe settings UI and save payload.
  - Backend compatibility fields remain untouched.
  - Impact: avoids admins continuing to configure the pre-Payment-v2 subscription page. User-facing `/purchase` payment flow is unchanged.
- Documentation:
  - Added this upstream audit matrix and Kimi review notes for deployment reference.
