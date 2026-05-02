import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import AccountBulkActionsBar from '../AccountBulkActionsBar.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('AccountBulkActionsBar', () => {
  it('disables filtered bulk edit until filters are active', async () => {
    const wrapper = mount(AccountBulkActionsBar, {
      props: {
        selectedIds: [],
        hasActiveFilters: false
      }
    })

    const button = wrapper.get('button.btn-primary')
    expect(button.attributes('disabled')).toBeDefined()

    await button.trigger('click')
    expect(wrapper.emitted('edit-filtered')).toBeUndefined()
  })

  it('emits filtered bulk edit when filters are active', async () => {
    const wrapper = mount(AccountBulkActionsBar, {
      props: {
        selectedIds: [],
        hasActiveFilters: true
      }
    })

    const button = wrapper.get('button.btn-primary')
    expect(button.attributes('disabled')).toBeUndefined()

    await button.trigger('click')
    expect(wrapper.emitted('edit-filtered')).toHaveLength(1)
  })
})
