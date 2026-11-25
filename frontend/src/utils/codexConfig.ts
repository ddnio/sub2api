import { buildOpenAICodexConfigToml, type CodexAuthMode } from '@/utils/codexAuthConfig'

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
  authMode?: CodexAuthMode
  model?: string
}

export function createCodexCliFiles(options: CreateCodexCliFilesOptions): CodexCliFile[] {
  const baseUrl = options.baseUrl.trim() || window.location.origin
  const configDir = options.os === 'windows' ? '%userprofile%\\.codex' : '~/.codex'
  const authMode: CodexAuthMode = options.authMode ?? 'api-key'

  const configContent = buildOpenAICodexConfigToml({
    baseUrl,
    apiKey: options.apiKey,
    authMode,
    model: options.model,
    websocket: options.websocket
  })

  const files: CodexCliFile[] = [
    {
      path: `${configDir}/config.toml`,
      content: configContent
    }
  ]

  if (authMode === 'legacy') {
    files.push({
      path: `${configDir}/auth.json`,
      content: JSON.stringify({ OPENAI_API_KEY: options.apiKey }, null, 2)
    })
  }

  return files
}
