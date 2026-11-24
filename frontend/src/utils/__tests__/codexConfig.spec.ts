import { describe, expect, it } from 'vitest'
import { createCodexCliFiles } from '@/utils/codexConfig'
import {
  DEFAULT_CODEX_MODEL,
  buildCodexQuickSetupCommand,
  buildOpenAICodexConfigToml
} from '@/utils/codexAuthConfig'

describe('createCodexCliFiles', () => {
  it('defaults to single-file api-key config matching UseKeyModal and /start', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com',
      apiKey: 'your-api-key-here',
      os: 'unix',
      providerName: 'Sub2API'
    })

    expect(files).toHaveLength(1)
    expect(files[0].path).toBe('~/.codex/config.toml')
    expect(files[0].content).toBe(
      buildOpenAICodexConfigToml({
        baseUrl: 'https://relay.example.com',
        apiKey: 'your-api-key-here'
      })
    )
    expect(files[0].content).toContain('model_provider = "OpenAI"')
    expect(files[0].content).toContain(`model = "${DEFAULT_CODEX_MODEL}"`)
    expect(files[0].content).toContain('base_url = "https://relay.example.com/v1"')
    expect(files[0].content).toContain('requires_openai_auth = false')
    expect(files[0].content).toContain('experimental_bearer_token = "your-api-key-here"')
    expect(files[0].content).toContain('http_headers = { "x-openai-actor-authorization" = "local-image-extension" }')
    expect(files[0].content).not.toContain('supports_websockets = true')

    // /start quick setup commands must embed the exact same config.toml
    const unixCmd = buildCodexQuickSetupCommand('https://relay.example.com', 'your-api-key-here', 'unix')
    const psCmd = buildCodexQuickSetupCommand('https://relay.example.com', 'your-api-key-here', 'powershell')
    expect(unixCmd).toContain(files[0].content)
    expect(psCmd).toContain(files[0].content)
  })

  it('uses Windows config path in default api-key mode', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com/v1',
      apiKey: 'sk-real-key',
      os: 'windows',
      providerName: 'Sub2API'
    })

    expect(files).toHaveLength(1)
    expect(files[0].path).toBe('%userprofile%\\.codex/config.toml')
    expect(files[0].content).toContain('base_url = "https://relay.example.com/v1"')
    expect(files[0].content).toContain('experimental_bearer_token = "sk-real-key"')
  })

  it('adds websocket flags only for advanced websocket mode', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com',
      apiKey: 'sk-websocket',
      os: 'unix',
      websocket: true,
      providerName: 'Sub2API'
    })

    expect(files[0].content).toContain('supports_websockets = true')
    expect(files[0].content).toContain('responses_websockets_v2 = true')
  })

  it('supports legacy auth.json mode when explicitly requested', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com',
      apiKey: 'sk-legacy-mode',
      os: 'unix',
      authMode: 'legacy',
      providerName: 'Sub2API'
    })

    expect(files).toHaveLength(2)
    expect(files[0].content).toContain('requires_openai_auth = true')
    expect(files[1].path).toBe('~/.codex/auth.json')
    expect(files[1].content).toContain('"OPENAI_API_KEY": "sk-legacy-mode"')
  })

  it('never turns the prerender site name placeholder into a literal provider key', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://api.prerender.invalid',
      apiKey: 'your-api-key-here',
      os: 'unix',
      providerName: '__PRERENDER_SITE_NAME__'
    })

    expect(files[0].content).toContain('model_provider = "OpenAI"')
    expect(files[0].content).not.toMatch(/prerender_site_name/i)
  })
})
