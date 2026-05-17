interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  fullPage?: boolean
}

const sizes = { sm: 'w-4 h-4', md: 'w-6 h-6', lg: 'w-10 h-10' }

export function Spinner({ size = 'md', fullPage }: SpinnerProps) {
  const spinner = (
    <span
      className={`${sizes[size]} border-2 border-[var(--border)] border-t-[var(--accent)] rounded-full animate-spin inline-block`}
    />
  )
  if (fullPage) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        {spinner}
      </div>
    )
  }
  return spinner
}