import { useEffect, useState } from 'react'
import { getAttendanceSummary, type AttendanceSummaryRow } from '../../api/admin'
import { Table } from '../../components/ui/Table'
import { Badge } from '../../components/ui/Badge'
import { Spinner } from '../../components/ui/Spinner'
import { EmptyState } from '../../components/ui/EmptyState'
import { listFromApi } from '../../types/api'

export default function AttendanceSummary() {
  const [rows, setRows] = useState<AttendanceSummaryRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    getAttendanceSummary()
      .then((r) => setRows(listFromApi(r.data)))
      .catch(() => setError('Failed to load attendance summary.'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <Spinner fullPage />
  if (error) return <p className="text-sm text-red-500">{error}</p>
  if (rows.length === 0) return <EmptyState icon="📊" title="No attendance data yet" />

  return (
    <div>
      <h1 className="text-xl font-bold text-[var(--text-h)] mb-6">Attendance Summary</h1>
      <Table
        keyExtractor={(r) => `${r.student_code}-${r.subject_name}-${r._i}`}
        data={rows.map((r, i) => ({ ...r, _i: i }))}
        columns={[
          { key: 'student_name', header: 'Student' },
          { key: 'student_code', header: 'Code' },
          { key: 'subject_name', header: 'Subject' },
          { key: 'present', header: 'Present' },
          { key: 'absent', header: 'Absent' },
          { key: 'late', header: 'Late' },
          { key: 'total', header: 'Total' },
          {
            key: 'percentage', header: 'Attendance %',
            render: (r) => {
              const pct = r.percentage ?? 0
              const variant = pct >= 75 ? 'success' : pct >= 50 ? 'warning' : 'danger'
              return <Badge label={`${pct.toFixed(1)}%`} variant={variant} />
            },
          },
        ]}
      />
    </div>
  )
}