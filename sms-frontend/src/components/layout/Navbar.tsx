import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { useNotifications } from '../../hooks/useNotifications'
import { Badge, roleBadgeVariant } from '../ui/Badge'

export function Navbar() {
  const { user, logout } = useAuth()
  const { unreadCount } = useNotifications()
  const navigate = useNavigate()
  const [menuOpen, setMenuOpen] = useState(false)

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  return (
    <header className="h-14 border-b border-[var(--border)] bg-[var(--bg)] flex items-center justify-between px-5 shrink-0">
      {/* Page title slot — empty here, pages can use a portal if needed */}
      <div />

      <div className="flex items-center gap-3">
        {/* Notification bell */}
        <button
          onClick={() => navigate('/notifications')}
          className="relative w-9 h-9 flex items-center justify-center rounded-lg text-[var(--text)] hover:bg-[var(--code-bg)] transition-colors"
        >
          🔔
          {unreadCount > 0 && (
            <span className="absolute top-1 right-1 min-w-[16px] h-4 flex items-center justify-center bg-[var(--accent)] text-white text-[10px] font-bold rounded-full px-1">
              {unreadCount > 99 ? '99+' : unreadCount}
            </span>
          )}
        </button>

        {/* User menu */}
        <div className="relative">
          <button
            onClick={() => setMenuOpen(!menuOpen)}
            className="flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-[var(--code-bg)] transition-colors"
          >
            <div className="w-7 h-7 rounded-full bg-[var(--accent-bg)] flex items-center justify-center text-[var(--accent)] text-xs font-bold">
              {user?.name?.[0]?.toUpperCase() ?? '?'}
            </div>
            <div className="text-left hidden sm:block">
              <p className="text-xs font-medium text-[var(--text-h)] leading-tight">{user?.name}</p>
              <p className="text-[10px] text-[var(--text)] leading-tight">{user?.email}</p>
            </div>
            {user?.role && (
              <Badge label={user.role} variant={roleBadgeVariant(user.role)} />
            )}
          </button>

          {menuOpen && (
            <>
              <div className="fixed inset-0 z-10" onClick={() => setMenuOpen(false)} />
              <div className="absolute right-0 top-full mt-1 w-44 bg-[var(--bg)] border border-[var(--border)] rounded-xl shadow-lg z-20 overflow-hidden">
                <button
                  onClick={() => { setMenuOpen(false); navigate('/profile') }}
                  className="w-full text-left px-4 py-2.5 text-sm text-[var(--text)] hover:bg-[var(--code-bg)] transition-colors"
                >
                  👤 Profile
                </button>
                <button
                  onClick={handleLogout}
                  className="w-full text-left px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/10 transition-colors"
                >
                  🚪 Logout
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </header>
  )
}