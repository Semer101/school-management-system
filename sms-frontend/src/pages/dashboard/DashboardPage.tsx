import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useAuth } from '../../hooks/useAuth'
import { useRole } from '../../hooks/useRole'
import { useNotifications } from '../../hooks/useNotifications'
import { Badge, roleBadgeVariant } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { PageHeader } from '../../components/ui/PageHeader'
import { GlassCard } from '../../components/ui/GlassCard'
import { navIcons, type NavIconKey } from '../../lib/nav-icons'
import { cn } from '../../lib/utils'

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
  { icon: 'enrollment', label: 'Enrollment', to: '/admin/enrollment', roles: ['Admin'] },
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

const container = {
  hidden: { opacity: 0 },
  show: { opacity: 1, transition: { staggerChildren: 0.05 } },
}

const item = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0 },
}

export default function DashboardPage() {
  const { user } = useAuth()
  const { role } = useRole()
  const { unreadCount } = useNotifications()
  const navigate = useNavigate()
  const visible = cards.filter((c) => role && c.roles.includes(role))

  return (
    <div className="max-w-5xl mx-auto">
      <PageHeader
        title={`Welcome, ${user?.name?.split(' ')[0] ?? 'User'}`}
        subtitle={user?.email}
        action={role ? <Badge label={role} variant={roleBadgeVariant(role)} /> : undefined}
      />

      {unreadCount > 0 && (
        <GlassCard
          hover
          className="mb-6 px-4 py-3 flex items-center justify-between cursor-pointer"
        >
          <motion.div
            role="button"
            tabIndex={0}
            onClick={() => navigate('/notifications')}
            onKeyDown={(e) => e.key === 'Enter' && navigate('/notifications')}
            className="flex items-center justify-between w-full"
          >
            <p className="text-sm text-accent font-medium">
              {unreadCount} unread notification{unreadCount > 1 ? 's' : ''}
            </p>
            <span className="text-xs text-accent font-mono">VIEW</span>
          </motion.div>
        </GlassCard>
      )}

      <motion.div
        variants={container}
        initial="hidden"
        animate="show"
        className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3"
      >
        {visible.map((card) => {
          const Icon = navIcons[card.icon]
          return (
            <motion.button
              key={card.to + card.label}
              variants={item}
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

      <div className="mt-8 flex justify-end">
        <Button variant="ghost" size="sm" onClick={() => navigate('/profile')}>
          Profile settings
        </Button>
      </div>
    </div>
  )
}
