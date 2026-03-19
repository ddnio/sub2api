<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <h1 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('adminPayment.plansTitle') }}</h1>
          <div class="flex flex-1 justify-end">
            <button @click="openCreateDialog" class="btn btn-primary">
              + {{ t('adminPayment.createPlan') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="items" :loading="loading">
          <template #cell-price="{ value }">
            ¥{{ (value as number).toFixed(2) }}
          </template>
          <template #cell-original_price="{ value }">
            <span v-if="value != null">¥{{ (value as number).toFixed(2) }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>
          <template #cell-is_active="{ value }">
            <span :class="['badge', value ? 'badge-success' : 'badge-gray']">
              {{ value ? t('common.enabled') : t('common.disabled') }}
            </span>
          </template>
          <template #cell-badge="{ value }">
            <span v-if="value" class="badge badge-warning">{{ value }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <button @click="openEditDialog(row as AdminPaymentPlan)" class="btn btn-secondary btn-sm">
                {{ t('common.edit') }}
              </button>
              <button @click="confirmDelete(row as AdminPaymentPlan)" class="btn btn-danger btn-sm">
                {{ t('common.delete') }}
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

    <!-- 创建/编辑弹窗 -->
    <Teleport to="body">
      <div v-if="showDialog" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="fixed inset-0 bg-black/50" @click="showDialog = false"></div>
        <div class="relative z-10 w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-6 shadow-xl dark:bg-gray-900" style="max-height:90vh">
          <h2 class="mb-5 text-lg font-semibold">
            {{ editingPlan ? t('adminPayment.editPlan') : t('adminPayment.createPlan') }}
          </h2>
          <form @submit.prevent="submitPlan" class="space-y-4">
            <div>
              <label class="input-label">{{ t('adminPayment.planName') }} *</label>
              <input v-model="form.name" required class="input" />
            </div>
            <div>
              <label class="input-label">{{ t('adminPayment.planDescription') }}</label>
              <textarea v-model="form.description" rows="2" class="input"></textarea>
            </div>
            <div>
              <label class="input-label">{{ t('adminPayment.planBadge') }}</label>
              <input v-model="form.badge" class="input" placeholder="推荐、热门..." />
            </div>
            <div>
              <label class="input-label">{{ t('adminPayment.planGroup') }} *</label>
              <select v-model="form.group_id" required class="input">
                <option value="">— {{ t('common.select') }} —</option>
                <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label">{{ t('adminPayment.planDuration') }} *</label>
                <input v-model.number="form.duration_days" type="number" min="1" required class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('adminPayment.planSortOrder') }}</label>
                <input v-model.number="form.sort_order" type="number" min="0" class="input" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label">{{ t('adminPayment.planPrice') }} *</label>
                <input v-model.number="form.price" type="number" min="0" step="0.01" required class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('adminPayment.planOriginalPrice') }}</label>
                <input v-model.number="form.original_price" type="number" min="0" step="0.01" class="input" placeholder="可选" />
              </div>
            </div>
            <div class="flex items-center gap-2">
              <input id="is_active" v-model="form.is_active" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600" />
              <label for="is_active" class="text-sm text-gray-700 dark:text-gray-300">{{ t('adminPayment.planIsActive') }}</label>
            </div>
            <div class="flex justify-end gap-3 pt-2">
              <button type="button" @click="showDialog = false" class="btn btn-secondary">{{ t('common.cancel') }}</button>
              <button type="submit" :disabled="submitting" class="btn btn-primary">
                {{ submitting ? t('common.saving') : t('common.save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- 删除确认 -->
    <Teleport to="body">
      <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="fixed inset-0 bg-black/50" @click="showDeleteConfirm = false"></div>
        <div class="relative z-10 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl dark:bg-gray-900">
          <h2 class="mb-2 text-base font-semibold">{{ t('adminPayment.deletePlan') }}</h2>
          <p class="mb-5 text-sm text-gray-500">
            {{ t('adminPayment.confirmDelete', { name: deletingPlan?.name }) }}
          </p>
          <div class="flex justify-end gap-3">
            <button @click="showDeleteConfirm = false" class="btn btn-secondary">{{ t('common.cancel') }}</button>
            <button @click="doDelete" :disabled="submitting" class="btn btn-danger">{{ t('common.delete') }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI, type AdminPaymentPlan } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { useTableLoader } from '@/composables/useTableLoader'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'

const { t } = useI18n()
const appStore = useAppStore()

const columns = [
  { key: 'name', label: t('adminPayment.planName') },
  { key: 'group_name', label: t('adminPayment.planGroup') },
  { key: 'duration_days', label: t('adminPayment.planDuration') },
  { key: 'price', label: t('adminPayment.planPrice') },
  { key: 'original_price', label: t('adminPayment.planOriginalPrice') },
  { key: 'badge', label: t('payment.badge') },
  { key: 'is_active', label: t('adminPayment.planIsActive') },
  { key: 'actions', label: '' }
]

const { items, loading, pagination, load, handlePageChange } = useTableLoader({
  fetchFn: (page, pageSize, params, options) =>
    adminAPI.payment.listPlans(page, pageSize, params, options),
  initialParams: {}
})

// 分组列表
interface Group { id: number; name: string }
const groups = ref<Group[]>([])
const loadGroups = async () => {
  try {
    const data = await adminAPI.groups.getAll()
    groups.value = data
  } catch {
    // ignore
  }
}

// 表单
const showDialog = ref(false)
const submitting = ref(false)
const editingPlan = ref<AdminPaymentPlan | null>(null)

const defaultForm = () => ({
  name: '',
  description: '',
  badge: '',
  group_id: '' as number | '',
  duration_days: 30,
  price: 0,
  original_price: null as number | null,
  sort_order: 0,
  is_active: true
})

const form = reactive(defaultForm())

const openCreateDialog = () => {
  editingPlan.value = null
  Object.assign(form, defaultForm())
  showDialog.value = true
}

const openEditDialog = (plan: AdminPaymentPlan) => {
  editingPlan.value = plan
  Object.assign(form, {
    name: plan.name,
    description: plan.description ?? '',
    badge: plan.badge ?? '',
    group_id: plan.group_id,
    duration_days: plan.duration_days,
    price: plan.price,
    original_price: plan.original_price ?? null,
    sort_order: plan.sort_order,
    is_active: plan.is_active
  })
  showDialog.value = true
}

const submitPlan = async () => {
  submitting.value = true
  try {
    const payload = {
      name: form.name,
      description: form.description,
      badge: form.badge || null,
      group_id: form.group_id as number,
      duration_days: form.duration_days,
      price: form.price,
      original_price: form.original_price || null,
      sort_order: form.sort_order,
      is_active: form.is_active
    }
    if (editingPlan.value) {
      await adminAPI.payment.updatePlan(editingPlan.value.id, payload)
    } else {
      await adminAPI.payment.createPlan(payload)
    }
    showDialog.value = false
    load()
    appStore.showSuccess(t('common.saved'))
  } catch (e: any) {
    appStore.showError(e?.response?.data?.message || t('common.error'))
  } finally {
    submitting.value = false
  }
}

// 删除
const showDeleteConfirm = ref(false)
const deletingPlan = ref<AdminPaymentPlan | null>(null)

const confirmDelete = (plan: AdminPaymentPlan) => {
  deletingPlan.value = plan
  showDeleteConfirm.value = true
}

const doDelete = async () => {
  if (!deletingPlan.value) return
  submitting.value = true
  try {
    await adminAPI.payment.deletePlan(deletingPlan.value.id)
    showDeleteConfirm.value = false
    load()
    appStore.showSuccess(t('common.deleted'))
  } catch (e: any) {
    appStore.showError(e?.response?.data?.message || t('common.error'))
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  load()
  loadGroups()
})
</script>
