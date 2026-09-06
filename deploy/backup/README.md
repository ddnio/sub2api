# Router 生产备份程序

这两份是已安装程序的版本化参考，修改仓库不等于已更新远端，也不能在开发机直接运行：

- `export-router-production.py` → 生产 `/home/nio/backups/fx-production/export-backup.py`，SSH forced command 只接受 backup-v1。
- `nas-pull-router-production.py` → NAS `/home/Nio/.local/share/nanafox-fx-production-backup/backup.py`；cron `/etc/cron.d/nanafox-fx-production` 每天北京时间 03:35，以 Nio 运行。

历史实际目录、容器名、MinIO 前缀保持兼容，不为调整环境称呼而迁移备份数据或密钥。程序运行会生成包含凭据的私有备份，不能输出文件内容或提交制品；SSH key 只留 NAS，严格校验已信任的服务器 host key。

检查最后成功时间、两个库名、对象数量和逐文件回读 SHA256，不把安装定时器视为验证通过。首轮仍在传输时后续触发会通过文件锁跳过。当前网络传输无断点续传，失连需要重跑；不能保证每日都能在下一个窗口前完成。

范围是两生产数据库、角色、配置、容器参数、证书和静态插件/主页/图标。尚未覆盖 Redis AOF、Docker 镜像层、Studio /data 卷及 R2 全量对象的日常异机备份。R2 是当前生产对象存储，模板封面读校验不等于用户作品备份。旧迁移快照只提供迁移时恢复点。

更新程序前核对远端差异及运行锁；正在传输时不停止当前任务。部署新版本后另记其 hash、首轮结果与 MinIO readback，不能将本地语法检查写成远端备份验收。
