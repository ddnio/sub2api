const GENERIC_SITE_SUBTITLES = new Set([
  'Subscription to API Conversion Platform',
  'AI API Gateway for Developers'
])

export function resolveSiteSubtitle(siteSubtitle: string | null | undefined, fallback: string): string {
  const normalized = siteSubtitle?.trim()
  if (!normalized || GENERIC_SITE_SUBTITLES.has(normalized)) {
    return fallback
  }

  return normalized
}
