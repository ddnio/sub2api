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

### 2. 创建工作分支

```bash
git checkout -b feature/<topic>
```

示例：

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

### 5. 有意识地合回主线

如果当前只使用 `main`：

- 经 review 后合回 `main`
- 保持 `main` 始终可发布

如果未来引入 `develop`：

- 功能先合到 `develop`
- 测试完成后再从 `develop` 合到 `main`

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
