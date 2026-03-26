<!-- frontend/src/components/docs/DocsSectionCodexCli.vue -->
<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">{{ guideCopy.title }}</h1>
      <p class="mt-3 text-sm leading-7 text-gray-600 dark:text-dark-400">{{ guideCopy.subtitle }}</p>
    </div>

    <!-- Platform selector -->
    <div class="inline-flex rounded-full border border-gray-200 bg-gray-50 p-1 dark:border-dark-700 dark:bg-dark-950">
      <button
        v-for="p in platforms"
        :key="p.id"
        type="button"
        class="rounded-full px-4 py-2 text-sm font-medium transition-colors"
        :class="activePlatform === p.id
          ? 'bg-gray-900 text-white dark:bg-white dark:text-dark-950'
          : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'"
        @click="activePlatform = p.id"
      >
        {{ p.label }}
      </button>
    </div>

    <!-- Step 1: Install Node.js -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">1</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ guideCopy.nodeTitle }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ guideCopy.nodeDescription }}</p>
        </div>
      </div>

      <!-- Node install methods -->
      <div class="space-y-4">
        <div v-for="method in nodeMethods" :key="method.title" class="space-y-2">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ method.title }}</h3>

          <DocsCodeBlock
            v-if="method.code"
            :tabs="[{ id: 'code', label: method.shell, path: method.shell, content: method.code }]"
          />

          <ul v-if="method.bullets?.length" class="space-y-2 text-sm leading-7 text-gray-600 dark:text-dark-400">
            <li v-for="bullet in method.bullets" :key="bullet" class="flex gap-2">
              <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500"></span>
              <span>{{ bullet }}</span>
            </li>
          </ul>

          <p v-if="method.note" class="text-sm leading-7 text-gray-600 dark:text-dark-400">{{ method.note }}</p>
        </div>
      </div>

      <!-- Verify Node.js -->
      <div class="space-y-2">
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ guideCopy.nodeVerifyTitle }}</h3>
        <DocsCodeBlock :tabs="[{ id: 'verify', label: currentShellLabel, path: currentShellLabel, content: nodeVerifyCommand }]" />
      </div>
    </section>

    <!-- Step 2: Install Codex CLI -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">2</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ guideCopy.installTitle }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ installLead }}</p>
        </div>
      </div>
      <DocsCodeBlock :tabs="[{ id: 'install', label: 'npm', path: 'npm', content: codexInstallCommand }]" />

      <div class="space-y-2">
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ guideCopy.installVerifyTitle }}</h3>
        <DocsCodeBlock :tabs="[{ id: 'verify', label: currentShellLabel, path: currentShellLabel, content: codexVerifyCommand }]" />
      </div>
    </section>

    <!-- Step 3: Configure -->
    <section class="space-y-3">
      <div class="flex items-start gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">3</span>
        <div class="min-w-0 flex-1">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ guideCopy.connectTitle }}</h2>
          <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ guideCopy.connectDescription }}</p>
          <p class="mt-2 text-sm text-gray-600 dark:text-dark-400">{{ configLocationText }}</p>
        </div>
      </div>

      <DocsCodeBlock
        v-for="file in displayCodexFiles"
        :key="file.displayPath"
        :tabs="[{ id: file.displayPath, label: file.displayPath, path: file.displayPath, content: file.content }]"
      />

      <p class="text-sm text-gray-600 dark:text-dark-400">{{ guideCopy.keyReplaceText }}</p>
      <p class="text-sm text-gray-600 dark:text-dark-400">{{ guideCopy.authNote }}</p>
    </section>

    <!-- Step 4 & 5: VSCode + Start (side by side on large screens) -->
    <div class="grid gap-4 lg:grid-cols-2">
      <!-- Step 4: VS Code -->
      <section class="space-y-3">
        <div class="flex items-start gap-3">
          <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">4</span>
          <div class="min-w-0 flex-1">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ guideCopy.vscodeTitle }}</h2>
            <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ guideCopy.vscodeDescription }}</p>
            <ul class="mt-3 space-y-2 text-sm leading-7 text-gray-600 dark:text-dark-400">
              <li v-for="item in vscodeItems" :key="item" class="flex gap-2">
                <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500"></span>
                <span>{{ item }}</span>
              </li>
            </ul>
          </div>
        </div>
      </section>

      <!-- Step 5: Start -->
      <section class="space-y-3">
        <div class="flex items-start gap-3">
          <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">5</span>
          <div class="min-w-0 flex-1">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ guideCopy.startTitle }}</h2>
            <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">{{ guideCopy.startDescription }}</p>
          </div>
        </div>
        <DocsCodeBlock :tabs="[{ id: 'start', label: currentShellLabel, path: currentShellLabel, content: startCommand }]" />
        <p class="text-sm text-gray-600 dark:text-dark-400">{{ guideCopy.startHint }}</p>
      </section>
    </div>

    <!-- FAQ -->
    <section class="space-y-4">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ guideCopy.faqTitle }}</h2>
      <div class="space-y-4">
        <article v-for="item in faqItems" :key="item.title" class="rounded-xl border border-gray-200/80 p-4 dark:border-dark-700">
          <h3 class="font-semibold text-gray-900 dark:text-white">{{ item.title }}</h3>

          <DocsCodeBlock
            v-if="item.code"
            class="mt-3"
            :tabs="[{ id: item.title, label: item.shell, path: item.shell, content: item.code }]"
          />

          <ul v-if="item.bullets?.length" class="mt-2 space-y-2 text-sm leading-7 text-gray-600 dark:text-dark-400">
            <li v-for="bullet in item.bullets" :key="bullet" class="flex gap-2">
              <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500"></span>
              <span>{{ bullet }}</span>
            </li>
          </ul>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import DocsCodeBlock from '@/components/docs/DocsCodeBlock.vue'
import { createCodexCliFiles, type CodexCliOS } from '@/utils/codexConfig'
import { useSyncedTabState } from '@/composables/useSyncedTabState'

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

const { locale } = useI18n()
const appStore = useAppStore()

const { activeTab: activePlatform } = useSyncedTabState({
  group: 'os',
  scope: 'docs',
  availableTabs: ['mac', 'windows', 'linux'],
  defaultTab: 'mac'
})

const isZh = computed(() => locale.value.startsWith('zh'))
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'NanaFox API')
const apiBaseRoot = computed(() => (appStore.apiBaseUrl || window.location.origin).replace(/\/v1\/?$/, '').replace(/\/+$/, ''))
const codexFileOs = computed<CodexCliOS>(() => (activePlatform.value === 'windows' ? 'windows' : 'unix'))
const currentShellLabel = computed(() => (activePlatform.value === 'windows' ? 'PowerShell' : 'bash'))

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
    switch (activePlatform.value as CodexGuidePlatform) {
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

  switch (activePlatform.value as CodexGuidePlatform) {
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
    return activePlatform.value === 'windows'
      ? '以管理员身份运行 PowerShell，执行：'
      : '执行下面的命令安装 Codex CLI：'
  }

  return activePlatform.value === 'windows'
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

  const prefix = activePlatform.value === 'windows'
    ? 'C:\\Users\\your-username\\.codex\\'
    : '~/.codex/'

  return files.map((file) => ({
    ...file,
    displayPath: `${prefix}${file.path.split(/[\\/]/).pop()}`
  }))
})

const configLocationText = computed(() => {
  if (isZh.value) {
    return activePlatform.value === 'windows'
      ? '配置文件位于 `C:\\Users\\你的用户名\\.codex\\` 目录（不存在则创建）。'
      : '配置文件位于 `~/.codex/` 目录（不存在则创建）。'
  }

  return activePlatform.value === 'windows'
    ? 'Store the files in `C:\\Users\\your-username\\.codex\\` and create the directory if it does not exist.'
    : 'Store the files in `~/.codex/` and create the directory if it does not exist.'
})

const vscodeItems = computed(() => {
  if (isZh.value) {
    return [
      '在 VS Code 扩展商店搜索并安装 Codex \u2013 OpenAI\u2019s coding agent',
      '确保已经按上面的步骤配置好 `config.toml` 和 `auth.json`',
      '如需扩展内读取密钥，可设置一个环境变量名并在扩展中引用该名称'
    ]
  }

  return [
    'Install Codex – OpenAI\u2019s coding agent from the VS Code marketplace',
    'Make sure `config.toml` and `auth.json` are already configured',
    'If the extension needs a key, set an environment variable and reference its name there'
  ]
})

const startCommand = computed(() => (
  activePlatform.value === 'windows'
    ? 'cd C:\\path\\to\\your\\project\ncodex'
    : 'cd /path/to/your/project\ncodex'
))

const faqItems = computed<FaqItem[]>(() => {
  if (isZh.value) {
    if (activePlatform.value === 'windows') {
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
        code: activePlatform.value === 'linux'
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

  if (activePlatform.value === 'windows') {
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
      code: activePlatform.value === 'linux'
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
</script>
