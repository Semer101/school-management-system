import { useEffect, useState } from 'react'
import { getReportCard, downloadReportCard } from '../../api/academics'
import type { ReportCard } from '../../types/academic'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Spinner } from '../../components/ui/Spinner'

export default function ReportCardPage() {
  const [report, setReport] = useState<ReportCard | null>(null)
  const [loading, setLoading] = useState(true)
  const [downloading, setDownloading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    // Backend resolves student from JWT
    getReportCard(0)
      .then((r) => setReport(r.data.data ?? null))
      .catch(() => setError('Failed to load report card.'))
      .finally(() => setLoading(false))
  }, [])

  const handleDownload = async () => {
    if (!report) return
    setDownloading(true)
    try {
      const res = await downloadReportCard(report.student.id)
      const url = URL.createObjectURL(new Blob([res.data], { type: 'application/pdf' }))
      const a = document.createElement('a')
      a.href = url
      a.download = `report-card-${report.student.student_code}.pdf`
      a.click()
      URL.revokeObjectURL(url)
    } finally {
      setDownloading(false)
    }
  }

  if (loading) return <Spinner fullPage />
  if (error) return <p className="text-sm text-red-500">{error}</p>
  if (!report) return null

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-bold text-[var(--text-h)]">Report Card</h1>
          <p className="text-sm text-[var(--text)] mt-0.5">
            {report.student.user?.name} · {report.term} {report.year}
          </p>
        </div>
        <Button variant="secondary" size="sm" loading={downloading} onClick={handleDownload}>
          ⬇ Download PDF
        </Button>
      </div>

      {/* Average */}
      <div className="bg-[var(--accent-bg)] border border-[var(--accent-border)] rounded-2xl px-6 py-4 flex items-center justify-between">
        <span className="text-sm font-medium text-[var(--accent)]">Overall Average</span>
        <span className="text-2xl font-bold text-[var(--accent)]">{report.average?.toFixed(1) ?? '—'}%</span>
      </div>

      {/* Grades */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl overflow-hidden">
        <div className="px-5 py-3 border-b border-[var(--border)] bg-[var(--code-bg)]">
          <h2 className="text-sm font-semibold text-[var(--text-h)]">Grades</h2>
        </div>
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[var(--border)]">
              {['Subject', 'Type', 'Score', 'Remarks'].map((h) => (
                <th key={h} className="text-left px-5 py-2.5 font-medium text-[var(--text-h)]">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {report.grades.map((g) => (
              <tr key={g.id} className="border-b border-[var(--border)] last:border-0">
                <td className="px-5 py-2.5 text-[var(--text)]">{g.subject?.name ?? `#${g.subject_id}`}</td>
                <td className="px-5 py-2.5 text-[var(--text)]">{g.grade_type}</td>
                <td className="px-5 py-2.5 font-medium text-[var(--text-h)]">{g.score}</td>
                <td className="px-5 py-2.5 text-[var(--text)]">{g.remarks || '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Attendance */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl overflow-hidden">
        <div className="px-5 py-3 border-b border-[var(--border)] bg-[var(--code-bg)]">
          <h2 className="text-sm font-semibold text-[var(--text-h)]">Attendance</h2>
        </div>
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[var(--border)]">
              {['Subject', 'Present', 'Total', '%'].map((h) => (
                <th key={h} className="text-left px-5 py-2.5 font-medium text-[var(--text-h)]">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {report.attendance.map((a) => (
              <tr key={a.subject_id} className="border-b border-[var(--border)] last:border-0">
                <td className="px-5 py-2.5 text-[var(--text)]">{a.subject_name}</td>
                <td className="px-5 py-2.5 text-[var(--text)]">{a.present}</td>
                <td className="px-5 py-2.5 text-[var(--text)]">{a.total}</td>
                <td className="px-5 py-2.5">
                  <Badge
                    label={`${a.percentage.toFixed(1)}%`}
                    variant={a.percentage >= 75 ? 'success' : a.percentage >= 50 ? 'warning' : 'danger'}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}