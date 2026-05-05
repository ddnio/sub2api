/**
 * User API endpoints
 * Handles user profile management and password changes
 */

import { apiClient } from './client'
import {
  resolveWeChatOAuthStartStrict,
  prepareOAuthBindAccessTokenCookie,
<<<<<<< HEAD
  type WeChatOAuthPublicSettings
} from './auth'
import type { UserProfile, ChangePasswordRequest, UserAuthProvider } from '@/types'
=======
  type WeChatOAuthPublicSettings,
} from './auth'
import type { User, ChangePasswordRequest, NotifyEmailEntry, UserAuthProvider } from '@/types'
>>>>>>> v0.1.116

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
<<<<<<< HEAD
}): Promise<UserProfile> {
  const { data } = await apiClient.put<UserProfile>('/user', profile)
=======
  balance_notify_extra_emails?: NotifyEmailEntry[]
}): Promise<User> {
  const { data } = await apiClient.put<User>('/user', profile)
>>>>>>> v0.1.116
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

<<<<<<< HEAD
=======
/**
 * Send verification code for adding a notify email
 * @param email - Email address to verify
 */
>>>>>>> v0.1.116
export async function sendNotifyEmailCode(email: string): Promise<void> {
  await apiClient.post('/user/notify-email/send-code', { email })
}

<<<<<<< HEAD
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
=======
/**
 * Verify and add a notify email
 * @param email - Email address to add
 * @param code - Verification code
 */
export async function verifyNotifyEmail(email: string, code: string): Promise<void> {
  await apiClient.post('/user/notify-email/verify', { email, code })
}

/**
 * Remove a notify email
 * @param email - Email address to remove
 */
export async function removeNotifyEmail(email: string): Promise<void> {
  await apiClient.delete('/user/notify-email', { data: { email } })
}

/**
 * Toggle a notify email's disabled state
 * @param email - Email address (empty string for primary email placeholder)
 * @param disabled - Whether to disable the email
 */
export async function toggleNotifyEmail(email: string, disabled: boolean): Promise<User> {
  const { data } = await apiClient.put<User>('/user/notify-email/toggle', { email, disabled })
  return data
}

export async function sendEmailBindingCode(email: string): Promise<void> {
  await apiClient.post('/user/account-bindings/email/send-code', { email })
}

export async function bindEmailIdentity(payload: {
  email: string
  verify_code: string
  password: string
}): Promise<User> {
  const { data } = await apiClient.post<User>('/user/account-bindings/email', payload)
  return data
}

export async function unbindAuthIdentity(provider: BindableOAuthProvider): Promise<User> {
  const { data } = await apiClient.delete<User>(`/user/account-bindings/${provider}`)
>>>>>>> v0.1.116
  return data
}

export type BindableOAuthProvider = Exclude<UserAuthProvider, 'email'>

<<<<<<< HEAD
export interface StartOAuthBindingResponse {
  provider: BindableOAuthProvider
  authorize_url: string
  method: string
  use_browser_redirect: boolean
}

=======
>>>>>>> v0.1.116
interface BuildOAuthBindingStartURLOptions {
  redirectTo?: string
  wechatOAuthSettings?: WeChatOAuthPublicSettings | null
}

export function resolveWeChatOAuthMode(): 'open' | 'mp' {
  if (typeof navigator === 'undefined') {
    return 'open'
  }
  return /MicroMessenger/i.test(navigator.userAgent) ? 'mp' : 'open'
}

function resolveWeChatOAuthBindingMode(
  settings?: WeChatOAuthPublicSettings | null
): 'open' | 'mp' | null {
  if (settings) {
    return resolveWeChatOAuthStartStrict(settings).mode
  }
  return resolveWeChatOAuthMode()
}

export function buildOAuthBindingStartURL(
  provider: BindableOAuthProvider,
  options: BuildOAuthBindingStartURLOptions = {}
): string | null {
  const redirectTo = options.redirectTo?.trim() || '/profile'
<<<<<<< HEAD
  if (provider !== 'wechat') {
    return null
  }
=======
>>>>>>> v0.1.116
  const apiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) || '/api/v1'
  const normalized = apiBase.replace(/\/$/, '')
  const params = new URLSearchParams({
    redirect: redirectTo,
    intent: 'bind_current_user'
  })

<<<<<<< HEAD
  const mode = resolveWeChatOAuthBindingMode(options.wechatOAuthSettings)
  if (!mode) {
    return null
  }
  params.set('mode', mode)

  return `${normalized}/auth/oauth/${provider}/start?${params.toString()}`
=======
  if (provider === 'wechat') {
    const mode = resolveWeChatOAuthBindingMode(options.wechatOAuthSettings)
    if (!mode) {
      return null
    }
    params.set('mode', mode)
  }

  return `${normalized}/auth/oauth/${provider}/bind/start?${params.toString()}`
>>>>>>> v0.1.116
}

export async function startOAuthBinding(
  provider: BindableOAuthProvider,
<<<<<<< HEAD
  options: BuildOAuthBindingStartURLOptions | string = {}
=======
  options: BuildOAuthBindingStartURLOptions = {}
>>>>>>> v0.1.116
): Promise<void> {
  if (typeof window === 'undefined') {
    return
  }
<<<<<<< HEAD

  const normalizedOptions =
    typeof options === 'string' ? { redirectTo: options } : options
  const redirectTo = normalizedOptions.redirectTo?.trim() || '/profile'

  prepareOAuthBindAccessTokenCookie()
  if (provider === 'wechat') {
    const startURL = buildOAuthBindingStartURL(provider, normalizedOptions)
    if (!startURL) {
      return
    }
    window.location.href = startURL
    return
  }

  const { data } = await apiClient.post<StartOAuthBindingResponse>('/user/auth-identities/bind/start', {
    provider,
    redirect_to: redirectTo,
  })
  window.location.href = data.authorize_url
}

export async function unbindAuthProvider(provider: Exclude<UserAuthProvider, 'email'>): Promise<UserProfile> {
  const { data } = await apiClient.delete<UserProfile>(`/user/account-bindings/${provider}`)
  return data
=======
  const startURL = buildOAuthBindingStartURL(provider, options)
  if (!startURL) {
    return
  }
  await prepareOAuthBindAccessTokenCookie()
  window.location.href = startURL
>>>>>>> v0.1.116
}

export const userAPI = {
  getProfile,
  updateProfile,
  changePassword,
  sendNotifyEmailCode,
  verifyNotifyEmail,
  removeNotifyEmail,
  toggleNotifyEmail,
<<<<<<< HEAD
  startOAuthBinding,
  unbindAuthProvider
=======
  sendEmailBindingCode,
  bindEmailIdentity,
  unbindAuthIdentity,
  buildOAuthBindingStartURL,
  startOAuthBinding
>>>>>>> v0.1.116
}

export default userAPI
