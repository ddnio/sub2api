<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('auth.oidc.callbackTitle', { providerName }) }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{
            isProcessing
              ? t('auth.oidc.callbackProcessing', { providerName })
              : t('auth.oidc.callbackHint')
          }}
        </p>
      </div>

      <transition name="fade">
        <div v-if="needsInvitation" class="space-y-4">
          <p class="text-sm text-gray-700 dark:text-gray-300">
            {{ t('auth.oidc.invitationRequired', { providerName }) }}
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
            {{
              isSubmitting
                ? t('auth.oidc.completing')
                : t('auth.oidc.completeRegistration')
            }}
          </button>
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
                {{ t('auth.oidc.backToLogin') }}
              </router-link>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore, useAppStore } from '@/stores'
import {
  completeOIDCOAuthRegistration,
  getPublicSettings
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

const needsInvitation = ref(false)
const pendingOAuthToken = ref('')
const invitationCode = ref('')
const isSubmitting = ref(false)
const invitationError = ref('')
const redirectTo = ref('/dashboard')
const providerName = ref('OIDC')

async function loadProviderName() {
  try {
    const settings = await getPublicSettings()
    const name = settings.oidc_oauth_provider_name?.trim()
    if (name) {
      providerName.value = name
    }
  } catch {
    // Ignore; fallback remains OIDC
  }
}

async function handleSubmitInvitation() {
  invitationError.value = ''
  if (!invitationCode.value.trim()) return

  isSubmitting.value = true
  try {
    const tokenData = await completeOIDCOAuthRegistration(
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
      err.response?.data?.message || err.message || t('auth.oidc.completeRegistrationFailed')
  } finally {
    isSubmitting.value = false
  }
}

onMounted(async () => {
  void loadProviderName()

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
        errorMessage.value = t('auth.oidc.invalidPendingToken')
        appStore.showError(errorMessage.value)
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
    errorMessage.value = t('auth.oidc.callbackMissingToken')
    appStore.showError(errorMessage.value)
    isProcessing.value = false
    return
  }

  try {
    if (refreshToken) {
      localStorage.setItem('refresh_token', refreshToken)
    }
    if (expiresIn && !isNaN(expiresIn)) {
      localStorage.setItem('token_expires_at', String(Date.now() + expiresIn * 1000))
    }

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
