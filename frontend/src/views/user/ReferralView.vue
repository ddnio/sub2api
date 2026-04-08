<template>
  <AppLayout>
  <div class="mx-auto max-w-4xl space-y-6">
    <!-- Page Header -->
    <div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('referral.title') }}</h1>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('referral.description') }}</p>
    </div>

    <!-- Referral Code Card -->
    <div class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('referral.myCode') }}</h2>
      <div class="mt-4 flex items-center gap-3">
        <div class="flex-1 rounded-lg bg-gray-50 px-4 py-3 font-mono text-lg font-bold tracking-wider text-gray-900 dark:bg-dark-700 dark:text-white">
          {{ referralInfo?.referral_code || '...' }}
        </div>
        <button
          class="rounded-lg bg-primary-600 px-4 py-3 text-sm font-medium text-white hover:bg-primary-700 transition-colors"
          @click="copyReferralLink"
        >
          {{ t('referral.copyLink') }}
        </button>
      </div>
      <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
        {{ referralLink }}
      </p>
    </div>

    <!-- Reward Info Card -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
      <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('referral.totalInvited') }}</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ referralInfo?.total_invited ?? 0 }}</p>
      </div>
      <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('referral.totalRewarded') }}</p>
        <p class="mt-1 text-2xl font-bold text-green-600 dark:text-green-400">${{ (referralInfo?.total_rewarded ?? 0).toFixed(2) }}</p>
      </div>
      <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('referral.rewardPerInvite') }}</p>
        <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
          <span v-if="referralInfo && (referralInfo.inviter_reward_amount > 0 || referralInfo.invitee_reward_amount > 0)">
            {{ t('referral.inviterGets') }} ${{ referralInfo.inviter_reward_amount.toFixed(2) }}
            <span v-if="referralInfo.invitee_reward_amount > 0" class="text-sm text-gray-500">
              / {{ t('referral.inviteeGets') }} ${{ referralInfo.invitee_reward_amount.toFixed(2) }}
            </span>
          </span>
          <span v-else class="text-sm text-gray-400">{{ t('referral.noRewardConfigured') }}</span>
        </p>
      </div>
    </div>

    <!-- Invitee List -->
    <div class="rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
      <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-700">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('referral.inviteeList') }}</h2>
      </div>
      <div v-if="loading" class="flex items-center justify-center py-12">
        <svg class="h-6 w-6 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>
      <div v-else-if="referrals.length === 0" class="py-12 text-center text-sm text-gray-500 dark:text-gray-400">
        {{ t('referral.noInvitees') }}
      </div>
      <table v-else class="w-full">
        <thead>
          <tr class="border-b border-gray-100 dark:border-dark-700">
            <th class="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">{{ t('referral.email') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">{{ t('referral.date') }}</th>
            <th class="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-gray-400">{{ t('referral.reward') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
          <tr v-for="record in referrals" :key="record.id">
            <td class="px-6 py-3 text-sm text-gray-900 dark:text-gray-200">{{ record.invitee_email }}</td>
            <td class="px-6 py-3 text-sm text-gray-500 dark:text-gray-400">{{ formatDate(record.created_at) }}</td>
            <td class="px-6 py-3 text-right text-sm font-medium text-green-600 dark:text-green-400">
              +${{ record.inviter_rewarded.toFixed(2) }}
            </td>
          </tr>
        </tbody>
      </table>
      <!-- Pagination -->
      <div v-if="pagination && pagination.pages > 1" class="flex items-center justify-between border-t border-gray-200 px-6 py-3 dark:border-dark-700">
        <button
          :disabled="pagination.page <= 1"
          class="rounded px-3 py-1 text-sm text-gray-600 hover:bg-gray-100 disabled:opacity-50 dark:text-gray-400 dark:hover:bg-dark-700"
          @click="loadReferrals(pagination.page - 1)"
        >
          {{ t('common.previous') }}
        </button>
        <span class="text-sm text-gray-500 dark:text-gray-400">
          {{ pagination.page }} / {{ pagination.pages }}
        </span>
        <button
          :disabled="pagination.page >= pagination.pages"
          class="rounded px-3 py-1 text-sm text-gray-600 hover:bg-gray-100 disabled:opacity-50 dark:text-gray-400 dark:hover:bg-dark-700"
          @click="loadReferrals(pagination.page + 1)"
        >
          {{ t('common.next') }}
        </button>
      </div>
    </div>

    <!-- Copy Success Toast -->
    <transition name="fade">
      <div v-if="showCopyToast" class="fixed bottom-4 right-4 rounded-lg bg-green-600 px-4 py-2 text-sm text-white shadow-lg">
        {{ t('referral.linkCopied') }}
      </div>
    </transition>
  </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { getReferralInfo, getReferralList } from '@/api/referral'
import type { ReferralInfo, ReferralRecord } from '@/api/referral'

const { t } = useI18n()

const loading = ref(true)
const referralInfo = ref<ReferralInfo | null>(null)
const referrals = ref<ReferralRecord[]>([])
const pagination = ref<{ total: number; page: number; page_size: number; pages: number } | null>(null)
const showCopyToast = ref(false)

const referralLink = computed(() => {
  if (!referralInfo.value?.referral_code) return ''
  const base = window.location.origin
  return `${base}/register?ref=${referralInfo.value.referral_code}`
})

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString()
}

async function copyReferralLink() {
  if (!referralLink.value) return
  try {
    await navigator.clipboard.writeText(referralLink.value)
    showCopyToast.value = true
    setTimeout(() => { showCopyToast.value = false }, 2000)
  } catch {
    // fallback
  }
}

async function loadReferrals(page = 1) {
  try {
    const res = await getReferralList(page)
    referrals.value = res.data || []
    pagination.value = res.pagination
  } catch (e) {
    console.error('Failed to load referrals:', e)
  }
}

onMounted(async () => {
  try {
    const [info] = await Promise.all([
      getReferralInfo(),
      loadReferrals()
    ])
    referralInfo.value = info
  } catch (e) {
    console.error('Failed to load referral info:', e)
  } finally {
    loading.value = false
  }
})
</script>
