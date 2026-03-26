<!-- frontend/src/components/docs/DocsNav.vue -->
<template>
  <!-- Desktop sidebar -->
  <nav class="hidden lg:block w-56 shrink-0">
    <div class="sticky top-24 space-y-1">
      <button
        v-for="item in navItems"
        :key="item.section"
        type="button"
        class="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition-colors"
        :class="active === item.section
          ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300'
          : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white'"
        @click="$emit('navigate', item.section)"
      >
        {{ item.label }}
      </button>
    </div>
  </nav>

  <!-- Mobile dropdown -->
  <div class="lg:hidden mb-4">
    <select
      :value="active"
      class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm font-medium text-gray-900 dark:border-dark-700 dark:bg-dark-900 dark:text-white"
      @change="$emit('navigate', ($event.target as HTMLSelectElement).value as DocsSection)"
    >
      <option v-for="item in navItems" :key="item.section" :value="item.section">
        {{ item.label }}
      </option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { DocsSection } from '@/composables/useDocsSectionRoute'

defineProps<{
  active: DocsSection
}>()

defineEmits<{
  (e: 'navigate', section: DocsSection): void
}>()

const { t } = useI18n()

interface NavItem {
  section: DocsSection
  label: string
}

const navItems = computed<NavItem[]>(() => [
  { section: 'quick-start', label: t('docs.nav.quickStart') },
  { section: 'claude-code', label: t('docs.nav.claudeCode') },
  { section: 'codex-cli', label: t('docs.nav.codexCli') },
  { section: 'opencode', label: t('docs.nav.opencode') },
  { section: 'api-usage', label: t('docs.nav.apiUsage') },
  { section: 'faq', label: t('docs.nav.faq') },
])
</script>
