/**
<<<<<<< HEAD
 * Usage request scheduler. Anthropic OAuth/setup-token accounts sharing the
 * same proxy exit are placed into a short serial queue to reduce upstream 429s.
=======
 * Usage request scheduler — throttles Anthropic API calls by proxy exit.
 *
 * Anthropic OAuth/setup-token accounts sharing the same proxy exit are placed
 * into a serial queue with a random 1–2s delay between requests, preventing
 * upstream 429 rate-limit errors.
 *
 * Proxy identity = host:port:username — two proxy records pointing to the
 * same exit share a single queue. Accounts without a proxy go into a
 * "direct" queue.
 *
 * All other platforms bypass the queue and execute immediately.
>>>>>>> v0.1.116
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

<<<<<<< HEAD
=======
/** Whether this account needs throttled queuing. */
>>>>>>> v0.1.116
function needsThrottle(account: Account): boolean {
  return (
    account.platform === 'anthropic' &&
    (account.type === 'oauth' || account.type === 'setup-token')
  )
}

<<<<<<< HEAD
=======
/** Build a queue key from proxy connection details. */
>>>>>>> v0.1.116
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
<<<<<<< HEAD
      await new Promise((resolve) => setTimeout(resolve, jitter))
=======
      await new Promise((r) => setTimeout(r, jitter))
>>>>>>> v0.1.116
    }
  }

  running.delete(groupKey)
  queues.delete(groupKey)
}

<<<<<<< HEAD
=======
/**
 * Schedule a usage fetch. Anthropic accounts are queued by proxy exit;
 * all other platforms execute immediately.
 */
>>>>>>> v0.1.116
export function enqueueUsageRequest<T>(
  account: Account,
  fn: () => Promise<T>
): Promise<T> {
<<<<<<< HEAD
=======
  // Non-Anthropic → fire immediately, no queuing
>>>>>>> v0.1.116
  if (!needsThrottle(account)) {
    return fn()
  }

  const key = buildGroupKey(account)
<<<<<<< HEAD
=======

>>>>>>> v0.1.116
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
