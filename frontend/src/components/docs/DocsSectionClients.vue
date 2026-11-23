<!-- frontend/src/components/docs/DocsSectionClients.vue -->
<!-- NanaFox fork: Cursor / Cherry Studio / Cline 等图形客户端的填法 -->
<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
        {{ t('docs.clients.title') }}
      </h1>
      <p class="mt-3 text-sm leading-7 text-gray-600 dark:text-dark-400">
        {{ t('docs.clients.subtitle') }}
      </p>
    </div>

    <DocsGroupHint platform="openai" />
    <DocsGroupHint platform="anthropic" />

    <section v-for="client in clients" :key="client.id" class="space-y-3" :data-docs-client="client.id">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ client.title }}</h2>
      <p class="text-sm text-gray-600 dark:text-dark-400">{{ client.desc }}</p>
      <DocsCodeBlock :tabs="client.tabs" :default-tab="client.tabs[0].id" />
      <p v-if="client.note" class="text-xs text-gray-500 dark:text-dark-400">{{ client.note }}</p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import DocsCodeBlock from './DocsCodeBlock.vue'
import DocsGroupHint from './DocsGroupHint.vue'
import type { CodeTab } from './DocsCodeBlock.vue'
import { DOCS_MODELS } from '@/utils/docsSnippets'

const { t } = useI18n()
const appStore = useAppStore()

const apiBase = computed(() =>
  (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
)

// 等宽字体里中文占两格，按显示宽度对齐
const displayWidth = (s: string) => [...s].reduce((n, ch) => n + (/[\u3000-\u9fff\uff00-\uffef]/.test(ch) ? 2 : 1), 0)

function fields(rows: [string, string][]): string {
  const width = Math.max(...rows.map(([k]) => displayWidth(k)))
  return rows.map(([k, v]) => `${k}${' '.repeat(width - displayWidth(k))}  ${v}`).join('\n')
}

const gptTab = (id: string, label: string, protocolLabel: string, protocol: string): CodeTab => ({
  id,
  label,
  content: fields([
    [protocolLabel, protocol],
    ['Base URL', `${apiBase.value}/v1`],
    ['API Key', `your-api-key-here  # ${t('docs.clients.openaiKey')}`],
    ['Model', DOCS_MODELS.openai]
  ])
})

const claudeTab = (id: string, label: string, protocolLabel: string, protocol: string): CodeTab => ({
  id,
  label,
  content: fields([
    [protocolLabel, protocol],
    ['Base URL', apiBase.value],
    ['API Key', `your-api-key-here  # ${t('docs.clients.anthropicKey')}`],
    ['Model', DOCS_MODELS.anthropic]
  ])
})

const clients = computed(() => [
  {
    id: 'cursor',
    title: 'Cursor',
    desc: t('docs.clients.cursorDesc'),
    note: t('docs.clients.cursorNote'),
    tabs: [
      {
        id: 'gpt',
        label: 'GPT',
        content: fields([
          ['Settings', 'Cursor Settings → Models'],
          ['OpenAI API Key', `your-api-key-here  # ${t('docs.clients.openaiKey')}`],
          ['Override OpenAI Base URL', `${apiBase.value}/v1`],
          ['Add model', DOCS_MODELS.openai]
        ])
      }
    ]
  },
  {
    id: 'cherry-studio',
    title: 'Cherry Studio / Chatbox / LobeChat',
    desc: t('docs.clients.cherryDesc'),
    note: t('docs.clients.anthropicNote'),
    tabs: [
      gptTab('gpt', 'GPT', t('docs.clients.providerType'), 'OpenAI'),
      claudeTab('claude', 'Claude', t('docs.clients.providerType'), 'Anthropic')
    ]
  },
  {
    id: 'cline',
    title: 'Cline / Roo Code',
    desc: t('docs.clients.clineDesc'),
    note: t('docs.clients.anthropicNote'),
    tabs: [
      gptTab('gpt', 'GPT', 'API Provider', 'OpenAI Compatible'),
      claudeTab('claude', 'Claude', 'API Provider', `Anthropic  # ${t('docs.clients.customBaseUrl')}`)
    ]
  }
])
</script>
