import { describe, expect, it } from 'vitest'
import { extractI18nErrorMessage } from '@/utils/apiError'

function t(key: string, params?: Record<string, unknown>): string {
  const messages: Record<string, string> = {
    'payment.errors.WXPAY_CONFIG_MISSING_KEY': 'Missing {key}',
    'payment.errors.WXPAY_CONFIG_INVALID_KEY_LENGTH': '{key} expected {expected}, got {actual}',
    'admin.settings.payment.field_certSerial': 'Certificate Serial',
    'admin.settings.payment.field_apiV3Key': 'API v3 Key',
  }
  let value = messages[key] ?? key
  for (const [paramKey, paramValue] of Object.entries(params ?? {})) {
    value = value.replace(`{${paramKey}}`, String(paramValue))
  }
  return value
}

t.te = (key: string) => key in {
  'payment.errors.WXPAY_CONFIG_MISSING_KEY': true,
  'payment.errors.WXPAY_CONFIG_INVALID_KEY_LENGTH': true,
  'admin.settings.payment.field_certSerial': true,
  'admin.settings.payment.field_apiV3Key': true,
}

describe('extractI18nErrorMessage', () => {
  it('localizes structured payment errors and config field metadata', () => {
    const message = extractI18nErrorMessage(
      {
        reason: 'WXPAY_CONFIG_MISSING_KEY',
        message: 'missing_required_key',
        metadata: { key: 'certSerial' },
      },
      t,
      'payment.errors',
      'fallback',
    )

    expect(message).toBe('Missing Certificate Serial')
  })

  it('interpolates numeric metadata for structured payment errors', () => {
    const message = extractI18nErrorMessage(
      {
        reason: 'WXPAY_CONFIG_INVALID_KEY_LENGTH',
        metadata: { key: 'apiV3Key', expected: '32', actual: '5' },
      },
      t,
      'payment.errors',
      'fallback',
    )

    expect(message).toBe('API v3 Key expected 32, got 5')
  })

  it('falls back to the backend message when no i18n key exists', () => {
    const message = extractI18nErrorMessage(
      { reason: 'UNKNOWN_PAYMENT_ERROR', message: 'backend message' },
      t,
      'payment.errors',
      'fallback',
    )

    expect(message).toBe('backend message')
  })
})
