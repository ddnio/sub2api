# Sub2API 生产上线计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 把当前只能在本机运行的项目，推进到可公网访问、可持续运营、具备基础安全、监控、支付、备份和恢复能力的生产级 Sub2API 服务。

**总体方案：** 第一阶段使用一台生产 Linux 服务器，在其上通过 Docker Compose 运行应用、PostgreSQL、Redis，并由反向代理负责 HTTPS。第一版优先追求简单、稳定、可恢复，不急着做高可用和复杂扩容。

**技术栈：** Go 后端、Vue 前端、Docker Compose、Caddy、PostgreSQL、Redis、Linux 运维、DNS/TLS、基于 Admin API 的支付对接。

---

## 前提假设

- 当前状态是：服务已经能在本机跑起来，但还不适合直接对外开放。
- 第一版推荐拓扑：单台云服务器 + Docker Compose。
- 推荐起步规格：4 vCPU / 8 GB RAM / 80 GB SSD 起步的 Linux VPS，带公网 IP。
- 正式上线前必须有正式域名和 HTTPS。
- 第一版必须使用 `RUN_MODE=standard`，不要误用 `simple`。

## 为什么需要服务器

本机运行足够做功能验证，但不适合长期公网运营。正式对外开放至少需要：

1. 一台有稳定公网 IP、可 24/7 在线的机器
2. PostgreSQL、Redis、应用数据的持久化与备份
3. 域名和 HTTPS 入口
4. 防火墙、反向代理和监控能力
5. 不依赖个人电脑的部署和恢复流程

建议顺序：

1. 购买 VPS
2. 绑定域名并配置 HTTPS
3. 用 Docker Compose 部署服务
4. 做安全加固和上线验证
5. 先小范围开放，再逐步放量

### 任务 1：准备生产基础设施

**文件：**
- 参考：`/Users/nio/project/github-likes/sub2api/README_CN.md`
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/README.md`
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/docker-compose.local.yml`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/production-inventory.md`

**步骤 1：确定第一版部署形态**

在下面两种方案中选一种：

- `方案 A（推荐）`：单台 VPS，Docker Compose，本地卷保存 PostgreSQL 和 Redis 数据
- `方案 B`：应用在 VPS，上数据库和 Redis 用托管服务

如果你的目标是尽快上线并减少变量，优先用 `方案 A`。

**步骤 2：购买并初始化服务器**

建议配置：

- Linux VPS，带公网 IP
- 至少 4 vCPU / 8 GB RAM / 80 GB SSD
- 开放 `80` 和 `443` 端口
- SSH 仅允许密钥登录

并记录到 `/Users/nio/project/github-likes/sub2api/docs/ops/production-inventory.md`：

- 云厂商
- 地域
- 机器规格
- 公网 IP
- SSH 接入方式
- 域名映射关系

**步骤 3：准备宿主机环境**

安装：

- Docker Engine
- Docker Compose v2
- Caddy（如果准备用宿主机反向代理）
- 基础工具：`curl`、`jq`、`openssl`、`ufw`、`fail2ban`

**步骤 4：验证服务器可用性**

在服务器上执行：

```bash
docker --version
docker compose version
openssl version
curl --version
```

预期：命令全部成功。

**步骤 5：提交**

```bash
git add docs/ops/production-inventory.md
git commit -m "docs: add production infrastructure inventory"
```

### 任务 2：准备生产配置

**文件：**
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/.env.example`
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/config.example.yaml`
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/docker-compose.local.yml`
- 修改：`/Users/nio/project/github-likes/sub2api/deploy/.env`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/production-secrets-checklist.md`

**步骤 1：生成生产环境 `.env`**

必须确认这些值：

- `SERVER_PORT`
- `SERVER_MODE=release`
- `RUN_MODE=standard`
- `POSTGRES_PASSWORD`
- `JWT_SECRET`
- `TOTP_ENCRYPTION_KEY`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`
- `LOG_ENV=production`

**步骤 2：生成强密钥**

执行：

```bash
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
```

分别用于：

- `POSTGRES_PASSWORD`
- `JWT_SECRET`
- `TOTP_ENCRYPTION_KEY`

**步骤 3：清除危险默认值**

确认：

- `JWT_SECRET` 不是空值
- `TOTP_ENCRYPTION_KEY` 不是空值
- 管理员密码不是弱口令
- 没有误设 `RUN_MODE=simple`

**步骤 4：记录密钥管理方式**

在 `/Users/nio/project/github-likes/sub2api/docs/ops/production-secrets-checklist.md` 中写清楚：

- 每个密钥存放在哪里
- 谁可以访问
- 如何轮换
- 轮换后的影响

**步骤 5：提交**

```bash
git add deploy/.env docs/ops/production-secrets-checklist.md
git commit -m "docs: add production configuration checklist"
```

### 任务 3：绑定域名并启用 HTTPS

**文件：**
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/Caddyfile`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/domain-and-tls.md`

**步骤 1：准备正式域名**

建议：

- 应用入口域名：`api.yourdomain.com`
- 可选官网域名：`www.yourdomain.com`

**步骤 2：配置 DNS**

创建 `A` 记录：

- `api.yourdomain.com -> <server-ip>`

**步骤 3：应用反向代理配置**

以 `/Users/nio/project/github-likes/sub2api/deploy/Caddyfile` 为基础，把示例域名替换成正式生产域名。

**步骤 4：验证 HTTPS 和健康检查**

执行：

```bash
curl -I https://api.yourdomain.com
curl https://api.yourdomain.com/health
```

预期：

- HTTPS 证书生效
- `/health` 返回 `200`

**步骤 5：提交**

```bash
git add docs/ops/domain-and-tls.md
git commit -m "docs: add domain and tls setup notes"
```

### 任务 4：在服务器上部署服务

**文件：**
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/docker-compose.local.yml`
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/docker-compose.yml`
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/README.md`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/deployment-runbook.md`

**步骤 1：选择 Compose 文件**

第一版生产部署建议使用 `/Users/nio/project/github-likes/sub2api/deploy/docker-compose.local.yml`，因为它把数据放在本地目录里，更利于备份和迁移。

**步骤 2：启动整套服务**

在服务器上执行：

```bash
docker compose -f docker-compose.local.yml up -d
docker compose -f docker-compose.local.yml ps
docker compose -f docker-compose.local.yml logs -f sub2api
```

预期：

- `sub2api`、`postgres`、`redis` 都正常启动
- 应用启动过程中没有 migration、配置、鉴权类错误

**步骤 3：完成首次登录验证**

至少验证：

- 管理员可以登录
- 管理后台能正常打开
- 注册/登录策略符合预期
- API Key 可以创建
- 能完成一次基础的上游转发请求

**步骤 4：形成部署和回滚文档**

在 `/Users/nio/project/github-likes/sub2api/docs/ops/deployment-runbook.md` 中写清：

- 部署命令
- 重启命令
- 回滚方式
- 日志位置
- 重启后如何验证服务正常

**步骤 5：提交**

```bash
git add docs/ops/deployment-runbook.md
git commit -m "docs: add deployment runbook"
```

### 任务 5：在正式开放前完成安全加固

**文件：**
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/.env.example`
- 参考：`/Users/nio/project/github-likes/sub2api/frontend/src/types/index.ts`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/security-baseline.md`

**步骤 1：收紧服务器入口**

执行：

- SSH 只允许密钥登录
- 关闭密码登录
- 防火墙仅放行 `22`、`80`、`443`
- 需要时启用 `fail2ban`

**步骤 2：收紧管理员入口**

执行：

- 使用强管理员密码
- 给管理员开启 TOTP
- 不多人共用同一个管理员账号
- 如果条件允许，考虑给后台加 IP 限制或 VPN 入口

**步骤 3：降低滥用面**

检查并配置：

- 邮箱验证策略
- 注册邮箱后缀白名单（如有需要）
- 若开放公开注册，启用 Turnstile/验证码
- API Key 默认配额和限流
- 用户、分组、账号层面的并发限制

**步骤 4：检查日志和密钥暴露**

确认：

- 启动日志里不打印敏感密钥
- 日志有轮转
- 管理员 API Key 被当作密钥处理，而不是普通文本

**步骤 5：提交**

```bash
git add docs/ops/security-baseline.md
git commit -m "docs: add production security baseline"
```

### 任务 6：准备支付和账务运营流程

**文件：**
- 参考：`/Users/nio/project/github-likes/sub2api/docs/ADMIN_PAYMENT_INTEGRATION_API.md`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/payment-operations.md`

**步骤 1：明确商业模式**

先决定：

- 只做预充值余额，还是余额 + 订阅并存
- 是否提供试用
- 退款规则
- 滥用和拒付处理规则

**步骤 2：围绕现有 Admin API 补齐支付侧流程**

正式上线前要明确：

- 订单支付成功状态
- 充值成功状态
- 充值失败后的重试流程
- 幂等 key 的生成规则
- 人工修正流程

**步骤 3：测试完整支付链路**

至少验证：

1. 用户完成支付
2. 支付回调验签成功
3. 充值 API 被调用
4. 用户余额正确变更
5. 重复回调不会重复加钱

**步骤 4：形成运营处理手册**

在 `/Users/nio/project/github-likes/sub2api/docs/ops/payment-operations.md` 中覆盖：

- 正常充值
- 重复回调
- 充值失败重试
- 人工余额修正
- 退款处理

**步骤 5：提交**

```bash
git add docs/ops/payment-operations.md
git commit -m "docs: add payment operations runbook"
```

### 任务 7：建立监控、告警和日常巡检

**文件：**
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/.env.example`
- 参考：`/Users/nio/project/github-likes/sub2api/frontend/src/router/index.ts`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/monitoring-and-alerting.md`

**步骤 1：先定义最小监控集**

至少关注：

- 服务健康状态
- 登录失败异常增高
- 上游 4xx/5xx 激增
- 延迟明显上升
- Redis 连接异常
- PostgreSQL 连接异常
- 磁盘空间不足
- 备份成功/失败

**步骤 2：优先利用项目内已有能力**

这个项目本身已经有 ops 监控和定时测试相关能力，第一阶段优先把这些内建能力真正用起来，再决定是否引入外部监控系统。

**步骤 3：补一条外部告警通道**

至少准备一种：

- 邮件
- Telegram
- 飞书
- Slack

**步骤 4：写清日检和周检项**

建议包括：

- 查看错误趋势
- 查看上游账号失败情况
- 查看充值和账务异常
- 查看定时测试失败项
- 查看磁盘和备份结果

**步骤 5：提交**

```bash
git add docs/ops/monitoring-and-alerting.md
git commit -m "docs: add monitoring and alerting plan"
```

### 任务 8：建立备份和恢复流程

**文件：**
- 参考：`/Users/nio/project/github-likes/sub2api/deploy/docker-compose.local.yml`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/backup-and-recovery.md`

**步骤 1：明确必须备份的内容**

至少包括：

- PostgreSQL 数据
- Redis 数据（如果业务上需要恢复）
- 应用数据目录
- 部署 `.env`
- 自定义 pricing/config 文件

**步骤 2：设定备份周期**

推荐起步策略：

- 数据库：每天备份
- 应用/配置：每天备份
- 异地副本：每天同步
- 保留策略：7 份日备份 + 4 份周备份

**步骤 3：做一次恢复演练**

至少在一个新环境中验证：

- 管理员可以登录
- 用户数据存在
- API Key 仍然可用
- 余额和用量数据完整

**步骤 4：写清恢复目标**

在文档中定义：

- 可接受恢复时间
- 可接受数据丢失窗口
- 恢复命令
- 责任人

**步骤 5：提交**

```bash
git add docs/ops/backup-and-recovery.md
git commit -m "docs: add backup and recovery runbook"
```

### 任务 9：准备对外用户文档

**文件：**
- 参考：`/Users/nio/project/github-likes/sub2api/README_CN.md`
- 参考：`/Users/nio/project/github-likes/sub2api/README.md`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/user/getting-started.md`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/user/billing-and-limits.md`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/user/faq.md`

**步骤 1：写用户入门文档**

至少说明：

- 如何注册
- 如何购买余额或套餐
- 如何创建 API Key
- 如何调用网关

**步骤 2：写计费和滥用规则**

至少说明：

- 计费单位
- 配额含义
- 限流规则
- 重试预期
- 禁止用途

**步骤 3：写支持和故障说明**

至少说明：

- 支持渠道
- 支持时间
- 响应预期
- 哪些行为会被视为滥用

**步骤 4：发布前做一次内部核对**

确保用户文档和实际部署行为一致，不要出现文档先行但产品没兑现的情况。

**步骤 5：提交**

```bash
git add docs/user/getting-started.md docs/user/billing-and-limits.md docs/user/faq.md
git commit -m "docs: add initial user-facing launch docs"
```

### 任务 10：分阶段开放流量

**文件：**
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/launch-checklist.md`
- 新建：`/Users/nio/project/github-likes/sub2api/docs/ops/incident-playbook.md`

**步骤 1：先做小范围内测**

优先开放给：

- 自己
- 少量可信用户
- 少数真实外部用户

目标是验证支付、配额、限流、支持流程是否真的能承受真实流量。

**步骤 2：复盘首周信号**

重点看：

- 平均延迟
- 上游可用性
- 用户投诉量
- 充值失败率
- 退款请求
- 滥用尝试

**步骤 3：建立事故处理流程**

在 `/Users/nio/project/github-likes/sub2api/docs/ops/incident-playbook.md` 中定义：

- 故障等级
- 服务降级动作
- 对用户的通告模板
- 谁可以关闭注册
- 谁可以暂停收款

**步骤 4：建立最终上线检查单**

在 `/Users/nio/project/github-likes/sub2api/docs/ops/launch-checklist.md` 中逐项检查：

- 基础设施
- 安全
- 支付
- 备份
- 支持
- 监控
- 文档

**步骤 5：提交**

```bash
git add docs/ops/launch-checklist.md docs/ops/incident-playbook.md
git commit -m "docs: add launch checklist and incident playbook"
```

## 建议时间线

### 第一阶段：本周内完成

- 买 VPS
- 绑域名
- 部署基础环境
- 固定密钥和管理员安全配置
- 验证服务健康和登录能力

### 第二阶段：接下来 3 到 7 天

- 跑通支付链路
- 做备份和恢复演练
- 建立监控和告警
- 完成基本运营文档

### 第三阶段：小流量开放

- 邀请第一批外部用户
- 观察错误、滥用、支付异常和支持压力
- 调整限流、配额和价格

### 第四阶段：正式公开

- 发布用户文档
- 发布价格、条款和支持说明
- 逐步开放注册

## 正式公开前的退出条件

在下面这些条件未全部满足前，不要大规模公开推广：

- 生产服务器已稳定在线
- 正式域名和 HTTPS 已正常工作
- JWT 和 TOTP 密钥已固定
- 管理员密码和 2FA 已启用
- 至少完成一次成功备份和一次成功恢复演练
- 支付充值链路具备幂等和重试能力
- 健康检查和告警可以在应用外部被观察到
- API Key 滥用防护和限额策略已配置
- 面向用户的计费和支持文档已准备完成
