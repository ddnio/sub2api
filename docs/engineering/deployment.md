# 部署文档

## 整体架构

```
ToC 服务器（108.160.133.141）                ToB 服务器（43.106.8.109）
├── 基础设施（docker-compose）               ├── 基础设施（docker-compose）
│   ├── sub2api-postgres（PostgreSQL 18）    │   ├── sub2api-postgres（PostgreSQL 18）
│   └── sub2api-redis（Redis 8）            │   └── sub2api-redis（Redis 8）
│                                            │
├── 测试环境                                 └── 生产环境
│   └── sub2api-test（端口 8081）                └── sub2api-prod（端口 8080）
│       ├── 域名：router-test.nanafox.com            ├── 域名：fx.nanafox.com
│       ├── 数据库：sub2api_test                     ├── 数据库：sub2api_tob
│       └── Redis：DB 1                             └── Redis：DB 0
│
│                                           └── Grok 上游（CLIProxyAPI）
│                                               ├── 域名：grok.nanafox.com
│                                               ├── 宿主机端口：127.0.0.1:8317
│                                               ├── 容器内网地址：http://cli-proxy-api:8317
│                                               └── 存储：本地文件（config.yaml + auth json），无 DB/Redis 依赖
│
└── 生产环境
    └── sub2api-prod（端口 8080）
        ├── 域名：router.nanafox.com
        ├── 数据库：sub2api
        └── Redis：DB 0

本地开发
└── go run（源码启动，连测试库，端口 8080）
    └── pnpm dev（前端，端口 3001）
```

---

## 一、本地开发

### 前置条件

- Go 1.26+（`~/go/bin/go1.26.x`）
- Node.js + pnpm
- Tailscale（用于连接服务器上的测试数据库）

### 配置文件

本地配置放在 `backend/config.yaml`（已在 `.gitignore`，不会提交）：

```yaml
database:
  host: "100.71.166.42"   # 服务器 Tailscale IP
  port: 5432
  user: "sub2api"
  password: "<见 /etc/sub2api/test.yaml>"
  dbname: "sub2api_test"
  sslmode: "disable"

redis:
  host: "100.71.166.42"
  port: 6379
  password: "<见 /etc/sub2api/test.yaml>"
  db: 1

jwt:
  secret: "<见 /etc/sub2api/test.yaml>"

totp:
  encryption_key: "<见 /etc/sub2api/test.yaml>"

server:
  frontend_url: "http://localhost:5173"
```

### 启动

```bash
# 后端（在 backend/ 目录下）
go run ./cmd/server

# 前端（另开终端）
pnpm --dir frontend dev
```

访问 `http://localhost:3001`。

---

## 二、服务器基础设施

### 目录结构

```
/data/service/
├── sub2api/          # 代码仓库（test 和 prod 共用）
└── cliproxyapi/      # ToB 服务器上的 Grok 上游（CLIProxyAPI）

/etc/sub2api/
├── test.yaml         # 测试环境配置
└── prod.yaml         # 生产环境配置

/opt/sub2api/deploy/  # docker-compose（基础设施）
```

### 基础设施管理

postgres 和 redis 通过 docker-compose 管理，正常情况不需要手动操作：

```bash
cd /opt/sub2api/deploy

# 查看状态
docker compose ps

# 重启
docker compose restart postgres redis
```

### 端口开放（防火墙）

5432 和 6379 仅对指定 IP 开放（本地开发机公网 IP），通过 ufw 管理：

```bash
ufw status
```

### Docker 网络

所有容器使用同一个 Docker 网络 `deploy_sub2api-network`，app 容器通过容器名访问 DB：
- 数据库 host：`sub2api-postgres`
- Redis host：`sub2api-redis`
- Grok 上游（CLIProxyAPI）host：`cli-proxy-api`（仅 ToB 服务器）

---

## 三、部署（测试 / 生产）

### 代码仓库

```bash
# 首次克隆（已完成）
git clone -b codex/workflow-docs https://github.com/ddnio/sub2api.git /data/service/sub2api
```

### 执行部署

```bash
cd /data/service/sub2api

# 部署测试环境
bash deploy/deploy-server.sh test

# 部署生产环境
bash deploy/deploy-server.sh prod
```

脚本会自动执行：
1. `git pull` 拉取最新代码
2. `docker build` 重新构建镜像（前后端一起打包进二进制）
3. 停止旧容器
4. 启动新容器（**迁移自动执行**，见第四节）

### 容器端口映射

| 容器 | 宿主机端口 | 域名 |
|------|-----------|------|
| `sub2api-test` | 127.0.0.1:8081 | https://router-test.nanafox.com |
| `sub2api-prod` | 127.0.0.1:8080 | https://router.nanafox.com |

---

## 四、数据库迁移

**迁移完全自动化**，无需手动执行 SQL。

应用启动时自动检测 `schema_migrations` 表，对比已应用的迁移记录，只执行 delta（新增的迁移文件）。已有数据和表结构完全保留。

机制特性：
- SHA256 校验和防止迁移文件被篡改
- PostgreSQL Advisory Lock 保证多实例并发安全
- `_notx.sql` 后缀的迁移在事务外执行（用于 `CREATE INDEX CONCURRENTLY`）

迁移文件位于 `backend/migrations/`，当前最大编号：`077_add_payment_tables.sql`。

---

## 五、Cloudflare + Caddy 反向代理

### Cloudflare CDN（2026-03-30 接入）

域名 `nanafox.com` 的 DNS 已迁移到 Cloudflare（NS: `eugene.ns.cloudflare.com` / `suzanne.ns.cloudflare.com`），用于优化中国大陆用户访问日本服务器的延迟。

架构：
```
用户 ──HTTPS──> Cloudflare 边缘节点 ──HTTPS──> Caddy(日本服务器) ──> Go 应用
```

Cloudflare 配置要点：
- SSL/TLS 模式：**Full (Strict)**（Caddy 有 Let's Encrypt 真实证书）
- `router` 和 `router-test` 的 A 记录均为 **Proxied**（橙色云）
- 缓存策略：使用默认规则（自动缓存静态资源，API 不缓存）
- 管理入口：https://dash.cloudflare.com（账号 Ddnio@outlook.com）

注意事项：
- 真实服务器 IP（108.160.133.141）已被 Cloudflare 隐藏，不要在公开渠道泄露
- Caddy 日志中看到的客户端 IP 是 Cloudflare 节点 IP，非用户真实 IP
- DNS 记录变更需在 Cloudflare Dashboard 操作，不再在阿里云万网管理

### Caddy 反向代理

配置文件：`/etc/caddy/Caddyfile`（服务器），代码库模板：`deploy/Caddyfile`

当前域名配置：

| 域名 | 指向 | 环境 |
|------|------|------|
| `router-test.nanafox.com` | `127.0.0.1:8081` | 测试 |
| `router.nanafox.com` | `127.0.0.1:8080` | 生产 |
| `fx.nanafox.com` | `127.0.0.1:8080` | ToB 生产 |
| `grok.nanafox.com` | `127.0.0.1:8317` | Grok 上游（CLIProxyAPI） |

Caddy 自动申请 Let's Encrypt 证书（Cloudflare 回源时会验证此证书）。

修改后重载：
```bash
echo <password> | sudo -S bash -c "caddy validate --config /etc/caddy/Caddyfile && systemctl restart caddy"
```

---

## 六、环境配置文件

配置文件存放在服务器 `/etc/sub2api/`，**不提交到 git**。

| 环境 | 文件 | 数据库 | Redis DB |
|------|------|--------|----------|
| 测试 | `/etc/sub2api/test.yaml` | `sub2api_test` | `1` |
| 生产 | `/etc/sub2api/prod.yaml` | `sub2api` | `0` |

关键配置项（test 和 prod 使用相同的 jwt.secret 和 totp.encryption_key）：

```yaml
jwt:
  secret: "固定随机值，不要更换"

totp:
  encryption_key: "固定随机值，不要更换"
```

> **注意**：`jwt.secret` 和 `totp.encryption_key` 一旦设定不要更改，否则所有用户需要重新登录，2FA 全部失效。

### 支付模块配置

支付相关密钥同样只放服务器配置文件，**不提交 git**。

```yaml
payment:
  provider: "wxpay"                               # 当前支持 wxpay / easypay
  callback_base_url: "https://router.nanafox.com" # 生产；测试改 router-test.nanafox.com

  # 微信支付商户信息（在微信支付商户平台获取）
  wxpay_app_id: "wxXXXXXXXXXXXXXXXX"        # 公众号/小程序 AppID
  wxpay_mch_id: "XXXXXXXXXX"                 # 商户号（10位）
  wxpay_api_v3_key: "32位字符串"              # APIv3 密钥（商户平台自行设置）
  wxpay_serial_no: "证书序列号"               # apiclient_cert.pem 中的 serialNumber

  # 商户 API 私钥（从商户平台下载 apiclient_key.pem 的内容）
  wxpay_private_key: |
    -----BEGIN PRIVATE KEY-----
    MIIEvAIBADANBgk...
    -----END PRIVATE KEY-----

  # 微信支付公钥模式（2024年后新商户）
  wxpay_public_key_id: "PUB_KEY_ID_XXXXXXXXXXXXXXXXXX"  # 注意必须保留 PUB_KEY_ID_ 前缀
  wxpay_public_key: |
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgk...
    -----END PUBLIC KEY-----
```

> **公钥模式 vs 证书模式**：2024 年后开通的商户默认使用公钥模式（回调 Header `Wechatpay-Serial` 以 `PUB_KEY_ID_` 开头）。此时必须配置 `wxpay_public_key_id` 和 `wxpay_public_key`。

---

## 七、迭代流程

```
本地改代码
  → git push origin <branch>
  → SSH 到服务器
  → cd /data/service/sub2api
  → git checkout <branch>               # 若服务器跟踪的分支不同，需先切换
  → bash deploy/deploy-server.sh test   # 验证测试环境
  → bash deploy/deploy-server.sh prod   # 确认后上生产
```

迁移自动执行，无需额外操作。

### 完整发布清单（main 合入后必做）

改动合并到 `main` 后，**需要同时部署到三个环境**，不能只部署其中一个。否则会出现不同业务线/域名代码版本不一致。

| 步骤 | 环境 | 服务器 | 域名 | 部署命令 |
|---|---|---|---|---|
| 1 | ToC 测试 | `108.160.133.141` | `router-test.nanafox.com` | `git checkout main && git pull && bash deploy/deploy-server.sh test` |
| 2 | ToC 生产 | `108.160.133.141` | `router.nanafox.com` | `git checkout main && git pull && bash deploy/deploy-server.sh prod` |
| 3 | ToB 生产 | `43.106.8.109` | `fx.nanafox.com` | `git checkout main && git pull && bash deploy/deploy-server.sh prod` |

验证方法：
```bash
# 在每台服务器上确认当前运行的代码 commit
cd /data/service/sub2api && git log --oneline -1
docker ps --format '{{.Names}} | {{.CreatedAt}}' | grep sub2api
```

**规则**：
- feature 分支验证通过 → 立即 cherry-pick / merge 到 main → **立即**按清单部署三环境
- 不要让 feature 分支长期悬挂（会与主线演进冲突，且容易遗漏部署）

---

## 八、ToB 服务器（43.106.8.109）

独立部署，与 ToC 服务器完全隔离，数据库独立。

### 服务器信息

| 项目 | 值 |
|------|---|
| IP | 43.106.8.109 |
| 用户 | nio |
| 域名 | https://fx.nanafox.com |
| 数据库 | sub2api_tob |
| Redis DB | 0 |
| 应用端口 | 8080 |

### 目录结构

```
/data/service/sub2api/    # 代码仓库（跟踪 main 分支）
/data/service/cliproxyapi/ # Grok 上游 CLIProxyAPI（docker-compose + config.yaml + auths + logs）
/etc/sub2api/prod.yaml    # 生产配置（含 DB、Redis、JWT、支付密钥）
/opt/sub2api/deploy/      # docker-compose（postgres + redis 基础设施）
```

### Grok 上游（CLIProxyAPI）

ToB 服务器的 Grok 上游用 [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 提供，旁路部署，**不依赖数据库**——纯文件型（config + 登录态 json + 日志），无需 PostgreSQL/Redis，也不需要 flaresolverr。Grok 走 xAI OAuth（Grok Build），消耗 SuperGrok 订阅额度（与 grok.com 网页端共享）。

> 历史：原先用 grok2api（爬 grok.com 网页 + flaresolverr 破 Cloudflare），2026-05-30 起替换为 CLIProxyAPI。

```text
/data/service/cliproxyapi/
├── docker-compose.yml    # 镜像 eceasy/cli-proxy-api:latest
├── config.yaml           # 含 api-keys + remote-management.secret-key，不提交
├── auths/
│   └── xai-*.json        # xAI OAuth 登录态（本机登录后拷贝上来，含 refresh_token 会自动刷新）
└── logs/
```

端口和访问方式：
- 宿主机只绑定 `127.0.0.1:8317:8317`，由 Caddy 反代 `grok.nanafox.com`。
- 容器加入 `deploy_sub2api-network`，`sub2api-prod` 用 `http://cli-proxy-api:8317/v1` 作为 OpenAI 兼容上游接入（渠道 API Key = config.yaml 的 `api-keys`）。
- Web 控制台 `grok.nanafox.com/management.html`，由 `remote-management.secret-key` 保护（`allow-remote: true` 才能经反代访问）。
- 登录新账号：服务器无浏览器，需在有桌面的机器上 `./cli-proxy-api -xai-login` 完成 OAuth，再把生成的 `~/.cli-proxy-api/*.json` 拷到 `auths/`。
- 控制台「Grok 额度」可能显示获取失败：查额度接口 `cli-chat-proxy.grok.com/v1/billing` 经 Cloudflare，纯显示问题，不影响推理。

### 部署

```bash
ssh nio@43.106.8.109
cd /data/service/sub2api
git pull                              # 或 git checkout <branch> 后再 pull
bash deploy/deploy-server.sh prod
```

### 基础设施管理

```bash
cd /opt/sub2api/deploy
docker compose ps       # 查看状态
docker compose restart  # 重启
```

### 配置文件关键项

```yaml
database:
  host: "sub2api-postgres"
  dbname: "sub2api_tob"       # 与 ToC 区分

server:
  frontend_url: "https://fx.nanafox.com"

payment:
  callback_base_url: "https://fx.nanafox.com"
```

> 完整密钥（DB 密码、JWT secret、TOTP key、微信支付）只在服务器 `/etc/sub2api/prod.yaml` 中，不提交 git。

---

## 历史域名说明

`sub.aibewinjpq.com`（原测试域名）已于 2026-03-23 下线，替换为 `router-test.nanafox.com`。
