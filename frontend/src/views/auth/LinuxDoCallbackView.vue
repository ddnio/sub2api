<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('auth.linuxdo.callbackTitle') }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ isProcessing ? t('auth.linuxdo.callbackProcessing') : t('auth.linuxdo.callbackHint') }}
        </p>
      </div>

      <transition name="fade">
        <div v-if="needsInvitation || needsCreateAccount || needsBindLogin" class="space-y-4">
          <template v-if="needsInvitation">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ t('auth.linuxdo.invitationRequired') }}
            </p>
            <div>
              <input
                v-model="invitationCode"
                type="text"
                class="input w-full"
                :placeholder="t('auth.invitationCodePlaceholder')"
                :disabled="isSubmitting"
                @keyup.enter="handleSubmitInvitation"
              />
            </div>
            <transition name="fade">
              <p v-if="invitationError" class="text-sm text-red-600 dark:text-red-400">
                {{ invitationError }}
              </p>
            </transition>
            <button
              class="btn btn-primary w-full"
              :disabled="isSubmitting || !invitationCode.trim()"
              @click="handleSubmitInvitation"
            >
              {{ isSubmitting ? t('auth.linuxdo.completing') : t('auth.linuxdo.completeRegistration') }}
            </button>
          </template>

          <template v-else-if="needsCreateAccount">
            <p v-if="pendingAccountRequiresInvitation" class="text-sm text-gray-700 dark:text-gray-300">
              {{ t('auth.linuxdo.invitationRequired') }}
            </p>
            <PendingOAuthCreateAccountForm
              test-id-prefix="linuxdo"
              :initial-email="pendingAccountEmail"
              :is-submitting="isSubmitting"
              :error-message="accountActionError"
              :show-invitation-code="pendingAccountRequiresInvitation"
              @submit="handleCreateAccount"
              @switch-to-bind="switchToBindLoginMode"
            />
          </template>

          <template v-else-if="needsBindLogin">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ t('auth.linuxdo.bindLoginRequired') }}
            </p>
            <div class="space-y-3">
              <input
                v-model="bindLoginEmail"
                data-testid="linuxdo-bind-login-email"
                type="email"
                class="input w-full"
                placeholder="you@example.com"
                :disabled="isSubmitting"
                @keyup.enter="handleBindLogin"
              />
              <input
                v-model="bindLoginPassword"
                data-testid="linuxdo-bind-login-password"
                type="password"
                class="input w-full"
                placeholder="Password"
                :disabled="isSubmitting"
                @keyup.enter="handleBindLogin"
              />
              <button
                data-testid="linuxdo-bind-login-submit"
                class="btn btn-primary w-full"
                :disabled="isSubmitting || !bindLoginEmail.trim() || !bindLoginPassword"
                @click="handleBindLogin"
              >
                {{ isSubmitting ? t('common.processing') : t('auth.linuxdo.bindLoginSubmit') }}
              </button>
              <button
                class="btn btn-secondary w-full"
                :disabled="isSubmitting"
                @click="switchToCreateAccountMode"
              >
                {{ t('auth.linuxdo.useDifferentEmail') }}
              </button>
            </div>
            <transition name="fade">
              <p v-if="accountActionError" class="text-sm text-red-600 dark:text-red-400">
                {{ accountActionError }}
              </p>
            </transition>
          </template>
        </div>
      </transition>

      <transition name="fade">
        <div
          v-if="errorMessage"
          class="rounded-xl border border-red-200 bg-red-50 p-4 dark:border-red-800/50 dark:bg-red-900/20"
        >
          <div class="flex items-start gap-3">
            <div class="flex-shrink-0">
              <Icon name="exclamationCircle" size="md" class="text-red-500" />
            </div>
            <div class="space-y-2">
              <p class="text-sm text-red-700 dark:text-red-400">
                {{ errorMessage }}
              </p>
              <router-link to="/login" class="btn btn-primary">
                {{ t('auth.linuxdo.backToLogin') }}
              </router-link>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import PendingOAuthCreateAccountForm, {
  type PendingOAuthCreateAccountPayload
} from '@/components/auth/PendingOAuthCreateAccountForm.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore, useAppStore } from '@/stores'
import {
  bindPendingOAuthLogin,
  completeLinuxDoOAuthRegistration,
  createPendingOAuthAccount,
  type PendingOAuthExchangeResponse,
  type PendingOAuthTokenPairResponse,
} from '@/api/auth'
import {
  hasAuthTokenPayload,
  isInvitationRequired,
  legacyPendingOAuthTokenFromFragment,
  parseFragmentParams,
  resolvePendingOAuthPayload,
  sanitizeRedirectPath,
} from './oauthPendingCallback'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const isProcessing = ref(true)
const errorMessage = ref('')

// Invitation code flow state
const needsInvitation = ref(false)
const pendingOAuthToken = ref('')
const invitationCode = ref('')
const isSubmitting = ref(false)
const invitationError = ref('')
const redirectTo = ref('/dashboard')
const pendingAccountAction = ref<'none' | 'create_account' | 'bind_login'>('none')
const pendingAccountEmail = ref('')
const pendingAccountRequiresInvitation = ref(false)
const bindLoginEmail = ref('')
const bindLoginPassword = ref('')
const accountActionError = ref('')

const needsCreateAccount = computed(() => pendingAccountAction.value === 'create_account')
const needsBindLogin = computed(() => pendingAccountAction.value === 'bind_login')

type PendingAccountResponse = PendingOAuthExchangeResponse | PendingOAuthTokenPairResponse

function getPendingAccountEmail(payload: PendingOAuthExchangeResponse): string {
  return (payload.email || '').trim()
}

function resolvePendingAccountAction(payload: PendingOAuthExchangeResponse): 'none' | 'create_account' | 'bind_login' {
  const raw = (payload.step || payload.error || '').trim().toLowerCase()
  if (raw === 'invitation_required' || raw === 'email_required' || raw === 'create_account_required' || raw === 'create_account') {
    return 'create_account'
  }
  if (raw === 'bind_login_required' || raw === 'bind_login') {
    return 'bind_login'
  }
  return 'none'
}

function applyPendingAccountAction(payload: PendingOAuthExchangeResponse) {
  const action = resolvePendingAccountAction(payload)
  pendingAccountAction.value = action
  pendingAccountRequiresInvitation.value = payload.error === 'invitation_required'
  accountActionError.value = ''

  const email = getPendingAccountEmail(payload)
  if (action === 'create_account') {
    pendingAccountEmail.value = email
    bindLoginEmail.value = email
    return
  }
  if (action === 'bind_login') {
    bindLoginEmail.value = email
    bindLoginPassword.value = ''
  }
}

function getRequestErrorMessage(error: unknown, fallback: string): string {
  const err = error as { message?: string; response?: { data?: { detail?: string; message?: string } } }
  return err.response?.data?.detail || err.response?.data?.message || err.message || fallback
}

function persistTokenContext(tokenData: PendingOAuthTokenPairResponse) {
  if (tokenData.refresh_token) {
    localStorage.setItem('refresh_token', tokenData.refresh_token)
  }
  if (tokenData.expires_in) {
    localStorage.setItem('token_expires_at', String(Date.now() + tokenData.expires_in * 1000))
  }
}

function isTokenPair(payload: PendingAccountResponse): payload is PendingOAuthTokenPairResponse {
  return typeof payload.access_token === 'string' && payload.access_token.trim().length > 0
}

async function finalizePendingAccountResponse(payload: PendingAccountResponse) {
  if (isTokenPair(payload)) {
    persistTokenContext(payload)
    await authStore.setToken(payload.access_token)
    appStore.showSuccess(t('auth.loginSuccess'))
    await router.replace(redirectTo.value)
    return
  }

  applyPendingAccountAction(payload)
}

async function handleSubmitInvitation() {
  invitationError.value = ''
  if (!invitationCode.value.trim()) return

  isSubmitting.value = true
  try {
    const tokenData = await completeLinuxDoOAuthRegistration(
      pendingOAuthToken.value,
      invitationCode.value.trim()
    )
    if (tokenData.refresh_token) {
      localStorage.setItem('refresh_token', tokenData.refresh_token)
    }
    if (tokenData.expires_in) {
      localStorage.setItem('token_expires_at', String(Date.now() + tokenData.expires_in * 1000))
    }
    await authStore.setToken(tokenData.access_token)
    appStore.showSuccess(t('auth.loginSuccess'))
    await router.replace(redirectTo.value)
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { message?: string } } }
    invitationError.value =
      err.response?.data?.message || err.message || t('auth.linuxdo.completeRegistrationFailed')
  } finally {
    isSubmitting.value = false
  }
}

async function handleCreateAccount(payload: PendingOAuthCreateAccountPayload) {
  accountActionError.value = ''
  isSubmitting.value = true
  try {
    const completion = await createPendingOAuthAccount({
      email: payload.email,
      password: payload.password,
      verify_code: payload.verifyCode || undefined,
      invitation_code: payload.invitationCode || undefined,
    })
    await finalizePendingAccountResponse(completion)
  } catch (e: unknown) {
    accountActionError.value = getRequestErrorMessage(e, t('auth.loginFailed'))
  } finally {
    isSubmitting.value = false
  }
}

async function handleBindLogin() {
  accountActionError.value = ''
  const email = bindLoginEmail.value.trim()
  const password = bindLoginPassword.value
  if (!email || !password) return

  isSubmitting.value = true
  try {
    const completion = await bindPendingOAuthLogin({ email, password })
    await finalizePendingAccountResponse(completion)
  } catch (e: unknown) {
    accountActionError.value = getRequestErrorMessage(e, t('auth.loginFailed'))
  } finally {
    isSubmitting.value = false
  }
}

function switchToBindLoginMode(nextEmail?: string) {
  pendingAccountAction.value = 'bind_login'
  bindLoginEmail.value = nextEmail?.trim() || pendingAccountEmail.value.trim()
  bindLoginPassword.value = ''
  accountActionError.value = ''
}

function switchToCreateAccountMode() {
  pendingAccountAction.value = 'create_account'
  pendingAccountEmail.value = pendingAccountEmail.value.trim() || bindLoginEmail.value.trim()
  accountActionError.value = ''
}

onMounted(async () => {
  const rawHash = typeof window !== 'undefined' ? window.location.hash : ''
  const params = parseFragmentParams(rawHash)
  const pendingPayload = await resolvePendingOAuthPayload(params)

  const token = pendingPayload.access_token || ''
  const refreshToken = pendingPayload.refresh_token || ''
  const expiresIn = pendingPayload.expires_in
  const redirect = sanitizeRedirectPath(
    pendingPayload.redirect || params.get('redirect') || (route.query.redirect as string | undefined) || '/dashboard'
  )
  const error = pendingPayload.error
  const errorDesc = params.get('error_description') || params.get('error_message') || ''

  if (error) {
    if (isInvitationRequired(pendingPayload)) {
      pendingOAuthToken.value = legacyPendingOAuthTokenFromFragment(params)
      redirectTo.value = redirect
      if (!pendingOAuthToken.value) {
        applyPendingAccountAction(pendingPayload)
        isProcessing.value = false
        return
      }
      needsInvitation.value = true
      isProcessing.value = false
      return
    }
    errorMessage.value = errorDesc || error
    appStore.showError(errorMessage.value)
    isProcessing.value = false
    return
  }

  if (!hasAuthTokenPayload(pendingPayload)) {
    errorMessage.value = t('auth.linuxdo.callbackMissingToken')
    appStore.showError(errorMessage.value)
    isProcessing.value = false
    return
  }

  try {
    // Store refresh token and expires_at (convert to timestamp) if provided
    persistTokenContext({
      access_token: token,
      refresh_token: refreshToken,
      expires_in: expiresIn || 0,
      token_type: pendingPayload.token_type || 'Bearer',
    })

    await authStore.setToken(token)
    appStore.showSuccess(t('auth.loginSuccess'))
    await router.replace(redirect)
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { detail?: string } } }
    errorMessage.value = err.response?.data?.detail || err.message || t('auth.loginFailed')
    appStore.showError(errorMessage.value)
    isProcessing.value = false
  }
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
