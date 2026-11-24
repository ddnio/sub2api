import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { ApiKey, Group, User } from '@/types'
import {
  DEFAULT_START_USE_CASE,
  START_USE_CASES,
  baseUrlFor,
  buildTestRequest,
  dismissFirstCallBanner,
  extractError,
  extractReply,
  firstCallBannerDismissKey,
  markStartDone,
  pickExistingKeyForUseCase,
  pickGroupForUseCase,
  rememberUseCase,
  rememberedUseCase,
  resetFirstCallBannerMemoryForTest,
  shouldOfferStart,
  shouldRedirectOAuthStartToDashboard,
  shouldShowFirstCallBanner,
  startDoneKey,
  useCaseFromQuery
} from '../startOnboarding'

const NOW = Date.parse('2026-11-23T12:00:00Z')

function user(overrides: Partial<User> = {}): User {
  return { id: 42, role: 'user', created_at: '2026-11-23T10:00:00Z', ...overrides } as User
}

function group(overrides: Partial<Group>): Group {
  return {
    id: 1,
    name: 'g',
    platform: 'openai',
    status: 'active',
    subscription_type: 'standard',
    is_exclusive: false,
    claude_code_only: false,
    ...overrides
  } as Group
}

// 新用户在生产上能看到的非专属分组（2026-11-23 只读查询）
const PROD_GROUPS = [
  group({ id: 3, name: 'OpenAI GPT', platform: 'openai' }),
  group({ id: 4, name: 'ClaudeCode专用', platform: 'anthropic' }),
  group({ id: 18, name: 'GTP-Image', platform: 'openai' }),
  group({ id: 26, name: 'grok', platform: 'grok' as Group['platform'] }),
  group({ id: 29, name: 'antigravity', platform: 'antigravity' }),
  group({ id: 33, name: 'Antigravity-Lite月卡', platform: 'antigravity', subscription_type: 'subscription' })
]

describe('shouldOfferStart', () => {
  beforeEach(() => localStorage.clear())

  it('offers to a fresh user', () => {
    expect(shouldOfferStart(user(), NOW)).toBe(true)
  })

  it('stops after 7 days', () => {
    expect(shouldOfferStart(user({ created_at: '2026-11-15T10:00:00Z' }), NOW)).toBe(false)
  })

  it('stops once marked done in this browser', () => {
    markStartDone(42)
    expect(localStorage.getItem(startDoneKey(42))).toBe('1')
    expect(shouldOfferStart(user(), NOW)).toBe(false)
    expect(shouldOfferStart(user({ id: 43 }), NOW)).toBe(true)
  })

  it('remembers done in memory when localStorage cannot be written', () => {
    const spy = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('QuotaExceededError')
    })
    try {
      markStartDone(77)
      expect(shouldOfferStart(user({ id: 77 }), NOW)).toBe(false)
    } finally {
      spy.mockRestore()
    }
  })

  it('does not offer when localStorage is unavailable', () => {
    const spy = vi.spyOn(window, 'localStorage', 'get').mockImplementation(() => {
      throw new Error('SecurityError')
    })
    try {
      expect(shouldOfferStart(user({ id: 78 }), NOW)).toBe(false)
    } finally {
      spy.mockRestore()
    }
  })

  it('never offers to admins or incomplete users', () => {
    expect(shouldOfferStart(user({ role: 'admin' }), NOW)).toBe(false)
    expect(shouldOfferStart(null, NOW)).toBe(false)
    expect(shouldOfferStart(user({ created_at: '' }), NOW)).toBe(false)
    expect(shouldOfferStart(user({ created_at: 'not a date' }), NOW)).toBe(false)
  })
})

describe('pickGroupForUseCase', () => {
  it('picks the named production groups', () => {
    expect(pickGroupForUseCase(PROD_GROUPS, 'claude')?.id).toBe(4)
    for (const uc of ['codex', 'client', 'code'] as const) {
      expect(pickGroupForUseCase(PROD_GROUPS, uc)?.id).toBe(3)
    }
  })

  it('falls back within the same platform when renamed', () => {
    const renamed = PROD_GROUPS.map((g) => (g.id === 4 ? { ...g, name: 'Claude 按量' } : g))
    expect(pickGroupForUseCase(renamed, 'claude')?.id).toBe(4)
  })

  it('never gives Claude an openai group or GPT an anthropic group', () => {
    const onlyOpenai = PROD_GROUPS.filter((g) => g.platform === 'openai')
    expect(pickGroupForUseCase(onlyOpenai, 'claude')).toBeNull()
    const onlyAnthropic = PROD_GROUPS.filter((g) => g.platform === 'anthropic')
    expect(pickGroupForUseCase(onlyAnthropic, 'code')).toBeNull()
  })

  it('claude_code_only groups are a fallback for Claude Code only', () => {
    const groups = [group({ id: 4, name: 'Claude 改名', platform: 'anthropic', claude_code_only: true })]
    expect(pickGroupForUseCase(groups, 'claude')?.id).toBe(4)
    const openaiOnly = [group({ id: 9, name: 'gpt 改名', claude_code_only: true })]
    expect(pickGroupForUseCase(openaiOnly, 'code')).toBeNull()
  })

  it('fallback skips subscription, exclusive and inactive groups', () => {
    const groups = [
      group({ id: 5, name: 'GPT-Lite', subscription_type: 'subscription' }),
      group({ id: 11, name: 'openai 兼容', is_exclusive: true }),
      group({ id: 12, name: 'old', status: 'inactive' })
    ]
    expect(pickGroupForUseCase(groups, 'code')).toBeNull()
  })
})

describe('urls and test request', () => {
  it('claude base url has no /v1, openai has /v1', () => {
    expect(baseUrlFor('https://router.nanafox.com', 'anthropic')).toBe('https://router.nanafox.com')
    expect(baseUrlFor('https://router.nanafox.com/', 'openai')).toBe('https://router.nanafox.com/v1')
    expect(baseUrlFor('https://router.nanafox.com/v1', 'anthropic')).toBe('https://router.nanafox.com')
  })

  it('anthropic test request', () => {
    const r = buildTestRequest('https://router.nanafox.com', 'anthropic', 'sk-a')
    expect(r.url).toBe('https://router.nanafox.com/v1/messages')
    expect((r.init.headers as Record<string, string>)['x-api-key']).toBe('sk-a')
    const body = JSON.parse(r.init.body as string)
    expect(body.model).toBe('claude-sonnet-5-5')
    expect(body.max_tokens).toBeLessThanOrEqual(64)
  })

  it('openai test request', () => {
    const r = buildTestRequest('https://router.nanafox.com/v1', 'openai', 'sk-o')
    expect(r.url).toBe('https://router.nanafox.com/v1/chat/completions')
    expect((r.init.headers as Record<string, string>).Authorization).toBe('Bearer sk-o')
    expect(JSON.parse(r.init.body as string).model).toBe('gpt-6.1-sol')
  })

  it('extracts replies and errors from both protocols', () => {
    expect(extractReply({ content: [{ type: 'text', text: '你好' }] })).toBe('你好')
    expect(extractReply({ choices: [{ message: { content: ' hi ' } }] })).toBe('hi')
    expect(extractReply(null)).toBe('')
    expect(extractError({ error: { message: 'insufficient balance' } })).toBe('insufficient balance')
    expect(extractError({ error: 'bad key' })).toBe('bad key')
  })
})

describe('useCaseFromQuery', () => {
  it('accepts known values only', () => {
    expect(useCaseFromQuery('claude')).toBe('claude')
    expect(useCaseFromQuery(['codex'])).toBe('codex')
    expect(useCaseFromQuery('evil')).toBeNull()
    expect(useCaseFromQuery(undefined)).toBeNull()
  })
})

describe('rememberUseCase', () => {
  beforeEach(() => sessionStorage.clear())

  it('keeps a valid use= for later and ignores junk', () => {
    expect(rememberedUseCase()).toBeNull()
    rememberUseCase('evil')
    expect(rememberedUseCase()).toBeNull()
    rememberUseCase(['claude'])
    expect(rememberedUseCase()).toBe('claude')
    rememberUseCase(undefined)
    expect(rememberedUseCase()).toBe('claude')
  })
})

describe('default start use case and order', () => {
  it('defaults to codex and lists codex first', () => {
    expect(DEFAULT_START_USE_CASE).toBe('codex')
    expect(START_USE_CASES).toEqual(['codex', 'claude', 'client', 'code'])
  })
})

describe('shouldShowFirstCallBanner', () => {
  beforeEach(() => {
    localStorage.clear()
    resetFirstCallBannerMemoryForTest()
  })

  it('shows for a user registered within 7 days who has a key and zero requests', () => {
    expect(shouldShowFirstCallBanner(user(), { total_api_keys: 1, total_requests: 0 }, NOW)).toBe(true)
  })

  it('hides when registered more than 7 days ago', () => {
    expect(
      shouldShowFirstCallBanner(
        user({ created_at: '2026-11-15T10:00:00Z' }),
        { total_api_keys: 1, total_requests: 0 },
        NOW
      )
    ).toBe(false)
  })

  it('hides when the user has no key or already made a call', () => {
    expect(shouldShowFirstCallBanner(user(), { total_api_keys: 0, total_requests: 0 }, NOW)).toBe(false)
    expect(shouldShowFirstCallBanner(user(), { total_api_keys: 2, total_requests: 1 }, NOW)).toBe(false)
  })

  it('hides after dismissal only for that user', () => {
    dismissFirstCallBanner(42)
    expect(localStorage.getItem(firstCallBannerDismissKey(42))).toBe('1')
    expect(shouldShowFirstCallBanner(user({ id: 42 }), { total_api_keys: 1, total_requests: 0 }, NOW)).toBe(false)
    expect(shouldShowFirstCallBanner(user({ id: 43 }), { total_api_keys: 1, total_requests: 0 }, NOW)).toBe(true)
  })

  it('hides for admins or missing data', () => {
    expect(shouldShowFirstCallBanner(user({ role: 'admin' }), { total_api_keys: 1, total_requests: 0 }, NOW)).toBe(false)
    expect(shouldShowFirstCallBanner(user(), null, NOW)).toBe(false)
    expect(shouldShowFirstCallBanner(null, { total_api_keys: 1, total_requests: 0 }, NOW)).toBe(false)
  })

  it('hides for subscribed or topped-up users', () => {
    expect(
      shouldShowFirstCallBanner(
        user({ subscriptions: [{ id: 1, status: 'active' } as any] }),
        { total_api_keys: 1, total_requests: 0 },
        NOW
      )
    ).toBe(false)
    expect(
      shouldShowFirstCallBanner(
        user({ balance: 25 }),
        { total_api_keys: 1, total_requests: 0 },
        { now: NOW, defaultSignupBalance: 5 }
      )
    ).toBe(false)
    expect(
      shouldShowFirstCallBanner(
        user({ balance: 5 }),
        { total_api_keys: 1, total_requests: 0 },
        { now: NOW, defaultSignupBalance: 5 }
      )
    ).toBe(true)
  })
})

describe('pickExistingKeyForUseCase', () => {
  const openaiGroup = group({ id: 3, name: 'OpenAI GPT', platform: 'openai' })
  const claudeGroup = group({ id: 4, name: 'ClaudeCode专用', platform: 'anthropic' })

  it('picks only an active key matching the selected platform', () => {
    const keys = [
      { id: 1, key: 'sk-inactive', status: 'inactive', group: openaiGroup },
      { id: 2, key: 'sk-claude', status: 'active', group: claudeGroup },
      { id: 3, key: 'sk-openai-ok', status: 'active', group: openaiGroup }
    ] as ApiKey[]
    expect(pickExistingKeyForUseCase(keys, 'codex')?.key).toBe('sk-openai-ok')
    expect(pickExistingKeyForUseCase(keys, 'claude')?.key).toBe('sk-claude')
  })

  it('returns null when only wrong-platform or inactive keys exist', () => {
    const keys = [
      { id: 1, key: 'sk-inactive', status: 'inactive', group: openaiGroup },
      { id: 2, key: 'sk-claude', status: 'active', group: claudeGroup }
    ] as ApiKey[]
    expect(pickExistingKeyForUseCase(keys, 'codex')).toBeNull()
  })
})

describe('shouldRedirectOAuthStartToDashboard', () => {
  beforeEach(() => localStorage.clear())

  it('keeps fresh OAuth users on /start and sends older or done accounts to /dashboard', () => {
    expect(shouldRedirectOAuthStartToDashboard('/start', { use: 'codex', from: 'oauth' }, user({ id: 99 }), NOW)).toBe(false)
    expect(
      shouldRedirectOAuthStartToDashboard(
        '/start',
        { use: 'codex', from: 'oauth' },
        user({ id: 100, created_at: '2026-11-01T00:00:00Z' }),
        NOW
      )
    ).toBe(true)
    expect(
      shouldRedirectOAuthStartToDashboard(
        '/start',
        { use: 'codex', from: 'oauth', step: 'config' },
        user({ id: 100, created_at: '2026-11-01T00:00:00Z' }),
        NOW
      )
    ).toBe(false)
    expect(
      shouldRedirectOAuthStartToDashboard(
        '/start',
        { use: 'codex' },
        user({ id: 100, created_at: '2026-11-01T00:00:00Z' }),
        NOW
      )
    ).toBe(false)
  })
})
