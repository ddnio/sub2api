<template>
  <AppLayout>
    <!-- 标签页切换 -->
    <div class="mb-6 border-b border-gray-200 dark:border-gray-700">
      <nav class="-mb-px flex space-x-6">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          :class="[
            'border-b-2 pb-3 text-sm font-medium transition-colors',
            activeTab === tab.key
              ? 'border-primary-500 text-primary-600 dark:text-primary-400'
              : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400'
          ]"
        >
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <!-- 订阅套餐 -->
    <div v-if="activeTab === 'plans'">
      <div v-if="plansLoading" class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="n in 3" :key="n" class="card animate-pulse p-6">
          <div class="mb-3 h-5 w-2/3 rounded bg-gray-200 dark:bg-gray-700"></div>
          <div class="mb-4 h-8 w-1/2 rounded bg-gray-200 dark:bg-gray-700"></div>
          <div class="h-9 w-full rounded bg-gray-200 dark:bg-gray-700"></div>
        </div>
      </div>
      <div v-else-if="plans.length === 0" class="card p-12 text-center text-gray-500">
        {{ t('payment.noOrders') }}
      </div>
      <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="plan in plans"
          :key="plan.id"
          class="card relative overflow-hidden p-6 transition-shadow hover:shadow-md"
        >
          <!-- 徽章 -->
          <div v-if="plan.badge" class="absolute right-4 top-4">
            <span class="badge badge-warning text-xs">{{ plan.badge }}</span>
          </div>
          <h3 class="mb-1 text-base font-semibold text-gray-900 dark:text-white">{{ plan.name }}</h3>
          <p v-if="plan.description" class="mb-3 text-sm text-gray-500">{{ plan.description }}</p>
          <div class="mb-1 flex items-baseline gap-1">
            <span class="text-2xl font-bold text-primary-600 dark:text-primary-400">¥{{ plan.price.toFixed(2) }}</span>
            <span class="text-sm text-gray-400">/ {{ plan.duration_days }} {{ t('payment.durationUnit') }}</span>
          </div>
          <p v-if="plan.original_price" class="mb-4 text-sm text-gray-400 line-through">
            {{ t('payment.originalPrice') }} ¥{{ plan.original_price.toFixed(2) }}
          </p>
          <div v-else class="mb-4"></div>
          <button
            class="btn btn-primary w-full"
            @click="openPlanPayment(plan)"
          >
            {{ t('payment.payNow') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 余额充值 -->
    <div v-if="activeTab === 'topup'">
      <div class="mx-auto max-w-md">
        <div class="card p-6">
          <div class="mb-5">
            <label class="input-label">{{ t('payment.topupAmount') }}</label>
            <div class="mt-1 flex items-center gap-2">
              <span class="text-lg font-medium text-gray-500">¥</span>
              <input
                v-model.number="topupAmount"
                type="number"
                min="1"
                step="0.01"
                :placeholder="t('payment.topupAmountPlaceholder')"
                class="input flex-1"
              />
            </div>
            <p class="mt-1 text-xs text-gray-400">
              {{ t('payment.topupMin', { min: 1 }) }} · {{ t('payment.topupMax', { max: 10000 }) }}
            </p>
          </div>
          <!-- 快捷金额 -->
          <div class="mb-5 flex flex-wrap gap-2">
            <button
              v-for="amount in quickAmounts"
              :key="amount"
              @click="topupAmount = amount"
              :class="[
                'rounded-lg border px-4 py-2 text-sm font-medium transition-colors',
                topupAmount === amount
                  ? 'border-primary-500 bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
                  : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-gray-700 dark:text-gray-400'
              ]"
            >
              ¥{{ amount }}
            </button>
          </div>
          <!-- 支付方式 -->
          <div class="mb-5">
            <label class="input-label">{{ t('payment.selectProvider') }}</label>
            <div class="mt-2 flex gap-3">
              <button
                v-for="p in providers"
                :key="p.value"
                @click="selectedProvider = p.value"
                :class="[
                  'flex flex-1 items-center justify-center gap-2 rounded-xl border-2 py-3 text-sm font-medium transition-colors',
                  selectedProvider === p.value
                    ? 'border-primary-500 bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
                    : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-gray-700 dark:text-gray-400'
                ]"
              >
                <span class="text-lg">{{ p.icon }}</span>
                {{ p.label }}
              </button>
            </div>
          </div>
          <button
            class="btn btn-primary w-full"
            :disabled="!topupAmount || topupAmount <= 0 || creatingOrder"
            @click="createTopupOrder"
          >
            {{ creatingOrder ? t('payment.paying') : t('payment.confirmPayment') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 订单记录 -->
    <div v-if="activeTab === 'orders'">
      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-gray-800">
              <tr>
                <th class="px-4 py-3 text-left font-medium text-gray-500">{{ t('payment.orderNo') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500">{{ t('payment.type') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500">{{ t('payment.amount') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500">{{ t('payment.status') }}</th>
                <th class="px-4 py-3 text-left font-medium text-gray-500">{{ t('payment.createdAt') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
              <tr v-if="ordersLoading">
                <td colspan="5" class="py-12 text-center text-gray-400">
                  <div class="mx-auto h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
                </td>
              </tr>
              <tr v-else-if="orders.length === 0">
                <td colspan="5" class="py-12 text-center text-gray-400">{{ t('payment.noOrders') }}</td>
              </tr>
              <tr v-for="order in orders" :key="order.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                <td class="px-4 py-3 font-mono text-xs text-gray-500">{{ order.order_no }}</td>
                <td class="px-4 py-3">
                  <span :class="['badge', order.type === 'plan' ? 'badge-primary' : 'badge-success']">
                    {{ t('payment.orderType.' + order.type) }}
                  </span>
                </td>
                <td class="px-4 py-3 font-medium">¥{{ order.amount.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <span :class="['badge', statusBadgeClass(order.status)]">
                    {{ t('payment.orderStatus.' + order.status) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-gray-500">{{ formatDateTime(order.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="orderPagination.total > orderPagination.page_size" class="border-t px-4 py-3">
          <Pagination
            :page="orderPagination.page"
            :total="orderPagination.total"
            :page-size="orderPagination.page_size"
            @update:page="handleOrderPageChange"
          />
        </div>
      </div>
    </div>

    <!-- 支付二维码弹窗 -->
    <Teleport to="body">
      <div v-if="showPayDialog" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="fixed inset-0 bg-black/50" @click="closePayDialog"></div>
        <div class="relative z-10 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl dark:bg-gray-900">
          <button @click="closePayDialog" class="absolute right-4 top-4 text-gray-400 hover:text-gray-600">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>

          <!-- 支付成功 -->
          <div v-if="payStatus === 'completed'" class="py-4 text-center">
            <div class="mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <svg class="h-8 w-8 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <p class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('payment.paymentSuccess') }}</p>
            <button class="btn btn-primary mt-4 w-full" @click="closePayDialog">{{ t('payment.closeDialog') }}</button>
          </div>

          <!-- 支付超时 -->
          <div v-else-if="payStatus === 'timeout'" class="py-4 text-center">
            <div class="mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
              <svg class="h-8 w-8 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <p class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('payment.paymentTimeout') }}</p>
            <button class="btn btn-secondary mt-4 w-full" @click="closePayDialog">{{ t('payment.closeDialog') }}</button>
          </div>

          <!-- 等待支付 -->
          <div v-else class="text-center">
            <p class="mb-1 text-base font-semibold text-gray-900 dark:text-white">
              {{ t('payment.scanQRCode', { provider: selectedProvider === 'wxpay' ? t('payment.wxpay') : t('payment.alipay') }) }}
            </p>
            <p class="mb-4 text-sm text-gray-500">
              ¥{{ currentOrder?.amount?.toFixed(2) }}
            </p>
            <!-- QR 码 canvas -->
            <div class="flex justify-center">
              <canvas ref="qrCanvas" class="rounded-xl border p-2" width="200" height="200"></canvas>
            </div>
            <!-- 倒计时 -->
            <div v-if="countdownSec > 0" class="mt-3 text-sm text-gray-500">
              {{ t('payment.qrCodeExpires') }}: <span class="font-mono font-medium text-gray-700 dark:text-gray-300">{{ formatCountdown(countdownSec) }}</span>
            </div>
            <!-- 轮询状态 -->
            <p class="mt-2 text-xs text-gray-400">{{ t('payment.waitingForPayment') }}</p>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import { paymentAPI, type PaymentPlan, type PaymentOrder } from '@/api'
import { useAppStore } from '@/stores/app'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'

const { t } = useI18n()
const appStore = useAppStore()

// 标签页
const tabs = computed(() => [
  { key: 'plans' as TabKey, label: t('payment.plans') },
  { key: 'topup' as TabKey, label: t('payment.topup') },
  { key: 'orders' as TabKey, label: t('payment.orderHistory') }
])
type TabKey = 'plans' | 'topup' | 'orders'
const activeTab = ref<TabKey>('plans')

// 套餐
const plans = ref<PaymentPlan[]>([])
const plansLoading = ref(false)

const loadPlans = async () => {
  plansLoading.value = true
  try {
    plans.value = await paymentAPI.listPlans()
  } catch {
    appStore.showError(t('common.error'))
  } finally {
    plansLoading.value = false
  }
}

// 充值
const topupAmount = ref<number | null>(null)
const quickAmounts = [10, 30, 50, 100, 200, 500]

// 支付方式
const selectedProvider = ref<'wxpay' | 'alipay'>('wxpay')
const providers = computed(() => [
  { value: 'wxpay' as const, label: t('payment.wxpay'), icon: '💚' }
])

// 订单列表
const orders = ref<PaymentOrder[]>([])
const ordersLoading = ref(false)
const orderPagination = ref({ page: 1, page_size: 20, total: 0 })

const loadOrders = async (page = 1) => {
  ordersLoading.value = true
  try {
    const res = await paymentAPI.listOrders(page, orderPagination.value.page_size, {})
    orders.value = res.items || []
    orderPagination.value.total = res.total || 0
    orderPagination.value.page = page
  } catch {
    appStore.showError(t('common.error'))
  } finally {
    ordersLoading.value = false
  }
}

const handleOrderPageChange = (page: number) => loadOrders(page)

// 创建订单 & 支付弹窗
const showPayDialog = ref(false)
const creatingOrder = ref(false)
const currentOrder = ref<PaymentOrder | null>(null)
const qrCodeURL = ref('')
const payStatus = ref<'waiting' | 'completed' | 'timeout'>('waiting')
const countdownSec = ref(0)
const qrCanvas = ref<HTMLCanvasElement | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null

const openPlanPayment = async (plan: PaymentPlan) => {
  creatingOrder.value = true
  try {
    const res = await paymentAPI.createOrder({
      type: 'plan',
      plan_id: plan.id,
      provider: selectedProvider.value
    })
    currentOrder.value = res.order
    qrCodeURL.value = res.qr_code_url
    openPayDialog()
  } catch (e: any) {
    appStore.showError(e?.response?.data?.message || t('common.error'))
  } finally {
    creatingOrder.value = false
  }
}

const createTopupOrder = async () => {
  if (!topupAmount.value || topupAmount.value <= 0) return
  creatingOrder.value = true
  try {
    const res = await paymentAPI.createOrder({
      type: 'topup',
      amount: topupAmount.value,
      provider: selectedProvider.value
    })
    currentOrder.value = res.order
    qrCodeURL.value = res.qr_code_url
    openPayDialog()
  } catch (e: any) {
    appStore.showError(e?.response?.data?.message || t('common.error'))
  } finally {
    creatingOrder.value = false
  }
}

const openPayDialog = () => {
  payStatus.value = 'waiting'
  // M1: Use order.expired_at for accurate countdown instead of hardcoded 1800s
  const expiredAt = currentOrder.value?.expired_at
  countdownSec.value = expiredAt
    ? Math.max(0, Math.floor((new Date(expiredAt).getTime() - Date.now()) / 1000))
    : 900
  showPayDialog.value = true
  nextTick(() => renderQRCode())
  startPolling()
  startCountdown()
}

const renderQRCode = async () => {
  if (!qrCanvas.value || !qrCodeURL.value) return
  try {
    await QRCode.toCanvas(qrCanvas.value, qrCodeURL.value, {
      width: 200,
      margin: 1,
      color: { dark: '#1a1a1a', light: '#ffffff' }
    })
  } catch (e) {
    console.error('QR code render error:', e)
  }
}

const startPolling = () => {
  if (pollTimer) clearInterval(pollTimer)
  pollTimer = setInterval(async () => {
    if (!currentOrder.value) return
    try {
      const res = await paymentAPI.getOrderStatus(currentOrder.value.id)
      if (res.status === 'completed') {
        payStatus.value = 'completed'
        stopPolling()
        loadOrders()
      }
    } catch {
      // 忽略轮询错误
    }
  }, 2500)
}

const startCountdown = () => {
  if (countdownTimer) clearInterval(countdownTimer)
  countdownTimer = setInterval(() => {
    countdownSec.value -= 1
    if (countdownSec.value <= 0) {
      payStatus.value = 'timeout'
      stopPolling()
    }
  }, 1000)
}

const stopPolling = () => {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
}

const closePayDialog = () => {
  stopPolling()
  showPayDialog.value = false
  currentOrder.value = null
  qrCodeURL.value = ''
  if (activeTab.value === 'orders') loadOrders()
}

const formatCountdown = (sec: number) => {
  const m = Math.floor(sec / 60).toString().padStart(2, '0')
  const s = (sec % 60).toString().padStart(2, '0')
  return `${m}:${s}`
}

const formatDateTime = (dt: string) => {
  return new Date(dt).toLocaleString()
}

const statusBadgeClass = (status: string) => {
  const map: Record<string, string> = {
    pending: 'badge-warning',
    paid: 'badge-primary',
    completed: 'badge-success',
    failed: 'badge-danger',
    expired: 'badge-gray',
    refunded: 'badge-gray'
  }
  return map[status] ?? 'badge-gray'
}

// 切换到订单页时自动加载
watch(activeTab, (tab) => {
  if (tab === 'orders') loadOrders()
})

onMounted(() => {
  loadPlans()
})

onUnmounted(() => {
  stopPolling()
})
</script>
