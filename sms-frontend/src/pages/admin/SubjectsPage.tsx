import { useEffect, useState, type FormEvent } from 'react'
import { Plus } from 'lucide-react'
import { RowActions } from '../../components/ui/RowActions'
import { getSubjects, createSubject, archiveSubject, updateSubject, getTeachers } from '../../api/admin'
import type { Subject, Teacher } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { DataTable } from '../../components/ui/DataTable'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Modal } from '../../components/ui/Modal'
import { ConfirmModal } from '../../components/ui/ConfirmModal'
import { PageHeader } from '../../components/ui/PageHeader'
import { Badge } from '../../components/ui/Badge'

export default function SubjectsPage() {
  const [subjects, setSubjects] = useState<Subject[]>([])
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [updateSubject_, setUpdateSubject] = useState<Subject | null>(null)
  const [editForm, setEditForm] = useState({ name: '', code: '', teacher_id: 0 })
  const [archiveId, setArchiveId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({
    name: '', code: '', grade_level: 9, stream: '', teacher_id: 0, status: 'Active',
  })
  const [apiStream, setApiStream] = useState<string | undefined>(undefined)

  const fetchAll = () => {
    setLoading(true)
    const subParams: { page_size: number; stream?: string } = { page_size: 50 }
    if (apiStream) subParams.stream = apiStream
    Promise.all([getSubjects(subParams), getTeachers({ page_size: 50 })])
      .then(([s, t]) => {
        setSubjects(listFromApi(s.data))
        setTeachers(listFromApi(t.data))
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchAll() }, [apiStream])

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      await createSubject(form)
      setModalOpen(false)
      setForm({ name: '', code: '', grade_level: 9, stream: '', teacher_id: 0, status: 'Active' })
      fetchAll()
    } catch {
      alert('Create failed — check unique code')
    } finally {
      setSaving(false)
    }
  }

  const handleUpdate = async (e: FormEvent) => {
    e.preventDefault()
    if (!updateSubject_) return
    setSaving(true)
    try {
      await updateSubject(updateSubject_.id, editForm)
      setUpdateSubject(null)
      fetchAll()
    } catch {
      alert('Update failed')
    } finally {
      setSaving(false)
    }
  }

  const handleArchive = async () => {
    if (!archiveId) return
    setSaving(true)
    try {
      await archiveSubject(archiveId)
      setArchiveId(null)
      fetchAll()
    } catch {
      alert('Archive failed')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <PageHeader title="Subjects" subtitle="Common subjects apply to both streams; Natural (Physics, Chemistry, Biology) and Social (History, Geography, Economics) are stream-specific"
        action={<Button onClick={() => setModalOpen(true)}><Plus className="w-4 h-4 mr-1" /> Create</Button>} />

      <div className="flex gap-2 mb-4 flex-wrap">
        {[
          { key: undefined, label: 'All streams' },
          { key: 'Common', label: 'Common only' },
          { key: 'Natural Science', label: 'Natural' },
          { key: 'Social Science', label: 'Social' },
        ].map(({ key, label }) => (
          <button
            key={label}
            type="button"
            onClick={() => setApiStream(key)}
            className={`px-3 py-1.5 rounded-lg text-sm border ${
              apiStream === key ? 'border-accent bg-accent/10 text-accent' : 'border-surface-border text-muted'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      <DataTable
        loading={loading}
        data={subjects}
        keyExtractor={(s) => s.id}
        searchKeys={['name', 'code', 'stream']}
        filters={[
          { key: 'grade_level', label: 'All grades', options: [9, 10, 11, 12].map((g) => ({ value: String(g), label: `Grade ${g}` })) },
          { key: 'status', label: 'Status', options: [{ value: 'Active', label: 'Active' }, { value: 'Inactive', label: 'Inactive' }] },
        ]}
        columns={[
          { key: 'code', header: 'Code' },
          { key: 'name', header: 'Name' },
          { key: 'grade_level', header: 'Grade', render: (s) => s.grade_level || '—' },
          {
            key: 'stream', header: 'Stream',
            render: (s) => (
              <span title={s.stream ? `Only ${s.stream} students` : 'All streams'}>
                {s.stream || 'Common (both streams)'}
              </span>
            ),
          },
          { key: 'teacher', header: 'Teacher', render: (s) => s.teacher?.user?.name ?? '—' },
          {
            key: 'status', header: 'Status',
            render: (s) => <Badge label={s.status ?? 'Active'} variant={s.status === 'Inactive' ? 'danger' : 'success'} />,
          },
          {
            key: 'actions', header: '',
            render: (s) => (
              <RowActions
                onUpdate={() => {
                  setUpdateSubject(s)
                  setEditForm({ name: s.name, code: s.code, teacher_id: s.teacher_id })
                }}
                onArchive={() => setArchiveId(s.id)}
              />
            ),
          },
        ]}
      />

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title="Create Subject"
        footer={<><Button variant="ghost" onClick={() => setModalOpen(false)}>Cancel</Button><Button loading={saving} type="submit" form="subject-form">Create</Button></>}>
        <form id="subject-form" onSubmit={handleCreate} className="space-y-3">
          <Input label="Subject name" placeholder="Mathematics" value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} required />
          <Input label="Code (unique)" placeholder="MATH-G10-NAT" value={form.code} onChange={(e) => setForm((f) => ({ ...f, code: e.target.value }))} required />
          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm text-muted">Grade level
              <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.grade_level}
                onChange={(e) => setForm((f) => ({ ...f, grade_level: Number(e.target.value) }))}>
                {[9, 10, 11, 12].map((g) => <option key={g} value={g}>{g}</option>)}
              </select>
            </label>
            <label className="text-sm text-muted">Stream (optional)
              <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.stream}
                onChange={(e) => setForm((f) => ({ ...f, stream: e.target.value }))}>
                <option value="">Common (both)</option>
                <option value="Natural Science">Natural Science</option>
                <option value="Social Science">Social Science</option>
              </select>
            </label>
          </div>
          <label className="text-sm text-muted">Teacher
            <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.teacher_id}
              onChange={(e) => setForm((f) => ({ ...f, teacher_id: Number(e.target.value) }))} required>
              <option value={0} disabled>Select teacher</option>
              {teachers.map((t) => <option key={t.id} value={t.id}>{t.user?.name}</option>)}
            </select>
          </label>
          <label className="text-sm text-muted">Status
            <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.status}
              onChange={(e) => setForm((f) => ({ ...f, status: e.target.value }))}>
              <option value="Active">Active</option>
              <option value="Inactive">Inactive</option>
            </select>
          </label>
        </form>
      </Modal>

      <Modal open={!!updateSubject_} onClose={() => setUpdateSubject(null)} title={`Update — ${updateSubject_?.name}`}
        footer={<><Button variant="ghost" onClick={() => setUpdateSubject(null)}>Cancel</Button><Button loading={saving} type="submit" form="subject-update-form">Save</Button></>}>
        <form id="subject-update-form" onSubmit={handleUpdate} className="space-y-3">
          <Input label="Subject name" value={editForm.name} onChange={(e) => setEditForm((f) => ({ ...f, name: e.target.value }))} required />
          <Input label="Code" value={editForm.code} onChange={(e) => setEditForm((f) => ({ ...f, code: e.target.value }))} required />
          <label className="text-sm text-muted">Teacher
            <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={editForm.teacher_id}
              onChange={(e) => setEditForm((f) => ({ ...f, teacher_id: Number(e.target.value) }))} required>
              {teachers.map((t) => <option key={t.id} value={t.id}>{t.user?.name}</option>)}
            </select>
          </label>
        </form>
      </Modal>

      <ConfirmModal open={!!archiveId} onClose={() => setArchiveId(null)} title="Archive Subject"
        message="Subject will be moved to Trash." confirmLabel="Archive" variant="danger" loading={saving} onConfirm={handleArchive} />
    </div>
  )
}
