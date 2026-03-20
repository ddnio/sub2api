import { apiClient } from './client'
import type { BasePaginationResponse, FetchOptions } from '@/types'

export interface PaymentPlan {
  id: number
  name: string
  description: string
  badge: string | null
  group_id: number
  group_name: string
  duration_days: number
  price: number
  original_price: number | null
  sort_order: number
  is_active: boolean
  created_at: string
}

export interface PaymentOrder {
  id: number
  order_no: string
  type: string
  plan_id: number | null
  plan_name: string | null
  amount: number
  credit_amount: number | null
  currency: string
  status: string
  provider: string | null
  provider_order_no: string | null
  paid_at: string | null
  completed_at: string | null
  expired_at: string
  created_at: string
}

export interface CreateOrderResponse {
  order: PaymentOrder
  qr_code_url: string
}

export interface OrderStatusResponse {
  status: string
}

async function listPlans(): Promise<PaymentPlan[]> {
  const { data } = await apiClient.get('/payment/plans')
  return data
}

async function createOrder(params: {
  type: 'plan' | 'topup'
  plan_id?: number
  amount?: number
  provider: 'wxpay'
}): Promise<CreateOrderResponse> {
  const { data } = await apiClient.post('/payment/orders', params)
  return data
}

async function listOrders(
  page: number,
  pageSize: number,
  params: { status?: string; type?: string },
  options?: FetchOptions
): Promise<BasePaginationResponse<PaymentOrder>> {
  const { data } = await apiClient.get('/payment/orders', {
    params: { page, page_size: pageSize, ...params },
    signal: options?.signal
  })
  return data
}

async function getOrderStatus(id: number): Promise<OrderStatusResponse> {
  const { data } = await apiClient.get(`/payment/orders/${id}/status`)
  return data
}

export const paymentAPI = {
  listPlans,
  createOrder,
  listOrders,
  getOrderStatus
}
