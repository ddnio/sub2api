<!-- frontend/src/views/DocsView.vue -->
<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <HomeHeader :is-dark="isDark" @toggle-theme="toggleTheme" />

    <main class="px-4 py-6 sm:px-6 sm:py-8">
      <div class="mx-auto flex max-w-5xl gap-8">
        <!-- Left nav (desktop: sidebar, mobile: dropdown rendered inside content) -->
        <DocsNav
          :active="activeSection"
          @navigate="navigateTo"
        />

        <!-- Right content -->
        <div class="min-w-0 flex-1">
          <KeepAlive>
            <component
              :is="currentSectionComponent"
              :key="activeSection"
              @navigate="navigateTo"
            />
          </KeepAlive>
        </div>
      </div>
    </main>

    <HomeFooter />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAuthStore, useAppStore } from '@/stores'
import { useDocsSectionRoute, type DocsSection } from '@/composables/useDocsSectionRoute'
import HomeFooter from '@/components/home/HomeFooter.vue'
import HomeHeader from '@/components/home/HomeHeader.vue'
import DocsNav from '@/components/docs/DocsNav.vue'
import DocsSectionQuickStart from '@/components/docs/DocsSectionQuickStart.vue'
import DocsSectionClaudeCode from '@/components/docs/DocsSectionClaudeCode.vue'
import DocsSectionCodexCli from '@/components/docs/DocsSectionCodexCli.vue'
import DocsSectionOpenCode from '@/components/docs/DocsSectionOpenCode.vue'
import DocsSectionApiUsage from '@/components/docs/DocsSectionApiUsage.vue'
import DocsSectionFaq from '@/components/docs/DocsSectionFaq.vue'

const authStore = useAuthStore()
const appStore = useAppStore()
const { activeSection, navigateTo } = useDocsSectionRoute()

const isDark = ref(document.documentElement.classList.contains('dark'))

const sectionComponents: Record<DocsSection, any> = {
  'quick-start': DocsSectionQuickStart,
  'claude-code': DocsSectionClaudeCode,
  'codex-cli': DocsSectionCodexCli,
  'opencode': DocsSectionOpenCode,
  'api-usage': DocsSectionApiUsage,
  'faq': DocsSectionFaq,
}

const currentSectionComponent = computed(() => sectionComponents[activeSection.value])

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  const prefersDark =
    typeof window.matchMedia === 'function'
      ? window.matchMedia('(prefers-color-scheme: dark)').matches
      : false
  if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
