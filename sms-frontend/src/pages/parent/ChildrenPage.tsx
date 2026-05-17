import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getMyChildren } from '../../api/parent'
import type { Student } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { Spinner } from '../../components/ui/Spinner'
import { Users } from 'lucide-react'
import { EmptyState } from '../../components/ui/EmptyState'
import { Button } from '../../components/ui/Button'

export default function ChildrenPage() {
  const [children, setChildren] = useState<Student[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    getMyChildren()
      .then((r) => setChildren(listFromApi(r.data)))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <Spinner fullPage />
  if (children.length === 0) {
    return <EmptyState icon={Users} title="No children linked" description="Contact the school administrator to link your children to your account." />
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-xl font-bold text-[var(--text-h)] mb-6">My Children</h1>
      <div className="space-y-3">
        {children.map((child) => (
          <div key={child.id} className="flex items-center justify-between px-5 py-4 bg-[var(--bg)] border border-[var(--border)] rounded-2xl hover:border-[var(--accent-border)] transition-colors">
            <div className="flex items-center gap-4">
              <div className="w-11 h-11 rounded-full bg-[var(--accent-bg)] flex items-center justify-center text-[var(--accent)] font-bold text-base">
                {child.user?.name?.[0]?.toUpperCase() ?? '?'}
              </div>
              <div>
                <p className="text-sm font-semibold text-[var(--text-h)]">{child.user?.name}</p>
                <p className="text-xs text-[var(--text)]">{child.student_code} · {child.class?.name ?? 'No class'}</p>
              </div>
            </div>
            <Button size="sm" variant="secondary" onClick={() => navigate(`/parent/children/${child.id}`)}>
              View →
            </Button>
          </div>
        ))}
      </div>
    </div>
  )
}