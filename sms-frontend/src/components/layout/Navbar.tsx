import { useNavigate } from 'react-router-dom'
import { Bell, LogOut } from 'lucide-react'
import { useAuth } from '../../hooks/useAuth'
import { useNotifications } from '../../hooks/useNotifications'
import { Badge, roleBadgeVariant } from '../ui/Badge'
import { Button } from '../ui/Button'
import { cn } from '../../lib/utils'

export function Navbar() {
  const { user, logout } = useAuth()
  const { unreadCount } = useNotifications()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  return (
    <header className="h-14 shrink-0 border-b border-surface-border bg-surface/40 backdrop-blur-xl flex items-center justify-between px-6">
      <div className="text-xs font-mono text-muted uppercase tracking-widest hidden sm:block">
        Session active
      </div>

      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={() => navigate('/notifications')}
          className={cn(
            'relative p-2 rounded-lg border border-surface-border text-muted',
            'hover:text-accent hover:border-accent/40 transition-colors'
          )}
          aria-label="Notifications"
        >
          <Bell className="w-4 h-4" />
          {unreadCount > 0 && (
            <span className="absolute -top-1 -right-1 min-w-[18px] h-[18px] px-1 flex items-center justify-center rounded-full bg-accent text-void text-[10px] font-semibold">
              {unreadCount > 99 ? '99+' : unreadCount}
            </span>
          )}
        </button>

        <div className="flex items-center gap-2 pl-2 border-l border-surface-border">
          <div className="hidden sm:block text-right">
            <p className="text-sm font-medium text-foreground leading-tight">{user?.name}</p>
            <p className="text-[10px] text-muted">{user?.email}</p>
          </div>
          {user?.role && <Badge label={user.role} variant={roleBadgeVariant(user.role)} />}
          <Button variant="ghost" size="sm" onClick={handleLogout} aria-label="Sign out">
            <LogOut className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </header>
  )
}
