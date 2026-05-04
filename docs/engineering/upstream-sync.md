# Upstream 同步指南

upstream 仓库：`Wei-Shaw/sub2api`（`upstream` remote）

我们的 fork：`ddnio/sub2api`（`origin` remote）

本仓库当前采用 release gate 同步模型。旧的“直接 merge `upstream/main` 后处理冲突”只保留为历史背景，不再作为日常同步流程。

## 当前原则

1. 按 upstream tag 顺序处理，例如 `v0.1.114..v0.1.115` 完整关闭后，才能进入 `v0.1.115..v0.1.116`。
2. release 是计划、验证、部署、打 fork marker tag 的边界。
3. 代码合入单元优先是 upstream first-parent commit / upstream PR merge commit；必要时可拆成更小的行为子项。
4. 不允许只因为冲突大、迁移复杂、产品含义复杂就跳过。每个 release item 最终必须有 `MERGED`、`ADAPTED`、`PRESENT`、`REJECTED`、`FROZEN` 或 `SKIP` 之一，并有证据。
5. `HOLD`、`PORT`、`PARTIAL`、`REOPENED` 都不是 release closeout 的最终状态。
6. 已推送的 `fork/vX.Y.Z` marker tag 不移动、不删除。发现旧 gate 有缺口时，在最新 `main` 上 forward-fix。
7. 默认一个完整 upstream release 一个 fork PR。只有 schema/migration、auth/payment/security/data-risk、大冲突、大 CI unblock 才拆 PR。
8. 常规 runtime 变更等整个 release gate 完成后统一部署 test/prod 验证；安全热修、迁移/schema、auth/payment/data-risk、紧急线上修复可以例外提前部署。

详细 ledger 和当前 gate 状态见：

- `docs/engineering/upstream-release-coverage-2026-05.md`
- `docs/plans/2026-05-03-upstream-release-sync-reset.md`
- `docs/plans/2026-05-04-upstream-release-continuation.md`

## 每个 release 的固定流程

### 1. 同步基线

```bash
git fetch origin
git fetch upstream
git switch main
git merge --ff-only origin/main
```

如果 GitHub HTTPS 出现 `HTTP2 framing layer`、`Empty reply from server` 或 timeout，不要反复重试同一路径。按 ledger 里的 repeated-issue log 切到：

```bash
git -c http.version=HTTP/1.1 fetch upstream
```

仍失败时，使用 SSH remote 或 GitHub API fallback。不能在 upstream ref 未确认时扩大 release scope。

### 2. 建立 release worktree

```bash
git worktree add -b sync/v0.1.115-release .claude/worktrees/release-v0.1.115 origin/main
```

根 checkout 不直接开发。每个 release 从最新 `origin/main` 开始，不从旧 worktree 或旧分支继续推断状态。

### 3. 生成 release commit 清单

必须同时记录 first-parent 和完整 commit 列表：

```bash
git log --oneline --first-parent --reverse v0.1.114..v0.1.115
git log --oneline --reverse v0.1.114..v0.1.115
```

first-parent 列表用于确认 upstream 主线 PR/merge 顺序；完整列表用于防止 merge PR 内部 commit 被漏审。每个 merge PR 还要展开内部 commit 或 PR branch 历史，确认其内部 commit 都映射到同一个 release 决策或子项。

### 4. 对每个 upstream item 做直接导入审计

处理顺序：

1. 先确认该 commit/PR 是否已经是 fork ancestor，或已经通过某个 fork PR 落地。
2. 如果未落地，在隔离 worktree 里 preview/cherry-pick upstream commit 或 merge PR。
3. 如果补丁干净、范围小、不会覆盖 fork 的产品/数据语义，按 upstream commit/PR 单元合入。
4. 如果冲突大或跨 fork 架构，先写 import audit，记录：
   - 可直接 port 的文件；
   - fork 已有等价行为；
   - 冲突区域；
   - schema/migration/Ent 影响；
   - 产品语义影响；
   - 必须补的测试；
   - 最小可拆子项。
5. 只有完成 audit 后，才允许手工 port 子项、标记 `PRESENT`、`ADAPTED`、`REJECTED` 或 `FROZEN`。

遇到大 diff、冲突多、branch/worktree 状态异常、或实现开始明显偏离 upstream 原始 commit 时，立即暂停实现并写 import audit。继续手写代码通常会扩大偏差。

### 5. 保留 fork 关键行为

每次合 upstream 都必须显式检查这些 fork 语义没有被覆盖：

| 模块 | 重点 |
| --- | --- |
| Payment B-2 | provider instance、provider snapshot、resume token、微信/Stripe/EasyPay flow、refund、充值/订阅订单、真实数据库迁移 |
| Auth/OAuth/OIDC | 现有用户登录、OIDC synthetic email、pending OAuth、profile binding、session/cookie 兼容 |
| Referral / Affiliate | fork 已有邀请/返利语义，不被 upstream affiliate 重构静默替换 |
| Channel / Routing | fork 的 channel/provider routing、OpenAI/Codex/Anthropic compatibility、model whitelist |
| Migrations / Ent | 不覆盖已上线 migration；新增 migration 前检查真实 DB 和 `schema_migrations` |
| Frontend admin/user UI | 保留 fork 已有支付、账户、设置、表格偏好、i18n 和部署路径 |

### 6. 数据兼容 gate

触及 auth identity、邀请码/referral/affiliate、payment、subscription、用户余额、provider 配置、Ent schema 或 migration 的 upstream item，必须先过数据兼容 gate：

1. 用只读 SQL 查 ToC test、ToC prod，以及共享 runtime 受影响时的 ToB prod，记录表结构、settings、row count、已处理/待处理状态。
2. 写清 fork 现有概念和 upstream 新概念的映射。没有一一映射时，不能硬替换，只能 `ADAPTED`、`REJECTED` 或 `FROZEN`。
3. 任何 accepted schema/data change 都要有 forward migration、backfill、幂等保护、回滚/restore 路径和测试。
4. 已有线上数据不能被删除、重新生成、重复发奖、重复扣款、静默解绑或失去审计记录。

邀请码/referral/affiliate 的固定边界：

- `redeem_codes(type='invitation')` 是 fork 的注册准入码，不等同于 upstream affiliate。
- `users.referral_code` + `user_referrals` 是 fork 的推荐归因和首次充值奖励状态。
- 如果后续采用 upstream `user_affiliates` / `user_affiliate_ledger`，必须从现有 `users.referral_code` 和 `user_referrals` 做兼容 backfill；已发奖记录必须标记为 processed，待发奖记录必须继续可发或有明确迁移语义。

## 验证规则

按变更范围选择验证，不用每个小 commit 都全量部署。

| 变更类型 | 本地验证 | 部署 |
| --- | --- | --- |
| docs/ledger only | `git diff --check`、self-review | 不部署 |
| backend runtime | targeted `go test -tags=unit ...`；共享服务改动再跑更广 package tests | release 完成后统一部署，除非高风险例外 |
| frontend runtime | targeted Vitest、`pnpm --dir frontend typecheck`，必要时 build | release 完成后统一部署 |
| payment/auth/data-risk | backend/frontend targeted regression、真实配置/DB preflight | 可提前 test/prod 部署验证，并记录 |
| migration/schema | migration test、真实 DB read-only precheck、backup plan | 必须 test 后 prod，记录 schema/log 验证 |

release-level deployment closeout 至少包含：

```bash
curl -s <test-or-prod>/health
curl -s -o /dev/null -w '%{http_code}' <test-or-prod>/v1/models
docker logs --since <deploy-time> <container> | egrep -i 'panic|fatal|error|migration|failed|traceback|异常'
```

预期：

- `/health` 返回 `{"status":"ok"}`；
- 未授权 `/v1/models` 返回 401；
- 严重日志扫描无异常。

## Closeout 检查清单

release 完成前必须确认：

1. `git log --first-parent` 的每个 mainline entry 都在 ledger 中有最终 outcome。
2. `git log --reverse` 的完整 commit 数已记录，merge PR 内部 commit 已展开或映射。
3. release 内没有未解决的 `HOLD`、`PORT`、`PARTIAL`、`REOPENED`。
4. 每个 runtime/code item 有本地测试证据，PR CI 已检查。
5. 上游同步/release-gate 相关 docs/runtime PR 默认需要 Kimi review；如果某次低风险变更不用 Kimi，必须在对应 release ledger 或 PR 说明中记录原因、替代 review 方式和验证证据。
6. release-level deploy 决策已记录；需要部署时 test/prod 均已验证。
7. `backend/cmd/server/VERSION` 已按 fork marker 语义更新。
8. `fork/vX.Y.Z` annotated tag 已创建并推送。
9. 合并后拉最新 `origin/main`，再从最新 main 创建下一个 release worktree。
10. 已合并的旧 worktree/branch 在确认无未提交改动后清理。

## 历史说明

早期曾尝试过跨多个 upstream release 的 slice-based 同步，也曾保留过“全量 merge upstream/main”指南。当前实际经验表明，这会掩盖未处理的 release item，也容易覆盖 fork 的 payment/auth/migration 语义。因此后续以 release gate + itemized ledger 为准。

历史记录仍可参考：

- `docs/engineering/upstream-sync-2026-04.md`
- `docs/engineering/upstream-sync-2026-05-phase2.md`
- `docs/engineering/payment-b2-deploy.md`
