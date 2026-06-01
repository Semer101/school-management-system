import { Archive, Pencil } from 'lucide-react'
import { Button } from './Button'

interface RowActionsProps {
  onUpdate: () => void
  onArchive: () => void
  archiveTitle?: string
}

export function RowActions({ onUpdate, onArchive, archiveTitle = 'Archive' }: RowActionsProps) {
  return (
    <div className="flex gap-1">
      <Button size="sm" variant="ghost" onClick={onUpdate} title="Update">
        <Pencil className="w-3.5 h-3.5" />
      </Button>
      <Button size="sm" variant="danger" onClick={onArchive} title={archiveTitle}>
        <Archive className="w-3.5 h-3.5" />
      </Button>
    </div>
  )
}
