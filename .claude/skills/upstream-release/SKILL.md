---
name: upstream-release
description: >-
  Standard 2-gate workflow for syncing fixed upstream Wei-Shaw/sub2api tags into the NanaFox fork,
  verifying fork invariants, running local rollback rehearsals, coordinating external code review,
  and deploying to production tmhhost. Use when the user asks to merge upstream sub2api versions,
  sync upstream releases, or deploy a new Router release.
---

# Sub2API 上游合并与发布标准流程（2-Gate）

本 Skill 将 `sub2api` 的上游合并与生产发布收拢为 **2 个明确审批门禁**，机械步骤全部交给 `deploy/release/` 下的确定性脚本，Agent 专注在**冲突决策、定制核对、审查判断与异常处理**。

## 原则与硬边界

- **固定 Tag**：只合并上游固定 release tag（如 `v0.2.24`），绝不合并漂移的 `upstream/main`。
- **Worktree 隔离**：所有合并、计划、构建在 `.claude/worktrees/release-vX.Y.Z` 完成，不原地切主目录分支。
- **不碰凭据**：严禁提交 `config.yaml`、密钥或打印生产环境变量中的密码。
- **只读无需额外打断**：本地预检、本地回滚演练、以及 `verify-prod.sh preflight`（全只读）都在 **Gate 1** 内自动完成，不再中途零碎请示。

---

## Gate 1：合并、验证、演练与外部审查（零生产写入）

收到用户"合并上游到 `vX.Y.Z`"后，**一口气执行完以下步骤**，最后只交付一份完整的 *Release Readiness* 报告等待审批：

### 1. 本机环境预检（先查资源再干活）
```bash
bash deploy/release/preflight-local.sh
```
- 磁盘低于 15 GiB、缺工具或 Docker/Colima 未启动时，立即停下报告，不要编到一半爆盘。
- 如使用 Colima，按输出提示导出 `DOCKER_HOST` 与 `TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE`。

### 2. 无损试合（Dry-Run）与规划
```bash
git fetch origin main && git fetch upstream --tags
git merge-tree --write-tree origin/main vX.Y.Z
```
- 检查冲突文件列表、上游 release notes、新增 SQL migration 数量，在 `docs/plans/YYYY-MM-DD-upstream-vX.Y.Z-merge.md` 写明简明计划。

### 3. 隔离 Worktree 合并与定制保护
```bash
git worktree add -b release/vX.Y.Z .claude/worktrees/release-vX.Y.Z origin/main
cd .claude/worktrees/release-vX.Y.Z
git merge --no-commit --no-ff vX.Y.Z
```
- **冲突解决**：保留 NanaFox 核心定制（Studio 登录/HMAC、微信支付/退款、邀请码、Opus 4.7 采样参数规则、OpenAI 错误快照顺序、图像创作插件），同时合入上游修复。
- **Migration 重编号**：上游新增 SQL 文件统一改编号接在当前最大编号之后（SQL 内容不动）；同步更新代码中所有硬编码的 `.sql` 文件名引用（如 `migrations_runner.go`）及测试。
- **Ent & Wire**：`ent/schema/` 手工合并后运行 `cd backend && go generate ./ent/...`；`wire_gen.go` 严禁跑 `wire`，只做手工核对。
- **运行确定性不变量检查**：
  ```bash
  bash deploy/release/check-merge-invariants.sh --base origin/main --upstream vX.Y.Z
  ```
  对输出中标出的高风险双改文件逐一 diff 核对（`git diff vX.Y.Z -- <file>`，确认 diff 里只有我们的定制改动）。

### 4. 编译与全量验证
- 后端：`go build`、`go vet`、单元测试、集成测试（带 `DOCKER_HOST`）、针对本次改动点的 `-race` 检查、`golangci-lint`、`govulncheck`。
- 数据库升级测试：冻结上一版本的升级测试集合，新增 `vPrev -> vX.Y.Z` 升级测试。
- 前端：`pnpm --dir frontend install --frozen-lockfile`、lint/typecheck、全量 Vitest、以及前端打包进后端的 `-tags embed` 构建验证。

### 5. 本地四轮回滚演练（零生产影响）
```bash
bash deploy/release/rehearse-rollback.sh --old-ref origin/main
```
自动完成 `旧版建库+种数据 -> 新版升级 -> 旧版重开升级后的库 -> 新版再启动`，并校验健康检查、401 鉴权、后台合规确认 + 4 个管理接口 200、迁移数、测试余额 `123.45` 不变、0 fatal 日志。

### 6. 生产只读预检（真实数据，禁止手抄）
```bash
bash deploy/release/verify-prod.sh preflight --host tmhhost
```
把真实输出的磁盘、内存预算（`GATEWAY_REQUEST_BODY_MEMORY_BUDGET_BYTES`）、`usage_logs` 行数/体积、当前迁移数与财务基线直接写入 `docs/plans/YYYY-MM-DD-upstream-vX.Y.Z-execution.md`。

### 7. 独立外部审查（给完整工作区上下文，不喂残片）
- 推送 `release/vX.Y.Z` 后，交给独立审查 Agent（如 Codex / agy）**直接在该 worktree 内读取完整文件**进行只读审查，限定审查范围为：
  1. `git diff vX.Y.Z..HEAD`（确认 NanaFox 定制完整）
  2. 冲突解决 commit 与重编号的 migration 文件（本次新增部分，不要混入历史重命名）
- 逐条核实发现：属于合并引入的直接修复并补单测；属于上游原版且无生产影响的记录在案。

### 8. 交付 Gate 1 报告并停止
向用户输出简洁的 **Release Readiness 报告**（冲突摘要、测试与演练结果、生产只读风险评估、审查结论），**停在这里等待用户一次性授权 Gate 2**（不碰生产、不合 `main`）。

---

## Gate 2：生产发布与收尾（用户一次确认后执行到底）

获得用户明确部署授权后，按顺序完成：

1. **发布前恢复点**：确认 NAS 最新完整备份正常，并在服务器私有目录 `/home/nio/releases/router-vX.Y.Z-YYYYMMDD/`（权限 `700`）留存 `pg_dump -Fc`、原 `compose.json` 与数据基线。
2. **构建与切换**：从干净的 `release/vX.Y.Z` 提交构建前端与 `linux/amd64` 二进制（`CGO=0`、`-tags embed`），核对 SHA256，更新镜像并只重建 Router（`--no-deps --force-recreate router`），不碰 Studio / PG / Redis / Caddy。
3. **切换后自动验收**：
   ```bash
   bash deploy/release/verify-prod.sh post-switch --host tmhhost
   ```
   确认迁移数、前后财务基线一致、0 重启、0 OOM、0 fatal 日志、公网域名 200/401 正常、前端 JS SHA256 与本机构建逐字节一致，并抽查自然流量计费与微信支付路由。
4. **代码与文档收尾**：
   - 将 `release/vX.Y.Z` 合并进 `main` 并推送；
   - 写发布记录 `docs/ops/YYYY-MM-DD-router-vX.Y.Z-deployment.md` 并更新 `docs/engineering/deployment.md`；
   - 同步更新 `CLAUDE.md` 中「上次同步」的版本与日期；
   - 在 `nanafox-ops` 提交并合并 `tmhhost/router/compose.json` 镜像 ID 变更 PR。
5. **环境清理**：同步本地 `main`，删除本次临时 worktree 与临时分支（保留远端 `release/vX.Y.Z` 分支、服务器旧镜像与恢复点目录），输出最终交付总结。
