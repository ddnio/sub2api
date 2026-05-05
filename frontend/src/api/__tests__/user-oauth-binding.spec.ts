import { beforeEach, describe, expect, it, vi } from 'vitest'

const post = vi.fn()
const prepareOAuthBindAccessTokenCookie = vi.fn()

vi.mock('@/api/client', () => ({
  apiClient: {
    post
  }
}))

vi.mock('@/api/auth', () => ({
  prepareOAuthBindAccessTokenCookie
}))

describe('user oauth binding api', () => {
  beforeEach(() => {
    post.mockReset()
    prepareOAuthBindAccessTokenCookie.mockReset()
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { href: 'http://localhost/profile' }
    })
  })

  it('routes WeChat binding through the configured backend OAuth start endpoint', async () => {
    const { startOAuthBinding } = await import('@/api/user')

    await startOAuthBinding('wechat', '/profile?tab=security')

    expect(prepareOAuthBindAccessTokenCookie).toHaveBeenCalled()
    expect(post).not.toHaveBeenCalled()
    expect(window.location.href).toBe(
      '/api/v1/auth/oauth/wechat/start?redirect=%2Fprofile%3Ftab%3Dsecurity&intent=bind_current_user&mode=open'
    )
  })
})
