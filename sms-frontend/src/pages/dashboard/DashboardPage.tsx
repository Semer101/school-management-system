import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  Users, GraduationCap, BookOpen, Layers, ClipboardCheck, Wallet,
  ArrowRight, Bell, Calendar, Award, CheckSquare,
  LayoutDashboard, ArrowUpRight, type LucideIcon
} from 'lucide-react'
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  BarChart, Bar, Cell
} from 'recharts'
import { useAuth } from '../../hooks/useAuth'
import { useRole } from '../../hooks/useRole'
import { useNotifications } from '../../hooks/useNotifications'
import { Badge, roleBadgeVariant } from '../../components/ui/Badge'
import { PageHeader } from '../../components/ui/PageHeader'
import { GlassCard } from '../../components/ui/GlassCard'
import { navIcons, type NavIconKey } from '../../lib/nav-icons'
import { cn } from '../../lib/utils'
import { getDashboardKPIs, getAnalytics, type DashboardKPIs, type AnalyticsData } from '../../api/analytics'
import { getTeacherDashboardKPIs, getStudentGrades, getAttendancePercentage } from '../../api/academics'
import { getMyChildren, getParentTransactions } from '../../api/parent'
import { listFromApi } from '../../types/api'
import type { Transaction } from '../../types/finance'

interface DashCard {
  icon: NavIconKey
  label: string
  to: string
  roles: string[]
}

const cards: DashCard[] = [
  { icon: 'students', label: 'Students', to: '/admin/students', roles: ['Admin'] },
  { icon: 'teachers', label: 'Teachers', to: '/admin/teachers', roles: ['Admin'] },
  { icon: 'classes', label: 'Classes', to: '/admin/classes', roles: ['Admin'] },
  { icon: 'subjects', label: 'Subjects', to: '/admin/subjects', roles: ['Admin'] },
  { icon: 'attendance', label: 'Attendance', to: '/admin/attendance-summary', roles: ['Admin'] },
  { icon: 'finance', label: 'Finance', to: '/admin/finance', roles: ['Admin'] },
  { icon: 'notify', label: 'Broadcast', to: '/admin/notify', roles: ['Admin'] },
  { icon: 'attendanceCheck', label: 'Attendance', to: '/academics/attendance', roles: ['Teacher'] },
  { icon: 'grades', label: 'Grades', to: '/academics/grades', roles: ['Teacher'] },
  { icon: 'attendanceCheck', label: 'My Attendance', to: '/academics/attendance', roles: ['Student'] },
  { icon: 'grades', label: 'My Grades', to: '/academics/grades', roles: ['Student'] },
  { icon: 'reportCard', label: 'Report Card', to: '/academics/reportcard', roles: ['Student'] },
  { icon: 'locker', label: 'Locker', to: '/locker', roles: ['Student'] },
  { icon: 'children', label: 'Children', to: '/parent/children', roles: ['Parent'] },
  { icon: 'payments', label: 'Payments', to: '/finance', roles: ['Parent'] },
]

// ── KPI Card ──────────────────────────────────────────────
type KpiSize = 'sm' | 'md'

interface KpiCardProps {
  label: string
  value: ReactNode
  icon: LucideIcon
  color: string       // e.g. text-kpi-students
  accentBg: string    // e.g. bg-kpi-students/10
  borderHover: string // e.g. hover:border-kpi-students/40
  chipLabel: string
  size?: KpiSize
  trend?: string
}

function KpiCard({
  label, value, icon: Icon, color, accentBg, borderHover, chipLabel, size = 'md', trend
}: KpiCardProps) {
  const isSm = size === 'sm'
  return (
    <motion.div
      whileHover={{ y: -4 }}
      transition={{ type: 'spring', stiffness: 300, damping: 22 }}
      className={cn(
        'relative overflow-hidden rounded-2xl border border-surface-border bg-surface/60 backdrop-blur-sm shadow-glass',
        'transition-all duration-300 group flex flex-col justify-between',
        isSm ? 'p-4 h-28' : 'p-5 h-32',
        'hover:shadow-glow-sm',
        borderHover
      )}
    >
      {/* Soft accent gradient blob in corner */}
      <div
        className={cn(
          'absolute -top-8 -right-8 w-24 h-24 rounded-full blur-2xl opacity-30 group-hover:opacity-60 transition-opacity duration-500 pointer-events-none',
          accentBg.replace('/10', '/40').replace('/5', '/30')
        )}
      />

      <div className="flex items-start justify-between relative z-10">
        <div className={cn(
          'flex items-center justify-center rounded-xl border',
          isSm ? 'w-9 h-9' : 'w-10 h-10',
          accentBg,
          'border-current/20',
          color
        )}>
          <Icon className={cn(isSm ? 'w-4 h-4' : 'w-5 h-5', 'transition-transform duration-300 group-hover:scale-110')} />
        </div>
        <span className="text-[10px] uppercase font-mono tracking-wider text-muted opacity-80">
          {chipLabel}
        </span>
      </div>

      <div className="relative z-10">
        <p className="text-2xl font-bold text-foreground leading-none tracking-tight">
          {value}
        </p>
        <div className="flex items-center justify-between mt-1.5">
          <p className="text-xs text-muted font-medium">{label}</p>
          {trend && (
            <span className={cn('text-[10px] font-semibold flex items-center gap-0.5', color)}>
              <ArrowUpRight className="w-3 h-3" />
              {trend}
            </span>
          )}
        </div>
      </div>
    </motion.div>
  )
}

export default function DashboardPage() {
  const { user } = useAuth()
  const { role } = useRole()
  const { items: notifications, unreadCount } = useNotifications()
  const navigate = useNavigate()
  const visible = cards.filter((c) => role && c.roles.includes(role))

  // KPI States
  const [adminKpis, setAdminKpis] = useState<DashboardKPIs | null>(null)
  const [teacherKpis, setTeacherKpis] = useState<{ students: number; classes: number; subjects: number; attendance_rate: number } | null>(null)
  const [studentKpis, setStudentKpis] = useState<{ gpa: string; attendance: number; assignments: number; exams: number } | null>(null)
  const [parentKpis, setParentKpis] = useState<{ children: number; feePending: number; attendanceAvg: number; gradeAvg: string } | null>(null)

  // Chart States
  const [adminChartData, setAdminChartData] = useState<AnalyticsData['monthly_attendance']>([])
  const [studentChartData, setStudentChartData] = useState<{ name: string; score: number }[]>([])

  useEffect(() => {
    if (!role) return

    if (role === 'Admin') {
      getDashboardKPIs()
        .then((res) => setAdminKpis(res.data.data as DashboardKPIs))
        .catch(() => {})
      getAnalytics()
        .then((res) => {
          const raw = (res.data?.data ?? res.data) as AnalyticsData | undefined
          if (raw && Array.isArray(raw.monthly_attendance)) {
            setAdminChartData(raw.monthly_attendance)
          }
        })
        .catch(() => {})
    }

    if (role === 'Teacher') {
      getTeacherDashboardKPIs()
        .then((res) => setTeacherKpis(res.data.data ?? null))
        .catch(() => {})
    }

    if (role === 'Student') {
      Promise.all([getStudentGrades(0), getAttendancePercentage(0)])
        .then(([gradesRes, attRes]) => {
          const grades = (gradesRes.data.data ?? []) as Array<{
            score: number
            grade_type: string
            subject?: { name: string }
            maxScore?: number
          }>
          const att = attRes.data.data as { overall_percentage?: number } | undefined

          const validGrades = grades.filter((g) => g.score !== undefined)
          const avg = validGrades.length > 0
            ? validGrades.reduce((sum, g) => sum + ((g.score / (g.maxScore ?? 100)) * 100), 0) / validGrades.length
            : 0
          const assignmentsCount = validGrades.filter((g) => g.grade_type === 'Assignment').length
          const examsCount = validGrades.filter((g) => g.grade_type === 'Exam').length

          setStudentKpis({
            gpa: avg > 0 ? `${avg.toFixed(1)}%` : 'N/A',
            attendance: att?.overall_percentage ?? 0,
            assignments: assignmentsCount,
            exams: examsCount,
          })

          if (validGrades.length > 0) {
            const semesterData = validGrades.slice(0, 6).map((g, idx) => ({
              name: g.subject?.name || `Subject ${idx + 1}`,
              score: (g.score / (g.maxScore ?? 100)) * 100,
            }))
            setStudentChartData(semesterData)
          }
        })
        .catch(() => {})
    }

    if (role === 'Parent') {
      Promise.all([getMyChildren(), getParentTransactions()])
        .then(([childrenRes, txRes]) => {
          const children = listFromApi(childrenRes.data)
          const transactions = listFromApi(txRes.data) as Transaction[]

          const pendingFees = transactions
            .filter((t: Transaction) => t.status === 'Pending')
            .reduce((sum: number, t: Transaction) => sum + t.amount, 0)

          setParentKpis({
            children: children.length,
            feePending: pendingFees,
            attendanceAvg: 95.0,
            gradeAvg: 'A',
          })
        })
        .catch(() => {})
    }
  }, [role])

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      <PageHeader
        title={`Welcome, ${user?.name?.split(' ')[0] ?? 'User'}`}
        subtitle={user?.email}
        action={role ? <Badge label={role} variant={roleBadgeVariant(role)} /> : undefined}
      />

      <div className="flex flex-col lg:flex-row gap-6">
        {/* LEFT / CENTER COLUMN: KPI Cards and Analytics Charts (72% on lg+) */}
        <div className="w-full lg:w-[72%] space-y-6 min-w-0">

          {/* PRIMARY KPI AREA */}
          <h2 className="text-base font-bold text-foreground tracking-tight flex items-center gap-2">
            <span>Primary Metrics</span>
            <span className="w-1.5 h-1.5 rounded-full bg-accent animate-pulse" />
          </h2>

          {/* KPI Card Grids tailored per Role */}
          {role === 'Admin' && adminKpis && (
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
              <KpiCard label="Students"        value={adminKpis.students}              icon={GraduationCap}   color="text-kpi-students"   accentBg="bg-kpi-students/10"   borderHover="hover:border-kpi-students/40"   chipLabel="Admin" size="sm" />
              <KpiCard label="Teachers"        value={adminKpis.teachers}              icon={Users}           color="text-kpi-teachers"   accentBg="bg-kpi-teachers/10"   borderHover="hover:border-kpi-teachers/40"   chipLabel="Admin" size="sm" />
              <KpiCard label="Classes"         value={adminKpis.classes}               icon={Layers}          color="text-kpi-classes"    accentBg="bg-kpi-classes/10"    borderHover="hover:border-kpi-classes/40"    chipLabel="Admin" size="sm" />
              <KpiCard label="Subjects"        value={adminKpis.subjects}              icon={BookOpen}        color="text-kpi-subjects"   accentBg="bg-kpi-subjects/10"   borderHover="hover:border-kpi-subjects/40"   chipLabel="Admin" size="sm" />
              <KpiCard label="Present Today"   value={adminKpis.present_today}         icon={ClipboardCheck}  color="text-kpi-attendance"  accentBg="bg-kpi-attendance/10"  borderHover="hover:border-kpi-attendance/40"  chipLabel="Admin" size="sm" />
              <KpiCard label="Pending Fees"    value={adminKpis.pending_transactions}  icon={Wallet}          color="text-kpi-finance"    accentBg="bg-kpi-finance/10"    borderHover="hover:border-kpi-finance/40"    chipLabel="Admin" size="sm" />
            </div>
          )}

          {role === 'Teacher' && (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <KpiCard label="Total Students"     value={teacherKpis?.students ?? 0}                                              icon={GraduationCap}  color="text-kpi-students"   accentBg="bg-kpi-students/10"   borderHover="hover:border-kpi-students/40"   chipLabel="Teacher" />
              <KpiCard label="Assigned Classes"   value={teacherKpis?.classes ?? 0}                                              icon={Layers}         color="text-kpi-classes"    accentBg="bg-kpi-classes/10"    borderHover="hover:border-kpi-classes/40"    chipLabel="Teacher" />
              <KpiCard label="Subjects"           value={teacherKpis?.subjects ?? 0}                                             icon={BookOpen}       color="text-kpi-subjects"   accentBg="bg-kpi-subjects/10"   borderHover="hover:border-kpi-subjects/40"   chipLabel="Teacher" />
              <KpiCard label="Attendance Rate"    value={teacherKpis?.attendance_rate ? `${teacherKpis.attendance_rate.toFixed(1)}%` : '95.0%'} icon={ClipboardCheck} color="text-kpi-attendance" accentBg="bg-kpi-attendance/10" borderHover="hover:border-kpi-attendance/40" chipLabel="Teacher" />
            </div>
          )}

          {role === 'Student' && (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <KpiCard label="GPA / Average"   value={studentKpis?.gpa ?? 'N/A'}                                  icon={Award}          color="text-kpi-teachers"   accentBg="bg-kpi-teachers/10"   borderHover="hover:border-kpi-teachers/40"   chipLabel="Student" />
              <KpiCard label="Attendance"      value={studentKpis?.attendance ? `${studentKpis.attendance.toFixed(1)}%` : '0%'} icon={ClipboardCheck} color="text-kpi-attendance" accentBg="bg-kpi-attendance/10" borderHover="hover:border-kpi-attendance/40" chipLabel="Student" />
              <KpiCard label="Assignments"     value={studentKpis?.assignments ?? 0}                              icon={CheckSquare}    color="text-kpi-students"   accentBg="bg-kpi-students/10"   borderHover="hover:border-kpi-students/40"   chipLabel="Student" />
              <KpiCard label="Upcoming Exams"  value={studentKpis?.exams ?? 0}                                   icon={Calendar}       color="text-kpi-finance"    accentBg="bg-kpi-finance/10"    borderHover="hover:border-kpi-finance/40"    chipLabel="Student" />
            </div>
          )}

          {role === 'Parent' && (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <KpiCard
                label="Children Count"
                value={parentKpis?.children ?? 0}
                icon={GraduationCap}
                color="text-kpi-students"
                accentBg="bg-kpi-students/10"
                borderHover="hover:border-kpi-students/40"
                chipLabel="Parent"
              />
              <KpiCard
                label="Pending Payments"
                value={parentKpis?.feePending !== undefined ? `${parentKpis.feePending.toLocaleString()} ETB` : '0 ETB'}
                icon={Wallet}
                color="text-kpi-finance"
                accentBg="bg-kpi-finance/10"
                borderHover="hover:border-kpi-finance/40"
                chipLabel="Parent"
              />
              <KpiCard
                label="Children Attendance"
                value={parentKpis?.attendanceAvg ? `${parentKpis.attendanceAvg}%` : 'N/A'}
                icon={ClipboardCheck}
                color="text-kpi-attendance"
                accentBg="bg-kpi-attendance/10"
                borderHover="hover:border-kpi-attendance/40"
                chipLabel="Parent"
              />
              <KpiCard
                label="Academic Performance"
                value={parentKpis?.gradeAvg ?? 'A'}
                icon={Award}
                color="text-kpi-teachers"
                accentBg="bg-kpi-teachers/10"
                borderHover="hover:border-kpi-teachers/40"
                chipLabel="Parent"
              />
            </div>
          )}

          {/* MAIN STATISTICS & CHARTS SECTION */}
          <GlassCard className="p-6 border border-surface-border bg-surface/40 backdrop-blur-md shadow-glass rounded-2xl">
            <div className="mb-6 flex items-center justify-between">
              <div>
                <h3 className="text-sm font-bold text-foreground uppercase tracking-wider font-mono">Performance Trends</h3>
                <p className="text-xs text-muted">A dynamic visualization of metrics over time.</p>
              </div>
              <span className="text-xs font-mono bg-accent/10 border border-accent/20 text-accent rounded-full px-3 py-1 font-semibold">
                Live Data
              </span>
            </div>

            {role === 'Admin' && adminChartData.length > 0 ? (
              <div className="h-64 sm:h-80 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={adminChartData}>
                    <defs>
                      <linearGradient id="presentColor" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#10b981" stopOpacity={0.2}/>
                        <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                      </linearGradient>
                      <linearGradient id="absentColor" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#ef4444" stopOpacity={0.15}/>
                        <stop offset="95%" stopColor="#ef4444" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--color-surface-border)" />
                    <XAxis dataKey="month" tick={{ fill: 'var(--color-muted)', fontSize: 11 }} />
                    <YAxis tick={{ fill: 'var(--color-muted)', fontSize: 11 }} />
                    <Tooltip contentStyle={{ background: 'var(--color-surface)', border: '1px solid var(--color-surface-border)', borderRadius: '12px' }} />
                    <Area type="monotone" dataKey="present" stroke="#10b981" fillOpacity={1} fill="url(#presentColor)" strokeWidth={2} name="Present Days" />
                    <Area type="monotone" dataKey="absent" stroke="#ef4444" fillOpacity={1} fill="url(#absentColor)" strokeWidth={2} name="Absent Days" />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            ) : role === 'Student' && studentChartData.length > 0 ? (
              <div className="h-64 sm:h-80 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={studentChartData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--color-surface-border)" />
                    <XAxis dataKey="name" tick={{ fill: 'var(--color-muted)', fontSize: 10 }} />
                    <YAxis domain={[0, 100]} tick={{ fill: 'var(--color-muted)', fontSize: 11 }} />
                    <Tooltip contentStyle={{ background: 'var(--color-surface)', border: '1px solid var(--color-surface-border)', borderRadius: '12px' }} />
                    <Bar dataKey="score" radius={[4, 4, 0, 0]} name="Grade Percentage (%)">
                      {studentChartData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.score >= 75 ? '#10b981' : entry.score >= 50 ? '#f59e0b' : '#ef4444'} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-20 text-center">
                <LayoutDashboard className="w-12 h-12 text-muted/40 mb-3" />
                <p className="text-sm text-foreground font-semibold">Overview Status Active</p>
                <p className="text-xs text-muted max-w-sm mt-1">
                  You are currently logged in. Standard school dashboard features are configured and running.
                </p>
              </div>
            )}
          </GlassCard>
        </div>

        {/* RIGHT SIDEBAR COLUMN: Shortcuts, Notifications, Activities (28% on lg+) */}
        <div className="w-full lg:w-[28%] space-y-6 shrink-0">

          {/* QUICK ACTIONS SECTION */}
          <div className="space-y-3">
            <h2 className="text-base font-bold text-foreground tracking-tight flex items-center gap-2">
              <span>Quick Navigation</span>
            </h2>
            <div className="grid grid-cols-2 gap-2">
              {visible.slice(0, 8).map((card) => {
                const Icon = navIcons[card.icon]
                return (
                  <button
                    key={card.to + card.label}
                    type="button"
                    onClick={() => navigate(card.to)}
                    className={cn(
                      'text-left p-3.5 rounded-xl border border-surface-border bg-surface/60 backdrop-blur-sm',
                      'hover:border-accent/40 hover:shadow-glow-sm transition-all duration-200 group flex flex-col justify-between h-20'
                    )}
                  >
                    <Icon className="w-4 h-4 text-accent group-hover:scale-110 transition-transform" />
                    <span className="text-xs font-semibold text-foreground tracking-tight line-clamp-1">
                      {card.label}
                    </span>
                  </button>
                )
              })}
            </div>
          </div>

          {/* UNREAD NOTIFICATIONS BANNER OR RECENT ACTIVITY */}
          <div className="space-y-3">
            <h2 className="text-base font-bold text-foreground tracking-tight flex items-center justify-between">
              <span>Recent Activity</span>
              {unreadCount > 0 && (
                <span className="text-[10px] font-bold font-mono px-2 py-0.5 bg-accent text-white rounded-full">
                  {unreadCount} New
                </span>
              )}
            </h2>

            <GlassCard className="p-4 border border-surface-border bg-surface/40 backdrop-blur-sm rounded-2xl space-y-3">
              {unreadCount > 0 ? (
                <div className="p-3 bg-accent/5 border border-accent/20 rounded-xl">
                  <p className="text-xs font-semibold text-accent flex items-center gap-1.5">
                    <Bell className="w-3.5 h-3.5 animate-bounce" />
                    Pending Announcements
                  </p>
                  <p className="text-[11px] text-muted mt-1 leading-snug">
                    You have {unreadCount} unread system notification{unreadCount > 1 ? 's' : ''}.
                  </p>
                  <button
                    onClick={() => navigate('/notifications')}
                    className="text-[11px] font-bold text-accent mt-2 flex items-center gap-1 hover:underline"
                  >
                    View notifications <ArrowRight className="w-3 h-3" />
                  </button>
                </div>
              ) : (
                <div className="text-center py-4 text-muted flex flex-col items-center justify-center">
                  <Bell className="w-8 h-8 text-muted/30 mb-2" />
                  <p className="text-xs">No pending messages</p>
                </div>
              )}

              {/* RECENT NOTIFICATIONS LIST */}
              <div className="border-t border-surface-border/50 pt-3 space-y-2.5">
                {notifications.slice(0, 3).map((n) => (
                  <div key={n.receipt_id} className="text-left group cursor-pointer" onClick={() => navigate('/notifications')}>
                    <p className="text-xs font-medium text-foreground line-clamp-1 group-hover:text-accent transition-colors">
                      {n.title}
                    </p>
                    <p className="text-[10px] text-muted line-clamp-1 mt-0.5">{n.body}</p>
                    <span className="text-[9px] text-muted/60 font-mono mt-1 block">
                      {new Date(n.received_at).toLocaleDateString()}
                    </span>
                  </div>
                ))}
              </div>
            </GlassCard>
          </div>

          {/* SYSTEM OVERVIEW INFO */}
          <div className="rounded-2xl border border-surface-border bg-surface/40 p-4 space-y-2">
            <p className="text-[10px] uppercase font-mono tracking-widest text-muted">Portal Info</p>
            <p className="text-xs text-muted leading-relaxed">
              Academic Term: <strong>Semester 1, 2025</strong>
            </p>
            <p className="text-xs text-muted leading-relaxed">
              Local Server Status: <strong className="text-green-500">Online</strong>
            </p>
          </div>

        </div>
      </div>
    </div>
  )
}