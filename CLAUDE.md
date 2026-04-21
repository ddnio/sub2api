# CLAUDE.md

Claude Code 在本仓库工作时必须先读本文件。

## 项目说明

Sub2API 是一个 AI API 网关平台。后端 Go，前端 Vue3，数据库 PostgreSQL，缓存 Redis。
发布时前端静态文件通过 `go:embed` 打包进后端二进制，单文件部署。

## 工程规范

详细规范见 `AGENTS.md`，核心要点：

- `origin` = 团队 fork（ddnio/sub2api），`upstream` = 上游（Wei-Shaw/sub2api）
- 日常功能在 `feature/<topic>` 分支开发，不直接推 `main`
- 变更部署方式时，同步更新 `docs/engineering/deployment.md`
- **不要提交任何 config.yaml、密钥、生产密码**

## 本地开发

```bash
# 后端（需要 Go 1.26+）
cd backend && go run ./cmd/server

# 前端
pnpm --dir frontend dev
```

配置文件放 `backend/config.yaml`（已在 `.gitignore`），连接服务器测试库（Tailscale IP：`100.71.166.42`）。
参考 `docs/engineering/deployment.md` 了解完整配置。

## 部署

```bash
# 在服务器 /data/service/sub2api 下执行
bash deploy/deploy-server.sh test   # 部署测试环境
bash deploy/deploy-server.sh prod   # 部署生产环境
```

脚本会自动 git pull → docker build → 重启容器。
服务器配置文件在 `/etc/sub2api/test.yaml` 和 `prod.yaml`，不在代码库里。

## 部署补充

- 服务器当前跟踪分支不一定是 main，`git pull` 只拉当前分支
- 部署新分支前需在服务器先 `git checkout <branch>` 再运行脚本
- 测试域名：`https://router-test.nanafox.com`（→ 127.0.0.1:8081）
- 生产域名：`https://router.nanafox.com`（→ 127.0.0.1:8080）
- 迁移**全自动**：启动时自动检测并执行 delta，无需手动跑 SQL

## 支付模块

- 已合并 main（2026-03-23）
- 支付服务商：**微信支付 Native Pay v3**（官方直连）
- Provider 实现：`backend/internal/repository/wxpay_provider.go`
- 备用 Provider：`backend/internal/repository/easypay_provider.go`（易支付，可切换）
- 通过 `payment.provider` 配置项切换（`wxpay` 或 `easypay`）
- 微信支付公钥模式（2024年后新商户默认），需配置 `wxpay_public_key_id`（带 `PUB_KEY_ID_` 前缀）
- Migration：`backend/migrations/077_add_payment_tables.sql`
- **密钥不要提交 git**，只放服务器 `/etc/sub2api/` 配置文件里
- 本地密钥文件放 `backend/config/`（已在 `.gitignore`）

## Wire DI

项目使用 Wire 做依赖注入，但 **Go 1.26.1 与 `wire` 生成工具不兼容**，无法运行 `wire gen`。
`backend/cmd/server/wire_gen.go` 需要**手动维护**：新增或修改 Provider 后直接编辑 `wire_gen.go`，不要尝试跑 `wire` 命令。

`ent` 代码生成工具**可以正常运行**（与 wire 不同）：
```bash
cd backend && go generate ./ent/...
```
修改 `backend/ent/schema/` 后需运行此命令重新生成。

## 上游同步

- 上游仓库：`upstream` remote → `https://github.com/Wei-Shaw/sub2api.git`
- 上次同步：2026-03-29，完整 merge upstream/main（v0.1.105），含 TLS 指纹、Privacy 模式、requested_model、Responses↔Anthropic 双向转换等
- **migration 编号说明**：`schema_migrations` 以 filename 为主键，077/078 号存在同号不同内容（我们的 payment + 上游的 requested_model），字典序排列，可安全并存
- **ent 冲突策略**：merge 时接受 upstream 所有 ent 非 schema 生成文件，手工合并 `ent/schema/`，然后运行 `go generate ./ent/...` 重新生成全部

## 关键文件

| 文件 | 说明 |
|------|------|
| `deploy/deploy-server.sh` | 服务器部署脚本 |
| `deploy/docker-compose.yml` | 基础设施（postgres + redis） |
| `deploy/docker-compose.override.yml` | 服务器本地覆盖（端口暴露，不提交） |
| `docs/engineering/deployment.md` | 完整部署文档 |
| `docs/engineering/git-workflow.md` | Git 工作流 |
| `backend/config.yaml` | 本地配置（不提交） |
| `backend/cmd/server/wire_gen.go` | 手动维护的 Wire 注入文件 |
| `backend/internal/repository/wxpay_provider.go` | 微信支付 Native Pay v3 Provider |
| `backend/internal/service/payment_service.go` | 支付业务逻辑（创单、回调、发放权益） |
| `backend/migrations/077_add_payment_tables.sql` | 支付模块 DB migration |

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.

<!-- code-review-graph MCP tools -->
## MCP Tools: code-review-graph

**IMPORTANT: This project has a knowledge graph. ALWAYS use the
code-review-graph MCP tools BEFORE using Grep/Glob/Read to explore
the codebase.** The graph is faster, cheaper (fewer tokens), and gives
you structural context (callers, dependents, test coverage) that file
scanning cannot.

### When to use graph tools FIRST

- **Exploring code**: `semantic_search_nodes` or `query_graph` instead of Grep
- **Understanding impact**: `get_impact_radius` instead of manually tracing imports
- **Code review**: `detect_changes` + `get_review_context` instead of reading entire files
- **Finding relationships**: `query_graph` with callers_of/callees_of/imports_of/tests_for
- **Architecture questions**: `get_architecture_overview` + `list_communities`

Fall back to Grep/Glob/Read **only** when the graph doesn't cover what you need.

### Key Tools

| Tool | Use when |
| ------ | ---------- |
| `detect_changes` | Reviewing code changes — gives risk-scored analysis |
| `get_review_context` | Need source snippets for review — token-efficient |
| `get_impact_radius` | Understanding blast radius of a change |
| `get_affected_flows` | Finding which execution paths are impacted |
| `query_graph` | Tracing callers, callees, imports, tests, dependencies |
| `semantic_search_nodes` | Finding functions/classes by name or keyword |
| `get_architecture_overview` | Understanding high-level codebase structure |
| `refactor_tool` | Planning renames, finding dead code |

### Workflow

1. The graph auto-updates on file changes (via hooks).
2. Use `detect_changes` for code review.
3. Use `get_affected_flows` to understand impact.
4. Use `query_graph` pattern="tests_for" to check coverage.
