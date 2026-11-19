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
- **只读授权约定**：调用本 Skill 即视为授权在 Gate 1 对 `tmhhost` 执行**严格只读**预检（`verify-prod.sh preflight`）；如用户未触发本 Skill，SSH 读生产前仍应先确认。任何生产写入动作一律等到 Gate 2。

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
- **Migration 重编号**：上游新增 SQL 文件统一改编号接在当前最大编号之后（SQL 内容不动）；同步更新代码中所有硬编码的 `.sql` 文件名引用（如 `migrations_runner.go`）及测试；**严禁修改或删除历史 migration 文件**。
- **Ent & Wire**：`ent/schema/` 手工合并后运行 `cd backend && go generate ./ent/...`；`wire_gen.go` 严禁跑 `wire`，只做手工核对。
- **运行确定性不变量检查**：
  ```bash
  bash deploy/release/check-merge-invariants.sh --base origin/main --upstream vX.Y.Z
  ```
  对输出中标出的高风险双改文件逐一 diff 核对（`git diff vX.Y.Z -- <file>`，确认 diff 里只有我们的定制改动）。

### 4. 编译与全量验证
- 后端：`go build`、`go vet`、单元测试、集成测试（带 `DOCKER_HOST`）、针对本次改动点的 `-race` 检查、`golangci-lint`、`govulncheck`。
- 数据库升级测试：冻结上一版本的升级测试集合，新增 `vPrev -> vX.Y.Z` 升级测试。
- 前端：`pnpm --dir frontend install --frozen-lockfile`、lint/typecheck、全量 Vitest、以及前端打包进后端的 `-tags embed` 构建验证（Vite 输出到 `backend/internal/web/dist`）。

### 5. 生产只读预检（先拿到生产实际运行版本与基线）
```bash
bash deploy/release/verify-prod.sh preflight --host tmhhost
```
- 从输出中的 `revision_label` 取得**生产当前实际运行的 commit SHA**（不要假设 `origin/main` 等于生产，合并 `main` 不会自动发布）。
- 脚本会自动保存基线 JSON 到 `/tmp/sub2api-prod-baseline.json`，把真实输出的磁盘、内存预算（`GATEWAY_REQUEST_BODY_MEMORY_BUDGET_BYTES`）、`usage_logs` 体量、当前迁移数与财务基线直接写入 `docs/plans/YYYY-MM-DD-upstream-vX.Y.Z-execution.md`（禁止手抄）。

### 6. 本地四轮回滚演练（零生产影响）
用上一步拿到的生产实际 commit SHA 作为 `--old-ref`：
```bash
bash deploy/release/rehearse-rollback.sh --old-ref <prod-revision-sha>
```
自动完成 `旧版 AUTO_SETUP 建库+种数据 -> 新版升级 -> 旧版重开升级后的库 -> 新版再启动`，并校验健康检查、401 鉴权、后台合规确认 + 4 个管理接口 200、迁移数、测试余额 `123.45` 不变、0 fatal 日志。

### 7. 独立外部审查（给完整工作区上下文，不喂残片）
- 推送 `release/vX.Y.Z` 后，交给独立审查 Agent（如 Codex / agy / Claude）**直接在该 worktree 内读取完整文件**进行只读审查，限定审查范围为：
  1. `git diff vX.Y.Z..HEAD`（确认 NanaFox 定制完整）
  2. 冲突解决 commit 与本次新增的重编号 migration 文件（不要混入历史重命名）
- 逐条核实发现：属于合并引入的直接修复并补单测；属于上游原版且无生产影响的记录在案。

### 8. 交付 Gate 1 报告并停止
向用户输出简洁的 **Release Readiness 报告**（冲突摘要、测试与演练结果、生产只读风险评估、审查结论），**停在这里等待用户一次性授权 Gate 2**（不碰生产、不合 `main`）。

---

## Gate 2：生产发布、验收与收尾（用户一次确认后执行到底）

获得用户明确部署授权后，按顺序完成：

1. **发布前恢复点**：确认 NAS 最新完整备份正常，并在服务器私有目录 `/home/nio/releases/router-vX.Y.Z-YYYYMMDD/`（权限 `700`）留存 `pg_dump -Fc`、原 `compose.json` 与数据基线。
2. **构建镜像与切换**：
   - 从干净的 `release/vX.Y.Z` 提交构建前端（输出到 `backend/internal/web/dist`）与 `linux/amd64` 二进制（`CGO=0`、`-tags embed`），记录二进制 SHA256；
   - 按现有方式制作生产镜像（如上游 Dockerfile 运行层无变化，可以当前生产镜像为底，仅替换 `/app/sub2api` 与 `/app/resources` 并更新 `version`/`revision` 标签，上传构建上下文校验 SHA256 后在主机 `docker build --pull=false`；若运行层有变则按完整 Dockerfile 构建）；
   - 验证镜像内 `/app/sub2api` SHA256 与本机一致且 `--version` 正确；
   - 更新 `/srv/nanafox-router-formal/compose.json` 的 Router 镜像 ID（diff 仅此一行），执行 `--no-deps --force-recreate router`（不碰 Studio / PG / Redis / Caddy）。
3. **切换后自动验收**：
   ```bash
   bash deploy/release/verify-prod.sh post-switch --host tmhhost
   ```
   自动校验 `check-router-production.py`、容器 `running=true` / `restarts=0` / `oom_killed=false`、近 15 分钟 0 fatal 日志、前后 DB 基线无回退、公网域名 200/401、以及线上首页入口 JS 文件名与 SHA256 与本机 `backend/internal/web/dist` 逐字节一致；随后抽查自然流量计费与微信支付路由。
4. **回滚预案（任一异常立即执行）**：
   - **触发条件**：`/health` 超时未就绪、日志出现 `panic`/`fatal`/`checksum mismatch`/迁移失败、容器重启或 OOM、财务/核心数据基线异常。
   - **回滚操作**：将 `/srv/nanafox-router-formal/compose.json` 里的 image 改回旧镜像 ID（旧镜像保留在宿主机上），再次执行同一条 `--no-deps --force-recreate router`。
   - **关键红线**：**回退后严禁还原旧数据库备份**（已通过本地回滚演练证实旧版可直接运行在升级后的库上，还原旧库会覆盖切换期间的新写入数据）。
5. **代码与文档收尾**：
   - 将 `release/vX.Y.Z` 合并进 `main` 并推送；
   - 写发布记录 `docs/ops/YYYY-MM-DD-router-vX.Y.Z-deployment.md` 并更新 `docs/engineering/deployment.md`；
   - 同步更新 `CLAUDE.md` 中「上次同步」的版本与日期；
   - 在 `nanafox-ops` 提交并合并 `tmhhost/router/compose.json` 镜像 ID 变更 PR。
6. **环境清理**：同步本地 `main`，删除本次临时 worktree 与临时分支（保留远端 `release/vX.Y.Z` 分支、服务器旧镜像与恢复点目录），输出最终交付总结。
