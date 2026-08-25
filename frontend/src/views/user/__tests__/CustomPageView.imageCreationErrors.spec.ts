import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import CustomPageView from '@/views/user/CustomPageView.vue'

const { issueImageCreationTicket, menuItem } = vi.hoisted(() => ({
  issueImageCreationTicket: vi.fn(),
  menuItem: {
    id: 'image-creation',
    label: '图像创作',
    icon_svg: '',
    url: 'http://localhost:3000/tools/image-playground/',
    visibility: 'user' as const,
    sort_order: 1,
  },
}))

vi.mock('@/api/imageCreation', () => ({
  issueImageCreationTicket,
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({
    name: 'CustomPage',
    params: { id: menuItem.id },
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const { ref } = await import('vue')
  return {
    ...await importOriginal<typeof import('vue-i18n')>(),
    useI18n: () => ({
      locale: ref('zh'),
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    publicSettingsLoaded: true,
    cachedPublicSettings: { custom_menu_items: [menuItem] },
    fetchPublicSettings: vi.fn(),
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ user: { id: 1 }, token: 'token' }),
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({ customMenuItems: [] }),
}))

function mountView() {
  return mount(CustomPageView, {
    global: {
      stubs: {
        AppLayout: {
          template: '<main><header><slot name="header-action" /></header><slot /></main>',
        },
        Icon: true,
      },
    },
  })
}

describe('CustomPageView 图像创作错误交互', () => {
  beforeEach(() => {
    issueImageCreationTicket.mockReset()
    vi.restoreAllMocks()
  })

  it('新窗口被拦截时保留已加载的 iframe，并提供对应的重试动作', async () => {
    issueImageCreationTicket.mockResolvedValue('iframe-ticket')
    const open = vi.spyOn(window, 'open').mockReturnValue(null)
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-testid="image-creation-frame"]').exists()).toBe(true)

    await wrapper.get('[data-testid="image-creation-open-new-window"]').trigger('click')
    await flushPromises()

    expect(open).toHaveBeenCalledTimes(1)
    expect(wrapper.find('[data-testid="image-creation-frame"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="image-creation-window-error"]').text()).toContain(
      'customPage.imageCreationPopupBlocked',
    )

    const popup = {
      opener: window,
      location: { replace: vi.fn() },
      close: vi.fn(),
    }
    open.mockReturnValue(popup as unknown as Window)
    issueImageCreationTicket.mockResolvedValueOnce('window-ticket')

    await wrapper.get('[data-testid="image-creation-retry-new-window"]').trigger('click')
    await flushPromises()

    expect(open).toHaveBeenCalledTimes(2)
    expect(popup.location.replace).toHaveBeenCalledWith(expect.stringContaining('#launch=window-ticket'))
    expect(wrapper.find('[data-testid="image-creation-window-error"]').exists()).toBe(false)
  })

  it('iframe 会话失败时只重试 iframe 会话', async () => {
    issueImageCreationTicket
      .mockRejectedValueOnce(new Error('session unavailable'))
      .mockResolvedValueOnce('retry-ticket')
    const open = vi.spyOn(window, 'open').mockReturnValue(null)
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-testid="image-creation-session-error"]').text()).toContain('session unavailable')
    expect(wrapper.find('[data-testid="image-creation-frame"]').exists()).toBe(false)

    await wrapper.get('[data-testid="image-creation-retry-session"]').trigger('click')
    await flushPromises()

    expect(open).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="image-creation-session-error"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="image-creation-frame"]').attributes('src')).toContain('#launch=retry-ticket')
  })
})
