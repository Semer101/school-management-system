import { useEffect, useState } from 'react'
import { getTeachers, archiveTeacher } from '../../api/admin'
import type { Teacher } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { Table } from '../../components/ui/Table'
import { Button } from '../../components/ui/Button'
import { School } from 'lucide-react'
import { EmptyState } from '../../components/ui/EmptyState'

export default function TeachersPage() {
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    getTeachers()
      .then((r) => setTeachers(listFromApi(r.data)))
      .catch(() => setError('Failed to load teachers.'))
      .finally(() => setLoading(false))
  }, [])

  const handleArchive = async (id: number) => {
    if (!confirm('Archive this teacher?')) return
    try {
      await archiveTeacher(id)
      setTeachers((prev) => prev.filter((t) => t.id !== id))
    } catch {
      alert('Failed to archive teacher.')
    }
  }

  if (!loading && teachers.length === 0 && !error) {
    return <EmptyState icon={School} title="No teachers yet" />
  }

  return (
    <div>
      <h1 className="text-xl font-bold text-[var(--text-h)] mb-6">Teachers</h1>
      {error && <p className="text-sm text-red-500 mb-4">{error}</p>}
      <Table
        loading={loading}
        keyExtractor={(t) => t.id}
        data={teachers}
        columns={[
          { key: 'teacher_code', header: 'Code' },
          { key: 'user', header: 'Name', render: (t) => t.user?.name ?? '—' },
          { key: 'email', header: 'Email', render: (t) => t.user?.email ?? '—' },
          { key: 'qualification', header: 'Qualification' },
          { key: 'joined_at', header: 'Joined', render: (t) => new Date(t.joined_at).toLocaleDateString() },
          {
            key: 'actions', header: '',
            render: (t) => (
              <Button size="sm" variant="danger" onClick={() => handleArchive(t.id)}>Archive</Button>
            ),
          },
        ]}
      />
    </div>
  )
}