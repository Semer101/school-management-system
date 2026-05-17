import api from './axiosClient'
import type { APIResponse } from '../types/api'
import type { NotificationReceipt } from '../types/notifications'

export const getMyNotifications = () =>
  api.get<APIResponse<NotificationReceipt[]>>('/api/notifications')

export const markAsRead = (id: number) =>
  api.patch<APIResponse>(`/api/notifications/${id}/read`)

export const issueSSEToken = () =>
  api.post<APIResponse<{ sse_token: string }>>('/api/notifications/sse-token')

// Opens a Server-Sent Events stream. Returns the EventSource.
export const openNotificationStream = (sseToken: string): EventSource => {
  const base = import.meta.env.VITE_API_BASE_URL ?? ''
  const path = `/api/notifications/stream?sse_token=${encodeURIComponent(sseToken)}`
  return new EventSource(base ? `${base}${path}` : path)
}
