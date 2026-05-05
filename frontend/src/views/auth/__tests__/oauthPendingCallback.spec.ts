import { describe, expect, it, vi } from 'vitest'

const exchangePendingOAuthCompletion = vi.hoisted(() => vi.fn())

vi.mock('@/api/auth', () => ({
  exchangePendingOAuthCompletion,
}))

import {
  legacyPendingOAuthTokenFromFragment,
  parseFragmentParams,
  pendingOAuthPayloadFromFragment,
  resolvePendingOAuthPayload,
  sanitizeRedirectPath,
  isOAuthBindCompletion,
} from '../oauthPendingCallback'

describe('oauth pending callback helpers', () => {
  it('prefers legacy fragment payload when pending_oauth_token exists', async () => {
    const params = parseFragmentParams(
      '#error=invitation_required&pending_oauth_token=legacy-token&redirect=%2Fprofile'
    )

    const payload = await resolvePendingOAuthPayload(params)

    expect(exchangePendingOAuthCompletion).not.toHaveBeenCalled()
    expect(payload).toEqual({
      access_token: undefined,
      refresh_token: undefined,
      expires_in: undefined,
      token_type: undefined,
      redirect: '/profile',
      error: 'invitation_required',
    })
    expect(legacyPendingOAuthTokenFromFragment(params)).toBe('legacy-token')
  })

  it('uses browser-bound pending exchange when no legacy token is present', async () => {
    exchangePendingOAuthCompletion.mockResolvedValueOnce({
      access_token: 'access',
      refresh_token: 'refresh',
      expires_in: 3600,
      redirect: '/dashboard',
    })

    const payload = await resolvePendingOAuthPayload(parseFragmentParams(''))

    expect(exchangePendingOAuthCompletion).toHaveBeenCalledTimes(1)
    expect(payload.access_token).toBe('access')
    expect(payload.refresh_token).toBe('refresh')
    expect(payload.expires_in).toBe(3600)
    expect(payload.redirect).toBe('/dashboard')
  })

  it('falls back to fragment tokens when pending exchange is unavailable', async () => {
    exchangePendingOAuthCompletion.mockRejectedValueOnce(new Error('missing pending session'))
    const params = parseFragmentParams(
      '#access_token=access&refresh_token=refresh&expires_in=1800&redirect=%2Fkeys'
    )

    const payload = await resolvePendingOAuthPayload(params)

    expect(payload).toEqual({
      access_token: 'access',
      refresh_token: 'refresh',
      expires_in: 1800,
      token_type: undefined,
      redirect: '/keys',
      error: undefined,
    })
  })

  it('sanitizes unsafe redirect values', () => {
    expect(sanitizeRedirectPath('/dashboard')).toBe('/dashboard')
    expect(sanitizeRedirectPath('https://evil.example')).toBe('/dashboard')
    expect(sanitizeRedirectPath('//evil.example')).toBe('/dashboard')
    expect(sanitizeRedirectPath('/safe\nLocation: evil')).toBe('/dashboard')
  })

  it('converts fragment payload fields without treating missing expires as zero', () => {
    const payload = pendingOAuthPayloadFromFragment(parseFragmentParams('#access_token=access'))

    expect(payload.access_token).toBe('access')
    expect(payload.expires_in).toBeUndefined()
  })

  it('recognizes browser-bound oauth bind completion without access token', () => {
    expect(isOAuthBindCompletion({ intent: 'bind_current_user', redirect: '/profile' })).toBe(true)
    expect(isOAuthBindCompletion({ auth_result: 'pending_session', redirect: '/profile' })).toBe(false)
    expect(isOAuthBindCompletion({ redirect: '/profile' })).toBe(false)
  })
})
