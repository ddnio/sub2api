// NanaFox fork: 注册后 /start 引导页的纯逻辑（选分组、拼测试请求、“已看过”标记）。
// 单独成文件，方便单测，也让上游文件的改动只剩路由里的几行。
import type { Group, User } from '@/types'

export type StartUseCase = 'claude' | 'codex' | 'client' | 'code'

export const START_USE_CASES: StartUseCase[] = ['claude', 'codex', 'client', 'code']

// 注册多少天内、还没看过引导的用户，去控制台时先带到 /start
export const START_OFFER_DAYS = 7

const CLAUDE_GROUP_NAME = 'ClaudeCode专用'
const OPENAI_GROUP_NAME = 'OpenAI GPT'

export const START_TEST_MODELS = {
  anthropic: 'claude-sonnet-5-5',
  openai: 'gpt-6.1-sol'
} as const

export function startDoneKey(userId: number): string {
  return `nf_start_done_${userId}`
}

function safeStorage(): Storage | null {
  try {
    return typeof window !== 'undefined' ? window.localStorage : null
  } catch {
    return null
  }
}

// 本次打开页面内的标记：localStorage 写不进去时，也不能让守卫把人反复拦回 /start
const doneInMemory = new Set<number>()

export function markStartDone(userId: number | undefined | null): void {
  if (!userId) return
  doneInMemory.add(userId)
  try {
    safeStorage()?.setItem(startDoneKey(userId), '1')
  } catch {
    // 存储写满或只读：靠内存标记，刷新后最多再看一次引导页
  }
}

export function isStartDone(userId: number): boolean {
  if (doneInMemory.has(userId)) return true
  try {
    const storage = safeStorage()
    // 没有 localStorage 就不引导，免得记不住“看过了”
    return !storage || storage.getItem(startDoneKey(userId)) === '1'
  } catch {
    return true
  }
}

// 同步判断，不发请求：路由守卫里用
export function shouldOfferStart(user: User | null | undefined, now: number = Date.now()): boolean {
  if (!user || user.role === 'admin' || !user.id || !user.created_at) return false
  const created = Date.parse(user.created_at)
  if (Number.isNaN(created)) return false
  if (now - created > START_OFFER_DAYS * 24 * 3600 * 1000) return false
  return !isStartDone(user.id)
}

export function useCaseFromQuery(value: unknown): StartUseCase | null {
  const v = Array.isArray(value) ? value[0] : value
  return typeof v === 'string' && (START_USE_CASES as string[]).includes(v) ? (v as StartUseCase) : null
}

// Claude 只走 ClaudeCode专用（Anthropic 协议）；其余用途一律 OpenAI GPT
export function platformForUseCase(useCase: StartUseCase): 'anthropic' | 'openai' {
  return useCase === 'claude' ? 'anthropic' : 'openai'
}

// 按名字找分组；名字改了就退回同平台第一个按量计费、非专属的分组；绝不跨平台
export function pickGroupForUseCase(groups: Group[], useCase: StartUseCase): Group | null {
  const platform = platformForUseCase(useCase)
  const wantName = platform === 'anthropic' ? CLAUDE_GROUP_NAME : OPENAI_GROUP_NAME
  const usable = groups.filter((g) => g.platform === platform && g.status !== 'inactive')
  const byName = usable.find((g) => g.name === wantName)
  if (byName) return byName
  // 「仅限 Claude Code 客户端」的分组只给 Claude Code 用途兜底
  return (
    usable.find(
      (g) => g.subscription_type === 'standard' && !g.is_exclusive && (useCase === 'claude' || !g.claude_code_only)
    ) ?? null
  )
}

export function baseUrlFor(origin: string, platform: 'anthropic' | 'openai'): string {
  const root = origin.replace(/\/+$/, '').replace(/\/v1$/, '')
  return platform === 'anthropic' ? root : `${root}/v1`
}

export interface StartTestRequest {
  url: string
  init: RequestInit
  model: string
}

export function buildTestRequest(origin: string, platform: 'anthropic' | 'openai', apiKey: string): StartTestRequest {
  const root = baseUrlFor(origin, 'anthropic')
  if (platform === 'anthropic') {
    const model = START_TEST_MODELS.anthropic
    return {
      model,
      url: `${root}/v1/messages`,
      init: {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'x-api-key': apiKey, 'anthropic-version': '2023-06-01' },
        body: JSON.stringify({ model, max_tokens: 32, messages: [{ role: 'user', content: '你好，用一句话回复我' }] })
      }
    }
  }
  const model = START_TEST_MODELS.openai
  return {
    model,
    url: `${root}/v1/chat/completions`,
    init: {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${apiKey}` },
      body: JSON.stringify({ model, max_tokens: 32, messages: [{ role: 'user', content: '你好，用一句话回复我' }] })
    }
  }
}

// 从两种协议的返回里取出回复文字或错误信息
export function extractReply(data: unknown): string {
  const d = data as Record<string, any> | null
  if (!d || typeof d !== 'object') return ''
  if (Array.isArray(d.content)) {
    return d.content.map((c: any) => (typeof c?.text === 'string' ? c.text : '')).join('').trim()
  }
  const msg = d.choices?.[0]?.message?.content
  return typeof msg === 'string' ? msg.trim() : ''
}

export function extractError(data: unknown): string {
  const d = data as Record<string, any> | null
  if (!d || typeof d !== 'object') return ''
  const e = d.error
  if (typeof e === 'string') return e
  if (e && typeof e.message === 'string') return e.message
  if (typeof d.message === 'string') return d.message
  return ''
}
