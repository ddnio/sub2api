# 大请求错误快照保护：受控生产部署

日期：2026-10-12，北京时间。范围：仅第一批快照入口保护。部署完成，功能验证通过；资源观察进行中。

## 授权与版本

用户在获知两个 85 MiB 请求并发未通过容量门槛后，进一步明确要求“先合并，没问题就先部署验证”。本次据此先合并再受控替换生产，作为单次发布顺序例外记录；不改变常规 release → 生产验证 → main 的仓库流程，也不把已知容量风险标为解决。

- PR：[ddnio/sub2api #64](https://github.com/ddnio/sub2api/pull/64)，已合并；main merge commit `e315e4bc1bcbbee99846a70eca4ff02a3427d885`。
- 应用源码提交 `d32154752f70e07c1ddd7a738a8e5ddf5e5514a4`。后续 PR 提交只补文档；合并后的 backend/frontend/Dockerfile/entrypoint 与已构建源码相同。
- 镜像 tag `sub2api:prod-v0.2.14-snapshot1-d32154752`，应用仍显示 0.2.14，binary commit 显示上述 hotfix。
- 服务器运行 image ID `sha256:841dd9d2b88a13922d8f7788608deff6433f37fad448acd6f33c09e348edd995`。本机 image ID 为 `sha256:a799ed0a9f6f4b9f73441ffb70b729424138cc8daa634368941f6795a6c2e862`；归档导入后两种 Docker 存储中的 ID 不同，逐一核对完整文件系统层、源码 label 与 binary version，未仅凭 tag 接受镜像。
- 传输归档 51130273 字节，SHA256 `07858513a6ad415f5439d765b9781c7e9cb790f96b491bf17a07910425161fac`；服务器回读一致。

## 合并前检查与既有 CI 失败

只修改一个生产代码文件及两个已有测试文件，未合并上游。相关测试重跑通过，此前 race、图片、工具与 reasoning 回归通过；候选镜像构建、原生 Linux 单请求验证证据保存在 [第一批验证记录](2026-10-12-ops-snapshot-guard-validation.md)。

PR CI 的三项集成测试失败与合并前 main 相同：两个 migration-239 测试找不到已更名的 fixture，scheduler-cache grace-period 测试取得 -1ns TTL。后端安全检查的 12 个 Go advisory ID 也与 main 最近检查相同；本次没有修改 Go toolchain、go.mod 或 go.sum。不能称全量 CI 通过，也没有为本修复混入依赖升级或其他测试修复。

## 替换与连续性

- 22:13:40 完成目标 `nio@43.106.8.109` 上的 `fx-production-router` 替换，耗时约 10.359 秒；选择 Redis 账号活跃槽位降为 0 的时刻开始停止原容器。该瞬时读数减少了中断风险，不保证两个步骤之间绝无新请求进入。
- 新 PID 3368383、restart=0。本机 health 首次恢复约 97 ms。
- 当前环境变量、entrypoint/cmd、只读 config 挂载、回环端口、两张网络、内存与 swap、CPU、安全选项、nofile 和日志配置逐项与原 inspect 核对一致。
- 仍为 1536 MiB 内存、内存加 swap 总上限 2048 MiB、1 CPU；未添加 GOMEMLIMIT/GOGC 或调整重试。
- 配置文件及 Caddy 文件 SHA256 均不变。Studio、PostgreSQL、Redis 的镜像和 PID 不变；生产连接仍为 router_app → sub2api、studio_app → nanafox_studio_prod。
- 迁移计数仍为 319；物理行数核对：users 174、accounts 49、groups 25、api_keys 255，替换前后相同。正常业务可继续写入，这不构成数据库恢复点。
- 最新 NAS 日常备份成功完成于 12:03:48，六对象已有回读记录；此次 MinIO manifest 再次回读 SHA256/大小一致。未重新回读全部 dump 或演练恢复。

全量 inspect、环境、生产配置、候选 create 参数、替换/回退程序及安全验证结果私有保存在 `/home/nio/releases/router-snapshot-guard-20261012`，不输出或提交密钥。本次按实际 inspect 构造 Docker API create/replacement 参数，没有调用已退役部署脚本。

## 正式入口功能验证

22:14:32 验证完成：

| 验证 | 结果 |
|---|---|
| 正式 Router health / 首页 / 图片插件静态页 | 均 200，约 8–10 ms |
| 未签名 Studio auth login | 401 |
| 签名 Studio auth resolve 已有身份 | 200，身份匹配；不输出用户信息或签名 |
| 现有 Studio API Key 的模型目录 | 200，28 项 |
| Studio health / ready | 均 200 |
| 合成 1126427 字节错误请求 | 本地校验 400，约 18 ms，未调用上游 |
| 合成请求对应错误快照 | error ID 58382；原体积准确、truncated=true；JSONB 文本仅 80 字节，包含 request_body_too_large 标记，不含 authorization |
| 正式 gpt-5.5 medium 短文本 | response.completed；首个可见文本 3.401 秒，总耗时 3.547 秒 |

合成请求标记 `snapshot-guard-10b1f57f2112`；模型请求标记 `snapshot-model-463979efe637`。只使用合成正文和既有测试凭据，不重放用户内容，不向生产发送 59/85 MiB 压测请求，不做真实支付或主动生图验收。

## 资源观察与剩余风险

22:14:30 开始 30 分钟观察，每 5 秒采样正式 health、memory.current、swap 和限额/OOM 事件，每分钟保存聚合结果。最终结果待本轮观察完成后填写。

两个 85 MiB 请求并发的隔离验证仍触及 1.5 GiB，不能因上线后短文本正常而标为完全解决。快照保护消除了诊断路径的额外全量处理；OAuth 完整解码、转换与重试驻留没有改变。短文本首字与发布前单个样本的差异不证明整体延迟改善。

## 回退

原容器保留为 `fx-production-router-rollback-v0.2.14-1de7382df-20261012-snapshot1`，停止且 restart=no；原 image ID `sha256:03e0e0aae0417452768e897667fcfe7c5ee6df05009cf4ec9f7efbe0b0014d1a` 保留。回退沿用当前配置与 1.5 GiB 额度，不恢复旧数据库、旧容量或旧生产主机。本次准备的程序包含目标/image 检查、失败时恢复原容器的路径；未主动执行回退演练。
