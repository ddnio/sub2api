import { apiClient } from './client'

interface ImageCreationTicket {
  ticket: string
  expires_in: number
}

export async function issueImageCreationTicket(admin: boolean) {
  const surface = admin ? 'admin' : 'user'
  const path = admin ? '/admin/image-creation/embed-tickets' : '/image-creation/embed-tickets'
  const { data } = await apiClient.post<ImageCreationTicket>(path, { surface })
  return data.ticket
}
