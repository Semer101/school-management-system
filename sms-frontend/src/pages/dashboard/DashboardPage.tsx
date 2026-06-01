import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  Users, GraduationCap, BookOpen, Layers, ClipboardCheck, Wallet,
} from 'lucide-react'
import { useAuth } from '../../hooks/useAuth'
import { useRole } from '../../hooks/useRole'
import { useNotifications } from '../../hooks/useNotifications'
import { Badge, roleBadgeVariant } from '../../components/ui/Badge'
import { PageHeader } from '../../components/ui/PageHeader'
import { GlassCard } from '../../components/ui/GlassCard'
import { navIcons, type NavIconKey } from '../../lib/nav-icons'
import { cn } from '../../lib/utils'
import { getDashboardKPIs, type DashboardKPIs } from '../../api/analytics'

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
  { icon: 'attendanceCheck', label: 'Attendance', to: '/academics/attendance', roles: ['Teacher', 'Student'] },
  { icon: 'grades', label: 'Grades', to: '/academics/grades', roles: ['Teacher', 'Student'] },
  { icon: 'reportCard', label: 'Report Card', to: '/academics/reportcard', roles: ['Student'] },
  { icon: 'payments', label: 'Finance', to: '/finance', roles: ['Student', 'Parent'] },
  { icon: 'locker', label: 'Locker', to: '/locker', roles: ['Student'] },
  { icon: 'children', label: 'Children', to: '/parent/children', roles: ['Parent'] },
]

const kpiConfig: { key: keyof DashboardKPIs; label: string; icon: typeof Users; color: string; bg: string }[] = [
  { key: 'students', label: 'Students', icon: GraduationCap, color: 'text-kpi-students', bg: 'bg-kpi-students/10 border-kpi-students/30' },
  { key: 'teachers', label: 'Teachers', icon: Users, color: 'text-kpi-teachers', bg: 'bg-kpi-teachers/10 border-kpi-teachers/30' },
  { key: 'classes', label: 'Classes', icon: Layers, color: 'text-kpi-classes', bg: 'bg-kpi-classes/10 border-kpi-classes/30' },
  { key: 'subjects', label: 'Subjects', icon: BookOpen, color: 'text-kpi-subjects', bg: 'bg-kpi-subjects/10 border-kpi-subjects/30' },
  { key: 'present_today', label: 'Present Today', icon: ClipboardCheck, color: 'text-kpi-attendance', bg: 'bg-kpi-attendance/10 border-kpi-attendance/30' },
  { key: 'pending_transactions', label: 'Pending Payments', icon: Wallet, color: 'text-kpi-finance', bg: 'bg-kpi-finance/10 border-kpi-finance/30' },
]

export default function DashboardPage() {
  const { user } = useAuth()
  const { role } = useRole()
  const { unreadCount } = useNotifications()
  const navigate = useNavigate()
  const visible = cards.filter((c) => role && c.roles.includes(role))
  const [kpis, setKpis] = useState<DashboardKPIs | null>(null)

  useEffect(() => {
    if (role !== 'Admin') return
    getDashboardKPIs()
      .then((res) => setKpis(res.data.data as DashboardKPIs))
      .catch(() => {})
  }, [role])

  return (
    <div className="max-w-6xl mx-auto">
      <PageHeader
        title={`Welcome, ${user?.name?.split(' ')[0] ?? 'User'}`}
        subtitle={user?.email}
        action={role ? <Badge label={role} variant={roleBadgeVariant(role)} /> : undefined}
      />

      {unreadCount > 0 && (
        <GlassCard hover className="mb-6 px-4 py-3 cursor-pointer">
          <button type="button" className="w-full text-left" onClick={() => navigate('/notifications')}>
            <p className="text-sm text-accent font-medium">
              {unreadCount} unread notification{unreadCount > 1 ? 's' : ''}
            </p>
          </button>
        </GlassCard>
      )}

      {role === 'Admin' && kpis && (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-8">
          {kpiConfig.map(({ key, label, icon: Icon, color, bg }) => (
            <motion.div
              key={key}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              className={cn('rounded-xl border p-4', bg)}
            >
              <Icon className={cn('w-5 h-5 mb-2', color)} />
              <p className="text-2xl font-bold text-foreground">{kpis[key] ?? 0}</p>
              <p className="text-xs text-muted mt-1">{label}</p>
            </motion.div>
          ))}
        </div>
      )}

      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3"
      >
        {visible.map((card) => {
          const Icon = navIcons[card.icon]
          return (
            <motion.button
              key={card.to + card.label}
              type="button"
              onClick={() => navigate(card.to)}
              className={cn(
                'text-left p-4 rounded-xl border border-surface-border bg-surface/60 backdrop-blur-sm',
                'hover:border-accent/40 hover:shadow-glow-sm transition-all duration-200 group'
              )}
            >
              <Icon className="w-5 h-5 text-accent mb-3 group-hover:scale-110 transition-transform" />
              <span className="text-sm font-medium text-foreground">{card.label}</span>
            </motion.button>
          )
        })}
      </motion.div>
    </div>
  )
}
