<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else
    class="relative flex min-h-screen flex-col overflow-hidden bg-[linear-gradient(180deg,_#f8fafc_0%,_#ffffff_45%,_#f8fafc_100%)] dark:bg-dark-950"
  >
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute -right-40 top-0 h-72 w-72 rounded-full bg-primary-400/8 blur-3xl"></div>
      <div class="absolute -bottom-40 -left-20 h-72 w-72 rounded-full bg-primary-500/8 blur-3xl"></div>
      <div class="absolute inset-0 bg-[linear-gradient(rgba(20,184,166,0.02)_1px,transparent_1px),linear-gradient(90deg,rgba(20,184,166,0.02)_1px,transparent_1px)] bg-[size:64px_64px]"></div>
    </div>

    <HomeHeader :is-dark="isDark" @toggle-theme="toggleTheme" />

    <main class="relative flex-1 px-4 pb-12 pt-6 sm:px-6 sm:pt-8 lg:pb-16 lg:pt-10">
      <div class="mx-auto flex w-full max-w-5xl flex-col gap-6 lg:gap-8">
        <section class="rounded-3xl border border-gray-200/60 bg-white/72 p-6 shadow-sm backdrop-blur-sm dark:border-dark-800/60 dark:bg-dark-900/50 lg:p-7">
          <HomeHero />
        </section>

        <section class="grid gap-3 sm:grid-cols-3">
          <div
            v-for="item in metrics"
            :key="item.title"
            class="rounded-2xl border border-gray-200/70 bg-white/75 px-5 py-4 shadow-sm dark:border-dark-800/70 dark:bg-dark-900/55"
          >
            <p class="text-sm font-medium text-gray-500 dark:text-dark-400">
              {{ item.title }}
            </p>
            <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">
              {{ item.value }}
            </p>
          </div>
        </section>
      </div>
    </main>

    <HomeFooter />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import HomeFooter from '@/components/home/HomeFooter.vue'
import HomeHeader from '@/components/home/HomeHeader.vue'
import HomeHero from '@/components/home/HomeHero.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const metrics = computed(() => [
  {
    title: t('home.metrics.compatibilityTitle'),
    value: t('home.metrics.compatibilityValue')
  },
  {
    title: t('home.metrics.routingTitle'),
    value: t('home.metrics.routingValue')
  },
  {
    title: t('home.metrics.billingTitle'),
    value: t('home.metrics.billingValue')
  }
])

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  const prefersDark =
    typeof window.matchMedia === 'function'
      ? window.matchMedia('(prefers-color-scheme: dark)').matches
      : false
  if (
    savedTheme === 'dark' ||
    (!savedTheme && prefersDark)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
