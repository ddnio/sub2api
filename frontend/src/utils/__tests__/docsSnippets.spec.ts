import { describe, expect, it } from 'vitest'
import { generateAnthropicApiExample, generateOpenCodeEnhancedSnippet } from '../docsSnippets'

describe('generateAnthropicApiExample', () => {
  it('uses the site root and x-api-key for curl', () => {
    const { content } = generateAnthropicApiExample('https://relay.test/v1/', 'sk-test', 'curl')
    expect(content).toContain('curl https://relay.test/v1/messages')
    expect(content).toContain('x-api-key: sk-test')
    expect(content).toContain('anthropic-version: 2023-06-01')
    expect(content).toContain('claude-sonnet-5-5')
  })

  it('passes the root base URL to the SDKs', () => {
    expect(generateAnthropicApiExample('https://relay.test', 'k', 'python').content)
      .toContain('base_url="https://relay.test"')
    expect(generateAnthropicApiExample('https://relay.test/', 'k', 'nodejs').content)
      .toContain('baseURL: "https://relay.test"')
  })
})

describe('generateOpenCodeEnhancedSnippet', () => {
  it('only lists GPT models in the OpenAI-compatible provider', () => {
    const config = JSON.parse(generateOpenCodeEnhancedSnippet('https://relay.test/v1', 'k').content)
    expect(config.model).toBe('openai/gpt-6.1-sol')
    expect(Object.keys(config.provider.openai.models)).toEqual(['gpt-6.1-sol', 'gpt-6.2-sol'])
  })
})
