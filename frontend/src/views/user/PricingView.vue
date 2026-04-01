<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Page Header -->
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('pricing.title') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('pricing.description') }}</p>
      </div>

      <!-- Toolbar -->
      <div class="card px-4 py-3">
        <div class="flex flex-wrap items-center gap-3">
          <!-- Group Filter -->
          <div class="min-w-[160px]">
            <select
              v-model="selectedGroupId"
              class="input"
              @change="onGroupChange"
            >
              <option :value="undefined">{{ t('pricing.groupAll') }}</option>
              <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>

          <!-- Search -->
          <div class="relative min-w-[200px] flex-1">
            <svg class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="searchQuery"
              type="text"
              class="input pl-9"
              :placeholder="t('pricing.searchPlaceholder')"
            />
          </div>

          <!-- Unit Toggle -->
          <div class="flex items-center rounded-xl border border-gray-200 dark:border-dark-600">
            <button
              :class="[
                'rounded-l-xl px-3 py-2 text-xs font-medium transition-colors',
                unit === 'million'
                  ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                  : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
              ]"
              @click="unit = 'million'"
            >
              {{ t('pricing.unitPerMillion') }}
            </button>
            <button
              :class="[
                'rounded-r-xl px-3 py-2 text-xs font-medium transition-colors',
                unit === 'thousand'
                  ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                  : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
              ]"
              @click="unit = 'thousand'"
            >
              {{ t('pricing.unitPerThousand') }}
            </button>
          </div>

          <!-- Rate Badge -->
          <span
            v-if="hasGroup"
            class="rounded-full bg-amber-100 px-2.5 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-400"
          >
            {{ t('pricing.currentRate') }}: {{ currentRate }}x
          </span>
        </div>
      </div>

      <!-- Loading Skeleton -->
      <div v-if="loading" class="card overflow-hidden">
        <div class="px-4 py-3">
          <div class="space-y-3">
            <div v-for="n in 8" :key="n" class="flex items-center gap-4">
              <div class="h-4 w-48 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
              <div class="h-4 w-20 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
              <div class="h-4 w-20 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
              <div class="h-4 w-20 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="loadError" class="card p-12 text-center">
        <svg class="mx-auto mb-4 h-12 w-12 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <p class="text-gray-500 dark:text-gray-400">{{ t('pricing.loadError') || 'Failed to load pricing data' }}</p>
        <button @click="loadPricing(selectedGroupId)" class="btn btn-primary mt-4">{{ t('common.retry') || 'Retry' }}</button>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredModels.length === 0" class="card p-12 text-center">
        <svg class="mx-auto mb-4 h-12 w-12 text-gray-300 dark:text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" />
        </svg>
        <p class="text-gray-500 dark:text-gray-400">{{ t('pricing.noModels') }}</p>
      </div>

      <template v-else>
        <!-- Desktop Table -->
        <div class="hidden md:block">
          <div class="card overflow-hidden">
            <div class="table-wrapper overflow-x-auto">
              <table class="min-w-full">
                <thead>
                  <tr class="bg-gray-50 dark:bg-dark-800">
                    <th class="sticky top-0 bg-gray-50 px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                      {{ t('pricing.modelName') }}
                    </th>
                    <th class="sticky top-0 bg-gray-50 px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                      {{ t('pricing.input') }}
                    </th>
                    <th class="sticky top-0 bg-gray-50 px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                      {{ t('pricing.output') }}
                    </th>
                    <th class="sticky top-0 bg-gray-50 px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                      {{ t('pricing.cacheRead') }}
                    </th>
                    <th class="sticky top-0 w-10 bg-gray-50 px-4 py-3 dark:bg-dark-800">
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-dark-700/50">
                  <template v-for="(catModels, category) in groupedModels" :key="category">
                    <!-- Category Header -->
                    <tr>
                      <td colspan="5" class="bg-gray-50/80 px-4 py-2 text-xs font-semibold uppercase tracking-wider text-gray-500 dark:bg-dark-800/80 dark:text-gray-400">
                        {{ formatCategory(category as string) }}
                        <span class="ml-1 text-gray-400 dark:text-gray-500">({{ catModels.length }})</span>
                      </td>
                    </tr>
                    <!-- Model Rows -->
                    <template v-for="model in catModels" :key="model.id">
                      <tr
                        class="transition-colors hover:bg-gray-50/50 dark:hover:bg-dark-800/30"
                        :class="{ 'bg-gray-50/30 dark:bg-dark-800/20': expandedModels.has(model.id) }"
                      >
                        <!-- Model Name -->
                        <td class="px-4 py-3">
                          <div class="text-sm font-medium text-gray-900 dark:text-white">{{ model.display_name }}</div>
                          <div class="text-xs text-gray-400 dark:text-gray-500">{{ model.id }}</div>
                        </td>
                        <!-- Input Price -->
                        <td class="px-4 py-3 text-right">
                          <template v-if="hasGroup && model.effective_pricing">
                            <div class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatPrice(model.effective_pricing.input_per_million) }}</div>
                            <div class="text-xs text-gray-400 dark:text-gray-500">{{ formatPrice(model.pricing.input_per_million) }}</div>
                          </template>
                          <span v-else class="text-sm text-gray-900 dark:text-white">{{ formatPrice(model.pricing.input_per_million) }}</span>
                        </td>
                        <!-- Output Price -->
                        <td class="px-4 py-3 text-right">
                          <template v-if="hasGroup && model.effective_pricing">
                            <div class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatPrice(model.effective_pricing.output_per_million) }}</div>
                            <div class="text-xs text-gray-400 dark:text-gray-500">{{ formatPrice(model.pricing.output_per_million) }}</div>
                          </template>
                          <span v-else class="text-sm text-gray-900 dark:text-white">{{ formatPrice(model.pricing.output_per_million) }}</span>
                        </td>
                        <!-- Cache Read Price -->
                        <td class="px-4 py-3 text-right">
                          <template v-if="model.pricing.cache_read_per_million == null">
                            <span class="text-sm text-gray-300 dark:text-gray-600">—</span>
                          </template>
                          <template v-else-if="hasGroup && model.effective_pricing">
                            <div class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatPrice(model.effective_pricing.cache_read_per_million) }}</div>
                            <div class="text-xs text-gray-400 dark:text-gray-500">{{ formatPrice(model.pricing.cache_read_per_million) }}</div>
                          </template>
                          <span v-else class="text-sm text-gray-900 dark:text-white">{{ formatPrice(model.pricing.cache_read_per_million) }}</span>
                        </td>
                        <!-- Expand Arrow -->
                        <td class="px-4 py-3 text-center">
                          <button
                            v-if="hasCacheCreation(model)"
                            class="inline-flex items-center justify-center rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:text-gray-500 dark:hover:bg-dark-700 dark:hover:text-gray-300"
                            :title="expandedModels.has(model.id) ? t('pricing.collapse') : t('pricing.expand')"
                            @click="toggleExpand(model.id)"
                          >
                            <svg
                              class="h-4 w-4 transition-transform duration-200"
                              :class="{ 'rotate-180': expandedModels.has(model.id) }"
                              fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"
                            >
                              <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                            </svg>
                          </button>
                        </td>
                      </tr>
                      <!-- Expanded Detail Row -->
                      <tr v-if="expandedModels.has(model.id)" class="bg-gray-50/50 dark:bg-dark-800/50">
                        <td colspan="5" class="px-4 py-3">
                          <div class="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
                            <div class="flex items-center gap-2">
                              <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('pricing.cacheCreation') }}:</span>
                              <template v-if="model.pricing.cache_creation_per_million == null">
                                <span class="text-sm text-gray-300 dark:text-gray-600">—</span>
                              </template>
                              <template v-else-if="hasGroup && model.effective_pricing">
                                <span class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatPrice(model.effective_pricing.cache_creation_per_million) }}</span>
                                <span class="ml-1 text-xs text-gray-400 dark:text-gray-500">{{ formatPrice(model.pricing.cache_creation_per_million) }}</span>
                              </template>
                              <span v-else class="text-sm text-gray-900 dark:text-white">{{ formatPrice(model.pricing.cache_creation_per_million) }}</span>
                            </div>
                            <div class="flex items-center gap-2">
                              <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('pricing.provider') }}:</span>
                              <span class="text-sm text-gray-700 dark:text-gray-300">{{ model.owned_by }}</span>
                            </div>
                          </div>
                        </td>
                      </tr>
                    </template>
                  </template>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Mobile Cards -->
        <div class="space-y-3 md:hidden">
          <template v-for="(catModels, category) in groupedModels" :key="category">
            <!-- Category Header -->
            <div class="px-1 pt-2">
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ formatCategory(category as string) }}
                <span class="ml-1 text-gray-400 dark:text-gray-500">({{ catModels.length }})</span>
              </h3>
            </div>
            <!-- Model Cards -->
            <div
              v-for="model in catModels"
              :key="model.id"
              class="card p-4"
            >
              <div class="mb-3 flex items-start justify-between">
                <div>
                  <div class="text-sm font-medium text-gray-900 dark:text-white">{{ model.display_name }}</div>
                  <div class="text-xs text-gray-400 dark:text-gray-500">{{ model.id }}</div>
                </div>
                <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-500 dark:bg-dark-700 dark:text-gray-400">
                  {{ model.owned_by }}
                </span>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <!-- Input -->
                <div>
                  <div class="mb-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('pricing.input') }}</div>
                  <template v-if="hasGroup && model.effective_pricing">
                    <div class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatPrice(model.effective_pricing.input_per_million) }}</div>
                    <div class="text-xs text-gray-400 dark:text-gray-500">{{ formatPrice(model.pricing.input_per_million) }}</div>
                  </template>
                  <div v-else class="text-sm text-gray-900 dark:text-white">{{ formatPrice(model.pricing.input_per_million) }}</div>
                </div>
                <!-- Output -->
                <div>
                  <div class="mb-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('pricing.output') }}</div>
                  <template v-if="hasGroup && model.effective_pricing">
                    <div class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatPrice(model.effective_pricing.output_per_million) }}</div>
                    <div class="text-xs text-gray-400 dark:text-gray-500">{{ formatPrice(model.pricing.output_per_million) }}</div>
                  </template>
                  <div v-else class="text-sm text-gray-900 dark:text-white">{{ formatPrice(model.pricing.output_per_million) }}</div>
                </div>
                <!-- Cache Read -->
                <div>
                  <div class="mb-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('pricing.cacheRead') }}</div>
                  <template v-if="model.pricing.cache_read_per_million == null">
                    <span class="text-sm text-gray-300 dark:text-gray-600">—</span>
                  </template>
                  <template v-else-if="hasGroup && model.effective_pricing">
                    <div class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatPrice(model.effective_pricing.cache_read_per_million) }}</div>
                    <div class="text-xs text-gray-400 dark:text-gray-500">{{ formatPrice(model.pricing.cache_read_per_million) }}</div>
                  </template>
                  <div v-else class="text-sm text-gray-900 dark:text-white">{{ formatPrice(model.pricing.cache_read_per_million) }}</div>
                </div>
                <!-- Cache Creation -->
                <div v-if="hasCacheCreation(model)">
                  <div class="mb-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('pricing.cacheCreation') }}</div>
                  <template v-if="model.pricing.cache_creation_per_million == null">
                    <span class="text-sm text-gray-300 dark:text-gray-600">—</span>
                  </template>
                  <template v-else-if="hasGroup && model.effective_pricing">
                    <div class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatPrice(model.effective_pricing.cache_creation_per_million) }}</div>
                    <div class="text-xs text-gray-400 dark:text-gray-500">{{ formatPrice(model.pricing.cache_creation_per_million) }}</div>
                  </template>
                  <div v-else class="text-sm text-gray-900 dark:text-white">{{ formatPrice(model.pricing.cache_creation_per_million) }}</div>
                </div>
              </div>
            </div>
          </template>
        </div>

        <!-- Notice -->
        <p class="text-center text-xs text-gray-400 dark:text-gray-500">
          {{ notice || t('pricing.notice') }}
        </p>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import pricingAPI from '@/api/pricing'
import { userGroupsAPI } from '@/api/groups'
import AppLayout from '@/components/layout/AppLayout.vue'
import type { ModelPricingItem, Group } from '@/types'

const { t } = useI18n()

// --------------- State ---------------
const loading = ref(false)
const models = ref<ModelPricingItem[]>([])
const groups = ref<Group[]>([])
const selectedGroupId = ref<number | undefined>(undefined)
const searchQuery = ref('')
const unit = ref<'million' | 'thousand'>('million')
const expandedModels = ref<Set<string>>(new Set())
const currentRate = ref(1.0)
const notice = ref('')
const loadError = ref(false)

// --------------- Computed ---------------
const hasGroup = computed(() => selectedGroupId.value !== undefined)

const filteredModels = computed(() => {
  const q = searchQuery.value.toLowerCase().trim()
  if (!q) return models.value
  return models.value.filter(
    (m) =>
      m.id.toLowerCase().includes(q) ||
      m.display_name.toLowerCase().includes(q) ||
      m.owned_by.toLowerCase().includes(q),
  )
})

const groupedModels = computed(() => {
  const categoryOrder = ['anthropic', 'openai', 'google', 'other']
  const grouped: Record<string, ModelPricingItem[]> = {}

  for (const model of filteredModels.value) {
    const cat = categoryOrder.includes(model.category) ? model.category : 'other'
    if (!grouped[cat]) grouped[cat] = []
    grouped[cat].push(model)
  }

  // Return in defined order, skipping empty categories
  const ordered: Record<string, ModelPricingItem[]> = {}
  for (const cat of categoryOrder) {
    if (grouped[cat]?.length) {
      ordered[cat] = grouped[cat]
    }
  }
  return ordered
})

// --------------- Methods ---------------
async function loadGroups() {
  try {
    groups.value = await userGroupsAPI.getAvailable()
  } catch (e) {
    console.error('Failed to load groups:', e)
  }
}

async function loadPricing(groupId?: number) {
  loading.value = true
  loadError.value = false
  try {
    const res = await pricingAPI.getModelPricing(groupId)
    models.value = res.models || []
    currentRate.value = res.group?.rate_multiplier ?? 1.0
    notice.value = res.notice || ''
  } catch (e) {
    console.error('Failed to load pricing:', e)
    models.value = []
    loadError.value = true
  } finally {
    loading.value = false
  }
}

function formatPrice(value: number | null | undefined): string {
  if (value == null) return '—'
  const v = unit.value === 'thousand' ? value / 1000 : value
  if (v === 0) return '$0'
  if (Math.abs(v) < 0.01) return `$${v.toFixed(4)}`
  if (Math.abs(v) < 1) return `$${v.toFixed(3)}`
  return `$${v.toFixed(2)}`
}

function hasCacheCreation(model: ModelPricingItem): boolean {
  return model.pricing.cache_creation_per_million != null
}

function toggleExpand(modelId: string) {
  const s = new Set(expandedModels.value)
  if (s.has(modelId)) s.delete(modelId)
  else s.add(modelId)
  expandedModels.value = s
}

function onGroupChange() {
  expandedModels.value = new Set()
  loadPricing(selectedGroupId.value)
}

function formatCategory(cat: string): string {
  const map: Record<string, string> = {
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    google: 'Google',
    other: 'Other',
  }
  return map[cat] || cat.charAt(0).toUpperCase() + cat.slice(1)
}

onMounted(() => {
  loadGroups()
  loadPricing()
})
</script>
