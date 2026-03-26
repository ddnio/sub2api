<template>
  <section class="text-center">
    <h2 class="mb-4 text-base font-bold text-gray-900 dark:text-white">
      {{ t('home.models.title') }}
    </h2>
    <div class="flex flex-wrap justify-center gap-2">
      <span
        v-for="model in models"
        :key="model.name"
        class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200/80 bg-white/80 px-3 py-1.5 text-xs font-medium text-gray-700 dark:border-dark-700/60 dark:bg-dark-800/60 dark:text-gray-300"
      >
        <span
          class="inline-block h-3 w-3 rounded"
          :style="{ background: model.color }"
        ></span>
        {{ model.name }}
      </span>
      <span
        v-for="soon in comingSoon"
        :key="soon"
        class="inline-flex items-center gap-1.5 rounded-lg border border-amber-200/60 bg-amber-50/80 px-3 py-1.5 text-xs font-medium text-amber-800 dark:border-amber-700/40 dark:bg-amber-900/20 dark:text-amber-400"
      >
        {{ soon }} · {{ t('home.models.comingSoon') }}
      </span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const models = [
  { name: 'GPT-5.4', color: '#10a37f' },
  { name: 'GPT-5.4 mini', color: '#10a37f' },
  { name: 'GPT-5.3', color: '#10a37f' },
  { name: 'Claude Opus 4.6', color: '#d97706' },
  { name: 'Claude Sonnet 4.6', color: '#d97706' },
  { name: 'Claude Haiku 4.5', color: '#d97706' },
]

const comingSoon = ['Gemini', 'DeepSeek']
</script>
