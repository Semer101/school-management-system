import type { ReactNode } from 'react'

interface EmptyStateProps {
  icon?: string
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ icon = '📭', title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-3 text-center">
      <span className="text-4xl">{icon}</span>
      <h3 className="text-base font-semibold text-[var(--text-h)]">{title}</h3>
      {description && (
        <p className="text-sm text-[var(--text)] max-w-sm">{description}</p>
      )}
      {action}
    </div>
  )
}