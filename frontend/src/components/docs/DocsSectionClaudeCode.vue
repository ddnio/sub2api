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

    <!-- Step 1: Install -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">1</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.claudeCode.installTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.claudeCode.installDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock
        :tabs="installTabs"
        sync-group="shell"
        default-tab="unix"
      />
      <p class="text-sm text-gray-600 dark:text-dark-400">
        {{ t('docs.claudeCode.vscodeExtNote') }}
      </p>
    </section>

    <!-- Step 2: Environment Variables -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">2</span>
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
      <div class="rounded-lg border border-amber-200/80 bg-amber-50/50 p-3 dark:border-amber-900/30 dark:bg-amber-900/10">
        <p class="text-xs text-amber-800 dark:text-amber-300">
          {{ t('docs.claudeCode.envNote') }}
        </p>
      </div>
    </section>

    <!-- Step 3: VSCode settings.json -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">3</span>
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

    <!-- Step 4: Verify -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">4</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.claudeCode.verifyTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.claudeCode.verifyDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock :tabs="[{ id: 'verify', label: 'Terminal', path: 'Terminal', content: verifyCommands }]" />
    </section>

    <!-- Troubleshooting -->
    <section class="space-y-4">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.claudeCode.troubleshootTitle') }}</h2>
      <div class="space-y-3">
        <article v-for="item in troubleshootItems" :key="item.q" class="rounded-xl border border-gray-200/80 p-4 dark:border-dark-700">
          <h3 class="font-semibold text-gray-900 dark:text-white">{{ item.q }}</h3>
          <p class="mt-2 text-sm text-gray-600 dark:text-dark-400">{{ item.a }}</p>
        </article>
      </div>
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
  generateClaudeCodeInstallSnippet,
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

const installTabs = computed<CodeTab[]>(() =>
  shells.map(s => {
    const snippet = generateClaudeCodeInstallSnippet(s.id)
    return { id: s.id, label: s.label, path: snippet.path, content: snippet.content }
  })
)

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

const verifyCommands = `# Check connection status
claude /status

# Or start a conversation to test
claude "hello"`

const troubleshootItems = computed(() => [
  { q: t('docs.claudeCode.troubleshoot.baseUrlQ'), a: t('docs.claudeCode.troubleshoot.baseUrlA') },
  { q: t('docs.claudeCode.troubleshoot.restartQ'), a: t('docs.claudeCode.troubleshoot.restartA') },
  { q: t('docs.claudeCode.troubleshoot.priorityQ'), a: t('docs.claudeCode.troubleshoot.priorityA') },
])
</script>
