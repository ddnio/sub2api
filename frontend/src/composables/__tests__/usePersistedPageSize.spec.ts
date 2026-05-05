import { afterEach, describe, expect, it } from 'vitest'

import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

describe('usePersistedPageSize', () => {
  afterEach(() => {
    localStorage.clear()
    delete window.__APP_CONFIG__
  })

<<<<<<< HEAD
  it('uses the persisted user page size when it is an allowed option', () => {
=======
  it('uses the system table default instead of stale localStorage state', () => {
>>>>>>> v0.1.116
    window.__APP_CONFIG__ = {
      table_default_page_size: 1000,
      table_page_size_options: [20, 50, 1000]
    } as any
    localStorage.setItem('table-page-size', '50')
<<<<<<< HEAD

    expect(getPersistedPageSize()).toBe(50)
=======
    localStorage.setItem('table-page-size-source', 'user')

    expect(getPersistedPageSize()).toBe(1000)
>>>>>>> v0.1.116
  })
})
