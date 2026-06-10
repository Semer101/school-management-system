import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useAuth } from '../hooks/useAuth'
import { Input } from '../components/ui/Input'
import { Button } from '../components/ui/Button'
import { GlassCard } from '../components/ui/GlassCard'
import { LoginIllustrationBanner } from '../components/auth/LoginIllustrationBanner'

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

      {/* Left side: Login Form */}
      <div className="flex-1 flex items-center justify-center p-6 sm:p-12 relative z-10 order-2 md:order-1">
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

      {/* Right side: Illustration Banner */}
      <LoginIllustrationBanner className="order-1 md:order-2" />
    </div>
  )
}
