import { useEffect, useState } from 'react'
import { getStudents, archiveStudent } from '../../api/admin'
import type { Student } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { Table } from '../../components/ui/Table'
import { Button } from '../../components/ui/Button'
import { GraduationCap } from 'lucide-react'
import { EmptyState } from '../../components/ui/EmptyState'

export default function StudentsPage() {
  const [students, setStudents] = useState<Student[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const fetchStudents = async () => {
    setLoading(true)
    try {
      const res = await getStudents()
      setStudents(listFromApi(res.data))
    } catch {
      setError('Failed to load students.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchStudents() }, [])

  const handleArchive = async (id: number) => {
    if (!confirm('Archive this student?')) return
    try {
      await archiveStudent(id)
      setStudents((prev) => prev.filter((s) => s.id !== id))
    } catch {
      alert('Failed to archive student.')
    }
  }

  if (!loading && students.length === 0 && !error) {
    return <EmptyState icon={GraduationCap} title="No students yet" description="Students will appear here once registered." />
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-[var(--text-h)]">Students</h1>
      </div>

      {error && <p className="text-sm text-red-500 mb-4">{error}</p>}

      <Table
        loading={loading}
        keyExtractor={(s) => s.id}
        data={students}
        columns={[
          { key: 'student_code', header: 'Code' },
          { key: 'user', header: 'Name', render: (s) => s.user?.name ?? '—' },
          { key: 'email', header: 'Email', render: (s) => s.user?.email ?? '—' },
          { key: 'class', header: 'Class', render: (s) => s.class?.name ?? '—' },
          { key: 'parent_name', header: 'Parent' },
          {
            key: 'actions',
            header: '',
            render: (s) => (
              <Button size="sm" variant="danger" onClick={() => handleArchive(s.id)}>
                Archive
              </Button>
            ),
          },
        ]}
      />
    </div>
  )
}