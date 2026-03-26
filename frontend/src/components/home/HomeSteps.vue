<template>
  <section>
    <h2 class="mb-4 text-center text-base font-bold text-gray-900 dark:text-white">
      {{ t('home.steps.title') }}
    </h2>
    <div class="grid gap-3 sm:grid-cols-3">
      <div
        v-for="(step, i) in steps"
        :key="i"
        class="rounded-2xl border border-gray-200/70 bg-white/75 px-4 py-5 text-center shadow-sm dark:border-dark-800/70 dark:bg-dark-900/55"
      >
        <div
          class="mx-auto flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-sm font-bold text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400"
        >
          {{ i + 1 }}
        </div>
        <p class="mt-3 text-sm font-semibold text-gray-900 dark:text-white">
          {{ step.title }}
        </p>
        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
          {{ step.desc }}
        </p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const steps = computed(() => [
  { title: t('home.steps.step1Title'), desc: t('home.steps.step1Desc') },
  { title: t('home.steps.step2Title'), desc: t('home.steps.step2Desc') },
  { title: t('home.steps.step3Title'), desc: t('home.steps.step3Desc') },
])
</script>
