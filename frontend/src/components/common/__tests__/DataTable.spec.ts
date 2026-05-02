import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import DataTable from '../DataTable.vue'

function mountDesktopTable(data: Array<Record<string, unknown>>) {
  const originalMatchMedia = window.matchMedia
  window.matchMedia = vi.fn().mockImplementation((query: string) => ({
    matches: query.includes('min-width: 768px'),
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })) as typeof window.matchMedia

  const i18n = createI18n({
    legacy: false,
    locale: 'en',
    messages: {
      en: {
        empty: { noData: 'No data' },
      },
    },
  })

  const wrapper = mount(DataTable, {
    props: {
      columns: [
        { key: 'id', label: 'ID' },
        { key: 'name', label: 'Name' },
      ],
      data,
    },
    global: {
      plugins: [i18n],
      stubs: {
        Icon: true,
      },
    },
    attachTo: document.body,
  })

  return {
    wrapper,
    restore: () => {
      window.matchMedia = originalMatchMedia
      wrapper.unmount()
    },
  }
}

describe('DataTable', () => {
  it('renders desktop rows even when the virtualizer has no visible items yet', async () => {
    const { wrapper, restore } = mountDesktopTable([
      { id: 1, name: 'First order' },
      { id: 2, name: 'Second order' },
    ])

    try {
      await wrapper.vm.$nextTick()

      expect(wrapper.text()).toContain('First order')
      expect(wrapper.text()).toContain('Second order')
      expect(wrapper.text()).not.toContain('No data')
    } finally {
      restore()
    }
  })
})
