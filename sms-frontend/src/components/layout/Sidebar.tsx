import { NavLink } from 'react-router-dom'
import { useRole } from '../../hooks/useRole'

interface NavItem {
  to: string
  label: string
  icon: string
  roles?: string[]
}

const navItems: NavItem[] = [
  { to: '/dashboard', label: 'Dashboard', icon: '⊞' },
  { to: '/profile', label: 'Profile', icon: '◉' },
  // Admin
  { to: '/admin/users', label: 'Users', icon: '👤', roles: ['Admin'] },
  { to: '/admin/students', label: 'Students', icon: '🎓', roles: ['Admin'] },
  { to: '/admin/teachers', label: 'Teachers', icon: '🏫', roles: ['Admin'] },
  { to: '/admin/classes', label: 'Classes', icon: '🏷️', roles: ['Admin'] },
  { to: '/admin/subjects', label: 'Subjects', icon: '📚', roles: ['Admin'] },
  { to: '/admin/enrollment', label: 'Enrollment', icon: '📋', roles: ['Admin'] },
  { to: '/admin/attendance-summary', label: 'Attendance', icon: '📊', roles: ['Admin'] },
  { to: '/admin/finance', label: 'Finance', icon: '💰', roles: ['Admin'] },
  { to: '/admin/notify', label: 'Notify', icon: '📣', roles: ['Admin'] },
  // Teacher
  { to: '/academics/attendance', label: 'Attendance', icon: '✅', roles: ['Teacher'] },
  { to: '/academics/grades', label: 'Grades', icon: '📝', roles: ['Teacher'] },
  // Student
  { to: '/academics/attendance', label: 'My Attendance', icon: '✅', roles: ['Student'] },
  { to: '/academics/grades', label: 'My Grades', icon: '📝', roles: ['Student'] },
  { to: '/academics/reportcard', label: 'Report Card', icon: '📄', roles: ['Student'] },
  { to: '/finance', label: 'Finance', icon: '💳', roles: ['Student'] },
  { to: '/locker', label: 'My Locker', icon: '🗂️', roles: ['Student'] },
  // Parent
  { to: '/parent/children', label: 'My Children', icon: '👨‍👧', roles: ['Parent'] },
  { to: '/finance', label: 'Payments', icon: '💳', roles: ['Parent'] },
  // Shared
  { to: '/notifications', label: 'Notifications', icon: '🔔' },
]

export function Sidebar() {
  const { role } = useRole()

  const visible = navItems.filter(
    (item) => !item.roles || (role && item.roles.includes(role))
  )

  return (
    <aside className="w-56 shrink-0 border-r border-[var(--border)] bg-[var(--bg)] flex flex-col min-h-screen">
      {/* Logo */}
      <div className="px-5 py-5 border-b border-[var(--border)]">
        <span className="text-base font-bold text-[var(--text-h)] tracking-tight">
          📐 SMS
        </span>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-2 py-3 space-y-0.5 overflow-y-auto">
        {visible.map((item) => (
          <NavLink
            key={item.to + (item.roles?.join('') ?? '')}
            to={item.to}
            className={({ isActive }) =>
              `flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors
               ${isActive
                ? 'bg-[var(--accent-bg)] text-[var(--accent)] font-medium'
                : 'text-[var(--text)] hover:bg-[var(--code-bg)] hover:text-[var(--text-h)]'
               }`
            }
          >
            <span className="text-base leading-none">{item.icon}</span>
            {item.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}