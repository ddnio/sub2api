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
| **review** | 每切片 PR 自动调 kimi review fork diff |
| **验证** | 每切片 merge 前必跑 `pnpm --dir frontend build` + `cd backend && go build ./...` + 必要时 `go generate ./ent/...` |

---

## 编译漂移风险与应对策略

KIMI-REVIEW 切片在 cherry-pick 上游 PR 时，可能依赖 HOLD 列表里的 wire/ent/model 改动，导致 cherry-pick 后**编译失败**。每切片处理流程必须包含以下检查：

| 风险 | 来源 | 应对 |
|---|---|---|
| **wire_gen.go 漂移** | HOLD #1977/#2051/#1828/#1897/#1850/#1815/#1785/#1572 都改 wire_gen | KIMI-REVIEW 切片若 cherry-pick 后引入新 service provider，**必须手改 wire_gen.go**（wire 命令 Go 1.26 不可用）；编译报 undefined provider 时优先查 wire_gen |
| **ent schema 漂移** | HOLD #1850/#1815/#1799/#1785/#1637/#1572 改 ent schema | KIMI-REVIEW 代码若调用新 schema 的 getter/setter，会缺方法。处理：cherry-pick 后跑 `cd backend && docker run --rm -v "$PWD":/app -w /app golang:1.26 go generate ./ent/...`；仍缺方法说明上游 schema 字段没拉过来，按 [`upstream-sync.md`](./upstream-sync.md) 第 4 节处理 |
| **model 删除** | KIMI-REVIEW #1766 (codex-drop-removed-models) 删除废弃 model | 已提到 slice-3 优先消化，避免后续 slice 引用废弃 model 失败 |
| **依赖升级** | axios/codex CLI 等 deps PR 在 HOLD 内 | KIMI-REVIEW 代码若调新 API 签名会编译失败；遇到则单独评估 deps PR 是否要先入 |

**每切片 cherry-pick 完最终验证清单**：

```bash
# 1. 后端编译
cd backend && go build ./... 2>&1 | head -50

# 2. 若涉及 ent schema 改动
docker run --rm -v "$PWD":/app -w /app golang:1.26 go generate ./ent/...

# 3. 后端测试（核心包）
go test ./internal/service/... 2>&1 | tail -20

# 4. 前端编译
cd ../frontend && pnpm build 2>&1 | tail -30
```

任何一步失败 → 切片整体回滚到 HOLD 重新评估。

---

## 分类汇总

| 标签 | 数量 | 处理 |
|---|---|---|
| **DONE（已合入 main）** | 23 | 全部 7 个 sync slice PR 已 merge |
| **HOLD（待人工对齐）** | 38 | 见 Phase 2 章节 |

注：
- AUTO-PICK 原 5：slice-0 #1624/#1635 sidebar 重构 skip → 实际 3
- KIMI-REVIEW 原 31 + 从 HOLD 提 #1766 = 32；slice-2 转 #1943/#1960，slice-3 转 #1766/#1895，slice-4 整片转 5 个（#1795/#1813/#1853/#2006/#2044）→ 23
- HOLD 原 24 + slice 累计转入 11（#1624/#1635/#1943/#1960/#1766/#1895/#1795/#1813/#1853/#2006/#2044）= 35

---

## 切片计划

| 切片 | 主题 | PR 数 | 估行数 | 状态 |
|---|---|---|---|---|
| **slice-0-easy** | AUTO-PICK 集合 | 5 | ~93 | pending |
| **slice-1-anthropic** | Anthropic SSE/cache 修复 | 6 | ~530 | pending |
| **slice-2-openai-core** | OpenAI stream/failover/account/quota | 7 | ~1230 | pending |
| **slice-3-codex** | Codex bridge/normalization/ids + model cleanup | 6 | ~1300 | pending |
| **slice-4-openai-image** | 全部 image API 改动统一片 | 5 | ~3940 | pending |
| **slice-5-cc-mimicry** | cc-mimicry-parity（大）| 1 | ~1200 | pending |
| **slice-6-ops-misc** | quota/retention/scheduling | 5 | ~600 | pending |
| **slice-7-frontend** | datatable/sidebar 纯前端 | 2 | ~340 | pending |

切片内部 cherry-pick 顺序：按 upstream merge 时间正序，避免依赖 PR 在依赖者后被合。

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

### KIMI-REVIEW (32)

| # | SHA | PR | Title | 文件 | +/- | 域 | 切片 | 状态 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| 1 | `4d676ddd` | [#2066](https://github.com/Wei-Shaw/sub2api/pull/2066) | anthropic-stream-eof-failover | 2 | +265/-6 | anthropic,gateway | slice-1 | pending | |
| 2 | `c92b88e3` | [#1996](https://github.com/Wei-Shaw/sub2api/pull/1996) | claude-code-read-empty-pages | 2 | +139/-2 | anthropic | slice-1 | pending | |
| 3 | `41d06573` | [#1970](https://github.com/Wei-Shaw/sub2api/pull/1970) | claude-openai-cache-usage | 2 | +106/-12 | anthropic,openai | slice-1 | pending | |
| 4 | `e70812f0` | [#1623](https://github.com/Wei-Shaw/sub2api/pull/1623) | anthropic-buffered-empty-output | 1 | +10/-1 | anthropic,gateway | slice-1 | pending | |
| 5 | `82b840c1` | [#1587](https://github.com/Wei-Shaw/sub2api/pull/1587) | anthropic-credit-balance-400 | 1 | +6/-1 | anthropic | slice-1 | pending | |
| 6 | `1afd81b0` | [#1920](https://github.com/Wei-Shaw/sub2api/pull/1920) | responses-web-search-tool-types | 1 | +1/-1 | anthropic | slice-1 | pending | |
| 7 | `22b12775` | [#1948](https://github.com/Wei-Shaw/sub2api/pull/1948) | openai-account-test-responses-stream | 5 | +87/-8 | gateway,openai | slice-2 | pending | |
| 8 | `aff98d5a` | [#1960](https://github.com/Wei-Shaw/sub2api/pull/1960) | openai-stream-keepalive-downstream-idle | 2 | +51/-5 | gateway,openai | slice-2 | pending | |
| 9 | `5d1c12e6` | [#1943](https://github.com/Wei-Shaw/sub2api/pull/1943) | openai-responses-preoutput-failover | 2 | +428/-18 | gateway,openai | slice-2 | pending | |
| 10 | `b95ffce2` | [#1772](https://github.com/Wei-Shaw/sub2api/pull/1772) | openai-test-state-reconciliation | 3 | +177/-6 | openai | slice-2 | pending | |
| 11 | `358ff6a6` | [#1683](https://github.com/Wei-Shaw/sub2api/pull/1683) | dev-main (gateway/openai) | 1 | +22/-0 | gateway,openai | slice-2 | pending | 标题模糊，实际是 gateway openai 小 fix |
| 12 | `41fbdba1` | [#1687](https://github.com/Wei-Shaw/sub2api/pull/1687) | upstream-response-limit-dedup | 5 | +107/-89 | gateway,openai | slice-2 | pending | |
| 13 | `70d0569f` | [#1668](https://github.com/Wei-Shaw/sub2api/pull/1668) | tyqy12/main (rate limit fix) | 7 | +80/-190 | gateway,openai | slice-2 | pending | OpenAI 账号限流回流误判 7d/5h 窗口逻辑修复，非 image |
| 14 | `76aae5aa` | [#1911](https://github.com/Wei-Shaw/sub2api/pull/1911) | codex-responses-payload-normalization | 2 | +372/-0 | codex,openai | slice-3 | pending | |
| 15 | `1ce9dc03` | [#1895](https://github.com/Wei-Shaw/sub2api/pull/1895) | codex-spark-limitations | 5 | +257/-2 | codex,gateway,openai | slice-3 | pending | |
| 16 | `15ce914a` | [#1910](https://github.com/Wei-Shaw/sub2api/pull/1910) | codex-tool-call-ids | 4 | +94/-7 | codex,openai | slice-3 | pending | |
| 17 | `ad6c3281` | [#1575](https://github.com/Wei-Shaw/sub2api/pull/1575) | cursor-responses-body-compat | 4 | +301/-12 | codex,gateway,openai | slice-3 | pending | |
| 18 | `e6e73b4f` | [#1690](https://github.com/Wei-Shaw/sub2api/pull/1690) | ws-codex-scheduler-cache-1662 | 6 | +108/-6 | codex,frontend,openai | slice-3 | pending | |
| 19 | `ffc9c387` | [#1766](https://github.com/Wei-Shaw/sub2api/pull/1766) | codex-drop-removed-models | 11 | +84/-265 | codex,frontend,openai | slice-3 | pending | 从 HOLD 提升：纯删除废弃 model，先消化避免后续 slice 漂移 |
| 20 | `bf43fb4e` | [#2044](https://github.com/Wei-Shaw/sub2api/pull/2044) | openai-image-apikey-versioned-base-url | 3 | +179/-1 | openai,image | slice-4 | pending | image 主题统一 |
| 21 | `ed0c85a1` | [#2006](https://github.com/Wei-Shaw/sub2api/pull/2006) | openai-images-explicit-session | 3 | +66/-13 | gateway,openai,image | slice-4 | pending | image 主题统一 |
| 22 | `ff08f9d7` | [#1853](https://github.com/Wei-Shaw/sub2api/pull/1853) | codex-image-generation-bridge | 8 | +406/-3 | codex,gateway,openai,image | slice-4 | pending | image 主题统一 |
| 23 | `327da8e2` | [#1813](https://github.com/Wei-Shaw/sub2api/pull/1813) | meteor041/fix-openai-image-handling | 8 | +484/-20 | codex,gateway,openai,image | slice-4 | pending | image 主题统一 |
| 24 | `32107b4f` | [#1795](https://github.com/Wei-Shaw/sub2api/pull/1795) | openai-image-api-sync | 17 | +2805/-46 | gateway,openai,image | slice-4 | pending | image 主题主体改动 |
| 25 | `6d20ab80` | [#1914](https://github.com/Wei-Shaw/sub2api/pull/1914) | cc-mimicry-parity | 13 | +1119/-76 | anthropic,gateway | slice-5 | pending | 单独片，规模大 |
| 26 | `a16c6650` | [#2090](https://github.com/Wei-Shaw/sub2api/pull/2090) | ops-retention-zero | 6 | +167/-35 | frontend,i18n | slice-6 | pending | |
| 27 | `8dbbd942` | [#1836](https://github.com/Wei-Shaw/sub2api/pull/1836) | account-daily-weekly-quota-cache | 2 | +107/-2 | quota | slice-6 | pending | |
| 28 | `e8be4344` | [#1752](https://github.com/Wei-Shaw/sub2api/pull/1752) | quota-exceeded-scheduling | 7 | +207/-35 | frontend,gateway | slice-6 | pending | |
| 29 | `061fd48d` | [#1749](https://github.com/Wei-Shaw/sub2api/pull/1749) | xhigh-reasoning-effort | 3 | +11/-3 | frontend,gateway | slice-6 | pending | |
| 30 | `c22d11ce` | [#1702](https://github.com/Wei-Shaw/sub2api/pull/1702) | outbox-watermark-context-dedup | 1 | +57/-25 | misc | slice-6 | pending | outbox 水位线去重 |
| 31 | `d949acb1` | [#1603](https://github.com/Wei-Shaw/sub2api/pull/1603) | datatable-mobile-double-render | 2 | +177/-18 | frontend | slice-7 | pending | |
| 32 | `16126a2c` | [#1545](https://github.com/Wei-Shaw/sub2api/pull/1545) | smooth-sidebar-collapse | 2 | +160/-47 | frontend | slice-7 | pending | |

### HOLD (24)

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
| 11 | `a8854947` | [#1764](https://github.com/Wei-Shaw/sub2api/pull/1764) | wxpay-pubkey-hardening | 20 | +335/-87 | payment | wxpay 公钥 |
| 12 | `51af8df3` | [#1731](https://github.com/Wei-Shaw/sub2api/pull/1731) | rate-billing-autofill-response-limit | 30 | +585/-186 | billing,payment | 计费 |
| 13 | `1db32d69` | [#1666](https://github.com/Wei-Shaw/sub2api/pull/1666) | account-cost-display | 16 | +191/-39 | migration | 账户成本 |
| 14 | `9bf079b7` | [#1655](https://github.com/Wei-Shaw/sub2api/pull/1655) | payment-fee-multiplier | 28 | +432/-182 | payment | 手续费 |
| 15 | `d402e722` | [#1637](https://github.com/Wei-Shaw/sub2api/pull/1637) | websearch-notify-pricing | 177 | +13643/-1201 | auth,billing,ci,ent_schema,migration,payment,wire_gen | 大功能 |
| 16 | `7d80b5ad` | [#1610](https://github.com/Wei-Shaw/sub2api/pull/1610) | alipay-wxpay-type-mapping | 5 | +22/-19 | payment | 支付 |
| 17 | `97f14b7a` | [#1572](https://github.com/Wei-Shaw/sub2api/pull/1572) | feat/payment-system-v2 | 174 | +45730/-3073 | ci,ent_schema,migration,payment,wire_gen | upstream payment v2 |
| 18 | `1ef3782d` | [#1538](https://github.com/Wei-Shaw/sub2api/pull/1538) | fix/bug-cleanup-main | 117 | +3961/-870 | auth | bug cleanup |
| 19 | `a0b5e5bf` | [#1973](https://github.com/Wei-Shaw/sub2api/pull/1973) | Nobody-Zhang/main (Zpay refund) | 2 | +371/-21 | payment(zpay) | 用户决定 hold |
| 20 | `79aff2df` | [#1810](https://github.com/Wei-Shaw/sub2api/pull/1810) | profile-auth-bindings-i18n | 20 | +845/-43 | auth | 用户决定 hold |
| 21 | `1da4bd72` | [#1802](https://github.com/Wei-Shaw/sub2api/pull/1802) | profile-auth-bindings-i18n | 8 | +139/-7 | auth | 用户决定 hold |
| 22 | `54490cf6` | [#1576](https://github.com/Wei-Shaw/sub2api/pull/1576) | feat/payment-docs | 9 | +578/-10 | payment | 用户决定 hold |
| 23 | `75908800` | [#1612](https://github.com/Wei-Shaw/sub2api/pull/1612) | qrcode-density | 3 | +4/-4 | payment(域) | 用户决定 hold |
| 24 | `6c73b621` | [#1734](https://github.com/Wei-Shaw/sub2api/pull/1734) | payment-recommend-kyren-topup | 2 | +14/-4 | payment(域) | 用户决定 hold |

---

## 切片执行日志

> 每个切片 merge 后回填这里。SQL/migration/wire/ent 改动单独详记到下方变更登记区。

### slice-0-easy
- **状态**: ✅ done（merged at d80a3827）
- **分支**: `sync/2026-04/slice-0-easy`
- **PR**: [ddnio/sub2api#9](https://github.com/ddnio/sub2api/pull/9)
- **计划包含 PR**: #1753, #1663, #1635, #1624, #1543
- **实际 cherry-pick**: #1543, #1663, #1753（3 个成功）
- **跳过**: #1624 / #1635 — 上游 sidebar version dropdown 修复，fork sidebar 已重构（floating-contact + sora 移除），冲突过大且功能不适用，downgrade 到 HOLD
- **kimi review**: approve（无 critical，2 个 nit 已记录）
- **build 验证**: ✅ go build / go test / pnpm build / vue-tsc 全部通过
- **冲突**: 仅 i18n/zh.ts 自动 merge 通过；sidebar 两个 PR 因 fork 重构 skip
- **Schema/Generated Code 改动**: 无
- **merge SHA**: d80a3827

### slice-1-anthropic
- **状态**: ✅ done（merged at a53527fa）
- **分支**: `sync/2026-04/slice-1-anthropic`
- **PR**: [ddnio/sub2api#10](https://github.com/ddnio/sub2api/pull/10)
- **包含 PR**: #1587, #1623, #1920, #1970, #1996, #2066（按 upstream 时间序）
- **kimi review**: approve（无 critical，3 个 suggestion 均为 pre-existing）
- **build 验证**: ✅ go build + pnpm build 全过
- **冲突**: 无（6 个 cherry-pick 全部 clean apply）
- **Schema/Generated Code 改动**: 无
- **merge SHA**: a53527fa

### slice-2-openai-core
- **状态**: ✅ done（merged at 2ce67ca4）
- **分支**: `sync/2026-04/slice-2-openai-core`
- **PR**: [ddnio/sub2api#11](https://github.com/ddnio/sub2api/pull/11)
- **计划包含 PR**: #1668, #1687, #1683, #1772, #1943, #1960, #1948
- **实际 cherry-pick**: #1668, #1687, #1683, #1772, #1948（5 个）
- **跳过**: #1943 / #1960 — 依赖未拉的上游 refactor（`handleStreamingResponsePassthrough` 增加 `originalModel/mappedModel` 参数），手术整合风险大，转 HOLD
- **额外 fix**: 2 处 test 签名适配（kimi review 反馈）
- **kimi review**: 3 轮 → approve（首轮 critical 修复后通过）
- **build 验证**: ✅ go build + pnpm build + 相关测试全过
- **Schema/Generated Code 改动**: 无
- **merge SHA**: 2ce67ca4

### slice-3-codex
- **状态**: ✅ done（merged at 60f10e5b）
- **分支**: `sync/2026-04/slice-3-codex`
- **PR**: [ddnio/sub2api#12](https://github.com/ddnio/sub2api/pull/12)
- **计划包含 PR**: #1575, #1690, #1766, #1910, #1895, #1911
- **实际 cherry-pick**: #1575, #1690, #1910, #1911（4 个）
- **跳过**: #1766 (codex-drop-removed-models)、#1895 (codex-spark-limitations) — useModelWhitelist.ts / openai_codex_transform.go 多处冲突，转 HOLD
- **额外**: 临时本地 `firstNonEmptyString` helper（合并后 slice-4 失败，helper 保留为永久）
- **kimi review**: approve
- **build 验证**: ✅ go build + pnpm build + go vet 全过
- **Schema/Generated Code 改动**: 无
- **merge SHA**: 60f10e5b

### slice-4-openai-image
- **状态**: ❌ 整片转 HOLD（5 个 PR 全部 cherry-pick 冲突）
- **分支**: 不创建 PR
- **计划包含 PR**: #1795, #1813, #1853, #2006, #2044
- **实际**: 全部失败
- **原因**：
  - #1795（openai-image-api-sync, +2805 行）冲突 `pkg/openai/constants.go` + `openai_account_scheduler.go`
  - #1813、#1853、#2006、#2044 都依赖 #1795 引入的 `openai_images.go` 和接口，#1795 没合则无意义
  - openai image API 在 fork 与 upstream 已严重分叉，需人工评估方向（保留 fork 自有实现 vs 整体切换上游）
- **副作用**: slice-3 的临时 `firstNonEmptyString` helper 因此保留（不再有上游 openai_images.go 替换）

### slice-5-cc-mimicry
- **状态**: ✅ done（merged at 4551da74）
- **分支**: `sync/2026-04/slice-5-cc-mimicry`
- **PR**: [ddnio/sub2api#13](https://github.com/ddnio/sub2api/pull/13)
- **包含 PR**: #1914（cc-mimicry-parity, 13 文件 +1119/-76）
- **kimi review**: approve（无 critical，3 条 nit suggestion）
- **build 验证**: ✅ go build + pnpm build 全过
- **冲突**: 无（clean apply）
- **Schema/Generated Code 改动**: 无
- **merge SHA**: 4551da74

### slice-6-ops-misc
- **状态**: ✅ done（merged at 11f5a6e3）
- **分支**: `sync/2026-04/slice-6-ops-misc`
- **PR**: [ddnio/sub2api#14](https://github.com/ddnio/sub2api/pull/14)
- **计划包含 PR**: #1702, #1749, #1752, #1836, #2090
- **实际 cherry-pick**: #1702, #1749, #2090（3 个）
- **跳过**: #1752 (i18n 冲突)、#1836 (usage_billing_repo 冲突)，转 HOLD
- **kimi review**: approve
- **build 验证**: ✅ go build + pnpm build 全过
- **Schema/Generated Code 改动**: 无
- **merge SHA**: 11f5a6e3

### slice-7-frontend
- **状态**: ✅ done（merged at a845041a）
- **分支**: `sync/2026-04/slice-7-frontend`
- **PR**: [ddnio/sub2api#15](https://github.com/ddnio/sub2api/pull/15)
- **计划包含 PR**: #1545, #1603
- **实际 cherry-pick**: #1603 datatable-mobile-double-render（1 个，2 文件 +177/-18）
- **跳过**: #1545 smooth-sidebar-collapse — 与 fork 重构后的 AppSidebar.vue 冲突，转 HOLD
- **kimi review**: approve
- **build 验证**: ✅ go build + pnpm build 全过
- **Schema/Generated Code 改动**: 无
- **merge SHA**: a845041a

---

## 🏁 同步周期闭环总结（2026-04-30）

**周期**: 2026-04-29 启动 → 2026-04-30 完成（约 2 个工作日）

### 实际成果

| 指标 | 数值 |
|---|---|
| **Upstream 增量** | 539 commits / 61 first-parent PR |
| **成功合入 fork** | **23 个 PR**（slice-0 3 + slice-1 6 + slice-2 5 + slice-3 4 + slice-5 1 + slice-6 3 + slice-7 1） |
| **HOLD 留待人工** | 38 个 PR |
| **完成切片** | 7 个 / 计划 8 个（slice-4 整片 HOLD） |
| **fork PR 总数** | 7 个 sync slice PR + 1 个台账 PR = 8 个 |
| **kimi review 轮次** | 共 9 轮（其中 slice-2 跑 3 轮迭代修复） |
| **生产/测试库 schema** | 未触动（无 migration 改动） |

### 切片执行结果一览

| 切片 | 计划 | 实际 | 跳过原因 | merge SHA |
|---|---|---|---|---|
| slice-0-easy | 5 | 3 | sidebar 重构（×2） | d80a3827 |
| slice-1-anthropic | 6 | 6 | — | a53527fa |
| slice-2-openai-core | 7 | 5 | signature 依赖（×2） | 2ce67ca4 |
| slice-3-codex | 6 | 4 | useModelWhitelist / openai_codex_transform 冲突（×2） | 60f10e5b |
| slice-4-openai-image | 5 | 0 | image API 整片分叉 | — |
| slice-5-cc-mimicry | 1 | 1 | — | 4551da74 |
| slice-6-ops-misc | 5 | 3 | i18n / billing repo 冲突（×2） | 11f5a6e3 |
| slice-7-frontend | 2 | 1 | sidebar 重构 | a845041a |
| **合计** | **37** | **23** | **14 转 HOLD（含 slice-4 整片 5 个）** | |

### Phase 2 待人工对齐（38 项 HOLD）

每行 = 1 个 PR。分组按主题，每个 PR 仅出现在一个组。

**Payment 域（9 项）** — 待 Q-C 决方向
- #1572 feat/payment-system-v2（45k+ 行）
- #1610 alipay-wxpay-type-mapping
- #1655 payment-fee-multiplier
- #1731 rate-billing-autofill-response-limit
- #1764 wxpay-pubkey-hardening
- #1973 Nobody-Zhang/main (Zpay refund)
- #1576 feat/payment-docs
- #1612 qrcode-density
- #1734 payment-recommend-kyren-topup

**Auth 域（6 项）**
- #1538 fix/bug-cleanup-main
- #1785 rebuild/auth-identity-foundation（80k 行级别）
- #1799 rebuild/auth-identity-foundation（相关 PR）
- #1810 profile-auth-bindings-i18n
- #1802 profile-auth-bindings-i18n（相关 PR）
- #1829 codex-oauth-proxy-message

**Affiliate 域（1 项）**
- #1897 codex/invite-affiliate-rebate

**大重构（4 项）**
- #1850 feat/channel-insights（35k 行）
- #1815 feat_rpm
- #1828 wx-11/main
- #1637 websearch-notify-pricing

**OpenAI image API（5 项 - slice-4 整片）**
- #1795 openai-image-api-sync
- #1813 fix-openai-image-handling
- #1853 codex-image-generation-bridge
- #2006 openai-images-explicit-session
- #2044 openai-image-apikey-versioned-base-url

**Vertex / Fast-Flex（2 项）**
- #1977 sholiverlee/vertex
- #2051 openai-fast-flex-policy

**Sidebar 已重构（3 项）**
- #1545 smooth-sidebar-collapse
- #1624 fix-issue-1613-version-dropdown
- #1635 fix-issue-1613-version-dropdown（v2）

**Signature/Codex 重叠（4 项）**
- #1943 openai-responses-preoutput-failover（passthrough signature 依赖）
- #1960 openai-stream-keepalive-downstream-idle（passthrough signature 依赖）
- #1766 codex-drop-removed-models（useModelWhitelist 冲突）
- #1895 codex-spark-limitations（codex transform 冲突）

**其他（4 项）**
- #1940 bump-codex-cli-version（auth 域 deps bump）
- #1752 quota-exceeded-scheduling（i18n 冲突）
- #1836 account-daily-weekly-quota-cache-invalidation（billing repo 冲突）
- #1666 account-cost-display（migration）

**核对**: 9 + 6 + 1 + 4 + 5 + 2 + 3 + 4 + 4 = **38** ✅

**未进切片的 24 个原始 HOLD**: 全部包含在上面 9 个分组里。剩余 14 个为各 slice 执行时新转入的（slice-0 转 #1624/#1635；slice-2 转 #1943/#1960；slice-3 转 #1766/#1895；slice-4 整片转 5 个；slice-6 转 #1752/#1836；slice-7 转 #1545）。

### 已知遗留

1. **fork 临时 helper**: `backend/internal/service/openai_codex_string_helpers.go`（slice-3 引入）
   - 因 slice-4 整片 HOLD，没有上游 `openai_images.go` 来替换它
   - 决定 payment/openai-image 方向时一并清理

2. **pre-existing 测试编译错误**（main 分支已有，非本周期引入）
   - `auth_service_*_test.go` NewAuthService 签名缺 PromoService/ReferralService/DefaultSubscriptionAssigner
   - `admin_service_create_user_test.go` defaultSubscriptionAssignerStub/settingRepoStub 未定义
   - 需要单独 PR 修复

3. **生产撞表风险未触发**
   - 本周期未碰任何 migration，`payment_orders` 表撞库风险待 Phase 2 处理

### 部署状态

- **测试环境**: 未部署（本周期合入的 slice 都已在 fork main，但服务器 git pull 未触发；需后续单独评估部署窗口）
- **生产环境**: 未部署
- **DB schema**: 未变更

### 下次同步建议

- **频率**: 上游 v0.1.119 之后再积累至 ~50 commit 再启动下次 sync
- **优先 Phase 2**: 在下次同步前先解决 payment/auth/affiliate 方向，否则 HOLD 会越滚越大

### slice-2-openai-core
- **状态**: pending
- **分支**: `sync/2026-04/slice-2-openai-core`
- **包含 PR**: #1948, #1960, #1943, #1772, #1683, #1687, #1668
- **PR**:
- **kimi review**:
- **build 验证（go/pnpm/ent generate）**:
- **冲突**:
- **Schema/Generated Code 改动**:
- **merge SHA**:

### slice-3-codex
- **状态**: pending
- **分支**: `sync/2026-04/slice-3-codex`
- **包含 PR**: #1911, #1895, #1910, #1575, #1690, #1766
- **PR**:
- **kimi review**:
- **build 验证（go/pnpm/ent generate）**:
- **冲突**:
- **Schema/Generated Code 改动**:
- **merge SHA**:

### slice-4-openai-image
- **状态**: pending
- **分支**: `sync/2026-04/slice-4-openai-image`
- **包含 PR**: #2044, #2006, #1853, #1813, #1795
- **PR**:
- **kimi review**:
- **build 验证（go/pnpm/ent generate）**:
- **冲突**:
- **Schema/Generated Code 改动**:
- **merge SHA**:

### slice-5-cc-mimicry
- **状态**: pending
- **分支**: `sync/2026-04/slice-5-cc-mimicry`
- **包含 PR**: #1914
- **PR**:
- **kimi review**:
- **build 验证（go/pnpm/ent generate）**:
- **冲突**:
- **Schema/Generated Code 改动**:
- **merge SHA**:

### slice-6-ops-misc
- **状态**: pending
- **分支**: `sync/2026-04/slice-6-ops-misc`
- **包含 PR**: #2090, #1836, #1752, #1749, #1702
- **PR**:
- **kimi review**:
- **build 验证（go/pnpm/ent generate）**:
- **冲突**:
- **Schema/Generated Code 改动**:
- **merge SHA**:

### slice-7-frontend
- **状态**: pending
- **分支**: `sync/2026-04/slice-7-frontend`
- **包含 PR**: #1603, #1545
- **PR**:
- **kimi review**:
- **build 验证（go/pnpm/ent generate）**:
- **冲突**:
- **Schema/Generated Code 改动**:
- **merge SHA**:

---

## Schema & Generated Code 变更登记

> 涉及以下任何文件的切片必须在此记录：
> - `backend/migrations/*.sql`
> - `backend/ent/schema/*.go`
> - `backend/cmd/server/wire_gen.go`
> - `backend/internal/repository/*_provider.go`（provider 接口变更）
>
> 部署前必须 pg_dump 留底（参见 memory: feedback_pre_deploy_dump.md）。
> migration 字典序并存策略对**同表 schema 不同**的情况无效（参见 077 vs 092 撞表风险）。

**当前**: 计划内 0 个 migration / 0 个 ent schema 改动（payment/auth/affiliate 全 hold）。

### 生产/测试库 schema_migrations 现状（2026-04-29 实测）

测试库（`sub2api_test`）和生产库（`sub2api`）**完全一致**，应用到 092：

| 编号 | 文件 | 来源 |
|---|---|---|
| 077 | `077_add_payment_tables.sql` | fork（payment） |
| 077 | `077_add_usage_log_requested_model.sql` | upstream |
| 078 | `078_add_refund_no.sql` | fork（refund） |
| 078 | `078_add_usage_log_requested_model_index_notx.sql` | upstream |
| 091 | `091_add_referral_system.sql` | fork（referral） |
| 091 | `091_add_group_messages_dispatch_model_config.sql` | upstream |
| 092 | `092_referral_deferred_reward.sql` | fork |

**实战验证**：077/078/091 双编号字典序并存策略已成功（test+prod 都跑过）。

**已知冲突点**：
- `payment_orders` 表已由 fork `077_add_payment_tables.sql` 创建并存在
- 上游若引入 `092_payment_orders.sql`（在 HOLD #1572 内）会 `CREATE TABLE` 失败 → **永远不能并存执行**
- 处理：Phase 2 决定 payment 方向时一并解决（要么走 fork 自研永远跳过，要么重构合并）

### Schema & Generated Code 切片登记

| 切片 | 文件 | 类型 | 改动摘要 | 编译验证 | pg_dump | 部署状态 |
|---|---|---|---|---|---|---|
| — | — | — | — | — | — | — |

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
