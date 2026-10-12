# Router 延迟与再次 503 恢复

时间：2026-10-12 19:46–20:26，北京时间。目标：43.106.8.109 上的 fx-production-router，v0.2.14，镜像 sha256:03e0e0aae0417452768e897667fcfe7c5ee6df05009cf4ec9f7efbe0b0014d1a。

## 已确认的故障与恢复

20:20:14 现场核对：正式域名首页和 /health 返回 503；本机 127.0.0.1:18080/health 超过 4 秒无响应。Router 仍 running，restart=0、OOMKilled=false。

容器硬内存上限 1 GiB，memory.current=1073725440，swap 已达到其 512 MiB 上限。memory.events max=2494519，oom/oom_kill=0；memory PSI full avg10=7.36%，I/O PSI full avg10=48.82%。宿主机可用约 1324 MiB，PG/Redis 容器运行。故障是容器额度内的回收/换页压力，不能由低 CPU 或没有 OOM 排除。

在核对目标、镜像、原资源上限后，仅在线调整 Router：

```bash
docker update --memory 1536m --memory-swap 2048m fx-production-router
```

新 Memory=1610612736、MemorySwap=2147483648，仍允许 512 MiB swap。PID 908683、restart=0 不变；未重启、替换镜像、改应用配置、账号、路由、Caddy、数据库或 Redis。

- 20:20:50 本机 health 已 200 / 19 ms；20:20:59 约 0.7 ms。
- 正式入口短暂仍 503，20:21:26 正式及直达源站 health 均恢复 200。Caddy 配置为每 30 秒检查 /health、超时 10 秒，恢复等待符合健康检查周期；没有取得其状态切换事件日志。
- 20:22:18 正式域名 /v1/responses 真实 gpt-5.5 medium 短文本完成，首个可见文本约 8.059 秒、总计 8.177 秒，收到 response.completed。请求标记 recovery-diag-8db5ea8850a54fc8；未输出凭据或用户内容。
- 20:22:54–20:23:35 连续五次本机及正式 health 为 200；容器内存约 905–1252 MiB、swap 约 6.2 MiB，memory/I/O PSI avg10 均约 0%。这只证明恢复窗口，不能证明长期容量充分。
- 20:26:12 再查：内存约 234 MiB、swap 约 6 MiB，memory PSI avg10=0、avg60=0.03%，限额命中数自恢复后未再增加；正式首页及 health 仍 200，宿主机可用约 2132 MiB。内存能释放支持请求峰值/驻留的解释，仍不足以排除其他泄漏。安全的资源前后记录保存在 /home/nio/incidents/20261012-router-503/2020-memory-recovery.json，不含凭据或完整 inspect。

## 首字延迟的定位

上游错误与本地资源压力同时存在，不能用其中一个解释全部请求。

1. 截图窗口 18:44:25–19:44:25 共 180 成功、50 失败；gpt-6.1-sol 40 条最终失败均关联上游 503，全部经历四次尝试。直接从生产宿主机调用也复现 server_is_overloaded。
2. 新启用账号 #38 确实接收请求；其 gpt-5.5 大上下文成功样本首字为 57–83 秒，包含约 160–166k 输入与缓存 token。近期原始请求体存在约 64–85 MiB 的样本；没有读取或输出其正文。两个账号共享 Router 的内存与出口，加账号不能解除这一层瓶颈。
3. 20:08 同一条约 231 字节 gpt-5.5 medium 短请求：直连 #38 首字 2.563 秒、直连 #44 为 2.060 秒、Router 本机 2.691 秒、正式域名 3.209 秒。这个小样本没有复现固定的几十秒网关开销。
4. 较大合成图片请求的成功推理对照未完成：上游拒绝或流中断，Router 遭重试后客户端 45 秒退出，不能将这些结果当成同等成功请求的性能基准。8 MiB 直连请求 curl 报上传完成约 10.5 秒，提示上传不可忽略；未确认云带宽上限。

## 代码中已定位的放大点

- OAuth/Codex 路径完整解码请求：openai_gateway_forward.go 中 ensureReqBody、UsesOpenAICodexProtocol 分支；修改后经 marshalOpenAIUpstreamJSON 再序列化。原始请求、解码对象、出站序列化缓冲会增加大请求的分配和驻留成本。
- ops_service.go 的 sanitizeAndTrimJSONPayload 先完整 json.Unmarshal、脱敏、json.Marshal，之后才检查快照大小上限。小日志快照并不保证处理过程占用小。
- 上游过载最多追加三次同账号重试，随后才尝试切换。大请求重发和长时间持有请求上下文会扩大容量风险。
- Forward 每次尝试重置计时，usage 的首字与 duration 通常不包括前面失败尝试；错误 routing_latency 又会累计此前上游耗时。看板值不能直接视为用户端到端等待或纯排队时间。

上述代码支持大请求与重试放大内存的解释；未取得线上 heap profile，不能宣称已证明内存泄漏或所有驻留对象的来源。

## 后续修复与边界

在线扩容恢复了入口可用性，未修复大请求保护、JSON 分配、重试预算及计时口径。恢复后实际用户成功样本仍有 10–23 秒首字，不能宣称首字延迟全部恢复。

优先实施按总在途字节保护大请求、在完整 JSON 解码前对超大错误快照降级、减少 OAuth 全量处理，并设置请求总重试预算；评估 Go 软内存上限时保留运行时外开销。环境变量变化或代码发布需受控替换容器，不把本次在线扩容视为已部署这些修复。不要自动减回 1 GiB，它已在真实流量下再次失去响应。

本次本地只更新运维文档；生产镜像与后端代码未变。另一个已有 incident worktree 的未提交文件保持原样。

## 补充：为什么下午扩到 1 GiB 后仍触顶

后续重新查询分钟指标、原始请求体积审计及请求完成日志；只输出体积、时间、模型及错误元数据。

| 时间，北京时间 | 直接证据 |
|---|---|
| 10 月 11 日 | Router 内存均值 145 MiB、P95 189 MiB、最大 233 MiB，容器限额为 512 MiB |
| 10 月 12 日 14:08 | 内存 213 MiB，限额 512 MiB |
| 14:08:50.153 | 收到 gpt-5.5 请求，原始 JSON 61878708 字节，约 59 MiB |
| 14:09 | 内存 511 MiB，触及原限额 |
| 14:27:54.816 | 又收到同样大小请求；第一条存在多次上游错误，长时间占用请求上下文 |
| 20:02:28–20:08:44 | 多次收到约 85 MiB 请求，分钟指标多次达到 1023 MiB / 1024 MiB |
| 20:20:57.760 | 20:08:44 的 85 MiB 请求结束，HTTP access 总耗时 746577 ms，约 12.44 分钟；逻辑失败 502，记录三个上游错误事件，HTTP 200 仅为 SSE 已提交 |
| 20:24:35.721 | 另一条 85 MiB 请求结束，总耗时 211128 ms，约 3.52 分钟 |
| 20:25–20:29 | 分钟内存依次回落到 328、199、202、205、195 MiB，Router 没有重启 |

并非所有历史时刻都低于 512 MiB：10 月 9 日也有 512 MiB 峰值；今天 14:08 之前同样有触顶采样。因此表中的 14:08 事件是下午明确对齐的一次突增，不能称为今天第一次大内存事件。

本地使用合成 59 MiB inline-image JSON 复现组件分配，通过 Go overlay 临时加载已有诊断 probe；未向生产发送大请求、未重放用户内容、未编辑其他 worktree。后端与生产部署提交无差异；本地 Go 1.27.0 / macOS arm64，不能替代 Linux 线上 heap profile。

- OAuth 完整解码、转换、序列化单次累计分配 378.8 MiB。
- 错误快照完整脱敏和编码单次累计分配 177.0 MiB，之后才裁剪为小快照。
- 连续组件调用观察到堆峰值 662.2 MiB、HeapSys 703.4 MiB。
- 释放请求并 GC 后 heap 为 197.4 MiB，其中测试仍有意持有 59 MiB 合成 body。
- 累计分配不能相加当作同时驻留；组件调用顺序也不是对生产全请求的完全模拟。该结果验证了分配放大风险，不给出线上每个请求的精确 RSS。

代码定位：getOpenAIRequestBodyMap 每次创建完整 map；OAuth Forward 分支调用该解码并重编码，重试重新进入 Forward；setOpsRequestBodySnapshot 将原始 body 保存在请求上下文；错误快照完整 JSON 处理在大小裁剪之前。生产未配置 GOMEMLIMIT，容器硬上限不能约束所有处理峰值。

综合判断：更大的请求、完整 JSON 分配及重试驻留是两次容量故障的主要触发链路；扩容只提高了可容纳峰值，没有改变每条请求的处理成本。请求结束后内存可回落到约 200 MiB，支持负载峰值的解释；尚不能完全排除其他驻留来源或泄漏。

## 上游比较与处理顺序

2026-10-12 20:35–20:41 只读查询 GitHub 并 fetch upstream refs/tags；未合并或切换运行版本。最新上游 main 为 9591136d45cef26453348ed9532a7ebbce7cf99c，最新发布为 v0.2.15（10 月 9 日）。

- 上游有类似症状报告：[内存约十个并发升至 4.5G，#7884](https://github.com/Wei-Shaw/sub2api/issues/7884)、[Codex 低并发下内存与 swap 快速增长，#4365](https://github.com/Wei-Shaw/sub2api/issues/4365)，查询时仍 open。这些用户报告不证明与本次完全同根因。
- [大 Responses 图片请求优化 #6783](https://github.com/Wei-Shaw/sub2api/pull/6783) 和 [大请求体保留优化 #2925](https://github.com/Wei-Shaw/sub2api/pull/2925) 已合并，生产源码已包含它们。后者明确保留 OAuth normalize 等复杂路径的完整解码；前者也注明剩余协议转换仍可能完整分配。最新 main/v0.2.15 的 OAuth Forward 仍调用 ensureReqBody。
- 更重要的 fork 差异：[上游 #2581](https://github.com/Wei-Shaw/sub2api/pull/2581) 在 5 月删除了 Ops 请求 replay/body 处理，避免错误日志随大 body 放大内存。我们的 fork 于 6 月 25 日通过 edfe0ea365c868007e190f99a1de559bd62f9e47 恢复错误诊断快照；该提交不在 upstream/main 中。PrepareOpsRequestBodyForQueue、setOpsRequestBodySnapshot、attachOpsRequestSnapshotToEntry 为本地恢复的路径，上游当前没有这些调用。
- 因此本地 probe 中快照额外分配 177 MiB 是我们 fork 的具体放大点，不能全部归因为原版上游。保留用于小错误正文的 sanitizeAndTrimJSONPayload 函数，并不代表上游还完整处理每个大请求正文。
- 单纯合并 v0.2.15 不会自动删除这段 fork 定制，也没有覆盖本次已复现的完整 OAuth 处理成本；不能把升级当作已经验证的修复。

建议优先在当前版本准备小范围 hotfix：

1. 错误诊断快照在完整解码之前检查请求体积。超大正文仅保留有界、脱敏的模型/体积/错误元数据；保持 RequestBodyBytes 和 truncated 标志，不为生成 256 KiB 快照先解析几十 MiB 全文。不能简单恢复异步队列持有完整原始 body。
2. 加总在途请求字节预算，超大请求限制并发；评估 Go 软内存目标并留出运行时/容器开销。软目标本身不能控制多个大请求的存活内存。
3. 上游过载按请求总预算重试，尽快切换经过验证的账号；普通与高推理请求分别设预算。减少重试可以缩短失败等待，不保证远端恢复。
4. 后续再优化 OAuth 路径的 raw patch/按需解码，保留工具调用、加密 reasoning、图片和指纹语义；不能通过直接删用户上下文来实现减内存。

验证需包含合成 59/85 MiB 请求、并发、重复过载恢复、正常长上下文与高推理请求；观察峰值及释放、health、客户端首字和总耗时。通过后再按受控发布流程部署。20:41:33 正式 health 仍 200 / 53 ms，内存约 208.5 MiB、swap 8 MiB、restart=0。本次上游调查未发布代码修复、未修改运行环境。

用户随后要求小范围规划。上面的四项是已识别的改进方向，具体执行范围以 [小范围修复计划](../plans/2026-10-12-router-memory-latency-hotfix.md) 为准：第一批只修快照入口的体积保护，验证后按证据单独处理过载重试或剩余容量风险，暂缓协议转换重写。
