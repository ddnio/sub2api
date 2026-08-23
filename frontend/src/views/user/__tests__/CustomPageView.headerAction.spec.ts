import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const view = readFileSync(resolve(dir, '../CustomPageView.vue'), 'utf8')
const layout = readFileSync(resolve(dir, '../../../components/layout/AppLayout.vue'), 'utf8')
const header = readFileSync(resolve(dir, '../../../components/layout/AppHeader.vue'), 'utf8')

describe('图像创作宿主页头操作', () => {
  it('把新窗口入口放在页标题旁，并移除 iframe 上方的重复工具条', () => {
    expect(view).toContain('<template v-if="isImageCreationMode" #header-action>')
    expect(view).not.toContain('custom-embed-toolbar')
    expect(layout).toContain('<slot name="header-action" />')
    expect(header).toContain('<slot name="title-action" />')
  })
})
