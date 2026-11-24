# Router v0.2.23 小版本发布记录（GitHub 一键注册/登录提到页面顶部）

日期：2026-11-24，北京时间。授权：运营转达老板批准上线；PR #89 经 agy 外部审查、按意见修改后合并。

## 范围与制品

- 只更新唯一生产 Router；Studio、PostgreSQL、Redis、Caddy、DNS 都没有发布或重启。
- 源码：`main` 合并提交 `e0abbc300654c8b3f925f57bb2c41734abcf323b`（PR #89）。**只改前端，无数据库迁移**。改动：
  - /register：开放注册且开启 GitHub 登录时，邮箱表单上方放醒目按钮「用 GitHub 注册，领 $5」（无赠送额度时为「用 GitHub 一键注册」）、小字「不用收邮箱验证码，授权后直接开始」、分隔线「或用邮箱注册」；底部第三方登录不再重复 GitHub。注册关闭时布局与原来一致。
  - /login：表单上方放「用 GitHub 登录」（后端模式下不显示）。
  - 上线前核实：GitHub 注册新用户拿全局默认额度 $5（线上 GitHub 单独赠送设置未启用，最近 8 个 GitHub 注册用户起始余额均为 $5）；OAuth 回调到 /dashboard，路由守卫把 7 天内新用户带到 /start。
  - 已知限制：若以后开启「登录协议」勾选，顶部按钮要等下方勾选后才可点；线上 `login_agreement_enabled=false`。
- 本机构建：pnpm 9.15.9（含 /docs 预渲染）；Go linux/amd64，CGO=0，`-tags embed`，`-trimpath`。构建时间 `2026-11-24T07:40:18Z`，程序 SHA256 `dda97ce88763504914a9607c8d37c079b71e1f0211eab98ef1cdf72169c0671a`。
- 以 `sub2api:prod-v0.2.23-7c965752b` 为底，镜像 `sub2api:prod-v0.2.23-e0abbc300`，ID `sha256:991e6ce878002f3c10a39a0f7cfff7400a023f30b65f01adcec274ecdb0c83eb`；构建上下文 SHA256 `ee7d333f…` 上传前后一致；镜像内程序与本机一致，`--version` 显示 commit `e0abbc300…`。

## 预检与备份

- `verify-prod.sh preflight` 通过；`-tags embed` 预渲染测试通过。
- 同盘恢复点 `/home/nio/releases/router-e0abbc3-20261124/`（700）：dump 223,791,779 字节，SHA256 `bbc8cc630486a89a509bb87a04f4397db2d20d27d279bb1d65e946cbbd215d82`，112 张表 `TABLE DATA` 可读；`compose.json.pre`、`router-inspect.private.json`（均 600）。

## 切换与验收

- 08:06:16Z 将 Router `image` 从 `sha256:3d3dabb7…` 改为 `sha256:991e6ce8…`（diff 仅此一行，比对前遮掉数据库密码），只重建 Router，08:06:29Z `healthy`。
- `post-switch` 通过：迁移 333 → 333、用户 277 → 277、订单 212 → 212、余额只有正常消耗；重启 0、无 OOM；0 fatal；入口 JS 与本机构建一致。探针 `probe=verify-e0abbc3-20261124` 命中。
- 线上 /register 注入 `github_oauth_enabled: true`、`registration_enabled: true`、`default_balance: 5`；`/api/v1/auth/oauth/github/start` 正常 302 到 GitHub，GitHub 接受回调地址。未登录无头浏览器截图确认注册页、登录页顶部按钮位置和文案正确（桌面与窄屏）。
- 同日运营把 GitHub OAuth 换成组织号 nana-fox 下的新应用（后台改 Client ID/Secret，不涉及本次镜像）。

## 回退

- 把 compose 里 `image` 改回 `sha256:3d3dabb7070d06f77506f4829304ce6eacc5770f83f2d852020c8614b3c09b59`（tag `sub2api:prod-v0.2.23-7c965752b`），再执行同一条只重建命令。无数据库迁移；不要用恢复点覆盖生产库。
