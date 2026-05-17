import { useState, useEffect, useCallback } from 'react'
import type { NotificationReceipt } from '../types/notifications'
import {
  getMyNotifications,
  markAsRead as apiMarkAsRead,
  issueSSEToken,
  openNotificationStream,
} from '../api/notifications'
import { useAuth } from './useAuth'

export function useNotifications() {
  const { user } = useAuth()
  const [receipts, setReceipts] = useState<NotificationReceipt[]>([])
  const [loading, setLoading] = useState(false)

  const unreadCount = receipts.filter((r) => !r.is_read).length

  const fetchAll = useCallback(async () => {
    if (!user) return
    setLoading(true)
    try {
      const res = await getMyNotifications()
      setReceipts(res.data.data ?? [])
    } finally {
      setLoading(false)
    }
  }, [user])

  const markAsRead = async (id: number) => {
    await apiMarkAsRead(id)
    setReceipts((prev) =>
      prev.map((r) =>
        r.notification_id === id ? { ...r, is_read: true, read_at: new Date().toISOString() } : r
      )
    )
  }

  // Start SSE stream
  useEffect(() => {
    if (!user) return

    fetchAll()

    let es: EventSource | null = null

    issueSSEToken()
      .then(({ data }) => {
        const token = data.data?.sse_token
        if (!token) return
        es = openNotificationStream(token)
        es.onmessage = () => {
          // Re-fetch on any new event
          fetchAll()
        }
        es.onerror = () => es?.close()
      })
      .catch(() => {})

    return () => {
      es?.close()
    }
  }, [user, fetchAll])

  return { receipts, unreadCount, loading, markAsRead, refetch: fetchAll }
}