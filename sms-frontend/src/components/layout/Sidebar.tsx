import { NavLink } from 'react-router-dom'
import { BarChart3, Trash2, ChevronLeft, ChevronRight } from 'lucide-react'
import { useRole } from '../../hooks/useRole'
import { navIcons, type NavIconKey } from '../../lib/nav-icons'
import { cn } from '../../lib/utils'
import { useAppSelector, useAppDispatch } from '../../store/hooks'
import { toggleSidebar, setSidebarCollapsed } from '../../store/uiSlice'

interface NavItem {
  to: string
  label: string
  icon: NavIconKey | 'analytics' | 'trash'
  roles?: string[]
}

const navItems: NavItem[] = [
  { to: '/dashboard', label: 'Dashboard', icon: 'dashboard' },
  { to: '/profile', label: 'Profile', icon: 'profile' },
  { to: '/admin/admins', label: 'Admins', icon: 'users', roles: ['Admin'] },
  { to: '/admin/parents', label: 'Parents', icon: 'users', roles: ['Admin'] },
  { to: '/admin/students', label: 'Students', icon: 'students', roles: ['Admin'] },
  { to: '/admin/promotion', label: 'Promotion', icon: 'promotion', roles: ['Admin'] },
  { to: '/admin/teachers', label: 'Teachers', icon: 'teachers', roles: ['Admin'] },
  { to: '/admin/classes', label: 'Classes', icon: 'classes', roles: ['Admin'] },
  { to: '/admin/subjects', label: 'Subjects', icon: 'subjects', roles: ['Admin'] },
  { to: '/admin/analytics', label: 'Analytics', icon: 'analytics', roles: ['Admin'] },
  { to: '/admin/attendance-summary', label: 'Attendance', icon: 'attendance', roles: ['Admin'] },
  { to: '/admin/finance', label: 'Finance', icon: 'finance', roles: ['Admin'] },
  { to: '/admin/notify', label: 'Broadcast', icon: 'notify', roles: ['Admin'] },
  { to: '/admin/trash', label: 'Trash', icon: 'trash', roles: ['Admin'] },
  { to: '/academics/attendance', label: 'Attendance', icon: 'attendanceCheck', roles: ['Teacher'] },
  { to: '/academics/grades', label: 'Grades', icon: 'grades', roles: ['Teacher'] },
  { to: '/academics/attendance', label: 'My Attendance', icon: 'attendanceCheck', roles: ['Student'] },
  { to: '/academics/grades', label: 'My Grades', icon: 'grades', roles: ['Student'] },
  { to: '/locker', label: 'Locker', icon: 'locker', roles: ['Student'] },
  { to: '/parent/children', label: 'Children', icon: 'children', roles: ['Parent'] },
  { to: '/finance', label: 'Payments', icon: 'payments', roles: ['Parent'] },
  { to: '/id-card', label: 'ID Card', icon: 'profile', roles: ['Admin'] },
  { to: '/notifications', label: 'Notifications', icon: 'notifications' },
]

export function Sidebar() {
  const { role } = useRole()
  const sidebarCollapsed = useAppSelector((state) => state.ui.sidebarCollapsed)
  const dispatch = useAppDispatch()
  
  const visible = navItems.filter((item) => !item.roles || (role && item.roles.includes(role)))

  const handleLinkClick = () => {
    // Auto collapse on mobile/tablet viewports
    if (window.innerWidth < 1024) {
      dispatch(setSidebarCollapsed(true))
    }
  }

  return (
    <aside
      className={cn(
        'fixed lg:sticky top-0 bottom-0 left-0 shrink-0 border-r border-surface-border bg-surface/90 lg:bg-surface/80 backdrop-blur-xl flex flex-col min-h-screen transition-all duration-300 ease-in-out z-50',
        sidebarCollapsed ? 'w-60 lg:w-16 -translate-x-full lg:translate-x-0' : 'w-60 translate-x-0'
      )}
    >
      {/* Sidebar Header / Logo */}
      <div className="px-4 py-5 border-b border-surface-border flex items-center justify-between min-h-[60px]">
        <div className={cn('flex items-center gap-2', sidebarCollapsed && 'lg:justify-center lg:w-full')}>
          <div className="w-8 h-8 shrink-0 rounded-lg bg-accent/20 border border-accent/40 flex items-center justify-center">
            <span className="text-accent font-mono text-xs font-bold">S</span>
          </div>
          <div className={cn('transition-all duration-300', sidebarCollapsed ? 'lg:opacity-0 lg:w-0 overflow-hidden' : 'opacity-100 w-auto')}>
            <p className="text-sm font-semibold text-foreground tracking-wide whitespace-nowrap">SMS</p>
            <p className="text-[10px] uppercase tracking-widest text-muted whitespace-nowrap">Ethiopia G9–12</p>
          </div>
        </div>
      </div>

      {/* Nav items */}
      <nav className="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
        {visible.map((item) => {
          const Icon = item.icon === 'analytics' ? BarChart3 : item.icon === 'trash' ? Trash2 : navIcons[item.icon as NavIconKey]
          return (
            <NavLink
              key={item.to + (item.roles?.join('') ?? '')}
              to={item.to}
              onClick={handleLinkClick}
              title={sidebarCollapsed ? item.label : undefined}
              className={({ isActive }) =>
                cn(
                  'flex items-center rounded-lg text-sm transition-all duration-200 border border-transparent',
                  sidebarCollapsed 
                    ? 'lg:justify-center lg:px-2 py-2.5' 
                    : 'px-3 py-2.5 gap-3',
                  isActive
                    ? 'bg-accent/10 text-accent border-accent/30 shadow-glow-sm'
                    : 'text-muted hover:text-foreground hover:bg-surface-elevated/60'
                )
              }
            >
              <Icon className="w-4 h-4 shrink-0" strokeWidth={1.75} />
              <span className={cn('transition-all duration-300 whitespace-nowrap', sidebarCollapsed ? 'lg:hidden lg:w-0' : 'opacity-100 w-auto')}>
                {item.label}
              </span>
            </NavLink>
          )
        })}
      </nav>

      {/* Desktop Collapse Toggle Handle */}
      <div className="hidden lg:flex p-3 border-t border-surface-border justify-center">
        <button
          type="button"
          onClick={() => dispatch(toggleSidebar())}
          className="p-1.5 rounded-lg border border-surface-border text-muted hover:text-foreground hover:bg-surface-elevated/60 transition-colors"
          aria-label={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {sidebarCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
        </button>
      </div>
    </aside>
  )
}
