import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Plus } from 'lucide-react'
import { getClasses, createClass, archiveClass, updateClass, getTeachers } from '../../api/admin'
import type { Class, Teacher } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { DataTable } from '../../components/ui/DataTable'
import { RowActions } from '../../components/ui/RowActions'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import { ConfirmModal } from '../../components/ui/ConfirmModal'
import { PageHeader } from '../../components/ui/PageHeader'
import { Input } from '../../components/ui/Input'
import { Badge } from '../../components/ui/Badge'
import { AlertModal } from '../../components/ui/AlertModal'

const SECTION_LETTERS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('')

function nextSectionLetter(used: string[]): string | null {
  for (const letter of SECTION_LETTERS) {
    if (!used.includes(letter)) return letter
  }
  return null
}

export default function ClassesPage() {
  const [classes, setClasses] = useState<Class[]>([])
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [updateClass_, setUpdateClass] = useState<Class | null>(null)
  const [archiveId, setArchiveId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({
    grade_level: 9, section: 'A', stream: '' as string,
    year: new Date().getFullYear(), teacher_id: 0, status: 'Active',
  })
  const [editForm, setEditForm] = useState({ teacher_id: 0, year: new Date().getFullYear(), status: 'Active' })
  const [alertState, setAlertState] = useState<{ open: boolean; title: string; message: string; type: 'success' | 'error' }>({
    open: false,
    title: '',
    message: '',
    type: 'success',
  })

  const fetchAll = () => {
    setLoading(true)
    Promise.all([getClasses({ page_size: 50 }), getTeachers({ page_size: 50 })])
      .then(([c, t]) => {
        setClasses(listFromApi(c.data))
        setTeachers(listFromApi(t.data))
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchAll() }, [])

  const usedSections = useMemo(() => {
    return classes
      .filter((c) => c.grade_level === form.grade_level && c.year === form.year)
      .map((c) => (c.section ?? '').toUpperCase())
  }, [classes, form.grade_level, form.year])

  const availableSection = useMemo(() => nextSectionLetter(usedSections), [usedSections])

  useEffect(() => {
    if (availableSection) setForm((f) => ({ ...f, section: availableSection }))
  }, [availableSection])

  const previewName = () => {
    let n = `${form.grade_level}${form.section}`
    if (form.grade_level >= 11 && form.stream) n += ` ${form.stream}`
    return n
  }

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    if (!availableSection) {
      setAlertState({ open: true, title: 'Validation Error', message: 'All sections (A–Z) already exist for this grade and year.', type: 'error' })
      return
    }
    const stream = form.grade_level >= 11 ? form.stream : ''
    setSaving(true)
    try {
      await createClass({
        grade_level: form.grade_level,
        section: form.section,
        stream,
        year: form.year,
        teacher_id: form.teacher_id,
        status: form.status,
      })
      setModalOpen(false)
      fetchAll()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setAlertState({ open: true, title: 'Create Failed', message: msg ?? 'Class may already exist', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleUpdate = async (e: FormEvent) => {
    e.preventDefault()
    if (!updateClass_) return
    setSaving(true)
    try {
      await updateClass(updateClass_.id, {
        teacher_id: editForm.teacher_id,
        year: editForm.year,
      })
      setUpdateClass(null)
      fetchAll()
      setAlertState({ open: true, title: 'Success', message: 'Class updated successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Update Failed', message: 'Failed to update class details', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleArchive = async () => {
    if (!archiveId) return
    setSaving(true)
    try {
      await archiveClass(archiveId)
      setArchiveId(null)
      fetchAll()
      setAlertState({ open: true, title: 'Success', message: 'Class archived successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Archive Failed', message: 'Ensure no students are assigned to this class.', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <PageHeader title="Classes" subtitle="Homeroom sections per grade (e.g. 9A, 9B — only unused sections can be created)"
        action={<Button onClick={() => setModalOpen(true)} disabled={!availableSection}><Plus className="w-4 h-4 mr-1" /> Create</Button>} />

      <DataTable
        loading={loading}
        data={classes}
        keyExtractor={(c) => c.id}
        searchKeys={['name', 'stream']}
        filters={[
          { key: 'grade_level', label: 'All grades', options: [9, 10, 11, 12].map((g) => ({ value: String(g), label: `Grade ${g}` })) },
          { key: 'status', label: 'All status', options: [{ value: 'Active', label: 'Active' }, { value: 'Inactive', label: 'Inactive' }] },
        ]}
        columns={[
          { key: 'name', header: 'Class' },
          { key: 'stream', header: 'Stream', render: (c) => c.stream || '—' },
          { key: 'year', header: 'Year' },
          { key: 'teacher', header: 'Homeroom', render: (c) => c.teacher?.user?.name ?? '—' },
          {
            key: 'status', header: 'Status',
            render: (c) => <Badge label={c.status ?? 'Active'} variant={c.status === 'Inactive' ? 'danger' : 'success'} />,
          },
          {
            key: 'actions', header: '',
            render: (c) => (
              <RowActions
                onUpdate={() => {
                  setUpdateClass(c)
                  setEditForm({ teacher_id: c.teacher_id, year: c.year, status: c.status ?? 'Active' })
                }}
                onArchive={() => setArchiveId(c.id)}
              />
            ),
          },
        ]}
      />

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title="Create Class"
        footer={<><Button variant="ghost" onClick={() => setModalOpen(false)}>Cancel</Button><Button loading={saving} type="submit" form="class-form" disabled={!availableSection}>Create</Button></>}>
        <form id="class-form" onSubmit={handleCreate} className="space-y-3">
          <p className="text-sm text-muted">Preview: <span className="text-foreground font-medium">{previewName()}</span></p>
          {usedSections.length > 0 && (
            <p className="text-xs text-muted">Existing sections for grade {form.grade_level}: {usedSections.join(', ')}</p>
          )}
          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm text-muted">Grade
              <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.grade_level}
                onChange={(e) => setForm((f) => ({
                  ...f,
                  grade_level: Number(e.target.value),
                  stream: Number(e.target.value) >= 11 ? 'Natural Science' : '',
                }))}>
                {[9, 10, 11, 12].map((g) => <option key={g} value={g}>{g}</option>)}
              </select>
            </label>
            <label className="text-sm text-muted">Section
              <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.section}
                onChange={(e) => setForm((f) => ({ ...f, section: e.target.value }))} required>
                {SECTION_LETTERS.filter((l) => !usedSections.includes(l)).map((l) => (
                  <option key={l} value={l}>{l}{l === availableSection ? ' — recommended' : ''}</option>
                ))}
              </select>
            </label>
          </div>
          {!availableSection && (
            <p className="text-sm text-danger">No sections available for this grade/year. All letters A–Z are in use.</p>
          )}
          {form.grade_level >= 11 ? (
            <label className="text-sm text-muted">Stream
              <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.stream}
                onChange={(e) => setForm((f) => ({ ...f, stream: e.target.value }))}>
                <option value="Natural Science">Natural Science</option>
                <option value="Social Science">Social Science</option>
              </select>
            </label>
          ) : (
            <p className="text-xs text-muted">Grade {form.grade_level} uses a common curriculum — no stream applies.</p>
          )}
          <Input label="Academic year" type="number" value={form.year} onChange={(e) => setForm((f) => ({ ...f, year: Number(e.target.value) }))} required />
          <label className="text-sm text-muted">Homeroom teacher
            <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.teacher_id}
              onChange={(e) => setForm((f) => ({ ...f, teacher_id: Number(e.target.value) }))} required>
              <option value={0} disabled>Select teacher</option>
              {teachers.map((t) => <option key={t.id} value={t.id}>{t.user?.name}</option>)}
            </select>
          </label>
        </form>
      </Modal>

      <Modal open={!!updateClass_} onClose={() => setUpdateClass(null)} title={`Update — ${updateClass_?.name}`}
        footer={<><Button variant="ghost" onClick={() => setUpdateClass(null)}>Cancel</Button><Button loading={saving} type="submit" form="class-update-form">Save</Button></>}>
        <form id="class-update-form" onSubmit={handleUpdate} className="space-y-3">
          <label className="text-sm text-muted">Homeroom teacher
            <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={editForm.teacher_id}
              onChange={(e) => setEditForm((f) => ({ ...f, teacher_id: Number(e.target.value) }))} required>
              {teachers.map((t) => <option key={t.id} value={t.id}>{t.user?.name}</option>)}
            </select>
          </label>
          <Input label="Academic year" type="number" value={editForm.year} onChange={(e) => setEditForm((f) => ({ ...f, year: Number(e.target.value) }))} required />
        </form>
      </Modal>

      <ConfirmModal open={!!archiveId} onClose={() => setArchiveId(null)} title="Archive Class"
        message="Class moves to Trash. All students must be reassigned first." confirmLabel="Archive" variant="danger" loading={saving} onConfirm={handleArchive} />

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