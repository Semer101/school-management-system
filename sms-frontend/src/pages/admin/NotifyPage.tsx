import { useState, type FormEvent } from 'react'
import { Megaphone, Bell } from 'lucide-react'
import { broadcastAnnouncement, notifyAbsentParents } from '../../api/admin'
import { Input } from '../../components/ui/Input'
import { Button } from '../../components/ui/Button'
import type { Role } from '../../types/user'

const ALL_ROLES: Role[] = ['Admin', 'Teacher', 'Student', 'Parent']

export default function NotifyPage() {
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [roles, setRoles] = useState<Role[]>(['Student', 'Parent'])
  const [sending, setSending] = useState(false)
  const [absentSending, setAbsentSending] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const toggleRole = (role: Role) =>
    setRoles((prev) =>
      prev.includes(role) ? prev.filter((r) => r !== role) : [...prev, role]
    )

  const handleBroadcast = async (e: FormEvent) => {
    e.preventDefault()
    if (roles.length === 0) { setError('Select at least one role.'); return }
    setError(''); setMessage(''); setSending(true)
    try {
      await broadcastAnnouncement({ title, body, target_roles: roles })
      setMessage('Announcement sent successfully.')
      setTitle(''); setBody('')
    } catch (err: unknown) {
      setError(
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Failed to send announcement.'
      )
    } finally {
      setSending(false)
    }
  }

  const handleNotifyAbsent = async () => {
    if (!confirm('Send absence notifications to all parents of absent students today?')) return
    setAbsentSending(true)
    try {
      await notifyAbsentParents()
      setMessage('Absence notifications sent.')
    } catch {
      setError('Failed to notify absent parents.')
    } finally {
      setAbsentSending(false)
    }
  }

  return (
    <div className="max-w-xl mx-auto space-y-6">
      <h1 className="text-xl font-bold text-[var(--text-h)]">Notifications</h1>

      {/* Broadcast */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <h2 className="text-base font-semibold text-[var(--text-h)] mb-4 flex items-center gap-2">
          <Megaphone className="w-4 h-4 text-accent" />
          Broadcast Announcement
        </h2>
        <form onSubmit={handleBroadcast} className="flex flex-col gap-4">
          <Input
            label="Title"
            placeholder="School closure tomorrow"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
          />
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Message</label>
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Write your announcement here..."
              rows={4}
              required
              className="w-full px-3 py-2 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] placeholder:text-[var(--text)] outline-none focus:border-[var(--accent)] resize-none"
            />
          </div>

          {/* Role checkboxes */}
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Send to</label>
            <div className="flex gap-3 flex-wrap">
              {ALL_ROLES.map((role) => (
                <label key={role} className="flex items-center gap-1.5 cursor-pointer text-sm text-[var(--text)]">
                  <input
                    type="checkbox"
                    checked={roles.includes(role)}
                    onChange={() => toggleRole(role)}
                    className="accent-[var(--accent)]"
                  />
                  {role}
                </label>
              ))}
            </div>
          </div>

          {error && <p className="text-sm text-red-500">{error}</p>}
          {message && <p className="text-sm text-green-600">{message}</p>}

          <Button type="submit" loading={sending}>Send Announcement</Button>
        </form>
      </div>

      {/* Notify absent parents */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <h2 className="text-base font-semibold text-[var(--text-h)] mb-2 flex items-center gap-2">
          <Bell className="w-4 h-4 text-accent" />
          Notify Absent Parents
        </h2>
        <p className="text-sm text-[var(--text)] mb-4">
          Automatically send notifications to parents of all students marked absent today.
        </p>
        <Button variant="secondary" loading={absentSending} onClick={handleNotifyAbsent}>
          Send Absence Alerts
        </Button>
      </div>
    </div>
  )
}