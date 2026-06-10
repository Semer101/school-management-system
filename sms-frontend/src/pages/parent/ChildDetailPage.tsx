import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ChevronLeft, Filter } from 'lucide-react'
import {
  getChildAttendance,
  getChildGrades,
  getChildReportCard,
  downloadChildReportCard,
} from '../../api/parent'
import type { Grade, ReportCard } from '../../types/academic'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Spinner } from '../../components/ui/Spinner'

type Tab = 'attendance' | 'grades' | 'reportcard'

const SEMESTERS = ['Semester 1', 'Semester 2', 'Semester 3']

interface OverallAttendance {
  overall_percentage: number
  total_days: number
  attended: number
  by_month?: { month: string; total: number; present: number; percentage: number }[]
}

export default function ChildDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const studentId = Number(id)
  const [tab, setTab] = useState<Tab>('attendance')
  const [attendance, setAttendance] = useState<OverallAttendance | null>(null)
  const [grades, setGrades] = useState<Grade[]>([])
  const [report, setReport] = useState<ReportCard | null>(null)
  const [loading, setLoading] = useState(true)
  const [downloading, setDownloading] = useState(false)
  const [filterSemester, setFilterSemester] = useState<string>('')

  useEffect(() => {
    Promise.all([
      getChildAttendance(studentId),
      getChildGrades(studentId, filterSemester || undefined),
      getChildReportCard(studentId),
    ])
      .then(([a, g, r]) => {
        const attData = a.data.data as OverallAttendance | null
        setAttendance(attData)
        setGrades(Array.isArray(g.data.data) ? g.data.data : [])
        setReport(r.data.data ?? null)
      })
      .finally(() => setLoading(false))
  }, [studentId, filterSemester])

  const handleDownload = async () => {
    setDownloading(true)
    try {
      const res = await downloadChildReportCard(studentId)
      const url = URL.createObjectURL(new Blob([res.data], { type: 'application/pdf' }))
      const a = document.createElement('a')
      a.href = url; a.download = `report-card-${studentId}.pdf`; a.click()
      URL.revokeObjectURL(url)
    } finally {
      setDownloading(false)
    }
  }

  const clearFilters = () => {
    setFilterSemester('')
  }

  if (loading) return <Spinner fullPage />

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-3">
        <button type="button" onClick={() => navigate('/parent/children')} className="text-muted hover:text-foreground p-1" aria-label="Back">
          <ChevronLeft className="w-5 h-5" />
        </button>
        <h1 className="text-xl font-bold text-[var(--text-h)]">
          {report?.student?.user?.name ?? `Student #${studentId}`}
        </h1>
      </div>

      <div className="flex gap-2">
        {(['attendance', 'grades', 'reportcard'] as Tab[]).map((t) => (
          <button key={t} onClick={() => setTab(t)}
            className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-colors capitalize
              ${tab === t ? 'bg-[var(--accent)] text-white' : 'bg-[var(--code-bg)] text-[var(--text)]'}`}
          >
            {t === 'reportcard' ? 'Report Card' : t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {tab === 'attendance' && (
        <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl overflow-hidden">
          <table className="w-full text-sm">
            <thead><tr className="border-b border-[var(--border)] bg-[var(--code-bg)]">
              {['Metric', 'Value'].map((h) => (
                <th key={h} className="text-left px-5 py-3 font-semibold text-[var(--text-h)]">{h}</th>
              ))}
            </tr></thead>
            <tbody>
              {attendance === null ? (
                <tr>
                  <td colSpan={2} className="px-5 py-8 text-center text-[var(--text)]">
                    No attendance records available
                  </td>
                </tr>
              ) : (
                <>
                  <tr className="border-b border-[var(--border)]">
                    <td className="px-5 py-2.5 text-[var(--text)]">Overall Attendance</td>
                    <td className="px-5 py-2.5">
                      <Badge label={`${attendance.overall_percentage.toFixed(1)}%`}
                        variant={attendance.overall_percentage >= 75 ? 'success' : attendance.overall_percentage >= 50 ? 'warning' : 'danger'} />
                    </td>
                  </tr>
                  <tr className="border-b border-[var(--border)]">
                    <td className="px-5 py-2.5 text-[var(--text)]">Days Attended</td>
                    <td className="px-5 py-2.5 text-[var(--text-h)] font-medium">{attendance.attended} / {attendance.total_days}</td>
                  </tr>
                </>
              )}
            </tbody>
          </table>
        </div>
      )}

      {tab === 'grades' && (
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <Filter className="w-4 h-4 text-muted" />
              <select
                value={filterSemester}
                onChange={(e) => setFilterSemester(e.target.value)}
                className="px-3 py-1.5 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none focus:border-accent/50"
              >
                <option value="">All Semesters</option>
                {SEMESTERS.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
            {filterSemester && (
              <button onClick={clearFilters} className="text-xs text-accent hover:underline">
                Clear
              </button>
            )}
          </div>

          <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl overflow-hidden">
            <table className="w-full text-sm">
              <thead><tr className="border-b border-[var(--border)] bg-[var(--code-bg)]">
                {['Subject', 'Score', 'Semester'].map((h) => (
                  <th key={h} className="text-left px-5 py-3 font-semibold text-[var(--text-h)]">{h}</th>
                ))}
              </tr></thead>
              <tbody>
                {grades.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="px-5 py-8 text-center text-[var(--text)]">
                      No grade records available
                    </td>
                  </tr>
                ) : grades.map((g) => (
                  <tr key={g.id} className="border-b border-[var(--border)] last:border-0">
                    <td className="px-5 py-2.5 text-[var(--text)]">{g.subject?.name ?? 'Unknown'}</td>
                    <td className="px-5 py-2.5 font-medium text-[var(--text-h)]">{g.score}/100</td>
                    <td className="px-5 py-2.5 text-[var(--text)]">{g.semester}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === 'reportcard' && report && (
        <div className="space-y-4">
          <div className="flex items-center justify-between px-5 py-3 bg-[var(--accent-bg)] border border-[var(--accent-border)] rounded-xl">
            <div>
              <span className="text-sm font-medium text-[var(--accent)]">Average: {report.average?.toFixed(1)}%</span>
              {report.semester && <span className="ml-2 text-sm text-[var(--text)]">({report.semester})</span>}
            </div>
            <Button size="sm" variant="secondary" loading={downloading} onClick={handleDownload}>
              ⬇ PDF
            </Button>
          </div>

          <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-5 space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-[var(--text-h)] mb-2">Student Information</h3>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div><span className="text-[var(--text)]">Name:</span> <span className="text-[var(--text-h)] font-medium">{report.student?.user?.name ?? '—'}</span></div>
                <div><span className="text-[var(--text)]">Code:</span> <span className="text-[var(--text-h)] font-medium">{report.student?.student_code ?? '—'}</span></div>
                <div><span className="text-[var(--text)]">Class:</span> <span className="text-[var(--text-h)] font-medium">{report.student?.class ?? '—'}</span></div>
                <div><span className="text-[var(--text)]">Year:</span> <span className="text-[var(--text-h)] font-medium">{report.year ?? '—'}</span></div>
              </div>
            </div>

            {report.subjects && report.subjects.length > 0 && (
              <div>
                <h3 className="text-sm font-semibold text-[var(--text-h)] mb-2">Subject Grades</h3>
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-[var(--border)]">
                      {['Subject', 'Score', 'Grade'].map((h) => (
                        <th key={h} className="text-left px-3 py-2 font-medium text-[var(--text-h)]">{h}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {report.subjects.map((s) => (
                      <tr key={s.subject_name} className="border-b border-[var(--border)] last:border-0">
                        <td className="px-3 py-2 text-[var(--text)]">{s.subject_name}</td>
                        <td className="px-3 py-2 font-medium text-[var(--text-h)]">{s.overall.toFixed(1)}</td>
                        <td className="px-3 py-2"><Badge label={s.letter_grade} variant="default" /></td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}