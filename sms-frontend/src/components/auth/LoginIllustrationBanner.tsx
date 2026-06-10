import { useEffect, useState, type ReactNode } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Bell,
  BarChart3,
  MessageCircle,
  GraduationCap,
  CalendarCheck,
  Users,
  BookOpen,
  Mail,
} from 'lucide-react'
import { cn } from '../../lib/utils'

const SLIDES = [
  {
    title: 'Elevating Academic Excellence',
    description:
      'Welcome to the digital heart of our school. Connect with students, teachers, parents, and administrative tools in one unified workspace.',
  },
  {
    title: 'Unified School Operations',
    description:
      'Manage attendance, grades, finance, and communications from a single premium dashboard built for Ethiopian G9–12 education.',
  },
  {
    title: 'Connect Every Stakeholder',
    description:
      'Keep parents informed, empower teachers, and give administrators real-time visibility across your entire institution.',
  },
] as const

function floatAnimation(delay: number) {
  return {
    y: [0, -6 - delay * 2, 0],
    transition: {
      duration: 3.5 + delay * 0.4,
      repeat: Infinity,
      ease: 'easeInOut' as const,
    },
  }
}

function FloatingCard({
  children,
  className,
  delay = 0,
}: {
  children: ReactNode
  className?: string
  delay?: number
}) {
  return (
    <motion.div
      animate={floatAnimation(delay)}
      className={cn(
        'absolute rounded-xl border border-white/25 bg-white/15 backdrop-blur-md shadow-[0_8px_32px_rgba(0,0,0,0.12)]',
        className,
      )}
    >
      {children}
    </motion.div>
  )
}

function CentralIllustration() {
  return (
    <svg
      viewBox="0 0 420 340"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="w-full max-w-md h-auto drop-shadow-[0_16px_40px_rgba(8,145,178,0.3)]"
      aria-hidden="true"
    >
      {/* Decorative wave lines */}
      <path
        d="M40 200 Q80 180 120 200 T200 200"
        stroke="rgba(255,255,255,0.25)"
        strokeWidth="1.5"
        fill="none"
      />
      <path
        d="M300 210 Q340 190 380 210"
        stroke="rgba(255,255,255,0.2)"
        strokeWidth="1.5"
        fill="none"
      />

      {/* Monitor base */}
      <rect x="118" y="248" width="184" height="12" rx="4" fill="rgba(255,255,255,0.15)" />
      <rect x="198" y="260" width="24" height="8" rx="2" fill="rgba(255,255,255,0.1)" />

      {/* Monitor screen */}
      <rect x="108" y="148" width="204" height="104" rx="10" fill="#0e7490" />
      <rect x="116" y="156" width="188" height="88" rx="6" fill="#164e63" />

      {/* Screen UI mockup */}
      <rect x="128" y="168" width="60" height="6" rx="2" fill="rgba(34,211,238,0.5)" />
      <rect x="128" y="182" width="148" height="4" rx="1" fill="rgba(255,255,255,0.15)" />
      <rect x="128" y="192" width="120" height="4" rx="1" fill="rgba(255,255,255,0.1)" />
      <rect x="128" y="210" width="40" height="22" rx="4" fill="rgba(34,211,238,0.25)" />
      <rect x="176" y="210" width="40" height="22" rx="4" fill="rgba(255,255,255,0.08)" />
      <rect x="224" y="210" width="40" height="22" rx="4" fill="rgba(255,255,255,0.08)" />

      {/* Person — head */}
      <circle cx="210" cy="108" r="28" fill="#f8fafc" />
      {/* Hair */}
      <path
        d="M182 100 C182 78 198 68 210 68 C222 68 238 78 238 100 C238 92 230 86 210 86 C190 86 182 92 182 100Z"
        fill="#0f172a"
      />
      {/* Headset */}
      <path
        d="M178 108 C178 92 192 82 210 82 C228 82 242 92 242 108"
        stroke="#0891b2"
        strokeWidth="3"
        fill="none"
      />
      <rect x="172" y="104" width="10" height="16" rx="4" fill="#0891b2" />
      <rect x="238" y="104" width="10" height="16" rx="4" fill="#0891b2" />
      <path d="M248 112 L258 114 L256 120" stroke="#0891b2" strokeWidth="2" fill="none" />

      {/* Body / shoulders */}
      <path
        d="M168 136 C168 136 178 128 210 128 C242 128 252 136 252 136 L260 200 C260 210 248 218 210 218 C172 218 160 210 160 200Z"
        fill="#0891b2"
      />
      <path
        d="M190 148 L210 168 L230 148"
        stroke="rgba(255,255,255,0.3)"
        strokeWidth="2"
        fill="none"
      />

      {/* OK hand gesture */}
      <circle cx="148" cy="172" r="14" fill="#f8fafc" />
      <circle cx="140" cy="164" r="5" fill="#f8fafc" />
      <circle cx="156" cy="164" r="5" fill="#f8fafc" />
      <path
        d="M148 178 C148 186 140 190 136 186"
        stroke="#f8fafc"
        strokeWidth="3"
        fill="none"
        strokeLinecap="round"
      />

      {/* Check badge */}
      <circle cx="88" cy="188" r="18" fill="white" />
      <path
        d="M80 188 L86 194 L96 182"
        stroke="#0891b2"
        strokeWidth="2.5"
        fill="none"
        strokeLinecap="round"
        strokeLinejoin="round"
      />

      {/* Decorative dots */}
      <circle cx="340" cy="120" r="3" fill="rgba(255,255,255,0.3)" />
      <circle cx="352" cy="132" r="2" fill="rgba(255,255,255,0.2)" />
      <circle cx="60" cy="140" r="2" fill="rgba(255,255,255,0.25)" />
      <circle cx="72" cy="152" r="3" fill="rgba(255,255,255,0.15)" />
    </svg>
  )
}

export function LoginIllustrationBanner({ className }: { className?: string }) {
  const [activeSlide, setActiveSlide] = useState(0)

  useEffect(() => {
    const timer = setInterval(() => {
      setActiveSlide((prev) => (prev + 1) % SLIDES.length)
    }, 6000)
    return () => clearInterval(timer)
  }, [])

  return (
    <aside
      className={cn(
        'hidden md:flex md:w-1/2 lg:w-3/5 flex-col justify-between p-8 lg:p-12 text-white relative z-10 overflow-hidden',
        'border-l border-white/10',
        className,
      )}
      aria-label="Platform highlights"
    >
      {/* Gradient background */}
      <div className="absolute inset-0 bg-gradient-to-br from-accent via-cyan-700 to-cyan-950 dark:from-cyan-950 dark:via-cyan-900 dark:to-void" />

      {/* Radial glow */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_40%,rgba(34,211,238,0.22),transparent_55%)] pointer-events-none" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_80%_70%,rgba(8,145,178,0.18),transparent_50%)] pointer-events-none" />

      {/* Grid pattern */}
      <div
        className="absolute inset-0 opacity-[0.07] pointer-events-none"
        style={{
          backgroundImage:
            'linear-gradient(rgba(255,255,255,0.4) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.4) 1px, transparent 1px)',
          backgroundSize: '48px 48px',
        }}
      />

      {/* Abstract decorative shapes */}
      <div className="absolute -top-24 -right-24 w-96 h-96 rounded-full bg-cyan-300/20 blur-3xl pointer-events-none" />
      <div className="absolute -bottom-32 -left-16 w-80 h-80 rounded-full bg-accent/25 blur-3xl pointer-events-none" />
      <div className="absolute top-1/3 right-8 w-24 h-24 rounded-full border border-white/10 pointer-events-none" />
      <div className="absolute bottom-1/4 left-12 w-16 h-16 rounded-full border border-white/8 pointer-events-none" />

      {/* Scattered ring pattern */}
      <svg
        className="absolute top-16 right-16 w-20 h-20 opacity-20 pointer-events-none"
        viewBox="0 0 80 80"
        aria-hidden="true"
      >
        <circle cx="40" cy="40" r="30" stroke="white" strokeWidth="1" fill="none" />
        <circle cx="40" cy="40" r="20" stroke="white" strokeWidth="1" fill="none" />
        <circle cx="40" cy="40" r="10" stroke="white" strokeWidth="1" fill="none" />
      </svg>

      {/* Brand header */}
      <div className="relative z-10 flex items-center gap-2 select-none">
        <div className="w-8 h-8 rounded-lg bg-white/20 border border-white/30 flex items-center justify-center backdrop-blur-sm">
          <span className="text-white font-mono text-sm font-bold">S</span>
        </div>
        <div>
          <p className="text-sm font-semibold tracking-wide">SMS Portal</p>
          <p className="text-[10px] uppercase tracking-widest opacity-75">Ethiopia G9–12</p>
        </div>
      </div>

      {/* Illustration scene */}
      <div className="relative z-10 flex-1 flex items-center justify-center py-6 lg:py-10 min-h-[280px]">
        {/* Floating UI elements */}
        <FloatingCard className="top-[8%] left-[6%] lg:left-[10%] p-2.5" delay={0}>
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg bg-white/20 flex items-center justify-center">
              <MessageCircle className="w-3.5 h-3.5 text-white" strokeWidth={2} />
            </div>
            <div className="pr-1">
              <p className="text-[9px] font-medium text-white/90 leading-none">New message</p>
              <p className="text-[8px] text-white/60 mt-0.5">Parent inquiry</p>
            </div>
          </div>
        </FloatingCard>

        <FloatingCard className="top-[4%] right-[8%] lg:right-[12%] p-2" delay={1}>
          <div className="flex items-center gap-1.5">
            <Bell className="w-4 h-4 text-white/90" strokeWidth={2} />
            <span className="text-[9px] font-medium text-white/80">3 alerts</span>
          </div>
        </FloatingCard>

        <FloatingCard className="top-[32%] left-[2%] lg:left-[5%] p-2.5" delay={2}>
          <div className="flex items-center gap-2">
            <BarChart3 className="w-4 h-4 text-cyan-200" strokeWidth={2} />
            <div>
              <p className="text-[9px] font-semibold text-white leading-none">94%</p>
              <p className="text-[8px] text-white/60">Attendance</p>
            </div>
          </div>
        </FloatingCard>

        <FloatingCard className="top-[28%] right-[4%] lg:right-[8%] p-2.5 w-[88px]" delay={3}>
          <div className="flex items-center gap-1.5 mb-1.5">
            <GraduationCap className="w-3.5 h-3.5 text-white/90" strokeWidth={2} />
            <span className="text-[8px] font-medium text-white/80">Grades</span>
          </div>
          <div className="flex items-end gap-1 h-6">
            <div className="w-2 bg-white/30 rounded-sm h-[40%]" />
            <div className="w-2 bg-white/50 rounded-sm h-[70%]" />
            <div className="w-2 bg-cyan-300/80 rounded-sm h-[100%]" />
            <div className="w-2 bg-white/40 rounded-sm h-[55%]" />
          </div>
        </FloatingCard>

        <FloatingCard className="bottom-[28%] left-[8%] lg:left-[12%] p-2" delay={1}>
          <CalendarCheck className="w-4 h-4 text-white/85" strokeWidth={2} />
        </FloatingCard>

        <FloatingCard className="bottom-[32%] right-[6%] lg:right-[10%] p-2.5" delay={2}>
          <div className="flex items-center gap-2">
            <Users className="w-3.5 h-3.5 text-white/90" strokeWidth={2} />
            <span className="text-[9px] text-white/80 font-medium">1,240 students</span>
          </div>
        </FloatingCard>

        <FloatingCard className="bottom-[12%] left-[18%] p-2" delay={0}>
          <BookOpen className="w-4 h-4 text-white/80" strokeWidth={2} />
        </FloatingCard>

        <FloatingCard className="bottom-[14%] right-[20%] p-2" delay={3}>
          <Mail className="w-4 h-4 text-white/80" strokeWidth={2} />
        </FloatingCard>

        {/* Icon cloud — upper arc */}
        <div className="absolute top-[2%] left-1/2 -translate-x-1/2 flex gap-3 opacity-40" aria-hidden="true">
          {[...Array(5)].map((_, i) => (
            <motion.div
              key={i}
              animate={{ opacity: [0.3, 0.6, 0.3] }}
              transition={{ duration: 2.5, repeat: Infinity, delay: i * 0.3 }}
              className="w-1 h-1 rounded-full bg-white"
            />
          ))}
        </div>

        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.6, ease: 'easeOut' }}
          className="relative z-10"
        >
          <CentralIllustration />
        </motion.div>
      </div>

      {/* Tagline carousel */}
      <div className="relative z-10 max-w-xl">
        <AnimatePresence mode="wait">
          <motion.div
            key={activeSlide}
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.4 }}
          >
            <h2 className="text-xl lg:text-2xl font-bold tracking-tight text-white sm:text-3xl">
              {SLIDES[activeSlide].title}
            </h2>
            <p className="text-sm text-cyan-100/80 leading-relaxed mt-2">
              {SLIDES[activeSlide].description}
            </p>
          </motion.div>
        </AnimatePresence>
      </div>

      {/* Carousel indicator — bottom right */}
      <div
        className="absolute bottom-8 right-8 lg:bottom-12 lg:right-12 z-20 flex items-center gap-2"
        aria-hidden="true"
      >
        {SLIDES.map((_, i) => (
          <button
            key={i}
            type="button"
            tabIndex={-1}
            onClick={() => setActiveSlide(i)}
            className={cn(
              'rounded-full transition-all duration-300',
              i === activeSlide
                ? 'w-6 h-2 bg-white'
                : 'w-2 h-2 bg-white/40 hover:bg-white/60',
            )}
          />
        ))}
      </div>
    </aside>
  )
}
