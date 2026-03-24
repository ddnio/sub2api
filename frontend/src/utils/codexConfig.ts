export type CodexCliOS = 'unix' | 'windows'

export interface CodexCliFile {
  path: string
  content: string
}

interface CreateCodexCliFilesOptions {
  apiKey: string
  baseUrl: string
  os: CodexCliOS
  providerName?: string
  websocket?: boolean
}

function normalizeBaseUrl(baseUrl: string): string {
  // Strip /v1 suffix and trailing slashes — Codex CLI handles the path internally
  const trimmed = baseUrl.trim().replace(/\/v1\/?$/, '').replace(/\/+$/, '')
  return trimmed || window.location.origin.replace(/\/+$/, '')
}

function toProviderKey(name: string): string {
  return (
    name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '_')
      .replace(/^_+|_+$/g, '') || 'nanafox'
  )
}

export function createCodexCliFiles(options: CreateCodexCliFilesOptions): CodexCliFile[] {
  const normalizedBaseUrl = normalizeBaseUrl(options.baseUrl)
  const configDir = options.os === 'windows' ? '%userprofile%\\.codex' : '~/.codex'
  const providerName = options.providerName?.trim() || 'NanaFox API'
  const providerKey = toProviderKey(providerName)
  const standardConfig = `model_provider = "${providerKey}"
model = "gpt-5.3-codex"
model_reasoning_effort = "high"
network_access = "enabled"
disable_response_storage = true
windows_wsl_setup_acknowledged = true
model_verbosity = "high"

[model_providers.${providerKey}]
name = "${providerName}"
base_url = "${normalizedBaseUrl}"
wire_api = "responses"
requires_openai_auth = true`
  const websocketExtension = options.websocket
    ? '\n\nsupports_websockets = true\n\n[features]\nresponses_websockets_v2 = true'
    : ''
  const windowsConfig = options.os === 'windows'
    ? '\n\n[windows]\nsandbox = "elevated"'
    : ''

  const configContent = `${standardConfig}${websocketExtension}${windowsConfig}`

  const authContent = `{
  "OPENAI_API_KEY": "${options.apiKey}"
}`

  return [
    {
      path: `${configDir}/config.toml`,
      content: configContent
    },
    {
      path: `${configDir}/auth.json`,
      content: authContent
    }
  ]
}
