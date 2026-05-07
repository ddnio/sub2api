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
  it('emits filtered bulk edit from the filtered bulk action', async () => {
    const wrapper = mount(AccountBulkActionsBar, {
      props: {
        selectedIds: []
      }
    })

    const button = wrapper.get('button.btn-primary')
    expect(button.attributes('disabled')).toBeUndefined()

    await button.trigger('click')
    expect(wrapper.emitted('edit-filtered')).toHaveLength(1)
  })
})
