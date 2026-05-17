import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  getChildAttendance,
  getChildGrades,
  getChildReportCard,
  downloadChildReportCard,
} from '../../api/parent'
import type { Grade, AttendancePercentage, ReportCard } from '../../types/academic'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Spinner } from '../../components/ui/Spinner'

type Tab = 'attendance' | 'grades' | 'reportcard'

export default function ChildDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const studentId = Number(id)
  const [tab, setTab] = useState<Tab>('attendance')
  const [attendance, setAttendance] = useState<AttendancePercentage[]>([])
  const [grades, setGrades] = useState<Grade[]>([])
  const [report, setReport] = useState<ReportCard | null>(null)
  const [loading, setLoading] = useState(true)
  const [downloading, setDownloading] = useState(false)

  useEffect(() => {
    Promise.all([
      getChildAttendance(studentId),
      getChildGrades(studentId),
      getChildReportCard(studentId),
    ])
      .then(([a, g, r]) => {
        setAttendance(a.data.data ?? []); setGrades(g.data.data ?? []); setReport(r.data.data ?? null)
      })
      .finally(() => setLoading(false))
  }, [studentId])

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

  if (loading) return <Spinner fullPage />

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-3">
        <button onClick={() => navigate('/parent/children')} className="text-[var(--text)] hover:text-[var(--text-h)]">←</button>
        <h1 className="text-xl font-bold text-[var(--text-h)]">
          {report?.student?.user?.name ?? `Student #${studentId}`}
        </h1>
      </div>

      {/* Tabs */}
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

      {/* Attendance */}
      {tab === 'attendance' && (
        <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl overflow-hidden">
          <table className="w-full text-sm">
            <thead><tr className="border-b border-[var(--border)] bg-[var(--code-bg)]">
              {['Subject', 'Present', 'Total', '%'].map((h) => (
                <th key={h} className="text-left px-5 py-3 font-semibold text-[var(--text-h)]">{h}</th>
              ))}
            </tr></thead>
            <tbody>
              {attendance.map((a) => (
                <tr key={a.subject_id} className="border-b border-[var(--border)] last:border-0">
                  <td className="px-5 py-2.5 text-[var(--text)]">{a.subject_name}</td>
                  <td className="px-5 py-2.5 text-[var(--text)]">{a.present}</td>
                  <td className="px-5 py-2.5 text-[var(--text)]">{a.total}</td>
                  <td className="px-5 py-2.5">
                    <Badge label={`${a.percentage.toFixed(1)}%`}
                      variant={a.percentage >= 75 ? 'success' : a.percentage >= 50 ? 'warning' : 'danger'} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Grades */}
      {tab === 'grades' && (
        <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl overflow-hidden">
          <table className="w-full text-sm">
            <thead><tr className="border-b border-[var(--border)] bg-[var(--code-bg)]">
              {['Subject', 'Type', 'Score', 'Term'].map((h) => (
                <th key={h} className="text-left px-5 py-3 font-semibold text-[var(--text-h)]">{h}</th>
              ))}
            </tr></thead>
            <tbody>
              {grades.map((g) => (
                <tr key={g.id} className="border-b border-[var(--border)] last:border-0">
                  <td className="px-5 py-2.5 text-[var(--text)]">{g.subject?.name}</td>
                  <td className="px-5 py-2.5 text-[var(--text)]">{g.grade_type}</td>
                  <td className="px-5 py-2.5 font-medium text-[var(--text-h)]">{g.score}</td>
                  <td className="px-5 py-2.5 text-[var(--text)]">{g.term}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Report Card */}
      {tab === 'reportcard' && report && (
        <div className="space-y-4">
          <div className="flex items-center justify-between px-5 py-3 bg-[var(--accent-bg)] border border-[var(--accent-border)] rounded-xl">
            <span className="text-sm font-medium text-[var(--accent)]">Average: {report.average?.toFixed(1)}%</span>
            <Button size="sm" variant="secondary" loading={downloading} onClick={handleDownload}>
              ⬇ PDF
            </Button>
          </div>
          <p className="text-sm text-[var(--text)]">{report.term} · {report.year}</p>
        </div>
      )}
    </div>
  )
}