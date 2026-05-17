import { type ReactNode, useEffect } from 'react'
import { X } from 'lucide-react'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
  footer?: ReactNode
}

export function Modal({ open, onClose, title, children, footer }: ModalProps) {
  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, onClose])

  if (!open) return null

  const overlay = 'fixed inset-0 z-50 flex items-center justify-center p-4'

  return (
    <div className={overlay} onClick={onClose} role="dialog" aria-modal="true" aria-labelledby="modal-title">
      <div className="absolute inset-0 bg-void/80 backdrop-blur-sm" aria-hidden />
      <div
        className="relative w-full max-w-md bg-surface border border-surface-border rounded-2xl shadow-glass"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-6 py-4 border-b border-surface-border">
          <h2 id="modal-title" className="text-base font-semibold text-foreground">
            {title}
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="w-7 h-7 flex items-center justify-center rounded-lg text-muted hover:text-foreground hover:bg-surface-elevated transition-colors"
            aria-label="Close"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="px-6 py-4">{children}</div>
        {footer ? (
          <div className="px-6 py-4 border-t border-surface-border flex justify-end gap-2">{footer}</div>
        ) : null}
      </div>
    </div>
  )
}
