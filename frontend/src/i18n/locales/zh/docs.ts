export default {
    title: '文档',
    backToHome: '返回首页',
    copy: '复制',
    copied: '已复制',
    nav: {
      quickStart: '快速开始',
      claudeCode: 'Claude Code',
      codexCli: 'Codex CLI',
      opencode: 'OpenCode',
      apiUsage: 'API 调用',
      faq: '常见问题'
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
      configNote: '将 model 字段改为你需要使用的模型（格式：provider_id/model_id）。models 对象中列出可用的模型名称。',
      verifyTitle: '验证并启动',
      verifyDescription: '检查安装是否成功，然后在项目目录下启动 OpenCode。',
      tipsTitle: '配置提示',
      tips: {
        envVar: 'API Key 支持环境变量引用：将 apiKey 值设为 "{env:YOUR_API_KEY}" 可从环境变量读取，避免在配置文件中硬编码密钥。',
        hierarchy: '配置优先级：项目级 opencode.json > 环境变量 OPENCODE_CONFIG > 全局 ~/.config/opencode/opencode.json。',
        models: '使用 opencode /models 命令可查看当前可用的模型列表。'
      }
    },
    quickStart: {
      title: '快速开始',
      subtitle: '选择你使用的工具，按教程完成配置即可接入。',
      baseUrlLabel: '平台基地址',
      getKeyStep: '在控制台创建 API Key',
      getKeyDesc: '注册登录后，在控制台中创建一个新的 API Key。',
      chooseToolStep: '选择接入工具',
      chooseToolDesc: '根据你的使用场景，选择下方对应的工具查看接入教程。',
      toolCards: {
        claudeCode: '通过环境变量快速接入，适合 Claude Code 用户。',
        codexCli: '完整的 Codex CLI 安装与配置教程。',
        opencode: '一个 JSON 配置文件即可接入。',
        apiUsage: '直接调用 API，支持 Python、cURL、Node.js。'
      }
    },
    faq: {
      title: '常见问题',
      generalTitle: '通用问题',
      items: {
        compatibilityQ: '需要修改 SDK 或请求结构吗？',
        compatibilityA: '通常不需要。本平台兼容 OpenAI 协议，替换 base_url 和 api_key 即可。',
        multiToolQ: '可以同时使用多个工具吗？',
        multiToolA: '可以。每个工具独立配置，共用同一个 API Key。',
        modelsQ: '支持哪些模型？',
        modelsA: '支持 GPT、Claude 全系列模型，具体可用模型以控制台显示为准。'
      }
    },
    mode: {
      api: '通用 API 接入',
      codex: 'Codex 接入'
    },
    api: {
      title: '通用 API 接入',
      subtitle: '保持 OpenAI SDK 的调用方式不变，只需要把基地址替换为当前站点的 `/v1` 端点并使用你创建的 API Key。',
      baseUrlLabel: '推荐基地址',
      examplesTitle: '示例请求',
      authTitle: '认证方式',
      authDescription: '所有请求都使用标准 Bearer Token 鉴权，请把控制台生成的 Key 放入 `Authorization` 请求头。',
      endpointsTitle: '支持的端点',
      endpoints: {
        chat: '兼容 OpenAI Chat Completions 协议，可直接替换现有聊天请求。',
        models: '返回当前可用模型列表，便于 SDK 或前端动态展示。'
      },
      sdkTitle: '安装 SDK',
      sdkDescription: '使用 OpenAI 官方 SDK 可以最方便地调用 API。选择你使用的语言安装对应的包。',
      streamingTitle: '流式响应',
      streamingDescription: '启用 stream 模式可以实时接收模型的输出，适合聊天界面等需要即时反馈的场景。',
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
          connectionDesc: '如果 CLI 无法连接，优先检查站点基地址是否正确、是否补上了 `/v1`，以及 API Key 是否仍然有效。',
          updateTitle: '升级 CLI 后重试',
          updateDesc: '如果命令可执行但行为异常，先升级 Codex CLI 包到最新版本，再重新读取配置验证。'
        }
      }
    }
}
