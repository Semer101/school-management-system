import { type ReactNode } from 'react'
import { Modal } from './Modal'
import { Button } from './Button'
import { Input } from './Input'

interface ConfirmModalProps {
  open: boolean
  onClose: () => void
  title: string
  message: ReactNode
  confirmLabel?: string
  variant?: 'danger' | 'primary'
  loading?: boolean
  onConfirm: () => void
  requirePassword?: boolean
  password?: string
  onPasswordChange?: (v: string) => void
}

export function ConfirmModal({
  open,
  onClose,
  title,
  message,
  confirmLabel = 'Confirm',
  variant = 'primary',
  loading,
  onConfirm,
  requirePassword,
  password = '',
  onPasswordChange,
}: ConfirmModalProps) {
  const canConfirm = !requirePassword || password.length >= 1

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={title}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button
            variant={variant === 'danger' ? 'danger' : 'primary'}
            loading={loading}
            disabled={!canConfirm}
            onClick={onConfirm}
          >
            {confirmLabel}
          </Button>
        </>
      }
    >
      <div className="text-sm text-muted space-y-4">{message}</div>
      {requirePassword && onPasswordChange && (
        <Input
          label="Enter your password to confirm"
          type="password"
          value={password}
          onChange={(e) => onPasswordChange(e.target.value)}
          required
          autoComplete="current-password"
        />
      )}
    </Modal>
  )
}
