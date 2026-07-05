import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import StripePaymentInline from '../StripePaymentInline.vue'

const routerResolve = vi.hoisted(() => vi.fn())
const loadStripe = vi.hoisted(() => vi.fn())
const stripeElements = vi.hoisted(() => ({
  create: vi.fn(),
}))
const stripePaymentElement = vi.hoisted(() => ({
  mount: vi.fn(),
  on: vi.fn(),
}))
const stripeInstance = vi.hoisted(() => ({
  elements: vi.fn(),
  confirmPayment: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ resolve: routerResolve }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    cancelOrder: vi.fn(),
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
  }),
}))

vi.mock('@stripe/stripe-js', () => ({
  loadStripe,
}))

function hrefForRoute(path: string, query: Record<string, unknown>): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined) {
      params.set(key, String(value))
    }
  }
  return `${path}?${params.toString()}`
}

describe('StripePaymentInline', () => {
  let changeHandler: ((event: { value: { type: string } }) => void) | null = null

  beforeEach(() => {
    changeHandler = null
    routerResolve.mockReset().mockImplementation(({ path, query }) => ({
      href: hrefForRoute(path, query || {}),
    }))
    loadStripe.mockReset().mockResolvedValue(stripeInstance)
    stripeElements.create.mockReset().mockReturnValue(stripePaymentElement)
    stripePaymentElement.mount.mockReset()
    stripePaymentElement.on.mockReset().mockImplementation((event: string, callback: unknown) => {
      if (event === 'ready' && typeof callback === 'function') {
        callback()
      }
      if (event === 'change' && typeof callback === 'function') {
        changeHandler = callback as (event: { value: { type: string } }) => void
      }
    })
    stripeInstance.elements.mockReset().mockReturnValue(stripeElements)
    stripeInstance.confirmPayment.mockReset()
    vi.spyOn(window, 'open').mockReturnValue({
      closed: false,
      postMessage: vi.fn(),
    } as unknown as Window)
  })

  it('carries out_trade_no into popup payment routes', async () => {
    const wrapper = mount(StripePaymentInline, {
      props: {
        orderId: 42,
        amount: 100,
        clientSecret: 'pi_secret_42',
        outTradeNo: 'sub2_stripe_42',
        publishableKey: 'pk_test',
        payAmount: 103,
        currency: 'CNY',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await flushPromises()
    expect(changeHandler).toBeTruthy()
    changeHandler?.({ value: { type: 'alipay' } })

    await wrapper.find('button.btn-stripe').trigger('click')

    expect(routerResolve).toHaveBeenCalledWith({
      path: '/payment/stripe-popup',
      query: {
        order_id: '42',
        method: 'alipay',
        amount: '103',
        out_trade_no: 'sub2_stripe_42',
      },
    })
    expect(window.open).toHaveBeenCalledWith(
      expect.stringContaining('out_trade_no=sub2_stripe_42'),
      'paymentPopup',
      expect.any(String),
    )
  })
})
