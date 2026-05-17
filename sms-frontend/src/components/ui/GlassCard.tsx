import { motion } from 'framer-motion'
import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

interface GlassCardProps {
  children: ReactNode
  className?: string
  hover?: boolean
}

export function GlassCard({ children, className, hover }: GlassCardProps) {
  return (
    <motion.div
      initial={false}
      whileHover={hover ? { y: -2 } : undefined}
      className={cn(
        'rounded-xl border border-surface-border bg-surface/80 backdrop-blur-xl shadow-glass',
        hover && 'transition-colors duration-300 hover:border-accent/40 hover:shadow-glow-sm',
        className
      )}
    >
      {children}
    </motion.div>
  )
}
