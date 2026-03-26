<!-- frontend/src/components/docs/DocsCodeBlock.vue -->
<template>
  <div class="relative">
    <!-- Hint -->
    <p v-if="currentTab?.hint" class="text-xs text-amber-600 dark:text-amber-400 mb-1.5 flex items-center gap-1">
      <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
      </svg>
      {{ currentTab.hint }}
    </p>

    <div class="overflow-hidden rounded-xl bg-gray-900 dark:bg-dark-900">
      <!-- Header with tabs + copy -->
      <div class="flex items-center justify-between border-b border-gray-700/60 dark:border-dark-700 px-4 py-2">
        <div class="flex items-center gap-1">
          <!-- Tab buttons (only if multiple tabs) -->
          <template v-if="tabs.length > 1">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              type="button"
              class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
              :class="activeTab === tab.id
                ? 'bg-gray-700 text-white'
                : 'text-gray-400 hover:text-gray-200'"
              @click="activeTab = tab.id"
            >
              {{ tab.label }}
            </button>
          </template>
          <!-- Single tab: show path as label -->
          <span v-else class="text-xs text-gray-400 font-mono">{{ currentTab?.path }}</span>
        </div>

        <button
          type="button"
          class="flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-lg transition-colors"
          :class="copied
            ? 'bg-green-500/20 text-green-400'
            : 'bg-gray-700 hover:bg-gray-600 text-gray-300 hover:text-white'"
          @click="copyCode"
        >
          <svg v-if="copied" class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
          <svg v-else class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184" />
          </svg>
          {{ copied ? t('docs.shared.copied') : t('docs.shared.copy') }}
        </button>
      </div>

      <!-- File path (when multi-tab, show path below tabs) -->
      <div v-if="tabs.length > 1 && currentTab?.path" class="px-4 pt-2">
        <span class="text-xs text-gray-500 font-mono">{{ currentTab.path }}</span>
      </div>

      <!-- Code content -->
      <pre class="p-4 text-sm font-mono text-gray-100 overflow-x-auto leading-relaxed"><code v-text="currentTab?.content ?? ''"></code></pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSyncedTabState } from '@/composables/useSyncedTabState'

export interface CodeTab {
  id: string
  label: string
  path?: string
  content: string
  hint?: string
}

const props = withDefaults(defineProps<{
  tabs: CodeTab[]
  syncGroup?: string
  defaultTab?: string
  scope?: string
}>(), {
  scope: 'docs'
})

const { t } = useI18n()
const copied = ref(false)

// Use synced state if syncGroup is provided, otherwise local state
const tabIds = computed(() => props.tabs.map(tab => tab.id))
const defaultTabId = computed(() => props.defaultTab || props.tabs[0]?.id || '')

const { activeTab } = props.syncGroup
  ? useSyncedTabState({
      group: props.syncGroup,
      scope: props.scope,
      availableTabs: tabIds.value,
      defaultTab: defaultTabId.value
    })
  : { activeTab: ref(defaultTabId.value) }

const currentTab = computed(() =>
  props.tabs.find(tab => tab.id === activeTab.value) || props.tabs[0]
)

async function copyCode() {
  const content = currentTab.value?.content
  if (!content) return

  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(content)
    } else {
      const textarea = document.createElement('textarea')
      textarea.value = content
      textarea.style.cssText = 'position:fixed;left:-9999px;top:-9999px'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // Silent fail
  }
}
</script>
