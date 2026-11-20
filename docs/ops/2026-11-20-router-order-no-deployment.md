# Router v0.2.23 小版本发布记录（订单号不再暴露自增 ID）

日期：2026-11-20，北京时间。用户授权：执行本次小版本上线部署（PR #70 已合并至 main）。

## 范围与制品

- 只更新唯一生产 Router；Studio、PostgreSQL、Redis、Caddy、DNS 都没有发布或重启。
- 源码：`main` 提交 `adcf2f2f34ae9d67e99eec25d46061563b03b16c`（PR #70）。自上一生产提交 `1177a0df6` 起仅包含两处业务代码变更（`af02ca0af`、`3c03c3b37`：付款邮件与余额充值兑换码改用订单编号 `out_trade_no`，不再暴露自增 ID），其余为文档与发布脚本。**无数据库迁移**。
- 构建在本机独立 worktree（`deploy-adcf2f2`）完成：Node v24.22.3 / pnpm 9.15.9 按锁文件构建前端；Go 1.27.4 原生交叉编译 linux/amd64，CGO=0，带 `-tags embed`。版本 `0.2.23`，构建时间 `2026-11-20T11:01:33Z`。程序 SHA256 `71f1ac283a1c73b9ccd269f26b79e230d7f15f8f669c66c45103831dc823c494`。
- 镜像做法沿用 2026-11-19 v0.2.23 方式：以当前生产镜像 `sub2api:prod-v0.2.23-1177a0df6`（`sha256:053a14a3d4a30f8253a145b3c8241fcf417283f18e532522926629bbf13694dc`）为底，只替换 `/app/sub2api` 和 `/app/resources`，并更新 `org.opencontainers.image.version="0.2.23"` 与 `org.opencontainers.image.revision="adcf2f2f34ae9d67e99eec25d46061563b03b16c"` 标签。
- 上传约 39 MB 的构建上下文（`context.tar.gz`，SHA256 `465bd410777c070ac47a799b324c9147d1f085b25c3497653c2eb2fdaab8cc62`，上传前后一致）。在生产主机上用 `docker build --pull=false` 叠加构建，未拉源码、未联网下载。
- 生产镜像 `sub2api:prod-v0.2.23-adcf2f2f3`，ID `sha256:81719120553b9609a84f78580dcbbed11002b15ce104c5c80ad1494cc2eae18c`，amd64。镜像内 `/app/sub2api` SHA256 与本机一致，`--version` 显示 `Sub2API 0.2.23 (commit: adcf2f2f34ae9d67e99eec25d46061563b03b16c, built: 2026-11-20T11:01:33Z)`。

## 预检与备份

- `deploy/check-router-production.py` 只读预检通过（通过 `sudo -n python3 -` 执行）。原 Router 状态健康，重启 0 次，无 OOM，内存约 239.7 MiB / 1 GiB，宿主机根盘剩余 24 G。
- NAS 最新完整备份 `20261119T193511Z`（北京时间 2026-11-20 03:36:48 完成，`stage=complete`，全部文件 `verified=true`），包含 `sub2api` 和 `nanafox_studio_prod` 两个库。
- 发布前在主机私有目录 `/home/nio/releases/router-adcf2f2-20261120/`（权限 700）保存同盘恢复点：
  - `pre-release-sub2api.dump`：`pg_dump -Fc`，216,058,141 字节，SHA256 `f18fa5de65821c94fb6206ec06d295b9e761e9edf5d98baf4b6588b1ff5c0929`，`pg_restore -l` 可读出 112 张表的 `TABLE DATA`。
  - 原 `compose.json`（`compose.json.pre`，600）与容器 inspect（`router-inspect.private.json`，600）。

  此目录为同盘副本，不能当作异机备份。

## 切换

- 11:04:30Z 将 `/srv/nanafox-router-formal/compose.json` 里 Router 的 `image` 从 `sha256:053a14a3…` 改为 `sha256:81719120…`，diff 仅此一行。
- 11:04:49Z 在 `/srv/nanafox-router-formal` 下执行只重建命令（`sudo -n docker compose -f compose.json up -d --no-deps --force-recreate router`），11:05:02Z 容器恢复 `healthy`，前后约 13 秒。
- 无数据库迁移：迁移计数维持 333 项（最新 `278_channel_pricing_ultrafast_multiplier.sql`），`schema_migrations` 指纹 `15c9af48099153b062a1eb4ba4bb1661` 前后完全一致。
- 切换前后数据对比：有效用户 `181 → 181`，余额合计 `15845.79963340 → 15845.79963340`，有效账号 `18 → 18`，有效分组 `27 → 27`，有效 API Key `223 → 223`，订单 `207 → 207`，全部一致。
- 启动日志里 `panic` / `fatal` / `checksum mismatch` / `migration failed` 为 0。请求体预算显示 `268435456 bytes (256.00 MB)`，显式配置生效。

## 验收

- `deploy/check-router-production.py` 只读校验通过；Router `running=true`、健康状态 `healthy`、重启 `0` 次、`OOMKilled=false`、内存约 151.1 MiB / 1 GiB。PG、Redis、Studio、Caddy 运行时间未中断。
- 正式域名访问：
  - Router `/health`、`/`、`/tools/image-playground/`、`/admin/payment/orders` 返回 200；
  - Studio `/api/health`、`/api/ready`、`/` 返回 200；
  - 未登录访问 `/v1/models`、`GET /api/v1/admin/payment/orders` 返回 401；
  - 未签名 `POST /internal/v1/studio-auth/login` 返回 401（非缺配置的 404）。
- 首页加载的 `/assets/index-BUYaZ3TQ.js` 与本机构建逐字节一致（SHA256 `6f2a462fa9b7f2a4d1ff14f8d68efadf6a524871591319dfa6e1869de2dd5fad`）。
- 带唯一标记（`probe=verify-adcf2f2-20261120`）的请求在 `/srv/nanafox-gateway/logs/router.log` 中命中，证明请求由新主机处理。
- 微信支付出口：容器内 `api.mch.weixin.qq.com` 解析为 `100.71.253.28`，访问微信接口返回 `HTTP/1.1 401 Unauthorized`，链路正常。
- 未执行范围：未持有管理员会话令牌，未做带鉴权的管理端订单列表请求或交互登录；未发起真实付款，未用客户 API Key 调用模型，未执行任何写库 SQL。

## 回退

- 把 `/srv/nanafox-router-formal/compose.json` 里的 `image` 改回 `sha256:053a14a3d4a30f8253a145b3c8241fcf417283f18e532522926629bbf13694dc`（tag `sub2api:prod-v0.2.23-1177a0df6`），再执行同一条只重建命令。旧镜像已保留在主机上。
- 本次无数据库迁移，回退不涉及数据结构变更；回退后不要用旧库快照覆盖生产库。
