import { describe, expect, it } from 'vitest'
import { enqueueUsageRequest } from '../usageLoadQueue'
import type { Account } from '@/types'

function makeAccount(
  platform: string,
  type: string = 'oauth',
  proxy?: { host: string; port: number; username?: string | null } | null
): Account {
  return {
    id: Math.floor(Math.random() * 10000),
    platform,
    type,
    name: 'test',
    status: 'active',
    proxy_id: proxy ? 1 : null,
    proxy: proxy
      ? { id: 1, name: 'p', protocol: 'http', host: proxy.host, port: proxy.port, username: proxy.username ?? null, status: 'active', created_at: '', updated_at: '' }
      : undefined,
    credentials: {},
    created_at: '',
    updated_at: ''
  } as unknown as Account
}

describe('usageLoadQueue', () => {
  it('serializes Anthropic accounts sharing the same proxy exit', async () => {
    const order: number[] = []
    const acc1 = makeAccount('anthropic', 'oauth', { host: '10.0.0.1', port: 3128, username: 'admin' })
    const acc2 = makeAccount('anthropic', 'setup-token', { host: '10.0.0.1', port: 3128, username: 'admin' })

    const p1 = enqueueUsageRequest(acc1, async () => {
      order.push(1)
      return 1
    })
    const p2 = enqueueUsageRequest(acc2, async () => {
      order.push(2)
      return 2
    })

    await Promise.all([p1, p2])
    expect(order).toEqual([1, 2])
  })

  it('keeps running queued tasks after one Anthropic task rejects', async () => {
    const results: string[] = []
    const acc = makeAccount('anthropic', 'oauth', { host: '99.99.99.99', port: 1234 })

    const p1 = enqueueUsageRequest(acc, async () => {
      throw new Error('fail')
    })
    const p2 = enqueueUsageRequest(acc, async () => {
      results.push('second')
      return 'ok'
    })

    await expect(p1).rejects.toThrow('fail')
    await p2
    expect(results).toEqual(['second'])
  })

  it('does not queue non-Anthropic platforms', async () => {
    const timestamps: number[] = []
    const makeFn = () => async () => {
      timestamps.push(Date.now())
      return 'ok'
    }

    const acc1 = makeAccount('gemini', 'oauth', { host: '1.2.3.4', port: 8080 })
    const acc2 = makeAccount('gemini', 'oauth', { host: '1.2.3.4', port: 8080 })

    await Promise.all([
      enqueueUsageRequest(acc1, makeFn()),
      enqueueUsageRequest(acc2, makeFn())
    ])

    expect(timestamps).toHaveLength(2)
    expect(Math.abs(timestamps[1] - timestamps[0])).toBeLessThan(50)
  })

  it('does not queue Anthropic API key accounts', async () => {
    const timestamps: number[] = []
    const makeFn = () => async () => {
      timestamps.push(Date.now())
      return 'ok'
    }

    const acc1 = makeAccount('anthropic', 'apikey', { host: '1.2.3.4', port: 8080 })
    const acc2 = makeAccount('anthropic', 'apikey', { host: '1.2.3.4', port: 8080 })

    await Promise.all([
      enqueueUsageRequest(acc1, makeFn()),
      enqueueUsageRequest(acc2, makeFn())
    ])

    expect(timestamps).toHaveLength(2)
    expect(Math.abs(timestamps[1] - timestamps[0])).toBeLessThan(50)
  })
})
