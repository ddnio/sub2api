<!-- NanaFox fork: 注册后“开始使用”引导页。只服务中文用户，文案直接写在页面里，不改上游 i18n。 -->
<template>
  <AppLayout>
    <div class="mx-auto max-w-3xl space-y-5" data-testid="start-view">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-gray-900 dark:text-white">4 步开始用 NanaFox Router</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            账户里已有 ${{ (user?.balance ?? 0).toFixed(2) }} 额度。建 key → 复制配置 → 发一句话试试。
          </p>
        </div>
        <button
          type="button"
          class="shrink-0 text-sm text-gray-500 underline-offset-4 transition-colors hover:text-gray-700 hover:underline dark:text-dark-400 dark:hover:text-gray-200"
          data-testid="start-skip"
          @click="skip"
        >
          跳过，去控制台
        </button>
      </div>

      <div v-if="loading" class="card p-6 text-sm text-gray-500 dark:text-dark-400">加载中…</div>

      <template v-else>
        <!-- 第 1 步：用途 -->
        <section class="card p-5">
          <h2 class="mb-3 text-base font-semibold text-gray-900 dark:text-white">1. 你打算用在哪？</h2>
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <button
              v-for="opt in useCaseOptions"
              :key="opt.value"
              type="button"
              :data-testid="`start-usecase-${opt.value}`"
              :disabled="!!createdKey || creating"
              class="rounded-xl border p-4 text-left transition disabled:cursor-not-allowed"
              :class="
                useCase === opt.value
                  ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                  : 'border-gray-200 hover:border-primary-300 dark:border-dark-700'
              "
              @click="useCase = opt.value"
            >
              <div class="font-medium text-gray-900 dark:text-white">{{ opt.title }}</div>
              <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ opt.desc }}</div>
            </button>
          </div>
          <p v-if="useCase === 'client'" class="mt-3 text-xs text-gray-500 dark:text-dark-400">
            客户端里想用 Claude 的话，以后到「API 密钥」再建一个「ClaudeCode专用」分组的 key，客户端选 Anthropic 类型。
          </p>
        </section>

        <!-- 第 2 步：建 key -->
        <section class="card p-5">
          <h2 class="mb-3 text-base font-semibold text-gray-900 dark:text-white">2. 一键建 key</h2>
          <div v-if="!createdKey">
            <p class="mb-3 text-sm text-gray-600 dark:text-dark-300">
              将使用分组：<b>{{ selectedGroup?.name || '（暂无可用分组）' }}</b>
            </p>
            <button
              type="button"
              class="btn btn-primary"
              data-testid="start-create"
              :disabled="!selectedGroup || creating"
              @click="createKey"
            >
              {{ creating ? '正在创建…' : '帮我建好 key' }}
            </button>
            <p v-if="!selectedGroup" class="mt-3 text-sm text-red-600">
              没找到可用的分组，请加微信 nanafox-hi，我们帮你开通。
            </p>
          </div>
          <div v-else class="space-y-2">
            <p class="text-sm text-gray-600 dark:text-dark-300">已建好「{{ createdKey.name }}」，分组 {{ selectedGroup?.name }}：</p>
            <div class="flex items-center gap-2">
              <code
                class="flex-1 break-all rounded-lg bg-gray-100 px-3 py-2 font-mono text-sm text-gray-900 dark:bg-dark-800 dark:text-gray-100"
                data-testid="start-key"
              >{{ createdKey.key }}</code>
              <button type="button" class="btn btn-secondary" @click="copy(createdKey.key)">复制</button>
            </div>
            <p class="text-xs text-gray-500 dark:text-dark-400">以后也能在左侧「API 密钥」里找到它。</p>
          </div>
        </section>

        <!-- 第 3 步：复制配置（Codex 直接内联 CC Switch + 一行写入命令；其他用途复用「使用密钥」弹窗） -->
        <section id="start-step-config" class="card p-5" :class="{ 'opacity-50': !createdKey }">
          <h2 class="mb-3 text-base font-semibold text-gray-900 dark:text-white">3. 复制配置</h2>

          <template v-if="useCase === 'codex'">
            <div class="space-y-4" data-testid="start-codex-setup">
              <div class="flex flex-wrap items-center gap-3">
                <button
                  v-if="!publicSettings?.hide_ccs_import_button"
                  type="button"
                  class="btn btn-primary"
                  data-testid="start-ccs"
                  :disabled="!createdKey"
                  @click="importToCcswitch"
                >
                  导入到 CC Switch
                </button>
                <button
                  type="button"
                  class="btn btn-secondary"
                  data-testid="start-use-key"
                  :disabled="!createdKey"
                  @click="showUseKeyModal = true"
                >
                  查看完整配置
                </button>
                <span
                  v-if="!publicSettings?.hide_ccs_import_button"
                  class="text-xs text-gray-500 dark:text-dark-400"
                >
                  已装 CC Switch 点一下即可自动写入配置
                </span>
              </div>

              <div class="space-y-2">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <p class="text-sm text-gray-600 dark:text-dark-300">
                    或复制下面命令到终端回车，自动写入 <code>~/.codex/config.toml</code>（与「使用密钥」一致）：
                  </p>
                  <div class="inline-flex rounded-lg border border-gray-200 bg-gray-50 p-0.5 text-xs dark:border-dark-700 dark:bg-dark-900">
                    <button
                      type="button"
                      data-testid="start-codex-shell-unix"
                      class="rounded-md px-2.5 py-1 font-medium transition-colors"
                      :class="
                        codexShell === 'unix'
                          ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-800 dark:text-white'
                          : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'
                      "
                      @click="codexShell = 'unix'"
                    >
                      macOS / Linux
                    </button>
                    <button
                      type="button"
                      data-testid="start-codex-shell-powershell"
                      class="rounded-md px-2.5 py-1 font-medium transition-colors"
                      :class="
                        codexShell === 'powershell'
                          ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-800 dark:text-white'
                          : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'
                      "
                      @click="codexShell = 'powershell'"
                    >
                      PowerShell
                    </button>
                  </div>
                </div>
                <div class="overflow-hidden rounded-xl bg-gray-900 dark:bg-dark-900">
                  <div class="flex items-center justify-between border-b border-gray-700 bg-gray-800 px-4 py-2 dark:border-dark-700 dark:bg-dark-800">
                    <span class="font-mono text-xs text-gray-400">
                      {{ codexShell === 'powershell' ? 'PowerShell' : 'macOS / Linux Terminal' }}
                    </span>
                    <button
                      type="button"
                      class="rounded-lg bg-gray-700 px-2.5 py-1 text-xs font-medium text-gray-200 transition-colors hover:bg-gray-600 hover:text-white disabled:opacity-50"
                      data-testid="start-codex-copy"
                      :disabled="!createdKey"
                      @click="copy(codexQuickCommand)"
                    >
                      复制命令
                    </button>
                  </div>
                  <pre class="overflow-x-auto p-4 font-mono text-xs text-gray-100"><code data-testid="start-codex-command">{{ codexQuickCommand }}</code></pre>
                </div>
              </div>
            </div>
          </template>

          <template v-else>
            <p class="mb-2 text-sm text-gray-600 dark:text-dark-300">
              接口地址（{{ platform === 'anthropic' ? 'Claude Code 不加 /v1' : 'OpenAI 兼容写法要加 /v1' }}）：
            </p>
            <div class="mb-4 flex items-center gap-2">
              <code
                class="flex-1 break-all rounded-lg bg-gray-100 px-3 py-2 font-mono text-sm text-gray-900 dark:bg-dark-800 dark:text-gray-100"
                data-testid="start-base-url"
              >{{ baseUrl }}</code>
              <button type="button" class="btn btn-secondary" @click="copy(baseUrl)">复制</button>
            </div>
            <div class="flex flex-wrap gap-3">
              <button
                type="button"
                class="btn btn-primary"
                data-testid="start-use-key"
                :disabled="!createdKey"
                @click="showUseKeyModal = true"
              >
                使用密钥（各客户端填法，key 已填好）
              </button>
              <button
                v-if="!publicSettings?.hide_ccs_import_button"
                type="button"
                class="btn btn-secondary"
                data-testid="start-ccs"
                :disabled="!createdKey"
                @click="importToCcswitch"
              >
                导入到 CCS
              </button>
            </div>
          </template>
        </section>

        <!-- 第 4 步：测一下 -->
        <section class="card p-5" :class="{ 'opacity-50': !createdKey }">
          <h2 class="mb-1 text-base font-semibold text-gray-900 dark:text-white">4. 发一句话测试</h2>
          <!-- 仅限 Claude Code 客户端的分组，浏览器直连一定被拒，改成去终端里测 -->
          <div v-if="selectedGroup?.claude_code_only" data-testid="start-test-cli-only">
            <p class="mb-3 text-sm text-gray-600 dark:text-dark-300">
              这个分组只接受 Claude Code 客户端，网页里测不了。按上一步配好后，在终端运行 <code>claude</code> 发一句话就行。
            </p>
            <button type="button" class="btn btn-secondary" :disabled="!createdKey" @click="finish">完成，去控制台</button>
          </div>
          <template v-else>
          <p class="mb-3 text-xs text-gray-500 dark:text-dark-400">
            用刚建的 key 让 {{ testModel }} 回一句话，只花几分钱，从你的额度里扣。
          </p>
          <button
            type="button"
            class="btn btn-primary"
            data-testid="start-test"
            :disabled="!createdKey || testing"
            @click="runTest"
          >
            {{ testing ? '测试中…' : '发一句话测试' }}
          </button>
          <div v-if="testResult" class="mt-3 rounded-lg p-3 text-sm" :class="testResult.ok ? 'bg-green-50 text-green-800 dark:bg-green-900/20 dark:text-green-300' : 'bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-300'" data-testid="start-test-result">
            <template v-if="testResult.ok">
              ✅ 接通了，用时 {{ testResult.seconds }} 秒。模型回复：{{ testResult.text || '（空）' }}
            </template>
            <template v-else>
              没接通：{{ testResult.text }}<br />
              卡住了？加微信 nanafox-hi，我们帮你看。
            </template>
          </div>
          <div v-if="testResult?.ok" class="mt-4 space-y-2">
            <p v-if="useCase === 'codex'" class="text-sm text-gray-600 dark:text-dark-300">
              现在回终端运行 <code>codex</code>，发一句“你好”即可开始写代码。
            </p>
            <button type="button" class="btn btn-secondary" @click="finish">完成，去控制台</button>
          </div>
          </template>
        </section>
      </template>
    </div>

    <UseKeyModal
      :show="showUseKeyModal"
      :api-key="createdKey?.key || ''"
      :base-url="publicSettings?.api_base_url || ''"
      :platform="selectedGroup?.platform || null"
      :claude-code-only="selectedGroup?.claude_code_only || false"
      :allow-messages-dispatch="selectedGroup?.allow_messages_dispatch || false"
      @close="showUseKeyModal = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { keysAPI, userGroupsAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import UseKeyModal from '@/components/keys/UseKeyModal.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore, useAuthStore } from '@/stores'
import type { ApiKey, Group, PublicSettings } from '@/types'
import { CC_SWITCH_USAGE_SCRIPT, buildCcSwitchImportDeeplink } from '@/utils/ccswitchImport'
import { buildCodexQuickSetupCommand, type CodexQuickSetupShell } from '@/utils/codexAuthConfig'
import {
  DEFAULT_START_USE_CASE,
  baseUrlFor,
  buildTestRequest,
  extractError,
  extractReply,
  markStartDone,
  pickGroupForUseCase,
  platformForUseCase,
  rememberedUseCase,
  useCaseFromQuery,
  type StartUseCase
} from './startOnboarding'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const user = computed(() => authStore.user)
const loading = ref(true)
const groups = ref<Group[]>([])
const publicSettings = ref<PublicSettings | null>(null)
const useCase = ref<StartUseCase>(useCaseFromQuery(route.query.use) ?? rememberedUseCase() ?? DEFAULT_START_USE_CASE)
const createdKey = ref<ApiKey | null>(null)
const creating = ref(false)
const showUseKeyModal = ref(false)
const codexShell = ref<CodexQuickSetupShell>('unix')
const testing = ref(false)
const testResult = ref<{ ok: boolean; text: string; seconds?: string } | null>(null)

const useCaseOptions: { value: StartUseCase; title: string; desc: string }[] = [
  { value: 'codex', title: 'Codex', desc: 'OpenAI 的编程工具，用 GPT 模型' },
  { value: 'claude', title: 'Claude Code', desc: '命令行写代码，用 Claude 模型' },
  { value: 'client', title: 'Cherry Studio / Cursor 等客户端', desc: 'OpenAI 兼容接口，用 GPT 模型' },
  { value: 'code', title: '自己写代码', desc: 'OpenAI SDK 等，用 GPT 模型' }
]

const platform = computed(() => platformForUseCase(useCase.value))
const selectedGroup = computed(() => pickGroupForUseCase(groups.value, useCase.value))
const origin = computed(() => publicSettings.value?.api_base_url || window.location.origin)
const baseUrl = computed(() => baseUrlFor(origin.value, platform.value))
const codexQuickCommand = computed(() =>
  buildCodexQuickSetupCommand(origin.value, createdKey.value?.key || 'sk-your-api-key', codexShell.value)
)
const testModel = computed(() => buildTestRequest(origin.value, platform.value, '').model)

function goDashboard() {
  markStartDone(user.value?.id)
  router.replace('/dashboard')
}

function skip() {
  goDashboard()
}

function finish() {
  goDashboard()
}

async function copy(text: string) {
  await copyToClipboard(text, '已复制')
}

async function createKey() {
  const group = selectedGroup.value
  if (!group || creating.value) return
  creating.value = true
  try {
    createdKey.value = await keysAPI.create('我的第一个 key', group.id)
    markStartDone(user.value?.id)
    if (useCase.value !== 'codex') {
      showUseKeyModal.value = true
    }
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || error?.message || '创建失败，请稍后重试')
  } finally {
    creating.value = false
  }
}

function importToCcswitch() {
  const key = createdKey.value
  if (!key) return
  const providerName = (publicSettings.value?.site_name || 'sub2api').trim() || 'sub2api'
  const deeplink = buildCcSwitchImportDeeplink({
    baseUrl: origin.value,
    platform: selectedGroup.value?.platform || 'anthropic',
    clientType: 'claude',
    providerName,
    apiKey: key.key,
    usageScript: CC_SWITCH_USAGE_SCRIPT
  })
  window.open(deeplink, '_self')
}

async function runTest() {
  const key = createdKey.value
  if (!key || testing.value) return
  testing.value = true
  testResult.value = null
  const req = buildTestRequest(origin.value, platform.value, key.key)
  const started = performance.now()
  try {
    const resp = await fetch(req.url, req.init)
    const data = await resp.json().catch(() => null)
    const seconds = ((performance.now() - started) / 1000).toFixed(1)
    if (resp.ok) {
      testResult.value = { ok: true, text: extractReply(data), seconds }
    } else {
      testResult.value = { ok: false, text: `HTTP ${resp.status} ${extractError(data)}`.trim() }
    }
  } catch (error: any) {
    testResult.value = { ok: false, text: error?.message || '网络错误' }
  } finally {
    testing.value = false
  }
}

onMounted(async () => {
  try {
    const resumeConfig = route.query.step === 'config'
    const existing = await keysAPI.list(1, resumeConfig ? 20 : 1)
    if ((existing?.total ?? 0) > 0 && !resumeConfig) {
      goDashboard()
      return
    }
    const [g, s] = await Promise.all([
      userGroupsAPI.getAvailable(),
      appStore.fetchPublicSettings().catch(() => null)
    ])
    groups.value = g || []
    publicSettings.value = s
    if (resumeConfig && existing?.items?.length) {
      const preferred =
        existing.items.find((k) => k.group?.platform === 'openai') ?? existing.items[0]
      createdKey.value = preferred
      if (!useCaseFromQuery(route.query.use) && preferred.group?.platform === 'anthropic') {
        useCase.value = 'claude'
      }
    }
  } catch (error: any) {
    appStore.showError(error?.message || '加载失败，可以先跳过，到「API 密钥」里手动创建')
  } finally {
    loading.value = false
  }
})
</script>
