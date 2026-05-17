import { useState, type FormEvent } from 'react'
import { useAuth } from '../../hooks/useAuth'
import { changePassword } from '../../api/me'
import { Input } from '../../components/ui/Input'
import { Button } from '../../components/ui/Button'
import { Badge, roleBadgeVariant } from '../../components/ui/Badge'

export default function ProfilePage() {
  const { user } = useAuth()
  const [current, setCurrent] = useState('')
  const [next, setNext] = useState('')
  const [confirm, setConfirm] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState('')
  const [error, setError] = useState('')

  const handlePassword = async (e: FormEvent) => {
    e.preventDefault()
    if (next !== confirm) { setError('Passwords do not match.'); return }
    setError(''); setSuccess(''); setLoading(true)
    try {
      await changePassword(current, next)
      setSuccess('Password updated successfully.')
      setCurrent(''); setNext(''); setConfirm('')
    } catch (err: unknown) {
      setError(
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Failed to update password.'
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-xl font-bold text-[var(--text-h)] mb-6">My Profile</h1>

      {/* Info card */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6 mb-6">
        <div className="flex items-center gap-4 mb-4">
          <div className="w-14 h-14 rounded-full bg-[var(--accent-bg)] flex items-center justify-center text-[var(--accent)] text-2xl font-bold">
            {user?.name?.[0]?.toUpperCase()}
          </div>
          <div>
            <h2 className="text-lg font-semibold text-[var(--text-h)]">{user?.name}</h2>
            <p className="text-sm text-[var(--text)]">{user?.email}</p>
            {user?.role && (
              <div className="mt-1">
                <Badge label={user.role} variant={roleBadgeVariant(user.role)} />
              </div>
            )}
          </div>
        </div>
        {user?.phone && (
          <p className="text-sm text-[var(--text)]">📞 {user.phone}</p>
        )}
        <p className="text-xs text-[var(--text)] mt-2">
          Member since {user?.created_at ? new Date(user.created_at).toLocaleDateString() : '—'}
        </p>
      </div>

      {/* Change password */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <h2 className="text-base font-semibold text-[var(--text-h)] mb-4">Change Password</h2>
        <form onSubmit={handlePassword} className="flex flex-col gap-4">
          <Input
            label="Current Password"
            type="password"
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
            required
          />
          <Input
            label="New Password"
            type="password"
            value={next}
            onChange={(e) => setNext(e.target.value)}
            required
          />
          <Input
            label="Confirm New Password"
            type="password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            required
            error={confirm && next !== confirm ? 'Passwords do not match' : undefined}
          />
          {error && <p className="text-sm text-red-500">{error}</p>}
          {success && <p className="text-sm text-green-600">{success}</p>}
          <Button type="submit" loading={loading}>Update Password</Button>
        </form>
      </div>
    </div>
  )
}