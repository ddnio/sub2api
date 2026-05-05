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

  it('routes WeChat binding through the frontend callback bootstrap', async () => {
    const { startOAuthBinding } = await import('@/api/user')

    await startOAuthBinding('wechat', '/profile?tab=security')

    expect(prepareOAuthBindAccessTokenCookie).toHaveBeenCalled()
    expect(post).not.toHaveBeenCalled()
    expect(window.location.href).toBe(
      '/auth/wechat/callback?wechat_bind_existing=1&redirect=%2Fprofile%3Ftab%3Dsecurity'
    )
  })
})
