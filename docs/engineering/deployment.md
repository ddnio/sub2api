# NanaFox 当前部署与运维

最后核对：2026-09-09。本文是生产目标的唯一部署入口；历史日期文档、迁移前双生产模板不能覆盖本文。运行前再次检查宿主机与 Docker 状态。

## 当前生产拓扑

唯一生产主机：阿里云新加坡 Router 生产环境，`nio@43.106.8.109`，hostname `iZt4nfdwwipifuqc4q2v01Z`，实际 2 vCPU / 4 GiB（系统约 3.41 GiB）。`router` 和 `fx` 是同一 Router 服务的两个入口，不是两套生产。

| 服务 | 域名 | 容器与宿主机回环端口 | 数据/配置 |
|---|---|---|---|
| Router v0.2.4 | router.nanafox.com、fx.nanafox.com | fx-production-router，18080 → 8080 | sub2api，role router_app |
| Studio v0.13.0 | studio.nanafox.com；studio-fx 为临时别名 | fx-production-studio，18789 → 8788 | nanafox_studio_prod，role studio_app |
| PostgreSQL 18 | 仅内部网络 | fx-production-postgres | fx-production-pgdata |
| Redis 8 | 仅内部网络 | fx-production-redis | fx-production-redisdata，AOF |
| 图像创作插件 | Router /tools/image-playground/ | Caddy 静态文件 + Router 内置 API | /srv/nanafox/image-playground/prod-current，release 8bba7e8 |
| Landing | nanafox.com、www.nanafox.com | Caddy 静态文件，无 Node/Docker | /srv/nanafox-landing/current |

`R=/home/nio/recovery/live-production-20260906` 是现有配置目录，不是可随意清理的临时目录：

- Router **实际挂载** `R/router.preview.json` → `/app/data/config.yaml:ro`。历史文件名虽叫 preview，内容已是完整生产配置；`router.production.json` 是生产副本。后续从实际挂载来源延续，不能按名字切回早期预览配置。
- Studio **实际 env 文件** `R/studio.production.env`；`studio.active.env` 是早期临时域名配置，不用于当前生产。
- Studio `STUDIO_PUBLIC_ORIGIN=https://studio.nanafox.com`、`ROUTER_AUTH_BASE_URL=https://router.nanafox.com`。生图、支付均保持生产开启，保留原签名及密文解密密钥。
- Router 网络 `fx-production-internal` + `fx-router-egress`；Studio 网络 `fx-production-internal` + `fx-studio-egress`，数据卷 `fx-production-studiodata:/data`。
- 两应用当前均限内存 512 MiB / 1 CPU、no-new-privileges，Router nofile=100000，Studio stop-timeout=600。复用运行参数，不从旧模板重新猜测。
- Logo 文件 `/srv/nanafox/fx-web-assets` 经 `/migration-assets/` 提供；不能漏备份或删除。
- Caddy 当前文件 `/etc/caddy/Caddyfile`，只读运行快照 [当前快照](../../deploy/Caddyfile.router-production.reference)。先读实时配置并合并目标站点，验证后 reload；不能覆盖其他已有站点。

## 旧服务与测试边界

旧主机 `108.160.133.141` 的 `sub2api-prod`、`nanafox-studio-prod` 已停止且 restart=no。旧 Caddy 的 Router/Studio 站点临时 HTTPS 转发到 Router 生产环境，处理 DNS 缓存。数据库、镜像和备份仍保留；**整台旧机没有退役**。

新主机上迁移前的独立服务 `sub2api-prod` 已停止，原 `sub2api_tob` 库、旧 PG/Redis 与其他业务保留。生产只连接上述 fx-production 栈；库名不同也不代表可复用旧容器或凭据。

`pick`、`newapi` 用户明确不迁移。当前没有 Router/Studio 测试环境；旧主机上遗留的 `sub2api-test`、测试配置、端口和历史域名都不是发布目标。不要因合并 main 自动部署、把遗留容器当成测试环境，或重新启动旧生产。

## 应用更新流程

1. 明确本次授权的项目、版本和 Router 生产环境目标，固定 release 提交与候选镜像 digest。先按 Git 工作流验证候选。当前没有测试环境；未明确提供新的验证目标时，只报告本地验证缺口，不登录旧机部署，也不擅自占用生产主机建测试栈。
2. 在 生产主机对当前运行态执行只读检查（脚本可从当前受审阅版本经 SSH stdin 运行）：

   ```bash
   ssh nio@43.106.8.109 'python3 -' < deploy/check-router-production.py
   ```

   错主机、错挂载、旧卷或网络不匹配即失败；通过只证明部署目标匹配，不等于应用验收或备份通过。
3. 私有保存当前 `docker inspect`、image ID、实际配置/环境文件和 Caddy；核对备份时间。不要把 env、密钥或 inspect 全量输出到终端/仓库。延续 Router JWT/TOTP、Studio HMAC、生图/支付解密和 R2 配置；候选数据库/Redis必须指向当前生产栈。
4. 按实际 inspect 参数准备本次候选启动命令，逐项核对镜像、挂载、两张网络、回环端口、资源限制、日志配置、restart 与停止窗口。只替换目标应用，保留上一镜像和可回退容器；不重建 PostgreSQL/Redis，不执行 compose down -v。Studio 配置连续性用 Studio 仓库 `deploy/check-config-continuity.mjs` 核验，保留 600 秒停止窗口并处理活动任务。
5. 使用已验证的同一镜像进行应用更新；不隐式 `git pull` 未固定分支再构建。**旧 `bash deploy/deploy-server.sh prod` 已禁用**，该脚本不能创建当前生产布局。当前仓库没有经验证的全自动 生产替换脚本，不能把只读预检误当作部署完成。
6. 检查下列验收项；失败先保全新数据，再恢复上一个兼容应用镜像及当前生产配置。不要恢复旧数据库覆盖新增业务记录。
7. 应用验证通过后按发布分支流程合并 main，记录已运行 image ID 与源码提交。常规应用发布不需要重切 DNS。仅文档/运维规则更新不重建应用。

`deploy/deploy-server.sh` 已完全退役：当前没有测试环境，且该脚本不能创建现有生产布局。遗留的 `sub2api-test`、`/etc/sub2api/test.yaml`、回环 8081 和 `router-test.nanafox.com` 不能作为当前发布依据。

## 验收与真实流量证明

- Router `/health`、首页、插件静态资源、用户/管理员插件 API；未签名 `POST /internal/v1/studio-auth/login` 应返回 401，404 表示 trusted adapter 配置缺失。
- Studio `/api/health`、`/api/ready`、正式深链、带正式 Origin 的认证请求、可信 Router 签名链路与 R2 文件读取。
- 核对账号、用户、分组、订单、余额、模板及关键配置；区分 API 可见计数和包含软删除记录的物理表行数。
- 模型调用用确实支持的模型；支付配置读取、下单与真实扣款分别记录。真实付款需要具体交易授权。未执行的生成/交互登录不能写成通过。
- 确认容器实际网络与 PG 连接：router_app→sub2api、studio_app→nanafox_studio_prod；账号长期数据在 PG，Redis 缓存/会话/锁不等于完整账号备份。
- DNS 核对权威 A 与递归缓存，再以唯一 query 标记请求正式域名，在 Router 生产环境 Caddy 日志查到同一标记。持续运行的 ping 只反映最初解析；旧 IP 的 HTTPS 可能经旧机桥接到 Router 生产环境。单看 IP 或 HTTP 200 不证明新应用处理了请求。
- 观察内存、磁盘、队列、错误和备份，4 GiB 不是已完成生产并发容量验证。

## 已部署版本与回退边界

当前生产版本：

- Router 于 2026-09-09 部署源码 `8caa9e0c6173f70824006299248e9ea008de36fb`，服务器镜像 `sha256:9ba863606057cc1d58f97b933748425afe4d20d52ae4838a265d044bfdcaef0f`（tag `sub2api:prod-v0.2.4-8caa9e0c6`），数据库迁移至 `257_add_minimax_platform.sql`。切换前容器保留为 `fx-production-router-rollback-v0.2.1-7c1552fca`，其镜像为 `sha256:543d5975d7f60d633cd31e64998b1813d5086ed83c882a39149beb85f22db38a`。
- Studio 源码 `7c69139d273da0feaf5a66339378cd0d3cee750b`，image `sha256:f5e7004ad8bb718499414db2860945436f4278b34847c6d3a77af2f21b2cf5d1`，schema 21。
- Landing deploy 产物 `c63fc7e7beb7ec3010a1f223dcd5bd7927c645fb`，43 文件逐一 SHA256 相同。

这是迁移时基线，不应固定为未来每次发布版本。迁移后生产已有新写入，不能重新导入迁移前快照或直接启动旧生产库回退。用户本次接受少量切换差异、不要求停写，不等于未来任意数据库变更都可忽略一致性。

## NAS MinIO 备份

Bucket `nanafox-postgres-backups`：

- 原 Router 生产环境 保全：`fx-pre-recovery/20260906T125728Z/`。
- 迁移源生产完整快照：`live-production-migration/20260906T151329Z/`；插件、Landing、证书、最终配置在其 `supplements/`，已回读 SHA256 核验。
- 新日常备份：`fx-production/<UTC时间>/`，NAS 北京时间 03:35 拉取两生产库、角色、配置、容器参数、证书和静态产物；不依赖旧生产机。
- Router 生产环境 导出 `/home/nio/backups/fx-production/export-backup.py`；NAS `/home/Nio/.local/share/nanafox-fx-production-backup/backup.py`；cron `/etc/cron.d/nanafox-fx-production`。
- NAS SSH key 仅允许固定 `backup-v1` 导出。只有全部上传和回读一致后才写 `last-success.json`；检查其时间和 MinIO 对象，不能用 cron 存在代替备份成功。
- 2026-09-07 本次记录时首轮导出成功、传输缓慢，**新日常备份端到端尚未确认完成**。旧快照与补备已确认，不混淆两者。
- 新任务不包含可恢复的 Redis AOF 快照、Docker 镜像层、Studio /data 卷或 R2 全量对象；原源 Redis 快照/应用镜像有迁移备份。账号以 PG 为准，会话可能重建；需要 Redis 恢复点或新镜像异机保全时另补验证。

执行证据与时间线见 [迁移记录](../plans/2026-09-07-router-production-migration.md)。旧流程通过 Git 历史查阅，不复制回当前操作入口。

## Caddy 快照限制

`.reference` 文件记录迁移后的实际配置，不是直接发布模板。当前运行快照仍含针对静态资源的 immutable 覆盖和 text/* 压缩，与仓库旧缓存/SSE 策略不同；本次仅记录，不宣称流式取消与所有缓存场景已验证。后续改 Caddy 应审查这些差异并单独验证流式请求，不能用非流式 200 代替。

备份程序版本化参考与未覆盖范围见 [备份说明](../../deploy/backup/README.md)。

## 后续核对项（不能标为已完成）

- 新每日任务首轮完成与回读、传输时长能否满足每日窗口；目前无断点续传。
- Redis AOF、Studio /data、R2 对象和后续镜像版本的独立恢复点；具体范围见备份说明。
- 图像菜单域名：用户于 2026-09-07 明确确认已改回，此项完成（依据用户确认）。
- Caddy 流式/SSE、缓存和客户端取消行为；已验证非流式调用不覆盖这些场景。
- 旧机仍有未迁移或明确不迁移的业务，不可销毁整机。远端旧仓库可能仍带旧部署脚本；本次仓库保护不等于旧服务器 checkout 已自动同步。

## Studio 独立备份（2026-09-07）

已新增独立 Studio 生产数据库备份，NAS 北京时间 03:45 执行 `backup.py --studio`，MinIO `nanafox-postgres-backups/studio-production/<UTC>/`。独立进程锁和 `studio/last-success.json`，不被 Router 大文件备份锁阻塞。固定 SSH 命令新增 `backup-studio-v1`，仅导出 nanafox_studio_prod、角色、私有容器信息、Studio 环境配置和 Caddy/Studio 证书；不执行 shell。首轮快照为北京时间 2026-09-07 01:29:52，于 01:31:15 完成全部 5 个文件的 MinIO SHA256 回读。Studio dump 1,877,745 字节，SHA256 `bb4acf9a8f0ef92909d667ba54181a514a45520be6754d1d1c556ce89a41b663`；前缀 `studio-production/20260906T172952Z/`。Router 大备份仍在传输，不混淆两项状态。
