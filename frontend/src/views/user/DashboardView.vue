<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-12"><LoadingSpinner /></div>
      <template v-else-if="stats">
        <div
          v-if="showFirstCallBanner"
          data-testid="first-call-banner"
          class="flex flex-col gap-3 rounded-xl border border-primary-200 bg-primary-50/80 px-4 py-3 text-sm sm:flex-row sm:items-center sm:justify-between dark:border-primary-800/60 dark:bg-primary-900/20"
        >
          <div class="flex items-center gap-2 text-gray-800 dark:text-gray-200">
            <span class="font-medium text-gray-900 dark:text-white">
              已创建 API Key，还没完成首次调用？
            </span>
            <router-link
              to="/start?step=config&use=codex"
              data-testid="first-call-banner-link"
              class="font-medium text-primary-700 underline-offset-4 hover:underline dark:text-primary-300"
            >
              打开 Codex 一键配置 / 导入 CC Switch →
            </router-link>
          </div>
          <button
            type="button"
            data-testid="first-call-banner-dismiss"
            class="self-end text-xs text-gray-500 transition-colors hover:text-gray-700 sm:self-auto dark:text-dark-400 dark:hover:text-gray-200"
            @click="dismissBanner"
          >
            关闭
          </button>
        </div>
        <UserDashboardStats :stats="stats" :balance="user?.balance || 0" :is-simple="authStore.isSimpleMode" :platform-quotas="platformQuotas" />
        <AffiliateInviteCard />
        <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :trend="trendData" :models="modelStats" @dateRangeChange="loadCharts" @granularityChange="loadCharts" @refresh="refreshAll" />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
          <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
        </div>
      </template>
      <div v-else-if="loadError" class="flex flex-col items-center justify-center py-16 text-center">
        <svg class="mb-4 h-12 w-12 text-gray-400 dark:text-dark-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <p class="text-gray-500 dark:text-dark-400">{{ t('dashboard.failedToLoad') }}</p>
        <button @click="refreshAll" class="mt-4 rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700">
          {{ t('dashboard.retry') }}
        </button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'; import { useI18n } from 'vue-i18n'; import { useAuthStore } from '@/stores/auth'; import { useAppStore } from '@/stores/app'; import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'; import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'; import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import AffiliateInviteCard from '@/components/user/AffiliateInviteCard.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'; import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat, PlatformQuotaItem } from '@/types'
import { getMyPlatformQuotas } from '@/api/user'
import { formatDateLocalInput } from '@/utils/format'
import { dismissFirstCallBanner, shouldShowFirstCallBanner } from './startOnboarding'

const { t } = useI18n()
const authStore = useAuthStore(); const appStore = useAppStore(); const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null); const loading = ref(false); const loadError = ref(false); const loadingUsage = ref(false); const loadingCharts = ref(false)
const bannerDismissed = ref(false)
const showFirstCallBanner = computed(() =>
  !bannerDismissed.value &&
  shouldShowFirstCallBanner(user.value, stats.value, {
    defaultSignupBalance: appStore.cachedPublicSettings?.default_balance
  })
)
const dismissBanner = () => { dismissFirstCallBanner(user.value?.id); bannerDismissed.value = true }
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const recentUsage = ref<UsageLog[]>([])
const platformQuotas = ref<PlatformQuotaItem[] | null>(null)

const startDate = ref(formatDateLocalInput(new Date(Date.now() - 6 * 86400000))); const endDate = ref(formatDateLocalInput(new Date())); const granularity = ref('day')

const loadStats = async () => { loading.value = true; loadError.value = false; try { await authStore.refreshUser(); stats.value = await usageAPI.getDashboardStats() } catch (error) { console.error('Failed to load dashboard stats:', error); loadError.value = true } finally { loading.value = false } }
const loadCharts = async () => { loadingCharts.value = true; try { const res = await Promise.all([usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }), usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })]); trendData.value = res[0].trend || []; modelStats.value = res[1].models || [] } catch (error) { console.error('Failed to load charts:', error) } finally { loadingCharts.value = false } }
const loadRecent = async () => { loadingUsage.value = true; try { const res = await usageAPI.getByDateRange(startDate.value, endDate.value); recentUsage.value = res.items.slice(0, 5) } catch (error) { console.error('Failed to load recent usage:', error) } finally { loadingUsage.value = false } }
const loadPlatformQuotas = async () => { try { const data = await getMyPlatformQuotas(); platformQuotas.value = data.platform_quotas ?? [] } catch (error) { console.warn('Failed to load platform quotas:', error); platformQuotas.value = [] } }
const refreshAll = () => { loadStats(); loadCharts(); loadRecent(); loadPlatformQuotas() }

onMounted(() => { refreshAll() })
</script>
