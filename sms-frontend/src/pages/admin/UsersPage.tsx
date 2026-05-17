import { useState, type FormEvent } from 'react'
import { registerUser } from '../../api/admin'
import type { Role } from '../../types/user'
import { Input } from '../../components/ui/Input'
import { Button } from '../../components/ui/Button'

export default function UsersPage() {
  const [form, setForm] = useState({ name: '', email: '', password: '', role: 'Student' as Role, phone: '' })
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState('')
  const [error, setError] = useState('')

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }))

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(''); setSuccess(''); setLoading(true)
    try {
      await registerUser(form)
      setSuccess(`User "${form.name}" registered as ${form.role}.`)
      setForm({ name: '', email: '', password: '', role: 'Student', phone: '' })
    } catch (err: unknown) {
      setError(
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Registration failed.'
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-xl font-bold text-[var(--text-h)] mb-6">Register User</h1>
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <Input label="Full Name" value={form.name} onChange={set('name')} required />
          <Input label="Email" type="email" value={form.email} onChange={set('email')} required />
          <Input label="Password" type="password" value={form.password} onChange={set('password')} required />
          <Input label="Phone (optional)" value={form.phone} onChange={set('phone')} />
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Role</label>
            <select
              value={form.role}
              onChange={set('role')}
              className="w-full px-3 py-2 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none focus:border-[var(--accent)]"
            >
              {(['Admin', 'Teacher', 'Student', 'Parent'] as Role[]).map((r) => (
                <option key={r} value={r}>{r}</option>
              ))}
            </select>
          </div>
          {error && <p className="text-sm text-red-500">{error}</p>}
          {success && <p className="text-sm text-green-600">{success}</p>}
          <Button type="submit" loading={loading}>Register User</Button>
        </form>
      </div>
    </div>
  )
}