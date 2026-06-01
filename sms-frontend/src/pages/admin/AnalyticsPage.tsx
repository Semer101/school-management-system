import { useCallback, useEffect, useState } from 'react'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell, LineChart, Line, Legend,
} from 'recharts'
import { getAnalytics, type AnalyticsData } from '../../api/analytics'
import { PageHeader } from '../../components/ui/PageHeader'
import { Spinner } from '../../components/ui/Spinner'
import { Button } from '../../components/ui/Button'

const COLORS = ['#3b82f6', '#8b5cf6', '#10b981', '#f59e0b', '#ec4899', '#06b6d4', '#ef4444']

function parseAnalytics(payload: unknown): AnalyticsData | null {
  if (!payload || typeof payload !== 'object') return null
  const p = payload as Record<string, unknown>
  if (p.kpis && p.students_by_grade) return p as unknown as AnalyticsData
  if (p.data && typeof p.data === 'object') return parseAnalytics(p.data)
  return null
}

export default function AnalyticsPage() {
  const [data, setData] = useState<AnalyticsData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    setLoading(true)
    setError('')
    getAnalytics()
      .then((res) => {
        const parsed = parseAnalytics(res.data?.data ?? res.data)
        if (!parsed) {
          setError('Invalid analytics response from server.')
          setData(null)
          return
        }
        setData(parsed)
      })
      .catch(() => {
        setError('Failed to load analytics.')
        setData(null)
      })
      .finally(() => setLoading(false))
  }, [])

  // Only run once on mount — removing location.key which caused the page
  // to flash/disappear every time the router key changed.
  useEffect(() => { load() }, [load])

  if (loading) return <div className="flex justify-center py-20"><Spinner /></div>
  if (!data) {
    return (
      <div className="max-w-6xl mx-auto text-center py-16">
        <p className="text-muted mb-4">{error || 'Failed to load analytics.'}</p>
        <Button onClick={load}>Retry</Button>
      </div>
    )
  }

  const gradeData = data.students_by_grade?.map((g) => ({
    name: `Grade ${g.grade}`,
    count: g.count,
  })) ?? []

  const streamData = data.students_by_stream ?? []
  const gradeAvg = data.grade_averages?.slice(0, 10) ?? []
  const attPie = data.attendance_breakdown ?? []
  const monthly = data.monthly_attendance ?? []

  return (
    <div className="max-w-6xl mx-auto space-y-8">
      <PageHeader title="Analytics" subtitle="School-wide insights — Ethiopian Grades 9–12" />

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        {[
          { label: 'Students', value: data.kpis?.students ?? 0, color: 'border-kpi-students/40 bg-kpi-students/5' },
          { label: 'Teachers', value: data.kpis?.teachers ?? 0, color: 'border-kpi-teachers/40 bg-kpi-teachers/5' },
          { label: 'Revenue (ETB)', value: (data.kpis?.revenue_etb ?? 0).toLocaleString(), color: 'border-kpi-finance/40 bg-kpi-finance/5' },
          { label: 'Notifications', value: data.kpis?.notifications ?? 0, color: 'border-kpi-attendance/40 bg-kpi-attendance/5' },
        ].map((k) => (
          <div key={k.label} className={`rounded-xl border p-4 ${k.color}`}>
            <p className="text-2xl font-bold text-foreground">{k.value}</p>
            <p className="text-xs text-muted">{k.label}</p>
          </div>
        ))}
      </div>

      <div className="grid lg:grid-cols-2 gap-6">
        <ChartCard title="Students by Grade (9–12)">
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={gradeData}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-surface-border)" />
              <XAxis dataKey="name" tick={{ fill: 'var(--color-muted)', fontSize: 12 }} />
              <YAxis tick={{ fill: 'var(--color-muted)', fontSize: 12 }} />
              <Tooltip contentStyle={{ background: 'var(--color-surface)', border: '1px solid var(--color-surface-border)' }} />
              <Bar dataKey="count" fill="#3b82f6" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Stream Distribution (Grades 11–12)">
          <ResponsiveContainer width="100%" height={260}>
            <PieChart>
              <Pie data={streamData} dataKey="count" nameKey="stream" cx="50%" cy="50%" outerRadius={90} label>
                {streamData.map((_, i) => (
                  <Cell key={i} fill={COLORS[i % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Average Grades by Subject">
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={gradeAvg} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis type="number" domain={[0, 100]} tick={{ fontSize: 11 }} />
              <YAxis type="category" dataKey="subject_name" width={100} tick={{ fontSize: 10 }} />
              <Tooltip />
              <Bar dataKey="average" fill="#10b981" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Attendance Breakdown">
          <ResponsiveContainer width="100%" height={260}>
            <PieChart>
              <Pie data={attPie} dataKey="count" nameKey="status" cx="50%" cy="50%" outerRadius={90} label>
                {attPie.map((_, i) => (
                  <Cell key={i} fill={COLORS[i % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Monthly Attendance Trend" className="lg:col-span-2">
          <ResponsiveContainer width="100%" height={280}>
            <LineChart data={monthly}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="month" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="present" stroke="#10b981" strokeWidth={2} />
              <Line type="monotone" dataKey="absent" stroke="#ef4444" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Promotion Status">
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={data.promotion_distribution ?? []}>
              <XAxis dataKey="status" tick={{ fontSize: 11 }} />
              <YAxis />
              <Tooltip />
              <Bar dataKey="count" fill="#8b5cf6" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Finance Overview (ETB)">
          <div className="grid grid-cols-2 gap-4 p-4">
            <Stat label="Verified Revenue" value={data.kpis?.revenue_etb ?? 0} />
            <Stat label="Pending Fees" value={data.kpis?.pending_etb ?? 0} />
            <Stat label="Payroll Paid" value={data.kpis?.payroll_paid ?? 0} />
            <Stat label="Payroll Pending" value={data.kpis?.payroll_pending ?? 0} />
          </div>
        </ChartCard>
      </div>
    </div>
  )
}

function ChartCard({ title, children, className = '' }: { title: string; children: React.ReactNode; className?: string }) {
  return (
    <div className={`rounded-xl border border-surface-border bg-surface p-4 ${className}`}>
      <h3 className="text-sm font-semibold text-foreground mb-4">{title}</h3>
      {children}
    </div>
  )
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg border border-surface-border p-3">
      <p className="text-lg font-bold text-foreground">{Number(value).toLocaleString()} ETB</p>
      <p className="text-xs text-muted">{label}</p>
    </div>
  )
}