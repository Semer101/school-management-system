import { useEffect, useState, type FormEvent } from 'react'
import { Plus } from 'lucide-react'
import { getAdmins, registerUser, updateAdmin, archiveAdmin } from '../../api/admin'
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
import type { AdminRow } from '../../api/admin'

export default function AdminsPage() {
  const [admins, setAdmins] = useState<AdminRow[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [updateRow, setUpdateRow] = useState<AdminRow | null>(null)
  const [archiveId, setArchiveId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({ name: '', email: '', password: 'Admin@1234', phone: '' })
  const [editForm, setEditForm] = useState({ name: '', email: '', phone: '' })
  const [alertState, setAlertState] = useState<{ open: boolean; title: string; message: string; type: 'success' | 'error' }>({
    open: false,
    title: '',
    message: '',
    type: 'success',
  })

  const fetchAdmins = () => {
    setLoading(true)
    getAdmins()
      .then((res) => {
        const payload = res.data.data as { data?: AdminRow[] } | AdminRow[]
        setAdmins(Array.isArray(payload) ? payload : (payload?.data ?? []))
      })
      .catch(() => setAdmins([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchAdmins() }, [])

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      await registerUser({ ...form, role: 'Admin' as Role })
      setCreateOpen(false)
      setForm({ name: '', email: '', password: 'Admin@1234', phone: '' })
      fetchAdmins()
      setAlertState({ open: true, title: 'Success', message: 'Administrator created successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Create Failed', message: 'Email may already exist or is invalid', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleUpdate = async (e: FormEvent) => {
    e.preventDefault()
    if (!updateRow) return
    setSaving(true)
    try {
      await updateAdmin(updateRow.id, editForm)
      setUpdateRow(null)
      fetchAdmins()
      setAlertState({ open: true, title: 'Success', message: 'Administrator updated successfully', type: 'success' })
    } catch (err: any) {
      const errMsg = err?.response?.data?.error || 'Failed to update administrator details'
      setAlertState({ open: true, title: 'Update Failed', message: errMsg, type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleArchive = async () => {
    if (!archiveId) return
    setSaving(true)
    try {
      await archiveAdmin(archiveId)
      setArchiveId(null)
      fetchAdmins()
      setAlertState({ open: true, title: 'Success', message: 'Administrator archived successfully', type: 'success' })
    } catch (err: any) {
      const errMsg = err?.response?.data?.error || 'Failed to archive administrator account'
      setAlertState({ open: true, title: 'Archive Failed', message: errMsg, type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const openUpdate = (a: AdminRow) => {
    setUpdateRow(a)
    setEditForm({ name: a.name, email: a.email, phone: a.phone ?? '' })
  }

  return (
    <div>
      <PageHeader
        title="Administrators"
        subtitle="System admin accounts — use sparingly"
        action={<Button onClick={() => setCreateOpen(true)}><Plus className="w-4 h-4 mr-1" /> Create</Button>}
      />

      <DataTable
        loading={loading}
        data={admins}
        keyExtractor={(a) => a.id}
        searchKeys={['name', 'email']}
        columns={[
          { key: 'name', header: 'Name' },
          { key: 'email', header: 'Email' },
          { key: 'phone', header: 'Phone', render: (a) => a.phone || <span className="text-warning text-xs">Not set</span> },
          {
            key: 'status', header: 'Status',
            render: (a) => <Badge label={a.status} variant={a.is_active ? 'success' : 'danger'} />,
          },
          {
            key: 'actions', header: '',
            render: (a) => (
              <RowActions onUpdate={() => openUpdate(a)} onArchive={() => setArchiveId(a.id)} />
            ),
          },
        ]}
      />

      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="Create Administrator"
        footer={<><Button variant="ghost" onClick={() => setCreateOpen(false)}>Cancel</Button><Button loading={saving} type="submit" form="admin-form">Create</Button></>}>
        <form id="admin-form" onSubmit={handleCreate} className="space-y-3">
          <Input label="Full name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <Input label="Email" type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} required />
          <Input label="Password" type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required />
          <Input label="Phone" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} placeholder="09xxxxxxxx" required />
        </form>
      </Modal>

      <Modal open={!!updateRow} onClose={() => setUpdateRow(null)} title="Update Administrator"
        footer={<><Button variant="ghost" onClick={() => setUpdateRow(null)}>Cancel</Button><Button loading={saving} type="submit" form="admin-update-form">Save</Button></>}>
        <form id="admin-update-form" onSubmit={handleUpdate} className="space-y-3">
          <Input label="Full name" value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} required />
          <Input label="Email" type="email" value={editForm.email} onChange={(e) => setEditForm({ ...editForm, email: e.target.value })} required />
          <Input label="Phone" value={editForm.phone} onChange={(e) => setEditForm({ ...editForm, phone: e.target.value })} placeholder="09xxxxxxxx" required />
        </form>
      </Modal>

      <ConfirmModal open={!!archiveId} onClose={() => setArchiveId(null)} title="Archive Administrator"
        message="This admin will be deactivated and moved to Trash." confirmLabel="Archive" variant="danger" loading={saving} onConfirm={handleArchive} />

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