import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { useRole } from '../../hooks/useRole'
import { useNotifications } from '../../hooks/useNotifications'
import { Badge, roleBadgeVariant } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'

interface DashCard {
  icon: string
  label: string
  to: string
  roles: string[]
}

const cards: DashCard[] = [
  { icon: '🎓', label: 'Students', to: '/admin/students', roles: ['Admin'] },
  { icon: '🏫', label: 'Teachers', to: '/admin/teachers', roles: ['Admin'] },
  { icon: '🏷️', label: 'Classes', to: '/admin/classes', roles: ['Admin'] },
  { icon: '📚', label: 'Subjects', to: '/admin/subjects', roles: ['Admin'] },
  { icon: '📋', label: 'Enrollment', to: '/admin/enrollment', roles: ['Admin'] },
  { icon: '📊', label: 'Attendance', to: '/admin/attendance-summary', roles: ['Admin'] },
  { icon: '💰', label: 'Finance', to: '/admin/finance', roles: ['Admin'] },
  { icon: '📣', label: 'Broadcast', to: '/admin/notify', roles: ['Admin'] },
  { icon: '✅', label: 'Attendance', to: '/academics/attendance', roles: ['Teacher', 'Student'] },
  { icon: '📝', label: 'Grades', to: '/academics/grades', roles: ['Teacher', 'Student'] },
  { icon: '📄', label: 'Report Card', to: '/academics/reportcard', roles: ['Student'] },
  { icon: '💳', label: 'Finance', to: '/finance', roles: ['Student', 'Parent'] },
  { icon: '🗂️', label: 'Locker', to: '/locker', roles: ['Student'] },
  { icon: '👨‍👧', label: 'My Children', to: '/parent/children', roles: ['Parent'] },
]

export default function DashboardPage() {
  const { user } = useAuth()
  const { role } = useRole()
  const { unreadCount } = useNotifications()
  const navigate = useNavigate()

  const visible = cards.filter((c) => role && c.roles.includes(role))

  return (
    <div className="max-w-4xl mx-auto">
      {/* Welcome */}
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-1">
          <h1 className="text-2xl font-bold text-[var(--text-h)]">
            Welcome back, {user?.name?.split(' ')[0]} 👋
          </h1>
          {role && <Badge label={role} variant={roleBadgeVariant(role)} />}
        </div>
        <p className="text-sm text-[var(--text)]">{user?.email}</p>
      </div>

      {/* Notification banner */}
      {unreadCount > 0 && (
        <div
          className="mb-6 flex items-center justify-between px-4 py-3 rounded-xl bg-[var(--accent-bg)] border border-[var(--accent-border)] cursor-pointer"
          onClick={() => navigate('/notifications')}
        >
          <p className="text-sm text-[var(--accent)] font-medium">
            🔔 You have {unreadCount} unread notification{unreadCount > 1 ? 's' : ''}
          </p>
          <span className="text-xs text-[var(--accent)]">View →</span>
        </div>
      )}

      {/* Quick access cards */}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
        {visible.map((card) => (
          <button
            key={card.to + card.label}
            onClick={() => navigate(card.to)}
            className="flex flex-col items-center justify-center gap-2 p-5 rounded-xl border border-[var(--border)] bg-[var(--bg)] hover:border-[var(--accent-border)] hover:bg-[var(--accent-bg)] transition-all duration-150 group"
          >
            <span className="text-2xl group-hover:scale-110 transition-transform">{card.icon}</span>
            <span className="text-sm font-medium text-[var(--text-h)]">{card.label}</span>
          </button>
        ))}
      </div>

      {/* Profile shortcut */}
      <div className="mt-6 flex justify-end">
        <Button variant="ghost" size="sm" onClick={() => navigate('/profile')}>
          ◉ My Profile
        </Button>
      </div>
    </div>
  )
}