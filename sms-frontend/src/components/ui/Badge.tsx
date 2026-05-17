interface BadgeProps {
  label: string
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'accent'
}

const variants = {
  default: 'bg-[var(--code-bg)] text-[var(--text)]',
  success: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
  warning: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
  danger: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  accent: 'bg-[var(--accent-bg)] text-[var(--accent)]',
}

export function Badge({ label, variant = 'default' }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${variants[variant]}`}
    >
      {label}
    </span>
  )
}

export function statusVariant(
  status: string
): BadgeProps['variant'] {
  switch (status) {
    case 'Verified':
    case 'Paid':
    case 'Present':
      return 'success'
    case 'Pending':
    case 'Late':
      return 'warning'
    case 'Rejected':
    case 'Absent':
      return 'danger'
    default:
      return 'default'
  }
}

export function roleBadgeVariant(role: string): BadgeProps['variant'] {
  switch (role) {
    case 'Admin': return 'danger'
    case 'Teacher': return 'accent'
    case 'Student': return 'success'
    case 'Parent': return 'warning'
    default: return 'default'
  }
}