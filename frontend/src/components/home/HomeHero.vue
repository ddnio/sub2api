<template>
  <section class="hero-shell">
    <div class="hero-copy">
      <div class="hero-tag">
        <span class="hero-tag-dot"></span>
        {{ t('home.badge') }}
      </div>

      <div class="hero-content">
        <p class="hero-overline">{{ siteSubtitle }}</p>
        <h1 class="hero-title">{{ siteName }}</h1>
        <p class="hero-desc">{{ t('home.heroDescription') }}</p>
      </div>

      <div class="hero-actions">
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="btn-cta">
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
          <svg class="btn-cta-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25" />
          </svg>
        </router-link>
        <a
          :href="effectiveDocUrl"
          :target="hasExternalDocUrl ? '_blank' : '_self'"
          :rel="hasExternalDocUrl ? 'noopener noreferrer' : undefined"
          class="btn-docs"
        >
          {{ t('home.hero.viewDocs') }}
        </a>
      </div>

      <div class="hero-base">
        <p class="hero-hint">
          <span class="hero-hint-dot"></span>
          {{ t('home.hero.baseUrlHint') }}
        </p>

        <div class="url-bar">
          <code class="url-code">{{ apiBase }}/v1</code>
          <button type="button" class="url-copy" @click="copyBaseUrl">
            <Icon :name="baseUrlCopied ? 'clipboard' : 'copy'" size="sm" />
            {{ baseUrlCopied ? t('home.hero.copiedBaseUrl') : t('home.hero.copyBaseUrl') }}
          </button>
        </div>
      </div>
    </div>

    <div class="hero-right">
      <div class="terminal-window">
        <div class="terminal-header">
          <div class="terminal-dots">
            <span class="dot dot-red"></span>
            <span class="dot dot-yellow"></span>
            <span class="dot dot-green"></span>
          </div>
          <span class="terminal-tab">Python</span>
        </div>

        <div class="terminal-body">
          <div class="code-line">
            <span class="t-comment"># {{ t('home.hero.snippetComment') }}</span>
          </div>
          <div class="code-line">
            <span class="t-keyword">from</span>
            <span class="t-plain"> openai </span>
            <span class="t-keyword">import</span>
            <span class="t-plain"> OpenAI</span>
          </div>
          <div class="code-line"><span class="t-plain">&nbsp;</span></div>
          <div class="code-line">
            <span class="t-plain">client = OpenAI(</span>
          </div>
          <div class="code-line">
            <span class="t-plain">&nbsp;&nbsp;base_url=</span>
            <span class="t-string">"{{ apiBase }}/v1"</span>
            <span class="t-plain">,</span>
          </div>
          <div class="code-line">
            <span class="t-plain">&nbsp;&nbsp;api_key=</span>
            <span class="t-string">"sk-..."</span>
            <span class="t-plain">,</span>
          </div>
          <div class="code-line"><span class="t-plain">)</span></div>
          <div class="code-line"><span class="t-plain">&nbsp;</span></div>
          <div class="code-line">
            <span class="t-plain">resp = client.chat.completions.create(</span>
          </div>
          <div class="code-line">
            <span class="t-plain">&nbsp;&nbsp;model=</span>
            <span class="t-string">"gpt-5.4"</span>
            <span class="t-plain">,</span>
          </div>
          <div class="code-line">
            <span class="t-plain">&nbsp;&nbsp;messages=[{</span>
            <span class="t-string">"role"</span>
            <span class="t-plain">: </span>
            <span class="t-string">"user"</span>
            <span class="t-plain">, </span>
            <span class="t-string">"content"</span>
            <span class="t-plain">: </span>
            <span class="t-string">"你好"</span>
            <span class="t-plain">}]</span>
          </div>
          <div class="code-line"><span class="t-plain">)</span></div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import { resolveSiteSubtitle } from '@/utils/siteBranding'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'NanaFox API')
const siteSubtitle = computed(() => resolveSiteSubtitle(appStore.cachedPublicSettings?.site_subtitle, t('home.heroSubtitle')))
// apiBase: root URL without /v1 — used for both the URL bar display and the terminal snippet
const apiBase = computed(() =>
  (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
)
const baseUrlCopied = ref(false)

const hasExternalDocUrl = computed(() => !!appStore.docUrl)
const effectiveDocUrl = computed(() => appStore.docUrl || '/docs')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))

async function copyBaseUrl() {
  if (!navigator.clipboard) {
    return
  }

  await navigator.clipboard.writeText(`${apiBase.value}/v1`)
  baseUrlCopied.value = true
  window.setTimeout(() => {
    baseUrlCopied.value = false
  }, 1500)
}
</script>

<style scoped>
.hero-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(320px, 460px);
  gap: 32px;
  align-items: start;
}

@media (max-width: 1024px) {
  .hero-shell {
    grid-template-columns: 1fr;
    gap: 24px;
  }
}

.hero-copy {
  max-width: 500px;
}

.hero-tag {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  border: 1px solid rgba(16, 185, 129, 0.16);
  border-radius: 999px;
  background: rgba(236, 253, 245, 0.72);
  font-size: 12px;
  font-weight: 600;
  color: #047857;
}

.dark .hero-tag {
  border-color: rgba(52, 211, 153, 0.18);
  background: rgba(15, 23, 42, 0.7);
  color: #6ee7b7;
}

.hero-tag-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: linear-gradient(135deg, #10b981, #34d399);
}

.hero-content {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.hero-overline {
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.01em;
  color: #0f766e;
}

.dark .hero-overline {
  color: #5eead4;
}

.hero-title {
  font-size: clamp(42px, 5vw, 64px);
  font-weight: 800;
  line-height: 1.02;
  letter-spacing: -0.04em;
  color: #0f172a;
}

.dark .hero-title {
  color: #f8fafc;
}

.hero-desc {
  max-width: 480px;
  font-size: 15px;
  line-height: 1.9;
  color: #64748b;
}

.dark .hero-desc {
  color: #94a3b8;
}

.hero-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 22px;
}

.btn-cta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, #34d399, #10b981);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  border-radius: 14px;
  text-decoration: none;
  box-shadow: 0 10px 24px rgba(16, 185, 129, 0.16);
}

.btn-cta:hover {
  transform: translateY(-1px);
}

.btn-cta-icon {
  width: 16px;
  height: 16px;
}

.btn-docs {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 12px 18px;
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(255, 255, 255, 0.75);
  color: #0f172a;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
}

.dark .btn-docs {
  background: rgba(15, 23, 42, 0.7);
  color: #e2e8f0;
  border-color: rgba(148, 163, 184, 0.16);
}

.hero-hint {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #64748b;
  font-size: 13px;
  font-weight: 600;
}

.dark .hero-hint {
  color: #94a3b8;
}

.hero-hint-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: linear-gradient(135deg, #10b981, #34d399);
  box-shadow: 0 0 0 3px rgba(52, 211, 153, 0.14);
}

.hero-base {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  margin-top: 22px;
}

.url-bar {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding: 10px 12px 10px 16px;
  border-radius: 999px;
  background: #0f172a;
  color: #fff;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.1);
  max-width: 100%;
}

.url-code {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  color: #a7f3d0;
  word-break: break-all;
}

.url-copy {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 0;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: #e2e8f0;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.hero-right {
  display: flex;
  justify-content: flex-end;
}

.terminal-window {
  width: min(100%, 460px);
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 20px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.dark .terminal-window {
  border-color: rgba(30, 41, 59, 0.9);
  background: rgba(15, 23, 42, 0.9);
}

.terminal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid rgba(226, 232, 240, 0.9);
}

.dark .terminal-header {
  border-color: rgba(51, 65, 85, 0.9);
}

.terminal-dots {
  display: flex;
  gap: 10px;
}

.dot {
  width: 14px;
  height: 14px;
  border-radius: 999px;
}

.dot-red { background: #fb7185; }
.dot-yellow { background: #fbbf24; }
.dot-green { background: #4ade80; }

.terminal-tab {
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #94a3b8;
}

.terminal-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 22px 20px 24px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  line-height: 1.75;
}

.code-line {
  color: #334155;
}

.dark .code-line {
  color: #cbd5e1;
}

.t-keyword { color: #1d4ed8; }
.t-plain { color: inherit; }
.t-string { color: #059669; }
.t-comment { color: #94a3b8; }

@media (max-width: 640px) {
  .hero-title {
    font-size: 44px;
  }

  .hero-desc {
    font-size: 15px;
    line-height: 1.8;
  }

  .hero-actions {
    width: 100%;
  }

  .btn-cta,
  .btn-docs {
    flex: 1 1 auto;
  }

  .url-bar {
    width: 100%;
    border-radius: 20px;
  }
}
</style>
