import { cn } from '../../lib/utils'

interface BadgeProps {
  label: string
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'accent'
}

const variants = {
  default: 'bg-surface-elevated text-muted border border-surface-border',
  success: 'bg-success/15 text-success border border-success/30',
  warning: 'bg-warning/15 text-warning border border-warning/30',
  danger: 'bg-danger/15 text-danger border border-danger/30',
  accent: 'bg-accent/15 text-accent border border-accent/30',
}

export function Badge({ label, variant = 'default' }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium font-mono uppercase tracking-wide',
        variants[variant]
      )}
    >
      {label}
    </span>
  )
}

export function statusVariant(status: string): BadgeProps['variant'] {
  switch (status) {
    case 'Verified':
    case 'Paid':
    case 'Present':
    case 'Active':
      return 'success'
    case 'Pending':
    case 'Late':
      return 'warning'
    case 'Rejected':
    case 'Absent':
    case 'Inactive':
      return 'danger'
    default:
      return 'default'
  }
}

export function roleBadgeVariant(role: string): BadgeProps['variant'] {
  switch (role) {
    case 'Admin':
      return 'danger'
    case 'Teacher':
      return 'accent'
    case 'Student':
      return 'success'
    case 'Parent':
      return 'warning'
    default:
      return 'default'
  }
}
