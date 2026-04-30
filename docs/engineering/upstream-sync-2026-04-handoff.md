# 上游同步 2026-04 交接文档

> 本文档用于跨 session 接续工作。读完此文即可无缝衔接。

**生成时间**：2026-04-30 09:50 (UTC+8)
**当前阶段**：Phase 1 完成，已部署生产 → 进入 Phase 2 准备期

---

## 1. 当前状态（必读）

### 代码与分支

| 位置 | 分支 | Commit | 说明 |
|---|---|---|---|
| origin/main | main | `8d7fb33c` | 生产分支，已含 23 个上游 PR |
| origin/staging/2026-04 | staging/2026-04 | `8d7fb33c` | 与 main 同 commit，作为"已验证"标签 |
| 本地 main（主 checkout）| main | `dc151a27` | **落后 9 commits**，需要 `git pull` |
| worktree | sync/2026-04/closure | `c8c74e3d` | 本周期工作 worktree，已无任务 |

**关键事实**：main 已经包含全部 23 个 sync PR 的代码。`origin/main` == `origin/staging/2026-04`。

### 部署状态

| 环境 | 容器 | 二进制 md5 | 状态 |
|---|---|---|---|
| 测试 (router-test.nanafox.com:8081) | `sub2api-test` | `c25070c1309eed549694ac3103250145` | ✅ Up healthy |
| 生产 (router.nanafox.com:8080) | `sub2api-prod` | `c25070c1309eed549694ac3103250145` | ✅ Up healthy（部署于 09:44）|

测试和生产**用同一个二进制**（Docker layer 共享）。

### DB 备份

| 环境 | 备份路径 | 大小 | 时间 |
|---|---|---|---|
| 测试库 | `/data/backups/sub2api_test_pre_sync_2026-04-30-0837.sql` | 42M | 部署前 |
| 生产库 | `/data/backups/sub2api_prod_pre_sync_2026-04-30-0942.sql` | 872M | 部署前 |

---

## 2. 已完成工作（Phase 1）

### 9 个本仓库 PR（全部已 merge 到 main）

```
#8  docs(upstream-sync): 2026-04 增量台账
#9  sync(slice-0-easy): #1543/#1663/#1753 i18n + test 小修复
#10 sync(slice-1-anthropic): 6 PR Anthropic SSE/cache/credit
#11 sync(slice-2-openai-core): 5 PR OpenAI stream/account/quota
#12 sync(slice-3-codex): 4 PR Codex 修复 + 临时 firstNonEmptyString helper
#13 sync(slice-5-cc-mimicry): #1914 cc-mimicry-parity（13 文件 +1119）
#14 sync(slice-6-ops-misc): 3 PR ops/quota/scheduling
#15 sync(slice-7-frontend): #1603 datatable-mobile-double-render
#16 docs(upstream-sync): 2026-04 周期闭环总结
```

合计：**23 个上游 PR 已 merge，38 个上游 PR 在 HOLD**。

### 关键变更

- **0 migration**、**0 wire_gen.go**、**0 ent/schema** 改动 — Phase 1 是纯应用层修复
- 76 文件，+5276/-543 行
- 含 1 个临时文件：`backend/internal/service/openai_codex_string_helpers.go`（注释中标记 slice-4 后可删，但 slice-4 全 HOLD，本文件**永久保留**）

### 验证证据

测试环境（admin user_id=1, balance 122.96 → 122.64）：
- ✅ `/v1/models` 返回正常
- ✅ 非流式 completion (gpt-5.2-pro)：usage_logs id=328
- ✅ 流式 SSE (gpt-5.2-pro)：usage_logs id=329
- ✅ 计费精确（小数级）
- ✅ 0 panic / 0 fatal
- ⚠️ 15 条 ERROR 全部来自 account_id=2 (claude-1-max) Anthropic OAuth/代理问题，**非本次 sync 引入**，failover 正常切到 account_id=9

生产环境（部署后日志）：
- ✅ `/v1/messages` 200，account_id=13 (kimi-1)，model mapping `claude-haiku-4-5-20251001 → kimi-for-coding` 工作正常
- ✅ `/api/v1/auth/me` 200, `/api/v1/subscriptions/active` 200
- ✅ 用户实时流量正在处理

---

## 3. Phase 2 待办（38 PR HOLD）

详见 `docs/engineering/upstream-sync-2026-04.md`。按域分组：

| 域 | 数量 | 关键决策 |
|---|---|---|
| Payment | 9 | **Q-C：保留 wxpay 自研 vs 采纳上游 payment v2？**（用户上一轮选"全部 HOLD 给最后"）|
| Auth | 6 | auth-identity-foundation 80k 行重构 |
| Affiliate | 1 | #1897 |
| 大重构 | 4 | channel-insights / feat_rpm / wx-11 / websearch-notify |
| OpenAI image | 5 | slice-4 整片冲突（fork 的 image API 已分叉）|
| Vertex/Fast-Flex | 2 | |
| Sidebar 重构 | 3 | |
| Signature/Codex 重叠 | 4 | |
| 其他 | 4 | |

**用户原话**："现在不决定，全部 hold 给最后"。

---

## 4. 已知遗留问题

### 待修（非本次 sync 引入，但部署前就存在）

1. **account_id=2 (claude-1-max) Proxy Auth 失败**
   - 日志：`Proxy Authentication Required` → 401 → 529
   - 代理：SG1 `202.170.76.182:12323`
   - 影响：单账号失效，但 failover 工作正常（自动切到 account_id=9）
   - 处理：用户决定，可以重置该账号的 OAuth token 或修代理配置

2. **测试库 main 分支预先存在的 test 编译错误**
   - 文件：`auth_service_*_test.go`、`admin_service_create_user_test.go`
   - 与本次 sync 无关，需要单独 fix PR

### 临时文件

- `backend/internal/service/openai_codex_string_helpers.go` — 14 行的 helper，永久保留（slice-4 全 HOLD 后无可去重 target）

---

## 5. 下一个 session 接续指南

### 立即可做（不需要决策）

```bash
# 1. 把本地 main 拉到最新（可能有用）
cd /Users/nio/project/nanafox/sub2api
git fetch origin
git checkout main
git pull   # dc151a27 → 8d7fb33c

# 2. 清理本周期 worktree（如果不再需要）
git worktree remove .claude/worktrees/upstream-sync-2026-04
git worktree prune

# 3. 30 分钟观察生产错误日志（建议）
sshpass -p 'nio2026.' ssh root@108.160.133.141 \
  "docker logs --since 30m sub2api-prod 2>&1 | grep -cE 'ERROR|FATAL|PANIC'"
```

### 需要用户决策才能开始的（Phase 2）

按优先级建议：

1. **Q-C 支付方向**（最关键）
   - 选项 A：保留 wxpay 自研 → Payment 9 个 HOLD 全部 drop
   - 选项 B：采纳上游 payment v2 → 需要规划数据迁移（fork `077_add_payment_tables.sql` ↔ 上游 `092_payment_orders.sql` 撞表）
   - 选项 C：双轨制 → 配置切换（最复杂）

2. **Auth-identity-foundation**（80k 行）
   - 选项 A：整体合入 → 需要专门一个 sync 周期
   - 选项 B：drop → 维持 fork 现有 auth

3. **预存 test 编译错误**
   - 单独开 PR fix，不阻塞 Phase 2

### 服务器连接信息

```
SSH: root@108.160.133.141
密码: nio2026.
代码路径: /data/service/sub2api
配置: /etc/sub2api/{test,prod}.yaml
DB 备份: /data/backups/
docker exec sub2api-postgres psql -U sub2api -d {sub2api,sub2api_test}
```

### admin API key（测试环境）

```
sk-d0d61e3cf8cd4ab544349be00805acc4001d22accc29bd21beb844719601c887  (#1 胡华翔的密钥1)
sk-1b6f3f60effd01ec650dc8b158cb5d8369b2f9caef06a6194f507153fd44abf3  (#3 胡华翔的claude)
sk-18da6c5e5d8858b1f3b74cad28ced5cd6afa48ad26bad3e609ec0bbf2cca6ab8  (#7 gpt-kimi)
```

### 回滚步骤（紧急时使用）

```bash
# 生产回滚到 sync 前
sshpass -p 'nio2026.' ssh root@108.160.133.141 << 'EOF'
cd /data/service/sub2api
git checkout dc151a27
bash deploy/deploy-server.sh prod
EOF

# 如果 DB 也要恢复
docker exec -i sub2api-postgres psql -U sub2api -d sub2api < /data/backups/sub2api_prod_pre_sync_2026-04-30-0942.sql
```

---

## 6. 关键文件索引

| 文件 | 用途 |
|---|---|
| `docs/engineering/upstream-sync-2026-04.md` | 周期主台账（PR 详细分组、风险、验证流水线）|
| `docs/engineering/upstream-sync.md` | 上游同步历史索引 |
| `docs/engineering/upstream-sync-2026-04-handoff.md` | **本文档（交接）** |
| `CLAUDE.md` | 项目工程规范（worktree、wire、ent、上游同步策略）|

---

## 7. 一句话总结

**Phase 1 完成：23/61 上游 PR 已合并 + 测试/生产部署成功，0 panic 0 fatal，用户流量正常处理。Phase 2 等待用户决策（payment 方向、auth 方向）后启动。**
