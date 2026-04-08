/**
 * Referral API endpoints
 */

import { apiClient } from './client'

export interface ReferralInfo {
  referral_code: string
  total_invited: number
  total_rewarded: number
  inviter_reward_amount: number
  invitee_reward_amount: number
}

export interface ReferralRecord {
  id: number
  inviter_id: number
  invitee_id: number
  invitee_email: string
  code: string
  inviter_rewarded: number
  invitee_rewarded: number
  created_at: string
}

export interface ReferralListResponse {
  data: ReferralRecord[]
  pagination: {
    total: number
    page: number
    page_size: number
    pages: number
  }
}

export async function getReferralInfo(): Promise<ReferralInfo> {
  const { data } = await apiClient.get<ReferralInfo>('/referral')
  return data
}

export async function getReferralList(page = 1, pageSize = 20): Promise<ReferralListResponse> {
  const { data } = await apiClient.get<ReferralListResponse>('/referral/list', {
    params: { page, page_size: pageSize }
  })
  return data
}
