import { useEffect, useState, type FormEvent } from 'react'
import { useRole } from '../../hooks/useRole'
import { useAuth } from '../../hooks/useAuth'
import {
  recordAttendance,
  getClassAttendance,
  getAttendancePercentage,
} from '../../api/academics'
import { getStudents, getSubjects, getClasses } from '../../api/admin'
import type { Attendance, AttendancePercentage, Student, Subject, Class } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { Table } from '../../components/ui/Table'
import { Badge, statusVariant } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Spinner } from '../../components/ui/Spinner'

export default function AttendancePage() {
  const { isTeacher, isAdmin } = useRole()
  const canRecord = isTeacher || isAdmin

  return canRecord ? <TeacherView /> : <StudentView />
}

// ── Teacher / Admin ───────────────────────────────────────
function TeacherView() {
  const [classes, setClasses] = useState<Class[]>([])
  const [students, setStudents] = useState<Student[]>([])
  const [subjects, setSubjects] = useState<Subject[]>([])
  const [classId, setClassId] = useState(0)
  const [records, setRecords] = useState<Attendance[]>([])
  const [loadingRecords, setLoadingRecords] = useState(false)

  // Form state
  const [form, setForm] = useState({
    student_id: 0,
    subject_id: 0,
    date: new Date().toISOString().split('T')[0],
    status: 'Present' as 'Present' | 'Absent' | 'Late',
  })
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    Promise.all([getClasses(), getStudents(), getSubjects()]).then(([c, s, sub]) => {
      setClasses(listFromApi(c.data)); setStudents(listFromApi(s.data)); setSubjects(listFromApi(sub.data))
    })
  }, [])

  const loadClassAttendance = async (id: number) => {
    setClassId(id); setLoadingRecords(true)
    try {
      const res = await getClassAttendance(id)
      setRecords(res.data.data ?? [])
    } finally {
      setLoadingRecords(false)
    }
  }

  const handleRecord = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true); setMessage('')
    try {
      const res = await recordAttendance(form)
      const created = res.data.data
      if (created) setRecords((prev) => [created, ...prev])
      setMessage('Attendance recorded.')
    } finally {
      setSaving(false)
    }
  }

  const sel = "w-full px-3 py-2 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none focus:border-[var(--accent)]"

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold text-[var(--text-h)]">Attendance</h1>

      {/* Record form */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <h2 className="text-base font-semibold text-[var(--text-h)] mb-4">Record Attendance</h2>
        <form onSubmit={handleRecord} className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Student</label>
            <select value={form.student_id} onChange={(e) => setForm((f) => ({ ...f, student_id: Number(e.target.value) }))} className={sel} required>
              <option value={0} disabled>Select student...</option>
              {students.map((s) => <option key={s.id} value={s.id}>{s.user?.name ?? s.student_code}</option>)}
            </select>
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Subject</label>
            <select value={form.subject_id} onChange={(e) => setForm((f) => ({ ...f, subject_id: Number(e.target.value) }))} className={sel} required>
              <option value={0} disabled>Select subject...</option>
              {subjects.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
            </select>
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Date</label>
            <input type="date" value={form.date} onChange={(e) => setForm((f) => ({ ...f, date: e.target.value }))}
              className={sel} required />
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Status</label>
            <select value={form.status} onChange={(e) => setForm((f) => ({ ...f, status: e.target.value as 'Present' | 'Absent' | 'Late' }))} className={sel}>
              {['Present', 'Absent', 'Late'].map((s) => <option key={s}>{s}</option>)}
            </select>
          </div>
          <div className="sm:col-span-2 flex items-center gap-3">
            <Button type="submit" loading={saving}>Record</Button>
            {message && <p className="text-sm text-green-600">{message}</p>}
          </div>
        </form>
      </div>

      {/* View by class */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <div className="flex items-center gap-3 mb-4">
          <h2 className="text-base font-semibold text-[var(--text-h)]">View by Class</h2>
          <select value={classId} onChange={(e) => loadClassAttendance(Number(e.target.value))}
            className="px-3 py-1.5 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none">
            <option value={0}>Select class...</option>
            {classes.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
        </div>
        {loadingRecords ? <Spinner /> : (
          <Table keyExtractor={(r) => r.id} data={records}
            columns={[
              { key: 'student', header: 'Student', render: (r) => r.student?.user?.name ?? `#${r.student_id}` },
              { key: 'subject', header: 'Subject', render: (r) => r.subject?.name ?? `#${r.subject_id}` },
              { key: 'date', header: 'Date', render: (r) => new Date(r.date).toLocaleDateString() },
              { key: 'status', header: 'Status', render: (r) => <Badge label={r.status} variant={statusVariant(r.status)} /> },
            ]}
          />
        )}
      </div>
    </div>
  )
}

// ── Student ───────────────────────────────────────────────
function StudentView() {
  const { user } = useAuth()
  const [rows, setRows] = useState<AttendancePercentage[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!user) return
    // The student's own attendance — backend resolves student ID from JWT
    getAttendancePercentage(0)
      .then((r) => setRows(r.data.data ?? []))
      .finally(() => setLoading(false))
  }, [user])

  if (loading) return <Spinner fullPage />

  return (
    <div>
      <h1 className="text-xl font-bold text-[var(--text-h)] mb-6">My Attendance</h1>
      <Table keyExtractor={(r) => r.subject_id} data={rows}
        columns={[
          { key: 'subject_name', header: 'Subject' },
          { key: 'present', header: 'Present' },
          { key: 'total', header: 'Total Classes' },
          {
            key: 'percentage', header: 'Attendance %',
            render: (r) => {
              const v = r.percentage >= 75 ? 'success' : r.percentage >= 50 ? 'warning' : 'danger'
              return <Badge label={`${r.percentage.toFixed(1)}%`} variant={v} />
            },
          },
        ]}
      />
    </div>
  )
}