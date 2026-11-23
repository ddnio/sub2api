<!-- frontend/src/components/docs/DocsGroupHint.vue -->
<!-- NanaFox fork: 每节开头提示建 key 选哪个分组、地址要不要加 /v1 -->
<template>
  <div
    class="rounded-lg border border-primary-200/80 bg-primary-50/60 p-3 text-sm text-primary-800 dark:border-primary-900/40 dark:bg-primary-900/10 dark:text-primary-300"
    :data-docs-group-hint="platform"
  >
    {{ text }}
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'

const props = defineProps<{ platform: 'anthropic' | 'openai'; keyOnly?: boolean }>()

const { t } = useI18n()
const appStore = useAppStore()

const apiBase = computed(() =>
  (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
)

const text = computed(() => {
  if (props.platform === 'anthropic') return t('docs.groups.anthropicHint', { base: apiBase.value })
  if (props.keyOnly) return t('docs.groups.openaiKeyOnlyHint')
  return t('docs.groups.openaiHint', { base: `${apiBase.value}/v1` })
})
</script>
