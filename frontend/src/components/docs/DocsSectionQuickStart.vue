<!-- frontend/src/components/docs/DocsSectionQuickStart.vue -->
<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
        {{ t('docs.quickStart.title') }}
      </h1>
      <p class="mt-3 text-sm leading-7 text-gray-600 dark:text-dark-400">
        {{ t('docs.quickStart.subtitle') }}
      </p>
    </div>

    <!-- Base URL -->
    <div class="rounded-xl border border-gray-200/80 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900">
      <p class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
        {{ t('docs.quickStart.baseUrlLabel') }}
      </p>
      <code class="mt-2 block text-sm font-mono text-primary-600 dark:text-primary-400">{{ apiBase }}/v1</code>
    </div>

    <!-- Steps -->
    <div class="space-y-4">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">1</span>
        <div>
          <h3 class="font-semibold text-gray-900 dark:text-white">{{ t('docs.quickStart.getKeyStep') }}</h3>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.quickStart.getKeyDesc') }}</p>
        </div>
      </div>

      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">2</span>
        <div>
          <h3 class="font-semibold text-gray-900 dark:text-white">{{ t('docs.quickStart.chooseToolStep') }}</h3>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.quickStart.chooseToolDesc') }}</p>
        </div>
      </div>
    </div>

    <!-- Tool cards -->
    <div class="grid gap-3 sm:grid-cols-2">
      <button
        v-for="tool in tools"
        :key="tool.section"
        type="button"
        class="rounded-xl border border-gray-200/80 bg-white p-4 text-left shadow-sm transition-colors hover:border-primary-300 hover:bg-primary-50/50 dark:border-dark-700 dark:bg-dark-900 dark:hover:border-primary-700 dark:hover:bg-primary-900/10"
        @click="$emit('navigate', tool.section)"
      >
        <h4 class="font-semibold text-gray-900 dark:text-white">{{ tool.title }}</h4>
        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ tool.desc }}</p>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import type { DocsSection } from '@/composables/useDocsSectionRoute'

defineEmits<{ (e: 'navigate', section: DocsSection): void }>()

const { t } = useI18n()
const appStore = useAppStore()

const apiBase = computed(() =>
  (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
)

const tools = computed(() => [
  { section: 'claude-code' as DocsSection, title: 'Claude Code', desc: t('docs.quickStart.toolCards.claudeCode') },
  { section: 'codex-cli' as DocsSection, title: 'Codex CLI', desc: t('docs.quickStart.toolCards.codexCli') },
  { section: 'opencode' as DocsSection, title: 'OpenCode', desc: t('docs.quickStart.toolCards.opencode') },
  { section: 'api-usage' as DocsSection, title: t('docs.nav.apiUsage'), desc: t('docs.quickStart.toolCards.apiUsage') },
])
</script>
