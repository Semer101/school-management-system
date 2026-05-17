import { useNotifications } from '../../hooks/useNotifications'
import { Spinner } from '../../components/ui/Spinner'
import { EmptyState } from '../../components/ui/EmptyState'
import { Button } from '../../components/ui/Button'

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

export default function NotificationsPage() {
  const { receipts, loading, markAsRead } = useNotifications()

  if (loading) return <Spinner fullPage />

  if (receipts.length === 0) {
    return <EmptyState icon="🔔" title="No notifications yet" description="Announcements from your school will appear here." />
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-[var(--text-h)]">Notifications</h1>
        <span className="text-sm text-[var(--text)]">
          {receipts.filter((r) => !r.is_read).length} unread
        </span>
      </div>

      <div className="space-y-2">
        {receipts.map((receipt) => (
          <div
            key={receipt.id}
            className={`px-5 py-4 rounded-xl border transition-colors
              ${receipt.is_read
                ? 'bg-[var(--bg)] border-[var(--border)]'
                : 'bg-[var(--accent-bg)] border-[var(--accent-border)]'
              }`}
          >
            <div className="flex items-start justify-between gap-3">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  {!receipt.is_read && (
                    <span className="w-2 h-2 rounded-full bg-[var(--accent)] shrink-0" />
                  )}
                  <p className="text-sm font-semibold text-[var(--text-h)] truncate">
                    {receipt.notification?.title ?? 'Notification'}
                  </p>
                </div>
                <p className="text-sm text-[var(--text)] leading-relaxed">
                  {receipt.notification?.body}
                </p>
                <p className="text-xs text-[var(--text)] mt-1.5">
                  {receipt.notification?.created_at
                    ? timeAgo(receipt.notification.created_at)
                    : ''}
                  {receipt.notification?.sender && (
                    <span className="ml-2">· from {receipt.notification.sender.name}</span>
                  )}
                </p>
              </div>
              {!receipt.is_read && (
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => markAsRead(receipt.notification_id)}
                >
                  Mark read
                </Button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}