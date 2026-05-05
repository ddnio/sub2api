<template>
  <form class="space-y-3" @submit.prevent="handleSubmit">
    <input
      v-model="email"
      :data-testid="`${testIdPrefix}-create-account-email`"
      type="email"
      class="input w-full"
<<<<<<< HEAD
      placeholder="you@example.com"
=======
      :placeholder="t('auth.emailPlaceholder')"
>>>>>>> v0.1.116
      :disabled="isSubmitting || isSendingCode"
    />
    <input
      v-model="password"
      :data-testid="`${testIdPrefix}-create-account-password`"
      type="password"
      class="input w-full"
<<<<<<< HEAD
      placeholder="Password"
      :disabled="isSubmitting"
    />
    <div class="flex gap-3">
      <input
        v-model="verifyCode"
        :data-testid="`${testIdPrefix}-create-account-verify-code`"
        type="text"
        inputmode="numeric"
        maxlength="6"
        class="input min-w-0 flex-1"
        placeholder="123456"
        :disabled="isSubmitting"
      />
=======
      :placeholder="t('auth.passwordPlaceholder')"
      :disabled="isSubmitting"
    />
    <div v-if="turnstileEnabled && turnstileSiteKey" class="space-y-2">
      <TurnstileWidget
        ref="turnstileRef"
        :site-key="turnstileSiteKey"
        @verify="onTurnstileVerify"
        @expire="onTurnstileExpire"
        @error="onTurnstileError"
      />
    </div>
    <div class="flex gap-3">
    <input
      v-model="verifyCode"
      :data-testid="`${testIdPrefix}-create-account-verify-code`"
      type="text"
        inputmode="numeric"
      maxlength="6"
      class="input min-w-0 flex-1"
      placeholder="123456"
      :disabled="isSubmitting"
    />
>>>>>>> v0.1.116
      <button
        :data-testid="`${testIdPrefix}-create-account-send-code`"
        type="button"
        class="btn btn-secondary shrink-0"
<<<<<<< HEAD
        :disabled="isSubmitting || isSendingCode || countdown > 0 || !email.trim()"
=======
        :disabled="isSubmitting || isSendingCode || countdown > 0 || !email.trim() || (turnstileEnabled && !turnstileToken)"
>>>>>>> v0.1.116
        @click="handleSendCode"
      >
        {{
          isSendingCode
            ? t('auth.sendingCode')
            : countdown > 0
              ? t('auth.resendCountdown', { countdown })
              : t('auth.sendCode')
        }}
      </button>
    </div>
<<<<<<< HEAD
    <input
      v-if="showInvitationCode"
=======
    <p v-if="sendCodeSuccess" class="text-sm text-green-600 dark:text-green-400">
      {{ t('auth.codeSentSuccess') }}
    </p>
    <p v-else class="text-xs text-gray-500 dark:text-dark-400">
      {{ t('auth.verificationCodeHint') }}
    </p>
    <input
      v-if="invitationCodeEnabled"
>>>>>>> v0.1.116
      v-model="invitationCode"
      :data-testid="`${testIdPrefix}-create-account-invitation-code`"
      type="text"
      class="input w-full"
      :placeholder="t('auth.invitationCodePlaceholder')"
      :disabled="isSubmitting"
    />
<<<<<<< HEAD
    <p v-if="sendCodeSuccess" class="text-sm text-green-600 dark:text-green-400">
      {{ t('auth.codeSentSuccess') }}
    </p>
    <p v-else class="text-xs text-gray-500 dark:text-dark-400">
      {{ t('auth.verificationCodeHint') }}
    </p>
=======
>>>>>>> v0.1.116
    <button
      :data-testid="`${testIdPrefix}-create-account-submit`"
      type="button"
      class="btn btn-primary w-full"
<<<<<<< HEAD
      :disabled="isSubmitting || !email.trim() || password.length < 6"
=======
      :disabled="isSubmitting || !email.trim() || password.length < 6 || (invitationCodeEnabled && !invitationCode.trim())"
>>>>>>> v0.1.116
      @click="handleSubmit"
    >
      {{ isSubmitting ? t('common.processing') : t('auth.createAccount') }}
    </button>
    <button
      type="button"
      class="btn btn-secondary w-full"
      :disabled="isSubmitting"
      @click="emitSwitchToBind"
    >
      {{ t('auth.alreadyHaveAccount') }}
    </button>
<<<<<<< HEAD
    <transition name="fade">
      <p v-if="sendCodeError" class="text-sm text-red-600 dark:text-red-400">
        {{ sendCodeError }}
      </p>
    </transition>
    <transition name="fade">
      <p v-if="errorMessage" class="text-sm text-red-600 dark:text-red-400">
        {{ errorMessage }}
      </p>
    </transition>
=======
>>>>>>> v0.1.116
  </form>
</template>

<script setup lang="ts">
<<<<<<< HEAD
import { onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { sendVerifyCode } from '@/api/auth'

export interface PendingOAuthCreateAccountPayload {
  email: string
  password: string
  verifyCode: string
  invitationCode: string
}

const props = withDefaults(
  defineProps<{
    initialEmail: string
    testIdPrefix: string
    isSubmitting: boolean
    errorMessage?: string
    showInvitationCode?: boolean
  }>(),
  {
    errorMessage: '',
    showInvitationCode: false,
  }
)
=======
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { getPublicSettings, sendPendingOAuthVerifyCode } from '@/api/auth'
import { useAppStore } from '@/stores'

export type PendingOAuthCreateAccountPayload = {
  email: string
  password: string
  verifyCode: string
  invitationCode?: string
}

const props = defineProps<{
  initialEmail: string
  testIdPrefix: string
  isSubmitting: boolean
  errorMessage?: string
}>()
>>>>>>> v0.1.116

const emit = defineEmits<{
  submit: [payload: PendingOAuthCreateAccountPayload]
  switchToBind: [email: string]
}>()

const { t } = useI18n()
<<<<<<< HEAD
=======
const appStore = useAppStore()
>>>>>>> v0.1.116

const email = ref('')
const password = ref('')
const verifyCode = ref('')
const invitationCode = ref('')
const isSendingCode = ref(false)
const sendCodeError = ref('')
const sendCodeSuccess = ref(false)
const countdown = ref(0)
<<<<<<< HEAD
=======
const invitationCodeEnabled = ref(false)
const turnstileEnabled = ref(false)
const turnstileSiteKey = ref('')
const turnstileToken = ref('')
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
>>>>>>> v0.1.116

let countdownTimer: ReturnType<typeof setInterval> | null = null

watch(
  () => props.initialEmail,
  value => {
    email.value = value || ''
  },
  { immediate: true }
)

<<<<<<< HEAD
=======
watch(sendCodeError, value => {
  if (value) {
    appStore.showError(value)
  }
})

watch(
  () => props.errorMessage,
  value => {
    if (value) {
      appStore.showError(value)
    }
  }
)

>>>>>>> v0.1.116
function clearCountdown() {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

function startCountdown(seconds: number) {
  clearCountdown()
  countdown.value = Math.max(0, seconds)

  if (countdown.value <= 0) {
    return
  }

  countdownTimer = setInterval(() => {
    if (countdown.value <= 1) {
      countdown.value = 0
      clearCountdown()
      return
    }
<<<<<<< HEAD
=======

>>>>>>> v0.1.116
    countdown.value -= 1
  }, 1000)
}

function getRequestErrorMessage(error: unknown, fallback: string): string {
  const err = error as { message?: string; response?: { data?: { detail?: string; message?: string } } }
  return err.response?.data?.detail || err.response?.data?.message || err.message || fallback
}

<<<<<<< HEAD
=======
function resetTurnstile() {
  turnstileToken.value = ''
  turnstileRef.value?.reset()
}

function onTurnstileVerify(token: string) {
  turnstileToken.value = token
  sendCodeError.value = ''
}

function onTurnstileExpire() {
  turnstileToken.value = ''
  sendCodeError.value = t('auth.turnstileExpired')
}

function onTurnstileError() {
  turnstileToken.value = ''
  sendCodeError.value = t('auth.turnstileFailed')
}

>>>>>>> v0.1.116
async function handleSendCode() {
  const trimmedEmail = email.value.trim()
  if (!trimmedEmail) {
    return
  }

<<<<<<< HEAD
=======
  if (turnstileEnabled.value && !turnstileToken.value) {
    sendCodeError.value = t('auth.completeVerification')
    return
  }

>>>>>>> v0.1.116
  isSendingCode.value = true
  sendCodeError.value = ''
  sendCodeSuccess.value = false

  try {
<<<<<<< HEAD
    const response = await sendVerifyCode({ email: trimmedEmail })
    sendCodeSuccess.value = true
    startCountdown(response.countdown)
=======
    const response = await sendPendingOAuthVerifyCode({
      email: trimmedEmail,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
    })
    sendCodeSuccess.value = true
    startCountdown(response.countdown)
    if (turnstileEnabled.value) {
      resetTurnstile()
    }
>>>>>>> v0.1.116
  } catch (error: unknown) {
    sendCodeError.value = getRequestErrorMessage(error, t('auth.sendCodeFailed'))
  } finally {
    isSendingCode.value = false
  }
}

function handleSubmit() {
  const trimmedEmail = email.value.trim()
  if (!trimmedEmail || password.value.length < 6) {
    return
  }

  emit('submit', {
    email: trimmedEmail,
    password: password.value,
    verifyCode: verifyCode.value.trim(),
<<<<<<< HEAD
    invitationCode: invitationCode.value.trim(),
=======
    invitationCode: invitationCode.value.trim() || undefined
>>>>>>> v0.1.116
  })
}

function emitSwitchToBind() {
  emit('switchToBind', email.value.trim())
}

<<<<<<< HEAD
=======
onMounted(async () => {
  try {
    const settings = await getPublicSettings()
    invitationCodeEnabled.value = settings.invitation_code_enabled === true
    turnstileEnabled.value = settings.turnstile_enabled === true
    turnstileSiteKey.value = settings.turnstile_site_key || ''
  } catch {
    invitationCodeEnabled.value = false
    turnstileEnabled.value = false
    turnstileSiteKey.value = ''
  }
})

>>>>>>> v0.1.116
onUnmounted(() => {
  clearCountdown()
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
