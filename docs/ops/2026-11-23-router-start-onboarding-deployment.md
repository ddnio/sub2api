# Router v0.2.23 小版本发布记录（新用户 /start 引导页）

日期：2026-11-23，北京时间。用户授权：本地真注册走通 → 外部审查 → 合并 PR → 生产只读预检后上线（PR #73 已合并至 main）。

## 范围与制品

- 只更新唯一生产 Router；Studio、PostgreSQL、Redis、Caddy、DNS 都没有发布或重启。
- 源码：`main` 提交 `5b911fe7c47b111276a729dc03f45d623a8b2d89`（PR #73）。自上一生产提交 `adcf2f2f3` 起，业务改动只有前端：新增 `/start` 页面（`views/user/StartView.vue`、`views/user/startOnboarding.ts`）和 `router/index.ts` 一处守卫；其余为文档与发布脚本（PR #71、#72）。**无数据库迁移**。
- 构建在本机独立 worktree（`deploy-5b911fe`）完成：pnpm 9.15.9 按锁文件构建前端；Go 1.27.4 交叉编译 linux/amd64，CGO=0，`-tags embed`，`-trimpath`。版本 `0.2.23`，构建时间 `2026-11-23T05:17:14Z`。程序 SHA256 `491124669d5e8684227d51abd7724aa9494f629c4fcfb095e7be4b686db81d0a`。
- 镜像沿用此前做法：以生产镜像 `sub2api:prod-v0.2.23-adcf2f2f3`（`sha256:81719120…`）为底，只替换 `/app/sub2api` 和 `/app/resources`，更新 `org.opencontainers.image.revision` 标签。构建上下文 `context.tar.gz`（约 40 MB，SHA256 `e775574f7f39740ff914082884e85b711ea524d4b1183e6a4f6b152ecb8cb428`，上传前后一致），在主机上 `docker build --pull=false`。
- 生产镜像 `sub2api:prod-v0.2.23-5b911fe7c`，ID `sha256:eb06963c35d2f3021fefbe05c28be6e38c160f93788b944b5db676f9da2163bb`，amd64。镜像内 `/app/sub2api` SHA256 与本机一致，`--version` 显示 `Sub2API 0.2.23 (commit: 5b911fe7c47b111276a729dc03f45d623a8b2d89, built: 2026-11-23T05:17:14Z)`。

## 预检与备份

- `deploy/release/verify-prod.sh preflight` 通过（含 `check-router-production.py`）。原 Router 运行镜像 `sha256:81719120…`，重启 0 次，无 OOM，内存约 140.8 MiB / 1 GiB，根盘剩余 23 G。
- NAS 最新完整备份 `20261122T193511Z`（北京时间 2026-11-23 03:36:51 完成，`stage=complete`），包含两个库。
- 发布前在主机私有目录 `/home/nio/releases/router-5b911fe-20261123/`（权限 700）保存同盘恢复点：
  - `pre-release-sub2api.dump`：`pg_dump -Fc`，218,976,907 字节，SHA256 `f890af84e5c50bb60fd5b7786f634f6683cf0d26647be744ae42babe6a068082`，`pg_restore -l` 可读出 112 张表的 `TABLE DATA`。
  - 原 `compose.json`（`compose.json.pre`，600）与容器 inspect（`router-inspect.private.json`，600）。

  此目录为同盘副本，不能当作异机备份。

## 切换

- 05:22:04Z 将 `/srv/nanafox-router-formal/compose.json` 里 Router 的 `image` 从 `sha256:81719120…` 改为 `sha256:eb06963c…`，diff 仅此一行；执行 `docker compose -f compose.json up -d --no-deps --force-recreate router`，05:22:17Z 恢复 `healthy`。
- 无数据库迁移：迁移计数维持 333 项（最新 `278_channel_pricing_ultrafast_multiplier.sql`），指纹 `15c9af48099153b062a1eb4ba4bb1661` 前后一致。
- 切换前后：用户 `270 → 270`，订单 `211 → 211`，余额合计 `16188.02133276 → 16187.82638876`（期间正常调用扣费）。

## 验收

- `deploy/release/verify-prod.sh post-switch` 通过：`running=true`、重启 0、`OOMKilled=false`；近 15 分钟 fatal/panic/迁移错误 0 行；Router `/health`、`/` 200，`/v1/models` 未登录 401，Studio `/health` 200；线上入口 `/assets/index-BsIiHodX.js` 与本机构建 SHA256 一致。
- `/assets/StartView-BQAt5uw5.js` 返回 200，内容为新版。
- 带唯一标记（`probe=verify-5b911fe-20261123`）的请求在 `/srv/nanafox-gateway/logs/router.log` 中命中。
- 上线前本地用临时库真实注册走完 4 步（注册后到 /start、建 key 进「ClaudeCode专用」、使用密钥弹窗地址不带 /v1、测试显示失败原因、打标记后不再拦截）。生产新用户默认余额 $5，生产「OpenAI GPT」「ClaudeCode专用」两个分组都未开「仅限 Claude Code 客户端」。
- 未执行范围：未在生产注册测试账号（邮箱注册和 GitHub 登录到 /start 的真实路径待用测试账号确认），未用客户 API Key 调用模型，未执行任何写库 SQL。

## 回退

- 把 `/srv/nanafox-router-formal/compose.json` 里的 `image` 改回 `sha256:81719120553b9609a84f78580dcbbed11002b15ce104c5c80ad1494cc2eae18c`（tag `sub2api:prod-v0.2.23-adcf2f2f3`），再执行同一条只重建命令。旧镜像已保留在主机上。
- 本次无数据库迁移，回退不涉及数据结构变更；回退后不要用旧库快照覆盖生产库。

## 补记：同日下午 d6ea0a51c（PR #75）

- 改动：路由守卫把推广链接里的 `?use=` 存进 sessionStorage，/start 在 URL 没带时读取。只改前端，无数据库迁移。外部审查（agy）无阻塞问题；前端全量 410 文件 / 3261 测试通过。
- 程序 SHA256 `c22215a2a86c2afd8f8fd3b84fe4fe6632db47bf81d612d790f1b812a714e9c4`，构建时间 `2026-11-23T07:25:02Z`；以 `sub2api:prod-v0.2.23-5b911fe7c` 为底叠加构建，镜像 `sub2api:prod-v0.2.23-d6ea0a51c`，ID `sha256:61cbb116435c2d9d5d41f7eb1ff8a64cd30a3280a09674c273b397efec8dd473`，镜像内程序哈希与本机一致。
- 预检通过；NAS 最新完整备份 `20261122T193511Z`；恢复点 `/home/nio/releases/router-d6ea0a5-20261123/`（`pg_dump -Fc` 219,171,010 字节，SHA256 `74d9c121e63056f73a770f06e0fd77fcd13e4d8908fd9143999b724b312d20d1`，112 张表数据可读，另有 compose 和 inspect）。
- 07:30:05Z 切换镜像（compose diff 仅 image 一行），07:30:18Z 恢复 healthy。post-switch 验收通过：迁移 333 不变，入口 JS `/assets/index-NaGcmd-b.js` 与本机一致，带标记 `probe=verify-d6ea0a5-20261123` 的请求在网关日志命中。切换前后新注册 2 人，余额合计随之增加，订单数不变。
- 回退目标：`sha256:eb06963c35d2f3021fefbe05c28be6e38c160f93788b944b5db676f9da2163bb`（tag `sub2api:prod-v0.2.23-5b911fe7c`）。
