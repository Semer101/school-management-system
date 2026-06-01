import { useEffect, useState, type FormEvent } from 'react'
import { X } from 'lucide-react'
import { useRole } from '../../hooks/useRole'
import { bulkGradeEntry, getSubjectGrades, getStudentGrades } from '../../api/academics'
import { getStudents, getSubjects } from '../../api/admin'
import type { Grade, Student, Subject } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { Table } from '../../components/ui/Table'
import { Button } from '../../components/ui/Button'
import { Spinner } from '../../components/ui/Spinner'

export default function GradesPage() {
  const { isTeacher, isAdmin } = useRole()
  return (isTeacher || isAdmin) ? <TeacherView /> : <StudentView />
}

// ── Teacher / Admin ───────────────────────────────────────
function TeacherView() {
  const [students, setStudents] = useState<Student[]>([])
  const [subjects, setSubjects] = useState<Subject[]>([])
  const [subjectId, setSubjectId] = useState(0)
  const [grades, setGrades] = useState<Grade[]>([])
  const [loadingGrades, setLoadingGrades] = useState(false)

  // Bulk entry row
  const emptyEntry = { student_id: 0, subject_id: 0, score: 0, grade_type: 'Exam', term: 'Term 1', remarks: '' }
  const [entries, setEntries] = useState([{ ...emptyEntry }])
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    Promise.all([getStudents({ page_size: 50 }), getSubjects({ page_size: 50 })]).then(([s, sub]) => {
      setStudents(listFromApi(s.data)); setSubjects(listFromApi(sub.data))
    })
  }, [])

  const loadGrades = async (id: number) => {
    setSubjectId(id); setLoadingGrades(true)
    const res = await getSubjectGrades(id)
    setGrades(res.data.data ?? [])
    setLoadingGrades(false)
  }

  const updateEntry = (i: number, key: string, val: string | number) =>
    setEntries((prev) => prev.map((e, idx) => idx === i ? { ...e, [key]: val } : e))

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault(); setSaving(true); setMessage('')
    try {
      await bulkGradeEntry(entries)
      setMessage('Grades saved successfully.')
      setEntries([{ ...emptyEntry }])
    } finally {
      setSaving(false)
    }
  }

  const sel = "w-full px-2 py-1.5 rounded-lg text-xs bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none focus:border-[var(--accent)]"

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold text-[var(--text-h)]">Grades</h1>

      {/* Bulk entry */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <h2 className="text-base font-semibold text-[var(--text-h)] mb-4">Bulk Grade Entry</h2>
        <form onSubmit={handleSubmit} className="space-y-3">
          {entries.map((entry, i) => (
            <div key={i} className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-2 items-end">
              <div>
                <label className="text-xs text-[var(--text)]">Student</label>
                <select value={entry.student_id} onChange={(e) => updateEntry(i, 'student_id', Number(e.target.value))} className={sel} required>
                  <option value={0} disabled>Student</option>
                  {students.map((s) => <option key={s.id} value={s.id}>{s.user?.name ?? s.student_code}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs text-[var(--text)]">Subject</label>
                <select value={entry.subject_id} onChange={(e) => updateEntry(i, 'subject_id', Number(e.target.value))} className={sel} required>
                  <option value={0} disabled>Subject</option>
                  {subjects.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs text-[var(--text)]">Score</label>
                <input type="number" min={0} max={100} value={entry.score}
                  onChange={(e) => updateEntry(i, 'score', Number(e.target.value))}
                  className={sel} required />
              </div>
              <div>
                <label className="text-xs text-[var(--text)]">Type</label>
                <select value={entry.grade_type} onChange={(e) => updateEntry(i, 'grade_type', e.target.value)} className={sel}>
                  {['Exam', 'Quiz', 'Assignment', 'Midterm', 'Final'].map((t) => <option key={t}>{t}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs text-[var(--text)]">Term</label>
                <select value={entry.term} onChange={(e) => updateEntry(i, 'term', e.target.value)} className={sel}>
                  {['Term 1', 'Term 2', 'Term 3'].map((t) => <option key={t}>{t}</option>)}
                </select>
              </div>
              <button type="button" onClick={() => setEntries((prev) => prev.filter((_, idx) => idx !== i))}
                className="text-danger hover:text-danger/80 p-1" aria-label="Remove row">
                <X className="w-4 h-4" />
              </button>
            </div>
          ))}

          <div className="flex items-center gap-3 pt-2">
            <Button type="button" variant="ghost" size="sm" onClick={() => setEntries((p) => [...p, { ...emptyEntry }])}>
              + Add Row
            </Button>
            <Button type="submit" loading={saving}>Save Grades</Button>
            {message && <p className="text-sm text-green-600">{message}</p>}
          </div>
        </form>
      </div>

      {/* View grades by subject */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <div className="flex items-center gap-3 mb-4">
          <h2 className="text-base font-semibold text-[var(--text-h)]">View by Subject</h2>
          <select value={subjectId} onChange={(e) => loadGrades(Number(e.target.value))}
            className="px-3 py-1.5 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none">
            <option value={0}>Select subject...</option>
            {subjects.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
          </select>
        </div>
        {loadingGrades ? <Spinner /> : (
          <Table keyExtractor={(g) => g.id} data={grades}
            columns={[
              { key: 'student', header: 'Student', render: (g) => g.student?.user?.name ?? `#${g.student_id}` },
              { key: 'score', header: 'Score' },
              { key: 'grade_type', header: 'Type' },
              { key: 'term', header: 'Term' },
              { key: 'remarks', header: 'Remarks' },
            ]}
          />
        )}
      </div>
    </div>
  )
}

// ── Student ───────────────────────────────────────────────
function StudentView() {
  const [grades, setGrades] = useState<Grade[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Backend resolves student ID from JWT
    getStudentGrades(0)
      .then((r) => setGrades(r.data.data ?? []))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <Spinner fullPage />

  return (
    <div>
      <h1 className="text-xl font-bold text-[var(--text-h)] mb-6">My Grades</h1>
      <Table keyExtractor={(g) => g.id} data={grades}
        columns={[
          { key: 'subject', header: 'Subject', render: (g) => g.subject?.name ?? `#${g.subject_id}` },
          { key: 'score', header: 'Score' },
          { key: 'grade_type', header: 'Type' },
          { key: 'term', header: 'Term' },
          { key: 'remarks', header: 'Remarks' },
          { key: 'created_at', header: 'Date', render: (g) => new Date(g.created_at).toLocaleDateString() },
        ]}
      />
    </div>
  )
}