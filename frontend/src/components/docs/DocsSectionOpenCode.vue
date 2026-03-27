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

    <!-- Step 1: Install -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">1</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.opencode.installTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.opencode.installDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock
        :tabs="installTabs"
        sync-group="shell"
        default-tab="unix"
      />
    </section>

    <!-- Step 2: Config file -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">2</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.opencode.configTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.opencode.configDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock :tabs="[configTab]" />
      <div class="rounded-lg border border-amber-200/80 bg-amber-50/50 p-3 dark:border-amber-900/30 dark:bg-amber-900/10">
        <p class="text-xs text-amber-800 dark:text-amber-300">
          {{ t('docs.opencode.configNote') }}
        </p>
      </div>
    </section>

    <!-- Step 3: Verify & Start -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">3</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.opencode.verifyTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.opencode.verifyDescription') }}</p>
        </div>
      </div>
      <DocsCodeBlock :tabs="[{ id: 'verify', label: 'Terminal', path: 'Terminal', content: verifyCommands }]" />
    </section>

    <!-- Config tips -->
    <section class="space-y-3">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.opencode.tipsTitle') }}</h2>
      <ul class="space-y-2 text-sm text-gray-600 dark:text-dark-400">
        <li class="flex gap-2">
          <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500"></span>
          <span>{{ t('docs.opencode.tips.envVar') }}</span>
        </li>
        <li class="flex gap-2">
          <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500"></span>
          <span>{{ t('docs.opencode.tips.hierarchy') }}</span>
        </li>
        <li class="flex gap-2">
          <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500"></span>
          <span>{{ t('docs.opencode.tips.models') }}</span>
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
import {
  generateOpenCodeInstallSnippet,
  generateOpenCodeEnhancedSnippet,
  type ShellType
} from '@/utils/docsSnippets'

const { t } = useI18n()
const appStore = useAppStore()

const apiBase = computed(() => {
  const root = (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
  return root.endsWith('/v1') ? root : `${root}/v1`
})

const shells: { id: ShellType; label: string }[] = [
  { id: 'unix', label: 'macOS / Linux' },
  { id: 'cmd', label: 'Windows CMD' },
  { id: 'powershell', label: 'PowerShell' }
]

const installTabs = computed<CodeTab[]>(() =>
  shells.map(s => {
    const snippet = generateOpenCodeInstallSnippet(s.id)
    return { id: s.id, label: s.label, path: snippet.path, content: snippet.content }
  })
)

const configTab = computed<CodeTab>(() => {
  const snippet = generateOpenCodeEnhancedSnippet(apiBase.value, 'your-api-key-here')
  return { id: 'opencode', label: 'opencode.json', path: snippet.path, content: snippet.content }
})

const verifyCommands = `# Check version
opencode --version

# Start OpenCode
opencode`
</script>
