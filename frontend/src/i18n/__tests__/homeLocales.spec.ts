import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

function readPath(obj: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((current, segment) => {
    if (!current || typeof current !== 'object') return undefined
    return (current as Record<string, unknown>)[segment]
  }, obj)
}

describe('home locale namespace', () => {
  const requiredKeys = [
    'home.badge',
    'home.nav.docs',
    'home.nav.keyUsage',
    'home.nav.support',
    'home.hero.viewDocs',
    'home.hero.baseUrlHint',
    'home.hero.copyBaseUrl',
    'home.hero.copiedBaseUrl',
    'home.hero.snippetTitle',
    'home.hero.runLabel',
    'home.metrics.compatibilityTitle',
    'home.metrics.compatibilityValue',
    'home.metrics.routingTitle',
    'home.metrics.routingValue',
    'home.metrics.billingTitle',
    'home.metrics.billingValue'
  ]

  it.each([
    ['zh', zh],
    ['en', en]
  ] as const)('contains split home keys in %s', (_locale, messages) => {
    for (const key of requiredKeys) {
      const value = readPath(messages, key)
      expect(value, key).toEqual(expect.any(String))
      expect(value, key).not.toBe('')
    }
  })
})
