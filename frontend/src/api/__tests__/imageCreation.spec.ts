import { beforeEach, describe, expect, it, vi } from 'vitest'

const { post } = vi.hoisted(() => ({ post: vi.fn() }))

vi.mock('@/api/client', () => ({ apiClient: { post } }))

import { issueImageCreationTicket } from '@/api/imageCreation'

describe('image creation ticket api', () => {
  beforeEach(() => {
    post.mockReset().mockResolvedValue({ data: { ticket: 'one-time-ticket', expires_in: 60 } })
  })

  it.each([
    [false, '/image-creation/embed-tickets', 'user'],
    [true, '/admin/image-creation/embed-tickets', 'admin'],
  ])('uses the isolated ticket endpoint for admin=%s', async (admin, path, surface) => {
    await expect(issueImageCreationTicket(admin)).resolves.toBe('one-time-ticket')
    expect(post).toHaveBeenCalledWith(path, { surface })
  })
})
