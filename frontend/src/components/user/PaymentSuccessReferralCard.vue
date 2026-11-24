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
import { useAppStore } from '@/stores'
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
const clipboard = hasPinia ? useClipboard() : null

const dismissed = ref(isPaymentSuccessReferralDismissed())
const detail = ref<UserAffiliateDetail | null>(null)

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
  if (!shareText.value || !clipboard) return
  await clipboard.copyToClipboard(shareText.value, t('firstPayNudges.shareTextCopied'))
}

function dismiss(): void {
  dismissed.value = true
  dismissPaymentSuccessReferral()
}

let unmounted = false
onBeforeUnmount(() => {
  unmounted = true
})

onMounted(async () => {
  if (dismissed.value || !appStore) return
  try {
    const settings = await appStore.fetchPublicSettings().catch(() => null)
    if (unmounted || !settings?.affiliate_enabled) return
    const result = await userAPI.getAffiliateDetail()
    if (!unmounted) detail.value = result
  } catch {
    // 拿不到返佣信息时静默不展示，不影响支付成功页
  }
})
</script>
