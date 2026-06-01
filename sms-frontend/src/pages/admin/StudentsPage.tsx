import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Plus, BookOpen, TrendingUp } from 'lucide-react'
import {
  getStudents, archiveStudent, createStudent, updateStudent, getClasses,
  enrollStudent, unenrollStudent, getStudentEnrollmentStatus, promoteStudent,
  registerUser, type CreateStudentPayload, type EnrollmentStatusRow,
} from '../../api/admin'
import { RowActions } from '../../components/ui/RowActions'
import type { Student, Class } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { DataTable } from '../../components/ui/DataTable'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import { Input } from '../../components/ui/Input'
import { ConfirmModal } from '../../components/ui/ConfirmModal'
import { PageHeader } from '../../components/ui/PageHeader'
import { Badge } from '../../components/ui/Badge'
import { AlertModal } from '../../components/ui/AlertModal'

export default function StudentsPage() {
  const [students, setStudents] = useState<Student[]>([])
  const [classes, setClasses] = useState<Class[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [updateStudent_, setUpdateStudent] = useState<Student | null>(null)
  const [editForm, setEditForm] = useState({ class_id: 0, stream: '' as CreateStudentPayload['stream'], grade_level: 9 })
  const [archiveId, setArchiveId] = useState<number | null>(null)
  const [enrollStudent_, setEnrollStudent] = useState<Student | null>(null)
  const [enrollRows, setEnrollRows] = useState<EnrollmentStatusRow[]>([])
  const [promoteId, setPromoteId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState<CreateStudentPayload & { newParent?: boolean; parentPassword?: string }>({
    name: '', email: '', password: 'Student@1234',
    class_id: 0, parent_id: 0, parent_name: '', parent_email: '', parent_phone: '',
    stream: '', grade_level: 9, newParent: false, parentPassword: 'Parent@1234',
  })
  const [alertState, setAlertState] = useState<{ open: boolean; title: string; message: string; type: 'success' | 'error' }>({
    open: false,
    title: '',
    message: '',
    type: 'success',
  })

  const fetchStudents = async () => {
    setLoading(true)
    try {
      const res = await getStudents({ page_size: 50 })
      const body = res.data
      const payload = body.data
      if (Array.isArray(payload)) {
        setStudents(payload as Student[])
      } else if (payload && typeof payload === 'object' && 'data' in payload) {
        setStudents((payload as { data: Student[] }).data)
      } else {
        setStudents(listFromApi(body))
      }
    } catch {
      setStudents([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchStudents()
    getClasses({ page_size: 50 }).then((r) => setClasses(listFromApi(r.data))).catch(() => {})
  }, [])

  // Classes filtered by the currently selected grade (and stream for grades 11-12)
  // For grades 9-10 we also exclude any class that has a stream set (old legacy data)
  const filteredClassesForForm = useMemo(() => {
    return classes.filter((c) => {
      if (c.grade_level !== form.grade_level) return false
      if (form.grade_level >= 11) return c.stream === form.stream
      return !c.stream // grades 9-10: only show stream-less classes
    })
  }, [classes, form.grade_level, form.stream])

  const filteredClassesForEdit = useMemo(() => {
    return classes.filter((c) => {
      if (c.grade_level !== editForm.grade_level) return false
      if (editForm.grade_level >= 11) return c.stream === editForm.stream
      return !c.stream // grades 9-10: only stream-less classes
    })
  }, [classes, editForm.grade_level, editForm.stream])

  const openEnroll = async (s: Student) => {
    setEnrollStudent(s)
    try {
      const res = await getStudentEnrollmentStatus(s.id)
      setEnrollRows(Array.isArray(res.data.data) ? res.data.data : [])
    } catch {
      setEnrollRows([])
    }
  }

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      let parentId = form.parent_id
      if (form.newParent && form.parent_email) {
        const pr = await registerUser({
          name: form.parent_name || 'Parent',
          email: form.parent_email,
          password: form.parentPassword || 'Parent@1234',
          role: 'Parent',
        })
        parentId = (pr.data.data as { id: number }).id
      }
      const payload = { ...form }
      delete (payload as Partial<CreateStudentPayload>).student_code
      await createStudent({ ...payload, parent_id: parentId })
      setCreateOpen(false)
      fetchStudents()
      setAlertState({ open: true, title: 'Success', message: 'Student created successfully', type: 'success' })
    } catch (err: unknown) {
      const errMsg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Create failed'
      setAlertState({ open: true, title: 'Create Failed', message: errMsg, type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleArchive = async () => {
    if (!archiveId) return
    setSaving(true)
    try {
      await archiveStudent(archiveId)
      setArchiveId(null)
      fetchStudents()
      setAlertState({ open: true, title: 'Success', message: 'Student archived successfully', type: 'success' })
    } catch {
      setAlertState({ open: true, title: 'Archive Failed', message: 'Failed to archive student profile', type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleEnrollSubject = async (subjectId: number) => {
    if (!enrollStudent_) return
    try {
      await enrollStudent(enrollStudent_.id, subjectId)
      openEnroll(enrollStudent_)
    } catch (err: unknown) {
      const errMsg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Enroll failed'
      setAlertState({ open: true, title: 'Enrollment Failed', message: errMsg, type: 'error' })
    }
  }

  const handleUnenrollSubject = async (subjectId: number) => {
    if (!enrollStudent_) return
    try {
      await unenrollStudent(enrollStudent_.id, subjectId)
      openEnroll(enrollStudent_)
    } catch (err: unknown) {
      const errMsg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Unenroll failed'
      setAlertState({ open: true, title: 'Unenrollment Failed', message: errMsg, type: 'error' })
    }
  }

  const handleUpdate = async (e: FormEvent) => {
    e.preventDefault()
    if (!updateStudent_) return
    setSaving(true)
    try {
      await updateStudent(updateStudent_.id, {
        class_id: editForm.class_id,
        stream: editForm.stream,
        grade_level: editForm.grade_level,
      })
      setUpdateStudent(null)
      fetchStudents()
      setAlertState({ open: true, title: 'Success', message: 'Student details updated successfully', type: 'success' })
    } catch (err: unknown) {
      const errMsg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Update failed'
      setAlertState({ open: true, title: 'Update Failed', message: errMsg, type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handlePromote = async () => {
    if (!promoteId) return
    setSaving(true)
    try {
      await promoteStudent(promoteId)
      setPromoteId(null)
      fetchStudents()
      setAlertState({ open: true, title: 'Success', message: 'Promotion processed — student enrolled in new year subjects if eligible.', type: 'success' })
    } catch (err: unknown) {
      const errMsg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Promotion failed'
      setAlertState({ open: true, title: 'Promotion Failed', message: errMsg, type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <PageHeader
        title="Students"
        subtitle="Ethiopian Grades 9–12 — Natural & Social Science streams"
        action={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="w-4 h-4 mr-1" /> Create
          </Button>
        }
      />

      <DataTable
        loading={loading}
        data={students}
        keyExtractor={(s) => s.id}
        searchPlaceholder="Search name, code, email..."
        filters={[
          { key: 'stream', label: 'All streams', options: [
            { value: 'Natural Science', label: 'Natural Science' },
            { value: 'Social Science', label: 'Social Science' },
          ]},
          { key: 'grade_level', label: 'All grades', options: [9, 10, 11, 12].map((g) => ({
            value: String(g), label: `Grade ${g}`,
          }))},
        ]}
        columns={[
          { key: 'student_code', header: 'Code' },
          { key: 'user', header: 'Name', render: (s) => s.user?.name ?? '—' },
          { key: 'stream', header: 'Stream', render: (s) => s.stream || '—' },
          { key: 'class', header: 'Class', render: (s) => s.class?.name ?? '—' },
          {
            key: 'promotion',
            header: 'Status',
            render: (s) => (
              <Badge
                label={s.promotion_status ?? 'normal'}
                variant={s.promotion_status === 'repeat' ? 'danger' : s.promotion_status === 'conditional' ? 'warning' : 'success'}
              />
            ),
          },
          {
            key: 'enroll',
            header: 'Enrollment',
            render: (s) => (
              <Button size="sm" variant="ghost" onClick={() => openEnroll(s)}>
                <BookOpen className="w-3 h-3 mr-1" /> Subjects
              </Button>
            ),
          },
          {
            key: 'actions',
            header: '',
            render: (s) => (
              <div className="flex gap-1 items-center">
                <RowActions
                  onUpdate={() => {
                    setUpdateStudent(s)
                    setEditForm({
                      class_id: s.class_id,
                      stream: (s.grade_level ?? 9) >= 11
                        ? ((s.stream as CreateStudentPayload['stream']) || 'Natural Science')
                        : '',
                      grade_level: s.grade_level ?? 9,
                    })
                  }}
                  onArchive={() => setArchiveId(s.id)}
                />
                <Button size="sm" variant="ghost" onClick={() => setPromoteId(s.id)} title="Year-end promotion">
                  <TrendingUp className="w-3 h-3" />
                </Button>
              </div>
            ),
          },
        ]}
      />

      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="Create Student"
        footer={<><Button variant="ghost" onClick={() => setCreateOpen(false)}>Cancel</Button><Button loading={saving} type="submit" form="student-form">Create</Button></>}>
        <form id="student-form" onSubmit={handleCreate} className="space-y-3 max-h-[60vh] overflow-y-auto pr-1">
          <Input label="Full name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <Input label="Email" type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} required />
          <Input label="Password" type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required />
          <p className="text-xs text-muted">Student code is assigned automatically on create (e.g. STU-2025-001).</p>
          <div className={`grid gap-3 ${form.grade_level >= 11 ? 'grid-cols-2' : 'grid-cols-1'}`}>
            <label className="text-sm text-muted">Grade
              <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.grade_level}
                onChange={(e) => {
                  const g = Number(e.target.value)
                  setForm({ ...form, grade_level: g, stream: g >= 11 ? 'Natural Science' : '', class_id: 0 })
                }}>
                {[9, 10, 11, 12].map((g) => <option key={g} value={g}>Grade {g}</option>)}
              </select>
            </label>
            {form.grade_level >= 11 && (
              <label className="text-sm text-muted">Stream
                <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.stream}
                  onChange={(e) => setForm({ ...form, stream: e.target.value as CreateStudentPayload['stream'], class_id: 0 })}>
                  <option value="Natural Science">Natural Science</option>
                  <option value="Social Science">Social Science</option>
                </select>
              </label>
            )}
          </div>
          <label className="text-sm text-muted">Class
            <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={form.class_id}
              onChange={(e) => setForm({ ...form, class_id: Number(e.target.value) })} required>
              <option value={0}>Select class</option>
              {filteredClassesForForm.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </label>
          {filteredClassesForForm.length === 0 && (
            <p className="text-xs text-warning">No classes exist for Grade {form.grade_level}{form.grade_level >= 11 ? ` — ${form.stream}` : ''}. Create one first.</p>
          )}
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={form.newParent} onChange={(e) => setForm({ ...form, newParent: e.target.checked })} />
            Create new parent account
          </label>
          {form.newParent ? (
            <>
              <Input label="Parent name" value={form.parent_name} onChange={(e) => setForm({ ...form, parent_name: e.target.value })} required />
              <Input label="Parent email" type="email" value={form.parent_email} onChange={(e) => setForm({ ...form, parent_email: e.target.value })} required />
              <Input label="Parent phone" value={form.parent_phone} onChange={(e) => setForm({ ...form, parent_phone: e.target.value })} placeholder="09xxxxxxxx" required />
            </>
          ) : (
            <Input label="Parent user ID" type="number" value={form.parent_id || ''} onChange={(e) => setForm({ ...form, parent_id: Number(e.target.value) })} required />
          )}
        </form>
      </Modal>

      <Modal open={!!updateStudent_} onClose={() => setUpdateStudent(null)} title={`Update — ${updateStudent_?.user?.name}`}
        footer={<><Button variant="ghost" onClick={() => setUpdateStudent(null)}>Cancel</Button><Button loading={saving} type="submit" form="student-update-form">Save</Button></>}>
        <form id="student-update-form" onSubmit={handleUpdate} className="space-y-3">
          <p className="text-sm text-muted">Code: <span className="font-mono text-foreground">{updateStudent_?.student_code}</span></p>
          <div className={`grid gap-3 ${editForm.grade_level >= 11 ? 'grid-cols-2' : 'grid-cols-1'}`}>
            <label className="text-sm text-muted">Grade
              <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={editForm.grade_level}
                onChange={(e) => {
                  const g = Number(e.target.value)
                  setEditForm({ ...editForm, grade_level: g, stream: g >= 11 ? 'Natural Science' : '', class_id: 0 })
                }}>
                {[9, 10, 11, 12].map((g) => <option key={g} value={g}>Grade {g}</option>)}
              </select>
            </label>
            {editForm.grade_level >= 11 && (
              <label className="text-sm text-muted">Stream
                <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={editForm.stream}
                  onChange={(e) => setEditForm({ ...editForm, stream: e.target.value as CreateStudentPayload['stream'], class_id: 0 })}>
                  <option value="Natural Science">Natural Science</option>
                  <option value="Social Science">Social Science</option>
                </select>
              </label>
            )}
          </div>
          <label className="text-sm text-muted">Class
            <select className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface" value={editForm.class_id}
              onChange={(e) => setEditForm({ ...editForm, class_id: Number(e.target.value) })} required>
              <option value={0} disabled>Select class</option>
              {filteredClassesForEdit.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </label>
          {filteredClassesForEdit.length === 0 && (
            <p className="text-xs text-warning">No classes for Grade {editForm.grade_level}{editForm.grade_level >= 11 ? ` — ${editForm.stream}` : ''}.</p>
          )}
        </form>
      </Modal>

      <Modal open={!!enrollStudent_} onClose={() => setEnrollStudent(null)}
        title={`Subjects — ${enrollStudent_?.user?.name}`}>
        <div className="mb-3 flex items-center justify-between">
          <p className="text-xs text-muted">
            Grade {enrollStudent_?.grade_level}
            {enrollStudent_?.grade_level && enrollStudent_.grade_level >= 11 && enrollStudent_?.stream
              ? ` — ${enrollStudent_.stream}` : ' — Common curriculum'}
          </p>
          <span className="text-xs text-muted">
            {enrollRows.filter((r) => r.enrolled).length} / {enrollRows.length} enrolled
          </span>
        </div>
        <div className="space-y-2 max-h-[50vh] overflow-y-auto">
          {enrollRows.length === 0 && (
            <p className="text-sm text-muted text-center py-6">No subjects available for this grade/stream.</p>
          )}
          {enrollRows.map((row) => (
            <div key={row.subject_id} className="flex items-center justify-between py-2 border-b border-surface-border last:border-0">
              <div>
                <span className="text-sm font-medium">{row.subject_name}</span>
                <span className="ml-2 text-muted font-mono text-xs">({row.subject_code})</span>
              </div>
              {row.enrolled ? (
                <div className="flex items-center gap-2">
                  <Badge label="Enrolled" variant="success" />
                  <Button size="sm" variant="ghost" onClick={() => handleUnenrollSubject(row.subject_id)}>
                    Remove
                  </Button>
                </div>
              ) : (
                <Button size="sm" onClick={() => handleEnrollSubject(row.subject_id)}>Enroll</Button>
              )}
            </div>
          ))}
        </div>
      </Modal>

      <ConfirmModal open={!!archiveId} onClose={() => setArchiveId(null)} title="Archive Student"
        message="This student will be moved to Trash. Enrollments will be removed but grades and attendance are kept."
        confirmLabel="Archive" variant="danger" loading={saving} onConfirm={handleArchive} />

      <ConfirmModal open={!!promoteId} onClose={() => setPromoteId(null)} title="Year-End Promotion"
        message="Checks all subject averages from the current year. Students with 3+ failures repeat; 1–2 get conditional promotion; others advance and are auto-enrolled in next grade subjects."
        confirmLabel="Run promotion" loading={saving} onConfirm={handlePromote} />

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