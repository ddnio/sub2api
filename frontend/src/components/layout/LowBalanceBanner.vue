<template>
  <div
    v-if="visible"
    data-testid="low-balance-banner"
    class="mb-4 flex flex-col gap-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-800/50 dark:bg-amber-900/20 dark:text-amber-300 sm:flex-row sm:items-center sm:justify-between"
  >
    <div class="flex items-center gap-2">
      <Icon name="exclamationCircle" size="md" class="flex-shrink-0 text-amber-500" />
      <span>
        {{ balance <= 0 ? t('firstPayNudges.balanceEmpty') : t('firstPayNudges.balanceLow', { amount: balanceText }) }}
      </span>
    </div>
    <router-link v-if="internalPath" :to="internalPath" class="btn btn-primary btn-sm flex-shrink-0 text-center">
      {{ t('firstPayNudges.recharge') }}
    </router-link>
    <a v-else :href="rechargeURL" class="btn btn-primary btn-sm flex-shrink-0 text-center">
      {{ t('firstPayNudges.recharge') }}
    </a>
  </div>
</template>

<script setup lang="ts">
// NanaFox fork: 余额用完或低于后台「低余额提醒」默认阈值时，在控制台顶部提示去充值。
// 阈值和充值地址跟提醒邮件用同一组设置；管理员和有生效中订阅的用户不提示。
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useSubscriptionStore } from '@/stores/subscriptions'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const subscriptionStore = useSubscriptionStore()

const balance = computed(() => Number(authStore.user?.balance ?? 0))
const balanceText = computed(() => `$${balance.value.toFixed(2)}`)
const threshold = computed(() => Number(appStore.cachedPublicSettings?.balance_low_notify_threshold) || 0)

const rechargeURL = computed(() => {
  const configured = appStore.cachedPublicSettings?.balance_low_notify_recharge_url?.trim() || ''
  return /^https?:\/\//i.test(configured) ? configured : '/purchase'
})

// 充值页在本站时走前端路由，不整页刷新
const internalPath = computed(() => {
  if (rechargeURL.value.startsWith('/')) return rechargeURL.value
  try {
    const url = new URL(rechargeURL.value)
    return url.origin === window.location.origin ? `${url.pathname}${url.search}${url.hash}` : ''
  } catch {
    return ''
  }
})

const visible = computed(() => {
  if (!authStore.user || authStore.user.role === 'admin') return false
  if (subscriptionStore.hasActiveSubscriptions) return false
  if (internalPath.value && route.path === internalPath.value.split(/[?#]/)[0]) return false
  return balance.value <= 0 || (threshold.value > 0 && balance.value < threshold.value)
})
</script>
