<!-- frontend/src/components/docs/DocsSectionApiUsage.vue -->
<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
        {{ t('docs.api.title') }}
      </h1>
      <p class="mt-3 text-sm leading-7 text-gray-600 dark:text-dark-400">
        {{ t('docs.api.subtitle') }}
      </p>
    </div>

    <!-- Base URL -->
    <div class="rounded-xl border border-gray-200/80 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900">
      <p class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
        {{ t('docs.api.baseUrlLabel') }}
      </p>
      <code class="mt-2 block text-sm font-mono text-primary-600 dark:text-primary-400">{{ apiBase }}/v1</code>
    </div>

    <!-- SDK Installation -->
    <section class="space-y-3">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.api.sdkTitle') }}</h2>
      <p class="text-sm text-gray-600 dark:text-dark-400">{{ t('docs.api.sdkDescription') }}</p>
      <DocsCodeBlock
        :tabs="sdkInstallTabs"
        sync-group="lang"
        default-tab="python"
      />
    </section>

    <!-- Chat Completions example -->
    <section class="space-y-3">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.api.examplesTitle') }}</h2>
      <DocsCodeBlock
        :tabs="exampleTabs"
        sync-group="lang"
        default-tab="python"
      />
    </section>

    <!-- Streaming example -->
    <section class="space-y-3">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.api.streamingTitle') }}</h2>
      <p class="text-sm text-gray-600 dark:text-dark-400">{{ t('docs.api.streamingDescription') }}</p>
      <DocsCodeBlock
        :tabs="streamingTabs"
        sync-group="lang"
        default-tab="python"
      />
    </section>

    <!-- Auth -->
    <section class="space-y-3">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.api.authTitle') }}</h2>
      <p class="text-sm text-gray-600 dark:text-dark-400">{{ t('docs.api.authDescription') }}</p>
    </section>

    <!-- Endpoints -->
    <section class="space-y-3">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.api.endpointsTitle') }}</h2>
      <ul class="space-y-2 text-sm text-gray-600 dark:text-dark-400">
        <li class="flex gap-2">
          <code class="shrink-0 text-primary-600 dark:text-primary-400">/v1/chat/completions</code>
          <span>— {{ t('docs.api.endpoints.chat') }}</span>
        </li>
        <li class="flex gap-2">
          <code class="shrink-0 text-primary-600 dark:text-primary-400">/v1/models</code>
          <span>— {{ t('docs.api.endpoints.models') }}</span>
        </li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import DocsCodeBlock from './DocsCodeBlock.vue'
import type { CodeTab } from './DocsCodeBlock.vue'
import { generateApiExample, generateApiStreamingExample, type ApiLanguage } from '@/utils/docsSnippets'

const { t } = useI18n()
const appStore = useAppStore()

const apiBase = computed(() =>
  (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
)

const languages: { id: ApiLanguage; label: string }[] = [
  { id: 'python', label: 'Python' },
  { id: 'curl', label: 'cURL' },
  { id: 'nodejs', label: 'Node.js' }
]

const sdkInstallTabs = computed<CodeTab[]>(() => [
  { id: 'python', label: 'Python', path: 'pip', content: 'pip install openai' },
  { id: 'curl', label: 'cURL', path: 'cURL', content: '# cURL is pre-installed on most systems\ncurl --version' },
  { id: 'nodejs', label: 'Node.js', path: 'npm', content: 'npm install openai' }
])

const exampleTabs = computed<CodeTab[]>(() =>
  languages.map(lang => {
    const snippet = generateApiExample(apiBase.value, 'your-api-key-here', lang.id)
    return { id: lang.id, label: lang.label, path: snippet.path, content: snippet.content }
  })
)

const streamingTabs = computed<CodeTab[]>(() =>
  languages.map(lang => {
    const snippet = generateApiStreamingExample(apiBase.value, 'your-api-key-here', lang.id)
    return { id: lang.id, label: lang.label, path: snippet.path, content: snippet.content }
  })
)
</script>
