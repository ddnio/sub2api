# Router v0.2.23 小版本发布记录（首单引导：余额不足提示、注册页、低余额横幅、返佣入口）

日期：2026-11-24，北京时间。授权：运营转达老板批准「测试 + 审查 + 恢复点后上线」；PR #83 经 agy 外部审查、按意见修改后合并。

## 范围与制品

- 只更新唯一生产 Router；Studio、PostgreSQL、Redis、Caddy、DNS 都没有发布或重启。
- 源码：`main` 合并提交 `34e75467c2ca1c86beae24eb04717b09cad328fe`（PR #83）。**无数据库迁移**。改动：
  - 余额不足报错（`INSUFFICIENT_BALANCE`，含 Gemini 格式）保留错误码和英文开头 `Insufficient account balance`，后面加中文和后台「低余额提醒 → 充值页面 URL」（只认 http(s)，缓存 1 分钟）。
  - 公共设置新增 `default_balance`；注册页按钮「注册，领 $5」、副标题与站点名相同时不显示、选填邀请码框只在带邀请参数时显示（否则显示「有邀请码？」可展开）。
  - 登录后页面顶部低余额横幅：余额用完，或开启了低余额提醒且余额低于阈值时显示；管理员、有生效订阅的用户和充值页不显示。控制台和充值页显示邀请返佣卡片（比例取 `/user/aff` 实际生效值）。
- 本机构建：pnpm 9.15.9 按锁文件构建前端（含 /docs 预渲染）；Go linux/amd64，CGO=0，`-tags embed`，`-trimpath`。版本 `0.2.23`，构建时间 `2026-11-24T02:32:26Z`，程序 SHA256 `3713f8e686ae59dafc71ae2109c25d1988fe7062201f2eb27ad032722c5ea211`。
- 镜像以 `sub2api:prod-v0.2.23-e59e0a6c8` 为底，只替换 `/app/sub2api`、`/app/resources` 和 revision 标签。构建上下文 SHA256 `2786220e…`，上传前后一致。生产镜像 `sub2api:prod-v0.2.23-34e75467c`，ID `sha256:62ee784f8733bd0de47c42d08ba0c82a2ab093db91777b31e173e326500ed199`；镜像内程序 SHA256 与本机一致，`--version` 显示 commit `34e75467c…`。

## 预检与备份

- `deploy/release/verify-prod.sh preflight` 通过。NAS 最新完整备份 `20261123T193511Z`（`stage=complete`）。
- 同盘恢复点 `/home/nio/releases/router-34e7546-20261124/`（700）：`pre-release-sub2api.dump`（`pg_dump -Fc`，222,992,122 字节，SHA256 `811666795d80c5d9566770fc7ae5a66339fac04b6ba07b7c905fbc9d3b1b9293`，`pg_restore -l` 可读 112 张表的 `TABLE DATA`）、`compose.json.pre`、`router-inspect.private.json`（均 600）。同盘副本不能当异机备份。

## 切换与验收

- 02:34:39Z 将 compose 里 Router `image` 从 `sha256:f587ce9d…` 改为 `sha256:62ee784f…`（diff 仅此一行，比对前遮掉数据库密码），只重建 Router，02:34:51Z `healthy`。
- `verify-prod.sh post-switch` 通过：迁移 333 → 333（指纹不变）、用户 277 → 277、订单 212 → 212、余额合计不变；重启 0、无 OOM；近 15 分钟 fatal/panic/迁移错误 0 行；线上入口 JS 与本机构建 SHA256 一致。
- 公共设置返回 `default_balance: 5`、`balance_low_notify_threshold: 1`；带标记 `probe=verify-34e7546-20261124` 的请求在网关日志命中。
- 未在生产用真实用户 API Key 触发余额不足报错；报错文本由单测覆盖。

## 同日设置变更

- 老板批准后，于后台「系统设置 → 邮件设置 → 余额不足提醒」把默认提醒阈值从 `0` 改为 `1`（美元）。改前改后全量导出设置比对，只有 `balance_low_notify_threshold` 一项变化。此前阈值为 0，提醒邮件从未发出；改后只对余额从 $1 以上降到 $1 以下的用户发一次。

## 回退

- 把 compose 里 `image` 改回 `sha256:f587ce9db62cc83deb001c1887acbe749bfa01d42f3249a75b121934eb8c4a4a`（tag `sub2api:prod-v0.2.23-e59e0a6c8`），再执行同一条只重建命令。无数据库迁移，回退不涉及数据结构；不要用恢复点覆盖生产库。
- 阈值回退：后台把默认提醒阈值改回 0（只改这一项）。
