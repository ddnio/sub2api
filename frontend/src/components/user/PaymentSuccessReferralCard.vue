<!-- NanaFox fork: 付费成功页一次性可关闭的“邀请同事一起用”提示卡 -->
<template>
  <div
    v-if="visible"
    data-testid="payment-success-referral-card"
    class="rounded-xl border border-primary-200/80 bg-primary-50/70 p-4 text-left dark:border-primary-800/60 dark:bg-primary-900/20"
  >
    <div class="flex items-start justify-between gap-3">
      <div class="flex items-start gap-2.5">
        <Icon name="gift" size="md" class="mt-0.5 flex-shrink-0 text-primary-500" />
        <div>
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('firstPayNudges.paymentSuccessInviteTitle') }}
          </p>
          <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-dark-300">
            {{ t('firstPayNudges.paymentSuccessInviteDesc') }}
          </p>
        </div>
      </div>
      <button
        type="button"
        data-testid="payment-referral-dismiss"
        class="shrink-0 text-xs text-gray-400 transition-colors hover:text-gray-600 dark:text-dark-400 dark:hover:text-gray-200"
        @click="dismiss"
      >
        {{ t('common.close') }}
      </button>
    </div>
    <div class="mt-3 flex flex-wrap items-center gap-2">
      <button
        type="button"
        data-testid="payment-referral-copy-share"
        class="btn btn-primary btn-sm"
        :disabled="copying"
        @click="copyShareText"
      >
        {{ t('firstPayNudges.copyShareText') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { getActivePinia } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import { userAPI } from '@/api/user'
import { useClipboard } from '@/composables/useClipboard'
import Icon from '@/components/icons/Icon.vue'
import type { UserAffiliateDetail } from '@/types'
import {
  buildReferralShareText,
  dismissPaymentSuccessReferral,
  isPaymentSuccessReferralDismissed
} from '@/utils/referralShare'

const i18n = useI18n()
const { t } = i18n
const hasPinia = Boolean(getActivePinia())
const appStore = hasPinia ? useAppStore() : null
const authStore = hasPinia ? useAuthStore() : null
const clipboard = hasPinia ? useClipboard() : null

function readStorageUserId(): number | null {
  try {
    if (typeof window === 'undefined' || !window.localStorage.getItem('auth_token')) {
      return null
    }
    const raw = window.localStorage.getItem('auth_user')
    if (!raw) return null
    const parsed = JSON.parse(raw) as { id?: unknown } | null
    return typeof parsed?.id === 'number' && parsed.id > 0 ? parsed.id : null
  } catch {
    return null
  }
}

function resolveAuthenticatedUserId(): number | null {
  // 优先读 localStorage 中的最新登录态（可感知同浏览器其他标签页切换账号或登出）
  const storageUserId = readStorageUserId()
  if (storageUserId) return storageUserId

  try {
    const hasStorageToken = typeof window !== 'undefined' && Boolean(window.localStorage.getItem('auth_token'))
    if (hasStorageToken) return null
  } catch {
    // ignore storage errors
  }

  // 测试环境或纯内存登录态：仅在 authStore 明确标记已登录且有有效 user.id 时放行
  if (authStore?.isAuthenticated && typeof authStore.user?.id === 'number' && authStore.user.id > 0) {
    return authStore.user.id
  }
  return null
}

const currentUserId = ref<number | null>(resolveAuthenticatedUserId())
const dismissed = ref(isPaymentSuccessReferralDismissed(currentUserId.value))
const detail = ref<UserAffiliateDetail | null>(null)
const copying = ref(false)

const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || 'zh')
  }
  return 'zh'
})

const rate = computed(() => Number(detail.value?.effective_rebate_rate_percent) || 0)
const shareText = computed(() =>
  detail.value?.aff_code
    ? buildReferralShareText(window.location.origin, detail.value.aff_code, localeCode.value)
    : ''
)
const visible = computed(() => !dismissed.value && rate.value > 0 && Boolean(shareText.value))

async function copyShareText(): Promise<void> {
  if (!clipboard || copying.value) return
  const activeUserId = resolveAuthenticatedUserId()
  if (!activeUserId) {
    detail.value = null
    return
  }
  currentUserId.value = activeUserId
  copying.value = true
  try {
    // 复制前按当前登录用户重新获取邀请码，避免多标签页切换账号后复制到旧用户的邀请码
    const freshDetail = await userAPI.getAffiliateDetail({ skipAuthRedirect: true })
    if (unmounted) return
    detail.value = freshDetail
    const freshRate = Number(freshDetail?.effective_rebate_rate_percent) || 0
    const freshText =
      freshRate > 0 && freshDetail?.aff_code
        ? buildReferralShareText(window.location.origin, freshDetail.aff_code, localeCode.value)
        : ''
    if (!freshText) return
    await clipboard.copyToClipboard(freshText, t('firstPayNudges.shareTextCopied'))
  } catch {
    // 请求失败（含 401）时静默隐藏卡片，不影响支付成功页
    if (!unmounted) detail.value = null
  } finally {
    if (!unmounted) copying.value = false
  }
}

function dismiss(): void {
  const activeUserId = resolveAuthenticatedUserId() ?? currentUserId.value
  dismissed.value = true
  dismissPaymentSuccessReferral(activeUserId)
}

let unmounted = false
onBeforeUnmount(() => {
  unmounted = true
})

onMounted(async () => {
  const activeUserId = resolveAuthenticatedUserId()
  currentUserId.value = activeUserId
  if (!activeUserId || isPaymentSuccessReferralDismissed(activeUserId) || !appStore) {
    dismissed.value = Boolean(activeUserId && isPaymentSuccessReferralDismissed(activeUserId))
    return
  }
  try {
    const settings = await appStore.fetchPublicSettings().catch(() => null)
    if (unmounted || !settings?.affiliate_enabled) return
    const result = await userAPI.getAffiliateDetail({ skipAuthRedirect: true })
    if (!unmounted) detail.value = result
  } catch {
    // 拿不到返佣信息（含 401）时静默不展示，不影响支付成功页
    if (!unmounted) detail.value = null
  }
})
</script>
