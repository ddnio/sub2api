<!-- frontend/src/components/docs/DocsSectionFaq.vue -->
<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
      {{ t('docs.faq.title') }}
    </h1>

    <!-- NanaFox fork: 计费、报错排查、联系方式 -->
    <div class="space-y-4" data-docs-faq-billing>
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.faq.billingTitle') }}</h2>
      <ul class="space-y-2 text-sm text-gray-600 dark:text-dark-400">
        <li v-for="(line, idx) in billingLines" :key="idx" class="flex gap-2">
          <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-primary-500"></span>
          <span>{{ line }}</span>
        </li>
      </ul>
    </div>

    <div class="space-y-4" data-docs-faq-errors>
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.faq.errorsTitle') }}</h2>
      <div class="overflow-x-auto rounded-xl border border-gray-200/80 dark:border-dark-700">
        <table class="w-full text-left text-sm">
          <tbody class="divide-y divide-gray-200/80 dark:divide-dark-700">
            <tr v-for="row in errorRows" :key="row.code" class="text-gray-700 dark:text-dark-300">
              <td class="whitespace-nowrap px-4 py-3 font-mono font-semibold text-gray-900 dark:text-white">{{ row.code }}</td>
              <td class="px-4 py-3">{{ row.fix }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

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

    <div class="space-y-2" data-docs-faq-contact>
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.faq.contactTitle') }}</h2>
      <p class="text-sm text-gray-600 dark:text-dark-400">{{ t('docs.faq.contactDesc', { contact, email: 'dev@nanafox.com' }) }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()

const contact = computed(() => appStore.contactInfo?.trim() || t('docs.faq.contactFallback'))

const billingLines = computed(() => [
  t('docs.faq.billing.payg'),
  t('docs.faq.billing.topup'),
  t('docs.faq.billing.monthly')
])

const errorRows = computed(() => [
  { code: '404', fix: t('docs.faq.errors.notFound') },
  { code: '401', fix: t('docs.faq.errors.unauthorized') },
  { code: '403', fix: t('docs.faq.errors.forbidden') },
  { code: 'model not found', fix: t('docs.faq.errors.model') }
])

const faqItems = computed(() => [
  { q: t('docs.faq.items.compatibilityQ'), a: t('docs.faq.items.compatibilityA') },
  { q: t('docs.faq.items.multiToolQ'), a: t('docs.faq.items.multiToolA') },
  { q: t('docs.faq.items.modelsQ'), a: t('docs.faq.items.modelsA') },
  { q: t('docs.api.faq.items.sdkQ'), a: t('docs.api.faq.items.sdkA') },
  { q: t('docs.api.faq.items.dataQ'), a: t('docs.api.faq.items.dataA') },
])
</script>
