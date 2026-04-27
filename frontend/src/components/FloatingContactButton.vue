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

// 抽屉态：桌面端默认半隐，hover 滑出，离开后自动收回
const expanded = ref(false)
let firstHintTimerShow: ReturnType<typeof setTimeout> | null = null
let firstHintTimerHide: ReturnType<typeof setTimeout> | null = null
let expandTimer: ReturnType<typeof setTimeout> | null = null
let collapseTimer: ReturnType<typeof setTimeout> | null = null

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
    /* ignore: private mode */
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

function clearExpandTimers() {
  if (expandTimer) {
    clearTimeout(expandTimer)
    expandTimer = null
  }
  if (collapseTimer) {
    clearTimeout(collapseTimer)
    collapseTimer = null
  }
}

function scheduleExpand() {
  if (collapseTimer) {
    clearTimeout(collapseTimer)
    collapseTimer = null
  }
  if (expanded.value) return
  if (expandTimer) clearTimeout(expandTimer)
  // 150ms 意图判定，避免鼠标掠过时抖动
  expandTimer = setTimeout(() => {
    expanded.value = true
  }, 150)
}

function scheduleCollapse() {
  if (expandTimer) {
    clearTimeout(expandTimer)
    expandTimer = null
  }
  if (!expanded.value) return
  if (collapseTimer) clearTimeout(collapseTimer)
  collapseTimer = setTimeout(() => {
    expanded.value = false
  }, 600)
}

async function open() {
  // R2: 打开弹窗时刷一次最新公开配置（store 内部不强制时走缓存，开销可控）
  try {
    await appStore.fetchPublicSettings(false)
  } catch {
    /* network error: render with current cache */
  }
  dismissFirstHint()
  clearExpandTimers()
  expanded.value = false // 弹窗打开时入口收回，弹窗本身承担可视化
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
    // 首次提示时短暂展开抽屉，让用户注意到入口
    expanded.value = true
    firstHintTimerHide = setTimeout(() => {
      showFirstHint.value = false
      expanded.value = false
      markFirstHintSeen()
    }, 7000)
  }, 3000)
})

onBeforeUnmount(() => {
  if (firstHintTimerShow) clearTimeout(firstHintTimerShow)
  if (firstHintTimerHide) clearTimeout(firstHintTimerHide)
  clearExpandTimers()
})
</script>

<template>
  <div v-if="shouldRender">
    <!-- 首次访问轻气泡（仅展示一次，localStorage 记忆） -->
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 translate-y-1"
    >
      <div
        v-if="showFirstHint && !isOpen"
        class="fixed right-6 bottom-[88px] z-[59] inline-flex max-w-[240px] cursor-pointer items-center gap-2.5 rounded-xl bg-dark-900 px-3 py-2.5 text-[13px] text-white shadow-glass ring-1 ring-black/10 dark:bg-dark-700 dark:ring-white/10 max-sm:right-4 max-sm:bottom-[80px]"
        @click="open"
      >
        <span class="flex-1 leading-snug">{{ t('contact.firstHint') }}</span>
        <button
          type="button"
          class="relative z-10 flex h-5 w-5 items-center justify-center rounded-full text-white/70 transition-colors hover:bg-white/10 hover:text-white"
          :aria-label="t('contact.close')"
          @click.stop="dismissFirstHint"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="h-3 w-3" aria-hidden="true">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
        <span class="absolute right-[22px] -bottom-1.5 h-3 w-3 rotate-45 rounded-[2px] bg-dark-900 dark:bg-dark-700 max-sm:right-[18px]" aria-hidden="true"></span>
      </div>
    </transition>

    <!-- 抽屉容器：fixed 在右下角，监听 hover 触发展开/收回 -->
    <div
      v-if="!isOpen"
      class="fixed right-0 bottom-6 z-[60] max-sm:right-0 max-sm:bottom-4"
      @mouseenter="scheduleExpand"
      @mouseleave="scheduleCollapse"
    >
      <button
        type="button"
        :aria-label="t('contact.openTooltip')"
        class="group relative inline-flex h-11 items-center gap-2 rounded-l-full rounded-r-none border border-r-0 border-primary-500/60 bg-white pl-3.5 pr-4 text-primary-600 shadow-card transition-all duration-300 ease-out hover:border-primary-500 hover:shadow-card-hover hover:shadow-glow focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2 dark:bg-dark-800 dark:text-primary-400 dark:ring-1 dark:ring-white/5 dark:hover:bg-dark-700 dark:focus-visible:ring-offset-dark-900 max-sm:h-12 max-sm:gap-0 max-sm:rounded-full max-sm:rounded-r-none max-sm:border-r-0 max-sm:p-0 max-sm:pl-3 max-sm:pr-2"
        :class="expanded ? 'translate-x-0' : 'translate-x-[58%] max-sm:translate-x-[40%]'"
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
          class="h-[18px] w-[18px] flex-shrink-0 max-sm:h-[22px] max-sm:w-[22px]"
          aria-hidden="true"
        >
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
          <circle cx="8.5" cy="12" r="0.9" fill="currentColor" stroke="none"/>
          <circle cx="12" cy="12" r="0.9" fill="currentColor" stroke="none"/>
          <circle cx="15.5" cy="12" r="0.9" fill="currentColor" stroke="none"/>
        </svg>
        <span
          class="text-sm font-medium tracking-wide whitespace-nowrap transition-opacity duration-200 max-sm:hidden"
          :class="expanded ? 'opacity-100' : 'opacity-0'"
        >
          {{ t('contact.label') }}
        </span>
        <!-- tooltip：仅在抽屉收回时展示，提示用户可悬停 -->
        <span
          v-if="!expanded"
          class="pointer-events-none absolute right-0 bottom-[calc(100%+8px)] whitespace-nowrap rounded-md bg-dark-900/95 px-2.5 py-1.5 text-xs text-white opacity-0 shadow-glass-sm transition-opacity duration-150 group-hover:opacity-100 dark:bg-dark-700 max-sm:hidden"
        >
          {{ t('contact.openTooltip') }}
        </span>
      </button>
    </div>

    <!-- 弹窗遮罩 -->
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isOpen"
        class="fixed inset-0 z-[70] flex items-center justify-center bg-black/45 p-4 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        @click.self="close"
      >
        <div
          class="w-full max-w-[360px] overflow-hidden rounded-3xl bg-white shadow-2xl ring-1 ring-black/5 animate-scale-in dark:bg-dark-800 dark:ring-white/10 max-sm:max-w-[90vw]"
        >
          <div class="flex items-center gap-2 px-3 pt-3">
            <div v-if="channels.length > 1" class="flex min-w-0 flex-1 flex-wrap gap-1">
              <button
                v-for="c in channels"
                :key="c.type"
                type="button"
                class="rounded-full px-3 py-1.5 text-[13px] font-medium transition-colors"
                :class="
                  c.type === activeType
                    ? 'bg-primary-500 text-white shadow-glow'
                    : 'text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-dark-700'
                "
                @click="selectTab(c.type)"
              >
                {{ tabLabel(c) }}
              </button>
            </div>
            <div
              v-else
              class="flex-1 px-1 py-1.5 text-sm font-semibold text-gray-900 dark:text-gray-100"
            >
              {{ activeChannel ? tabLabel(activeChannel) : '' }}
            </div>
            <button
              type="button"
              class="ml-auto inline-flex h-8 w-8 items-center justify-center rounded-full text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-gray-100"
              :aria-label="t('contact.close')"
              @click="close"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4" aria-hidden="true">
                <line x1="18" y1="6" x2="6" y2="18"/>
                <line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>

          <div v-if="activeChannel" class="flex flex-col items-center px-5 pb-6 pt-4 text-center">
            <img
              :src="activeChannel.qr_image"
              alt=""
              class="h-[220px] w-[220px] rounded-xl bg-gray-50 object-contain p-2 ring-1 ring-black/5 dark:bg-gray-100 dark:ring-white/10"
            />
            <p
              v-if="activeChannel.description"
              class="mt-3.5 whitespace-pre-line break-words text-sm leading-relaxed text-gray-700 dark:text-gray-300"
            >
              {{ activeChannel.description }}
            </p>
            <p
              v-if="activeChannel.extra_info"
              class="mt-2 break-all text-xs text-gray-500 dark:text-gray-400"
            >
              {{ activeChannel.extra_info }}
            </p>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>
