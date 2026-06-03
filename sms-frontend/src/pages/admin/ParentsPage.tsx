import { useEffect, useState, type FormEvent } from 'react'
import { Plus } from 'lucide-react'
import { getParents, registerUser, archiveParent, updateParent, getStudents, type ParentRow } from '../../api/admin'
import type { Student } from '../../types/academic'
import { listFromApi } from '../../types/api'
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
  const [students, setStudents] = useState<Student[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [updateRow, setUpdateRow] = useState<ParentRow | null>(null)
  const [archiveId, setArchiveId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({ name: '', email: '', password: 'Parent@1234', phone: '', student_id: 0 })
  const [editForm, setEditForm] = useState({ name: '', email: '', phone: '' })
  const [alertState, setAlertState] = useState<{ open: boolean; title: string; message: string; type: 'success' | 'error' }>({
    open: false,
    title: '',
    message: '',
    type: 'success',
  })

  const fetchData = () => {
    setLoading(true)
    Promise.all([
      getParents(),
      getStudents({ page_size: 200 })
    ])
      .then(([p, s]) => {
        const payload = p.data.data as { data?: ParentRow[] } | ParentRow[]
        setParents(Array.isArray(payload) ? payload : (payload?.data ?? []))
        setStudents(listFromApi(s.data))
      })
      .catch(() => setParents([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    async function load() {
      setLoading(true)
      try {
        const [p, s] = await Promise.all([getParents(), getStudents({ page_size: 200 })])
        const payload = p.data.data as { data?: ParentRow[] } | ParentRow[]
        setParents(Array.isArray(payload) ? payload : (payload?.data ?? []))
        setStudents(listFromApi(s.data))
      } catch {
        setParents([])
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      const res = await registerUser({ ...form, role: 'Parent' as Role })
      // If a student was selected, update the student's parent_id
      if (form.student_id && res?.data?.data) {
        const parentUserId = (res.data.data as { id: number }).id
        await import('../../api/admin').then(mod => {
          mod.updateStudent(form.student_id, { parent_id: parentUserId } as Partial<Student>)
        })
      }
      setCreateOpen(false)
      setForm({ name: '', email: '', password: 'Parent@1234', phone: '', student_id: 0 })
      fetchData()
      setAlertState({ open: true, title: 'Success', message: 'Parent account created and linked to student successfully', type: 'success' })
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
      fetchData()
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
      fetchData()
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
        searchKeys={['name', 'email', 'student_names']}
        columns={[
          { key: 'name', header: 'Parent Name' },
          { key: 'student_names', header: 'Student Name', render: (p) => p.student_names || '—' },
          { key: 'email', header: 'Email' },
          { key: 'phone', header: 'Phone', render: (p) => p.phone || '—' },
          {
            key: 'children_count',
            header: 'Children',
            render: (p) => (
              <span className="relative group cursor-pointer underline decoration-dotted font-medium text-accent hover:text-accent-hover">
                {p.children_count}
                {p.children && p.children.length > 0 && (
                  <div className="absolute z-50 hidden group-hover:block bg-surface-elevated border border-surface-border text-foreground text-xs rounded-lg shadow-xl p-3 min-w-[260px] left-1/2 -translate-x-1/2 mt-1">
                    <div className="font-semibold border-b border-surface-border pb-1.5 mb-1.5 text-accent text-left">
                      Children List
                    </div>
                    <div className="space-y-2">
                      {p.children.map((c) => (
                        <div key={c.id} className="text-left">
                          <div className="font-medium text-foreground">{c.name}</div>
                          <div className="text-muted text-[10px] flex justify-between gap-2 mt-0.5 font-mono">
                            <span>ID: {c.student_code}</span>
                            <span>{c.grade} {c.section ? `(${c.section})` : ''}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </span>
            ),
          },
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
          <label className="text-sm text-muted">Link to existing student
            <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.student_id}
              onChange={(e) => setForm({ ...form, student_id: Number(e.target.value) })}>
              <option value={0}>— No link (create parent only) —</option>
              {students.filter(s => !s.parent_id || s.parent_id === 0).map((s) => (
                <option key={s.id} value={s.id}>{s.user?.name ?? `Student #${s.id}`} ({s.student_code})</option>
              ))}
            </select>
            <p className="text-xs text-muted mt-1">Optionally link this parent to an existing student. Only students without a parent are shown.</p>
          </label>
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