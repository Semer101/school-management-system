import { useEffect, useState } from 'react'
import { listTrash, restoreTrash, permanentDelete, type TrashEntity } from '../../api/trash'
import { DataTable } from '../../components/ui/DataTable'
import { PageHeader } from '../../components/ui/PageHeader'
import { Button } from '../../components/ui/Button'
import { ConfirmModal } from '../../components/ui/ConfirmModal'
import { AlertModal } from '../../components/ui/AlertModal'

const entities: { id: TrashEntity; label: string }[] = [
  { id: 'students', label: 'Students' },
  { id: 'teachers', label: 'Teachers' },
  { id: 'classes', label: 'Classes' },
  { id: 'subjects', label: 'Subjects' },
  { id: 'users', label: 'Users' },
]

export default function TrashPage() {
  const [entity, setEntity] = useState<TrashEntity>('students')
  const [rows, setRows] = useState<Record<string, unknown>[]>([])
  const [loading, setLoading] = useState(true)
  const [deleteTarget, setDeleteTarget] = useState<{ id: number; label: string } | null>(null)
  const [password, setPassword] = useState('')
  const [deleting, setDeleting] = useState(false)
  const [alertState, setAlertState] = useState<{ open: boolean; title: string; message: string; type: 'success' | 'error' }>({
    open: false,
    title: '',
    message: '',
    type: 'success',
  })

  const fetchTrash = async () => {
    setLoading(true)
    try {
      const res = await listTrash(entity)
      const payload = res.data.data as { data?: unknown[] } | unknown[]
      const list = Array.isArray(payload) ? payload : (payload?.data ?? [])
      setRows(list as Record<string, unknown>[])
    } catch {
      setRows([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchTrash() }, [entity])

  const handleRestore = async (id: number) => {
    try {
      await restoreTrash(entity, id)
      fetchTrash()
      setAlertState({ open: true, title: 'Success', message: 'Record restored successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Restore Failed', message: 'Failed to restore archived record', type: 'error' })
    }
  }

  const handlePermanentDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await permanentDelete(entity, deleteTarget.id, password)
      setDeleteTarget(null)
      setPassword('')
      fetchTrash()
      setAlertState({ open: true, title: 'Success', message: 'Record permanently deleted', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Delete Failed', message: 'Delete failed — check your password', type: 'error' })
    } finally {
      setDeleting(false)
    }
  }

  const label = (row: Record<string, unknown>) =>
    String(row.student_code ?? row.teacher_code ?? row.name ?? row.email ?? row.id)

  return (
    <div>
      <PageHeader title="Trash" subtitle="Restore archived records or permanently delete" />

      <div className="flex gap-2 mb-6 flex-wrap">
        {entities.map((e) => (
          <button
            key={e.id}
            type="button"
            onClick={() => setEntity(e.id)}
            className={`px-3 py-1.5 rounded-lg text-sm border transition-colors ${
              entity === e.id
                ? 'border-accent bg-accent/10 text-accent'
                : 'border-surface-border text-muted hover:text-foreground'
            }`}
          >
            {e.label}
          </button>
        ))}
      </div>

      <DataTable
        loading={loading}
        data={rows}
        keyExtractor={(r) => Number(r.id)}
        searchKeys={['name', 'email', 'student_code', 'teacher_code', 'code']}
        columns={[
          { key: 'id', header: 'ID' },
          { key: 'label', header: 'Record', render: (r) => label(r) },
          {
            key: 'actions',
            header: 'Actions',
            render: (r) => (
              <div className="flex gap-2">
                <Button size="sm" variant="primary" onClick={() => handleRestore(Number(r.id))}>
                  Restore
                </Button>
                <Button
                  size="sm"
                  variant="danger"
                  onClick={() => setDeleteTarget({ id: Number(r.id), label: label(r) })}
                >
                  Delete Forever
                </Button>
              </div>
            ),
          },
        ]}
      />

      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => { setDeleteTarget(null); setPassword('') }}
        title="Permanent Delete"
        message={
          <p>
            This will permanently remove <strong>{deleteTarget?.label}</strong>. This action cannot be undone.
          </p>
        }
        confirmLabel="Delete permanently"
        variant="danger"
        loading={deleting}
        requirePassword
        password={password}
        onPasswordChange={setPassword}
        onConfirm={handlePermanentDelete}
      />

      <AlertModal
        open={alertState.open}
        onClose={() => setAlertState({ ...alertState, open: false })}
        title={alertState.title}
        message={alertState.message}
        type={alertState.type}
      />
    </div>
  )
}
