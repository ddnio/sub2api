/**
 * User API endpoints
 * Handles user profile management and password changes
 */

import { apiClient } from './client'
import { prepareOAuthBindAccessTokenCookie } from './auth'
import type { UserProfile, ChangePasswordRequest, UserAuthProvider } from '@/types'

/**
 * Get current user profile
 * @returns User profile data
 */
export async function getProfile(): Promise<UserProfile> {
  const { data } = await apiClient.get<UserProfile>('/user/profile')
  return data
}

/**
 * Update current user profile
 * @param profile - Profile data to update
 * @returns Updated user profile data
 */
export async function updateProfile(profile: {
  username?: string
  avatar_url?: string | null
  balance_notify_enabled?: boolean
  balance_notify_threshold?: number | null
}): Promise<UserProfile> {
  const { data } = await apiClient.put<UserProfile>('/user', profile)
  return data
}

/**
 * Change current user password
 * @param passwords - Old and new password
 * @returns Success message
 */
export async function changePassword(
  oldPassword: string,
  newPassword: string
): Promise<{ message: string }> {
  const payload: ChangePasswordRequest = {
    old_password: oldPassword,
    new_password: newPassword
  }

  const { data } = await apiClient.put<{ message: string }>('/user/password', payload)
  return data
}

export async function sendNotifyEmailCode(email: string): Promise<void> {
  await apiClient.post('/user/notify-email/send-code', { email })
}

export async function verifyNotifyEmail(email: string, code: string): Promise<UserProfile> {
  const { data } = await apiClient.post<UserProfile>('/user/notify-email/verify', { email, code })
  return data
}

export async function removeNotifyEmail(email: string): Promise<UserProfile> {
  const { data } = await apiClient.delete<UserProfile>('/user/notify-email', { data: { email } })
  return data
}

export async function toggleNotifyEmail(email: string, disabled: boolean): Promise<UserProfile> {
  const { data } = await apiClient.put<UserProfile>('/user/notify-email/toggle', { email, disabled })
  return data
}

export type BindableOAuthProvider = Exclude<UserAuthProvider, 'email' | 'wechat'>

export interface StartOAuthBindingResponse {
  provider: BindableOAuthProvider
  authorize_url: string
  method: string
  use_browser_redirect: boolean
}

export async function startOAuthBinding(
  provider: BindableOAuthProvider,
  redirectTo = '/profile'
): Promise<void> {
  if (typeof window === 'undefined') {
    return
  }

  prepareOAuthBindAccessTokenCookie()
  const { data } = await apiClient.post<StartOAuthBindingResponse>('/user/auth-identities/bind/start', {
    provider,
    redirect_to: redirectTo,
  })
  window.location.href = data.authorize_url
}

export const userAPI = {
  getProfile,
  updateProfile,
  changePassword,
  sendNotifyEmailCode,
  verifyNotifyEmail,
  removeNotifyEmail,
  toggleNotifyEmail,
  startOAuthBinding
}

export default userAPI
