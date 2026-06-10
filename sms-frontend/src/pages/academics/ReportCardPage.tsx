import { useEffect, useRef, useState } from 'react'
import { useRole } from '../../hooks/useRole'
import { getReportCard, downloadReportCard } from '../../api/academics'
import { getMyChildren, getChildReportCard } from '../../api/parent'
import { listFromApi } from '../../types/api'
import type { ReportCard, ReportCardAttendance, ReportCardSubject } from '../../types/academic'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Spinner } from '../../components/ui/Spinner'

function studentDisplayName(report: ReportCard): string {
  return report.student.user?.name ?? report.student.name ?? `Student #${report.student.id}`
}

function reportYear(report: ReportCard): number {
  if (report.year) return report.year
  if (report.academic_year) return Number(report.academic_year)
  return new Date().getFullYear()
}

function isOverallAttendance(
  att: ReportCard['attendance']
): att is ReportCardAttendance {
  return !!att && !Array.isArray(att) && 'overall_percentage' in att
}

export default function ReportCardPage() {
  const { isParent } = useRole()
  const [report, setReport] = useState<ReportCard | null>(null)
  const [childList, setChildList] = useState<{ id: number; name: string }[]>([])
  const [selectedChild, setSelectedChild] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)
  const [downloading, setDownloading] = useState(false)
  const [error, setError] = useState('')
  const printRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (isParent) {
      getMyChildren()
        .then((res) => {
          const kids = listFromApi(res.data)
          setChildList(kids.map((k: { id: number; user?: { name?: string } }) => ({
            id: k.id,
            name: k.user?.name ?? `Student #${k.id}`,
          })))
          if (kids.length > 0) setSelectedChild(kids[0].id)
        })
    }
  }, [isParent])

  useEffect(() => {
    if (!isParent) {
      setLoading(true)
      setError('')
      getReportCard(0)
        .then((r) => setReport(r.data.data ?? null))
        .catch(() => setError('Failed to load report card.'))
        .finally(() => setLoading(false))
    } else if (selectedChild) {
      setLoading(true)
      setError('')
      getChildReportCard(selectedChild)
        .then((r) => setReport(r.data.data ?? null))
        .catch(() => setError('Failed to load report card.'))
        .finally(() => setLoading(false))
    }
  }, [isParent, selectedChild])

  const handleDownload = async () => {
    if (!report) return
    setDownloading(true)
    try {
      const studentId = report.student.id
      const res = await downloadReportCard(studentId)
      const url = URL.createObjectURL(new Blob([res.data], { type: 'application/pdf' }))
      const a = document.createElement('a')
      a.href = url
      a.download = `report-card-${report.student.student_code ?? studentId}.pdf`
      a.click()
      URL.revokeObjectURL(url)
    } finally {
      setDownloading(false)
    }
  }

  const handlePrint = () => {
    const printWindow = window.open('', '_blank')
    if (!printWindow) { window.print(); return }
    const cardEl = printRef.current
    if (!cardEl) return
    printWindow.document.write(`
      <html><head><title>Report Card</title>
      <style>
        body { margin: 0; font-family: system-ui, sans-serif; padding: 24px; }
        @media print { @page { margin: 15mm; } body { padding: 0; -webkit-print-color-adjust: exact; print-color-adjust: exact; } }
        table { width: 100%; border-collapse: collapse; margin: 12px 0; }
        th, td { border: 1px solid #e5e7eb; padding: 8px 12px; text-align: left; font-size: 13px; }
        th { background: #f1f5f9; font-weight: 600; }
      </style></head><body>${cardEl.innerHTML}</body></html>`)
    printWindow.document.close()
    printWindow.focus()
    setTimeout(() => { printWindow.print() }, 500)
  }

  if (loading) return <Spinner fullPage />
  if (error) return <p className="text-sm text-red-500">{error}</p>
  if (!report) return null

  const subjects: ReportCardSubject[] = report.subjects ?? []
  const overallAtt = isOverallAttendance(report.attendance) ? report.attendance : null

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-bold text-foreground">Report Card</h1>
          <p className="text-sm text-muted mt-0.5">
            {studentDisplayName(report)}
            {report.semester ? ` · ${report.semester}` : ''} {reportYear(report)}
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="secondary" size="sm" onClick={handlePrint}>
            Print
          </Button>
          <Button variant="secondary" size="sm" loading={downloading} onClick={handleDownload}>
            Download PDF
          </Button>
        </div>
      </div>

      {isParent && childList.length > 0 && (
        <div className="bg-surface border border-surface-border rounded-xl p-4">
          <label className="text-sm font-medium text-foreground">Select Child</label>
          <select
            value={selectedChild ?? ''}
            onChange={(e) => setSelectedChild(Number(e.target.value))}
            className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface text-sm focus:outline-none focus:border-accent/50"
          >
            {childList.map((c) => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>
      )}

      <div ref={printRef}>
        <div className="bg-accent/5 border border-accent/20 rounded-2xl px-6 py-4 flex items-center justify-between">
          <span className="text-sm font-medium text-accent">Overall Average</span>
          <span className="text-2xl font-bold text-accent">{report.average?.toFixed(1) ?? '—'}%</span>
        </div>

        <div className="bg-surface border border-surface-border rounded-2xl overflow-hidden mt-4">
          <div className="px-5 py-3 border-b border-surface-border bg-surface-elevated">
            <h2 className="text-sm font-semibold text-foreground">Subject Grades</h2>
          </div>
          {subjects.length > 0 ? (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-surface-border">
                  {['Subject', 'Midterm', 'Final', 'Overall', 'Grade'].map((h) => (
                    <th key={h} className="text-left px-5 py-2.5 font-medium text-foreground">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {subjects.map((s) => (
                  <tr key={s.subject_name} className="border-b border-surface-border last:border-0">
                    <td className="px-5 py-2.5 text-muted">{s.subject_name}</td>
                    <td className="px-5 py-2.5 text-muted">{s.midterm > 0 ? s.midterm.toFixed(1) : '—'}</td>
                    <td className="px-5 py-2.5 text-muted">{s.final > 0 ? s.final.toFixed(1) : '—'}</td>
                    <td className="px-5 py-2.5 font-medium text-foreground">{s.overall.toFixed(1)}</td>
                    <td className="px-5 py-2.5">
                      <Badge label={s.letter_grade} variant="default" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <p className="px-5 py-6 text-sm text-muted">No grade records for this academic year.</p>
          )}
        </div>

        <div className="bg-surface border border-surface-border rounded-2xl overflow-hidden mt-4">
          <div className="px-5 py-3 border-b border-surface-border bg-surface-elevated">
            <h2 className="text-sm font-semibold text-foreground">Attendance</h2>
          </div>
          {overallAtt ? (
            <div className="px-5 py-4 text-sm text-muted space-y-1">
              <p>
                Overall: <strong className="text-foreground">{overallAtt.overall_percentage.toFixed(1)}%</strong>
              </p>
              <p>
                Days attended: <strong className="text-foreground">{overallAtt.attended}</strong> / {overallAtt.total_days}
              </p>
            </div>
          ) : Array.isArray(report.attendance) && report.attendance.length > 0 ? (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-surface-border">
                  {['Subject', 'Present', 'Total', '%'].map((h) => (
                    <th key={h} className="text-left px-5 py-2.5 font-medium text-foreground">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {report.attendance.map((a: any) => (
                  <tr key={a.subject_id} className="border-b border-surface-border last:border-0">
                    <td className="px-5 py-2.5 text-muted">{a.subject_name}</td>
                    <td className="px-5 py-2.5 text-muted">{a.present}</td>
                    <td className="px-5 py-2.5 text-muted">{a.total}</td>
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
          ) : (
            <p className="px-5 py-6 text-sm text-muted">No attendance records available.</p>
          )}
        </div>
      </div>
    </div>
  )
}
