# Router v0.2.23 小版本发布记录（/docs 顶部「注册，领 $5」按钮）

日期：2026-11-24，北京时间。授权：运营转达老板优先做 /docs 注册按钮；PR #85 经 agy 外部审查、按意见修改后合并。

## 范围与制品

- 只更新唯一生产 Router；Studio、PostgreSQL、Redis、Caddy、DNS 都没有发布或重启。
- 源码：`main` 合并提交 `e960b70a967abfab6af905f1e480e9a8c2ae6187`（PR #85）。**只改前端，无数据库迁移**。改动：
  - /docs 正文上方加注册引导条：未登录、且公共设置 `registration_enabled === true` 时显示；有赠送额度时标题「注册即送 $5 体验额度，照着文档马上用」、按钮「注册，领 $5」，链到 `/register`。
  - 预渲染 HTML 不含引导条（客户端按设置渲染，避免已登录用户闪现、关注册后仍显示）。
- 本机构建：pnpm 9.15.9 按锁文件构建前端（含 /docs 预渲染）；Go linux/amd64，CGO=0，`-tags embed`，`-trimpath`。版本 `0.2.23`，构建时间 `2026-11-24T02:49:59Z`，程序 SHA256 `c3afc9da9a2717d10eb61f34619dcc69496328bbc9f8367e26868bad465d10ee`。
- 镜像以 `sub2api:prod-v0.2.23-34e75467c` 为底，只替换 `/app/sub2api`、`/app/resources` 和 revision 标签。构建上下文 SHA256 `a0d75a40…`，上传前后一致。生产镜像 `sub2api:prod-v0.2.23-e960b70a9`，ID `sha256:b3123d151ce80c8aaa5fee949869f813f38af4652dc41b42bb454e6ad63954b8`；镜像内程序 SHA256 与本机一致，`--version` 显示 commit `e960b70a9…`。

## 预检与备份

- `deploy/release/verify-prod.sh preflight` 通过；`-tags embed` 预渲染页面测试通过。
- 同盘恢复点 `/home/nio/releases/router-e960b70-20261124/`（700）：`pre-release-sub2api.dump`（`pg_dump -Fc`，223,024,670 字节，SHA256 `0ec11ef9d5acfc33c1df7e5216c0aa165abc6ec816d60f16eede16d72d7841dc`，`pg_restore -l` 可读 112 张表的 `TABLE DATA`）、`compose.json.pre`、`router-inspect.private.json`（均 600）。同盘副本不能当异机备份。

## 切换与验收

- 02:54:56Z 将 compose 里 Router `image` 从 `sha256:62ee784f…` 改为 `sha256:b3123d15…`（diff 仅此一行，比对前遮掉数据库密码），只重建 Router，02:55:12Z `healthy`。
- `verify-prod.sh post-switch` 通过：迁移 333 → 333、用户 277 → 277、订单 212 → 212、余额合计只有正常消耗；重启 0、无 OOM；近 15 分钟 fatal/panic/迁移错误 0 行；线上入口 JS 与本机构建 SHA256 一致。
- 线上 /docs 注入 `registration_enabled: true`、`default_balance: 5`；线上 JS 包含引导条文案；带标记 `probe=verify-e960b70-20261124` 的请求在网关日志命中。

## 回退

- 把 compose 里 `image` 改回 `sha256:62ee784f8733bd0de47c42d08ba0c82a2ab093db91777b31e173e326500ed199`（tag `sub2api:prod-v0.2.23-34e75467c`），再执行同一条只重建命令。无数据库迁移；不要用恢复点覆盖生产库。
