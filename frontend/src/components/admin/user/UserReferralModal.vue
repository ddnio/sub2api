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
      <div class="grid grid-cols-2 gap-3">
        <div class="rounded-lg border border-gray-200 px-4 py-3 dark:border-dark-600">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.users.referralInviteCount') }}</p>
          <p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{{ referralData?.invite_count ?? 0 }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 px-4 py-3 dark:border-dark-600">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.users.referralInvitedBy') }}</p>
          <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
            {{ referralData?.invited_by?.invitee_email || '-' }}
          </p>
        </div>
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

watch(() => props.show, async (val) => {
  if (val && props.user) {
    loading.value = true
    try {
      const { data } = await apiClient.get(`/admin/users/${props.user.id}/referral`)
      referralData.value = data
    } catch {
      referralData.value = null
    } finally {
      loading.value = false
    }
  }
})
</script>
