import { describe, expect, it } from 'vitest'
import { createCodexCliFiles } from '@/utils/codexConfig'

describe('createCodexCliFiles', () => {
  it('builds codexcn-style standard config for macOS/Linux', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com',
      apiKey: 'your-api-key-here',
      os: 'unix',
      providerName: 'Sub2API'
    })

    expect(files[0].path).toBe('~/.codex/config.toml')
    expect(files[0].content).toContain('model_provider = "sub2api"')
    expect(files[0].content).toContain('model = "gpt-5.3-codex"')
    expect(files[0].content).toContain('[model_providers.sub2api]')
    expect(files[0].content).toContain('name = "Sub2API"')
    expect(files[0].content).toContain('base_url = "https://relay.example.com/v1"')
    expect(files[0].content).not.toContain('supports_websockets = true')
    expect(files[1].path).toBe('~/.codex/auth.json')
    expect(files[1].content).toContain('"OPENAI_API_KEY": "your-api-key-here"')
  })

  it('uses Windows config paths and includes windows sandbox settings', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com/v1',
      apiKey: 'sk-real-key',
      os: 'windows',
      providerName: 'Sub2API'
    })

    expect(files[0].path).toBe('%userprofile%\\.codex/config.toml')
    expect(files[0].content).toContain('[windows]')
    expect(files[0].content).toContain('sandbox = "elevated"')
    expect(files[1].path).toBe('%userprofile%\\.codex/auth.json')
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
})
