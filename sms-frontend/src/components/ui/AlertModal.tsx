import { Modal } from './Modal'
import { Button } from './Button'
import { CheckCircle2, AlertTriangle } from 'lucide-react'

interface AlertModalProps {
  open: boolean
  onClose: () => void
  title: string
  message: string
  type?: 'success' | 'error'
}

export function AlertModal({ open, onClose, title, message, type = 'success' }: AlertModalProps) {
  return (
    <Modal
      open={open}
      onClose={onClose}
      title={title}
      footer={
        <Button variant="primary" onClick={onClose}>
          OK
        </Button>
      }
    >
      <div className="flex items-start gap-4">
        {type === 'success' ? (
          <CheckCircle2 className="w-8 h-8 text-success shrink-0" />
        ) : (
          <AlertTriangle className="w-8 h-8 text-danger shrink-0" />
        )}
        <div className="text-sm text-foreground leading-relaxed">{message}</div>
      </div>
    </Modal>
  )
}
