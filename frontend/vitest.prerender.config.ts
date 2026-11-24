// NanaFox fork: 构建后的 /docs 预渲染（src/prerender/docs.prerender.ts），不属于单元测试
import { defineConfig } from 'vitest/config'
import base from './vitest.config'

export default defineConfig({
  ...base,
  test: {
    ...base.test,
    include: ['src/prerender/**/*.prerender.ts'],
    environmentOptions: { jsdom: { url: 'https://site.prerender.invalid/' } }
  }
})
