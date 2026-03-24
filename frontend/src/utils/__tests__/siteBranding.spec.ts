import { describe, expect, it } from 'vitest'
import { resolveSiteSubtitle } from '@/utils/siteBranding'

describe('resolveSiteSubtitle', () => {
  it('falls back to the localized subtitle for generic default values', () => {
    expect(resolveSiteSubtitle('Subscription to API Conversion Platform', '统一接入你的 AI 上游')).toBe('统一接入你的 AI 上游')
    expect(resolveSiteSubtitle('AI API Gateway for Developers', '统一接入你的 AI 上游')).toBe('统一接入你的 AI 上游')
  })

  it('keeps custom white-label subtitles intact', () => {
    expect(resolveSiteSubtitle('Acme API Gateway', '统一接入你的 AI 上游')).toBe('Acme API Gateway')
  })
})
