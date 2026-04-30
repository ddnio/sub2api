import { apiClient } from '../client'
import type { BasePaginationResponse, FetchOptions } from '@/types'
import type { PaymentPlan, PaymentOrder } from '../payment'

export interface AdminPaymentPlan extends PaymentPlan {
  group_name: string
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

export interface ProviderInstance {
  id: number
  provider_key: string
  name: string
  config: Record<string, string>
  supported_types: string[]
  limits: string
  enabled: boolean
  refund_enabled: boolean
  allow_user_refund: boolean
  sort_order: number
  payment_mode: string
}

export interface PaymentConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  order_timeout_minutes: number
  max_pending_orders: number
  enabled_payment_types: string[]
  balance_disabled: boolean
  balance_recharge_multiplier: number
  recharge_fee_rate: number
  load_balance_strategy: string
  product_name_prefix: string
  product_name_suffix: string
  help_image_url: string
  help_text: string
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
  group_id: number
  validity_days: number
  validity_unit?: string
  price: number
  original_price?: number | null
  features?: string
  product_name?: string
  for_sale?: boolean
  sort_order?: number
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
  params: { status?: string; order_type?: string },
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

async function listProviders(): Promise<ProviderInstance[]> {
  const { data } = await apiClient.get('/admin/payment/providers')
  return data
}

async function createProvider(req: {
  provider_key: string
  name: string
  config: Record<string, string>
  supported_types?: string[]
  enabled?: boolean
  payment_mode?: string
  sort_order?: number
  refund_enabled?: boolean
  allow_user_refund?: boolean
}): Promise<ProviderInstance> {
  const { data } = await apiClient.post('/admin/payment/providers', req)
  return data
}

async function updateProvider(id: number, updates: {
  name?: string
  config?: Record<string, string>
  supported_types?: string[]
  enabled?: boolean
  payment_mode?: string
  sort_order?: number
  refund_enabled?: boolean
  allow_user_refund?: boolean
}): Promise<ProviderInstance> {
  const { data } = await apiClient.put(`/admin/payment/providers/${id}`, updates)
  return data
}

async function deleteProvider(id: number): Promise<void> {
  await apiClient.delete(`/admin/payment/providers/${id}`)
}

async function getPaymentConfig(): Promise<PaymentConfig> {
  const { data } = await apiClient.get('/admin/payment/config')
  return data
}

async function updatePaymentConfig(updates: Partial<PaymentConfig>): Promise<void> {
  await apiClient.put('/admin/payment/config', updates)
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
  getOrderStats,
  listProviders,
  createProvider,
  updateProvider,
  deleteProvider,
  getPaymentConfig,
  updatePaymentConfig
}

export default adminPaymentAPI
