# 2026-11-24 Router 发布记录：/docs 小节独立地址 + robots.txt / sitemap.xml

## 版本

- 源码：`c48de501d28ff36c0f4d8fb56764bdb05d5b1913`（PR #79 合并提交）。
- 改动：/docs 每个小节有独立地址（`/docs/claude-code`、`/docs/codex-cli`、`/docs/clients`、`/docs/opencode`、`/docs/api-usage`、`/docs/faq`，快速开始仍为 `/docs`），各自的页面标题、描述和 canonical；旧的 `/docs#小节` 链接自动跳到新地址；侧边栏改为真实链接。新增 `frontend/public/robots.txt`、`sitemap.xml`，由后端当静态文件返回（不再回落成首页 HTML）。只改前端和一个后端测试，无数据库迁移。
- 外部审查（agy）3 处必须改已修；前端全量 412 文件 / 3283 测试通过，`go test -tags embed ./internal/web/` 通过。

## 构建与镜像

- 程序 SHA256 `fea63987d21e67342db7ef9f7d8cfde37baff1f12c483f7336c6f3bc0884169f`，构建时间 `2026-11-23T23:44:01Z`。
- 以 `sub2api:prod-v0.2.23-c3d50c2a3` 为底叠加构建，镜像 `sub2api:prod-v0.2.23-c48de501d`，ID `sha256:b9db12acf3eeaf724e6ac84a7c1625e5c66f8272cfc07aa8f67b4d24ff13a478`，revision 标签为上述源码；镜像内程序哈希与本机一致，`--version` 显示同一提交。构建上下文 SHA256 `5ef85c6fb9c089de993d1ad050fce3381d2d3599e02c368ee41c68be1a17e77e`，上传前后一致。

## 发布前

- `deploy/release/verify-prod.sh preflight` 通过。
- NAS 最新完整备份 `20261123T193511Z`（stage complete）。
- 恢复点 `/home/nio/releases/router-c48de50-20261124/`：`pg_dump -Fc`（库 `sub2api`）222,828,412 字节，SHA256 `9c4e0736b5472a4dc323b009519d183faeb61978c85c5c3b06d6df6b29c0a744`，112 张表数据可读；另有 `compose.json.pre`、`router-inspect.private.json`。

## 切换

- compose diff 只有 image 一行（`f84484dc…` → `b9db12ac…`），只重建 router；容器 23:48:38Z 启动并恢复 healthy。

## 验收

- `verify-prod.sh post-switch` 通过：重启 0、无 OOM；近 15 分钟 fatal/panic/迁移错误 0 行；迁移 333 不变，用户 276 → 276，订单 212 → 212，余额不变；`/health`、`/` 200，`/v1/models` 未登录 401，Studio `/health` 200；入口 `/assets/index-BpYM0Jv5.js` 与本机构建一致。
- `https://router.nanafox.com/robots.txt`：`text/plain`，Cloudflare 托管规则在前，我们的规则（`Allow: /`、`Disallow: /api/ /v1/ /admin`、`Sitemap:`）接在后面。
- `https://router.nanafox.com/sitemap.xml`：`text/xml`，8 个 `<loc>`。
- 浏览器打开线上 `/docs/clients`（标题「图形客户端接入 - NanaFox Router」，canonical 为本地址）、`/docs#faq`（跳到 `/docs/faq`，标题「计费与常见问题 - NanaFox Router」）、`/docs`（canonical `/docs`），描述均为对应小节。
- 带标记 `probe=verify-c48de50-20261124` 的请求在 `/srv/nanafox-gateway/logs/router.log` 中命中。
- 未执行范围：未改 Cloudflare 托管 robots（区域级设置，需老板确认）；未向搜索引擎提交 sitemap。

## 回退

- 把 `/srv/nanafox-router-formal/compose.json` 的 `image` 改回 `sha256:f84484dc1288f7fbea52f44c2c326bc650c755e0a8a7978363f771011becdacb`（tag `sub2api:prod-v0.2.23-c3d50c2a3`），再执行同一条只重建命令。旧镜像保留在主机上。
- 无数据库迁移；回退后不要用旧库快照覆盖生产库。
