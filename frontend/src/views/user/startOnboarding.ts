// NanaFox fork: 注册后 /start 引导页的纯逻辑（选分组、拼测试请求、“已看过”标记）。
// 单独成文件，方便单测，也让上游文件的改动只剩路由里的几行。
import type { ApiKey, Group, User } from '@/types'

export type StartUseCase = 'codex' | 'claude' | 'client' | 'code'

export const START_USE_CASES: StartUseCase[] = ['codex', 'claude', 'client', 'code']
export const DEFAULT_START_USE_CASE: StartUseCase = 'codex'

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

// 注册页 OAuth 回调带 /start?use=...&from=oauth：老账号（不满足新手引导条件）改回 /dashboard
export function shouldRedirectOAuthStartToDashboard(
  path: string,
  query: Record<string, unknown> | undefined,
  user: User | null | undefined,
  now: number = Date.now()
): boolean {
  if (path !== '/start') return false
  if (query?.from !== 'oauth' || query?.step === 'config') return false
  return !shouldOfferStart(user, now)
}

export function firstCallBannerDismissKey(userId: number): string {
  return `nf_first_call_banner_dismissed_${userId}`
}

const dismissedBannerInMemory = new Set<number>()

export function dismissFirstCallBanner(userId: number | undefined | null): void {
  if (!userId) return
  dismissedBannerInMemory.add(userId)
  try {
    safeStorage()?.setItem(firstCallBannerDismissKey(userId), '1')
  } catch {
    // 存储不可写时回退到内存标记
  }
}

export function isFirstCallBannerDismissed(userId: number): boolean {
  if (dismissedBannerInMemory.has(userId)) return true
  try {
    const storage = safeStorage()
    return storage?.getItem(firstCallBannerDismissKey(userId)) === '1'
  } catch {
    return false
  }
}

export function resetFirstCallBannerMemoryForTest(): void {
  dismissedBannerInMemory.clear()
}

export interface FirstCallBannerStats {
  total_api_keys?: number
  total_requests?: number
}

export interface FirstCallBannerOptions {
  now?: number
  defaultSignupBalance?: number | null
}

function hasActiveSubscription(user: User): boolean {
  if (!Array.isArray(user.subscriptions) || user.subscriptions.length === 0) return false
  return user.subscriptions.some((s) => !s.status || s.status === 'active')
}

// 控制台顶部轻提醒：注册 7 天内、未订阅/未充值、有 Key、从没成功调用过、且未关闭横幅时显示
export function shouldShowFirstCallBanner(
  user: User | null | undefined,
  stats: FirstCallBannerStats | null | undefined,
  nowOrOptions: number | FirstCallBannerOptions = Date.now()
): boolean {
  const options: FirstCallBannerOptions =
    typeof nowOrOptions === 'number' ? { now: nowOrOptions } : nowOrOptions
  const now = options.now ?? Date.now()
  if (!user || user.role === 'admin' || !user.id || !user.created_at || !stats) return false
  const created = Date.parse(user.created_at)
  if (Number.isNaN(created)) return false
  if (now - created > START_OFFER_DAYS * 24 * 3600 * 1000) return false
  if (hasActiveSubscription(user)) return false
  if (
    typeof options.defaultSignupBalance === 'number' &&
    options.defaultSignupBalance > 0 &&
    typeof user.balance === 'number' &&
    user.balance > options.defaultSignupBalance + 1e-6
  ) {
    return false
  }
  if ((stats.total_api_keys ?? 0) <= 0) return false
  if ((stats.total_requests ?? 0) > 0) return false
  return !isFirstCallBannerDismissed(user.id)
}

// 横幅/文档带 ?step=config 回到 /start 时，按用途找一把状态为 active 且平台匹配的 Key；找不到就返回 null 停在建 Key 步骤
export function pickExistingKeyForUseCase(keys: ApiKey[], useCase: StartUseCase): ApiKey | null {
  const platform = platformForUseCase(useCase)
  return (
    keys.find(
      (k) =>
        k.status === 'active' &&
        k.group?.platform === platform &&
        (useCase === 'claude' || !k.group?.claude_code_only)
    ) ?? null
  )
}

export function useCaseFromQuery(value: unknown): StartUseCase | null {
  const v = Array.isArray(value) ? value[0] : value
  return typeof v === 'string' && (START_USE_CASES as string[]).includes(v) ? (v as StartUseCase) : null
}

const START_USE_KEY = 'nf_start_use'

// 本标签页内记住推广链接里的用途（注册、验证邮箱、GitHub 登录都会换页，query 会丢）
export function rememberUseCase(value: unknown): void {
  const useCase = useCaseFromQuery(value)
  if (!useCase) return
  try {
    window.sessionStorage.setItem(START_USE_KEY, useCase)
  } catch {
    // 存不了就不预选，不影响使用
  }
}

export function rememberedUseCase(): StartUseCase | null {
  try {
    return useCaseFromQuery(window.sessionStorage.getItem(START_USE_KEY))
  } catch {
    return null
  }
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
