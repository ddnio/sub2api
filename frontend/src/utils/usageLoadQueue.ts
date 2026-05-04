/**
 * Usage request scheduler. Anthropic OAuth/setup-token accounts sharing the
 * same proxy exit are placed into a short serial queue to reduce upstream 429s.
 */

import type { Account } from '@/types'

const GROUP_DELAY_MIN_MS = 1000
const GROUP_DELAY_MAX_MS = 2000

type Task<T> = {
  fn: () => Promise<T>
  resolve: (value: T) => void
  reject: (reason: unknown) => void
}

const queues = new Map<string, Task<unknown>[]>()
const running = new Set<string>()

function needsThrottle(account: Account): boolean {
  return (
    account.platform === 'anthropic' &&
    (account.type === 'oauth' || account.type === 'setup-token')
  )
}

function buildGroupKey(account: Account): string {
  const proxy = account.proxy
  const proxyIdentity = proxy
    ? `${proxy.host}:${proxy.port}:${proxy.username || ''}`
    : 'direct'
  return `anthropic:${proxyIdentity}`
}

async function drain(groupKey: string) {
  if (running.has(groupKey)) return
  running.add(groupKey)

  const queue = queues.get(groupKey)
  while (queue && queue.length > 0) {
    const task = queue.shift()!
    try {
      const result = await task.fn()
      task.resolve(result)
    } catch (err) {
      task.reject(err)
    }
    if (queue.length > 0) {
      const jitter = GROUP_DELAY_MIN_MS + Math.random() * (GROUP_DELAY_MAX_MS - GROUP_DELAY_MIN_MS)
      await new Promise((resolve) => setTimeout(resolve, jitter))
    }
  }

  running.delete(groupKey)
  queues.delete(groupKey)
}

export function enqueueUsageRequest<T>(
  account: Account,
  fn: () => Promise<T>
): Promise<T> {
  if (!needsThrottle(account)) {
    return fn()
  }

  const key = buildGroupKey(account)
  return new Promise<T>((resolve, reject) => {
    let queue = queues.get(key)
    if (!queue) {
      queue = []
      queues.set(key, queue)
    }
    queue.push({ fn, resolve, reject } as Task<unknown>)
    drain(key)
  })
}
