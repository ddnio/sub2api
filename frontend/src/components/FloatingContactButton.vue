<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useI18n } from 'vue-i18n'
import type { ContactChannel } from '@/types'

const FIRST_HINT_SEEN_KEY = 'contact_first_hint_seen'

// 排除路径前缀（admin / 安装向导 / 登录注册）— 避免遮挡管理界面与表单流
const EXCLUDED_PATH_PREFIXES = ['/admin', '/setup', '/login', '/register']

const route = useRoute()
const appStore = useAppStore()
const { t } = useI18n()

const isOpen = ref(false)
const activeType = ref<string>('')
const showFirstHint = ref(false)
let firstHintTimerShow: ReturnType<typeof setTimeout> | null = null
let firstHintTimerHide: ReturnType<typeof setTimeout> | null = null

const channels = computed<ContactChannel[]>(() => {
  const list = appStore.cachedPublicSettings?.contact_channels ?? []
  return [...list]
    .filter((c) => c && c.enabled && c.qr_image)
    .sort((a, b) => a.priority - b.priority)
})

const isExcludedRoute = computed(() => {
  const p = route.path || '/'
  return EXCLUDED_PATH_PREFIXES.some((prefix) => p === prefix || p.startsWith(prefix + '/'))
})

const shouldRender = computed(() => {
  return !isExcludedRoute.value && channels.value.length > 0
})

const activeChannel = computed<ContactChannel | undefined>(() => {
  if (channels.value.length === 0) return undefined
  return channels.value.find((c) => c.type === activeType.value) ?? channels.value[0]
})

watch(
  channels,
  (list) => {
    if (list.length === 0) {
      activeType.value = ''
      return
    }
    if (!list.some((c) => c.type === activeType.value)) {
      activeType.value = list[0].type
    }
  },
  { immediate: true }
)

function readFirstHintSeen(): boolean {
  try {
    return localStorage.getItem(FIRST_HINT_SEEN_KEY) === '1'
  } catch {
    return true
  }
}

function markFirstHintSeen() {
  try {
    localStorage.setItem(FIRST_HINT_SEEN_KEY, '1')
  } catch {
    // 忽略 storage 不可用
  }
}

function dismissFirstHint() {
  showFirstHint.value = false
  markFirstHintSeen()
  if (firstHintTimerHide) {
    clearTimeout(firstHintTimerHide)
    firstHintTimerHide = null
  }
}

async function open() {
  // R2: 打开弹窗时刷一次最新公开配置（store 内部不强制时走缓存，开销可控）
  try {
    await appStore.fetchPublicSettings(false)
  } catch {
    // 网络错误时仍使用现有缓存渲染
  }
  dismissFirstHint()
  isOpen.value = true
}

function close() {
  isOpen.value = false
}

function selectTab(type: string) {
  activeType.value = type
}

function tabLabel(c: ContactChannel): string {
  if (c.label) return c.label
  return t(`contact.channelTypes.${c.type}`)
}

onMounted(() => {
  if (readFirstHintSeen()) return
  firstHintTimerShow = setTimeout(() => {
    if (!shouldRender.value || isOpen.value) return
    showFirstHint.value = true
    firstHintTimerHide = setTimeout(() => {
      showFirstHint.value = false
      markFirstHintSeen()
    }, 7000)
  }, 3000)
})

onBeforeUnmount(() => {
  if (firstHintTimerShow) clearTimeout(firstHintTimerShow)
  if (firstHintTimerHide) clearTimeout(firstHintTimerHide)
})
</script>

<template>
  <div v-if="shouldRender">
    <!-- 首次访问的轻气泡提示（仅展示一次，localStorage 记忆） -->
    <transition name="fc-fade">
      <div
        v-if="showFirstHint && !isOpen"
        class="floating-contact-hint"
        @click="open"
      >
        <span class="floating-contact-hint-text">{{ t('contact.firstHint') }}</span>
        <button
          type="button"
          class="floating-contact-hint-close"
          :aria-label="t('contact.close')"
          @click.stop="dismissFirstHint"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="h-3 w-3" aria-hidden="true">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
    </transition>

    <!-- 悬浮入口：桌面胶囊 + 移动端圆形 -->
    <button
      v-if="!isOpen"
      type="button"
      :aria-label="t('contact.openTooltip')"
      class="floating-contact-btn"
      @click="open"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="floating-contact-icon"
        aria-hidden="true"
      >
        <!-- users icon: 社群语义比单纯 chat-bubble 准确 -->
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
        <circle cx="9" cy="7" r="4"/>
        <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
        <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
      </svg>
      <span class="floating-contact-label">{{ t('contact.label') }}</span>
      <!-- 自定义 tooltip：hover 立即出现，不依赖原生 title -->
      <span class="floating-contact-tooltip">{{ t('contact.openTooltip') }}</span>
    </button>

    <!-- 弹窗遮罩 -->
    <div
      v-if="isOpen"
      class="floating-contact-overlay"
      role="dialog"
      aria-modal="true"
      @click.self="close"
    >
      <div class="floating-contact-panel">
        <!-- 顶部：tab + 关闭 -->
        <div class="floating-contact-header">
          <div v-if="channels.length > 1" class="floating-contact-tabs">
            <button
              v-for="c in channels"
              :key="c.type"
              type="button"
              class="floating-contact-tab"
              :class="{ 'floating-contact-tab-active': c.type === activeType }"
              @click="selectTab(c.type)"
            >
              {{ tabLabel(c) }}
            </button>
          </div>
          <div v-else class="floating-contact-single-title">
            {{ activeChannel ? tabLabel(activeChannel) : '' }}
          </div>
          <button
            type="button"
            class="floating-contact-icon-btn ml-auto"
            :aria-label="t('contact.close')"
            @click="close"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4" aria-hidden="true">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <!-- 内容：二维码 + 文案（纯文本插值，禁用 v-html，防 XSS） -->
        <div v-if="activeChannel" class="floating-contact-body">
          <img
            class="floating-contact-qr"
            :src="activeChannel.qr_image"
            alt=""
          />
          <p v-if="activeChannel.description" class="floating-contact-desc">
            {{ activeChannel.description }}
          </p>
          <p v-if="activeChannel.extra_info" class="floating-contact-extra">
            {{ activeChannel.extra_info }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* === 桌面端：白底胶囊（图标 + 文字），微信绿仅作 accent === */
.floating-contact-btn {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 60;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 16px 0 14px;
  border-radius: 9999px;
  background-color: #ffffff;
  color: #111827;
  border: 1px solid rgba(17, 24, 39, 0.08);
  box-shadow: 0 8px 24px -10px rgba(17, 24, 39, 0.25);
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease, background-color 0.15s ease;
}

.floating-contact-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 14px 32px -10px rgba(7, 193, 96, 0.35);
  background-color: #f9fafb;
}

:global(.dark) .floating-contact-btn {
  background-color: #1f2937;
  color: #f3f4f6;
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow: 0 8px 24px -10px rgba(0, 0, 0, 0.6);
}

:global(.dark) .floating-contact-btn:hover {
  background-color: #111827;
  box-shadow: 0 14px 32px -10px rgba(7, 193, 96, 0.45);
}

.floating-contact-icon {
  width: 18px;
  height: 18px;
  color: #07c160; /* 微信绿仅落在图标上，作 accent */
  flex-shrink: 0;
}

.floating-contact-label {
  font-size: 14px;
  font-weight: 500;
  letter-spacing: 0.02em;
}

/* === 自定义 tooltip：hover 立即出现 === */
.floating-contact-tooltip {
  position: absolute;
  right: 0;
  bottom: calc(100% + 8px);
  padding: 6px 10px;
  font-size: 12px;
  color: #fff;
  background-color: rgba(17, 24, 39, 0.92);
  border-radius: 6px;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transform: translateY(4px);
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.floating-contact-btn:hover .floating-contact-tooltip,
.floating-contact-btn:focus-visible .floating-contact-tooltip {
  opacity: 1;
  transform: translateY(0);
}

/* === 移动端：缩成紧凑圆形按钮 === */
@media (max-width: 640px) {
  .floating-contact-btn {
    height: 48px;
    width: 48px;
    padding: 0;
    justify-content: center;
    right: 16px;
    bottom: 16px;
    background-color: #07c160;
    color: #fff;
    border-color: transparent;
  }
  .floating-contact-icon {
    width: 22px;
    height: 22px;
    color: #fff;
  }
  .floating-contact-label,
  .floating-contact-tooltip {
    display: none;
  }
}

/* === 首次轻气泡提示 === */
.floating-contact-hint {
  position: fixed;
  right: 24px;
  bottom: 80px;
  z-index: 59;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px 10px 14px;
  background-color: #111827;
  color: #fff;
  border-radius: 12px;
  font-size: 13px;
  box-shadow: 0 12px 28px -10px rgba(0, 0, 0, 0.4);
  cursor: pointer;
  max-width: 240px;
}

.floating-contact-hint::after {
  content: '';
  position: absolute;
  right: 22px;
  bottom: -6px;
  width: 12px;
  height: 12px;
  background-color: #111827;
  transform: rotate(45deg);
  border-radius: 2px;
}

.floating-contact-hint-text {
  flex: 1 1 auto;
  line-height: 1.4;
}

.floating-contact-hint-close {
  position: relative;
  z-index: 1;
  border: 0;
  background-color: transparent;
  color: rgba(255, 255, 255, 0.7);
  width: 22px;
  height: 22px;
  border-radius: 9999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.floating-contact-hint-close:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: #fff;
}

@media (max-width: 640px) {
  .floating-contact-hint {
    right: 16px;
    bottom: 72px;
  }
  .floating-contact-hint::after {
    right: 18px;
  }
}

.fc-fade-enter-active,
.fc-fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.fc-fade-enter-from,
.fc-fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}

/* === 弹窗（沿用原结构） === */
.floating-contact-overlay {
  position: fixed;
  inset: 0;
  z-index: 70;
  background-color: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.floating-contact-panel {
  width: 100%;
  max-width: 360px;
  background-color: #fff;
  border-radius: 16px;
  box-shadow: 0 16px 40px -8px rgba(0, 0, 0, 0.3);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

@media (max-width: 640px) {
  .floating-contact-panel {
    max-width: 90vw;
  }
}

:global(.dark) .floating-contact-panel {
  background-color: #1f2937;
  color: #f3f4f6;
}

.floating-contact-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 12px 0 12px;
}

.floating-contact-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  flex: 1 1 auto;
  min-width: 0;
}

.floating-contact-tab {
  padding: 6px 12px;
  font-size: 13px;
  border-radius: 9999px;
  border: 0;
  background-color: transparent;
  color: #6b7280;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.floating-contact-tab:hover {
  background-color: rgba(0, 0, 0, 0.04);
}

.floating-contact-tab-active {
  background-color: #07c160;
  color: #fff;
}

.floating-contact-tab-active:hover {
  background-color: #07c160;
}

.floating-contact-single-title {
  flex: 1 1 auto;
  font-size: 14px;
  font-weight: 600;
  color: #111827;
  padding: 6px 4px;
}

:global(.dark) .floating-contact-single-title {
  color: #f9fafb;
}

.floating-contact-icon-btn {
  border: 0;
  background-color: transparent;
  width: 32px;
  height: 32px;
  border-radius: 9999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.floating-contact-icon-btn:hover {
  background-color: rgba(0, 0, 0, 0.06);
  color: #111827;
}

:global(.dark) .floating-contact-icon-btn:hover {
  background-color: rgba(255, 255, 255, 0.08);
  color: #f9fafb;
}

.floating-contact-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 16px 20px 24px;
}

.floating-contact-qr {
  width: 220px;
  height: 220px;
  object-fit: contain;
  border-radius: 12px;
  background-color: #f9fafb;
  padding: 8px;
}

:global(.dark) .floating-contact-qr {
  background-color: #f3f4f6;
}

.floating-contact-desc {
  margin-top: 14px;
  font-size: 14px;
  line-height: 1.5;
  color: #374151;
  white-space: pre-line;
  word-break: break-word;
}

:global(.dark) .floating-contact-desc {
  color: #d1d5db;
}

.floating-contact-extra {
  margin-top: 8px;
  font-size: 12px;
  color: #6b7280;
  word-break: break-all;
}
</style>
