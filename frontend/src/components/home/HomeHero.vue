<template>
  <section class="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(320px,460px)] lg:gap-8">
    <!-- Left: Copy -->
    <div class="max-w-full sm:max-w-[500px]">
      <!-- Badge -->
      <div
        class="inline-flex items-center gap-2 rounded-full border border-emerald-500/[0.16] bg-emerald-50/[0.72] px-3 py-[7px] text-xs font-semibold text-emerald-700 dark:border-emerald-400/[0.18] dark:bg-dark-900/70 dark:text-emerald-300"
      >
        <span class="h-2.5 w-2.5 rounded-full bg-gradient-to-br from-emerald-500 to-emerald-400"></span>
        {{ t('home.badge') }}
      </div>

      <!-- Title block -->
      <div class="mt-4 flex flex-col gap-2.5">
        <p class="text-[13px] font-semibold tracking-[0.01em] text-teal-700 dark:text-teal-300">
          {{ siteSubtitle }}
        </p>
        <h1
          class="text-[clamp(32px,8vw,44px)] font-extrabold leading-[1.02] tracking-[-0.04em] text-dark-900 dark:text-dark-50 sm:text-[clamp(42px,5vw,64px)]"
        >
          {{ siteName }}
        </h1>
        <p class="max-w-[480px] text-sm leading-[1.9] text-gray-500 dark:text-gray-400 sm:text-[15px]">
          {{ t('home.heroDescription') }}
        </p>
      </div>

      <!-- CTA buttons -->
      <div class="mt-5 flex w-full flex-wrap items-center gap-3 sm:w-auto">
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="inline-flex flex-1 items-center justify-center gap-2 rounded-[14px] bg-gradient-to-br from-emerald-400 to-emerald-500 px-5 py-3 text-sm font-semibold text-white shadow-[0_10px_24px_rgba(16,185,129,0.16)] transition-transform hover:-translate-y-px sm:flex-none"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25" />
          </svg>
        </router-link>
        <a
          :href="effectiveDocUrl"
          :target="hasExternalDocUrl ? '_blank' : '_self'"
          :rel="hasExternalDocUrl ? 'noopener noreferrer' : undefined"
          class="inline-flex flex-1 items-center justify-center rounded-[14px] border border-gray-400/[0.22] bg-white/75 px-[18px] py-3 text-sm font-semibold text-dark-900 no-underline dark:border-gray-400/[0.16] dark:bg-dark-900/70 dark:text-gray-200 sm:flex-none"
        >
          {{ t('home.hero.viewDocs') }}
        </a>
      </div>

      <!-- Base URL -->
      <div class="mt-5 flex flex-col items-start gap-2.5">
        <p class="inline-flex items-center gap-2 text-[13px] font-semibold text-gray-500 dark:text-gray-400">
          <span
            class="h-2.5 w-2.5 rounded-full bg-gradient-to-br from-emerald-500 to-emerald-400 shadow-[0_0_0_3px_rgba(52,211,153,0.14)]"
          ></span>
          {{ t('home.hero.baseUrlHint') }}
        </p>

        <div
          class="inline-flex max-w-full flex-wrap items-center gap-2 rounded-2xl bg-dark-900 py-2 pl-3.5 pr-2.5 text-white shadow-[0_8px_20px_rgba(15,23,42,0.1)] sm:gap-3 sm:rounded-full sm:py-2.5 sm:pl-4 sm:pr-3"
        >
          <code class="min-w-0 break-all font-mono text-xs text-emerald-200 sm:text-[13px]">{{ apiBase }}</code>
          <button
            type="button"
            class="inline-flex shrink-0 items-center gap-1.5 rounded-full border-0 bg-white/[0.08] px-2.5 py-1.5 text-[11px] font-semibold text-gray-200 sm:px-3 sm:py-2 sm:text-xs"
            @click="copyBaseUrl"
          >
            <Icon :name="baseUrlCopied ? 'clipboard' : 'copy'" size="sm" />
            {{ baseUrlCopied ? t('home.hero.copiedBaseUrl') : t('home.hero.copyBaseUrl') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Right: Terminal -->
    <div class="flex justify-end">
      <div
        class="w-full max-w-[460px] overflow-hidden rounded-[20px] border border-gray-200/95 bg-white/[0.92] shadow-[0_18px_40px_rgba(15,23,42,0.06)] dark:border-dark-700/90 dark:bg-dark-900/90"
      >
        <!-- Terminal header -->
        <div
          class="flex items-center justify-between border-b border-gray-200/90 px-3.5 py-2.5 dark:border-dark-700/90 sm:px-[18px] sm:py-3.5"
        >
          <div class="flex gap-2 sm:gap-2.5">
            <span class="h-[11px] w-[11px] rounded-full bg-rose-400 sm:h-3.5 sm:w-3.5"></span>
            <span class="h-[11px] w-[11px] rounded-full bg-amber-400 sm:h-3.5 sm:w-3.5"></span>
            <span class="h-[11px] w-[11px] rounded-full bg-green-400 sm:h-3.5 sm:w-3.5"></span>
          </div>
          <span class="text-[10px] font-extrabold uppercase tracking-[0.18em] text-gray-400 sm:text-xs">
            Terminal
          </span>
        </div>

        <!-- Terminal body -->
        <div
          class="flex flex-col gap-2 px-3.5 py-4 font-mono text-[11.5px] leading-[1.75] sm:px-5 sm:pb-6 sm:pt-[22px] sm:text-[13px]"
        >
          <div class="text-gray-700 dark:text-gray-300">
            <span class="text-gray-400 dark:text-gray-500">// {{ t('home.hero.snippetTitle') }}</span>
          </div>
          <div class="text-gray-700 dark:text-gray-300">
            <span class="text-blue-700 dark:text-blue-400">export</span>
            <span> OPENAI_BASE_URL=</span>
            <span class="text-emerald-600 dark:text-emerald-400">"{{ apiBase }}"</span>
          </div>
          <div class="text-gray-700 dark:text-gray-300">
            <span class="text-blue-700 dark:text-blue-400">export</span>
            <span> OPENAI_API_KEY=</span>
            <span class="text-emerald-600 dark:text-emerald-400">"your-api-key-here"</span>
          </div>
          <div class="text-gray-700 dark:text-gray-300">&nbsp;</div>
          <div class="text-gray-700 dark:text-gray-300">
            <span class="text-gray-400 dark:text-gray-500"># {{ t('home.hero.runLabel') }}</span>
          </div>
          <div class="text-gray-700 dark:text-gray-300">$ python app.py</div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import { resolveSiteSubtitle } from '@/utils/siteBranding'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'NanaFox API')
const siteSubtitle = computed(() => resolveSiteSubtitle(appStore.cachedPublicSettings?.site_subtitle, t('home.heroSubtitle')))
// apiBase: root URL without /v1 — used for both the URL bar display and the terminal snippet
const apiBase = computed(() =>
  (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
)
const baseUrlCopied = ref(false)

const hasExternalDocUrl = computed(() => !!appStore.docUrl)
const effectiveDocUrl = computed(() => appStore.docUrl || '/docs')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))

async function copyBaseUrl() {
  if (!navigator.clipboard) {
    return
  }

  await navigator.clipboard.writeText(apiBase.value)
  baseUrlCopied.value = true
  window.setTimeout(() => {
    baseUrlCopied.value = false
  }, 1500)
}
</script>
