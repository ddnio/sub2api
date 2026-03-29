<template>
  <header class="px-6 py-4 backdrop-blur-sm dark:bg-dark-950/80">
    <nav class="mx-auto flex max-w-5xl items-center justify-between gap-6">
      <router-link to="/home" class="flex items-center gap-2.5">
        <div class="h-9 w-9 overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
        </div>
        <span class="text-base font-semibold text-gray-900 dark:text-white">{{ siteName }}</span>
      </router-link>

      <div class="flex items-center gap-2">
        <a
          :href="effectiveDocUrl"
          :target="hasExternalDocUrl ? '_blank' : '_self'"
          :rel="hasExternalDocUrl ? 'noopener noreferrer' : undefined"
          class="hidden text-sm text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white md:block mr-1"
        >
          {{ t('home.nav.docs') }}
        </a>
        <LocaleSwitcher />
        <button
          @click="$emit('toggle-theme')"
          class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
        >
          <Icon v-if="isDark" name="sun" size="md" />
          <Icon v-else name="moon" size="md" />
        </button>
        <router-link
          v-if="isAuthenticated"
          :to="dashboardPath"
          class="inline-flex items-center gap-1.5 rounded-full bg-gray-900 py-1 pl-1 pr-2.5 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
        >
          <span
            class="flex h-5 w-5 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[10px] font-semibold text-white"
          >
            {{ userInitial }}
          </span>
          <span>{{ t('home.dashboard') }}</span>
          <svg class="h-3 w-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25" />
          </svg>
        </router-link>
        <router-link
          v-else
          to="/login"
          class="inline-flex items-center rounded-full bg-gray-900 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
        >
          {{ t('home.login') }}
        </router-link>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  isDark: boolean
}>()

defineEmits<{
  'toggle-theme': []
}>()

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'NanaFox API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const hasExternalDocUrl = computed(() => !!appStore.docUrl)
const effectiveDocUrl = computed(() => appStore.docUrl || '/docs')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})
</script>
