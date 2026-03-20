# 部署文档

## 整体架构

```
服务器（108.160.133.141）
├── 基础设施（docker-compose，一直运行）
│   ├── sub2api-postgres（PostgreSQL 18，端口 5432）
│   └── sub2api-redis（Redis 8，端口 6379）
│
├── 测试环境
│   └── sub2api-test（Docker 独立容器，端口 8081）
│       ├── 数据库：sub2api_test
│       └── Redis：DB 1
│
└── 生产环境（待配置）
    └── sub2api-prod（Docker 独立容器，端口 8080）
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
└── sub2api/          # 代码仓库（test 和 prod 共用）

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

# 部署生产环境（配置好后使用）
bash deploy/deploy-server.sh prod
```

脚本会自动执行：
1. `git pull` 拉取最新代码
2. `docker build` 重新构建镜像（前后端一起打包进二进制）
3. 停止旧容器
4. 启动新容器

### 容器端口映射

| 容器 | 宿主机端口 | 域名 |
|------|-----------|------|
| `sub2api-test` | 127.0.0.1:8081 | https://sub.aibewinjpq.com |
| `sub2api-prod` | 127.0.0.1:8080 | 待配置 |

---

## 四、Caddy 反向代理

配置文件：`/etc/caddy/Caddyfile`

当前配置：`sub.aibewinjpq.com` → `127.0.0.1:8081`（测试环境）

修改后重载：
```bash
caddy reload --config /etc/caddy/Caddyfile
```

生产环境上线时，在 Caddyfile 新增一个域名块指向 8080 即可，测试环境配置不需要改动。

---

## 五、环境配置文件

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

### 支付模块配置（feature/payment-module 分支起）

支付相关密钥同样只放服务器配置文件，**不提交 git**。私钥文件也不要提交（已加入 `.gitignore`）。

```yaml
payment:
  provider: "wxpay"                          # 当前支持 wxpay / easypay
  callback_base_url: "https://sub.aibewinjpq.com"  # 测试环境；生产环境改对应域名

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
  # 在商户平台「账户中心 → API安全 → 微信支付公钥」下载 pub_key.pem
  wxpay_public_key_id: "PUB_KEY_ID_XXXXXXXXXXXXXXXXXX"  # 注意必须保留 PUB_KEY_ID_ 前缀
  wxpay_public_key: |
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgk...
    -----END PUBLIC KEY-----
```

> **公钥模式 vs 证书模式**：2024 年后开通的商户默认使用公钥模式（回调 Header `Wechatpay-Serial` 以 `PUB_KEY_ID_` 开头）。此时必须配置 `wxpay_public_key_id` 和 `wxpay_public_key`；若不配置则回退到平台证书模式（旧商户）。

---

## 六、迭代流程

```
本地改代码
  → git push origin <branch>
  → SSH 到服务器
  → cd /data/service/sub2api
  → git checkout <branch>               # 若服务器跟踪的分支不同，需先切换
  → bash deploy/deploy-server.sh test   # 验证测试环境
  → bash deploy/deploy-server.sh prod   # 确认后上生产
```

### 包含 DB Migration 时

新功能有 migration SQL（`backend/migrations/0XX_*.sql`）时，部署后需手动执行：

```bash
# 在服务器上连接对应数据库执行
psql -h localhost -U sub2api -d sub2api_test -f /data/service/sub2api/backend/migrations/0XX_feature.sql
```

按文件编号从小到大顺序执行，不要跳号。当前各功能对应的 migration：

| 分支 / 功能 | Migration 文件 |
|------------|----------------|
| 支付模块 | `077_add_payment_tables.sql` |
