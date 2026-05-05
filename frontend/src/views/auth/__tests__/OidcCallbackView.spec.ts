import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import OidcCallbackView from '../OidcCallbackView.vue'

const exchangePendingOAuthCompletion = vi.hoisted(() => vi.fn())
const createPendingOAuthAccount = vi.hoisted(() => vi.fn())
const bindPendingOAuthLogin = vi.hoisted(() => vi.fn())
const completeOIDCOAuthRegistration = vi.hoisted(() => vi.fn())
const getPublicSettings = vi.hoisted(() => vi.fn())
const login2FA = vi.hoisted(() => vi.fn())
const setToken = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const replace = vi.hoisted(() => vi.fn())

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
  useRouter: () => ({ replace }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/stores', () => ({
  useAuthStore: () => ({ setToken }),
  useAppStore: () => ({ showSuccess, showError }),
}))

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
    exchangePendingOAuthCompletion,
    createPendingOAuthAccount,
    bindPendingOAuthLogin,
    completeOIDCOAuthRegistration,
    getPublicSettings,
    login2FA,
  }
})

function mountView() {
  return mount(OidcCallbackView, {
    global: {
      stubs: {
        AuthLayout: { template: '<div><slot /></div>' },
        Icon: true,
        RouterLink: { template: '<a><slot /></a>' },
        transition: false,
      },
    },
  })
}

describe('OidcCallbackView pending OAuth flow', () => {
  beforeEach(() => {
    exchangePendingOAuthCompletion.mockReset()
    createPendingOAuthAccount.mockReset()
    bindPendingOAuthLogin.mockReset()
    completeOIDCOAuthRegistration.mockReset()
    getPublicSettings.mockReset()
    login2FA.mockReset()
    setToken.mockReset()
    showSuccess.mockReset()
    showError.mockReset()
    replace.mockReset()
    localStorage.clear()
    window.location.hash = ''
    getPublicSettings.mockResolvedValue({})
  })

  it('shows create-account form for DB-backed invitation-required pending sessions', async () => {
    exchangePendingOAuthCompletion.mockResolvedValue({
      error: 'invitation_required',
      email: 'new@example.com',
      redirect: '/dashboard',
    })

    const wrapper = mountView()

    await flushPromises()

    expect(wrapper.find('[data-testid="oidc-create-account-email"]').exists()).toBe(true)
    expect((wrapper.get('[data-testid="oidc-create-account-email"]').element as HTMLInputElement).value).toBe(
      'new@example.com'
    )
    expect(wrapper.text()).not.toContain('auth.oidc.invalidPendingToken')
  })

  it('keeps the legacy fragment invitation flow compatible', async () => {
    window.location.hash =
      '#error=invitation_required&pending_oauth_token=legacy-token&redirect=%2Flegacy'
    completeOIDCOAuthRegistration.mockResolvedValue({
      access_token: 'legacy-access',
      refresh_token: 'legacy-refresh',
      expires_in: 1800,
      token_type: 'Bearer',
    })
    setToken.mockResolvedValue({})

    const wrapper = mountView()

    await flushPromises()
    expect(wrapper.find('[data-testid="oidc-create-account-email"]').exists()).toBe(false)

    await wrapper.get('input[type="text"]').setValue('INVITE-LEGACY')
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()

    expect(completeOIDCOAuthRegistration).toHaveBeenCalledWith(
      'legacy-token',
      'INVITE-LEGACY'
    )
    expect(localStorage.getItem('refresh_token')).toBe('legacy-refresh')
    expect(setToken).toHaveBeenCalledWith('legacy-access')
    expect(replace).toHaveBeenCalledWith('/legacy')
  })

  it('creates a local account and persists returned token context', async () => {
    exchangePendingOAuthCompletion.mockResolvedValue({
      error: 'invitation_required',
      email: 'new@example.com',
      redirect: '/dashboard',
    })
    createPendingOAuthAccount.mockResolvedValue({
      access_token: 'access-token',
      refresh_token: 'refresh-token',
      expires_in: 3600,
      token_type: 'Bearer',
    })
    setToken.mockResolvedValue({})

    const wrapper = mountView()

    await flushPromises()
    await wrapper.get('[data-testid="oidc-create-account-password"]').setValue('secret-123')
    await wrapper.get('[data-testid="oidc-create-account-verify-code"]').setValue('246810')
    await wrapper.get('[data-testid="oidc-create-account-invitation-code"]').setValue('INVITE-1')
    await wrapper.get('[data-testid="oidc-create-account-submit"]').trigger('click')
    await flushPromises()

    expect(createPendingOAuthAccount).toHaveBeenCalledWith({
      email: 'new@example.com',
      password: 'secret-123',
      verify_code: '246810',
      invitation_code: 'INVITE-1',
      adopt_display_name: false,
      adopt_avatar: false,
    })
    expect(localStorage.getItem('refresh_token')).toBe('refresh-token')
    expect(setToken).toHaveBeenCalledWith('access-token')
    expect(replace).toHaveBeenCalledWith('/dashboard')
  })

  it('switches to bind-login when create-account reports an existing email', async () => {
    exchangePendingOAuthCompletion.mockResolvedValue({
      error: 'invitation_required',
      email: 'new@example.com',
      redirect: '/profile',
    })
    createPendingOAuthAccount.mockResolvedValue({
      step: 'bind_login_required',
      email: 'new@example.com',
      redirect: '/profile',
    })
    bindPendingOAuthLogin.mockResolvedValue({
      access_token: 'bound-access',
      refresh_token: 'bound-refresh',
      expires_in: 3600,
      token_type: 'Bearer',
    })
    setToken.mockResolvedValue({})

    const wrapper = mountView()

    await flushPromises()
    await wrapper.get('[data-testid="oidc-create-account-password"]').setValue('secret-123')
    await wrapper.get('[data-testid="oidc-create-account-verify-code"]').setValue('246810')
    await wrapper.get('[data-testid="oidc-create-account-submit"]').trigger('click')
    await flushPromises()

    await wrapper.get('[data-testid="oidc-bind-login-password"]').setValue('existing-password')
    await wrapper.get('[data-testid="oidc-bind-login-submit"]').trigger('click')
    await flushPromises()

    expect(bindPendingOAuthLogin).toHaveBeenCalledWith({
      email: 'new@example.com',
      password: 'existing-password',
      adopt_display_name: false,
      adopt_avatar: false,
    })
    expect(setToken).toHaveBeenCalledWith('bound-access')
    expect(replace).toHaveBeenCalledWith('/profile')
  })

  it('finishes bind-login after a pending OAuth bind 2FA challenge', async () => {
    exchangePendingOAuthCompletion.mockResolvedValue({
      error: 'invitation_required',
      email: 'new@example.com',
      redirect: '/profile',
    })
    createPendingOAuthAccount.mockResolvedValue({
      step: 'bind_login_required',
      email: 'new@example.com',
      redirect: '/profile',
    })
    bindPendingOAuthLogin.mockResolvedValue({
      requires_2fa: true,
      temp_token: 'temp-token',
      user_email_masked: 'n***@example.com',
    })
    login2FA.mockResolvedValue({
      access_token: 'bound-access',
      refresh_token: 'bound-refresh',
      expires_in: 3600,
      token_type: 'Bearer',
    })

    const wrapper = mountView()

    await flushPromises()
    await wrapper.get('[data-testid="oidc-create-account-password"]').setValue('secret-123')
    await wrapper.get('[data-testid="oidc-create-account-verify-code"]').setValue('246810')
    await wrapper.get('[data-testid="oidc-create-account-submit"]').trigger('click')
    await flushPromises()

    await wrapper.get('[data-testid="oidc-bind-login-password"]').setValue('existing-password')
    await wrapper.get('[data-testid="oidc-bind-login-submit"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="oidc-bind-login-totp"]').exists()).toBe(true)
    await wrapper.get('[data-testid="oidc-bind-login-totp"]').setValue('123456')
    await wrapper.get('[data-testid="oidc-bind-login-totp-submit"]').trigger('click')
    await flushPromises()

    expect(login2FA).toHaveBeenCalledWith({
      temp_token: 'temp-token',
      totp_code: '123456',
    })
    expect(setToken).toHaveBeenCalledWith('bound-access')
    expect(replace).toHaveBeenCalledWith('/profile')
  })
})
