# Git 工作流

本仓库建立在上游项目之上，当前本地仓库应保持如下远端结构：

- `origin` = 我们自己的 fork
- `upstream` = 原始上游仓库

编写本文档时，远端为：

- `origin`: `https://github.com/ddnio/sub2api.git`
- `upstream`: `https://github.com/Wei-Shaw/sub2api.git`

## 目标

- 把我们的定制功能沉淀在自己的 fork 上。
- 保持按需同步上游更新的能力。
- 避免共享分支上的高风险历史改写。
- 让发布、热修和后续维护更可控。

## 分支策略

### 稳定分支

- `main`
  - 生产稳定分支
  - 应始终保持可发布状态
  - 在未引入发布分支模型前，默认作为上线基线

- `develop`（可选）
  - 只有当并行开发增多、确实需要共享集成分支时再引入
  - 团队还小、节奏还简单时，不要为了“规范”而额外加复杂度

### 工作分支

- `feature/<topic>`
  - 新功能、产品迭代、内部增强

- `fix/<topic>`
  - 非紧急缺陷修复

- `hotfix/<topic>`
  - 生产紧急修复

- `chore/<topic>`
  - 文档、工具、CI、维护性调整

## 日常开发流程

### 1. 从 `main` 开始

```bash
git checkout main
git pull --ff-only origin main
```

### 2. 创建工作分支（默认走 worktree）

新功能、修复、文档调整一律用 git worktree 隔离。详见下文「Worktree 工作流」。

如果确实只想原地切分支（极小改动、不会跑 dev、几分钟搞完），也可以：

```bash
git checkout -b feature/<topic>
```

示例分支命名：

- `feature/payment-webhook`
- `feature/admin-usage-dashboard`
- `fix/rate-limit-window`
- `chore/update-docs`

### 3. 本地开发并验证

推送前至少完成这些动作：

- 运行与本次改动直接相关的最小验证
- 必要时运行格式化、lint、测试
- 自己先看一遍 diff

### 4. 推送到自己的 fork

```bash
git push -u origin feature/<topic>
```

不要把功能分支推到 `upstream`。

### 5. 在 fork 内开 PR

```bash
gh pr create --repo ddnio/sub2api --base main --head ddnio:<branch> ...
```

注意：`gh pr create` 默认目标是 fork 的 parent（即 upstream Wei-Shaw/sub2api）。**必须显式指定 `--repo ddnio/sub2api`**，否则 PR 会错开到上游。

### 6. 有意识地合回主线

如果当前只使用 `main`：

- 经 review 后合回 `main`
- 保持 `main` 始终可发布

如果未来引入 `develop`：

- 功能先合到 `develop`
- 测试完成后再从 `develop` 合到 `main`

## Worktree 工作流

### 为什么默认用 worktree

- 多分支并行不需要 `git stash` 来回切
- Claude Code 多个 session 可以同时挂在不同 worktree 上互不干扰
- 切分支不影响当前正在跑的 dev / build

### 目录约定

Claude Code 用原生 `EnterWorktree` 工具，默认建在仓库内 `.claude/worktrees/<name>/`。该路径已被根 `.gitignore` 的 `.claude` 规则覆盖，不会污染 git 状态。

手动场景也用同一路径：

```bash
git worktree add .claude/worktrees/<feature> -b feature/<topic>
```

### 新建 worktree 后必做的 bootstrap

每个 worktree 都是干净的工作区，不带本地不入库的文件。新建后立即：

```bash
cd .claude/worktrees/<feature>

# 1. 拷配置（gitignored，主仓库才有）
cp ../../backend/config.yaml backend/config.yaml

# 2. 装前端依赖（pnpm 用 content-addressable store，磁盘开销小，但符号链接结构必须重建）
pnpm --dir frontend install --frozen-lockfile
```

Go build cache 默认在 `~/Library/Caches/go-build`，**全局共享**，不需要每个 worktree 单独管。

### 并行跑 dev 时的端口冲突

后端默认 `8080`、前端 vite 默认 `5173`。多 worktree 同时跑 dev 必须错开端口：

```bash
# 后端
SERVER_PORT=8090 go run ./cmd/server

# 前端
pnpm --dir frontend dev -- --port 5183
```

建议为每个长期 worktree 在 README 顶部记录它分配的端口。

### 跨 worktree 风险点

| 文件 | 风险 | 防御 |
|------|------|------|
| `backend/cmd/server/wire_gen.go` | 手动维护，多个 worktree 同时改 Provider 难合 | 改前 `git worktree list` 看其他 worktree 是否动过 |
| `backend/ent/schema/` + 生成产物 | merge 后要重跑 `go generate ./ent/...` | 同上，且生成产物冲突要全量 regenerate |
| `backend/migrations/` | 同号不同内容的 migration 冲突 | 取下一个未占号；上游同步时按字典序并存 |

### 退出与清理

任务完成、PR 合并后立即清理，避免长尾 worktree 越攒越多：

```bash
# Claude Code 内
ExitWorktree(action="remove")

# 命令行
git worktree remove .claude/worktrees/<feature>

# 定期清理元数据残留（worktree 目录已删但 .git 里还记录的）
git worktree prune
git worktree list   # 确认现状
```

如果只是临时离开 worktree、之后还要回来：`ExitWorktree(action="keep")`，目录和分支都保留。

### 长期改进（可选）

- 整理一份 `backend/config.example.yaml`（脱敏 + 注释），新 worktree bootstrap 改成 `cp config.example.yaml backend/config.yaml` 再补本地密钥，比拷主仓库更可持续

## 同步上游更新

当上游仓库有我们需要吸收的修复或功能时，按这个流程操作。

### 1. 拉取上游最新代码

```bash
git fetch upstream
```

### 2. 更新本地 `main`

```bash
git checkout main
git merge upstream/main
```

除非团队明确决定切换到 rebase 工作流，否则默认使用 merge，原因是更直观、风险更低。

### 3. 推送到自己的 fork

```bash
git push origin main
```

### 4. 让工作分支跟上新的 `main`

小型、个人分支可使用 rebase：

```bash
git checkout feature/<topic>
git rebase main
```

多人共享分支或风险较高时，优先 merge：

```bash
git checkout feature/<topic>
git merge main
```

## 热修流程

线上问题需要尽快修复时，按下面流程执行。

### 1. 从 `main` 切分支

```bash
git checkout main
git pull --ff-only origin main
git checkout -b hotfix/<topic>
```

### 2. 只做最小、安全的修复

- 不要顺手夹带重构
- 只验证受影响路径和明显回归点

### 3. 快速推送并合并

```bash
git push -u origin hotfix/<topic>
```

### 4. 确保其他长期分支也带上修复

如果后面引入了 `develop`，线上热修合入 `main` 后，要立即同步到 `develop`。

## 远端参考

查看当前远端：

```bash
git remote -v
```

预期应为：

```text
origin   https://github.com/ddnio/sub2api.git (fetch)
origin   https://github.com/ddnio/sub2api.git (push)
upstream https://github.com/Wei-Shaw/sub2api.git (fetch)
upstream https://github.com/Wei-Shaw/sub2api.git (push)
```

## 保护规则

- 不要 force-push `main`。
- 不要提交生产凭据和敏感信息。
- 不要把同步上游和产品开发混在同一个分支里做。
- 不要随意修改生成文件，先搞清楚它们的生成方式。
- 工作流有变化时，文档要在同一任务里一起更新。

## 当前建议

现阶段保持流程简单即可：

- `main` 作为稳定分支
- 所有新需求都走 `feature/*`
- 需要时再把 `upstream/main` 合入 `main`
- 一切开发推送都走 `origin`

只有当并行开发和发版协调真的开始带来痛点时，再引入 `develop`。
