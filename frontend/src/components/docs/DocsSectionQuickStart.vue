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

    <!-- NanaFox fork: 工具 → 分组 → 地址 对照表 -->
    <div class="overflow-x-auto rounded-xl border border-gray-200/80 dark:border-dark-700" data-docs-group-table>
      <table class="w-full text-left text-sm">
        <thead class="bg-gray-50 text-xs uppercase tracking-wider text-gray-500 dark:bg-dark-900 dark:text-dark-400">
          <tr>
            <th class="px-4 py-2 font-semibold">{{ t('docs.groups.colTool') }}</th>
            <th class="px-4 py-2 font-semibold">{{ t('docs.groups.colGroup') }}</th>
            <th class="px-4 py-2 font-semibold">{{ t('docs.groups.colBase') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200/80 dark:divide-dark-700">
          <tr v-for="row in groupRows" :key="row.id" class="text-gray-700 dark:text-dark-300">
            <td class="px-4 py-3">{{ row.tools }}</td>
            <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{{ row.group }}</td>
            <td class="px-4 py-3"><code class="font-mono text-primary-600 dark:text-primary-400">{{ row.base }}</code></td>
          </tr>
        </tbody>
      </table>
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
          <h3 class="font-semibold text-gray-900 dark:text-white">{{ t('docs.quickStart.useKeyStep') }}</h3>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.quickStart.useKeyDesc') }}</p>
        </div>
      </div>

      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">3</span>
        <div>
          <h3 class="font-semibold text-gray-900 dark:text-white">{{ t('docs.quickStart.testKeyStep') }}</h3>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ t('docs.quickStart.testKeyDesc') }}</p>
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

const groupRows = computed(() => [
  { id: 'anthropic', tools: t('docs.groups.anthropicTools'), group: t('docs.groups.anthropicName'), base: apiBase.value },
  { id: 'openai', tools: t('docs.groups.openaiTools'), group: t('docs.groups.openaiName'), base: `${apiBase.value}/v1` }
])

const tools = computed(() => [
  { section: 'claude-code' as DocsSection, title: 'Claude Code', desc: t('docs.quickStart.toolCards.claudeCode') },
  { section: 'codex-cli' as DocsSection, title: 'Codex CLI', desc: t('docs.quickStart.toolCards.codexCli') },
  { section: 'clients' as DocsSection, title: t('docs.nav.clients'), desc: t('docs.quickStart.toolCards.clients') },
  { section: 'opencode' as DocsSection, title: 'OpenCode', desc: t('docs.quickStart.toolCards.opencode') },
  { section: 'api-usage' as DocsSection, title: t('docs.nav.apiUsage'), desc: t('docs.quickStart.toolCards.apiUsage') },
])
</script>
