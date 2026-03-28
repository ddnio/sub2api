import { apiClient } from '../client'
import type { BasePaginationResponse, FetchOptions } from '@/types'
import type { PaymentPlan, PaymentOrder } from '../payment'

export interface AdminPaymentPlan extends PaymentPlan {
  deleted_at: string | null
}

export interface AdminPaymentOrder extends PaymentOrder {
  user_id: number
  email: string
  admin_note: string | null
  refunded_at: string | null
  callback_raw: string | null
}

export interface OrderStats {
  total_orders: number
  total_amount: number
  paid_orders: number
  paid_amount: number
  completed_orders: number
  completed_amount: number
  breakdown: StatsBreakdown[]
}

export interface StatsBreakdown {
  date: string
  count: number
  amount: number
}

async function listPlans(
  page: number,
  pageSize: number,
  _params: Record<string, any>,
  options?: FetchOptions
): Promise<BasePaginationResponse<AdminPaymentPlan>> {
  const { data } = await apiClient.get('/admin/payment/plans', {
    params: { page, page_size: pageSize },
    signal: options?.signal
  })
  return data
}

async function createPlan(plan: {
  name: string
  description?: string
  badge?: string | null
  group_id: number
  duration_days: number
  price: number
  original_price?: number | null
  sort_order?: number
  is_active?: boolean
}): Promise<AdminPaymentPlan> {
  const { data } = await apiClient.post('/admin/payment/plans', plan)
  return data
}

async function updatePlan(id: number, updates: Record<string, any>): Promise<AdminPaymentPlan> {
  const { data } = await apiClient.put(`/admin/payment/plans/${id}`, updates)
  return data
}

async function deletePlan(id: number): Promise<void> {
  await apiClient.delete(`/admin/payment/plans/${id}`)
}

async function listOrders(
  page: number,
  pageSize: number,
  params: { status?: string; type?: string },
  options?: FetchOptions
): Promise<BasePaginationResponse<AdminPaymentOrder>> {
  const { data } = await apiClient.get('/admin/payment/orders', {
    params: { page, page_size: pageSize, ...params },
    signal: options?.signal
  })
  return data
}

async function getOrder(id: number): Promise<AdminPaymentOrder> {
  const { data } = await apiClient.get(`/admin/payment/orders/${id}`)
  return data
}

async function completeOrder(id: number, adminNote?: string): Promise<void> {
  await apiClient.post(`/admin/payment/orders/${id}/complete`, { admin_note: adminNote })
}

async function refundOrder(id: number, adminNote?: string): Promise<void> {
  await apiClient.post(`/admin/payment/orders/${id}/refund`, { admin_note: adminNote })
}

async function getOrderStats(params: {
  start_date?: string
  end_date?: string
  group_by?: string
}): Promise<OrderStats> {
  const { data } = await apiClient.get('/admin/payment/orders/stats', { params })
  return data
}

const adminPaymentAPI = {
  listPlans,
  createPlan,
  updatePlan,
  deletePlan,
  listOrders,
  getOrder,
  completeOrder,
  refundOrder,
  getOrderStats
}

export default adminPaymentAPI
