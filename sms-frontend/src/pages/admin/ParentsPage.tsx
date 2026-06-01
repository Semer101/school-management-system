import { useEffect, useState, type FormEvent } from 'react'
import { Plus } from 'lucide-react'
import { getParents, registerUser, archiveParent, updateParent, type ParentRow } from '../../api/admin'
import { DataTable } from '../../components/ui/DataTable'
import { RowActions } from '../../components/ui/RowActions'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import { Input } from '../../components/ui/Input'
import { ConfirmModal } from '../../components/ui/ConfirmModal'
import { PageHeader } from '../../components/ui/PageHeader'
import { Badge } from '../../components/ui/Badge'
import { AlertModal } from '../../components/ui/AlertModal'
import type { Role } from '../../types/user'

export default function ParentsPage() {
  const [parents, setParents] = useState<ParentRow[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [updateRow, setUpdateRow] = useState<ParentRow | null>(null)
  const [archiveId, setArchiveId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({ name: '', email: '', password: 'Parent@1234', phone: '' })
  const [editForm, setEditForm] = useState({ name: '', email: '', phone: '' })
  const [alertState, setAlertState] = useState<{ open: boolean; title: string; message: string; type: 'success' | 'error' }>({
    open: false,
    title: '',
    message: '',
    type: 'success',
  })

  const fetchParents = () => {
    setLoading(true)
    getParents()
      .then((res) => {
        const payload = res.data.data as { data?: ParentRow[] } | ParentRow[]
        setParents(Array.isArray(payload) ? payload : (payload?.data ?? []))
      })
      .catch(() => setParents([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchParents() }, [])

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      await registerUser({ ...form, role: 'Parent' as Role })
      setCreateOpen(false)
      setForm({ name: '', email: '', password: 'Parent@1234', phone: '' })
      fetchParents()
      setAlertState({ open: true, title: 'Success', message: 'Parent account created successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Create Failed', message: 'Failed to create parent account', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleUpdate = async (e: FormEvent) => {
    e.preventDefault()
    if (!updateRow) return
    setSaving(true)
    try {
      await updateParent(updateRow.id, editForm)
      setUpdateRow(null)
      fetchParents()
      setAlertState({ open: true, title: 'Success', message: 'Parent account updated successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Update Failed', message: 'Failed to update parent details', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleArchive = async () => {
    if (!archiveId) return
    setSaving(true)
    try {
      await archiveParent(archiveId)
      setArchiveId(null)
      fetchParents()
      setAlertState({ open: true, title: 'Success', message: 'Parent account archived successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Archive Failed', message: 'Failed to archive parent account', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <PageHeader title="Parents" subtitle="Guardian accounts linked to students"
        action={<Button onClick={() => setCreateOpen(true)}><Plus className="w-4 h-4 mr-1" /> Create</Button>} />

      <DataTable
        loading={loading}
        data={parents}
        keyExtractor={(p) => p.id}
        searchKeys={['name', 'email']}
        columns={[
          { key: 'name', header: 'Name' },
          { key: 'email', header: 'Email' },
          { key: 'phone', header: 'Phone', render: (p) => p.phone || '—' },
          { key: 'children_count', header: 'Children' },
          {
            key: 'status', header: 'Status',
            render: (p) => <Badge label={p.status} variant={p.is_active ? 'success' : 'danger'} />,
          },
          {
            key: 'actions', header: '',
            render: (p) => (
              <RowActions onUpdate={() => { setUpdateRow(p); setEditForm({ name: p.name, email: p.email, phone: p.phone ?? '' }) }} onArchive={() => setArchiveId(p.id)} />
            ),
          },
        ]}
      />

      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="Create Parent"
        footer={<><Button variant="ghost" onClick={() => setCreateOpen(false)}>Cancel</Button><Button loading={saving} type="submit" form="parent-form">Create</Button></>}>
        <form id="parent-form" onSubmit={handleCreate} className="space-y-3">
          <Input label="Full name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <Input label="Email" type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} required />
          <Input label="Password" type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required />
          <Input label="Phone" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} placeholder="09xxxxxxxx" required />
        </form>
      </Modal>

      <Modal open={!!updateRow} onClose={() => setUpdateRow(null)} title="Update Parent"
        footer={<><Button variant="ghost" onClick={() => setUpdateRow(null)}>Cancel</Button><Button loading={saving} type="submit" form="parent-update-form">Save</Button></>}>
        <form id="parent-update-form" onSubmit={handleUpdate} className="space-y-3">
          <Input label="Full name" value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} required />
          <Input label="Email" type="email" value={editForm.email} onChange={(e) => setEditForm({ ...editForm, email: e.target.value })} required />
          <Input label="Phone" value={editForm.phone} onChange={(e) => setEditForm({ ...editForm, phone: e.target.value })} placeholder="09xxxxxxxx" required />
        </form>
      </Modal>

      <ConfirmModal open={!!archiveId} onClose={() => setArchiveId(null)} title="Archive Parent"
        message="Parent will be deactivated and moved to Trash." confirmLabel="Archive" variant="danger" loading={saving} onConfirm={handleArchive} />

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