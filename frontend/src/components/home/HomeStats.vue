<template>
  <section class="grid gap-3 sm:grid-cols-3">
    <div
      v-for="item in stats"
      :key="item.title"
      class="rounded-2xl border border-gray-200/70 bg-white/75 px-5 py-4 text-center shadow-sm dark:border-dark-800/70 dark:bg-dark-900/55"
    >
      <p class="text-lg font-extrabold text-gray-900 dark:text-white">
        {{ item.title }}
      </p>
      <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
        {{ item.desc }}
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const stats = computed(() => [
  { title: t('home.stats.gptTitle'), desc: t('home.stats.gptDesc') },
  { title: t('home.stats.claudeTitle'), desc: t('home.stats.claudeDesc') },
  { title: t('home.stats.billingTitle'), desc: t('home.stats.billingDesc') },
])
</script>
