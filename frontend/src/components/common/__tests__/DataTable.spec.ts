import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'

import DataTable from '../DataTable.vue'

const createTestI18n = () => createI18n({
  legacy: false,
  locale: 'en',
  messages: {
    en: {
      empty: { noData: 'No data' },
    },
  },
})

const stubDesktopMatchMedia = () => {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: query.includes('min-width: 768px'),
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  })
}

function mountDesktopTable(props: Record<string, unknown> = {}) {
  return mount(DataTable, {
    props: {
      columns: [
        { key: 'id', label: 'ID' },
        { key: 'name', label: 'Name' },
      ],
      data: [],
      ...props,
    },
    global: {
      plugins: [createTestI18n()],
      stubs: {
        Icon: true,
      },
    },
    attachTo: document.body,
  })
}

describe('DataTable', () => {
  beforeEach(() => {
    stubDesktopMatchMedia()
    localStorage.clear()
  })

  it('renders desktop rows even when the virtualizer has no visible items yet', async () => {
    const wrapper = mountDesktopTable({
      data: [
        { id: 1, name: 'First order' },
        { id: 2, name: 'Second order' },
      ],
    })

    try {
      await wrapper.vm.$nextTick()

      expect(wrapper.text()).toContain('First order')
      expect(wrapper.text()).toContain('Second order')
      expect(wrapper.text()).not.toContain('No data')
    } finally {
      wrapper.unmount()
    }
  })

  it('keeps desktop tables wide enough to expose horizontal overflow', async () => {
    const wrapper = mountDesktopTable({
      data: [{ id: 1, name: 'First order' }],
    })

    try {
      await wrapper.vm.$nextTick()

      const table = wrapper.find('table')
      expect(table.classes()).toContain('w-full')
      expect(table.classes()).toContain('min-w-max')
    } finally {
      wrapper.unmount()
    }
  })

  it('renders paired sort arrows and highlights the active direction', async () => {
    const wrapper = mountDesktopTable({
      columns: [
        { key: 'name', label: 'Name', sortable: true },
        { key: 'created_at', label: 'Created', sortable: true },
      ],
      data: [
        { id: 1, name: 'Beta', created_at: '2026-01-02T00:00:00Z' },
        { id: 2, name: 'Alpha', created_at: '2026-01-01T00:00:00Z' },
      ],
      defaultSortKey: 'name',
      defaultSortOrder: 'asc',
    })

    try {
      await wrapper.vm.$nextTick()

      const nameHeader = wrapper.findAll('th')[0]
      expect(nameHeader.attributes('aria-sort')).toBe('ascending')
      expect(nameHeader.findAll('svg')).toHaveLength(2)
      expect(nameHeader.findAll('svg')[0].classes()).toContain('text-primary-600')
      expect(nameHeader.findAll('svg')[1].classes()).toContain('text-gray-300')

      await nameHeader.trigger('click')
      await wrapper.vm.$nextTick()

      expect(nameHeader.attributes('aria-sort')).toBe('descending')
      expect(nameHeader.findAll('svg')[0].classes()).toContain('text-gray-300')
      expect(nameHeader.findAll('svg')[1].classes()).toContain('text-primary-600')
    } finally {
      wrapper.unmount()
    }
  })

  it('renders every row with no virtual padding spacer for small datasets (virtualization off)', async () => {
    const data = Array.from({ length: 8 }, (_, i) => ({ id: i + 1, name: `Row ${i + 1}` }))
    const wrapper = mountDesktopTable({
      columns: [{ key: 'name', label: 'Name' }],
      data,
    })

    try {
      await wrapper.vm.$nextTick()

      // Virtualization is OFF for a small list…
      expect((wrapper.vm as any).shouldVirtualize).toBe(false)
      // …every row is in the DOM…
      expect(wrapper.findAll('tbody tr[data-index]')).toHaveLength(data.length)
      // …and there are no aria-hidden virtual padding spacer rows.
      expect(wrapper.findAll('tbody tr[aria-hidden="true"]')).toHaveLength(0)
    } finally {
      wrapper.unmount()
    }
  })

  it('switches to windowed rendering once row count exceeds virtualizeThreshold', async () => {
    const data = Array.from({ length: 12 }, (_, i) => ({ id: i + 1, name: `Row ${i + 1}` }))
    const wrapper = mountDesktopTable({
      columns: [{ key: 'name', label: 'Name' }],
      data,
      virtualizeThreshold: 3,
    })

    try {
      await wrapper.vm.$nextTick()

      // Virtualization is ON: the mode-switch decision flipped…
      expect((wrapper.vm as any).shouldVirtualize).toBe(true)
      // …and the virtualizer drives off the full row count.
      const exposed = (wrapper.vm as any).virtualizer
      const instance = exposed?.value ?? exposed
      expect(instance.options.count).toBe(data.length)
    } finally {
      wrapper.unmount()
    }
  })

  it('keys the virtualizer size cache by row identity, not index (avoids stale heights on sort/filter)', async () => {
    const data = Array.from({ length: 12 }, (_, i) => ({ id: 100 + i, name: `Row ${i + 1}` }))
    const wrapper = mountDesktopTable({
      columns: [{ key: 'name', label: 'Name' }],
      data,
      rowKey: 'id',
      virtualizeThreshold: 3,
    })

    try {
      await wrapper.vm.$nextTick()

      const exposed = (wrapper.vm as any).virtualizer
      const instance = exposed?.value ?? exposed
      // getItemKey must resolve to the row's stable key (id), not the positional index.
      expect(instance.options.getItemKey(0)).toBe(100)
      expect(instance.options.getItemKey(5)).toBe(105)
    } finally {
      wrapper.unmount()
    }
  })
})
