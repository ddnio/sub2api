<template>
  <BaseDialog :show="show" :title="t('admin.users.referralInfoTitle')" width="normal" @close="$emit('close')">
    <div v-if="loading" class="flex items-center justify-center py-8">
      <svg class="h-6 w-6 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </div>
    <div v-else class="space-y-4">
      <!-- Referral Code -->
      <div class="flex items-center justify-between rounded-lg bg-gray-50 px-4 py-3 dark:bg-dark-700">
        <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.users.referralCode') }}</span>
        <span class="font-mono font-bold text-gray-900 dark:text-white">{{ referralData?.referral_code || '-' }}</span>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <div class="rounded-lg border border-gray-200 px-4 py-3 dark:border-dark-600">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.users.referralInviteCount') }}</p>
          <p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{{ referralData?.invite_count ?? 0 }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 px-4 py-3 dark:border-dark-600">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('referral.totalRewarded') }}</p>
          <p class="mt-1 text-lg font-bold text-green-600 dark:text-green-400">${{ (referralData?.total_rewarded ?? 0).toFixed(2) }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 px-4 py-3 dark:border-dark-600">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('referral.pendingCount') }}</p>
          <p class="mt-1 text-lg font-bold text-amber-500 dark:text-amber-400">{{ referralData?.pending_count ?? 0 }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 px-4 py-3 dark:border-dark-600">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.users.referralInvitedBy') }}</p>
          <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
            {{ referralData?.invited_by?.inviter_email || '-' }}
          </p>
        </div>
      </div>

      <!-- Invite Records -->
      <div class="rounded-lg border border-gray-200 dark:border-dark-600">
        <div class="border-b border-gray-200 px-4 py-3 dark:border-dark-600">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('referral.inviteeList') }}</h3>
        </div>
        <div v-if="!referralData?.invite_records?.length" class="py-6 text-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('referral.noInvitees') }}
        </div>
        <table v-else class="w-full">
          <thead>
            <tr class="border-b border-gray-100 dark:border-dark-700">
              <th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">{{ t('referral.email') }}</th>
              <th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">{{ t('referral.date') }}</th>
              <th class="px-4 py-2 text-center text-xs font-medium uppercase text-gray-500 dark:text-gray-400">{{ t('referral.status') }}</th>
              <th class="px-4 py-2 text-right text-xs font-medium uppercase text-gray-500 dark:text-gray-400">{{ t('referral.reward') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="record in referralData.invite_records" :key="record.id">
              <td class="px-4 py-2 text-sm text-gray-900 dark:text-gray-200">{{ record.invitee_email }}</td>
              <td class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">{{ formatDate(record.created_at) }}</td>
              <td class="px-4 py-2 text-center text-sm">
                <span v-if="record.reward_granted_at" class="inline-flex items-center rounded-full bg-green-50 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/20 dark:text-green-400">
                  {{ t('referral.statusGranted') }}
                </span>
                <span v-else class="inline-flex items-center rounded-full bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900/20 dark:text-amber-400">
                  {{ t('referral.statusPending') }}
                </span>
              </td>
              <td class="px-4 py-2 text-right text-sm font-medium">
                <span v-if="record.reward_granted_at" class="text-green-600 dark:text-green-400">
                  +${{ record.inviter_rewarded.toFixed(2) }}
                </span>
                <span v-else class="text-gray-400 dark:text-gray-500">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end">
        <button @click="$emit('close')" class="btn btn-secondary">{{ t('common.close') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '@/api/client'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  user: AdminUser | null
}>()

defineEmits<{ close: [] }>()

const loading = ref(false)
const referralData = ref<any>(null)

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString()
}

watch(() => props.show, async (val) => {
  if (val && props.user) {
    loading.value = true
    try {
      const { data } = await apiClient.get(`/admin/users/${props.user.id}/referral`, { params: { page_size: 100 } })
      referralData.value = data
    } catch {
      referralData.value = null
    } finally {
      loading.value = false
    }
  }
})
</script>
