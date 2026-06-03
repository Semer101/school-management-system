import { useCallback, useEffect, useState } from 'react'
import { getAttendanceSummary, getClasses, type AttendanceSummaryRow } from '../../api/admin'
import type { Class } from '../../types/academic'
import { Badge, statusVariant } from '../../components/ui/Badge'
import { Spinner } from '../../components/ui/Spinner'
import { BarChart3 } from 'lucide-react'
import { EmptyState } from '../../components/ui/EmptyState'
import { listFromApi } from '../../types/api'
import { Button } from '../../components/ui/Button'

export default function AttendanceSummary() {
  const [rows, setRows] = useState<AttendanceSummaryRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [date, setDate] = useState('')
  const [grade, setGrade] = useState('')
  const [section, setSection] = useState('')
  const [classes, setClasses] = useState<Class[]>([])
  const [classId, setClassId] = useState('')

  const load = useCallback(() => {
    setLoading(true)
    setError('')
    const params: { date?: string; grade_level?: string; section?: string; class_id?: string } = {}
    if (date) params.date = date
    if (grade) params.grade_level = grade
    if (section) params.section = section
    if (classId) params.class_id = classId
    getAttendanceSummary(params)
      .then((r) => {
        const payload = r.data?.data
        setRows(Array.isArray(payload) ? payload : listFromApi(r.data))
      })
      .catch(() => setError('Failed to load attendance summary.'))
      .finally(() => setLoading(false))
  }, [date, grade, section, classId])

  useEffect(() => {
    getClasses({ page_size: 100 })
      .then((res) => {
        setClasses(listFromApi(res.data))
      })
      .catch(() => {})
  }, [])

  useEffect(() => { load() }, [load])

  if (loading && rows.length === 0) return <Spinner fullPage />

  return (
    <div>
      <h1 className="text-xl font-bold text-[var(--text-h)] mb-4">Daily Attendance</h1>
      <p className="text-sm text-muted mb-4">One record per student per school day (not per subject).</p>

      <div className="flex flex-wrap gap-3 mb-4 items-end">
        <label className="text-sm text-muted flex flex-col gap-1">
          Date
          <input type="date" value={date} onChange={(e) => setDate(e.target.value)}
            className="px-3 py-2 rounded-lg border border-surface-border bg-surface text-sm" />
        </label>
        <label className="text-sm text-muted flex flex-col gap-1">
          Class
          <select value={classId} onChange={(e) => setClassId(e.target.value)}
            className="px-3 py-2 rounded-lg border border-surface-border bg-surface text-sm">
            <option value="">All</option>
            {classes.map((c) => (
              <option key={c.id} value={String(c.id)}>{c.name}</option>
            ))}
          </select>
        </label>
        <label className="text-sm text-muted flex flex-col gap-1">
          Grade Level
          <select value={grade} onChange={(e) => setGrade(e.target.value)}
            className="px-3 py-2 rounded-lg border border-surface-border bg-surface text-sm">
            <option value="">All</option>
            {[9, 10, 11, 12].map((g) => <option key={g} value={String(g)}>Grade {g}</option>)}
          </select>
        </label>
        <label className="text-sm text-muted flex flex-col gap-1">
          Section
          <select value={section} onChange={(e) => setSection(e.target.value)}
            className="px-3 py-2 rounded-lg border border-surface-border bg-surface text-sm">
            <option value="">All</option>
            {'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('').map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </label>
        <Button size="sm" variant="secondary" onClick={load}>Apply</Button>
        {(date || grade || section || classId) && (
          <Button size="sm" variant="ghost" onClick={() => { setDate(''); setGrade(''); setSection(''); setClassId('') }}>Clear</Button>
        )}
      </div>

      {error && <p className="text-sm text-red-500 mb-4">{error}</p>}
      {!loading && rows.length === 0 && !error && (
        <EmptyState icon={BarChart3} title="No attendance data" description="Try different filters or record daily attendance first." />
      )}

      {rows.length > 0 && (
        <div className="rounded-xl border border-surface-border overflow-hidden">
          <div className="max-h-[480px] overflow-y-auto">
            <table className="w-full text-sm">
              <thead className="sticky top-0 bg-surface border-b border-surface-border z-10">
                <tr>
                  <th className="text-left px-4 py-3 text-muted font-medium">Student</th>
                  <th className="text-left px-4 py-3 text-muted font-medium">Code</th>
                  <th className="text-left px-4 py-3 text-muted font-medium">Class</th>
                  <th className="text-left px-4 py-3 text-muted font-medium">Date</th>
                  <th className="text-left px-4 py-3 text-muted font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r, i) => (
                  <tr key={`${r.student_code}-${r.date}-${i}`} className="border-b border-surface-border/60 hover:bg-surface/50">
                    <td className="px-4 py-3 text-foreground">{r.student_name}</td>
                    <td className="px-4 py-3 text-muted font-mono text-xs">{r.student_code}</td>
                    <td className="px-4 py-3 text-muted">{r.class_name || '—'}</td>
                    <td className="px-4 py-3 text-muted">{r.date}</td>
                    <td className="px-4 py-3"><Badge label={r.status} variant={statusVariant(r.status)} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="text-xs text-muted px-4 py-2 border-t border-surface-border">{rows.length} record(s)</p>
        </div>
      )}
    </div>
  )
}
