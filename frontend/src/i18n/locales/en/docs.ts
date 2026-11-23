export default {
    title: 'Documentation',
    backToHome: 'Back to Home',
    copy: 'Copy',
    copied: 'Copied',
    nav: {
      quickStart: 'Quick Start',
      claudeCode: 'Claude Code',
      codexCli: 'Codex CLI',
      clients: 'Desktop Clients',
      opencode: 'OpenCode',
      apiUsage: 'API Usage',
      faq: 'Billing & FAQ'
    },
    shared: {
      copy: 'Copy',
      copied: 'Copied',
      optional: 'Optional',
      verify: 'Verify',
      viewFullGuide: 'View full guide',
      replaceKeyHint: 'Replace your-api-key-here with the API key from your dashboard.',
      baseUrlHint: 'Replace the base URL below with your actual site address.'
    },
    groups: {
      anthropicName: 'ClaudeCode专用',
      openaiName: 'OpenAI GPT (or the matching GPT monthly-plan group)',
      anthropicTools: 'Claude Code, Anthropic SDK, Claude in desktop clients',
      openaiTools: 'Codex CLI, Cursor, OpenCode, OpenAI SDK, GPT in desktop clients',
      colTool: 'Your tool',
      colGroup: 'Group for the API key',
      colBase: 'Base URL',
      anthropicHint: 'For Claude models: create the API key in the "ClaudeCode专用" group and use {base} (no /v1).',
      openaiKeyOnlyHint: 'For GPT models: create the API key in the "OpenAI GPT" group or a matching GPT monthly-plan group. Use the base URL from the generated config below.',
      openaiHint: 'For GPT models: create the API key in the "OpenAI GPT" group or a matching GPT monthly-plan group and use {base} (with /v1).'
    },
    clients: {
      title: 'Desktop Client Setup',
      subtitle: 'Cursor, Cherry Studio, Chatbox, LobeChat, Cline, Roo Code and other clients work as long as they accept a custom base URL and key. Below are the fields to fill in for each.',
      openaiKey: 'key from "OpenAI GPT" or a GPT monthly-plan group',
      anthropicKey: 'key from the "ClaudeCode专用" group',
      providerType: 'Provider type',
      customBaseUrl: 'enable "Use custom base URL"',
      cursorDesc: 'Cursor only supports the OpenAI-compatible mode, so only GPT models work here.',
      cursorNote: 'With Override OpenAI Base URL enabled, some built-in Cursor features (such as Tab completion) still use the official service and are unaffected.',
      cherryDesc: 'Add a new provider: choose the OpenAI type for GPT, or the Anthropic type for Claude. The two types need keys from different groups.',
      anthropicNote: 'With the Anthropic type, if the root URL returns 404, the client does not append /v1 itself; add /v1 to the end of the base URL.',
      clineDesc: 'Pick the API Provider in the extension settings: OpenAI Compatible for GPT, or Anthropic with a custom base URL for Claude.'
    },
    claudeCode: {
      title: 'Claude Code Setup',
      subtitle: 'Install Claude Code and configure it to connect to this platform. Supports both CLI and VSCode extension.',
      installTitle: 'Install Claude Code',
      installDescription: 'Choose the installation method for your system. Once installed, Claude Code is available directly in your terminal.',
      vscodeExtNote: 'For VSCode users, you can also search for "Claude Code" in the Extensions marketplace for an integrated IDE experience.',
      envTitle: 'Set Environment Variables',
      envDescription: 'Set the following environment variables in your terminal to point Claude Code at this platform. Add them to your shell profile (~/.zshrc or ~/.bashrc) for persistence.',
      envNote: 'Important: Do not include /v1/messages in the base URL — Claude Code appends the API path automatically.',
      settingsTitle: 'Configure settings.json (Optional)',
      settingsDescription: 'If you use the Claude Code extension in VSCode, you can also configure via ~/.claude/settings.json. Environment variables take precedence over settings.json.',
      verifyTitle: 'Verify Connection',
      verifyDescription: 'Run the following commands to check that Claude Code is connected to this platform. The /status command shows current auth status and API endpoint info.',
      troubleshootTitle: 'Troubleshooting',
      troubleshoot: {
        baseUrlQ: 'Wrong Base URL format?',
        baseUrlA: 'Make sure the URL does not include /v1/messages. Claude Code appends the API path automatically. Correct format example: https://api.example.com',
        restartQ: 'Config changes not taking effect?',
        restartA: 'Reopen your terminal after changing environment variables. Restart VSCode after editing settings.json. Environment variables take precedence over settings.json.',
        priorityQ: 'Both env vars and settings.json configured?',
        priorityA: 'Environment variables win. If both are set, env var values override settings.json. We recommend picking one method to avoid confusion.'
      }
    },
    opencode: {
      title: 'OpenCode Setup',
      subtitle: 'Install OpenCode and connect to this platform via an opencode.json config file.',
      installTitle: 'Install OpenCode',
      installDescription: 'Install the OpenCode CLI via npm, Homebrew, or curl.',
      configTitle: 'Create Config File',
      configDescription: 'Create opencode.json in your project root. The config includes provider info, model list, and API connection parameters.',
      configNote: 'Change the model field to the model you want to use (format: provider_id/model_id). This config uses the OpenAI-compatible protocol, so only GPT models work. To use Claude in OpenCode, add a separate {\'@\'}ai-sdk/anthropic provider with a key from the "ClaudeCode专用" group. Note that the baseURL of this provider must include /v1 (unlike Claude Code and the official Anthropic SDK).',
      verifyTitle: 'Verify & Start',
      verifyDescription: 'Check that the installation succeeded, then start OpenCode in your project directory.',
      tipsTitle: 'Configuration Tips',
      tips: {
        envVar: 'API keys support env var references: set apiKey to "{\'{\'}env:YOUR_API_KEY{\'}\'}" to read from environment variables instead of hardcoding secrets in config.',
        hierarchy: 'Config precedence: project opencode.json > OPENCODE_CONFIG env var > global ~/.config/opencode/opencode.json.',
        models: 'Run opencode /models to see the list of currently available models.'
      }
    },
    quickStart: {
      title: 'Quick Start',
      subtitle: 'Three steps: create a key in the right group for your tool, put the config into the tool, and send one test message. New users are taken to the onboarding guide automatically after signing in.',
      getKeyStep: 'Create an API key in the right group',
      getKeyDesc: 'Create a key on the dashboard API Keys page and pick the group from the table above. The group decides which models you can use; a wrong group shows a "model not found" error.',
      useKeyStep: 'Put the config into your tool',
      useKeyDesc: 'Click "Use Key" in the key list to see a config with your real base URL and key filled in. CC Switch users can click "Import to CCS". You can also follow the per-tool guides below.',
      testKeyStep: 'Send a test message',
      testKeyDesc: 'Ask anything in your tool. A normal reply means you are connected. For errors, see the error table in Billing & FAQ.',
      toolCards: {
        claudeCode: 'Quick setup via environment variables for Claude Code users.',
        codexCli: 'Full Codex CLI installation and configuration guide.',
        clients: 'Settings for Cursor, Cherry Studio, Cline and more.',
        opencode: 'Connect with a single JSON config file.',
        apiUsage: 'Call the API directly with Python, cURL, or Node.js.'
      }
    },
    faq: {
      title: 'Billing & FAQ',
      billingTitle: 'Billing',
      billing: {
        payg: 'Usage is deducted from your balance as you go. Per-request usage and cost are shown in the dashboard.',
        topup: 'When your balance runs low, top up in CNY via WeChat Pay on the dashboard top-up page.',
        monthly: 'Heavy GPT users can also buy a monthly plan. See the top-up page for available plans and prices.'
      },
      errorsTitle: 'Common errors',
      errors: {
        notFound: 'Wrong base URL. Claude tools use the base URL without /v1; GPT tools and the OpenAI SDK need /v1.',
        unauthorized: 'Wrong API key: check it was copied in full and has not been deleted or disabled.',
        forbidden: 'Insufficient balance or the key quota is used up. Top up and retry.',
        model: 'Wrong group: Claude models need a key from the "ClaudeCode专用" group; GPT models need a key from "OpenAI GPT" or a GPT monthly-plan group.'
      },
      contactTitle: 'Contact',
      contactDesc: 'Need help? Reach us at {contact} or email {email}.',
      contactFallback: 'WeChat: nanafox-hi',
      generalTitle: 'General',
      items: {
        compatibilityQ: 'Do I need to change the SDK or request format?',
        compatibilityA: 'Usually no. GPT uses the OpenAI protocol and Claude uses the Anthropic protocol; just swap the base URL and key.',
        multiToolQ: 'Can I use multiple tools at the same time?',
        multiToolA: 'Yes. Tools using the same model family can share a key. To use both Claude and GPT, create one key in each group.',
        modelsQ: 'Which models are supported?',
        modelsA: 'GPT models (such as gpt-6.1-sol) and Claude models (such as claude-sonnet-5-5). The key group decides which models you get: Claude is only in the "ClaudeCode专用" group over the Anthropic protocol; GPT is in "OpenAI GPT" and GPT monthly-plan groups over the OpenAI protocol.'
      }
    },
    mode: {
      api: 'Generic API Access',
      codex: 'Codex Access'
    },
    api: {
      title: 'Generic API Access',
      subtitle: 'For GPT, keep the same OpenAI SDK request shape and point the base URL at this site’s `/v1` endpoint. For Claude, use the Anthropic SDK; see the example below.',
      baseUrlLabel: 'Recommended base URL',
      examplesTitle: 'Request Examples',
      authTitle: 'Authentication',
      authDescription: 'OpenAI-protocol requests send the key in the `Authorization: Bearer` header; Anthropic-protocol requests send it in the `x-api-key` header. The official SDKs handle this once api_key is set.',
      endpointsTitle: 'Supported Endpoints',
      endpoints: {
        chat: 'OpenAI-compatible Chat Completions endpoint for direct drop-in replacement.',
        messages: 'Anthropic Messages-compatible endpoint for Claude models.',
        models: 'Lists currently available models so SDKs and frontends can render them dynamically.'
      },
      sdkTitle: 'Install SDK',
      sdkDescription: 'The OpenAI official SDK is the easiest way to call the API. Install the package for your language.',
      streamingTitle: 'Streaming Responses',
      streamingDescription: 'Enable stream mode to receive model output in real time — ideal for chat UIs and other scenarios that benefit from instant feedback.',
      anthropicTitle: 'Calling Claude (Anthropic protocol)',
      anthropicDescription: 'Install the official Anthropic SDK (pip install anthropic or npm install {\'@\'}anthropic-ai/sdk). Use the site root as the base URL; the SDK appends /v1/messages itself.',
      faqTitle: 'FAQ',
      quickStart: {
        registerTitle: 'Create an Account',
        registerDesc: 'Sign up and log in first. The public page stays minimal; actual key management happens in the dashboard.',
        keyTitle: 'Create an API Key',
        keyDesc: 'Create a new API key in the dashboard. That key is then reused across SDKs, scripts, and local tools.',
        baseUrlTitle: 'Replace the Base URL',
        baseUrlDesc: 'Point your existing OpenAI client at this site’s `/v1` endpoint. In most cases the rest of the request body stays the same.'
      },
      faq: {
        items: {
          compatibilityQ: 'Do I need to change the SDK or payload format?',
          compatibilityA: 'Usually no. If your client already speaks the OpenAI protocol, swapping `base_url` and `api_key` is enough.',
          sdkQ: 'Which languages and frameworks are supported?',
          sdkA: 'Python, Node.js, Go, Java, and any OpenAI-compatible SDK can connect directly.',
          dataQ: 'Does the platform store my conversation content?',
          dataA: 'The public docs focus on routing and metering behavior. If your deployment enables extra audit or logging policy, follow your actual site settings.'
        }
      }
    },
    codex: {
      title: 'Codex CLI Access',
      label: 'Codex Access',
      subtitle: 'Follow the standard setup flow for Node.js, Codex CLI, API key, and config files, then start coding immediately.',
      os: {
        mac: 'macOS',
        windows: 'Windows',
        linux: 'Linux'
      },
      nodeTitle: 'Install Node.js',
      nodeDescription: 'Codex CLI requires Node.js. Install the current LTS release before installing the CLI itself.',
      nodeGuideMac: 'Install the current Node.js LTS release with Homebrew before moving on to Codex CLI.',
      nodeGuideWindows: 'Install Node.js with Chocolatey or Scoop, then reopen your terminal before continuing.',
      nodeGuideLinux: 'Install the current Node.js LTS release from NodeSource before moving on to Codex CLI.',
      nodeVerifyTitle: 'Verify versions',
      installTitle: 'Install Codex CLI',
      installDescription: 'Install the Codex CLI package globally. If npm is slow in your region, you can keep the mirror registry flag.',
      installVerifyTitle: 'Verify CLI',
      keyTitle: 'Prepare an API Key',
      keyDescription: 'Create an API key in the dashboard first. Codex CLI will read that key through environment-based auth.',
      keyHint: 'Replace the placeholder value here with your own API key. The in-product usage modal will inject the real value automatically.',
      configureTitle: 'Write the config files',
      configureDescription: 'Place the two files below inside your local `.codex` directory. Standard mode is the default path and should be used first.',
      startTitle: 'Start Codex',
      startDescription: 'Enter your project directory and run `codex`. The CLI will load the config files and route requests through this site’s OpenAI-compatible endpoint.',
      vscodeTitle: 'Use it inside your editor',
      vscodeDescription: 'If you run Codex CLI from VS Code or other editors, reuse the same `.codex` files instead of maintaining a second config.',
      advancedTitle: 'Optional WebSocket Mode',
      advancedDescription: 'Only add the snippet below when you explicitly need a WebSocket response path on top of the standard setup.',
      advancedSnippetTitle: 'Append to config.toml',
      troubleshootTitle: 'Troubleshooting',
      troubleshoot: {
        items: {
          commandTitle: 'Verify the CLI is installed',
          commandDesc: 'Run `node --version`, `npm --version`, and `codex --version` first to confirm the environment and CLI are actually available.',
          connectionTitle: 'Verify the base URL',
          connectionDesc: 'If the CLI cannot connect, first check that base_url in config.toml is the site address, the API key is still valid, and the key belongs to a GPT group.',
          updateTitle: 'Upgrade the CLI and retry',
          updateDesc: 'If the command runs but behaves incorrectly, upgrade the Codex CLI package first and reload the config before testing again.'
        }
      }
    }
}
