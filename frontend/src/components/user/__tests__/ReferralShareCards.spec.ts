import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AffiliateInviteCard from '../AffiliateInviteCard.vue'
import PaymentSuccessReferralCard from '../PaymentSuccessReferralCard.vue'
import { resetPaymentSuccessReferralForTest } from '@/utils/referralShare'

const { fetchPublicSettings, getAffiliateDetail, copyToClipboard } = vi.hoisted(() => ({
  fetchPublicSettings: vi.fn(),
  getAffiliateDetail: vi.fn(),
  copyToClipboard: vi.fn()
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'zh' }
  })
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ fetchPublicSettings })
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

  it('PaymentSuccessReferralCard shows once, copies share text, and stays hidden after dismissal', async () => {
    const wrapper = mount(PaymentSuccessReferralCard, {
      global: { stubs: { Icon: true } }
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="payment-success-referral-card"]').exists()).toBe(true)
    await wrapper.get('[data-testid="payment-referral-copy-share"]').trigger('click')
    expect(copyToClipboard).toHaveBeenCalledWith(
      expect.stringContaining(`${window.location.origin}/register?use=codex&aff=CODE42`),
      'firstPayNudges.shareTextCopied'
    )

    await wrapper.get('[data-testid="payment-referral-dismiss"]').trigger('click')
    expect(wrapper.find('[data-testid="payment-success-referral-card"]').exists()).toBe(false)

    // Mounting again should not show the prompt a second time
    const secondMount = mount(PaymentSuccessReferralCard, {
      global: { stubs: { Icon: true } }
    })
    await flushPromises()
    expect(secondMount.find('[data-testid="payment-success-referral-card"]').exists()).toBe(false)
  })
})
