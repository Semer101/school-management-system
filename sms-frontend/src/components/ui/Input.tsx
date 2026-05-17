import { type InputHTMLAttributes, forwardRef } from 'react'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, className = '', ...props }, ref) => {
    return (
      <div className="flex flex-col gap-1.5">
        {label && (
          <label className="text-sm font-medium text-[var(--text-h)]">
            {label}
          </label>
        )}
        <input
          ref={ref}
          className={`
            w-full px-3 py-2 rounded-lg text-sm
            bg-[var(--bg)] border text-[var(--text-h)]
            placeholder:text-[var(--text)]
            transition-colors duration-150
            outline-none
            ${error
              ? 'border-red-500 focus:border-red-500'
              : 'border-[var(--border)] focus:border-[var(--accent)]'
            }
            ${className}
          `}
          {...props}
        />
        {error && <p className="text-xs text-red-500">{error}</p>}
      </div>
    )
  }
)

Input.displayName = 'Input'