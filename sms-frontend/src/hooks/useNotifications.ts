import { useState, useEffect, useCallback } from 'react'
import {
  getMyNotifications,
  markAsRead as apiMarkAsRead,
  issueSSEToken,
  openNotificationStream,
} from '../api/notifications'
import { useAuth } from './useAuth'

export interface NotificationItem {
  receipt_id: number
  id: number
  title: string
  body: string
  sender_name: string
  is_read: boolean
  read_at: string | null
  received_at: string
}

function normalizeReceipts(payload: unknown): NotificationItem[] {
  if (!Array.isArray(payload)) return []
  return payload.map((raw) => {
    const r = raw as Record<string, unknown>
    const nested = r.notification as Record<string, unknown> | undefined
    return {
      receipt_id: Number(r.receipt_id ?? r.id),
      id: Number(r.id ?? nested?.id ?? r.notification_id),
      title: String(r.title ?? nested?.title ?? 'Notification'),
      body: String(r.body ?? nested?.body ?? ''),
      sender_name: String(r.sender_name ?? (nested?.sender as Record<string, unknown>)?.name ?? ''),
      is_read: Boolean(r.is_read),
      read_at: (r.read_at as string) ?? null,
      received_at: String(r.received_at ?? nested?.created_at ?? ''),
    }
  })
}

export function useNotifications() {
  const { user } = useAuth()
  const [items, setItems] = useState<NotificationItem[]>([])
  const [loading, setLoading] = useState(false)

  const unreadCount = items.filter((r) => !r.is_read).length

  const fetchAll = useCallback(async () => {
    if (!user) return
    setLoading(true)
    try {
      const res = await getMyNotifications()
      setItems(normalizeReceipts(res.data.data ?? []))
    } finally {
      setLoading(false)
    }
  }, [user])

  const markAsRead = async (receiptId: number) => {
    await apiMarkAsRead(receiptId)
    setItems((prev) =>
      prev.map((r) =>
        r.receipt_id === receiptId ? { ...r, is_read: true, read_at: new Date().toISOString() } : r
      )
    )
  }

  useEffect(() => {
    if (!user) return

    fetchAll()

    let es: EventSource | null = null

    issueSSEToken()
      .then(({ data }) => {
        const token = data.data?.sse_token
        if (!token) return
        es = openNotificationStream(token)
        es.onmessage = () => fetchAll()
        es.onerror = () => es?.close()
      })
      .catch(() => {})

    return () => {
      es?.close()
    }
  }, [user, fetchAll])

  return { receipts: items, items, unreadCount, loading, markAsRead, refetch: fetchAll }
}
