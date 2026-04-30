import { apiClient } from './client'
import type { BasePaginationResponse, FetchOptions } from '@/types'

export interface PaymentPlan {
  id: number
  group_id: number
  group_platform: string
  name: string
  description: string
  price: number
  original_price: number | null
  validity_days: number
  validity_unit: string
  features: string
  product_name: string
  for_sale: boolean
  sort_order: number
}

export interface PaymentOrder {
  id: number
  out_trade_no: string
  order_type: string
  plan_id: number | null
  amount: number
  pay_amount: number
  payment_type: string
  payment_trade_no: string | null
  qr_code: string | null
  pay_url: string | null
  status: string
  paid_at: string | null
  completed_at: string | null
  expires_at: string
  created_at: string
}

export interface CreateOrderResponse {
  order_id: number
  amount: number
  pay_amount: number
  fee_rate: number
  status: string
  result_type: string
  payment_type: string
  out_trade_no: string
  pay_url: string
  qr_code: string
  expires_at: string
  payment_mode: string
}

export interface OrderStatusResponse {
  status: string
}

async function listPlans(): Promise<PaymentPlan[]> {
  const { data } = await apiClient.get('/payment/plans')
  return data
}

async function createOrder(params: {
  order_type: 'balance' | 'subscription'
  plan_id?: number
  amount?: number
  payment_type: string
}): Promise<CreateOrderResponse> {
  const { data } = await apiClient.post('/payment/orders', params)
  return data
}

async function listOrders(
  page: number,
  pageSize: number,
  params: { status?: string; order_type?: string },
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
