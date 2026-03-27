export type ShellType = 'unix' | 'cmd' | 'powershell'
export type ApiLanguage = 'python' | 'curl' | 'nodejs'

export interface SnippetFile {
  path: string
  content: string
  hint?: string
}

// ---------- Claude Code ----------

export function generateClaudeCodeEnvSnippet(
  baseUrl: string,
  apiKey: string,
  shell: ShellType
): SnippetFile {
  const base = baseUrl.replace(/\/+$/, '')
  switch (shell) {
    case 'unix':
      return {
        path: 'Terminal',
        content: `export ANTHROPIC_BASE_URL="${base}"\nexport ANTHROPIC_AUTH_TOKEN="${apiKey}"`
      }
    case 'cmd':
      return {
        path: 'Command Prompt',
        content: `set ANTHROPIC_BASE_URL=${base}\nset ANTHROPIC_AUTH_TOKEN=${apiKey}`
      }
    case 'powershell':
      return {
        path: 'PowerShell',
        content: `$env:ANTHROPIC_BASE_URL="${base}"\n$env:ANTHROPIC_AUTH_TOKEN="${apiKey}"`
      }
  }
}

export function generateClaudeCodeSettingsSnippet(
  baseUrl: string,
  apiKey: string,
  shell: ShellType
): SnippetFile {
  const base = baseUrl.replace(/\/+$/, '')
  const path = shell === 'unix'
    ? '~/.claude/settings.json'
    : '%userprofile%\\.claude\\settings.json'
  const content = JSON.stringify({
    env: {
      ANTHROPIC_BASE_URL: base,
      ANTHROPIC_AUTH_TOKEN: apiKey,
      CLAUDE_CODE_ATTRIBUTION_HEADER: '0'
    }
  }, null, 2)
  return { path, content, hint: 'VSCode Claude Code' }
}

// ---------- OpenCode ----------

export function generateOpenCodeSnippet(
  platform: string,
  baseUrl: string,
  apiKey: string,
  pathLabel?: string
): SnippetFile {
  const provider: Record<string, any> = {
    [platform]: {
      options: { baseURL: baseUrl, apiKey }
    }
  }

  if (platform === 'gemini' || platform === 'antigravity-gemini') {
    provider[platform].npm = '@ai-sdk/google'
  } else if (platform === 'anthropic' || platform === 'antigravity-claude') {
    provider[platform].npm = '@ai-sdk/anthropic'
  }

  const content = JSON.stringify(
    { provider, $schema: 'https://opencode.ai/config.json' },
    null,
    2
  )

  return {
    path: pathLabel ?? 'opencode.json',
    content,
    hint: undefined
  }
}

// ---------- Claude Code Install ----------

export function generateClaudeCodeInstallSnippet(shell: ShellType): SnippetFile {
  switch (shell) {
    case 'unix':
      return {
        path: 'Terminal',
        content: `# Install Claude Code CLI
curl -fsSL https://claude.ai/install.sh | bash

# Or via Homebrew
brew install --cask claude-code`
      }
    case 'cmd':
      return {
        path: 'Command Prompt',
        content: `@rem Install Claude Code CLI
curl -fsSL https://claude.ai/install.cmd -o install.cmd && install.cmd && del install.cmd`
      }
    case 'powershell':
      return {
        path: 'PowerShell',
        content: `# Install Claude Code CLI
irm https://claude.ai/install.ps1 | iex

# Or via WinGet
winget install Anthropic.ClaudeCode`
      }
  }
}

// ---------- OpenCode Install ----------

export function generateOpenCodeInstallSnippet(shell: ShellType): SnippetFile {
  switch (shell) {
    case 'unix':
      return {
        path: 'Terminal',
        content: `# Via npm (recommended)
npm install -g opencode-ai@latest

# Or via Homebrew
brew install anomalyco/tap/opencode

# Or via curl
curl -fsSL https://opencode.ai/install | bash`
      }
    case 'cmd':
      return {
        path: 'Command Prompt',
        content: `@rem Via npm (recommended)
npm install -g opencode-ai@latest`
      }
    case 'powershell':
      return {
        path: 'PowerShell',
        content: `# Via npm (recommended)
npm install -g opencode-ai@latest

# Or via WinGet
winget install opencode`
      }
  }
}

// ---------- OpenCode Enhanced Config ----------

export function generateOpenCodeEnhancedSnippet(
  baseUrl: string,
  apiKey: string
): SnippetFile {
  const config = {
    $schema: 'https://opencode.ai/config.json',
    model: 'openai/gpt-4o',
    provider: {
      openai: {
        npm: '@ai-sdk/openai-compatible',
        name: 'Custom Provider',
        options: {
          baseURL: baseUrl,
          apiKey: apiKey
        },
        models: {
          'gpt-4o': { name: 'GPT-4o' },
          'claude-sonnet-4-6': { name: 'Claude Sonnet 4.6' }
        }
      }
    }
  }
  return {
    path: 'opencode.json',
    content: JSON.stringify(config, null, 2)
  }
}

// ---------- API Streaming Examples ----------

export function generateApiStreamingExample(
  baseUrl: string,
  apiKey: string,
  language: ApiLanguage
): SnippetFile {
  const base = baseUrl.replace(/\/+$/, '')
  const baseV1 = base.endsWith('/v1') ? base : `${base}/v1`

  switch (language) {
    case 'python':
      return {
        path: 'Python (Streaming)',
        content: `from openai import OpenAI

client = OpenAI(
    base_url="${baseV1}",
    api_key="${apiKey}",
)

stream = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello"}],
    stream=True,
)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="", flush=True)`
      }
    case 'curl':
      return {
        path: 'cURL (Streaming)',
        content: `curl ${baseV1}/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer ${apiKey}" \\
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'`
      }
    case 'nodejs':
      return {
        path: 'Node.js (Streaming)',
        content: `import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "${baseV1}",
  apiKey: "${apiKey}",
});

const stream = await client.chat.completions.create({
  model: "gpt-4o",
  messages: [{ role: "user", content: "Hello" }],
  stream: true,
});
for await (const chunk of stream) {
  const content = chunk.choices[0]?.delta?.content;
  if (content) process.stdout.write(content);
}`
      }
  }
}

// ---------- API Examples ----------

export function generateApiExample(
  baseUrl: string,
  apiKey: string,
  language: ApiLanguage
): SnippetFile {
  const base = baseUrl.replace(/\/+$/, '')
  const baseV1 = base.endsWith('/v1') ? base : `${base}/v1`

  switch (language) {
    case 'python':
      return {
        path: 'Python',
        content: `from openai import OpenAI

client = OpenAI(
    base_url="${baseV1}",
    api_key="${apiKey}",
)

response = client.chat.completions.create(
    model="gpt-5.4",
    messages=[{"role": "user", "content": "Hello"}],
)
print(response.choices[0].message.content)`
      }
    case 'curl':
      return {
        path: 'cURL',
        content: `curl ${baseV1}/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer ${apiKey}" \\
  -d '{
    "model": "gpt-5.4",
    "messages": [{"role": "user", "content": "Hello"}]
  }'`
      }
    case 'nodejs':
      return {
        path: 'Node.js',
        content: `import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "${baseV1}",
  apiKey: "${apiKey}",
});

const response = await client.chat.completions.create({
  model: "gpt-5.4",
  messages: [{ role: "user", content: "Hello" }],
});
console.log(response.choices[0].message.content);`
      }
  }
}
