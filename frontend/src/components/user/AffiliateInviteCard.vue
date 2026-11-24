<template>
  <div
    v-if="rate > 0 && inviteLink"
    data-testid="affiliate-invite-card"
    class="card flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between"
  >
    <div class="flex items-start gap-3">
      <Icon name="gift" size="md" class="mt-0.5 flex-shrink-0 text-primary-500" />
      <div>
        <p class="font-medium text-gray-900 dark:text-white">{{ t('firstPayNudges.inviteTitle', { rate: rateText }) }}</p>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('firstPayNudges.inviteDesc', { rate: rateText }) }}</p>
      </div>
    </div>
    <div class="flex flex-shrink-0 flex-wrap gap-2">
      <button
        type="button"
        data-testid="affiliate-copy-share-text"
        class="btn btn-primary btn-sm"
        @click="copyShareText"
      >
        {{ t('firstPayNudges.copyShareText') }}
      </button>
      <button
        type="button"
        data-testid="affiliate-copy-link"
        class="btn btn-secondary btn-sm"
        @click="copyLink"
      >
        {{ t('firstPayNudges.copyLink') }}
      </button>
      <router-link to="/affiliate" class="btn btn-secondary btn-sm">{{ t('firstPayNudges.details') }}</router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
// NanaFox fork: 在控制台和充值页露出邀请返佣入口；比例取接口里实际生效的返佣比例，不写死。
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { userAPI } from '@/api/user'
import { useClipboard } from '@/composables/useClipboard'
import Icon from '@/components/icons/Icon.vue'
import type { UserAffiliateDetail } from '@/types'
import { buildCodexReferralLink, buildReferralShareText } from '@/utils/referralShare'

const i18n = useI18n()
const { t } = i18n
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

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
const rateText = computed(() => String(Math.round(rate.value * 100) / 100))
const inviteLink = computed(() =>
  detail.value?.aff_code ? buildCodexReferralLink(window.location.origin, detail.value.aff_code) : ''
)
const shareText = computed(() =>
  detail.value?.aff_code
    ? buildReferralShareText(window.location.origin, detail.value.aff_code, localeCode.value)
    : ''
)

async function copyShareText(): Promise<void> {
  if (shareText.value) await copyToClipboard(shareText.value, t('firstPayNudges.shareTextCopied'))
}

async function copyLink(): Promise<void> {
  if (inviteLink.value) await copyToClipboard(inviteLink.value, t('firstPayNudges.linkCopied'))
}

let unmounted = false
onBeforeUnmount(() => {
  unmounted = true
})

onMounted(async () => {
  const settings = await appStore.fetchPublicSettings().catch(() => null)
  if (unmounted || !settings?.affiliate_enabled) return
  try {
    const result = await userAPI.getAffiliateDetail()
    if (!unmounted) detail.value = result
  } catch {
    // 拿不到就不显示，不打扰用户
  }
})
</script>
