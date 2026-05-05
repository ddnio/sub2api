import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PendingOAuthCreateAccountForm from '../PendingOAuthCreateAccountForm.vue'

<<<<<<< HEAD
const sendVerifyCode = vi.hoisted(() => vi.fn())
=======
const sendVerifyCode = vi.fn()
const sendPendingOAuthVerifyCode = vi.fn()
const getPublicSettings = vi.fn()
const showError = vi.fn()
>>>>>>> v0.1.116

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
<<<<<<< HEAD
      t: (key: string, params?: Record<string, unknown>) =>
        params?.countdown ? `${key}:${params.countdown}` : key,
    }),
=======
      t: (key: string) => key
    })
>>>>>>> v0.1.116
  }
})

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
<<<<<<< HEAD
    sendVerifyCode,
  }
})

describe('PendingOAuthCreateAccountForm', () => {
  beforeEach(() => {
    sendVerifyCode.mockReset()
  })

  it('emits trimmed email, password, verify code, and invitation code', async () => {
    const wrapper = mount(PendingOAuthCreateAccountForm, {
      props: {
        testIdPrefix: 'linuxdo',
        initialEmail: 'prefill@example.com',
        isSubmitting: false,
        showInvitationCode: true,
      },
=======
    sendVerifyCode: (...args: any[]) => sendVerifyCode(...args),
    sendPendingOAuthVerifyCode: (...args: any[]) => sendPendingOAuthVerifyCode(...args),
    getPublicSettings: (...args: any[]) => getPublicSettings(...args)
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError
  })
}))

describe('PendingOAuthCreateAccountForm', () => {
  beforeEach(() => {
    sendVerifyCode.mockReset()
    sendPendingOAuthVerifyCode.mockReset()
    getPublicSettings.mockReset()
    showError.mockReset()
    getPublicSettings.mockResolvedValue({
      turnstile_enabled: false,
      turnstile_site_key: ''
    })
  })

  it('emits trimmed email, password, and verify code on submit', async () => {
    const wrapper = mount(PendingOAuthCreateAccountForm, {
      props: {
        providerName: 'LinuxDo',
        testIdPrefix: 'linuxdo',
        initialEmail: 'prefill@example.com',
        isSubmitting: false
      }
>>>>>>> v0.1.116
    })

    await wrapper.get('[data-testid="linuxdo-create-account-email"]').setValue('  user@example.com  ')
    await wrapper.get('[data-testid="linuxdo-create-account-password"]').setValue('secret-123')
    await wrapper.get('[data-testid="linuxdo-create-account-verify-code"]').setValue(' 246810 ')
<<<<<<< HEAD
    await wrapper.get('[data-testid="linuxdo-create-account-invitation-code"]').setValue(' INVITE-1 ')
=======
>>>>>>> v0.1.116
    await wrapper.get('form').trigger('submit.prevent')

    expect(wrapper.emitted('submit')).toEqual([
      [
        {
          email: 'user@example.com',
          password: 'secret-123',
<<<<<<< HEAD
          verifyCode: '246810',
          invitationCode: 'INVITE-1',
        },
      ],
    ])
  })

  it('sends a verify code for the trimmed email value', async () => {
    sendVerifyCode.mockResolvedValue({
      countdown: 60,
=======
          verifyCode: '246810'
        }
      ]
    ])
  })

  it('renders action labels through i18n keys', () => {
    const wrapper = mount(PendingOAuthCreateAccountForm, {
      props: {
        testIdPrefix: 'linuxdo',
        initialEmail: '',
        isSubmitting: false
      }
    })

    expect(wrapper.text()).toContain('auth.createAccount')
    expect(wrapper.text()).toContain('auth.alreadyHaveAccount')
  })

  it('shows and emits invitation code when invitation-only signup is enabled', async () => {
    getPublicSettings.mockResolvedValue({
      invitation_code_enabled: true,
      turnstile_enabled: false,
      turnstile_site_key: ''
>>>>>>> v0.1.116
    })

    const wrapper = mount(PendingOAuthCreateAccountForm, {
      props: {
<<<<<<< HEAD
        testIdPrefix: 'oidc',
        initialEmail: '',
        isSubmitting: false,
      },
    })

    await wrapper.get('[data-testid="oidc-create-account-email"]').setValue('  user@example.com  ')
    await wrapper.get('[data-testid="oidc-create-account-send-code"]').trigger('click')
    await flushPromises()

    expect(sendVerifyCode).toHaveBeenCalledWith({
      email: 'user@example.com',
    })
    expect(wrapper.text()).toContain('auth.codeSentSuccess')
=======
        providerName: 'LinuxDo',
        testIdPrefix: 'linuxdo',
        initialEmail: 'prefill@example.com',
        isSubmitting: false
      }
    })

    await flushPromises()
    await wrapper.get('[data-testid="linuxdo-create-account-password"]').setValue('secret-123')
    await wrapper.get('[data-testid="linuxdo-create-account-verify-code"]').setValue('246810')
    await wrapper.get('[data-testid="linuxdo-create-account-invitation-code"]').setValue(' INVITE123 ')
    await wrapper.get('form').trigger('submit.prevent')

    expect(wrapper.emitted('submit')).toEqual([
      [
        {
          email: 'prefill@example.com',
          password: 'secret-123',
          verifyCode: '246810',
          invitationCode: 'INVITE123'
        }
      ]
    ])
  })

  it('sends a verify code for the trimmed email value', async () => {
    sendPendingOAuthVerifyCode.mockResolvedValue({
      message: 'sent',
      countdown: 60
    })

    const wrapper = mount(PendingOAuthCreateAccountForm, {
      props: {
        providerName: 'LinuxDo',
        testIdPrefix: 'linuxdo',
        initialEmail: '',
        isSubmitting: false
      }
    })

    await wrapper.get('[data-testid="linuxdo-create-account-email"]').setValue('  user@example.com  ')
    await wrapper.get('[data-testid="linuxdo-create-account-send-code"]').trigger('click')
    await flushPromises()

    expect(sendPendingOAuthVerifyCode).toHaveBeenCalledWith({
      email: 'user@example.com'
    })
  })

  it('shows send-code failures via toast without rendering inline error text', async () => {
    sendPendingOAuthVerifyCode.mockRejectedValue(new Error('send failed'))

    const wrapper = mount(PendingOAuthCreateAccountForm, {
      props: {
        testIdPrefix: 'linuxdo',
        initialEmail: '',
        isSubmitting: false
      }
    })

    await wrapper.get('[data-testid="linuxdo-create-account-email"]').setValue('user@example.com')
    await wrapper.get('[data-testid="linuxdo-create-account-send-code"]').trigger('click')
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('send failed')
    expect(wrapper.text()).not.toContain('send failed')
  })

  it('requires a turnstile token before sending a verify code when turnstile is enabled', async () => {
    getPublicSettings.mockResolvedValue({
      turnstile_enabled: true,
      turnstile_site_key: 'site-key'
    })
    sendPendingOAuthVerifyCode.mockResolvedValue({
      message: 'sent',
      countdown: 60
    })

    const wrapper = mount(PendingOAuthCreateAccountForm, {
      props: {
        providerName: 'LinuxDo',
        testIdPrefix: 'linuxdo',
        initialEmail: '',
        isSubmitting: false
      },
      global: {
        stubs: {
          TurnstileWidget: {
            template: '<button data-testid="turnstile-verify" @click="$emit(\'verify\', \'turnstile-token\')">verify</button>'
          }
        }
      }
    })

    await flushPromises()
    await wrapper.get('[data-testid="linuxdo-create-account-email"]').setValue('  user@example.com  ')

    expect(wrapper.get('[data-testid="linuxdo-create-account-send-code"]').attributes('disabled')).toBeDefined()

    await wrapper.get('[data-testid="turnstile-verify"]').trigger('click')
    await wrapper.get('[data-testid="linuxdo-create-account-send-code"]').trigger('click')
    await flushPromises()

    expect(sendPendingOAuthVerifyCode).toHaveBeenCalledWith({
      email: 'user@example.com',
      turnstile_token: 'turnstile-token'
    })
>>>>>>> v0.1.116
  })
})
