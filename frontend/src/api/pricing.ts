import { apiClient } from './client'
import type { ModelPricingResponse } from '@/types'

export async function getModelPricing(groupId?: number): Promise<ModelPricingResponse> {
  const params: Record<string, string> = {}
  if (groupId !== undefined) {
    params.group_id = String(groupId)
  }
  const { data } = await apiClient.get<ModelPricingResponse>('/pricing/models', { params })
  return data
}

export const pricingAPI = {
  getModelPricing,
}

export default pricingAPI
