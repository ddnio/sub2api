import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import StartView from '../StartView.vue'
import { startDoneKey } from '../startOnboarding'
import UseKeyModal from '@/components/keys/UseKeyModal.vue'
import { buildOpenAICodexConfigToml } from '@/utils/codexAuthConfig'
import { createCodexCliFiles } from '@/utils/codexConfig'

const { listKeys, createKey, getAvailable, fetchPublicSettings, showError, replace, routeQuery } = vi.hoisted(() => ({
  listKeys: vi.fn(),
  createKey: vi.fn(),
  getAvailable: vi.fn(),
  fetchPublicSettings: vi.fn(),
  showError: vi.fn(),
  replace: vi.fn(),
  routeQuery: { value: {} as Record<string, string> }
}))

vi.mock('@/api', () => ({
  keysAPI: { list: listKeys, create: createKey },
  userGroupsAPI: { getAvailable }
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: routeQuery.value }),
  useRouter: () => ({ replace, push: replace })
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({ user: { id: 7, role: 'user', balance: 5, created_at: '2026-11-23T00:00:00Z' } }),
  useAppStore: () => ({ fetchPublicSettings, showError, showSuccess: vi.fn() })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn() })
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

const GROUPS = [
  { id: 3, name: 'OpenAI GPT', platform: 'openai', status: 'active', subscription_type: 'standard', is_exclusive: false, claude_code_only: false },
  { id: 4, name: 'ClaudeCode专用', platform: 'anthropic', status: 'active', subscription_type: 'standard', is_exclusive: false, claude_code_only: false }
]

function mountView() {
  return mount(StartView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        UseKeyModal: {
          props: ['show', 'apiKey', 'baseUrl', 'platform', 'claudeCodeOnly', 'allowMessagesDispatch'],
          template: '<div data-testid="use-key-modal" :data-show="show" :data-platform="platform" :data-key="apiKey" />'
        }
      }
    }
  })
}

describe('StartView', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
    vi.clearAllMocks()
    routeQuery.value = {}
    listKeys.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 1, pages: 0 })
    getAvailable.mockResolvedValue(GROUPS)
    fetchPublicSettings.mockResolvedValue({ api_base_url: '', site_name: 'NanaFox Router', hide_ccs_import_button: false })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('sends users who already have a key straight to the dashboard', async () => {
    listKeys.mockResolvedValue({ items: [{ id: 1 }], total: 1, page: 1, page_size: 1, pages: 1 })
    mountView()
    await flushPromises()
    expect(replace).toHaveBeenCalledWith('/dashboard')
    expect(localStorage.getItem(startDoneKey(7))).toBe('1')
    expect(getAvailable).not.toHaveBeenCalled()
  })

  it('skip marks done and leaves', async () => {
    const w = mountView()
    await flushPromises()
    await w.get('[data-testid="start-skip"]').trigger('click')
    expect(replace).toHaveBeenCalledWith('/dashboard')
    expect(localStorage.getItem(startDoneKey(7))).toBe('1')
  })

  it('defaults to codex with codex-first option order and inline setup without /v1 hint', async () => {
    createKey.mockResolvedValue({ id: 9, name: '我的第一个 key', key: 'sk-codex-123' })
    const w = mountView()
    await flushPromises()

    const useCaseIds = w.findAll('[data-testid^="start-usecase-"]').map((el) => el.attributes('data-testid'))
    expect(useCaseIds).toEqual([
      'start-usecase-codex',
      'start-usecase-claude',
      'start-usecase-client',
      'start-usecase-code'
    ])
    expect(w.text()).toContain('OpenAI GPT')
    expect(w.find('[data-testid="start-base-url"]').exists()).toBe(false)
    expect(w.text()).not.toContain('要加 /v1')
    expect(w.get('[data-testid="start-skip"]').classes()).not.toContain('btn-secondary')

    await w.get('[data-testid="start-create"]').trigger('click')
    await flushPromises()
    expect(w.get('[data-testid="use-key-modal"]').attributes('data-show')).toBe('false')
    expect(w.get('[data-testid="start-ccs"]').classes()).toContain('btn-primary')

    const expectedToml = buildOpenAICodexConfigToml({
      baseUrl: window.location.origin,
      apiKey: 'sk-codex-123'
    })
    const unixCmd = w.get('[data-testid="start-codex-command"]').text()
    expect(unixCmd).toContain('CODEX_HOME')
    expect(unixCmd).toContain(expectedToml)
    const safetyHint = w.get('[data-testid="start-codex-safety-hint"]').text()
    expect(safetyHint).toContain('自动备份')
    expect(safetyHint).toContain('不要发给别人')
    expect(safetyHint).toContain('公司受监控的电脑')

    await w.get('[data-testid="start-codex-shell-powershell"]').trigger('click')
    const psCmd = w.get('[data-testid="start-codex-command"]').text()
    expect(psCmd).toContain('Set-Content -LiteralPath $t -Encoding UTF8')
    expect(psCmd).toContain(expectedToml)

    // Cross-check directly against UseKeyModal's rendered OpenAI Codex config
    const realModal = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-codex-123',
        baseUrl: window.location.origin,
        platform: 'openai'
      },
      global: {
        mocks: { $t: (k: string) => k },
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          Icon: { template: '<span />' },
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })
    const modalToml = realModal.get('pre code').text()
    const docsFiles = createCodexCliFiles({
      baseUrl: window.location.origin,
      apiKey: 'sk-codex-123',
      os: 'unix'
    })
    expect(docsFiles).toHaveLength(1)
    expect(docsFiles[0].content).toBe(modalToml)
    expect(unixCmd).toContain(modalToml)
    expect(psCmd).toContain(modalToml)
  })

  it('shows /v1 base url hint when switching to "write code"', async () => {
    const w = mountView()
    await flushPromises()
    await w.get('[data-testid="start-usecase-code"]').trigger('click')
    expect(w.get('[data-testid="start-base-url"]').text()).toBe(`${window.location.origin}/v1`)
    expect(w.text()).toContain('要加 /v1')
  })

  it('resumes directly at step 3 with an existing active OpenAI key when ?step=config is passed', async () => {
    routeQuery.value = { step: 'config', use: 'codex' }
    listKeys.mockResolvedValue({
      items: [
        { id: 10, name: 'disabled-key', key: 'sk-disabled', status: 'inactive', group: GROUPS[0] },
        { id: 11, name: 'existing-key', key: 'sk-existing-999', status: 'active', group: GROUPS[0] }
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1
    })
    const w = mountView()
    await flushPromises()
    expect(listKeys).toHaveBeenCalledWith(1, 50, { status: 'active' })
    expect(replace).not.toHaveBeenCalled()
    expect(w.get('[data-testid="start-key"]').text()).toBe('sk-existing-999')
    expect(w.get('[data-testid="start-codex-command"]').text()).toContain('sk-existing-999')
  })

  it('continues past page 1 when ?step=config finds matching active key on page 2', async () => {
    routeQuery.value = { step: 'config', use: 'codex' }
    const page1Items = Array.from({ length: 50 }, (_, i) => ({
      id: i + 1,
      name: `claude-${i}`,
      key: `sk-claude-${i}`,
      status: 'active',
      group: GROUPS[1]
    }))
    listKeys
      .mockResolvedValueOnce({ items: page1Items, total: 51, page: 1, page_size: 50, pages: 2 })
      .mockResolvedValueOnce({
        items: [{ id: 51, name: 'codex-page2', key: 'sk-codex-page2', status: 'active', group: GROUPS[0] }],
        total: 51,
        page: 2,
        page_size: 50,
        pages: 2
      })
    const w = mountView()
    await flushPromises()
    expect(listKeys).toHaveBeenNthCalledWith(1, 1, 50, { status: 'active' })
    expect(listKeys).toHaveBeenNthCalledWith(2, 2, 50, { status: 'active' })
    expect(w.get('[data-testid="start-key"]').text()).toBe('sk-codex-page2')
  })

  it('stays at key creation step when ?step=config&use=codex has only Claude or inactive keys', async () => {
    routeQuery.value = { step: 'config', use: 'codex' }
    listKeys.mockResolvedValue({
      items: [
        { id: 12, name: 'claude-only', key: 'sk-claude-only', status: 'active', group: GROUPS[1] },
        { id: 13, name: 'disabled-openai', key: 'sk-disabled', status: 'inactive', group: GROUPS[0] }
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1
    })
    const w = mountView()
    await flushPromises()
    expect(replace).not.toHaveBeenCalled()
    expect(w.find('[data-testid="start-key"]').exists()).toBe(false)
    expect(w.get('[data-testid="start-create"]').exists()).toBe(true)
    expect(w.get('[data-testid="start-codex-command"]').text()).not.toContain('sk-claude-only')
  })

  it('preselects Claude Code from ?use=claude and creates the key in ClaudeCode专用', async () => {
    routeQuery.value = { use: 'claude' }
    createKey.mockResolvedValue({ id: 9, name: '我的第一个 key', key: 'sk-test-123' })
    const w = mountView()
    await flushPromises()
    expect(w.get('[data-testid="start-base-url"]').text()).toBe(window.location.origin)
    await w.get('[data-testid="start-create"]').trigger('click')
    await flushPromises()
    expect(createKey).toHaveBeenCalledWith('我的第一个 key', 4)
    expect(w.get('[data-testid="start-key"]').text()).toBe('sk-test-123')
    const modal = w.get('[data-testid="use-key-modal"]')
    expect(modal.attributes('data-show')).toBe('true')
    expect(modal.attributes('data-platform')).toBe('anthropic')
    expect(modal.attributes('data-key')).toBe('sk-test-123')
    expect(localStorage.getItem(startDoneKey(7))).toBe('1')
  })

  it('falls back to the use case remembered from the signup link', async () => {
    sessionStorage.setItem('nf_start_use', 'claude')
    createKey.mockResolvedValue({ id: 9, name: 'k', key: 'sk-a' })
    const w = mountView()
    await flushPromises()
    await w.get('[data-testid="start-create"]').trigger('click')
    await flushPromises()
    expect(createKey).toHaveBeenCalledWith('我的第一个 key', 4)
  })

  it('switching use case before creating changes the group', async () => {
    createKey.mockResolvedValue({ id: 9, name: 'k', key: 'sk-x' })
    const w = mountView()
    await flushPromises()
    await w.get('[data-testid="start-usecase-claude"]').trigger('click')
    await w.get('[data-testid="start-create"]').trigger('click')
    await flushPromises()
    expect(createKey).toHaveBeenCalledWith('我的第一个 key', 4)
  })

  it('test button calls the gateway with the new key and shows the reply', async () => {
    createKey.mockResolvedValue({ id: 9, name: 'k', key: 'sk-o' })
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ choices: [{ message: { content: '你好！' } }] })
    })
    vi.stubGlobal('fetch', fetchMock)
    const w = mountView()
    await flushPromises()
    await w.get('[data-testid="start-create"]').trigger('click')
    await flushPromises()
    await w.get('[data-testid="start-test"]').trigger('click')
    await flushPromises()
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe(`${window.location.origin}/v1/chat/completions`)
    expect(init.headers.Authorization).toBe('Bearer sk-o')
    expect(w.get('[data-testid="start-test-result"]').text()).toContain('接通了')
    expect(w.get('[data-testid="start-test-result"]').text()).toContain('你好！')
  })

  it('shows status and error message when the test fails', async () => {
    createKey.mockResolvedValue({ id: 9, name: 'k', key: 'sk-o' })
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 403,
      json: () => Promise.resolve({ error: { message: 'insufficient balance' } })
    }))
    const w = mountView()
    await flushPromises()
    await w.get('[data-testid="start-create"]').trigger('click')
    await flushPromises()
    await w.get('[data-testid="start-test"]').trigger('click')
    await flushPromises()
    const text = w.get('[data-testid="start-test-result"]').text()
    expect(text).toContain('HTTP 403')
    expect(text).toContain('insufficient balance')
    expect(text).toContain('nanafox-hi')
  })

  it('locks the use case cards while the key is being created', async () => {
    let resolve: (v: unknown) => void = () => {}
    createKey.mockReturnValue(new Promise((r) => (resolve = r)))
    const w = mountView()
    await flushPromises()
    await w.get('[data-testid="start-create"]').trigger('click')
    expect(w.get('[data-testid="start-usecase-claude"]').attributes('disabled')).toBeDefined()
    resolve({ id: 9, name: 'k', key: 'sk-x' })
    await flushPromises()
  })

  it('skips the browser test for Claude Code client-only groups', async () => {
    routeQuery.value = { use: 'claude' }
    getAvailable.mockResolvedValue(GROUPS.map((g) => (g.id === 4 ? { ...g, claude_code_only: true } : g)))
    createKey.mockResolvedValue({ id: 9, name: 'k', key: 'sk-a' })
    const w = mountView()
    await flushPromises()
    await w.get('[data-testid="start-create"]').trigger('click')
    await flushPromises()
    expect(w.find('[data-testid="start-test"]').exists()).toBe(false)
    expect(w.get('[data-testid="start-test-cli-only"]').text()).toContain('claude')
  })

  it('hides the CCS button when the site setting says so', async () => {
    fetchPublicSettings.mockResolvedValue({ api_base_url: '', hide_ccs_import_button: true })
    const w = mountView()
    await flushPromises()
    expect(w.find('[data-testid="start-ccs"]').exists()).toBe(false)
  })

  it('shows a contact hint when no usable group exists', async () => {
    getAvailable.mockResolvedValue([])
    const w = mountView()
    await flushPromises()
    expect(w.get('[data-testid="start-create"]').attributes('disabled')).toBeDefined()
    expect(w.text()).toContain('nanafox-hi')
  })
})
