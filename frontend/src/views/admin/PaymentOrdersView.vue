<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <Select
            v-model="params.status"
            :options="statusOptions"
            class="w-36"
            @change="reload"
          />
          <Select
            v-model="params.order_type"
            :options="typeOptions"
            class="w-32"
            @change="reload"
          />
          <div class="flex flex-1 justify-end gap-2">
            <button @click="showStats = !showStats" class="btn btn-secondary">
              {{ t('adminPayment.statsTitle') }}
            </button>
            <button @click="reload" :disabled="loading" class="btn btn-secondary">
              <svg :class="['h-4 w-4', loading ? 'animate-spin' : '']" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>
          </div>
        </div>
      </template>

      <!-- 统计卡片 -->
      <template v-if="showStats && stats" #header>
        <div class="mb-4 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
          <div class="card p-4 text-center">
            <p class="text-xs text-gray-500">{{ t('adminPayment.totalOrders') }}</p>
            <p class="text-xl font-bold">{{ stats.total_orders }}</p>
          </div>
          <div class="card p-4 text-center">
            <p class="text-xs text-gray-500">{{ t('adminPayment.totalAmount') }}</p>
            <p class="text-xl font-bold text-primary-600">¥{{ stats.total_amount.toFixed(2) }}</p>
          </div>
          <div class="card p-4 text-center">
            <p class="text-xs text-gray-500">{{ t('adminPayment.paidOrders') }}</p>
            <p class="text-xl font-bold">{{ stats.paid_orders }}</p>
          </div>
          <div class="card p-4 text-center">
            <p class="text-xs text-gray-500">{{ t('adminPayment.paidAmount') }}</p>
            <p class="text-xl font-bold text-blue-600">¥{{ stats.paid_amount.toFixed(2) }}</p>
          </div>
          <div class="card p-4 text-center">
            <p class="text-xs text-gray-500">{{ t('adminPayment.completedOrders') }}</p>
            <p class="text-xl font-bold">{{ stats.completed_orders }}</p>
          </div>
          <div class="card p-4 text-center">
            <p class="text-xs text-gray-500">{{ t('adminPayment.completedAmount') }}</p>
            <p class="text-xl font-bold text-green-600">¥{{ stats.completed_amount.toFixed(2) }}</p>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="items" :loading="loading">
          <template #cell-amount="{ value }">¥{{ (value as number).toFixed(2) }}</template>
          <template #cell-order_type="{ value }">
            <span :class="['badge', value === 'subscription' ? 'badge-primary' : 'badge-success']">
              {{ t('payment.orderType.' + (value as string)) }}
            </span>
          </template>
          <template #cell-status="{ value }">
            <span :class="['badge', statusBadge(value as string)]">
              {{ t('payment.orderStatus.' + (value as string).toLowerCase()) }}
            </span>
          </template>
          <template #cell-created_at="{ value }">
            {{ formatDate(value as string) }}
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <button
                v-if="(row as AdminPaymentOrder).status === 'PAID'"
                @click="openAction(row as AdminPaymentOrder, 'complete')"
                class="btn btn-success btn-sm"
              >
                {{ t('adminPayment.completeOrder') }}
              </button>
              <button
                v-if="['PAID', 'COMPLETED'].includes((row as AdminPaymentOrder).status)"
                @click="openAction(row as AdminPaymentOrder, 'refund')"
                class="btn btn-danger btn-sm"
              >
                {{ t('adminPayment.refundOrder') }}
              </button>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > pagination.page_size"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
        />
      </template>
    </TablePageLayout>

    <!-- 操作确认弹窗 -->
    <Teleport to="body">
      <div v-if="showActionDialog" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="fixed inset-0 bg-black/50" @click="showActionDialog = false"></div>
        <div class="relative z-10 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl dark:bg-gray-900">
          <h2 class="mb-2 text-base font-semibold">
            {{ actionType === 'complete' ? t('adminPayment.completeOrder') : t('adminPayment.refundOrder') }}
          </h2>
          <p class="mb-3 text-sm text-gray-500">
            {{ actionType === 'complete' ? t('adminPayment.confirmComplete') : t('adminPayment.confirmRefund') }}
          </p>
          <div class="mb-1 text-xs text-gray-400">{{ t('payment.orderNo') }}: {{ actionOrder?.out_trade_no }}</div>
          <div class="mb-4 text-sm font-medium">¥{{ actionOrder?.amount?.toFixed(2) }}</div>
          <div class="mb-4">
            <label class="input-label">{{ t('adminPayment.adminNote') }}</label>
            <input v-model="actionNote" class="input" :placeholder="t('adminPayment.adminNote')" />
          </div>
          <div class="flex justify-end gap-3">
            <button @click="showActionDialog = false" class="btn btn-secondary">{{ t('common.cancel') }}</button>
            <button @click="doAction" :disabled="submitting" :class="['btn', actionType === 'complete' ? 'btn-success' : 'btn-danger']">
              {{ submitting ? t('common.saving') : t('common.confirm') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI, type AdminPaymentOrder, type OrderStats } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { useTableLoader } from '@/composables/useTableLoader'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'

const { t } = useI18n()
const appStore = useAppStore()

const columns = [
  { key: 'out_trade_no', label: t('payment.orderNo') },
  { key: 'email', label: t('adminPayment.orderUser') },
  { key: 'order_type', label: t('payment.type') },
  { key: 'amount', label: t('payment.amount') },
  { key: 'status', label: t('payment.status') },
  { key: 'payment_type', label: t('payment.provider') },
  { key: 'created_at', label: t('payment.createdAt') },
  { key: 'actions', label: '' }
]

const statusOptions = [
  { value: '', label: t('common.all') },
  { value: 'PENDING', label: t('payment.orderStatus.pending') },
  { value: 'PAID', label: t('payment.orderStatus.paid') },
  { value: 'COMPLETED', label: t('payment.orderStatus.completed') },
  { value: 'FAILED', label: t('payment.orderStatus.failed') },
  { value: 'EXPIRED', label: t('payment.orderStatus.expired') },
  { value: 'REFUNDED', label: t('payment.orderStatus.refunded') }
]

const typeOptions = [
  { value: '', label: t('common.all') },
  { value: 'subscription', label: t('payment.orderType.subscription') },
  { value: 'balance', label: t('payment.orderType.balance') }
]

const { items, loading, params, pagination, load, reload, handlePageChange } = useTableLoader({
  fetchFn: (page, pageSize, p, options) =>
    adminAPI.payment.listOrders(page, pageSize, p, options),
  initialParams: { status: '', order_type: '' }
})

// 统计
const showStats = ref(false)
const stats = ref<OrderStats | null>(null)
const loadStats = async () => {
  try {
    stats.value = await adminAPI.payment.getOrderStats({})
  } catch {
    // ignore
  }
}

// 操作
const showActionDialog = ref(false)
const actionType = ref<'complete' | 'refund'>('complete')
const actionOrder = ref<AdminPaymentOrder | null>(null)
const actionNote = ref('')
const submitting = ref(false)

const openAction = (order: AdminPaymentOrder, type: 'complete' | 'refund') => {
  actionOrder.value = order
  actionType.value = type
  actionNote.value = ''
  showActionDialog.value = true
}

const doAction = async () => {
  if (!actionOrder.value) return
  submitting.value = true
  try {
    if (actionType.value === 'complete') {
      await adminAPI.payment.completeOrder(actionOrder.value.id, actionNote.value || undefined)
    } else {
      await adminAPI.payment.refundOrder(actionOrder.value.id, actionNote.value || undefined)
    }
    showActionDialog.value = false
    reload()
    loadStats()
    appStore.showSuccess(t('common.saved'))
  } catch (e: any) {
    appStore.showError(e?.response?.data?.message || t('common.error'))
  } finally {
    submitting.value = false
  }
}

const statusBadge = (status: string) => {
  const map: Record<string, string> = {
    PENDING: 'badge-warning',
    PAID: 'badge-primary',
    COMPLETED: 'badge-success',
    FAILED: 'badge-danger',
    EXPIRED: 'badge-gray',
    REFUNDED: 'badge-gray'
  }
  return map[status] ?? 'badge-gray'
}

const formatDate = (dt: string) => new Date(dt).toLocaleString()

onMounted(() => {
  load()
  loadStats()
})
</script>
