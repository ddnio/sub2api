<template>
  <footer class="border-t border-gray-200/70 px-6 py-5 dark:border-dark-700/50">
    <div class="mx-auto flex max-w-5xl flex-col items-center justify-between gap-3 text-center text-sm sm:flex-row sm:text-left">
      <p class="text-sm text-gray-500 dark:text-dark-400">
        &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
      </p>
      <div class="flex items-center gap-5 text-sm">
        <a
          :href="effectiveDocUrl"
          :target="hasExternalDocUrl ? '_blank' : '_self'"
          :rel="hasExternalDocUrl ? 'noopener noreferrer' : undefined"
          class="text-sm text-gray-500 transition-colors hover:text-gray-700 dark:text-dark-400 dark:hover:text-white"
        >
          {{ t('home.docs') }}
        </a>
        <a
          v-if="contactInfo"
          :href="contactInfoHref"
          class="text-sm text-gray-500 transition-colors hover:text-gray-700 dark:text-dark-400 dark:hover:text-white"
        >
          {{ t('home.nav.support') }}
        </a>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'NanaFox API')
const contactInfo = computed(() => appStore.contactInfo)
const contactInfoHref = computed(() => {
  const info = contactInfo.value
  if (!info) return '#'
  if (info.startsWith('http')) return info
  if (info.includes('@')) return `mailto:${info}`
  return info
})

const hasExternalDocUrl = computed(() => !!appStore.docUrl)
const effectiveDocUrl = computed(() => appStore.docUrl || '/docs')
const currentYear = computed(() => new Date().getFullYear())
</script>
