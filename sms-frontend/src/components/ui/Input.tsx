import { type InputHTMLAttributes, forwardRef } from 'react'
import { cn } from '../../lib/utils'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, className = '', ...props }, ref) => (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label className="text-xs font-medium uppercase tracking-wider text-muted">{label}</label>
      )}
      <input
        ref={ref}
        className={cn(
          'w-full px-3 py-2.5 rounded-lg text-sm text-foreground',
          'bg-void/60 border border-surface-border backdrop-blur-sm',
          'placeholder:text-muted/60 transition-colors duration-200',
          'focus:outline-none focus:border-accent/60 focus:ring-1 focus:ring-accent/30',
          error && 'border-danger/60 focus:border-danger/60 focus:ring-danger/30',
          className
        )}
        {...props}
      />
      {error && <p className="text-xs text-danger">{error}</p>}
    </div>
  )
)

Input.displayName = 'Input'