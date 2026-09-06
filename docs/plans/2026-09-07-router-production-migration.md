# Router 生产环境迁移记录

状态：Router、Studio、Landing Page 的正式域名已由用户切至 Router 生产环境，服务与数据库路径实测通过；旧 Router/Studio 生产容器已停止并禁用自动重启。旧机为 DNS 缓存客户端临时转发 Router/Studio 到 Router 生产环境。每日 NAS 备份任务已配置，首轮 Router 生产环境 数据库备份仍在慢速传输，原生产完整快照及最终配置已在 NAS 校验通过。

## 历史约定（迁移已完成，以当前部署说明为准）

- 使用已恢复原机的最新生产库与原生产运行镜像；凌晨备份和 Studio 空用户重建仅作兜底。
- Router 生产环境 原数据库 sub2api_tob 与迁移生产库 sub2api 名称不同；仍使用独立 PostgreSQL 容器、数据卷、应用容器、Redis 和网络。
- 原 Router 生产环境 数据和配置先完整异机备份到 NAS MinIO，验证成功后才替换其业务入口。
- 用户自行控制内部操作，接受备份之后少量数据差异，不要求停写或最终强制同步。
- 正式 router / studio 域名保持原指向。只有用户明确说切换后才变更；切换时停止旧服务和后台任务，避免重复刷新 OAuth 凭证、执行任务或处理支付。
- 先前代理自行增加了 Router 生产环境 上游调用、生成与支付限制，用户明确反对。2026-09-07 已撤销所有额外功能限制，恢复适配 Router 生产环境 数据库连接后的原生产配置；真实请求验收另行记录，不能将健康检查等同完整切换验收。

## 当时执行顺序（不得重复恢复）

1. 导出 Router 与 Studio 生产库、原环境配置、容器参数、Redis 快照、Studio 数据卷及运行镜像，设置私有文件权限。
2. 使用两服务器之间临时、目录受限的 SSH 传输密钥传输，仅允许读取本次源备份目录；单独授权 Router 生产环境 原备份暂存上传目录。完成后删除临时授权与密钥。
3. SHA256 比对源/目标/MinIO 文件；数据库恢复在独立数据卷，使用 PostgreSQL 18，与旧实例隔离。
4. 使用当前生产镜像 digest；保留账号、JWT、TOTP、Studio HMAC、生图/支付解密及 R2 密钥，改动目标连接地址和预览入口配置。
5. 默认新建 Redis，持久账号数据从 PostgreSQL 读取，保全源 Redis 快照备查。登录会话可能重建，不复制旧运行锁或并发占用。
6. 先验证库内容和密文，再启动预览应用；核对数据变化与预期配置差异。旧 Router 生产环境 服务停止但不删除，Caddy fx 入口切换可回退。
7. 对比源/迁移库业务记录、凭证摘要、余额、订单、模板、配置、schema；验证 health/ready、Studio 签名登录、两步验证可解密、R2 内容和预览入口。
8. 记录剩余真实链路验证及资源占用；等待正式切换指令。

## 门禁与回退

- 文件校验失败或备份未异机保存：不替换 Router 生产环境 入口。
- 数据库恢复错误或关键密钥不匹配：不开放候选入口，保留错误日志及源备份。
- 已迁移副本启动后可能执行本地过期清理，因此记录启动前对比结果，不能只用后来实时计数判断复制完整性。
- Router 生产环境 配置变化失败：恢复本次保存的 Caddy 配置，启动原 Router 生产环境 容器。不得删除其数据卷。
- 正式 Router 生产环境 接收新写入后，不能通过直接切回旧数据库回退；需先保全并处理新增数据。

## 当前已确认事实

- Router 生产 v0.2.1，服务器仓库提交 3acd6da0fb0567ab5e68c1d9655dc77938ad2fe6，数据库 migration 254。
- Studio 生产 v0.13，schema 21，575 模板、5 用户、7 订单、3 额度授予、12 生成任务，核查时活动任务为 0。
- 原 Router 生产环境 sub2api_tob、数据卷及旧应用容器保留。旧应用已停止，fx.nanafox.com 已转向独立的迁移 Router；原 Router 生产环境 完整备份已先进入 NAS 并回读校验。
- 原生产 Router health 200，未签名 Studio Auth 401；Studio health/ready 200。
- Router 生产环境 当前系统可见约 3.41 GiB 内存；恢复后约 2.3 GiB 可用内存、13 GiB 剩余磁盘。此为预览负载，尚不代表正式并发容量测试。

运行状态与非敏感验收结果保存于当前任务的本地交付目录；实际配置、数据库、凭据与密钥不进入 Git。

## 执行补充

- 服务器直连传输 Router 约 835 MiB 快照耗时约 70 秒，替代 Mac 慢速中转。
- 最新生产备份和 Router 生产环境 原环境备份已经全部进入 NAS，文件回读 SHA256 一致。
- Router 生产环境 对全部 575 张 R2 模板封面完成读取和字节数/SHA256 校验。
- Studio 前端路由并未完整支持子目录，放弃 `/studio/` 预览构建，不将该实验镜像用于服务。继续使用原生产根路径镜像，临时域名为 `studio-fx.nanafox.com`，Caddy 已配置到 `127.0.0.1:18789`。用户已经登录 Cloudflare，但浏览器控制通道持续超时，已请求用户添加单条 A 记录；正式记录不变。
- Router 数据库恢复及索引耗时约 1173 秒，Studio 约 6 秒。两库完整恢复已完成。
- API 可见账号 13、用户 156、分组 23，与原生产一致；物理表含隐藏/删除记录，账号总行数为 49，不能混用两种计数。
- 启动前使用 UTC 规范化逐行摘要核对：Router 账号凭证、分组、代理、订阅、支付订单/实例、设置、schema 完全一致；已枚举 Studio 业务表完全一致。
- 两次采样之间原生产继续写入：新增 21 条使用记录，1 个用户余额变化，少量最近使用时间、更新时间及用量状态变化。用户接受此类差异，不要求最终停写同步。
- 原生产 JWT、TOTP、支付、Studio HMAC 与 R2 相关密钥保留；1 条 TOTP 实际解密通过；Studio 到 Router 生产环境 Router 的真实 HTTP 签名认证链路到达账号校验。未使用用户真实密码验证交互登录。
- 生产证书、Router 定价缓存/运行数据、Router 生产环境 目标配置及 Docker 参数另行补备至 NAS，回读校验通过。临时两条源 SSH 授权和 Router 生产环境 四个传输密钥文件均已清理。

## 历史预览运行与正式切换

独立容器为 `fx-production-router`、`fx-production-studio`、`fx-production-postgres`、`fx-production-redis`。新库为 `sub2api` 与 `nanafox_studio_prod`，旧 Router 生产环境 库为 `sub2api_tob`。旧数据库与 Redis 保留，不混用。

2026-09-07 已撤销代理自行加入的预览限制：Router 增加独立出站网络 `fx-router-egress`，Caddy 使用已实际验证的回环发布端口 `127.0.0.1:18080`；恢复源生产 token refresh、统计聚合、用量清理设置。Studio 使用 `studio.active.env`，生成/支付开关与源生产相同，保留 Router 生产环境 临时 origin 和 Router 地址。两入口的 `@migrationWrites` 拦截规则均已移除。数据库/Redis 仍独立，正式域名不变。

支付 config/plans/limits/admin config GET 200，支付配置与原生产 API 返回完全一致；不携带认证的订单及模型 POST 正常到达 Router 401 认证边界。未创建支付订单或扣款。同一管理员 Key 的 gpt-5.2 在 Router 生产环境 和原 Router 都返回 502，上游日志为账号套餐不支持该模型；未修改模型或账号配置。使用该 Key 最近成功使用的 gpt-5.5，在 Router 生产环境 完成 JSON Schema 结构化问候实测，HTTP 200，耗时 2.85 秒，返回有效中文 JSON。此结果不代表所有模型/支付扣款/生图链路都已验收。用户反馈的具体支付 URL 和错误内容仍待补充，已请求其提供不含密钥的信息。

以下为当时切换清单，现已完成的事项见文末核验；不可当作待执行命令：

1. 检查目标健康、DNS 控制能力、最新备份及活动任务，记录当前差异；遵循用户接受小量差异的决定，不强制停写同步。
2. 停止源 Router/Studio 应用及会重复使用生产凭证的后台任务，保留数据库、镜像、配置和备份；迁移初期不销毁原云机。
3. 核对已恢复的 production 功能配置及 Router 出站/回环端口；切换 Studio 的正式 origin 与 Router 签名地址，保持 PG/Redis 独立。
4. 安装已保全的正式证书和两个正式域名 Caddy 配置；先用指定 Router 生产环境 IP 的 HTTPS 请求核验，再修改正式 DNS。明确临时域名后续用途；不得再自行加入功能拦截。
5. 验证实际登录、API 上游调用、Studio 生图和文件读取、订单/支付配置及回调目标；实际支付不得在没有具体交易授权时扣款测试。
6. 将原生产 `nanafox-db-backup.timer` 的备份目标迁移到 Router 生产环境 独立 PG，并完成一次上传/回读核验。当前原机定时任务仍在原机运行，每日 UTC 19:30 加最多 10 分钟随机延迟；本次只完成了迁移快照备份，不能宣称 Router 生产环境 定时备份已接管。
7. 观察正式流量、错误、内存与磁盘，确认备份接管后再决定原服务器退役；Router 生产环境 接收新写入后回退须保全这些数据。

NAS bucket 为 `nanafox-postgres-backups`；原 Router 生产环境 备份前缀为 `fx-pre-recovery/20260906T125728Z/`，最新生产备份为 `live-production-migration/20260906T151329Z/`，目标配置/证书/缓存位于其 `supplements/` 下。实际密钥均不进入仓库。


## 2026-09-07 图像插件与首屏修复

- 用户验收确认模型调用成功后，发现图像创作/模板管理缺失。原生产独立静态插件 `/srv/nanafox/image-playground/prod-current`（release `8bba7e8`，63 文件）此前未随 Docker 镜像迁移；现已原样补齐 Router 生产环境 文件与 Caddy 同源路由，未改插件代码。
- 用户已手动将两个菜单改成 Router 生产环境 域名，保留这次用户设置。正式切换后提醒改回 Router 的提醒已创建（automation-2）；未改正式 DNS。
- 大图标 data URL 在首页内嵌两次，HTML 原始约 636 KB，Router 生产环境 本机下载测试约 21.9 秒。将完全相同的 PNG（SHA256 `934e82950d1d7e9e15d964b1882a4f8fc7b43ca55e4b6cc1acd72902c99a7735`）发布为内容寻址的静态文件，只通过单字段设置更新 `site_logo` 为 `/migration-assets/site-logo-934e82950d1d7e9e.png`；设置表前后摘要确认仅此字段变化。
- 修复后 HTML 原始约 4.3 KB，传输约 1.9 KB，同一客户端测得首页 0.56 秒、图像创作 HTML 0.30 秒、管理员菜单页 HTML 0.42 秒。此为 HTTP 资源测量，不能当成真实浏览器完整渲染时间；跨境 JS/图标首次下载仍受线路影响。
- 后续备份/恢复必须覆盖插件发布目录、静态图标目录、Caddy 路由和 `site_logo` 设置；切换正式域名时这两个相对路径无需修改。

- 图像插件 API 验证：普通用户和管理员票据签发/兑换均为 200；普通端模板列表 99，管理端 101；两个自定义菜单的 API 数据均返回用户已保存的 Router 生产环境 地址。原生产发布包 63 个文件逐个 SHA256 一致，第一页模板封面 HTTP 读取另行校验。管理员侧栏此前失败的会话状态可通过完整刷新重新加载。

- 静态包 63/63 文件摘要一致，首屏 20 个模板封面逐个本机 HTTP 200，共约 10.9 MB。封面已有一年 immutable 缓存，但首次加载仍会受 Router 生产环境 线路/带宽影响；本次未改封面内容或生成缩略图，不能宣称整页图片下载也在 0.56 秒内完成。


## 2026-09-07 正式切换准备与 DNS 权限阻塞

- 用户明确说“好 没发现什么问题 可以切换了”，已获得正式切换授权；不再增加功能限制或要求最终停写同步。
- Router 生产环境 Caddy 的原 Router 生产环境 站点增加 `router.nanafox.com`，原 Studio Router 生产环境 站点增加 `studio.nanafox.com`，保留图像插件和静态 logo 路由。复制原机 Caddy 的受管证书、私钥和元数据到 Router 生产环境 的证书存储，保留自动续签机制。
- Studio 使用同一原生产镜像与数据卷，环境文件切换为 `studio.production.env`；仅变更 `STUDIO_PUBLIC_ORIGIN=https://studio.nanafox.com` 和 `ROUTER_AUTH_BASE_URL=https://router.nanafox.com`。前一容器停止并保留为 `fx-production-studio-before-formal-cutover`，restart=no。
- 指定 Router 生产环境 IP 的正式 HTTPS 请求：Router `/health`、首页、图像创作入口，Studio `/api/ready`、首页均 HTTP 200，证书校验成功。此时未变更 DNS，不代表公网切换完成。
- Cloudflare 插件已接通；读取 zone 和 DNS 成功。只修改 Router A 记录 content 的 PATCH 被 Cloudflare 拒绝（10000 Authentication error），随后复查 Router/Studio 仍是旧 IP，没有部分成功。用户正处理重新授权；不要将接口拒绝误报为 DNS 切换成功。
- 原生产应用仍运行，避免 DNS 尚未改变时中断现网。其数据、配置、容器和回退条件保持原约定。

### 每日备份接管方式

- 发现原机旧脚本只列出 `sub2api`、`sub2api_test`、`nanafox_studio_test`，没有 Studio 生产库；同时 Router 生产环境 无原机的 Tailscale 网络，不能直接连接 NAS 私网 S3 地址。
- 改为 NAS 每天北京时间 03:35 主动拉取 Router 生产环境 两个生产库 `sub2api`、`nanafox_studio_prod`，以及数据库角色、容器镜像/参数、生产配置、受管证书、图像插件和静态 logo；无需依赖原生产服务器。
- NAS 程序 `/home/Nio/.local/share/nanafox-fx-production-backup/backup.py`；Router 生产环境 导出程序 `/home/nio/backups/fx-production/export-backup.py`。NAS 私钥仅保存在 NAS，Router 生产环境 公钥授权使用 `restrict` 和固定导出命令，仅接受 `backup-v1`，不允许远程 shell 或端口转发。
- 每次导出先验证 PostgreSQL 自定义归档目录，生成 SHA256 清单；NAS 上传到 `nanafox-postgres-backups/fx-production/<UTC时间>/` 后逐文件回读验证，成功才更新 `last-success.json`，失败写入 `last-failure.json`。任务通过文件锁避免并发备份。
- 首次完整任务已启动；在其上传/回读完成前，只能声称任务已配置，不能声称每日备份已验证接管。原机旧备份任务未关闭，测试库仍留在原机。

- NAS 用户 crontab 写入因系统 spool 权限失败，未改其权限；已通过 `/etc/cron.d/nanafox-fx-production` 安装 root 所有、以 Nio 用户运行的每日任务，保留其余 cron 项，cron 服务 active。首次备份仍在 Router 生产环境 导出中。


## 2026-09-07 正式切换核验结果

- Cloudflare DNS 只读复查确认：Router 00:46:47、Studio 00:46:56、主页 00:57:45、www 00:57:51（北京时间）均被用户切至 Router 生产环境；保留仅 DNS、自动 TTL。
- 正式 Router 的管理员接口：账号 13、用户 156、分组 23，均 HTTP 200；支付配置、套餐、限制和管理端支付配置 GET 均 200。使用现有管理员 Key 实测 gpt-5.5 JSON Schema 请求 HTTP 200，2.49 秒，72 tokens。
- Studio 正式 `/api/ready` 200；带真实浏览器 Origin 的不存在账号登录请求返回 401 INVALID_CREDENTIALS，表明 Studio 到 Router 可信认证链路到达账号校验。首次探测未带 Origin，得到 403 ORIGIN_REJECTED，纠正探测后通过，没有修改服务安全校验。未使用真实账号密码测试交互登录，未发起真实扣款。
- 旧生产 Studio 活动生成任务为 0，停止 `nanafox-studio-prod` 和 `sub2api-prod` 并设置 restart=no。数据库、Redis、数据卷、镜像、测试和 New API 保留。
- 本机局域网递归 DNS 仍返回旧地址，Cloudflare 权威 DNS 和部分公共递归 DNS 已返回 Router 生产环境。旧机 Caddy 的两个正式应用站点改为经证书验证的 HTTPS 转发到 Router 生产环境；其他站点保持原样。这样缓存旧 IP 的请求也进入 Router 生产环境。
- 通过正式域名发送唯一标记请求，两条请求均在 Router 生产环境 的对应 Caddy 日志中找到并为 200；当时本机仍连旧 IP，实证旧入口已转发到 Router 生产环境。证据文件 `live-cutover-proof.json`。
- 两个图像菜单恢复 Router 域名的提醒已发送，automation-2 已暂停，避免重复；菜单值由用户手动修改，代理没有代改。
- 用户补充授权迁移 Landing Page，旧生产 deploy 提交 c63fc7e7beb7ec3010a1f223dcd5bd7927c645fb，43 文件、3,367,503 字节，全部 SHA256 一致。Router 生产环境 `/srv/nanafox-landing/current` 静态托管，不新增常驻进程；主页和 www 的正式 HTTPS 首页 SHA256 均与文件一致。源码和深圳预览未改；部署文档在独立 Landing worktree 更新。
- 用户明确 pick 与 newapi 不迁移；测试暂不部署 Router 生产环境。旧机尚未关闭或销毁，保留其他工作负载与备查数据。
- 当前配置、容器参数、原机转发及回退 Caddy、备份程序另存 NAS `.../supplements/formal-cutover/formal-cutover-configs.private.tar.gz`，16,221 字节，SHA256 c0bd0292b5585e1a1fa2efff64de1a99f00e8ec73c707a6aa1f4782859fb485d，回读一致。
- 首轮 Router 生产环境 两库快照时间 2026-09-07 00:39:59，北京时间，Router 876,126,159 字节、Studio 1,877,695 字节。导出与归档目录验证完成，Router 生产环境 到 NAS 直传缓慢，完整上传和回读仍待完成；不能把已安装每日任务当成新备份已验证成功。Landing 与最终配置补备已独立完成 NAS 回读。
