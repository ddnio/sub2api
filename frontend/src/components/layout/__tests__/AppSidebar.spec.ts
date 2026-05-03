import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar smooth collapse styles', () => {
  it('keeps sidebar text nodes mounted and hides them with collapse classes', () => {
    expect(componentSource).toContain('sidebar-brand-collapsed')
    expect(componentSource).toContain('sidebar-label-collapsed')
    expect(componentSource).toContain('sidebar-link-collapsed')
    expect(componentSource).toContain('sidebar-section-title-collapsed')
    expect(componentSource).not.toContain('<span v-if="!sidebarCollapsed"')
  })
})
