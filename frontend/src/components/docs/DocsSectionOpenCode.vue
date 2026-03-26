<!-- frontend/src/components/docs/DocsSectionOpenCode.vue -->
<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
        {{ t('docs.opencode.title') }}
      </h1>
      <p class="mt-3 text-sm leading-7 text-gray-600 dark:text-dark-400">
        {{ t('docs.opencode.subtitle') }}
      </p>
    </div>

    <!-- Step 1: Config file -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">1</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.opencode.configTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.opencode.configDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock :tabs="[configTab]" />
    </section>

    <!-- Step 2: Start -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">2</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.opencode.startTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.opencode.startDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock :tabs="[{ id: 'start', label: 'Terminal', path: 'Terminal', content: 'opencode' }]" />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import DocsCodeBlock from './DocsCodeBlock.vue'
import type { CodeTab } from './DocsCodeBlock.vue'
import { generateOpenCodeSnippet } from '@/utils/docsSnippets'

const { t } = useI18n()
const appStore = useAppStore()

const apiBase = computed(() => {
  const root = (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
  return root.endsWith('/v1') ? root : `${root}/v1`
})

const configTab = computed<CodeTab>(() => {
  const snippet = generateOpenCodeSnippet('openai', apiBase.value, 'your-api-key-here')
  return { id: 'opencode', label: 'opencode.json', path: snippet.path, content: snippet.content }
})
</script>
