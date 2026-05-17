import { useEffect, useState, type FormEvent } from 'react'
import { getStudents, getSubjects, enrollStudent, unenrollStudent } from '../../api/admin'
import type { Student, Subject } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { Button } from '../../components/ui/Button'

export default function EnrollmentPage() {
  const [students, setStudents] = useState<Student[]>([])
  const [subjects, setSubjects] = useState<Subject[]>([])
  const [studentId, setStudentId] = useState(0)
  const [subjectId, setSubjectId] = useState(0)
  const [mode, setMode] = useState<'enroll' | 'unenroll'>('enroll')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([getStudents(), getSubjects()]).then(([s, sub]) => {
      setStudents(listFromApi(s.data)); setSubjects(listFromApi(sub.data))
    })
  }, [])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setMessage(''); setError(''); setLoading(true)
    try {
      if (mode === 'enroll') {
        await enrollStudent(studentId, subjectId)
        setMessage('Student enrolled successfully.')
      } else {
        await unenrollStudent(studentId, subjectId)
        setMessage('Student unenrolled successfully.')
      }
    } catch (err: unknown) {
      setError(
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Operation failed.'
      )
    } finally {
      setLoading(false)
    }
  }

  const selectClass = "w-full px-3 py-2 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none focus:border-[var(--accent)]"

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-xl font-bold text-[var(--text-h)] mb-6">Enrollment</h1>

      {/* Mode toggle */}
      <div className="flex gap-2 mb-6">
        {(['enroll', 'unenroll'] as const).map((m) => (
          <button key={m} onClick={() => setMode(m)}
            className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-colors capitalize
              ${mode === m ? 'bg-[var(--accent)] text-white' : 'bg-[var(--code-bg)] text-[var(--text)]'}`}
          >
            {m}
          </button>
        ))}
      </div>

      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Student</label>
            <select value={studentId} onChange={(e) => setStudentId(Number(e.target.value))} className={selectClass} required>
              <option value={0} disabled>Select student...</option>
              {students.map((s) => (
                <option key={s.id} value={s.id}>{s.user?.name ?? s.student_code}</option>
              ))}
            </select>
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Subject</label>
            <select value={subjectId} onChange={(e) => setSubjectId(Number(e.target.value))} className={selectClass} required>
              <option value={0} disabled>Select subject...</option>
              {subjects.map((s) => (
                <option key={s.id} value={s.id}>{s.name} ({s.code})</option>
              ))}
            </select>
          </div>

          {error && <p className="text-sm text-red-500">{error}</p>}
          {message && <p className="text-sm text-green-600">{message}</p>}

          <Button type="submit" loading={loading} variant={mode === 'unenroll' ? 'danger' : 'primary'}>
            {mode === 'enroll' ? 'Enroll Student' : 'Unenroll Student'}
          </Button>
        </form>
      </div>
    </div>
  )
}