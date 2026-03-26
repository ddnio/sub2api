<template>
  <section class="grid gap-3 sm:grid-cols-2">
    <div
      v-for="feat in features"
      :key="feat.icon"
      class="flex items-center gap-3 rounded-2xl border border-gray-200/70 bg-white/75 px-4 py-3.5 shadow-sm dark:border-dark-800/70 dark:bg-dark-900/55"
    >
      <span class="text-base">{{ feat.icon }}</span>
      <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ feat.text }}
      </span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const features = computed(() => [
  { icon: '🔌', text: t('home.features.compat') },
  { icon: '⚡', text: t('home.features.failover') },
  { icon: '💰', text: t('home.features.billing') },
  { icon: '📊', text: t('home.features.usage') },
])
</script>
