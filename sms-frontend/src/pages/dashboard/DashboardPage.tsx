import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  Users, GraduationCap, BookOpen, Layers, ClipboardCheck, Wallet,
  ArrowRight, Bell, Calendar, Award, CheckSquare,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { useAuth } from '../../hooks/useAuth'
import { useRole } from '../../hooks/useRole'
import { useNotifications } from '../../hooks/useNotifications'
import { Badge, roleBadgeVariant } from '../../components/ui/Badge'
import { PageHeader } from '../../components/ui/PageHeader'
import { GlassCard } from '../../components/ui/GlassCard'
import { navIcons, type NavIconKey } from '../../lib/nav-icons'
import { gradeTypeLabel } from '../../lib/grades'
import { cn } from '../../lib/utils'
import { getDashboardKPIs, type DashboardKPIs } from '../../api/analytics'
import { getTeacherDashboardKPIs, getStudentGrades, getAttendancePercentage } from '../../api/academics'
import { getParentDashboardKPIs } from '../../api/parent'
import { getPortalContext } from '../../api/portal'

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

interface KpiCardProps {
  label: string
  value: ReactNode
  icon: LucideIcon
  color: string
  accentBg: string
  borderHover: string
  chipLabel: string
}

function KpiCard({
  label, value, icon: Icon, color, accentBg, borderHover, chipLabel,
}: KpiCardProps) {
  return (
    <motion.div
      whileHover={{ y: -4 }}
      transition={{ type: 'spring', stiffness: 300, damping: 22 }}
      className={cn(
        'relative overflow-hidden rounded-2xl border border-surface-border bg-surface/60 backdrop-blur-sm shadow-glass',
        'transition-all duration-300 group flex flex-col justify-between',
        'p-6 h-52 hover:shadow-glow-sm',
        borderHover
      )}
    >
      <div
        className={cn(
          'absolute -top-10 -right-10 w-32 h-32 rounded-full blur-2xl opacity-30 group-hover:opacity-60 transition-opacity duration-500 pointer-events-none',
          accentBg.replace('/10', '/40').replace('/5', '/30')
        )}
      />

      <div className="flex items-start justify-between relative z-10">
        <div className={cn(
          'flex items-center justify-center rounded-2xl border w-16 h-16',
          accentBg,
          'border-current/20',
          color
        )}>
          <Icon className="w-8 h-8 transition-transform duration-300 group-hover:scale-110" />
        </div>
        <span className="text-[10px] uppercase font-mono tracking-wider text-muted opacity-80 text-right">
          {chipLabel}
        </span>
      </div>

      <div className="relative z-10 mt-4 pb-2">
        <p className="text-4xl font-bold text-foreground leading-none tracking-tight">
          {value}
        </p>
        <p className="text-sm text-muted font-medium mt-2">{label}</p>
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

  const [adminKpis, setAdminKpis] = useState<DashboardKPIs | null>(null)
  const [teacherKpis, setTeacherKpis] = useState<{ students: number; classes: number; subjects: number; attendance_rate: number } | null>(null)
  const [studentKpis, setStudentKpis] = useState<{ gpa: string; attendance: number; midterms: number; finals: number } | null>(null)
  const [parentKpis, setParentKpis] = useState<{ children: number; feePending: number; attendanceAvg: number; gradeAvg: string } | null>(null)

  const [portalYear, setPortalYear] = useState<number | null>(null)
  const [portalSemester, setPortalSemester] = useState<string>('')

  useEffect(() => {
    getPortalContext()
      .then((res) => {
        const ctx = res.data.data
        if (ctx) {
          setPortalYear(ctx.academic_year)
          setPortalSemester(ctx.active_semester)
        }
      })
      .catch(() => {})
  }, [])

  useEffect(() => {
    if (!role) return

    if (role === 'Admin') {
      getDashboardKPIs()
        .then((res) => setAdminKpis(res.data.data as DashboardKPIs))
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
            grade_type?: string
            type?: string
            subject?: { name: string }
            max_score?: number
            maxScore?: number
          }>
          const att = attRes.data.data as { overall_percentage?: number } | undefined

          const validGrades = grades.filter((g) => g.score !== undefined)
          const avg = validGrades.length > 0
            ? validGrades.reduce((sum, g) => {
                const max = g.max_score ?? g.maxScore ?? 100
                return sum + ((g.score / max) * 100)
              }, 0) / validGrades.length
            : 0
          const midtermsCount = validGrades.filter((g) => gradeTypeLabel(g as never) === 'Midterm').length
          const finalsCount = validGrades.filter((g) => gradeTypeLabel(g as never) === 'Final').length

          setStudentKpis({
            gpa: avg > 0 ? `${avg.toFixed(1)}%` : 'N/A',
            attendance: att?.overall_percentage ?? 0,
            midterms: midtermsCount,
            finals: finalsCount,
          })
        })
        .catch(() => {})
    }

    if (role === 'Parent') {
      getParentDashboardKPIs()
        .then((res) => {
          const kpis = res.data.data ?? { children: 0, attendance_avg: 0, grade_avg: '', fee_pending: 0 }
          setParentKpis({
            children: kpis.children,
            feePending: kpis.fee_pending,
            attendanceAvg: kpis.attendance_avg,
            gradeAvg: kpis.grade_avg,
          })
        })
        .catch(() => {})
    }
  }, [role])

  const today = new Date().toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      <PageHeader
        title={`Welcome, ${user?.name?.split(' ')[0] ?? 'User'}`}
        subtitle={user?.email}
        action={role ? <Badge label={role} variant={roleBadgeVariant(role)} /> : undefined}
      />

      <div className="flex flex-col lg:flex-row gap-6">
        <div className="w-full lg:w-[72%] space-y-6 min-w-0">
          <h2 className="text-base font-bold text-foreground tracking-tight flex items-center gap-2">
            <span>Primary Metrics</span>
            <span className="w-1.5 h-1.5 rounded-full bg-accent animate-pulse" />
          </h2>

          {role === 'Admin' && adminKpis && (
            <div className="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-6 gap-4">
              <KpiCard label="Students" value={adminKpis.students} icon={GraduationCap} color="text-kpi-students" accentBg="bg-kpi-students/10" borderHover="hover:border-kpi-students/40" chipLabel="Admin" />
              <KpiCard label="Teachers" value={adminKpis.teachers} icon={Users} color="text-kpi-teachers" accentBg="bg-kpi-teachers/10" borderHover="hover:border-kpi-teachers/40" chipLabel="Admin" />
              <KpiCard label="Classes" value={adminKpis.classes} icon={Layers} color="text-kpi-classes" accentBg="bg-kpi-classes/10" borderHover="hover:border-kpi-classes/40" chipLabel="Admin" />
              <KpiCard label="Subjects" value={adminKpis.subjects} icon={BookOpen} color="text-kpi-subjects" accentBg="bg-kpi-subjects/10" borderHover="hover:border-kpi-subjects/40" chipLabel="Admin" />
              <KpiCard label="Present Today" value={adminKpis.present_today} icon={ClipboardCheck} color="text-kpi-attendance" accentBg="bg-kpi-attendance/10" borderHover="hover:border-kpi-attendance/40" chipLabel="Admin" />
              <KpiCard label="Pending Fees" value={adminKpis.pending_transactions} icon={Wallet} color="text-kpi-finance" accentBg="bg-kpi-finance/10" borderHover="hover:border-kpi-finance/40" chipLabel="Admin" />
            </div>
          )}

          {role === 'Teacher' && (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <KpiCard label="Total Students" value={teacherKpis?.students ?? 0} icon={GraduationCap} color="text-kpi-students" accentBg="bg-kpi-students/10" borderHover="hover:border-kpi-students/40" chipLabel="Teacher" />
              <KpiCard label="Assigned Classes" value={teacherKpis?.classes ?? 0} icon={Layers} color="text-kpi-classes" accentBg="bg-kpi-classes/10" borderHover="hover:border-kpi-classes/40" chipLabel="Teacher" />
              <KpiCard label="Subjects" value={teacherKpis?.subjects ?? 0} icon={BookOpen} color="text-kpi-subjects" accentBg="bg-kpi-subjects/10" borderHover="hover:border-kpi-subjects/40" chipLabel="Teacher" />
              <KpiCard label="Attendance Rate" value={teacherKpis?.attendance_rate ? `${teacherKpis.attendance_rate.toFixed(1)}%` : '0%'} icon={ClipboardCheck} color="text-kpi-attendance" accentBg="bg-kpi-attendance/10" borderHover="hover:border-kpi-attendance/40" chipLabel="Teacher" />
            </div>
          )}

          {role === 'Student' && (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <KpiCard label="GPA / Average" value={studentKpis?.gpa ?? 'N/A'} icon={Award} color="text-kpi-teachers" accentBg="bg-kpi-teachers/10" borderHover="hover:border-kpi-teachers/40" chipLabel="Student" />
              <KpiCard label="Attendance" value={studentKpis?.attendance ? `${studentKpis.attendance.toFixed(1)}%` : '0%'} icon={ClipboardCheck} color="text-kpi-attendance" accentBg="bg-kpi-attendance/10" borderHover="hover:border-kpi-attendance/40" chipLabel="Student" />
              <KpiCard label="Midterm Grades" value={studentKpis?.midterms ?? 0} icon={CheckSquare} color="text-kpi-students" accentBg="bg-kpi-students/10" borderHover="hover:border-kpi-students/40" chipLabel="Student" />
              <KpiCard label="Final Grades" value={studentKpis?.finals ?? 0} icon={Calendar} color="text-kpi-finance" accentBg="bg-kpi-finance/10" borderHover="hover:border-kpi-finance/40" chipLabel="Student" />
            </div>
          )}

          {role === 'Parent' && (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <KpiCard label="Children Count" value={parentKpis?.children ?? 0} icon={GraduationCap} color="text-kpi-students" accentBg="bg-kpi-students/10" borderHover="hover:border-kpi-students/40" chipLabel="Parent" />
              <KpiCard label="Pending Payments" value={parentKpis?.feePending !== undefined ? `${parentKpis.feePending.toLocaleString()} ETB` : '0 ETB'} icon={Wallet} color="text-kpi-finance" accentBg="bg-kpi-finance/10" borderHover="hover:border-kpi-finance/40" chipLabel="Parent" />
              <KpiCard label="Children Attendance" value={parentKpis?.attendanceAvg ? `${parentKpis.attendanceAvg}%` : 'N/A'} icon={ClipboardCheck} color="text-kpi-attendance" accentBg="bg-kpi-attendance/10" borderHover="hover:border-kpi-attendance/40" chipLabel="Parent" />
              <KpiCard label="Academic Performance" value={parentKpis?.gradeAvg ?? 'A'} icon={Award} color="text-kpi-teachers" accentBg="bg-kpi-teachers/10" borderHover="hover:border-kpi-teachers/40" chipLabel="Parent" />
            </div>
          )}
        </div>

        <div className="w-full lg:w-[28%] space-y-6 shrink-0">
          <div className="space-y-3">
            <h2 className="text-base font-bold text-foreground tracking-tight">
              Quick Navigation
            </h2>
            <div className="flex flex-col gap-2">
              {visible.slice(0, 8).map((card) => {
                const Icon = navIcons[card.icon]
                return (
                  <button
                    key={card.to + card.label}
                    type="button"
                    onClick={() => navigate(card.to)}
                    className={cn(
                      'flex items-center gap-2 px-3 py-2 rounded-xl border border-surface-border bg-surface/60 backdrop-blur-sm',
                      'hover:border-accent/40 hover:shadow-glow-sm transition-all duration-200 group justify-start'
                    )}
                  >
                    <Icon className="w-4 h-4 text-accent group-hover:scale-110 transition-transform shrink-0" />
                    <span className="text-xs font-semibold text-foreground tracking-tight">
                      {card.label}
                    </span>
                  </button>
                )
              })}
            </div>
          </div>

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

          <div className="rounded-2xl border border-surface-border bg-surface/40 p-4 space-y-2">
            <p className="text-[10px] uppercase font-mono tracking-widest text-muted">Portal Info</p>
            <p className="text-xs text-muted leading-relaxed">
              Today: <strong className="text-foreground">{today}</strong>
            </p>
            <p className="text-xs text-muted leading-relaxed">
              Academic Year: <strong className="text-foreground">{portalYear ?? '—'}</strong>
            </p>
            <p className="text-xs text-muted leading-relaxed">
              Active Semester: <strong className="text-foreground">{portalSemester || '—'}</strong>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
