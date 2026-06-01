import { useNotifications } from '../../hooks/useNotifications'
import { Spinner } from '../../components/ui/Spinner'
import { Bell } from 'lucide-react'
import { EmptyState } from '../../components/ui/EmptyState'
import { Button } from '../../components/ui/Button'

function timeAgo(dateStr: string) {
  if (!dateStr) return ''
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

export default function NotificationsPage() {
  const { items, loading, markAsRead } = useNotifications()

  if (loading) return <Spinner fullPage />

  if (items.length === 0) {
    return <EmptyState icon={Bell} title="No notifications yet" description="Announcements from your school will appear here." />
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-[var(--text-h)]">Notifications</h1>
        <span className="text-sm text-[var(--text)]">
          {items.filter((r) => !r.is_read).length} unread
        </span>
      </div>

      <div className="space-y-3">
        {items.map((item) => (
          <article
            key={item.receipt_id}
            className={`px-5 py-4 rounded-xl border transition-colors
              ${item.is_read
                ? 'bg-[var(--bg)] border-[var(--border)]'
                : 'bg-[var(--accent-bg)] border-[var(--accent-border)]'
              }`}
          >
            <div className="flex items-start justify-between gap-3">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-2">
                  {!item.is_read && (
                    <span className="w-2 h-2 rounded-full bg-[var(--accent)] shrink-0" />
                  )}
                  <h2 className="text-sm font-semibold text-[var(--text-h)]">
                    {item.title}
                  </h2>
                </div>
                <p className="text-sm text-[var(--text)] leading-relaxed whitespace-pre-wrap">
                  {item.body || '(No message body)'}
                </p>
                <p className="text-xs text-[var(--text)] mt-2">
                  {item.received_at ? timeAgo(item.received_at) : ''}
                  {item.sender_name && (
                    <span className="ml-2">· from {item.sender_name}</span>
                  )}
                </p>
              </div>
              {!item.is_read && (
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => markAsRead(item.receipt_id)}
                >
                  Mark read
                </Button>
              )}
            </div>
          </article>
        ))}
      </div>
    </div>
  )
}
