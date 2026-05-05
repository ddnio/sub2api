import { beforeEach, describe, expect, it, vi } from 'vitest'

const post = vi.fn()

vi.mock('@/api/client', () => ({
  apiClient: {
    post
  }
}))

describe('oauth adoption auth api', () => {
  beforeEach(() => {
    post.mockReset()
    post.mockResolvedValue({ data: {} })
    localStorage.clear()
    document.cookie = 'oauth_bind_access_token=; Max-Age=0; path=/'
  })

  it('posts adoption decisions when exchanging pending oauth completion', async () => {
    const { exchangePendingOAuthCompletion } = await import('@/api/auth')

    await exchangePendingOAuthCompletion({
      adoptDisplayName: false,
      adoptAvatar: true
    })

    expect(post).toHaveBeenCalledWith('/auth/oauth/pending/exchange', {
      adopt_display_name: false,
      adopt_avatar: true
    })
  })

  it('posts bind-login decisions when finalizing pending oauth bind flow', async () => {
    const { completePendingOAuthBindLogin } = await import('@/api/auth')

    await completePendingOAuthBindLogin({
      adoptDisplayName: true,
      adoptAvatar: false
    })

    expect(post).toHaveBeenCalledWith('/auth/oauth/pending/exchange', {
      adopt_display_name: true,
      adopt_avatar: false
    })
  })

<<<<<<< HEAD
  it('keeps linuxdo legacy invitation completion token-based', async () => {
    const { completeLinuxDoOAuthRegistration } = await import('@/api/auth')

    await completeLinuxDoOAuthRegistration('pending-token', 'invite-code')

    expect(post).toHaveBeenCalledWith('/auth/oauth/linuxdo/complete-registration', {
      pending_oauth_token: 'pending-token',
      invitation_code: 'invite-code'
    })
  })

  it('posts generic linuxdo create-account completion with adoption decisions', async () => {
    const { createPendingOAuthAccount } = await import('@/api/auth')

    await createPendingOAuthAccount({
      email: 'linuxdo@example.com',
      password: 'secret-123',
      invitation_code: 'invite-code',
      adopt_display_name: false,
      adopt_avatar: true
    })

    expect(post).toHaveBeenCalledWith('/auth/oauth/pending/create-account', {
      email: 'linuxdo@example.com',
      password: 'secret-123',
      invitation_code: 'invite-code',
      adopt_display_name: false,
      adopt_avatar: true
    })
  })

  it('keeps oidc legacy invitation completion token-based', async () => {
    const { completeOIDCOAuthRegistration } = await import('@/api/auth')

    await completeOIDCOAuthRegistration('pending-token', 'invite-code')

    expect(post).toHaveBeenCalledWith('/auth/oauth/oidc/complete-registration', {
      pending_oauth_token: 'pending-token',
      invitation_code: 'invite-code'
    })
  })

  it('posts generic oidc create-account completion with adoption decisions', async () => {
    const { createPendingOAuthAccount } = await import('@/api/auth')

    await createPendingOAuthAccount({
      email: 'oidc@example.com',
      password: 'secret-123',
=======
  it('posts linuxdo invitation completion with adoption decisions', async () => {
    const { completeLinuxDoOAuthRegistration } = await import('@/api/auth')

    await completeLinuxDoOAuthRegistration('invite-code', {
      adoptDisplayName: true,
      adoptAvatar: false
    })

    expect(post).toHaveBeenCalledWith('/auth/oauth/linuxdo/complete-registration', {
>>>>>>> v0.1.116
      invitation_code: 'invite-code',
      adopt_display_name: true,
      adopt_avatar: false
    })
<<<<<<< HEAD

    expect(post).toHaveBeenCalledWith('/auth/oauth/pending/create-account', {
      email: 'oidc@example.com',
      password: 'secret-123',
=======
  })

  it('posts linuxdo create-account completion with adoption decisions', async () => {
    const { createPendingLinuxDoOAuthAccount } = await import('@/api/auth')

    await createPendingLinuxDoOAuthAccount('invite-code', {
      adoptDisplayName: false,
      adoptAvatar: true
    })

    expect(post).toHaveBeenCalledWith('/auth/oauth/linuxdo/complete-registration', {
      invitation_code: 'invite-code',
      adopt_display_name: false,
      adopt_avatar: true
    })
  })

  it('posts oidc invitation completion with adoption decisions', async () => {
    const { completeOIDCOAuthRegistration } = await import('@/api/auth')

    await completeOIDCOAuthRegistration('invite-code', {
      adoptDisplayName: false,
      adoptAvatar: true
    })

    expect(post).toHaveBeenCalledWith('/auth/oauth/oidc/complete-registration', {
      invitation_code: 'invite-code',
      adopt_display_name: false,
      adopt_avatar: true
    })
  })

  it('posts oidc create-account completion with adoption decisions', async () => {
    const { createPendingOIDCOAuthAccount } = await import('@/api/auth')

    await createPendingOIDCOAuthAccount('invite-code', {
      adoptDisplayName: true,
      adoptAvatar: false
    })

    expect(post).toHaveBeenCalledWith('/auth/oauth/oidc/complete-registration', {
>>>>>>> v0.1.116
      invitation_code: 'invite-code',
      adopt_display_name: true,
      adopt_avatar: false
    })
  })

  it('posts wechat invitation completion with adoption decisions', async () => {
    const { completeWeChatOAuthRegistration } = await import('@/api/auth')

    await completeWeChatOAuthRegistration('invite-code', {
      adoptDisplayName: true,
      adoptAvatar: true
    })

    expect(post).toHaveBeenCalledWith('/auth/oauth/wechat/complete-registration', {
      invitation_code: 'invite-code',
      adopt_display_name: true,
      adopt_avatar: true
    })
  })

  it('posts wechat create-account completion with adoption decisions', async () => {
    const { createPendingWeChatOAuthAccount } = await import('@/api/auth')

    await createPendingWeChatOAuthAccount('invite-code', {
      adoptDisplayName: false,
      adoptAvatar: false
    })

    expect(post).toHaveBeenCalledWith('/auth/oauth/wechat/complete-registration', {
      invitation_code: 'invite-code',
      adopt_display_name: false,
      adopt_avatar: false
    })
  })

  it('classifies oauth completion results as login or bind', async () => {
    const { getOAuthCompletionKind } = await import('@/api/auth')

    expect(getOAuthCompletionKind({ access_token: 'access-token' })).toBe('login')
    expect(getOAuthCompletionKind({ redirect: '/profile' })).toBe('bind')
  })

  it('provides bind-login utility helpers for invitation and suggested profile states', async () => {
    const {
      getPendingOAuthBindLoginKind,
      hasPendingOAuthSuggestedProfile,
      isPendingOAuthCreateAccountRequired
    } = await import('@/api/auth')

    expect(getPendingOAuthBindLoginKind({ access_token: 'access-token' })).toBe('login')
    expect(getPendingOAuthBindLoginKind({ redirect: '/profile' })).toBe('bind')
    expect(
      isPendingOAuthCreateAccountRequired({
        error: 'invitation_required'
      })
    ).toBe(true)
    expect(
      isPendingOAuthCreateAccountRequired({
        error: 'other'
      })
    ).toBe(false)
    expect(
      hasPendingOAuthSuggestedProfile({
        suggested_display_name: 'OAuth Nick'
      })
    ).toBe(true)
    expect(
      hasPendingOAuthSuggestedProfile({
        suggested_avatar_url: 'https://cdn.example/avatar.png'
      })
    ).toBe(true)
    expect(hasPendingOAuthSuggestedProfile({})).toBe(false)
  })

<<<<<<< HEAD
  it('prepares an oauth bind access token cookie before redirect binding', async () => {
    localStorage.setItem('auth_token', 'access-token-value')
    const setCookie = vi.fn()
    Object.defineProperty(document, 'cookie', {
      configurable: true,
      get: () => '',
      set: setCookie
    })

    const { prepareOAuthBindAccessTokenCookie } = await import('@/api/auth')

    prepareOAuthBindAccessTokenCookie()

    expect(setCookie).toHaveBeenCalledTimes(1)
    expect(setCookie.mock.calls[0]?.[0]).toContain('oauth_bind_access_token=access-token-value')
=======
  it('requests an HttpOnly oauth bind cookie before redirect binding', async () => {
    localStorage.setItem('auth_token', 'access-token-value')
    const { prepareOAuthBindAccessTokenCookie } = await import('@/api/auth')

    await prepareOAuthBindAccessTokenCookie()

    expect(post).toHaveBeenCalledWith('/auth/oauth/bind-token')
>>>>>>> v0.1.116
  })
})
