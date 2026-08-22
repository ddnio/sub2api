import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  buildEmbeddedUrl,
  buildImageCreationEmbeddedUrl,
  detectTheme,
  isImageCreationEmbedUrl,
} from '../embedded-url'

describe('embedded-url', () => {
  const originalLocation = window.location

  beforeEach(() => {
    Object.defineProperty(window, 'location', {
      value: {
        origin: 'https://app.example.com',
        href: 'https://app.example.com/user/purchase',
      },
      writable: true,
      configurable: true,
    })
  })

  afterEach(() => {
    Object.defineProperty(window, 'location', {
      value: originalLocation,
      writable: true,
      configurable: true,
    })
    document.documentElement.classList.remove('dark')
    vi.restoreAllMocks()
  })

  it('adds embedded query parameters including locale and source context', () => {
    const result = buildEmbeddedUrl(
      'https://pay.example.com/checkout?plan=pro',
      42,
      'token-123',
      'dark',
      'zh-CN',
    )

    const url = new URL(result)
    expect(url.searchParams.get('plan')).toBe('pro')
    expect(url.searchParams.get('user_id')).toBe('42')
    expect(url.searchParams.get('token')).toBe('token-123')
    expect(url.searchParams.get('theme')).toBe('dark')
    expect(url.searchParams.get('lang')).toBe('zh-CN')
    expect(url.searchParams.get('ui_mode')).toBe('embedded')
    expect(url.searchParams.get('src_host')).toBe('https://app.example.com')
    expect(url.searchParams.get('src_url')).toBe('https://app.example.com/user/purchase')
  })

  it('omits optional params when they are empty', () => {
    const result = buildEmbeddedUrl('https://pay.example.com/checkout', undefined, '', 'light')

    const url = new URL(result)
    expect(url.searchParams.get('theme')).toBe('light')
    expect(url.searchParams.get('ui_mode')).toBe('embedded')
    expect(url.searchParams.has('user_id')).toBe(false)
    expect(url.searchParams.has('token')).toBe(false)
    expect(url.searchParams.has('lang')).toBe(false)
  })

  it('returns original string for invalid url input', () => {
    expect(buildEmbeddedUrl('not a url', 1, 'token')).toBe('not a url')
  })

  it('puts the image creation launch ticket in a scrub-friendly fragment without JWT context', () => {
    const result = buildImageCreationEmbeddedUrl(
      'https://app.example.com/tools/image-playground/',
      'one-time-ticket',
      'dark',
      'zh-CN',
    )

    const url = new URL(result)
    const fragment = new URLSearchParams(url.hash.slice(1))
    expect(url.search).toBe('')
    expect(fragment.get('launch')).toBe('one-time-ticket')
    expect(fragment.get('theme')).toBe('dark')
    expect(fragment.get('lang')).toBe('zh-CN')
    expect(fragment.get('ui_mode')).toBe('embedded')
    expect(fragment.get('src_host')).toBe('https://app.example.com')
    expect(fragment.get('src_url')).toBe('https://app.example.com/user/purchase')
    expect(result).not.toContain('user_id=')
    expect(result).not.toContain('token=')
  })

  it('only selects the stable image playground path for scoped launch tickets', () => {
    expect(isImageCreationEmbedUrl('https://app.example.com/tools/image-playground/')).toBe(true)
    expect(isImageCreationEmbedUrl('https://app.example.com/tools/image-playground')).toBe(true)
    expect(isImageCreationEmbedUrl('https://app.example.com/tools/image-studio/')).toBe(false)
    expect(isImageCreationEmbedUrl('https://other.example.com/tools/image-playground/')).toBe(false)
  })

  it('detects dark mode from document root class', () => {
    document.documentElement.classList.add('dark')
    expect(detectTheme()).toBe('dark')
  })
})
