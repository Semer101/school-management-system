import type { User } from './user'

export interface Notification {
  id: number
  title: string
  body: string
  target_roles: string
  sender_id: number
  sender?: User
  created_at: string
}

export interface NotificationReceipt {
  id: number
  user_id: number
  user?: User
  notification_id: number
  notification?: Notification
  is_read: boolean
  read_at: string | null
}