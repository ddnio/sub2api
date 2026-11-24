<template>
  <div
    v-if="visible"
    data-testid="docs-signup-cta"
    class="mb-6 flex flex-col gap-3 rounded-xl border border-primary-200 bg-primary-50 px-4 py-3 dark:border-primary-800/50 dark:bg-primary-900/20 sm:flex-row sm:items-center sm:justify-between"
  >
    <div>
      <p class="font-medium text-gray-900 dark:text-white">
        {{ bonusText ? t('docs.signupCta.titleWithBonus', { amount: bonusText }) : t('docs.signupCta.title') }}
      </p>
      <p class="mt-1 text-sm text-gray-600 dark:text-dark-300">{{ t('docs.signupCta.desc') }}</p>
    </div>
    <router-link to="/register" class="btn btn-primary flex-shrink-0 text-center">
      {{ bonusText ? t('docs.signupCta.buttonWithBonus', { amount: bonusText }) : t('docs.signupCta.button') }}
    </router-link>
  </div>
</template>

<script setup lang="ts">
// NanaFox fork: 文档页顶部给未登录访客一个显眼的注册入口；金额取后台的新用户默认余额，没配时不写金额。
// 只在拿到公共设置且确认开放注册时显示：预渲染的静态页里没有它，已登录用户也不会先闪一下。
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const visible = computed(
  () => !authStore.isAuthenticated && appStore.cachedPublicSettings?.registration_enabled === true
)
const bonusText = computed(() => {
  const amount = Number(Number(appStore.cachedPublicSettings?.default_balance || 0).toFixed(2))
  return amount > 0 ? `$${amount}` : ''
})
</script>
