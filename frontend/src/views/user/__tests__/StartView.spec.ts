import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import StartView from '../StartView.vue'
import { startDoneKey } from '../startOnboarding'

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

  it('defaults to "write code" with the OpenAI group and /v1 address', async () => {
    const w = mountView()
    await flushPromises()
    expect(w.text()).toContain('OpenAI GPT')
    expect(w.get('[data-testid="start-base-url"]').text()).toBe(`${window.location.origin}/v1`)
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
