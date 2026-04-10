# Upstream 同步指南

upstream 仓库：`Wei-Shaw/sub2api`（`upstream` remote）
我们的 fork：`ddnio/sub2api`（`origin` remote）

---

## 什么时候同步

**不建议频繁小批量同步**，推荐积累到以下情况再合并：

- 落后 upstream 超过 50 commits
- upstream 有明确的目标 feature 需要引入（如 channel 定价）
- upstream 有重要 bugfix 影响线上（可 cherry-pick 单个 commit）

---

## 检查落后情况

```bash
git fetch upstream
git log upstream/main --oneline | wc -l   # upstream 总提交数
git log main..upstream/main --oneline      # 我们落后的提交列表
git log main..upstream/main --oneline | wc -l  # 落后数量
```

---

## 全量同步流程

### 1. 新建同步分支

```bash
git checkout main
git checkout -b sync/upstream-YYYYMMDD
```

### 2. 执行 merge（不要 rebase）

```bash
git merge upstream/main --no-commit --no-ff
```

`--no-commit` 让你在提交前先处理冲突。

### 3. 查看冲突文件

```bash
git diff --name-only --diff-filter=U
```

### 4. 解决冲突（核心原则）

| 区域 | 策略 |
|------|------|
| `ent/schema/*.go` | **手动合并**：保留我们的 payment/referral schema，接受 upstream 的新字段 |
| `ent/` 生成文件 | `git checkout upstream/main -- <files>` 清除冲突，然后 `go generate ./ent/...` 重建 |
| `wire_gen.go` | **手动合并**：同时保留 payment/channel provider 注入 |
| `service/wire.go` `handler/wire.go` | **手动合并**：保留我们的 payment/referral + 接受 upstream 新增 provider |
| `handler/handler.go` | **手动合并**：AdminHandlers 和 Handlers struct 两边字段都要 |
| `service/settings_view.go` `setting_service.go` | 两边字段都要（ReferralEnabled + OIDCConnect 等） |
| `i18n/en.ts` `i18n/zh.ts` | 删除 Sora section，保留我们的 payment/referral section |
| `repository/api_key_repo.go` | 保留 ReferralCode，删除 SoraStorage 字段 |

### 5. ent 重新生成（必须步骤）

本地 Go 版本可能不满足要求，用 Docker：

```bash
docker run --rm -v "$PWD/backend":/app -w /app golang:1.26 go generate ./ent/...
```

**注意**：运行前确保所有 ent 冲突文件已处理，否则编译失败。
步骤：先 `git checkout upstream/main` 所有 ent 生成文件 → 再运行 generate。

### 6. 编译验证

```bash
docker run --rm -v "$PWD/backend":/app -w /app golang:1.26 go build ./...
```

### 7. 测试验证

```bash
docker run --rm -v "$PWD/backend":/app -w /app golang:1.26 go test ./...
```

### 8. 提交 merge

```bash
git add .
git commit  # 填写 merge commit message
```

### 9. 部署测试环境验证

```bash
git push origin sync/upstream-YYYYMMDD
# 服务器切换分支
ssh nio@108.160.133.141 "cd /data/service/sub2api && git fetch origin && git checkout sync/upstream-YYYYMMDD && bash deploy/deploy-server.sh test"
```

### 10. 合并到 main

验证通过后：

```bash
git checkout main
git merge sync/upstream-YYYYMMDD --no-ff
git push origin main
# 同步部署 prod 和 fx
```

---

## 我们自定义的核心模块（每次合并必须保留）

| 模块 | 文件 | 说明 |
|------|------|------|
| 微信支付 | `repository/wxpay_provider.go`, `easypay_provider.go` | 支付 Provider |
| 支付业务 | `service/payment_service.go` | 创单、回调、权益发放 |
| 支付 DI | `service/wire.go` 中 ProvidePaymentService/ProvidePaymentExpiryService | Wire 注入 |
| 邀请系统 | `service/referral_service.go`，ent schema referral_code | 用户邀请码 |
| Migrations | `077_add_payment_tables.sql`, `078_add_refund_no.sql`, `091/092` | 数据库表 |

---

## 自动通知：GitHub Actions 周检查

在 `.github/workflows/upstream-check.yml` 添加：

```yaml
name: Check upstream updates
on:
  schedule:
    - cron: '0 9 * * 1'  # 每周一 09:00
  workflow_dispatch:

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Add upstream remote
        run: |
          git remote add upstream https://github.com/Wei-Shaw/sub2api.git
          git fetch upstream

      - name: Check commits behind
        run: |
          BEHIND=$(git log HEAD..upstream/main --oneline | wc -l)
          echo "落后 upstream $BEHIND 个 commits"
          if [ "$BEHIND" -gt 0 ]; then
            echo "## Upstream 有更新" >> $GITHUB_STEP_SUMMARY
            echo "落后 **$BEHIND** 个 commits" >> $GITHUB_STEP_SUMMARY
            git log HEAD..upstream/main --oneline | head -20 >> $GITHUB_STEP_SUMMARY
          fi
```

这样每周一 GitHub Actions Summary 会显示落后了多少，不需要人手盯。

---

## 历史同步记录

| 日期 | 版本范围 | 提交数 | 主要内容 | 分支 |
|------|---------|-------|---------|------|
| 2026-03-29 | ~ v0.1.105 | — | TLS 指纹、Privacy 模式、Responses↔Anthropic 双向转换 | main |
| 2026-04-11 | v0.1.105 → v0.1.130 | 225 | channel 定价、OIDC 登录、删除 Sora、LoadFactor fix | feature/upstream-sync-channel-pricing |

---

## 常见问题

**Q: `go generate ./ent/...` 失败，提示 undefined PaymentOrderMutation**

A: 需要先清除所有 ent 生成文件中的冲突标记，再 generate：
```bash
# 方法：checkout upstream 的 ent 生成文件（不含 schema）
git ls-tree -r upstream/main --name-only | grep "^backend/ent/" | grep -v "/schema/" | xargs git checkout upstream/main --
# 然后删除我们有但 upstream 没有的 payment/referral 生成文件
rm backend/ent/paymentorder*.go backend/ent/paymentplan*.go backend/ent/userreferral*.go
# 再 generate
docker run --rm -v "$PWD/backend":/app -w /app golang:1.26 go generate ./ent/...
```

**Q: wire_gen.go 如何处理**

A: `wire gen` 在 Go 1.26 下不可用，必须手动维护。
合并策略：以 upstream 为基础，把我们的 payment/referral provider 注入手动加回去。
参考行：`paymentService`, `paymentExpiryService`, `referralService` 的初始化。

**Q: DashboardView.vue 编译报错 retryAll not found**

A: upstream 删除了 `retryAll` 函数，统一改为 `refreshAll`。
把模板里的 `@click="retryAll"` 改为 `@click="refreshAll"`。
