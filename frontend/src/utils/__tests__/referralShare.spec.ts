import { beforeEach, describe, expect, it } from 'vitest'
import {
  PAYMENT_SUCCESS_REFERRAL_DISMISSED_KEY,
  buildCodexReferralLink,
  buildReferralShareText,
  dismissPaymentSuccessReferral,
  isPaymentSuccessReferralDismissed,
  resetPaymentSuccessReferralForTest
} from '@/utils/referralShare'

describe('referralShare utils', () => {
  beforeEach(() => {
    localStorage.clear()
    resetPaymentSuccessReferralForTest()
  })

  it('builds a codex-first referral link with both use=codex and aff params', () => {
    expect(buildCodexReferralLink('https://router.nanafox.com/', 'ABC123')).toBe(
      'https://router.nanafox.com/register?use=codex&aff=ABC123'
    )
    expect(buildCodexReferralLink('https://router.nanafox.com', '   ')).toBe('')
  })

  it('builds product-focused share copy without forbidden claims or rebate numbers', () => {
    const zh = buildReferralShareText('https://router.nanafox.com', 'INV88', 'zh')
    expect(zh).toContain('https://router.nanafox.com/register?use=codex&aff=INV88')
    expect(zh).toContain('Codex')
    expect(zh).not.toMatch(/终身|不封顶|返佣|%|ddnio\.com/i)

    const en = buildReferralShareText('https://router.nanafox.com', 'INV88', 'en')
    expect(en).toContain('https://router.nanafox.com/register?use=codex&aff=INV88')
    expect(en).not.toMatch(/lifetime|uncapped|rebate|%|ddnio\.com/i)
  })

  it('persists one-time dismissal for the payment success referral prompt', () => {
    expect(isPaymentSuccessReferralDismissed()).toBe(false)
    dismissPaymentSuccessReferral()
    expect(localStorage.getItem(PAYMENT_SUCCESS_REFERRAL_DISMISSED_KEY)).toBe('1')
    expect(isPaymentSuccessReferralDismissed()).toBe(true)
  })
})
