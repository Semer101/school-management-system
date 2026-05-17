import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { Inbox } from 'lucide-react'
import { cn } from '../../lib/utils'

interface EmptyStateProps {
  icon?: LucideIcon
  title: string
  description?: string
  action?: ReactNode
  className?: string
}

export function EmptyState({
  icon: Icon = Inbox,
  title,
  description,
  action,
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center py-16 gap-3 text-center',
        className
      )}
    >
      <div className="w-12 h-12 rounded-xl border border-surface-border bg-surface/60 flex items-center justify-center text-muted">
        <Icon className="w-6 h-6" strokeWidth={1.5} />
      </div>
      <h3 className="text-base font-semibold text-foreground">{title}</h3>
      {description && <p className="text-sm text-muted max-w-sm">{description}</p>}
      {action}
    </div>
  )
}
