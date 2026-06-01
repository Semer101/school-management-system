import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { forgotPassword, resetPasswordWithOTP } from '../api/auth'
import { Input } from '../components/ui/Input'
import { Button } from '../components/ui/Button'
import { GlassCard } from '../components/ui/GlassCard'

export default function ForgotPasswordPage() {
  const navigate = useNavigate()
  const [step, setStep] = useState<'email' | 'otp'>('email')
  const [email, setEmail] = useState('')
  const [otp, setOtp] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  const sendOTP = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await forgotPassword(email)
      setMessage('If your email is registered, a 6-digit OTP has been sent.')
      setStep('otp')
    } catch {
      setError('Could not send OTP. Try again.')
    } finally {
      setLoading(false)
    }
  }

  const resetPwd = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await resetPasswordWithOTP(email, otp, password)
      navigate('/login', { state: { message: 'Password reset. Please sign in.' } })
    } catch {
      setError('Invalid OTP or password too weak.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center app-grid-bg px-4">
      <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="w-full max-w-md">
        <h1 className="text-2xl font-semibold text-foreground text-center mb-2">Reset Password</h1>
        <p className="text-sm text-muted text-center mb-6">Recover access via email OTP</p>
        <GlassCard className="p-6">
          {step === 'email' ? (
            <form onSubmit={sendOTP} className="space-y-4">
              <Input label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
              {error && <p className="text-sm text-danger">{error}</p>}
              {message && <p className="text-sm text-success">{message}</p>}
              <Button type="submit" loading={loading} className="w-full">Send OTP</Button>
            </form>
          ) : (
            <form onSubmit={resetPwd} className="space-y-4">
              <Input label="6-digit OTP" value={otp} onChange={(e) => setOtp(e.target.value)} maxLength={6} required />
              <Input label="New password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
              {error && <p className="text-sm text-danger">{error}</p>}
              <Button type="submit" loading={loading} className="w-full">Reset password</Button>
              <button type="button" className="text-sm text-accent w-full" onClick={() => setStep('email')}>Resend OTP</button>
            </form>
          )}
          <Link to="/login" className="block text-center text-sm text-muted mt-4 hover:text-accent">Back to login</Link>
        </GlassCard>
      </motion.div>
    </div>
  )
}
