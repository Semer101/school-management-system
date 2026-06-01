import { useEffect, useState, type FormEvent } from 'react'
import { Plus } from 'lucide-react'
import { getTeachers, archiveTeacher, createTeacher, updateTeacher, type CreateTeacherPayload } from '../../api/admin'
import type { Teacher } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { DataTable } from '../../components/ui/DataTable'
import { RowActions } from '../../components/ui/RowActions'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import { Input } from '../../components/ui/Input'
import { ConfirmModal } from '../../components/ui/ConfirmModal'
import { PageHeader } from '../../components/ui/PageHeader'
import { AlertModal } from '../../components/ui/AlertModal'

export default function TeachersPage() {
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [updateTeacher_, setUpdateTeacher] = useState<Teacher | null>(null)
  const [archiveId, setArchiveId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState<CreateTeacherPayload>({
    name: '', email: '', password: 'Teacher@1234', teacher_code: '', qualification: '', phone: '',
  })
  const [editForm, setEditForm] = useState({ qualification: '', phone: '' })
  const [alertState, setAlertState] = useState<{ open: boolean; title: string; message: string; type: 'success' | 'error' }>({
    open: false,
    title: '',
    message: '',
    type: 'success',
  })

  const fetchTeachers = () => {
    setLoading(true)
    getTeachers({ page_size: 50 })
      .then((r) => setTeachers(listFromApi(r.data)))
      .catch(() => setTeachers([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchTeachers() }, [])

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      await createTeacher(form)
      setCreateOpen(false)
      fetchTeachers()
      setAlertState({ open: true, title: 'Success', message: 'Teacher profile created successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Create Failed', message: 'Failed to create teacher account', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleUpdate = async (e: FormEvent) => {
    e.preventDefault()
    if (!updateTeacher_) return
    setSaving(true)
    try {
      await updateTeacher(updateTeacher_.id, { qualification: editForm.qualification, phone: editForm.phone })
      setUpdateTeacher(null)
      fetchTeachers()
      setAlertState({ open: true, title: 'Success', message: 'Teacher profile updated successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Update Failed', message: 'Failed to update teacher details', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleArchive = async () => {
    if (!archiveId) return
    setSaving(true)
    try {
      await archiveTeacher(archiveId)
      setArchiveId(null)
      fetchTeachers()
      setAlertState({ open: true, title: 'Success', message: 'Teacher archived successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Archive Failed', message: 'Failed to archive teacher profile', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <PageHeader title="Teachers" action={
        <Button onClick={() => setCreateOpen(true)}><Plus className="w-4 h-4 mr-1" /> Create</Button>
      } />

      <DataTable
        loading={loading}
        data={teachers}
        keyExtractor={(t) => t.id}
        searchKeys={['teacher_code', 'qualification']}
        columns={[
          { key: 'teacher_code', header: 'Code' },
          { key: 'user', header: 'Name', render: (t) => t.user?.name ?? '—' },
          { key: 'email', header: 'Email', render: (t) => t.user?.email ?? '—' },
          { key: 'phone', header: 'Phone', render: (t) => t.user?.phone || '—' },
          { key: 'qualification', header: 'Qualification' },
          { key: 'joined_at', header: 'Joined', render: (t) => new Date(t.joined_at).toLocaleDateString() },
          {
            key: 'actions', header: '',
            render: (t) => (
              <RowActions
                onUpdate={() => { setUpdateTeacher(t); setEditForm({ qualification: t.qualification ?? '', phone: t.user?.phone ?? '' }) }}
                onArchive={() => setArchiveId(t.id)}
              />
            ),
          },
        ]}
      />

      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="Create Teacher"
        footer={<><Button variant="ghost" onClick={() => setCreateOpen(false)}>Cancel</Button><Button loading={saving} type="submit" form="teacher-form">Create</Button></>}>
        <form id="teacher-form" onSubmit={handleCreate} className="space-y-3">
          <Input label="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <Input label="Email" type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} required />
          <Input label="Password" type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required />
          <Input label="Teacher code" value={form.teacher_code} onChange={(e) => setForm({ ...form, teacher_code: e.target.value })} required />
          <Input label="Qualification" value={form.qualification} onChange={(e) => setForm({ ...form, qualification: e.target.value })} />
          <Input label="Phone" value={form.phone ?? ''} onChange={(e) => setForm({ ...form, phone: e.target.value })} placeholder="09xxxxxxxx" required />
        </form>
      </Modal>

      <Modal open={!!updateTeacher_} onClose={() => setUpdateTeacher(null)} title="Update Teacher"
        footer={<><Button variant="ghost" onClick={() => setUpdateTeacher(null)}>Cancel</Button><Button loading={saving} type="submit" form="teacher-update-form">Save</Button></>}>
        <form id="teacher-update-form" onSubmit={handleUpdate} className="space-y-3">
          <p className="text-sm text-muted">{updateTeacher_?.user?.name} — {updateTeacher_?.user?.email}</p>
          <Input label="Qualification" value={editForm.qualification} onChange={(e) => setEditForm({ ...editForm, qualification: e.target.value })} required />
          <Input label="Phone" value={editForm.phone} onChange={(e) => setEditForm({ ...editForm, phone: e.target.value })} placeholder="09xxxxxxxx" />
        </form>
      </Modal>

      <ConfirmModal open={!!archiveId} onClose={() => setArchiveId(null)} title="Archive Teacher"
        message="Teacher will be moved to Trash." confirmLabel="Archive" variant="danger" loading={saving} onConfirm={handleArchive} />

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