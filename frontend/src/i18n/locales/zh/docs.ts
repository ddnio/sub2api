export default {
    title: '文档',
    signupCta: {
      title: '还没有账号？注册就能照着文档马上用',
      titleWithBonus: '注册即送 {amount} 体验额度，照着文档马上用',
      desc: '一个 API Key 接入 Claude Code、Codex、Cursor 等客户端，微信充值，¥1 = $1。',
      button: '免费注册',
      buttonWithBonus: '注册，领 {amount}'
    },
    backToHome: '返回首页',
    copy: '复制',
    copied: '已复制',
    nav: {
      quickStart: '快速开始',
      claudeCode: 'Claude Code',
      codexCli: 'Codex CLI',
      clients: '图形客户端',
      opencode: 'OpenCode',
      apiUsage: 'API 调用',
      faq: '计费与常见问题'
    },
    seo: {
      quickStart: '三步接入 NanaFox Router：按工具选对分组创建 API Key，填进 Claude Code、Codex、Cursor 等工具，发一句话测试。',
      claudeCode: 'Claude Code 接入教程：安装、配置 ANTHROPIC_BASE_URL 和 API Key、验证连通，地址不加 /v1。',
      codexCli: 'Codex CLI 接入教程：安装 Node.js 和 Codex CLI，写入 config.toml 与 auth.json，使用 GPT 模型。',
      clients: 'Cursor、Cherry Studio、Chatbox、LobeChat、Cline、Roo Code 接入填法，GPT 与 Claude 分别配置。',
      opencode: 'OpenCode 接入教程：opencode.json 配置示例，GPT 走 OpenAI 兼容，Claude 走 Anthropic provider。',
      apiUsage: '用 OpenAI SDK 调用 GPT、用 Anthropic SDK 调用 Claude 的 Python、cURL、Node.js 示例。',
      faq: '计费方式、微信充值、月卡说明，以及 401、403、404、模型不存在等常见报错的排查方法。'
    },
    shared: {
      copy: '复制',
      copied: '已复制',
      optional: '可选',
      verify: '验证',
      viewFullGuide: '查看完整教程',
      replaceKeyHint: '将 your-api-key-here 替换为你在控制台创建的 API Key。',
      baseUrlHint: '将下方的基地址替换为实际的站点地址。'
    },
    groups: {
      anthropicName: 'ClaudeCode专用',
      openaiName: 'OpenAI GPT（或对应的 GPT 月卡分组）',
      anthropicTools: 'Claude Code、Anthropic SDK、客户端里的 Claude',
      openaiTools: 'Codex CLI、Cursor、OpenCode、OpenAI SDK、客户端里的 GPT',
      colTool: '你用的工具',
      colGroup: '创建 Key 时选的分组',
      colBase: '填写的地址',
      anthropicHint: '用 Claude 模型：创建 API Key 时选「ClaudeCode专用」分组，地址填 {base}（不要加 /v1）。',
      openaiKeyOnlyHint: '用 GPT 模型：创建 API Key 时选「OpenAI GPT」或对应的 GPT 月卡分组。地址照下面生成的配置填即可。',
      openaiHint: '用 GPT 模型：创建 API Key 时选「OpenAI GPT」或对应的 GPT 月卡分组，地址填 {base}（要带 /v1）。'
    },
    clients: {
      title: '图形客户端接入',
      subtitle: 'Cursor、Cherry Studio、Chatbox、LobeChat、Cline、Roo Code 等客户端，只要能填自定义地址和 Key 就能用。下面是各客户端里要填的字段。',
      openaiKey: '「OpenAI GPT」或 GPT 月卡分组的 Key',
      anthropicKey: '「ClaudeCode专用」分组的 Key',
      providerType: '服务商类型',
      customBaseUrl: '勾选「使用自定义 Base URL」',
      cursorDesc: 'Cursor 只支持 OpenAI 兼容方式，所以这里只能用 GPT 模型。',
      cursorNote: '打开 Override OpenAI Base URL 后，Cursor 自带的部分功能（如 Tab 补全）仍走官方，不受影响。',
      cherryDesc: '新建一个服务商：用 GPT 选 OpenAI 类型，用 Claude 选 Anthropic 类型。两种类型要用不同分组的 Key。',
      anthropicNote: '用 Anthropic 类型时，如果填根地址后报 404，说明这个客户端不会自己拼 /v1，在地址后面加上 /v1 即可。',
      clineDesc: '在插件设置里选择 API Provider：GPT 用 OpenAI Compatible，Claude 用 Anthropic 并填自定义地址。'
    },
    claudeCode: {
      title: 'Claude Code 接入',
      subtitle: '安装 Claude Code 并配置连接本平台，支持 CLI 和 VSCode 扩展两种方式。',
      installTitle: '安装 Claude Code',
      installDescription: '选择适合你系统的方式安装 Claude Code CLI。安装完成后可直接在终端使用。',
      vscodeExtNote: '如果使用 VSCode，也可以在扩展商店搜索 "Claude Code" 安装扩展，获得 IDE 内的集成体验。',
      envTitle: '配置环境变量',
      envDescription: '在终端中设置以下环境变量，将 Claude Code 指向本平台。可添加到 shell 配置文件（~/.zshrc 或 ~/.bashrc）中永久生效。',
      envNote: '注意：Base URL 不要包含 /v1/messages 路径后缀，Claude Code 会自动拼接。',
      settingsTitle: '配置 settings.json（可选）',
      settingsDescription: '如果你使用 VSCode 中的 Claude Code 扩展，也可以通过 ~/.claude/settings.json 配置。环境变量优先级高于 settings.json。',
      verifyTitle: '验证连通',
      verifyDescription: '运行以下命令检查 Claude Code 是否正确连接到本平台。/status 命令会显示当前的认证状态和 API 端点信息。',
      troubleshootTitle: '常见问题排查',
      troubleshoot: {
        baseUrlQ: 'Base URL 格式不对？',
        baseUrlA: '确保 URL 不包含 /v1/messages 后缀。Claude Code 会自动拼接 API 路径。正确格式示例：https://api.example.com',
        restartQ: '修改配置后不生效？',
        restartA: '环境变量修改后需要重新打开终端。如果修改了 settings.json，需要重启 VSCode。环境变量的优先级高于 settings.json。',
        priorityQ: '同时设置了环境变量和 settings.json？',
        priorityA: '环境变量优先。如果两处都配置了，环境变量的值会覆盖 settings.json 中的设置。建议只选择一种方式配置，避免混淆。'
      }
    },
    opencode: {
      title: 'OpenCode 接入',
      subtitle: '安装 OpenCode 并通过 opencode.json 配置文件连接本平台。',
      installTitle: '安装 OpenCode',
      installDescription: '通过 npm、Homebrew 或 curl 安装 OpenCode CLI。',
      configTitle: '创建配置文件',
      configDescription: '在项目根目录创建 opencode.json 文件。配置中包含 provider 信息、模型列表和 API 连接参数。',
      configNote: '将 model 字段改为你需要使用的模型（格式：provider_id/model_id）。这份配置走 OpenAI 兼容协议，只能用 GPT 模型；要在 OpenCode 里用 Claude，需另加一个 {\'@\'}ai-sdk/anthropic 类型的 provider，并使用「ClaudeCode专用」分组的 Key。注意这个 provider 的 baseURL 要带 /v1（和 Claude Code、Anthropic 官方 SDK 不同）。',
      verifyTitle: '验证并启动',
      verifyDescription: '检查安装是否成功，然后在项目目录下启动 OpenCode。',
      tipsTitle: '配置提示',
      tips: {
        envVar: 'API Key 支持环境变量引用：将 apiKey 值设为 "{\'{\'}env:YOUR_API_KEY{\'}\'}" 可从环境变量读取，避免在配置文件中硬编码密钥。',
        hierarchy: '配置优先级：项目级 opencode.json > 环境变量 OPENCODE_CONFIG > 全局 ~/.config/opencode/opencode.json。',
        models: '使用 opencode /models 命令可查看当前可用的模型列表。'
      }
    },
    quickStart: {
      title: '快速开始',
      subtitle: '三步接入：按你用的工具选对分组建 Key，把配置填进工具，发一句话测试。新注册用户登录后会自动进入新手引导，按提示走完即可。',
      getKeyStep: '按工具选分组，创建 API Key',
      getKeyDesc: '在控制台「API 密钥」页面新建 Key，分组按上表选。分组决定能用哪些模型，选错会提示模型不存在。',
      useKeyStep: '把配置填进工具',
      useKeyDesc: '在 Key 列表点「使用密钥」可以看到填好真实地址和 Key 的配置；用 CC Switch 的可以点「导入到 CCS」一键导入。也可以看下方各工具的教程。',
      testKeyStep: '发一句话测试',
      testKeyDesc: '在工具里随便问一句，能正常回复就说明接通了。报错可以对照「计费与常见问题」里的错误码排查。',
      toolCards: {
        claudeCode: '通过环境变量快速接入，适合 Claude Code 用户。',
        codexCli: '完整的 Codex CLI 安装与配置教程。',
        clients: 'Cursor、Cherry Studio、Cline 等客户端的填法。',
        opencode: '一个 JSON 配置文件即可接入。',
        apiUsage: '直接调用 API，支持 Python、cURL、Node.js。'
      }
    },
    faq: {
      title: '计费与常见问题',
      billingTitle: '怎么计费',
      billing: {
        payg: '按实际用量从账户余额扣费，每次调用的用量和费用可以在控制台查看。',
        topup: '余额不足时，在控制台「充值」页面用微信支付充值人民币。',
        monthly: '常用 GPT 的也可以购买月卡，可选套餐和价格以充值页面显示为准。'
      },
      errorsTitle: '报错怎么办',
      errors: {
        notFound: '地址写错了。Claude 系工具地址不加 /v1，GPT 系工具和 OpenAI SDK 地址要带 /v1。',
        unauthorized: 'API Key 不对：检查是否复制完整、是否已被删除或停用。',
        forbidden: '余额不足或 Key 额度用完，充值后重试。',
        model: '分组选错了：Claude 模型要用「ClaudeCode专用」分组的 Key，GPT 模型要用「OpenAI GPT」或 GPT 月卡分组的 Key。'
      },
      contactTitle: '联系我们',
      contactDesc: '遇到问题可以联系 {contact}，或发邮件到 {email}。',
      contactFallback: '微信：nanafox-hi',
      generalTitle: '通用问题',
      items: {
        compatibilityQ: '需要修改 SDK 或请求结构吗？',
        compatibilityA: '通常不需要。GPT 走 OpenAI 协议，Claude 走 Anthropic 协议，替换地址和 Key 即可。',
        multiToolQ: '可以同时使用多个工具吗？',
        multiToolA: '可以。同一类模型的工具可以共用一个 Key；同时用 Claude 和 GPT 时，分别在两个分组各建一个 Key。',
        modelsQ: '支持哪些模型？',
        modelsA: 'GPT 系列（如 gpt-6.1-sol）和 Claude 系列（如 claude-sonnet-5-5）。能用哪些模型由 Key 所在分组决定：Claude 只在「ClaudeCode专用」分组，走 Anthropic 协议；GPT 在「OpenAI GPT」和 GPT 月卡分组，走 OpenAI 协议。'
      }
    },
    mode: {
      api: '通用 API 接入',
      codex: 'Codex 接入'
    },
    api: {
      title: '通用 API 接入',
      subtitle: 'GPT 模型保持 OpenAI SDK 的调用方式不变，把基地址换成当前站点的 `/v1` 端点即可；Claude 模型用 Anthropic SDK，示例见下方。',
      baseUrlLabel: '推荐基地址',
      examplesTitle: '示例请求',
      authTitle: '认证方式',
      authDescription: 'OpenAI 协议的请求把 Key 放进 `Authorization: Bearer` 请求头；Anthropic 协议的请求把 Key 放进 `x-api-key` 请求头。官方 SDK 会自动处理，只要填好 api_key 即可。',
      endpointsTitle: '支持的端点',
      endpoints: {
        chat: '兼容 OpenAI Chat Completions 协议，可直接替换现有聊天请求。',
        messages: '兼容 Anthropic Messages 协议，用于 Claude 模型。',
        models: '返回当前可用模型列表，便于 SDK 或前端动态展示。'
      },
      sdkTitle: '安装 SDK',
      sdkDescription: '使用 OpenAI 官方 SDK 可以最方便地调用 API。选择你使用的语言安装对应的包。',
      streamingTitle: '流式响应',
      streamingDescription: '启用 stream 模式可以实时接收模型的输出，适合聊天界面等需要即时反馈的场景。',
      anthropicTitle: '调用 Claude（Anthropic 协议）',
      anthropicDescription: '安装 Anthropic 官方 SDK（pip install anthropic 或 npm install {\'@\'}anthropic-ai/sdk），地址填站点根地址，SDK 会自动拼上 /v1/messages。',
      faqTitle: '常见问题',
      quickStart: {
        registerTitle: '创建账号',
        registerDesc: '先注册并登录站点，公开页保持极简，真正的密钥管理在控制台完成。',
        keyTitle: '创建 API Key',
        keyDesc: '进入控制台创建一个新的 API Key，后续所有 SDK、脚本与客户端都使用这把 Key。',
        baseUrlTitle: '替换 Base URL',
        baseUrlDesc: '把原来指向 OpenAI 的 `base_url` 改成当前站点的 `/v1` 地址，其余请求参数基本保持不变。'
      },
      faq: {
        items: {
          compatibilityQ: '需要改 SDK 或请求结构吗？',
          compatibilityA: '通常不需要。只要你的客户端本身兼容 OpenAI 协议，替换 `base_url` 和 `api_key` 即可。',
          sdkQ: '支持哪些语言和框架？',
          sdkA: 'Python、Node.js、Go、Java 以及任何 OpenAI 兼容 SDK 都可以直接接入。',
          dataQ: '平台会保存我的对话内容吗？',
          dataA: '平台重点提供转发与计量能力，默认文档不承诺保存内容；如有自定义审计或日志策略，请以你的站点实际配置为准。'
        }
      }
    },
    codex: {
      title: 'Codex CLI 接入',
      label: 'Codex 接入',
      subtitle: '按标准流程完成 Node.js、Codex CLI、API Key 和配置文件后，即可直接开始使用。',
      os: {
        mac: 'macOS',
        windows: 'Windows',
        linux: 'Linux'
      },
      nodeTitle: '安装 Node.js',
      nodeDescription: 'Codex CLI 依赖 Node.js 运行环境。先安装 LTS 版本，再继续安装命令行工具。',
      nodeGuideMac: '推荐先通过 Homebrew 安装 Node.js LTS，再继续安装 Codex CLI。',
      nodeGuideWindows: '推荐使用 Chocolatey 或 Scoop 安装 Node.js，安装完成后重新打开终端。',
      nodeGuideLinux: '推荐先安装 NodeSource 提供的 Node.js LTS 版本，再继续安装 Codex CLI。',
      nodeVerifyTitle: '检查版本',
      installTitle: '安装 Codex CLI',
      installDescription: '全局安装 Codex CLI 包。如果你的环境访问 npm 较慢，可以保留镜像源参数。',
      installVerifyTitle: '验证 CLI',
      keyTitle: '准备 API Key',
      keyDescription: '先在控制台创建 API Key，后续 Codex CLI 会通过环境变量读取这把 Key。',
      keyHint: '把这里的占位值替换成你自己的 API Key。控制台内的“如何使用”弹窗会自动写入真实值。',
      configureTitle: '写入配置文件',
      configureDescription: '把下面两个文件放到本机 `.codex` 目录中。标准模式是默认接入方式，直接复用即可。',
      startTitle: '启动 Codex',
      startDescription: '进入你的项目目录后执行 `codex`。CLI 会读取配置文件并通过当前站点的 OpenAI 兼容端点发起请求。',
      vscodeTitle: '在编辑器中使用',
      vscodeDescription: '如果你在 VS Code 等编辑器里调用 Codex CLI，也复用同一份 `.codex` 配置文件，无需再维护第二套接入参数。',
      advancedTitle: '可选 WebSocket 模式',
      advancedDescription: '只有在你明确需要 WebSocket 响应链路时，才在标准配置基础上追加下面这段配置。',
      advancedSnippetTitle: '追加到 config.toml',
      troubleshootTitle: '排查建议',
      troubleshoot: {
        items: {
          commandTitle: '先确认 CLI 已安装',
          commandDesc: '先执行 `node --version`、`npm --version` 和 `codex --version`，确保环境与 CLI 都已经正确安装。',
          connectionTitle: '确认基地址可用',
          connectionDesc: '如果 CLI 无法连接，先检查 config.toml 里的 base_url 是否为站点地址、API Key 是否仍然有效，以及 Key 是否建在 GPT 分组。',
          updateTitle: '升级 CLI 后重试',
          updateDesc: '如果命令可执行但行为异常，先升级 Codex CLI 包到最新版本，再重新读取配置验证。'
        }
      }
    }
}
