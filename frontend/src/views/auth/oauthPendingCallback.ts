import type { PendingOAuthExchangeResponse } from '@/api/auth'
import { exchangePendingOAuthCompletion } from '@/api/auth'

export function parseFragmentParams(hash: string): URLSearchParams {
  const raw = hash.startsWith('#') ? hash.slice(1) : hash
  return new URLSearchParams(raw)
}

export function sanitizeRedirectPath(path: string | null | undefined): string {
  if (!path) return '/dashboard'
  if (!path.startsWith('/')) return '/dashboard'
  if (path.startsWith('//')) return '/dashboard'
  if (path.includes('://')) return '/dashboard'
  if (path.includes('\n') || path.includes('\r')) return '/dashboard'
  return path
}

function stringParam(value: FormDataEntryValue | string | null | undefined): string {
  return typeof value === 'string' ? value : ''
}

export function pendingOAuthPayloadFromFragment(params: URLSearchParams): PendingOAuthExchangeResponse {
  const error = params.get('error') || undefined
  return {
    access_token: params.get('access_token') || undefined,
    refresh_token: params.get('refresh_token') || undefined,
    expires_in: params.get('expires_in') ? Number(params.get('expires_in')) : undefined,
    token_type: params.get('token_type') || undefined,
    redirect: params.get('redirect') || undefined,
    error,
  }
}

export function legacyPendingOAuthTokenFromFragment(params: URLSearchParams): string {
  return stringParam(params.get('pending_oauth_token')).trim()
}

export function isInvitationRequired(payload: PendingOAuthExchangeResponse): boolean {
  return payload.error === 'invitation_required'
}

export function hasAuthTokenPayload(payload: PendingOAuthExchangeResponse): boolean {
  return Boolean(payload.access_token)
}

export function isOAuthBindCompletion(payload: PendingOAuthExchangeResponse): boolean {
	const intent = (payload.intent || '').trim().toLowerCase()
	return intent === 'bind_current_user'
}

export async function resolvePendingOAuthPayload(params: URLSearchParams): Promise<PendingOAuthExchangeResponse> {
  if (params.get('pending_oauth_token')) {
    return pendingOAuthPayloadFromFragment(params)
  }

  try {
    return await exchangePendingOAuthCompletion()
  } catch {
    return pendingOAuthPayloadFromFragment(params)
  }
}
