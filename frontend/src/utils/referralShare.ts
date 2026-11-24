// NanaFox fork: 控制台返佣卡片与付费成功页共用的推荐语、邀请链接与按用户隔离的关闭状态

const PAYMENT_SUCCESS_REFERRAL_DISMISSED_PREFIX = 'nf_payment_success_referral_dismissed_'

const dismissedUsersInMemory = new Set<number>()

export function paymentSuccessReferralDismissKey(userId: number): string {
  return `${PAYMENT_SUCCESS_REFERRAL_DISMISSED_PREFIX}${userId}`
}

export function buildCodexReferralLink(origin: string, affCode: string): string {
  const code = affCode.trim()
  if (!code) return ''
  const normalizedOrigin = origin.trim().replace(/\/+$/, '')
  return `${normalizedOrigin}/register?use=codex&aff=${encodeURIComponent(code)}`
}

export function buildReferralShareText(origin: string, affCode: string, locale = 'zh'): string {
  const link = buildCodexReferralLink(origin, affCode)
  if (!link) return ''
  if (locale.toLowerCase().startsWith('zh')) {
    return `最近写代码主力在用 Codex，这个中转原生支持 Responses 接口，也能直接导入 CC Switch，支持微信按量或月卡，我一直在用：${link}`
  }
  return `I've been using this relay for Codex CLI — it natively supports the Responses API, works with 1-click CC Switch import, and supports pay-as-you-go or monthly plans: ${link}`
}

function safeLocalStorage(): Storage | null {
  try {
    return typeof window !== 'undefined' ? window.localStorage : null
  } catch {
    return null
  }
}

export function isPaymentSuccessReferralDismissed(userId: number | null | undefined): boolean {
  if (!userId || userId <= 0) return false
  if (dismissedUsersInMemory.has(userId)) return true
  try {
    return safeLocalStorage()?.getItem(paymentSuccessReferralDismissKey(userId)) === '1'
  } catch {
    return false
  }
}

export function dismissPaymentSuccessReferral(userId: number | null | undefined): void {
  if (!userId || userId <= 0) return
  dismissedUsersInMemory.add(userId)
  try {
    safeLocalStorage()?.setItem(paymentSuccessReferralDismissKey(userId), '1')
  } catch {
    // localStorage 不可写时回退到内存标记
  }
}

export function resetPaymentSuccessReferralForTest(): void {
  dismissedUsersInMemory.clear()
}
