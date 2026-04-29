# Upstream Sync 2026-04 增量台账

**周期**: 2026-04-29 启动
**Range**: `origin/main..upstream/main` = 539 commits / 61 first-parent merge PRs
**上次同步**: 2026-03-29 → v0.1.105
**目标版本**: v0.1.119（中间态：拿大部分 fix + 部分 feat，hold payment/auth/affiliate）
**Worktree**: `.claude/worktrees/upstream-sync-2026-04/`
**Branch**: `worktree-upstream-sync-2026-04`

---

## 决策依据

| 项 | 决定 |
|---|---|
| **同步目标** | 中间态：fix + 部分 feat |
| **payment 域** | 全 hold（长期方向待最后定） |
| **auth 域** | 全 hold |
| **affiliate 域** | 全 hold |
| **粒度** | 按 upstream PR（first-parent merge）单元 cherry-pick |
| **打包方式** | 同主题切片，每切片一个 fork PR |
| **review** | 切片内最终 fork diff 给 kimi review |
| **验证** | 每切片 merge 前必跑 `pnpm --dir frontend build` + `cd backend && go build ./...` |

---

## 分类汇总

| 标签 | 数量 | 处理 |
|---|---|---|
| **AUTO-PICK** | 5 | 一个切片，自审 + kimi 抽审 |
| **KIMI-REVIEW** | 31 | 7 个主题切片，每片 kimi review |
| **HOLD** | 25 | 留给最后人工逐个对齐 |

---

## 切片计划

| 切片 | 主题 | PR 数 | 估行数 | 状态 |
|---|---|---|---|---|
| **slice-0-easy** | AUTO-PICK 集合 | 5 | ~93 | pending |
| **slice-1-anthropic** | Anthropic SSE/cache 修复 | 5 | ~530 | pending |
| **slice-2-openai-core** | OpenAI stream/failover/account | 8 | ~1300 | pending |
| **slice-3-codex** | Codex bridge/normalization/ids | 7 | ~1500 | pending |
| **slice-4-openai-image** | OpenAI image API（大）| 2 | ~2900 | pending |
| **slice-5-cc-mimicry** | cc-mimicry-parity（大）| 1 | ~1200 | pending |
| **slice-6-ops-misc** | quota/retention/scheduling | 5 | ~600 | pending |
| **slice-7-frontend** | datatable/sidebar 纯前端 | 2 | ~340 | pending |

注：切片划分初稿，实际 cherry-pick 时若遇依赖会调整。

---

## 完整 PR 台账

### AUTO-PICK (5)

| # | SHA | PR | Title | 文件 | +/- | 切片 | 状态 |
|---|---|---|---|---|---|---|---|
| 1 | `f5ee9379` | [#1753](https://github.com/Wei-Shaw/sub2api/pull/1753) | fix-orphaned-scheduled-tests | 1 | +3/-0 | slice-0 | pending |
| 2 | `8fd29082` | [#1663](https://github.com/Wei-Shaw/sub2api/pull/1663) | test-dialog-close-during-stream | 2 | +40/-34 | slice-0 | pending |
| 3 | `7c671b53` | [#1635](https://github.com/Wei-Shaw/sub2api/pull/1635) | fix-issue-1613-version-dropdown | 2 | +4/-1 | slice-0 | pending |
| 4 | `66bea2b5` | [#1624](https://github.com/Wei-Shaw/sub2api/pull/1624) | fix-issue-1613-version-dropdown | 2 | +11/-1 | slice-0 | pending |
| 5 | `9b7b3755` | [#1543](https://github.com/Wei-Shaw/sub2api/pull/1543) | messages-dispatch-i18n | 2 | +34/-6 | slice-0 | pending |

### KIMI-REVIEW (31)

| # | SHA | PR | Title | 文件 | +/- | 域 | 切片 | 状态 |
|---|---|---|---|---|---|---|---|---|
| 1 | `4d676ddd` | [#2066](https://github.com/Wei-Shaw/sub2api/pull/2066) | anthropic-stream-eof-failover | 2 | +265/-6 | anthropic,gateway | slice-1 | pending |
| 2 | `c92b88e3` | [#1996](https://github.com/Wei-Shaw/sub2api/pull/1996) | claude-code-read-empty-pages | 2 | +139/-2 | anthropic | slice-1 | pending |
| 3 | `41d06573` | [#1970](https://github.com/Wei-Shaw/sub2api/pull/1970) | claude-openai-cache-usage | 2 | +106/-12 | anthropic,openai | slice-1 | pending |
| 4 | `e70812f0` | [#1623](https://github.com/Wei-Shaw/sub2api/pull/1623) | anthropic-buffered-empty-output | 1 | +10/-1 | anthropic,gateway | slice-1 | pending |
| 5 | `82b840c1` | [#1587](https://github.com/Wei-Shaw/sub2api/pull/1587) | anthropic-credit-balance-400 | 1 | +6/-1 | anthropic | slice-1 | pending |
| 6 | `1afd81b0` | [#1920](https://github.com/Wei-Shaw/sub2api/pull/1920) | responses-web-search-tool-types | 1 | +1/-1 | anthropic | slice-1 | pending |
| 7 | `bf43fb4e` | [#2044](https://github.com/Wei-Shaw/sub2api/pull/2044) | openai-image-apikey-versioned-base-url | 3 | +179/-1 | openai | slice-2 | pending |
| 8 | `ed0c85a1` | [#2006](https://github.com/Wei-Shaw/sub2api/pull/2006) | openai-images-explicit-session | 3 | +66/-13 | gateway,openai | slice-2 | pending |
| 9 | `22b12775` | [#1948](https://github.com/Wei-Shaw/sub2api/pull/1948) | openai-account-test-responses-stream | 5 | +87/-8 | gateway,openai | slice-2 | pending |
| 10 | `aff98d5a` | [#1960](https://github.com/Wei-Shaw/sub2api/pull/1960) | openai-stream-keepalive-downstream-idle | 2 | +51/-5 | gateway,openai | slice-2 | pending |
| 11 | `5d1c12e6` | [#1943](https://github.com/Wei-Shaw/sub2api/pull/1943) | openai-responses-preoutput-failover | 2 | +428/-18 | gateway,openai | slice-2 | pending |
| 12 | `b95ffce2` | [#1772](https://github.com/Wei-Shaw/sub2api/pull/1772) | openai-test-state-reconciliation | 3 | +177/-6 | openai | slice-2 | pending |
| 13 | `358ff6a6` | [#1683](https://github.com/Wei-Shaw/sub2api/pull/1683) | dev-main (gateway/openai) | 1 | +22/-0 | gateway,openai | slice-2 | pending |
| 14 | `41fbdba1` | [#1687](https://github.com/Wei-Shaw/sub2api/pull/1687) | upstream-response-limit-dedup | 5 | +107/-89 | gateway,openai | slice-2 | pending |
| 15 | `76aae5aa` | [#1911](https://github.com/Wei-Shaw/sub2api/pull/1911) | codex-responses-payload-normalization | 2 | +372/-0 | codex,openai | slice-3 | pending |
| 16 | `1ce9dc03` | [#1895](https://github.com/Wei-Shaw/sub2api/pull/1895) | codex-spark-limitations | 5 | +257/-2 | codex,gateway,openai | slice-3 | pending |
| 17 | `15ce914a` | [#1910](https://github.com/Wei-Shaw/sub2api/pull/1910) | codex-tool-call-ids | 4 | +94/-7 | codex,openai | slice-3 | pending |
| 18 | `ff08f9d7` | [#1853](https://github.com/Wei-Shaw/sub2api/pull/1853) | codex-image-generation-bridge | 8 | +406/-3 | codex,gateway,openai | slice-3 | pending |
| 19 | `327da8e2` | [#1813](https://github.com/Wei-Shaw/sub2api/pull/1813) | meteor041/fix-openai-image-handling | 8 | +484/-20 | codex,gateway,openai | slice-3 | pending |
| 20 | `ad6c3281` | [#1575](https://github.com/Wei-Shaw/sub2api/pull/1575) | cursor-responses-body-compat | 4 | +301/-12 | codex,gateway,openai | slice-3 | pending |
| 21 | `e6e73b4f` | [#1690](https://github.com/Wei-Shaw/sub2api/pull/1690) | ws-codex-scheduler-cache-1662 | 6 | +108/-6 | codex,frontend,openai | slice-3 | pending |
| 22 | `32107b4f` | [#1795](https://github.com/Wei-Shaw/sub2api/pull/1795) | openai-image-api-sync | 17 | +2805/-46 | gateway,openai | slice-4 | pending |
| 23 | `70d0569f` | [#1668](https://github.com/Wei-Shaw/sub2api/pull/1668) | tyqy12/main | 7 | +80/-190 | frontend,gateway,openai | slice-4 | pending |
| 24 | `6d20ab80` | [#1914](https://github.com/Wei-Shaw/sub2api/pull/1914) | cc-mimicry-parity | 13 | +1119/-76 | anthropic,gateway | slice-5 | pending |
| 25 | `a16c6650` | [#2090](https://github.com/Wei-Shaw/sub2api/pull/2090) | ops-retention-zero | 6 | +167/-35 | frontend,i18n | slice-6 | pending |
| 26 | `8dbbd942` | [#1836](https://github.com/Wei-Shaw/sub2api/pull/1836) | account-daily-weekly-quota-cache | 2 | +107/-2 | quota | slice-6 | pending |
| 27 | `e8be4344` | [#1752](https://github.com/Wei-Shaw/sub2api/pull/1752) | quota-exceeded-scheduling | 7 | +207/-35 | frontend,gateway | slice-6 | pending |
| 28 | `061fd48d` | [#1749](https://github.com/Wei-Shaw/sub2api/pull/1749) | xhigh-reasoning-effort | 3 | +11/-3 | frontend,gateway | slice-6 | pending |
| 29 | `c22d11ce` | [#1702](https://github.com/Wei-Shaw/sub2api/pull/1702) | outbox-watermark-context-dedup | 1 | +57/-25 | misc | slice-6 | pending |
| 30 | `d949acb1` | [#1603](https://github.com/Wei-Shaw/sub2api/pull/1603) | datatable-mobile-double-render | 2 | +177/-18 | frontend | slice-7 | pending |
| 31 | `16126a2c` | [#1545](https://github.com/Wei-Shaw/sub2api/pull/1545) | smooth-sidebar-collapse | 2 | +160/-47 | frontend | slice-7 | pending |

### HOLD (25)

| # | SHA | PR | Title | 文件 | +/- | 原因 | 备注 |
|---|---|---|---|---|---|---|---|
| 1 | `63ef2310` | [#1977](https://github.com/Wei-Shaw/sub2api/pull/1977) | sholiverlee/vertex | 19 | +1330/-36 | wire_gen | Vertex Service Account |
| 2 | `b0a2252e` | [#2051](https://github.com/Wei-Shaw/sub2api/pull/2051) | openai-fast-flex-policy | 23 | +2820/-10 | wire_gen | Fast/Flex Policy |
| 3 | `641e6107` | [#1940](https://github.com/Wei-Shaw/sub2api/pull/1940) | bump-codex-cli-version | 4 | +7/-7 | auth | codex CLI 0.104→0.125 |
| 4 | `5b5db885` | [#1897](https://github.com/Wei-Shaw/sub2api/pull/1897) | codex/invite-affiliate-rebate | 33 | +1744/-42 | affiliate,auth,migration,payment,wire_gen | 邀请返利系统 |
| 5 | `ac114738` | [#1850](https://github.com/Wei-Shaw/sub2api/pull/1850) | feat/channel-insights | 151 | +35316/-615 | ent_schema,migration,payment,wire_gen | 大重构 |
| 6 | `827a4498` | [#1829](https://github.com/Wei-Shaw/sub2api/pull/1829) | codex-oauth-proxy-message | 6 | +99/-3 | auth | OAuth 代理 |
| 7 | `6b0cf466` | [#1815](https://github.com/Wei-Shaw/sub2api/pull/1815) | feat_rpm | 79 | +2831/-140 | auth,billing,ent_schema,migration,wire_gen | RPM 计费 |
| 8 | `27ffc7f3` | [#1828](https://github.com/Wei-Shaw/sub2api/pull/1828) | wx-11/main | 18 | +2032/-363 | wire_gen | gateway/openai |
| 9 | `ddf80f5e` | [#1799](https://github.com/Wei-Shaw/sub2api/pull/1799) | rebuild/auth-identity-foundation | 140 | +11032/-1181 | auth,ci,deploy,ent_schema,migration,payment | 身份重构 |
| 10 | `8eb3f9e7` | [#1785](https://github.com/Wei-Shaw/sub2api/pull/1785) | rebuild/auth-identity-foundation | 279 | +79981/-12626 | auth,billing,ent_schema,migration,payment,wire_gen | 巨型 |
| 11 | `ffc9c387` | [#1766](https://github.com/Wei-Shaw/sub2api/pull/1766) | codex-drop-removed-models | 11 | +84/-265 | billing | 删 model |
| 12 | `a8854947` | [#1764](https://github.com/Wei-Shaw/sub2api/pull/1764) | wxpay-pubkey-hardening | 20 | +335/-87 | payment | wxpay 公钥 |
| 13 | `51af8df3` | [#1731](https://github.com/Wei-Shaw/sub2api/pull/1731) | rate-billing-autofill-response-limit | 30 | +585/-186 | billing,payment | 计费 |
| 14 | `1db32d69` | [#1666](https://github.com/Wei-Shaw/sub2api/pull/1666) | account-cost-display | 16 | +191/-39 | migration | 账户成本 |
| 15 | `9bf079b7` | [#1655](https://github.com/Wei-Shaw/sub2api/pull/1655) | payment-fee-multiplier | 28 | +432/-182 | payment | 手续费 |
| 16 | `d402e722` | [#1637](https://github.com/Wei-Shaw/sub2api/pull/1637) | websearch-notify-pricing | 177 | +13643/-1201 | auth,billing,ci,ent_schema,migration,payment,wire_gen | 大功能 |
| 17 | `7d80b5ad` | [#1610](https://github.com/Wei-Shaw/sub2api/pull/1610) | alipay-wxpay-type-mapping | 5 | +22/-19 | payment | 支付 |
| 18 | `97f14b7a` | [#1572](https://github.com/Wei-Shaw/sub2api/pull/1572) | feat/payment-system-v2 | 174 | +45730/-3073 | ci,ent_schema,migration,payment,wire_gen | upstream payment v2 |
| 19 | `1ef3782d` | [#1538](https://github.com/Wei-Shaw/sub2api/pull/1538) | fix/bug-cleanup-main | 117 | +3961/-870 | auth | bug cleanup |
| 20 | `a0b5e5bf` | [#1973](https://github.com/Wei-Shaw/sub2api/pull/1973) | Nobody-Zhang/main (Zpay refund) | 2 | +371/-21 | payment(zpay) | 用户决定 hold |
| 21 | `79aff2df` | [#1810](https://github.com/Wei-Shaw/sub2api/pull/1810) | profile-auth-bindings-i18n | 20 | +845/-43 | auth | 用户决定 hold |
| 22 | `1da4bd72` | [#1802](https://github.com/Wei-Shaw/sub2api/pull/1802) | profile-auth-bindings-i18n | 8 | +139/-7 | auth | 用户决定 hold |
| 23 | `54490cf6` | [#1576](https://github.com/Wei-Shaw/sub2api/pull/1576) | feat/payment-docs | 9 | +578/-10 | payment | 用户决定 hold |
| 24 | `75908800` | [#1612](https://github.com/Wei-Shaw/sub2api/pull/1612) | qrcode-density | 3 | +4/-4 | payment(域) | 用户决定 hold |
| 25 | `6c73b621` | [#1734](https://github.com/Wei-Shaw/sub2api/pull/1734) | payment-recommend-kyren-topup | 2 | +14/-4 | payment(域) | 用户决定 hold |

---

## 切片执行日志

> 每个切片 merge 后回填这里。SQL/migration 改动单独详记。

### slice-0-easy
- **状态**: pending
- **分支**: `sync/2026-04/slice-0-easy`
- **PR**:
- **kimi review**:
- **build 验证**:
- **冲突**:
- **SQL 改动**: 无（AUTO-PICK 不含 migration）
- **merge SHA**:

### slice-1-anthropic
- **状态**: pending
- **分支**: `sync/2026-04/slice-1-anthropic`
- **PR**:
- **kimi review**:
- **build 验证**:
- **冲突**:
- **SQL 改动**:
- **merge SHA**:

### slice-2-openai-core
- **状态**: pending
（同上结构）

### slice-3-codex
（同上结构）

### slice-4-openai-image
（同上结构）

### slice-5-cc-mimicry
（同上结构）

### slice-6-ops-misc
（同上结构）

### slice-7-frontend
（同上结构）

---

## SQL/Migration 变更登记

> 任何切片涉及 `backend/migrations/` 或 `backend/ent/schema/` 必须在此记录。
> 部署前必须 pg_dump 留底（参见 memory: feedback_pre_deploy_dump.md）。

**当前**: 计划内 0 个 migration（payment/auth/affiliate 全 hold，主流 slice 都不涉 migration）。

| 切片 | 文件 | 表/列 | DDL 摘要 | 部署状态 | pg_dump |
|---|---|---|---|---|---|
| — | — | — | — | — | — |

---

## 待人工对齐的关键决策（Phase 2）

延期到 KIMI-REVIEW 全部 merge 后处理：

1. **payment 域长期方向**（Q-C 未决）
   - 选项 A：永久跳过上游 payment，长期维护 wxpay 自研
   - 选项 B：切到上游 payment v2 框架 + 加 wxpay provider
   - 涉及 PR：#1572, #1610, #1655, #1731, #1764, #1973, #1576, #1612, #1734
   - **撞表风险**：upstream `092_payment_orders.sql` vs fork `077_add_payment_tables.sql` 都建 `payment_orders`，schema 不同，**不能并存**。需要先确认生产 schema_migrations 已应用情况。

2. **auth-identity-foundation**（#1785/#1799）
   - upstream 重构身份层（80k 行级别），可能影响 OAuth/微信登录/OIDC
   - 我们已有自研 referral_code，需评估是否能融合

3. **affiliate-rebate**（#1897）
   - 我们 fork 已有自研 referral_service.go + ent schema referral_code
   - 上游 affiliate 是另一套设计，需选择保留哪边

4. **大 feat**：channel-insights #1850, websearch-notify-pricing #1637, feat_rpm #1815, vertex #1977, fast-flex #2051
   - 单独评估业务价值后决定是否引入

---

## 历史链接

- 上游同步通用流程：[`upstream-sync.md`](./upstream-sync.md)
- Git workflow：[`git-workflow.md`](./git-workflow.md)
- 部署：[`deployment.md`](./deployment.md)
