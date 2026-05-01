import { describe, expect, it } from 'vitest'
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

const srcRoot = join(process.cwd(), 'src')
const stylePath = join(srcRoot, 'style.css')
const sourceExtensions = new Set(['.vue', '.ts', '.css'])

function collectSourceFiles(dir: string, files: string[] = []): string[] {
  for (const entry of readdirSync(dir)) {
    const fullPath = join(dir, entry)
    const stat = statSync(fullPath)
    if (stat.isDirectory()) {
      if (entry === 'node_modules' || entry === 'dist') continue
      collectSourceFiles(fullPath, files)
      continue
    }
    if (sourceExtensions.has(fullPath.slice(fullPath.lastIndexOf('.')))) {
      files.push(fullPath)
    }
  }
  return files
}

function collectButtonClassUsages(): Set<string> {
  const classes = new Set<string>()
  for (const file of collectSourceFiles(srcRoot)) {
    const content = readFileSync(file, 'utf8')
    for (const match of content.matchAll(/\bbtn-[A-Za-z0-9_-]+\b/g)) {
      classes.add(match[0])
    }
  }
  return classes
}

function collectDefinedButtonClasses(): Set<string> {
  const content = readFileSync(stylePath, 'utf8')
  return new Set(
    [...content.matchAll(/\.btn-[A-Za-z0-9_-]+\b/g)].map(match => match[0].slice(1)),
  )
}

describe('global button classes', () => {
  it('defines every custom btn-* class used under src', () => {
    const used = collectButtonClassUsages()
    const defined = collectDefinedButtonClasses()
    const missing = [...used].filter(className => !defined.has(className)).sort()

    expect(missing, `Missing definitions in ${relative(process.cwd(), stylePath)}`).toEqual([])
  })
})
