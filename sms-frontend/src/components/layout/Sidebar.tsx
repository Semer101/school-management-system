import { NavLink } from 'react-router-dom'
import { useRole } from '../../hooks/useRole'
import { navIcons, type NavIconKey } from '../../lib/nav-icons'
import { cn } from '../../lib/utils'

interface NavItem {
  to: string
  label: string
  icon: NavIconKey
  roles?: string[]
}

const navItems: NavItem[] = [
  { to: '/dashboard', label: 'Dashboard', icon: 'dashboard' },
  { to: '/profile', label: 'Profile', icon: 'profile' },
  { to: '/admin/users', label: 'Users', icon: 'users', roles: ['Admin'] },
  { to: '/admin/students', label: 'Students', icon: 'students', roles: ['Admin'] },
  { to: '/admin/teachers', label: 'Teachers', icon: 'teachers', roles: ['Admin'] },
  { to: '/admin/classes', label: 'Classes', icon: 'classes', roles: ['Admin'] },
  { to: '/admin/subjects', label: 'Subjects', icon: 'subjects', roles: ['Admin'] },
  { to: '/admin/enrollment', label: 'Enrollment', icon: 'enrollment', roles: ['Admin'] },
  { to: '/admin/attendance-summary', label: 'Attendance', icon: 'attendance', roles: ['Admin'] },
  { to: '/admin/finance', label: 'Finance', icon: 'finance', roles: ['Admin'] },
  { to: '/admin/notify', label: 'Broadcast', icon: 'notify', roles: ['Admin'] },
  { to: '/academics/attendance', label: 'Attendance', icon: 'attendanceCheck', roles: ['Teacher'] },
  { to: '/academics/grades', label: 'Grades', icon: 'grades', roles: ['Teacher'] },
  { to: '/academics/attendance', label: 'My Attendance', icon: 'attendanceCheck', roles: ['Student'] },
  { to: '/academics/grades', label: 'My Grades', icon: 'grades', roles: ['Student'] },
  { to: '/academics/reportcard', label: 'Report Card', icon: 'reportCard', roles: ['Student'] },
  { to: '/finance', label: 'Finance', icon: 'payments', roles: ['Student'] },
  { to: '/locker', label: 'Locker', icon: 'locker', roles: ['Student'] },
  { to: '/parent/children', label: 'Children', icon: 'children', roles: ['Parent'] },
  { to: '/finance', label: 'Payments', icon: 'payments', roles: ['Parent'] },
  { to: '/notifications', label: 'Notifications', icon: 'notifications' },
]

export function Sidebar() {
  const { role } = useRole()
  const visible = navItems.filter((item) => !item.roles || (role && item.roles.includes(role)))

  return (
    <aside className="w-60 shrink-0 border-r border-surface-border bg-surface/50 backdrop-blur-xl flex flex-col min-h-screen">
      <div className="px-5 py-6 border-b border-surface-border">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-accent/20 border border-accent/40 flex items-center justify-center">
            <span className="text-accent font-mono text-xs font-bold">S</span>
          </div>
          <div>
            <p className="text-sm font-semibold text-foreground tracking-wide">SMS</p>
            <p className="text-[10px] uppercase tracking-widest text-muted">Control</p>
          </div>
        </div>
      </div>

      <nav className="flex-1 px-2 py-4 space-y-0.5 overflow-y-auto">
        {visible.map((item) => {
          const Icon = navIcons[item.icon]
          return (
            <NavLink
              key={item.to + (item.roles?.join('') ?? '')}
              to={item.to}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all duration-200',
                  isActive
                    ? 'bg-accent/10 text-accent border border-accent/30 shadow-glow-sm'
                    : 'text-muted hover:text-foreground hover:bg-surface-elevated/60 border border-transparent'
                )
              }
            >
              <Icon className="w-4 h-4 shrink-0" strokeWidth={1.75} />
              <span>{item.label}</span>
            </NavLink>
          )
        })}
      </nav>
    </aside>
  )
}
