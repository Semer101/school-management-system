import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useAuth } from '../hooks/useAuth'
import { Input } from '../components/ui/Input'
import { Button } from '../components/ui/Button'
import { GlassCard } from '../components/ui/GlassCard'

export default function LoginPage() {
  const { login, loading, error, clearError } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    clearError()
    try {
      await login(email, password)
      navigate('/dashboard')
    } catch {
      /* error in store */
    }
  }

  return (
    <div className="min-h-screen flex flex-col md:flex-row bg-void relative overflow-hidden">
      {/* Grid background on the entire screen under elements */}
      <div className="absolute inset-0 app-grid-bg pointer-events-none opacity-40 z-0" />

      {/* Left side: Illustration Banner (60% Desktop, 50% Tablet, Hidden on Mobile) */}
      <div className="hidden md:flex md:w-1/2 lg:w-3/5 bg-gradient-to-br from-accent/80 via-accent/90 to-cyan-950 dark:from-cyan-950 dark:via-cyan-900 dark:to-void flex-col justify-between p-12 text-white relative z-10 border-r border-surface-border">
        {/* Glow effect overlay */}
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(8,145,178,0.18),transparent_60%)] pointer-events-none" />
        {/* Subtle grid pattern overlay for depth */}
        <div className="absolute inset-0 opacity-[0.07] pointer-events-none [background-image:linear-gradient(rgba(255,255,255,0.4)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.4)_1px,transparent_1px)] [background-size:48px_48px]" />
        {/* Top-right soft glow blob */}
        <div className="absolute -top-24 -right-24 w-96 h-96 rounded-full bg-cyan-400/20 blur-3xl pointer-events-none" />
        {/* Bottom-left soft glow blob */}
        <div className="absolute -bottom-32 -left-16 w-80 h-80 rounded-full bg-accent/30 blur-3xl pointer-events-none" />
REPLACE


        <div className="flex items-center gap-2 select-none">
          <div className="w-8 h-8 rounded-lg bg-white/20 border border-white/30 flex items-center justify-center">
            <span className="text-white font-mono text-sm font-bold">S</span>
          </div>
          <div>
            <p className="text-sm font-semibold tracking-wide">SMS Portal</p>
            <p className="text-[10px] uppercase tracking-widest opacity-75">Ethiopia G9–12</p>
          </div>
        </div>

        {/* Center Illustration */}
        <div className="flex-1 flex flex-col items-center justify-center py-10">
          <motion.img
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.6, ease: 'easeOut' }}
            src="/login_illustration_cyan.png"
            alt="School Management Portal Illustration"
            className="max-w-[75%] max-h-[45vh] object-contain drop-shadow-[0_12px_24px_rgba(8,145,178,0.25)] select-none pointer-events-none"
            loading="eager"
          />
        </div>

        {/* Footer/Tagline info */}
        <div className="space-y-2 relative z-10 max-w-xl">
          <h2 className="text-2xl font-bold tracking-tight text-white sm:text-3xl">
            Elevating Academic Excellence
          </h2>
          <p className="text-sm text-cyan-100/80 leading-relaxed">
            Welcome to the digital heart of our school. Connect with students, teachers, parents, and administrative tools in a unified, premium space designed for collaboration.
          </p>
        </div>
      </div>

      {/* Right side: Login Form (40% Desktop, 50% Tablet, 100% Mobile) */}
      <div className="flex-1 flex items-center justify-center p-6 sm:p-12 relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
          className="w-full max-w-md"
        >
          <div className="text-center md:text-left mb-8">
            {/* School logo on mobile only */}
            <div className="flex md:hidden items-center justify-center gap-2 mb-4 select-none">
              <div className="w-10 h-10 rounded-lg bg-accent/10 border border-accent/30 flex items-center justify-center">
                <span className="text-accent font-mono text-base font-bold">S</span>
              </div>
              <div className="text-left">
                <p className="text-base font-bold text-foreground leading-tight">SMS Portal</p>
                <p className="text-[10px] uppercase tracking-widest text-muted">Ethiopia G9–12</p>
              </div>
            </div>

            <h1 className="text-2xl font-bold text-foreground tracking-tight sm:text-3xl">
              Welcome back
            </h1>
            <p className="text-sm text-muted mt-1.5">
              Enter your credentials below to access your workspace.
            </p>
          </div>

          <GlassCard className="p-8 shadow-glass border border-surface-border bg-surface/50 backdrop-blur-md">
            <form onSubmit={handleSubmit} className="flex flex-col gap-5">
              <Input
                label="Email Address"
                type="email"
                placeholder="e.g. name@school.et"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                className="w-full"
              />
              <Input
                label="Password"
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
                className="w-full"
              />

              {error && (
                <div className="text-xs font-medium text-danger bg-danger/10 border border-danger/20 rounded-xl px-4 py-3 animate-shake">
                  {error}
                </div>
              )}

              <div className="flex justify-end -mt-1">
                <Link
                  to="/forgot-password"
                  className="text-xs font-semibold text-accent hover:text-accent-hover transition-colors"
                >
                  Forgot password?
                </Link>
              </div>

              <Button
                type="submit"
                loading={loading}
                className="w-full py-2.5 rounded-xl bg-accent text-white hover:bg-accent-hover active:scale-[0.98] transition-all duration-200 mt-2 text-sm font-semibold"
              >
                Sign in
              </Button>
            </form>
          </GlassCard>

          <p className="text-xs text-center text-muted mt-8">
            By signing in, you agree to our terms of service and security policy.
          </p>
        </motion.div>
      </div>
    </div>
  )
}
