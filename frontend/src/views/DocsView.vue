<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <HomeHeader :is-dark="isDark" @toggle-theme="toggleTheme" />

    <main class="px-6 py-6 sm:py-8">
      <div class="mx-auto flex max-w-4xl flex-col gap-4 sm:gap-6">
        <section class="rounded-2xl border border-gray-200/80 bg-white/90 p-5 shadow-sm dark:border-dark-800 dark:bg-dark-900 sm:p-6">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary-600 dark:text-primary-400">
                {{ guideCopy.kicker }}
              </p>
              <h1 class="mt-3 text-3xl font-bold text-gray-900 dark:text-white sm:text-4xl">
                {{ guideCopy.title }}
              </h1>
              <p class="mt-4 max-w-3xl text-sm leading-7 text-gray-600 dark:text-dark-400">
                {{ guideCopy.subtitle }}
              </p>
            </div>

            <div class="mt-1 shrink-0 inline-flex rounded-full border border-gray-200 bg-gray-50 p-1 dark:border-dark-700 dark:bg-dark-950">
              <button
                v-for="platform in platforms"
                :key="platform.id"
                type="button"
                class="rounded-full px-4 py-2 text-sm font-medium transition-colors"
                :class="codexPlatform === platform.id
                  ? 'bg-gray-900 text-white dark:bg-white dark:text-dark-950'
                  : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'"
                @click="codexPlatform = platform.id"
              >
                {{ platform.label }}
              </button>
            </div>
          </div>
        </section>

        <section class="rounded-2xl border border-gray-200/80 bg-white/90 p-5 shadow-sm dark:border-dark-800 dark:bg-dark-900 sm:p-6">
          <div class="flex items-start gap-4">
            <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-semibold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">1</span>
            <div class="min-w-0 flex-1">
              <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ guideCopy.nodeTitle }}</h2>
              <p class="mt-2 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ guideCopy.nodeDescription }}</p>

              <div class="mt-6 space-y-6">
                <div v-for="method in nodeMethods" :key="method.title" class="space-y-3">
                  <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ method.title }}</h3>

                  <div v-if="method.code" class="overflow-hidden rounded-xl bg-gray-900">
                    <div class="flex items-center justify-between border-b border-gray-700/60 px-4 py-3">
                      <span class="text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">{{ method.shell }}</span>
                      <button type="button" class="text-xs text-gray-400 transition-colors hover:text-white" @click="copyText(method.code, method.title)">
                        {{ copiedKey === method.title ? t('docs.copied') : t('docs.copy') }}
                      </button>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm leading-7 text-gray-200"><code>{{ method.code }}</code></pre>
                  </div>

                  <ul v-if="method.bullets?.length" class="space-y-2 text-sm leading-7 text-gray-600 dark:text-dark-400">
                    <li v-for="bullet in method.bullets" :key="bullet" class="flex gap-2">
                      <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500"></span>
                      <span>{{ bullet }}</span>
                    </li>
                  </ul>

                  <p v-if="method.note" class="text-sm leading-7 text-gray-600 dark:text-dark-400">
                    {{ method.note }}
                  </p>
                </div>

                <div class="space-y-3">
                  <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ guideCopy.nodeVerifyTitle }}</h3>
                  <div class="overflow-hidden rounded-xl bg-gray-900">
                    <div class="flex items-center justify-between border-b border-gray-700/60 px-4 py-3">
                      <span class="text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">{{ currentShellLabel }}</span>
                      <button type="button" class="text-xs text-gray-400 transition-colors hover:text-white" @click="copyText(nodeVerifyCommand, 'node-verify')">
                        {{ copiedKey === 'node-verify' ? t('docs.copied') : t('docs.copy') }}
                      </button>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm leading-7 text-gray-200"><code>{{ nodeVerifyCommand }}</code></pre>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-2xl border border-gray-200/80 bg-white/90 p-5 shadow-sm dark:border-dark-800 dark:bg-dark-900 sm:p-6">
          <div class="flex items-start gap-4">
            <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-semibold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">2</span>
            <div class="min-w-0 flex-1">
              <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ guideCopy.installTitle }}</h2>
              <p class="mt-2 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ installLead }}</p>

              <div class="mt-6 overflow-hidden rounded-xl bg-gray-900">
                <div class="flex items-center justify-between border-b border-gray-700/60 px-4 py-3">
                  <span class="text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">npm</span>
                  <button type="button" class="text-xs text-gray-400 transition-colors hover:text-white" @click="copyText(codexInstallCommand, 'codex-install')">
                    {{ copiedKey === 'codex-install' ? t('docs.copied') : t('docs.copy') }}
                  </button>
                </div>
                <pre class="overflow-x-auto p-4 text-sm leading-7 text-gray-200"><code>{{ codexInstallCommand }}</code></pre>
              </div>

              <div class="mt-6 space-y-3">
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ guideCopy.installVerifyTitle }}</h3>
                <div class="overflow-hidden rounded-xl bg-gray-900">
                  <div class="flex items-center justify-between border-b border-gray-700/60 px-4 py-3">
                    <span class="text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">{{ currentShellLabel }}</span>
                    <button type="button" class="text-xs text-gray-400 transition-colors hover:text-white" @click="copyText(codexVerifyCommand, 'codex-verify')">
                      {{ copiedKey === 'codex-verify' ? t('docs.copied') : t('docs.copy') }}
                    </button>
                  </div>
                  <pre class="overflow-x-auto p-4 text-sm leading-7 text-gray-200"><code>{{ codexVerifyCommand }}</code></pre>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-2xl border border-gray-200/80 bg-white/90 p-5 shadow-sm dark:border-dark-800 dark:bg-dark-900 sm:p-6">
          <div class="flex items-start gap-4">
            <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-semibold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">3</span>
            <div class="min-w-0 flex-1">
              <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ guideCopy.connectTitle }}</h2>
              <p class="mt-2 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ guideCopy.connectDescription }}</p>
              <p class="mt-3 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ configLocationText }}</p>

              <div class="mt-6 space-y-5">
                <div v-for="file in displayCodexFiles" :key="file.displayPath" class="overflow-hidden rounded-xl bg-gray-900">
                  <div class="flex items-center justify-between border-b border-gray-700/60 px-4 py-3">
                    <span class="text-xs font-semibold text-gray-400">{{ file.displayPath }}</span>
                    <button type="button" class="text-xs text-gray-400 transition-colors hover:text-white" @click="copyText(file.content, file.displayPath)">
                      {{ copiedKey === file.displayPath ? t('docs.copied') : t('docs.copy') }}
                    </button>
                  </div>
                  <pre class="overflow-x-auto p-4 text-sm leading-7 text-gray-200"><code>{{ file.content }}</code></pre>
                </div>
              </div>

              <p class="mt-5 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ guideCopy.keyReplaceText }}</p>
              <p class="mt-2 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ guideCopy.authNote }}</p>
            </div>
          </div>
        </section>

        <section class="grid gap-4 lg:grid-cols-2 lg:gap-6">
          <article class="rounded-2xl border border-gray-200/80 bg-white/90 p-5 shadow-sm dark:border-dark-800 dark:bg-dark-900 sm:p-6">
            <div class="flex items-start gap-4">
              <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-semibold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">4</span>
              <div class="min-w-0 flex-1">
                <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ guideCopy.vscodeTitle }}</h2>
                <p class="mt-2 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ guideCopy.vscodeDescription }}</p>

                <ul class="mt-5 space-y-2 text-sm leading-7 text-gray-600 dark:text-dark-400">
                  <li v-for="item in vscodeItems" :key="item" class="flex gap-2">
                    <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500"></span>
                    <span>{{ item }}</span>
                  </li>
                </ul>
              </div>
            </div>
          </article>

          <article class="rounded-2xl border border-gray-200/80 bg-white/90 p-5 shadow-sm dark:border-dark-800 dark:bg-dark-900 sm:p-6">
            <div class="flex items-start gap-4">
              <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-semibold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">5</span>
              <div class="min-w-0 flex-1">
                <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ guideCopy.startTitle }}</h2>
                <p class="mt-2 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ guideCopy.startDescription }}</p>

                <div class="mt-5 overflow-hidden rounded-xl bg-gray-900">
                  <div class="flex items-center justify-between border-b border-gray-700/60 px-4 py-3">
                    <span class="text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">{{ currentShellLabel }}</span>
                    <button type="button" class="text-xs text-gray-400 transition-colors hover:text-white" @click="copyText(startCommand, 'start-command')">
                      {{ copiedKey === 'start-command' ? t('docs.copied') : t('docs.copy') }}
                    </button>
                  </div>
                  <pre class="overflow-x-auto p-4 text-sm leading-7 text-gray-200"><code>{{ startCommand }}</code></pre>
                </div>

                <p class="mt-5 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ guideCopy.startHint }}</p>
              </div>
            </div>
          </article>
        </section>

        <section class="rounded-2xl border border-gray-200/80 bg-white/90 p-5 shadow-sm dark:border-dark-800 dark:bg-dark-900 sm:p-6">
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ guideCopy.faqTitle }}</h2>

          <div class="mt-6 space-y-6">
            <article v-for="item in faqItems" :key="item.title" class="space-y-3 rounded-xl border border-gray-200/80 p-5 dark:border-dark-700/80">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ item.title }}</h3>

              <div v-if="item.code" class="overflow-hidden rounded-xl bg-gray-900">
                <div class="flex items-center justify-between border-b border-gray-700/60 px-4 py-3">
                  <span class="text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">{{ item.shell }}</span>
                  <button type="button" class="text-xs text-gray-400 transition-colors hover:text-white" @click="copyText(item.code, item.title)">
                    {{ copiedKey === item.title ? t('docs.copied') : t('docs.copy') }}
                  </button>
                </div>
                <pre class="overflow-x-auto p-4 text-sm leading-7 text-gray-200"><code>{{ item.code }}</code></pre>
              </div>

              <ul v-if="item.bullets?.length" class="space-y-2 text-sm leading-7 text-gray-600 dark:text-dark-400">
                <li v-for="bullet in item.bullets" :key="bullet" class="flex gap-2">
                  <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500"></span>
                  <span>{{ bullet }}</span>
                </li>
              </ul>
            </article>
          </div>
        </section>
      </div>
    </main>

    <HomeFooter />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import HomeFooter from '@/components/home/HomeFooter.vue'
import HomeHeader from '@/components/home/HomeHeader.vue'
import { createCodexCliFiles, type CodexCliOS } from '@/utils/codexConfig'

type CodexGuidePlatform = 'mac' | 'windows' | 'linux'

interface GuideMethod {
  title: string
  shell: string
  code?: string
  bullets?: string[]
  note?: string
}

interface FaqItem {
  title: string
  shell: string
  code?: string
  bullets?: string[]
}

const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const copiedKey = ref<string | null>(null)
const codexPlatform = ref<CodexGuidePlatform>('mac')
const isDark = ref(document.documentElement.classList.contains('dark'))

const isZh = computed(() => locale.value.startsWith('zh'))
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'NanaFox API')
const apiBaseRoot = computed(() => (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, ''))
const codexFileOs = computed<CodexCliOS>(() => (codexPlatform.value === 'windows' ? 'windows' : 'unix'))
const currentShellLabel = computed(() => (codexPlatform.value === 'windows' ? 'PowerShell' : 'bash'))

const guideCopy = computed(() => {
  if (isZh.value) {
    return {
      kicker: '教程',
      title: 'Codex CLI 使用指南',
      subtitle: `配置 Codex CLI 连接 ${siteName.value} API 服务，替换端点和 API Key 后即可开箱即用。`,
      nodeTitle: '安装 Node.js',
      nodeDescription: 'Codex CLI 需要 Node.js v18 或更高版本。',
      nodeVerifyTitle: '安装完成后验证',
      installTitle: '安装 Codex CLI',
      installVerifyTitle: '验证安装',
      connectTitle: `连接 ${siteName.value} API 服务`,
      connectDescription: '需要创建两个配置文件：`config.toml` 和 `auth.json`。',
      keyReplaceText: '将 `your-api-key-here` 替换为你在控制台创建的 API Key。',
      authNote: '此方式通过 `auth.json` 文件存储 API 密钥，`config.toml` 中无需配置 `env_key` 字段。',
      vscodeTitle: 'VS Code 扩展配置（可选）',
      vscodeDescription: '如果你使用 VS Code，可以安装 Codex 扩展获得更好的 IDE 集成体验。',
      startTitle: '启动 Codex',
      startDescription: '在项目目录下运行：',
      startHint: '首次启动时，Codex 会进行初始化配置。如果连接正常，你会看到交互界面。',
      faqTitle: '常见问题'
    }
  }

  return {
    kicker: 'Guide',
    title: 'Codex CLI Guide',
    subtitle: `Configure Codex CLI to connect to the ${siteName.value} API service. Replace the endpoint and API key, then start immediately.`,
    nodeTitle: 'Install Node.js',
    nodeDescription: 'Codex CLI requires Node.js v18 or later.',
    nodeVerifyTitle: 'Verify the installation',
    installTitle: 'Install Codex CLI',
    installVerifyTitle: 'Verify the CLI',
    connectTitle: `Connect to the ${siteName.value} API service`,
    connectDescription: 'Create two config files: `config.toml` and `auth.json`.',
    keyReplaceText: 'Replace `your-api-key-here` with the API key you created in the dashboard.',
    authNote: 'This setup stores the API key in `auth.json`, so `config.toml` does not need an `env_key` field.',
    vscodeTitle: 'VS Code extension setup (optional)',
    vscodeDescription: 'If you use VS Code, install the Codex extension for a better IDE workflow.',
    startTitle: 'Start Codex',
    startDescription: 'Run this inside your project directory:',
    startHint: 'On first launch Codex performs initial setup. If the connection works, the interactive UI appears.',
    faqTitle: 'FAQ'
  }
})

const platforms = computed(() => ([
  { id: 'mac' as const, label: 'macOS' },
  { id: 'windows' as const, label: 'Windows' },
  { id: 'linux' as const, label: 'Linux' }
]))

const nodeMethods = computed<GuideMethod[]>(() => {
  if (isZh.value) {
    switch (codexPlatform.value) {
      case 'windows':
        return [
          {
            title: '方法一：官网下载（推荐）',
            shell: 'Windows',
            bullets: [
              '访问 nodejs.org 下载 LTS 版本',
              '双击 `.msi` 文件，按向导安装并保持默认设置'
            ]
          },
          {
            title: '方法二：使用包管理器',
            shell: 'PowerShell',
            code: '# 使用 Chocolatey\nchoco install nodejs\n\n# 或使用 Scoop\nscoop install nodejs',
            note: '建议使用 PowerShell 而非 CMD，以获得更好的体验。'
          }
        ]
      case 'linux':
        return [
          {
            title: '方法一：使用官方仓库（推荐）',
            shell: 'bash',
            code: '# 添加 NodeSource 仓库\ncurl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -\n# 安装 Node.js\nsudo apt-get install -y nodejs'
          },
          {
            title: '方法二：使用系统包管理器',
            shell: 'bash',
            code: '# Ubuntu / Debian\nsudo apt update && sudo apt install nodejs npm\n\n# CentOS / RHEL / Fedora\nsudo dnf install nodejs npm'
          }
        ]
      default:
        return [
          {
            title: '方法一：使用 Homebrew（推荐）',
            shell: 'bash',
            code: '# 更新 Homebrew\nbrew update\n# 安装 Node.js\nbrew install node'
          },
          {
            title: '方法二：官网下载',
            shell: 'macOS',
            bullets: [
              '访问 nodejs.org 下载 macOS LTS 版本',
              '打开 `.pkg` 文件，按安装向导完成安装'
            ]
          }
        ]
    }
  }

  switch (codexPlatform.value) {
    case 'windows':
      return [
        {
          title: 'Method 1: download from nodejs.org',
          shell: 'Windows',
          bullets: [
            'Download the current LTS installer from nodejs.org',
            'Open the `.msi` file and keep the default options'
          ]
        },
        {
          title: 'Method 2: use a package manager',
          shell: 'PowerShell',
          code: '# Chocolatey\nchoco install nodejs\n\n# Or Scoop\nscoop install nodejs',
          note: 'PowerShell is recommended over CMD for a smoother setup flow.'
        }
      ]
    case 'linux':
      return [
        {
          title: 'Method 1: use the official repository',
          shell: 'bash',
          code: '# Add the NodeSource repository\ncurl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -\n# Install Node.js\nsudo apt-get install -y nodejs'
        },
        {
          title: 'Method 2: use the system package manager',
          shell: 'bash',
          code: '# Ubuntu / Debian\nsudo apt update && sudo apt install nodejs npm\n\n# CentOS / RHEL / Fedora\nsudo dnf install nodejs npm'
        }
      ]
    default:
      return [
        {
          title: 'Method 1: use Homebrew',
          shell: 'bash',
          code: '# Update Homebrew\nbrew update\n# Install Node.js\nbrew install node'
        },
        {
          title: 'Method 2: download from nodejs.org',
          shell: 'macOS',
          bullets: [
            'Download the current macOS LTS package from nodejs.org',
            'Open the `.pkg` file and finish the installer'
          ]
        }
      ]
  }
})

const installLead = computed(() => {
  if (isZh.value) {
    return codexPlatform.value === 'windows'
      ? '以管理员身份运行 PowerShell，执行：'
      : '执行下面的命令安装 Codex CLI：'
  }

  return codexPlatform.value === 'windows'
    ? 'Run PowerShell as Administrator and execute:'
    : 'Run the following command to install Codex CLI:'
})

const nodeVerifyCommand = 'node --version\nnpm --version'
const codexInstallCommand = computed(() =>
  isZh.value
    ? 'npm i -g @openai/codex --registry=https://registry.npmmirror.com'
    : 'npm i -g @openai/codex'
)
const codexVerifyCommand = 'codex --version'

const displayCodexFiles = computed(() => {
  const files = createCodexCliFiles({
    baseUrl: apiBaseRoot.value,
    apiKey: 'your-api-key-here',
    os: codexFileOs.value,
    providerName: siteName.value
  })

  const prefix = codexPlatform.value === 'windows'
    ? 'C:\\Users\\your-username\\.codex\\'
    : '~/.codex/'

  return files.map((file) => ({
    ...file,
    displayPath: `${prefix}${file.path.split(/[\\/]/).pop()}`
  }))
})

const configLocationText = computed(() => {
  if (isZh.value) {
    return codexPlatform.value === 'windows'
      ? '配置文件位于 `C:\\Users\\你的用户名\\.codex\\` 目录（不存在则创建）。'
      : '配置文件位于 `~/.codex/` 目录（不存在则创建）。'
  }

  return codexPlatform.value === 'windows'
    ? 'Store the files in `C:\\Users\\your-username\\.codex\\` and create the directory if it does not exist.'
    : 'Store the files in `~/.codex/` and create the directory if it does not exist.'
})

const vscodeItems = computed(() => {
  if (isZh.value) {
    return [
      '在 VS Code 扩展商店搜索并安装 Codex – OpenAI’s coding agent',
      '确保已经按上面的步骤配置好 `config.toml` 和 `auth.json`',
      '如需扩展内读取密钥，可设置一个环境变量名并在扩展中引用该名称'
    ]
  }

  return [
    'Install Codex – OpenAI’s coding agent from the VS Code marketplace',
    'Make sure `config.toml` and `auth.json` are already configured',
    'If the extension needs a key, set an environment variable and reference its name there'
  ]
})

const startCommand = computed(() => (
  codexPlatform.value === 'windows'
    ? 'cd C:\\path\\to\\your\\project\ncodex'
    : 'cd /path/to/your/project\ncodex'
))

const faqItems = computed<FaqItem[]>(() => {
  if (isZh.value) {
    if (codexPlatform.value === 'windows') {
      return [
        {
          title: '1. 命令未找到',
          shell: 'PowerShell',
          bullets: [
            '确保 npm 全局路径（通常 `C:\\Users\\你的用户名\\AppData\\Roaming\\npm`）已添加到系统 PATH',
            '重新打开 PowerShell 窗口后再执行 `codex --version`'
          ]
        },
        {
          title: '2. API 连接失败',
          shell: 'PowerShell',
          code: `# 测试 API 端点\ncurl.exe -I ${apiBaseRoot.value}\n\n# 检查配置文件\nGet-Content $HOME\\.codex\\config.toml\nGet-Content $HOME\\.codex\\auth.json`
        },
        {
          title: '3. 更新 Codex CLI',
          shell: 'PowerShell',
          code: codexInstallCommand.value
        }
      ]
    }

    return [
      {
        title: '1. 命令未找到',
        shell: 'bash',
        code: codexPlatform.value === 'linux'
          ? '# 检查 npm 全局安装路径\nnpm config get prefix\n\n# 如果路径不在 PATH 中，添加到 ~/.bashrc\necho \'export PATH="$HOME/.npm-global/bin:$PATH"\' >> ~/.bashrc\nsource ~/.bashrc'
          : '# 检查 npm 全局安装路径\nnpm config get prefix\n\n# 如果路径不在 PATH 中，添加到 ~/.zshrc\necho \'export PATH="$HOME/.npm-global/bin:$PATH"\' >> ~/.zshrc\nsource ~/.zshrc'
      },
      {
        title: '2. API 连接失败',
        shell: 'bash',
        code: `# 测试 API 端点\ncurl -I ${apiBaseRoot.value}\n\n# 检查配置文件\ncat ~/.codex/config.toml\ncat ~/.codex/auth.json`
      },
      {
        title: '3. 更新 Codex CLI',
        shell: 'bash',
        code: codexInstallCommand.value
      }
    ]
  }

  if (codexPlatform.value === 'windows') {
    return [
      {
        title: '1. Command not found',
        shell: 'PowerShell',
        bullets: [
          'Make sure the npm global path (usually `C:\\Users\\your-username\\AppData\\Roaming\\npm`) is in PATH',
          'Open a new PowerShell window and run `codex --version` again'
        ]
      },
      {
        title: '2. API connection failed',
        shell: 'PowerShell',
        code: `# Test the API endpoint\ncurl.exe -I ${apiBaseRoot.value}\n\n# Inspect config files\nGet-Content $HOME\\.codex\\config.toml\nGet-Content $HOME\\.codex\\auth.json`
      },
      {
        title: '3. Update Codex CLI',
        shell: 'PowerShell',
        code: codexInstallCommand.value
      }
    ]
  }

  return [
    {
      title: '1. Command not found',
      shell: 'bash',
      code: codexPlatform.value === 'linux'
        ? '# Check the npm global install path\nnpm config get prefix\n\n# If the path is missing from PATH, add it to ~/.bashrc\necho \'export PATH="$HOME/.npm-global/bin:$PATH"\' >> ~/.bashrc\nsource ~/.bashrc'
        : '# Check the npm global install path\nnpm config get prefix\n\n# If the path is missing from PATH, add it to ~/.zshrc\necho \'export PATH="$HOME/.npm-global/bin:$PATH"\' >> ~/.zshrc\nsource ~/.zshrc'
    },
    {
      title: '2. API connection failed',
      shell: 'bash',
      code: `# Test the API endpoint\ncurl -I ${apiBaseRoot.value}\n\n# Inspect config files\ncat ~/.codex/config.toml\ncat ~/.codex/auth.json`
    },
    {
      title: '3. Update Codex CLI',
      shell: 'bash',
      code: codexInstallCommand.value
    }
  ]
})

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  const prefersDark =
    typeof window.matchMedia === 'function'
      ? window.matchMedia('(prefers-color-scheme: dark)').matches
      : false
  if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

async function copyText(value: string, key: string) {
  if (!navigator.clipboard) {
    return
  }

  await navigator.clipboard.writeText(value)
  copiedKey.value = key
  window.setTimeout(() => {
    copiedKey.value = null
  }, 1500)
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
