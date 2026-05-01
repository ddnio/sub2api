import { describe, expect, it } from 'vitest'
import zh from '../locales/zh'
import en from '../locales/en'
import { normalizeLocaleMessages } from '../index'

function readPath(obj: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((current, segment) => {
    if (!current || typeof current !== 'object') return undefined
    return (current as Record<string, unknown>)[segment]
  }, obj)
}

describe('payment admin locale namespace', () => {
  const requiredKeys = [
    'admin.settings.payment.providerManagement',
    'admin.settings.payment.providerWxpay',
    'admin.settings.payment.enabledPaymentTypes',
    'admin.settings.payment.modeQRCode',
    'admin.settings.payment.limitSingleMin',
  ]

  it.each([
    ['zh', zh],
    ['en', en],
  ] as const)('resolves provider management keys in %s', (_locale, messages) => {
    normalizeLocaleMessages(messages)
    for (const key of requiredKeys) {
      expect(readPath(messages, key), key).toEqual(expect.any(String))
      expect(readPath(messages, key), key).not.toBe('')
    }
  })
})
