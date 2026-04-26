<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useI18n } from 'vue-i18n'
import type { ContactChannel } from '@/types'

const SESSION_DISMISSED_KEY = 'contact_dismissed'

// 排除路径前缀（admin / 安装向导 / 登录注册）— 避免遮挡管理界面与表单流
const EXCLUDED_PATH_PREFIXES = ['/admin', '/setup', '/login', '/register']

const route = useRoute()
const appStore = useAppStore()
const { t } = useI18n()

const isOpen = ref(false)
const dismissed = ref(readDismissed())
const activeType = ref<string>('')

function readDismissed(): boolean {
  try {
    return sessionStorage.getItem(SESSION_DISMISSED_KEY) === '1'
  } catch {
    return false
  }
}

function markDismissed() {
  try {
    sessionStorage.setItem(SESSION_DISMISSED_KEY, '1')
  } catch {
    // 忽略 storage 不可用（私密模式等）
  }
  dismissed.value = true
  isOpen.value = false
}

const channels = computed<ContactChannel[]>(() => {
  const list = appStore.cachedPublicSettings?.contact_channels ?? []
  // 后端已过滤 enabled + 排序，这里再防御一次
  return [...list]
    .filter((c) => c && c.enabled && c.qr_image)
    .sort((a, b) => a.priority - b.priority)
})

const isExcludedRoute = computed(() => {
  const p = route.path || '/'
  return EXCLUDED_PATH_PREFIXES.some((prefix) => p === prefix || p.startsWith(prefix + '/'))
})

const shouldRender = computed(() => {
  return !dismissed.value && !isExcludedRoute.value && channels.value.length > 0
})

const activeChannel = computed<ContactChannel | undefined>(() => {
  if (channels.value.length === 0) return undefined
  return channels.value.find((c) => c.type === activeType.value) ?? channels.value[0]
})

// 同步 activeType 到当前可用渠道列表的第一个
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

async function open() {
  // R2: 打开弹窗时刷一次最新公开配置（store 内部不强制时走缓存，开销可控）
  try {
    await appStore.fetchPublicSettings(false)
  } catch {
    // 网络错误时仍使用现有缓存渲染
  }
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
  // 渠道未填 label 时给默认 i18n 兜底
  return t(`contact.channelTypes.${c.type}`)
}
</script>

<template>
  <div v-if="shouldRender">
    <!-- 悬浮按钮 -->
    <button
      v-if="!isOpen"
      type="button"
      :title="t('contact.openTooltip')"
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
        <path
          d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"
        />
      </svg>
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
          <div class="ml-auto flex items-center gap-1">
            <button
              type="button"
              class="floating-contact-icon-btn"
              :title="t('contact.dismissSession')"
              :aria-label="t('contact.dismissSession')"
              @click="markDismissed"
            >
              <!-- bell-off 表示本会话不再展示 -->
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4" aria-hidden="true">
                <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
                <path d="M18.63 13A17.89 17.89 0 0 1 18 8"/>
                <path d="M6.26 6.26A5.86 5.86 0 0 0 6 8c0 7-3 9-3 9h14"/>
                <path d="M18 8a6 6 0 0 0-9.33-5"/>
                <line x1="1" y1="1" x2="23" y2="23"/>
              </svg>
            </button>
            <button
              type="button"
              class="floating-contact-icon-btn"
              :title="t('contact.close')"
              :aria-label="t('contact.close')"
              @click="close"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4" aria-hidden="true">
                <line x1="18" y1="6" x2="6" y2="18"/>
                <line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
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
.floating-contact-btn {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 60;
  width: 56px;
  height: 56px;
  border-radius: 9999px;
  background-color: #07c160;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 20px -4px rgba(7, 193, 96, 0.45);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
  border: 0;
  cursor: pointer;
}

.floating-contact-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 24px -4px rgba(7, 193, 96, 0.55);
}

.floating-contact-icon {
  width: 26px;
  height: 26px;
}

@media (max-width: 640px) {
  .floating-contact-btn {
    width: 48px;
    height: 48px;
    right: 16px;
    bottom: 16px;
  }
  .floating-contact-icon {
    width: 22px;
    height: 22px;
  }
}

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
  white-space: pre-line; /* 支持 admin 文案换行 */
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
