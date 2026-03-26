<!-- frontend/src/components/docs/DocsSectionFaq.vue -->
<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
      {{ t('docs.faq.title') }}
    </h1>

    <div class="space-y-4">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.faq.generalTitle') }}</h2>
      <div class="space-y-4">
        <article
          v-for="(item, idx) in faqItems"
          :key="idx"
          class="rounded-xl border border-gray-200/80 p-4 dark:border-dark-700"
        >
          <h3 class="font-semibold text-gray-900 dark:text-white">{{ item.q }}</h3>
          <p class="mt-2 text-sm text-gray-600 dark:text-dark-400">{{ item.a }}</p>
        </article>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const faqItems = computed(() => [
  { q: t('docs.faq.items.compatibilityQ'), a: t('docs.faq.items.compatibilityA') },
  { q: t('docs.faq.items.multiToolQ'), a: t('docs.faq.items.multiToolA') },
  { q: t('docs.faq.items.modelsQ'), a: t('docs.faq.items.modelsA') },
  { q: t('docs.api.faq.items.sdkQ'), a: t('docs.api.faq.items.sdkA') },
  { q: t('docs.api.faq.items.dataQ'), a: t('docs.api.faq.items.dataA') },
])
</script>
