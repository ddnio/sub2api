<!-- frontend/src/components/docs/DocsSectionClaudeCode.vue -->
<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
        {{ t('docs.claudeCode.title') }}
      </h1>
      <p class="mt-3 text-sm leading-7 text-gray-600 dark:text-dark-400">
        {{ t('docs.claudeCode.subtitle') }}
      </p>
    </div>

    <!-- Step 1: Environment Variables -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">1</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.claudeCode.envTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.claudeCode.envDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock
        :tabs="envTabs"
        sync-group="shell"
        default-tab="unix"
      />
    </section>

    <!-- Step 2: VSCode settings.json -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">2</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.claudeCode.settingsTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.claudeCode.settingsDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock
        :tabs="settingsTabs"
        sync-group="shell"
        default-tab="unix"
      />
    </section>

    <!-- Step 3: Verify -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">3</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.claudeCode.verifyTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.claudeCode.verifyDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock :tabs="[{ id: 'verify', label: 'Terminal', path: 'Terminal', content: 'claude' }]" />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import DocsCodeBlock from './DocsCodeBlock.vue'
import type { CodeTab } from './DocsCodeBlock.vue'
import {
  generateClaudeCodeEnvSnippet,
  generateClaudeCodeSettingsSnippet,
  type ShellType
} from '@/utils/docsSnippets'

const { t } = useI18n()
const appStore = useAppStore()

const apiBase = computed(() =>
  (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
)
const apiKey = 'your-api-key-here'

const shells: { id: ShellType; label: string }[] = [
  { id: 'unix', label: 'macOS / Linux' },
  { id: 'cmd', label: 'Windows CMD' },
  { id: 'powershell', label: 'PowerShell' }
]

const envTabs = computed<CodeTab[]>(() =>
  shells.map(s => {
    const snippet = generateClaudeCodeEnvSnippet(apiBase.value, apiKey, s.id)
    return { id: s.id, label: s.label, path: snippet.path, content: snippet.content }
  })
)

const settingsTabs = computed<CodeTab[]>(() =>
  shells.map(s => {
    const snippet = generateClaudeCodeSettingsSnippet(apiBase.value, apiKey, s.id)
    return { id: s.id, label: s.label, path: snippet.path, content: snippet.content, hint: snippet.hint }
  })
)
</script>
