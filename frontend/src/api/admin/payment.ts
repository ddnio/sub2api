/**
 * Admin Payment API endpoints
 * Handles payment management operations for administrators
 */

import { apiClient } from '../client'
import type {
  DashboardStats,
  PaymentOrder,
  PaymentChannel,
  SubscriptionPlan,
  ProviderInstance
} from '@/types/payment'
import type { BasePaginationResponse, FetchOptions } from '@/types'

export type AdminPaymentPlan = SubscriptionPlan & {
  group_name?: string
  deleted_at?: string | null
}

export type AdminPaymentOrder = PaymentOrder & {
  user_email?: string
  user_name?: string
  user_notes?: string
  email?: string
  admin_note?: string | null
  callback_raw?: string | null
}

export type StatsBreakdown = DashboardStats['daily_series'][number]
export type PaymentConfig = AdminPaymentConfig
export type { DashboardStats, ProviderInstance }

/** Admin-facing payment config returned by GET /admin/payment/config */
export interface AdminPaymentConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  order_timeout_minutes: number
  max_pending_orders: number
  enabled_payment_types: string[]
  balance_disabled: boolean
  balance_recharge_multiplier: number
  load_balance_strategy: string
  product_name_prefix: string
  product_name_suffix: string
  help_image_url: string
  help_text: string
}

/** Fields accepted by PUT /admin/payment/config (all optional via pointer semantics) */
export interface UpdatePaymentConfigRequest {
  enabled?: boolean
  min_amount?: number
  max_amount?: number
  daily_limit?: number
  order_timeout_minutes?: number
  max_pending_orders?: number
  enabled_payment_types?: string[]
  balance_disabled?: boolean
  balance_recharge_multiplier?: number
  load_balance_strategy?: string
  product_name_prefix?: string
  product_name_suffix?: string
  help_image_url?: string
  help_text?: string
}

export const adminPaymentAPI = {
  // ==================== Config ====================

  /** Get payment configuration (admin view) */
  getConfig() {
    return apiClient.get<AdminPaymentConfig>('/admin/payment/config')
  },

  /** Update payment configuration */
  updateConfig(data: UpdatePaymentConfigRequest) {
    return apiClient.put('/admin/payment/config', data)
  },

  // ==================== Dashboard ====================

  /** Get payment dashboard statistics */
  getDashboard(days?: number) {
    return apiClient.get<DashboardStats>('/admin/payment/dashboard', {
      params: days ? { days } : undefined
    })
  },

  // ==================== Orders ====================

  /** Get all orders (paginated, with filters) */
  getOrders(params?: {
    page?: number
    page_size?: number
    status?: string
    payment_type?: string
    user_id?: number
    keyword?: string
    start_date?: string
    end_date?: string
    order_type?: string
  }) {
    return apiClient.get<BasePaginationResponse<AdminPaymentOrder>>('/admin/payment/orders', { params }).then(normalizeOrderPage)
  },

  /** Get a specific order by ID */
  getOrder(id: number) {
    return apiClient.get<AdminPaymentOrder | { order?: AdminPaymentOrder; auditLogs?: unknown[]; audit_logs?: unknown[] }>(`/admin/payment/orders/${id}`).then((res) => {
      const data = res.data
      if (data && typeof data === 'object' && 'order' in data && data.order) {
        return { ...res, data: { ...data, order: normalizeAdminOrder(data.order as AdminPaymentOrder) } }
      }
      return { ...res, data: normalizeAdminOrder(data as AdminPaymentOrder) }
    })
  },

  /** Cancel an order (admin) */
  cancelOrder(id: number) {
    return apiClient.post(`/admin/payment/orders/${id}/cancel`)
  },

  /** Retry recharge for a failed order */
  retryRecharge(id: number) {
    return apiClient.post(`/admin/payment/orders/${id}/retry`)
  },

  /** Process a refund */
  refundOrder(id: number, data: { amount: number; reason: string; deduct_balance?: boolean; force?: boolean }) {
    return apiClient.post(`/admin/payment/orders/${id}/refund`, data)
  },

  // ==================== Channels ====================

  /** Get all payment channels */
  getChannels() {
    return apiClient.get<PaymentChannel[]>('/admin/payment/channels')
  },

  /** Create a payment channel */
  createChannel(data: Partial<PaymentChannel>) {
    return apiClient.post<PaymentChannel>('/admin/payment/channels', data)
  },

  /** Update a payment channel */
  updateChannel(id: number, data: Partial<PaymentChannel>) {
    return apiClient.put<PaymentChannel>(`/admin/payment/channels/${id}`, data)
  },

  /** Delete a payment channel */
  deleteChannel(id: number) {
    return apiClient.delete(`/admin/payment/channels/${id}`)
  },

  // ==================== Subscription Plans ====================

  /** Get all subscription plans */
  getPlans() {
    return apiClient.get<SubscriptionPlan[]>('/admin/payment/plans')
  },

  /** Create a subscription plan */
  createPlan(data: Record<string, unknown>) {
    return apiClient.post<SubscriptionPlan>('/admin/payment/plans', data)
  },

  /** Update a subscription plan */
  updatePlan(id: number, data: Record<string, unknown>) {
    return apiClient.put<SubscriptionPlan>(`/admin/payment/plans/${id}`, data)
  },

  /** Delete a subscription plan */
  deletePlan(id: number) {
    return apiClient.delete(`/admin/payment/plans/${id}`)
  },

  // ==================== Provider Instances ====================

  /** Get all provider instances */
  getProviders() {
    return apiClient.get<ProviderInstance[]>('/admin/payment/providers')
  },

  /** Create a provider instance */
  createProvider(data: Partial<ProviderInstance>) {
    return apiClient.post<ProviderInstance>('/admin/payment/providers', data)
  },

  /** Update a provider instance */
  updateProvider(id: number, data: Partial<ProviderInstance>) {
    return apiClient.put<ProviderInstance>(`/admin/payment/providers/${id}`, data)
  },

  /** Delete a provider instance */
  deleteProvider(id: number) {
    return apiClient.delete(`/admin/payment/providers/${id}`)
  },

  // Legacy aliases kept for existing fork views that still participate in type checking.
  async listPlans(options?: FetchOptions): Promise<AdminPaymentPlan[]> {
    const res = await apiClient.get<AdminPaymentPlan[]>('/admin/payment/plans', { signal: options?.signal })
    return res.data
  },

  async listOrders(
    page: number,
    pageSize: number,
    params: { status?: string; order_type?: string },
    options?: FetchOptions
  ): Promise<BasePaginationResponse<AdminPaymentOrder>> {
    const res = await apiClient.get<BasePaginationResponse<AdminPaymentOrder>>('/admin/payment/orders', {
      params: { page, page_size: pageSize, ...params },
      signal: options?.signal
    })
    return normalizeOrderPage(res).data
  },

  retryOrder(id: number) {
    return apiClient.post(`/admin/payment/orders/${id}/retry`)
  },

  async listProviders(): Promise<ProviderInstance[]> {
    const res = await apiClient.get<ProviderInstance[]>('/admin/payment/providers')
    return res.data
  },

  async getPaymentConfig(): Promise<AdminPaymentConfig> {
    const res = await apiClient.get<AdminPaymentConfig>('/admin/payment/config')
    return res.data
  },

  updatePaymentConfig(data: UpdatePaymentConfigRequest) {
    return apiClient.put('/admin/payment/config', data)
  }
}

function normalizeOrderPage<T extends { data: BasePaginationResponse<AdminPaymentOrder> }>(res: T): T {
  return {
    ...res,
    data: {
      ...res.data,
      items: (res.data.items || []).map(normalizeAdminOrder)
    }
  }
}

function normalizeAdminOrder(order: AdminPaymentOrder): AdminPaymentOrder {
  return {
    ...order,
    email: order.email || order.user_email || order.user_name || '',
    amount: Number(order.amount || 0),
    pay_amount: Number(order.pay_amount || 0),
    fee_rate: Number(order.fee_rate || 0),
    refund_amount: Number(order.refund_amount || 0)
  }
}

export default adminPaymentAPI
