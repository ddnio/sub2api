import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AffiliateInviteCard from '../AffiliateInviteCard.vue'
import PaymentSuccessReferralCard from '../PaymentSuccessReferralCard.vue'
import { resetPaymentSuccessReferralForTest } from '@/utils/referralShare'

const { fetchPublicSettings, getAffiliateDetail, copyToClipboard, authState } = vi.hoisted(() => ({
  fetchPublicSettings: vi.fn(),
  getAffiliateDetail: vi.fn(),
  copyToClipboard: vi.fn(),
  authState: {
    isAuthenticated: true,
    user: { id: 101 } as { id: number } | null
  }
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'zh' }
  })
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ fetchPublicSettings }),
  useAuthStore: () => authState
}))

vi.mock('@/api/user', () => ({
  userAPI: { getAffiliateDetail }
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard })
}))

describe('Referral share cards', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    resetPaymentSuccessReferralForTest()
    vi.clearAllMocks()
    authState.isAuthenticated = true
    authState.user = { id: 101 }
    localStorage.setItem('auth_token', 'token-a')
    localStorage.setItem('auth_user', JSON.stringify({ id: 101 }))
    fetchPublicSettings.mockResolvedValue({ affiliate_enabled: true })
    getAffiliateDetail.mockResolvedValue({
      aff_code: 'CODE42',
      effective_rebate_rate_percent: 20
    })
  })

  it('AffiliateInviteCard copies share text and ?use=codex&aff link', async () => {
    const wrapper = mount(AffiliateInviteCard, {
      global: {
        stubs: {
          Icon: true,
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })
    await flushPromises()

    await wrapper.get('[data-testid="affiliate-copy-share-text"]').trigger('click')
    expect(copyToClipboard).toHaveBeenCalledWith(
      expect.stringContaining(`${window.location.origin}/register?use=codex&aff=CODE42`),
      'firstPayNudges.shareTextCopied'
    )
    const copiedShare = copyToClipboard.mock.calls[0][0] as string
    expect(copiedShare).not.toMatch(/终身|不封顶|返佣|%|ddnio\.com/i)

    await wrapper.get('[data-testid="affiliate-copy-link"]').trigger('click')
    expect(copyToClipboard).toHaveBeenLastCalledWith(
      `${window.location.origin}/register?use=codex&aff=CODE42`,
      'firstPayNudges.linkCopied'
    )
  })

  it('PaymentSuccessReferralCard skips /user/aff when unauthenticated', async () => {
    localStorage.clear()
    authState.isAuthenticated = false
    authState.user = null

    const wrapper = mount(PaymentSuccessReferralCard, {
      global: { stubs: { Icon: true } }
    })
    await flushPromises()

    expect(getAffiliateDetail).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="payment-success-referral-card"]').exists()).toBe(false)
  })

  it('PaymentSuccessReferralCard stays hidden silently on 401 from /user/aff', async () => {
    getAffiliateDetail.mockRejectedValueOnce({ status: 401, message: 'Unauthorized' })

    const wrapper = mount(PaymentSuccessReferralCard, {
      global: { stubs: { Icon: true } }
    })
    await flushPromises()

    expect(getAffiliateDetail).toHaveBeenCalledWith({ skipAuthRedirect: true })
    expect(wrapper.find('[data-testid="payment-success-referral-card"]').exists()).toBe(false)
  })

  it('PaymentSuccessReferralCard isolates dismissal per user ID and refreshes aff_code before copy', async () => {
    const wrapper = mount(PaymentSuccessReferralCard, {
      global: { stubs: { Icon: true } }
    })
    await flushPromises()

    expect(getAffiliateDetail).toHaveBeenCalledWith({ skipAuthRedirect: true })
    expect(wrapper.find('[data-testid="payment-success-referral-card"]').exists()).toBe(true)

    // Simulate switching to user B in another tab before clicking copy
    localStorage.setItem('auth_token', 'token-b')
    localStorage.setItem('auth_user', JSON.stringify({ id: 202 }))
    getAffiliateDetail.mockResolvedValueOnce({
      aff_code: 'CODE_USER_B',
      effective_rebate_rate_percent: 20
    })

    await wrapper.get('[data-testid="payment-referral-copy-share"]').trigger('click')
    await flushPromises()
    expect(copyToClipboard).toHaveBeenCalledWith(
      expect.stringContaining(`${window.location.origin}/register?use=codex&aff=CODE_USER_B`),
      'firstPayNudges.shareTextCopied'
    )

    // Dismiss as user B (202): user B stays hidden, user A (101) can still see the card
    await wrapper.get('[data-testid="payment-referral-dismiss"]').trigger('click')
    expect(wrapper.find('[data-testid="payment-success-referral-card"]').exists()).toBe(false)

    const secondMountUserB = mount(PaymentSuccessReferralCard, {
      global: { stubs: { Icon: true } }
    })
    await flushPromises()
    expect(secondMountUserB.find('[data-testid="payment-success-referral-card"]').exists()).toBe(false)

    localStorage.setItem('auth_token', 'token-a')
    localStorage.setItem('auth_user', JSON.stringify({ id: 101 }))
    const mountUserA = mount(PaymentSuccessReferralCard, {
      global: { stubs: { Icon: true } }
    })
    await flushPromises()
    expect(mountUserA.find('[data-testid="payment-success-referral-card"]').exists()).toBe(true)
  })
})
