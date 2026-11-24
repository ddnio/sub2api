import { beforeEach, describe, expect, it } from 'vitest'
import {
  buildCodexReferralLink,
  buildReferralShareText,
  dismissPaymentSuccessReferral,
  isPaymentSuccessReferralDismissed,
  paymentSuccessReferralDismissKey,
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

  it('persists one-time dismissal isolated per user ID', () => {
    expect(isPaymentSuccessReferralDismissed(101)).toBe(false)
    expect(isPaymentSuccessReferralDismissed(202)).toBe(false)
    dismissPaymentSuccessReferral(101)
    expect(localStorage.getItem(paymentSuccessReferralDismissKey(101))).toBe('1')
    expect(isPaymentSuccessReferralDismissed(101)).toBe(true)
    expect(isPaymentSuccessReferralDismissed(202)).toBe(false)
    expect(isPaymentSuccessReferralDismissed(null)).toBe(false)
  })
})
