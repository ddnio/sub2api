<!-- frontend/src/components/docs/DocsSectionFaq.vue -->
<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
      {{ t('docs.faq.title') }}
    </h1>

    <!-- NanaFox fork: 计费、报错排查、联系方式 -->
    <div class="space-y-4" data-docs-faq-billing>
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.faq.billingTitle') }}</h2>
      <ul class="space-y-2 text-sm text-gray-600 dark:text-dark-400">
        <li v-for="(line, idx) in billingLines" :key="idx" class="flex gap-2">
          <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-primary-500"></span>
          <span>{{ line }}</span>
        </li>
      </ul>
    </div>

    <div class="space-y-4" data-docs-faq-errors>
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.faq.errorsTitle') }}</h2>
      <div class="overflow-x-auto rounded-xl border border-gray-200/80 dark:border-dark-700">
        <table class="w-full text-left text-sm">
          <tbody class="divide-y divide-gray-200/80 dark:divide-dark-700">
            <tr v-for="row in errorRows" :key="row.code" class="text-gray-700 dark:text-dark-300">
              <td class="whitespace-nowrap px-4 py-3 font-mono font-semibold text-gray-900 dark:text-white">{{ row.code }}</td>
              <td class="px-4 py-3">{{ row.fix }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- NanaFox fork: Codex 常见报错排查（B5） -->
    <section id="codex-errors" class="space-y-4" data-docs-faq-codex-errors>
      <div>
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ codexFaqCopy.title }}</h2>
        <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ codexFaqCopy.intro }}</p>
      </div>

      <div class="space-y-2 rounded-xl border border-gray-200/80 bg-gray-50/70 p-4 dark:border-dark-700 dark:bg-dark-900/60">
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ codexFaqCopy.configTitle }}</h3>
        <p class="text-xs text-gray-600 dark:text-dark-400">{{ codexFaqCopy.configDesc }}</p>
        <DocsCodeBlock
          :tabs="[{ id: 'codex-faq-config', label: '~/.codex/config.toml', path: '~/.codex/config.toml', content: standardCodexConfigToml }]"
        />
        <p class="text-xs text-gray-500 dark:text-dark-400">{{ codexFaqCopy.ccsTip }}</p>
      </div>

      <div class="space-y-3">
        <article
          v-for="item in codexErrorItems"
          :key="item.id"
          class="rounded-xl border border-gray-200/80 p-4 dark:border-dark-700"
        >
          <h3 class="font-semibold text-gray-900 dark:text-white">{{ item.title }}</h3>
          <div class="mt-2 space-y-2 text-sm text-gray-600 dark:text-dark-400">
            <div>
              <p class="font-medium text-gray-800 dark:text-gray-200">{{ codexFaqCopy.symptomLabel }}</p>
              <ul class="mt-1 space-y-1">
                <li v-for="(s, idx) in item.symptoms" :key="idx" class="font-mono text-xs break-all text-gray-700 dark:text-gray-300">
                  {{ s }}
                </li>
              </ul>
            </div>
            <p><span class="font-medium text-gray-800 dark:text-gray-200">{{ codexFaqCopy.causeLabel }}</span>{{ item.cause }}</p>
            <div>
              <p class="font-medium text-gray-800 dark:text-gray-200">{{ codexFaqCopy.fixLabel }}</p>
              <ul class="mt-1 space-y-1">
                <li v-for="(step, idx) in item.fixes" :key="idx" class="flex gap-2">
                  <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-primary-500"></span>
                  <span>{{ step }}</span>
                </li>
              </ul>
            </div>
          </div>
        </article>
      </div>
    </section>

    <div class="space-y-4">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.faq.generalTitle') }}</h2>
      <div class="space-y-4">
        <article
          v-for="(item, idx) in faqItems"
          :key="idx"
          class="rounded-xl border border-gray-200/80 p-4 dark:border-dark-700"
        >
          <h3 class="font-semibold text-gray-900 dark:text-white">{{ item.q }}</h3>
          <p class="mt-2 text-sm text-gray-600 dark:text-dark-400">{{ item.a }}</p>
        </article>
      </div>
    </div>

    <div class="space-y-2" data-docs-faq-contact>
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.faq.contactTitle') }}</h2>
      <p class="text-sm text-gray-600 dark:text-dark-400">{{ t('docs.faq.contactDesc', { contact, email: 'dev@nanafox.com' }) }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import DocsCodeBlock from '@/components/docs/DocsCodeBlock.vue'
import { DEFAULT_CODEX_MODEL, buildOpenAICodexConfigToml, ensureOpenAIV1BaseUrl } from '@/utils/codexAuthConfig'

const { t, locale } = useI18n()
const appStore = useAppStore()

const isZh = computed(() => String(locale.value || '').toLowerCase().startsWith('zh'))
const apiBaseRoot = computed(() =>
  (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, '')
)
const codexApiBase = computed(() => ensureOpenAIV1BaseUrl(apiBaseRoot.value))
const standardCodexConfigToml = computed(() =>
  buildOpenAICodexConfigToml({
    baseUrl: apiBaseRoot.value,
    apiKey: 'your-api-key-here'
  })
)

const contact = computed(() => appStore.contactInfo?.trim() || t('docs.faq.contactFallback'))

const codexFaqCopy = computed(() => {
  if (isZh.value) {
    return {
      title: 'Codex 常见报错',
      intro: '排查前先核对下方标准单文件 config.toml（与控制台「使用密钥」生成的默认配置完全一致）。',
      configTitle: '标准 Codex 配置速查',
      configDesc: 'macOS / Linux 位于 ~/.codex/config.toml，Windows 位于 %userprofile%\\.codex\\config.toml（若设置了 CODEX_HOME 则位于该目录）。',
      ccsTip: '已安装 CC Switch 的用户，建议直接在控制台「API 密钥」里点「导入到 CCS」，免手改文件。',
      symptomLabel: '现象：',
      causeLabel: '原因：',
      fixLabel: '怎么改：'
    }
  }
  return {
    title: 'Common Codex Errors',
    intro: 'Before troubleshooting, verify your config against the standard single-file config.toml below (identical to the dashboard "Use Key" default).',
    configTitle: 'Standard Codex Config Quick Check',
    configDesc: 'Located at ~/.codex/config.toml on macOS/Linux or %userprofile%\\.codex\\config.toml on Windows (or inside CODEX_HOME if set).',
    ccsTip: 'If you use CC Switch, click "Import to CCS" in the dashboard API Keys page to skip manual editing.',
    symptomLabel: 'Symptom:',
    causeLabel: 'Cause: ',
    fixLabel: 'How to fix:'
  }
})

interface CodexErrorItem {
  id: string
  title: string
  symptoms: string[]
  cause: string
  fixes: string[]
}

const codexErrorItems = computed<CodexErrorItem[]>(() => {
  if (isZh.value) {
    return [
      {
        id: 'model-unsupported',
        title: '1. 模型不支持或已下线（400 / 404）',
        symptoms: [
          "400 invalid_request_error: The 'gpt-5.4' model is not supported when using Codex with a ChatGPT account.",
          '404 model_not_found: Model "gpt-5.5" targets GPT-5.5, which retired from Codex... Select an available replacement such as GPT-6.1 Sol or GPT-6.2 Sol',
          '404 model_not_found: Model "xxx" is not supported by any configured account in this group'
        ],
        cause: 'config.toml 里填了已下线或通道不支持的模型名，或用了「OpenAI GPT」分组的 Key 去调 Claude / 其他平台模型。',
        fixes: [
          `把 config.toml 中的 model 改为当前可用模型（例如 ${DEFAULT_CODEX_MODEL} 或 gpt-6.2-sol），保存后重启 Codex；`,
          '如需调用 Claude 模型，请到控制台「API 密钥」新建一个「ClaudeCode专用」分组的 Key（OpenAI GPT 分组仅支持 GPT/Codex 模型）。'
        ]
      },
      {
        id: 'insufficient-balance',
        title: '2. 余额不足（403）',
        symptoms: [
          '403 billing_error: insufficient balance',
          '403 api_error: Insufficient account balance'
        ],
        cause: '账户体验额度或按量余额已用完。',
        fixes: [
          '前往控制台「充值 / 购买」页面补充按量余额或开通 GPT 月卡；',
          '若已开通月卡仍报此错，请到控制台「API 密钥」确认当前 config.toml 里填的是否为对应月卡分组的 Key。'
        ]
      },
      {
        id: 'quota-limit',
        title: '3. 套餐日限额或周限额已满（429 / 403）',
        symptoms: [
          '429 api_error: DAILY_LIMIT_EXCEEDED / WEEKLY_LIMIT_EXCEEDED',
          '403 billing_error: daily usage limit exceeded / weekly usage limit exceeded'
        ],
        cause: '当前订阅套餐的每日或每周额度已用完。',
        fixes: [
          '等待周期自动重置；',
          '若需继续使用，可在「充值 / 购买」页升级更高档位套餐，或在控制台切回按量分组的 Key。'
        ]
      },
      {
        id: 'no-active-subscription',
        title: '4. Key 所在分组没有有效订阅（403）',
        symptoms: [
          '403 api_error: No active subscription found for this group'
        ],
        cause: '当前使用的 API Key 绑定了订阅制分组，但账户在该分组下没有有效订阅（未开通或已过期）。',
        fixes: [
          '到控制台「API 密钥」确认该 Key 所属分组；',
          '按量用户请新建或换用「OpenAI GPT」分组的 Key 并更新 config.toml；月卡用户请续费对应分组。'
        ]
      },
      {
        id: 'concurrency-limit',
        title: '5. 并发请求数超限（502）',
        symptoms: [
          '502 upstream_error: Concurrency limit exceeded for user, please retry later',
          '502 upstream_error: Concurrency limit exceeded for account, please retry later'
        ],
        cause: '同一时间并发进行的请求数超过当前账户或通道上限（例如同时开多个 Codex 窗口、子代理并行或脚本并发调用）。',
        fixes: [
          '减少同时运行的 Codex 窗口或并发任务数量，等待几秒后重试。'
        ]
      },
      {
        id: 'responses-continuation',
        title: '6. Responses 接口续传协议不兼容（400）',
        symptoms: [
          '400 invalid_request_error: function_call_output requires call_id on HTTP requests; continuation via previous_response_id is only supported on Responses WebSocket v2'
        ],
        cause: '客户端在 HTTP 模式下发送了依赖 WebSocket 续传（previous_response_id）的工具调用上下文，常见于旧版 Codex 客户端或中间代理改写了请求。',
        fixes: [
          '将 Codex CLI / Desktop 升级到最新版本；',
          `核对 config.toml 中 wire_api = "responses" 且 base_url = "${codexApiBase.value}"（不要经过会转写为 chat/completions 的旧代理）；`,
          '开启一个新会话重试。'
        ]
      },
      {
        id: 'context-window',
        title: '7. 输入超出模型上下文窗口（400）',
        symptoms: [
          '400 invalid_request_error: Your input exceeds the context window of this model. Please adjust your input and try again.'
        ],
        cause: '当前会话历史或一次性附带的代码文件过长，超出模型最大上下文长度。',
        fixes: [
          '在 Codex 中执行 /compact 压缩历史对话，或新开会话并仅附带当前修改相关的文件。'
        ]
      },
      {
        id: 'upstream-transient',
        title: '8. 流式响应中断、上游限流或请求超时（502 / 503 / 429 / 408）',
        symptoms: [
          '502: Upstream response stream was interrupted / Upstream HTTP/2 stream failed / Upstream service temporarily unavailable',
          '503 service_unavailable_error: Our servers are currently overloaded. Please try again later.',
          '429 rate_limit_error: All available accounts are currently rate-limited. Please retry later.',
          '408 invalid_request_error: Timed out reading request body'
        ],
        cause: '上游模型接口短暂抖动或高峰期限流；408 多为本地网络上传超大请求体超时。',
        fixes: [
          '遇到 502 / 503 / 429 rate_limit_error，等待 5–30 秒重试通常即可恢复；',
          '遇到 408，检查本地网络稳定性并避免单次粘贴超大日志或文件；',
          '若多次重试仍持续报错，请前往控制台查看公告或通过页面底部联系方式联系我们。'
        ]
      }
    ]
  }

  return [
    {
      id: 'model-unsupported',
      title: '1. Unsupported or retired model (400 / 404)',
      symptoms: [
        "400 invalid_request_error: The 'gpt-5.4' model is not supported when using Codex with a ChatGPT account.",
        '404 model_not_found: Model "gpt-5.5" retired from Codex...',
        '404 model_not_found: Model "xxx" is not supported by any configured account in this group'
      ],
      cause: 'config.toml references a retired/unsupported model, or an OpenAI GPT key is being used for Claude or another platform.',
      fixes: [
        `Update model in config.toml to a supported model such as ${DEFAULT_CODEX_MODEL} or gpt-6.2-sol and restart Codex.`,
        'For Claude models, create a separate key in the "ClaudeCode专用" group.'
      ]
    },
    {
      id: 'insufficient-balance',
      title: '2. Insufficient balance (403)',
      symptoms: [
        '403 billing_error: insufficient balance',
        '403 api_error: Insufficient account balance'
      ],
      cause: 'Trial credit or pay-as-you-go balance has been used up.',
      fixes: [
        'Top up your balance or subscribe to a GPT monthly plan on the Purchase page.',
        'If you already have a monthly plan, check that config.toml uses the API key from that plan group.'
      ]
    },
    {
      id: 'quota-limit',
      title: '3. Daily or weekly plan limit reached (429 / 403)',
      symptoms: [
        '429 api_error: DAILY_LIMIT_EXCEEDED / WEEKLY_LIMIT_EXCEEDED',
        '403 billing_error: daily usage limit exceeded / weekly usage limit exceeded'
      ],
      cause: 'Your subscription daily or weekly quota has been reached.',
      fixes: [
        'Wait for the quota window to reset automatically.',
        'Or upgrade your plan on the Purchase page / switch to a pay-as-you-go group key.'
      ]
    },
    {
      id: 'no-active-subscription',
      title: '4. No active subscription for this group (403)',
      symptoms: ['403 api_error: No active subscription found for this group'],
      cause: 'The API key belongs to a subscription group without an active subscription.',
      fixes: [
        'Check the key group on the dashboard API Keys page.',
        'Use an "OpenAI GPT" pay-as-you-go key in config.toml, or renew the subscription for that group.'
      ]
    },
    {
      id: 'concurrency-limit',
      title: '5. Concurrency limit exceeded (502)',
      symptoms: [
        '502 upstream_error: Concurrency limit exceeded for user, please retry later',
        '502 upstream_error: Concurrency limit exceeded for account, please retry later'
      ],
      cause: 'Too many concurrent requests are running at the same time (multiple Codex windows, parallel subagents, or scripts).',
      fixes: ['Reduce concurrent Codex windows/tasks and retry after a few seconds.']
    },
    {
      id: 'responses-continuation',
      title: '6. Responses continuation protocol mismatch (400)',
      symptoms: [
        '400 invalid_request_error: function_call_output requires call_id on HTTP requests; continuation via previous_response_id is only supported on Responses WebSocket v2'
      ],
      cause: 'An older Codex client or intermediate proxy sent WebSocket continuation fields over HTTP.',
      fixes: [
        'Upgrade Codex CLI / Desktop to the latest version.',
        `Ensure config.toml has wire_api = "responses" and base_url = "${codexApiBase.value}".`,
        'Start a fresh session and retry.'
      ]
    },
    {
      id: 'context-window',
      title: '7. Input exceeds context window (400)',
      symptoms: ['400 invalid_request_error: Your input exceeds the context window of this model.'],
      cause: 'Conversation history or attached files exceed the model context length.',
      fixes: ['Run /compact in Codex or start a new session with only the relevant files.']
    },
    {
      id: 'upstream-transient',
      title: '8. Stream interrupted, rate limited, or request timeout (502 / 503 / 429 / 408)',
      symptoms: [
        '502: Upstream response stream was interrupted / Upstream HTTP/2 stream failed',
        '503 service_unavailable_error: Our servers are currently overloaded.',
        '429 rate_limit_error: All available accounts are currently rate-limited.',
        '408 invalid_request_error: Timed out reading request body'
      ],
      cause: 'Brief upstream congestion or peak-hour rate limits; 408 is usually a slow upload of a very large payload.',
      fixes: [
        'Wait 5–30 seconds and retry for 502 / 503 / 429 errors.',
        'For 408, check local network stability and avoid pasting oversized logs or files.',
        'If errors persist, check dashboard announcements or reach out via the contact details below.'
      ]
    }
  ]
})

const billingLines = computed(() => [
  t('docs.faq.billing.payg'),
  t('docs.faq.billing.topup'),
  t('docs.faq.billing.monthly')
])

const errorRows = computed(() => [
  { code: '404', fix: t('docs.faq.errors.notFound') },
  { code: '401', fix: t('docs.faq.errors.unauthorized') },
  { code: '403', fix: t('docs.faq.errors.forbidden') },
  { code: 'model not found', fix: t('docs.faq.errors.model') }
])

const faqItems = computed(() => [
  { q: t('docs.faq.items.compatibilityQ'), a: t('docs.faq.items.compatibilityA') },
  { q: t('docs.faq.items.multiToolQ'), a: t('docs.faq.items.multiToolA') },
  { q: t('docs.faq.items.modelsQ'), a: t('docs.faq.items.modelsA') },
  { q: t('docs.api.faq.items.sdkQ'), a: t('docs.api.faq.items.sdkA') },
  { q: t('docs.api.faq.items.dataQ'), a: t('docs.api.faq.items.dataA') },
])
</script>
