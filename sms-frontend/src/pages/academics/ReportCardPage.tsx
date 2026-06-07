import { useEffect, useRef, useState } from 'react'
import { useRole } from '../../hooks/useRole'
import { getReportCard, downloadReportCard } from '../../api/academics'
import { getMyChildren, getChildReportCard } from '../../api/parent'
import { listFromApi } from '../../types/api'
import type { ReportCard } from '../../types/academic'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Spinner } from '../../components/ui/Spinner'

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
          setChildList(kids.map((k: any) => ({ id: k.id, name: k.user?.name ?? `Student #${k.id}` })))
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

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-bold text-foreground">Report Card</h1>
          <p className="text-sm text-muted mt-0.5">
            {report.student.user?.name} · {report.semester} {report.year}
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
            <h2 className="text-sm font-semibold text-foreground">Grades</h2>
          </div>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                {['Subject', 'Type', 'Score', 'Remarks'].map((h) => (
                  <th key={h} className="text-left px-5 py-2.5 font-medium text-foreground">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {report.grades.map((g) => (
                <tr key={g.id} className="border-b border-surface-border last:border-0">
                  <td className="px-5 py-2.5 text-muted">{g.subject?.name ?? `#${g.subject_id}`}</td>
                  <td className="px-5 py-2.5 text-muted">{g.grade_type}</td>
                  <td className="px-5 py-2.5 font-medium text-foreground">{g.score}</td>
                  <td className="px-5 py-2.5 text-muted">{g.remarks || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="bg-surface border border-surface-border rounded-2xl overflow-hidden mt-4">
          <div className="px-5 py-3 border-b border-surface-border bg-surface-elevated">
            <h2 className="text-sm font-semibold text-foreground">Attendance</h2>
          </div>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                {['Subject', 'Present', 'Total', '%'].map((h) => (
                  <th key={h} className="text-left px-5 py-2.5 font-medium text-foreground">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {report.attendance.map((a) => (
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
        </div>
      </div>
    </div>
  )
}