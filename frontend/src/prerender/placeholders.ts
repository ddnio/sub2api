// NanaFox fork: 预渲染页面里取自线上设置的值先用这些占位符，后端按 manifest 里的同一份表替换。
export const PRERENDER_SITE_ORIGIN = 'https://site.prerender.invalid'

export const PRERENDER_PLACEHOLDERS = {
  siteOrigin: PRERENDER_SITE_ORIGIN,
  apiBase: 'https://api.prerender.invalid',
  siteName: '__PRERENDER_SITE_NAME__',
  contact: '__PRERENDER_CONTACT__'
}

export const PRERENDER_DIR = 'prerender'
export const PRERENDER_MANIFEST = 'manifest.json'
