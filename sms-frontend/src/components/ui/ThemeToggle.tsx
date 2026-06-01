import { Moon, Sun } from 'lucide-react'
import { useAppDispatch, useAppSelector } from '../../store/hooks'
import { toggleTheme } from '../../store/uiSlice'
import { cn } from '../../lib/utils'

export function ThemeToggle({ className }: { className?: string }) {
  const theme = useAppSelector((s) => s.ui.theme)
  const dispatch = useAppDispatch()

  return (
    <button
      type="button"
      onClick={() => dispatch(toggleTheme())}
      className={cn(
        'p-2 rounded-lg border border-surface-border text-muted',
        'hover:text-accent hover:border-accent/40 transition-colors',
        className
      )}
      aria-label={theme === 'light' ? 'Switch to dark mode' : 'Switch to light mode'}
    >
      {theme === 'light' ? <Moon className="w-4 h-4" /> : <Sun className="w-4 h-4" />}
    </button>
  )
}
