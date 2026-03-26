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
