import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const view = readFileSync(resolve(dir, '../CustomPageView.vue'), 'utf8')
const router = readFileSync(resolve(dir, '../../../router/index.ts'), 'utf8')
const sidebar = readFileSync(resolve(dir, '../../../components/layout/AppSidebar.vue'), 'utf8')

describe('自定义页面 surface 边界', () => {
  it('用户入口和管理员入口使用不同路由', () => {
    expect(router).toContain("path: '/custom/:id'")
    expect(router).toContain("props: { surface: 'user' }")
    expect(router).toContain("path: '/admin/custom/:id'")
    expect(router).toContain("name: 'AdminCustomPage'")
    expect(router).toContain("props: { surface: 'admin' }")
    expect(sidebar).toContain('`/admin/custom/${cm.id}`')
  })

  it('票据 scope 由入口决定，不由当前账号角色决定', () => {
    expect(view).toContain("surface?: 'user' | 'admin'")
    expect(view).toContain("issueImageCreationTicket(props.surface === 'admin')")
    expect(view).not.toContain('issueImageCreationTicket(authStore.isAdmin)')
  })
})
