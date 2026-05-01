import { beforeEach, describe, expect, it, vi } from 'vitest'
import { paymentAPI } from '@/api/payment'
import adminPaymentAPI from '@/api/admin/payment'
import { apiClient } from '@/api/client'

vi.mock('@/api/client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

const mockedClient = vi.mocked(apiClient)

describe('payment API contracts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads the current user orders through the upstream /payment/orders/my endpoint', async () => {
    mockedClient.get.mockResolvedValueOnce({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 1 } })

    await paymentAPI.listOrders(1, 20, { status: 'PENDING' })

    expect(mockedClient.get).toHaveBeenCalledWith('/payment/orders/my', {
      params: { page: 1, page_size: 20, status: 'PENDING' },
      signal: undefined
    })
  })

  it('gets order status by reading the order resource, not a non-existent /status route', async () => {
    mockedClient.get.mockResolvedValueOnce({ data: { status: 'COMPLETED' } })

    const result = await paymentAPI.getOrderStatus(42)

    expect(mockedClient.get).toHaveBeenCalledWith('/payment/orders/42')
    expect(result.status).toBe('COMPLETED')
  })

  it('loads checkout info from the upstream payment v2 aggregate endpoint', async () => {
    mockedClient.get.mockResolvedValueOnce({ data: { global_min: 1, global_max: 10000, methods: {}, plans: [] } })

    const result = await paymentAPI.getCheckoutInfo()

    expect(mockedClient.get).toHaveBeenCalledWith('/payment/checkout-info')
    expect(result.data.global_min).toBe(1)
  })

  it('uses upstream admin order operations', async () => {
    mockedClient.get.mockResolvedValueOnce({ data: { total_count: 3 } })
    mockedClient.post.mockResolvedValueOnce({ data: undefined })
    mockedClient.post.mockResolvedValueOnce({ data: undefined })

    await adminPaymentAPI.getDashboard()
    await adminPaymentAPI.retryOrder(7)
    await adminPaymentAPI.refundOrder(7, { amount: 1.25, reason: 'operator refund', deduct_balance: true })

    expect(mockedClient.get).toHaveBeenCalledWith('/admin/payment/dashboard', { params: undefined })
    expect(mockedClient.post).toHaveBeenNthCalledWith(1, '/admin/payment/orders/7/retry')
    expect(mockedClient.post).toHaveBeenNthCalledWith(2, '/admin/payment/orders/7/refund', {
      amount: 1.25,
      reason: 'operator refund',
      deduct_balance: true
    })
  })

  it('normalizes legacy admin orders with missing amount and user_email', async () => {
    mockedClient.get.mockResolvedValueOnce({
      data: {
        items: [{ id: 55, user_email: 'user@example.com', status: 'FAILED' }],
        total: 1,
        page: 1,
        page_size: 20,
        pages: 1
      }
    })

    const result = await adminPaymentAPI.listOrders(1, 20, {})

    expect(result.items[0].email).toBe('user@example.com')
    expect(result.items[0].amount).toBe(0)
    expect(result.items[0].pay_amount).toBe(0)
  })

  it('admin plans API returns the array contract from /admin/payment/plans', async () => {
    const plans = [{ id: 1, name: 'Basic' }]
    mockedClient.get.mockResolvedValueOnce({ data: plans })

    const result = await adminPaymentAPI.listPlans()

    expect(mockedClient.get).toHaveBeenCalledWith('/admin/payment/plans', { signal: undefined })
    expect(result).toBe(plans)
  })
})
