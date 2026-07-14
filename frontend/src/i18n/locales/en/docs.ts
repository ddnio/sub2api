export default {
    title: 'Documentation',
    backToHome: 'Back to Home',
    copy: 'Copy',
    copied: 'Copied',
    nav: {
      quickStart: 'Quick Start',
      claudeCode: 'Claude Code',
      codexCli: 'Codex CLI',
      opencode: 'OpenCode',
      apiUsage: 'API Usage',
      faq: 'FAQ'
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
      configNote: 'Change the model field to the model you want to use (format: provider_id/model_id). List available model names in the models object.',
      verifyTitle: 'Verify & Start',
      verifyDescription: 'Check that the installation succeeded, then start OpenCode in your project directory.',
      tipsTitle: 'Configuration Tips',
      tips: {
        envVar: 'API keys support env var references: set apiKey to "{env:YOUR_API_KEY}" to read from environment variables instead of hardcoding secrets in config.',
        hierarchy: 'Config precedence: project opencode.json > OPENCODE_CONFIG env var > global ~/.config/opencode/opencode.json.',
        models: 'Run opencode /models to see the list of currently available models.'
      }
    },
    quickStart: {
      title: 'Quick Start',
      subtitle: 'Choose your tool below and follow the setup guide to connect.',
      baseUrlLabel: 'Platform Base URL',
      getKeyStep: 'Create an API Key',
      getKeyDesc: 'Sign up, log in, and create a new API key in the dashboard.',
      chooseToolStep: 'Choose Your Tool',
      chooseToolDesc: 'Select the tool you use below to see the setup instructions.',
      toolCards: {
        claudeCode: 'Quick setup via environment variables for Claude Code users.',
        codexCli: 'Full Codex CLI installation and configuration guide.',
        opencode: 'Connect with a single JSON config file.',
        apiUsage: 'Call the API directly with Python, cURL, or Node.js.'
      }
    },
    faq: {
      title: 'FAQ',
      generalTitle: 'General',
      items: {
        compatibilityQ: 'Do I need to change the SDK or request format?',
        compatibilityA: 'Usually no. This platform is OpenAI-compatible — just swap base_url and api_key.',
        multiToolQ: 'Can I use multiple tools at the same time?',
        multiToolA: 'Yes. Each tool is configured independently and shares the same API key.',
        modelsQ: 'Which models are supported?',
        modelsA: 'GPT and Claude full model lineup. Check the dashboard for currently available models.'
      }
    },
    mode: {
      api: 'Generic API Access',
      codex: 'Codex Access'
    },
    api: {
      title: 'Generic API Access',
      subtitle: 'Keep the same OpenAI SDK request shape. Replace the base URL with this site’s `/v1` endpoint and use your generated API key.',
      baseUrlLabel: 'Recommended base URL',
      examplesTitle: 'Request Examples',
      authTitle: 'Authentication',
      authDescription: 'All requests use standard Bearer token auth. Put the API key from your dashboard into the `Authorization` header.',
      endpointsTitle: 'Supported Endpoints',
      endpoints: {
        chat: 'OpenAI-compatible Chat Completions endpoint for direct drop-in replacement.',
        models: 'Lists currently available models so SDKs and frontends can render them dynamically.'
      },
      sdkTitle: 'Install SDK',
      sdkDescription: 'The OpenAI official SDK is the easiest way to call the API. Install the package for your language.',
      streamingTitle: 'Streaming Responses',
      streamingDescription: 'Enable stream mode to receive model output in real time — ideal for chat UIs and other scenarios that benefit from instant feedback.',
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
          connectionDesc: 'If the CLI cannot connect, first confirm the base URL is correct, includes `/v1`, and the API key is still valid.',
          updateTitle: 'Upgrade the CLI and retry',
          updateDesc: 'If the command runs but behaves incorrectly, upgrade the Codex CLI package first and reload the config before testing again.'
        }
      }
    }
}
