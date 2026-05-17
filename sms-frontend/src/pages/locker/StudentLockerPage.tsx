import { useEffect, useState } from 'react'
import { getStudents } from '../../api/admin'
import { getStudentPublicFiles } from '../../api/locker'
import type { Student } from '../../types/academic'
import { listFromApi } from '../../types/api'
import type { LockerFile } from '../../types/locker'
import { EmptyState } from '../../components/ui/EmptyState'
import { Spinner } from '../../components/ui/Spinner'
import { Badge } from '../../components/ui/Badge'

function formatBytes(bytes: number) {
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export default function StudentLockerPage() {
  const [students, setStudents] = useState<Student[]>([])
  const [selectedId, setSelectedId] = useState(0)
  const [files, setFiles] = useState<LockerFile[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => { getStudents().then((r) => setStudents(listFromApi(r.data))) }, [])

  const loadFiles = async (id: number) => {
    setSelectedId(id); setLoading(true)
    try {
      const res = await getStudentPublicFiles(id)
      setFiles(res.data.data ?? [])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <h1 className="text-xl font-bold text-[var(--text-h)]">Student Locker</h1>

      <div className="flex flex-col gap-1.5">
        <label className="text-sm font-medium text-[var(--text-h)]">Select Student</label>
        <select value={selectedId} onChange={(e) => loadFiles(Number(e.target.value))}
          className="w-full max-w-xs px-3 py-2 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none focus:border-[var(--accent)]">
          <option value={0}>Select student...</option>
          {students.map((s) => <option key={s.id} value={s.id}>{s.user?.name ?? s.student_code}</option>)}
        </select>
      </div>

      {loading ? <Spinner /> : selectedId === 0 ? null : files.length === 0 ? (
        <EmptyState icon="🗂️" title="No public files" description="This student has no public locker files." />
      ) : (
        <div className="space-y-2">
          {files.map((file) => (
            <div key={file.id} className="flex items-center gap-3 px-4 py-3 bg-[var(--bg)] border border-[var(--border)] rounded-xl">
              <span className="text-lg shrink-0">
                {file.file_type === 'pdf' ? '📄' : ['jpg', 'jpeg', 'png'].includes(file.file_type) ? '🖼️' : '📎'}
              </span>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-[var(--text-h)] truncate">{file.file_name}</p>
                <p className="text-xs text-[var(--text)]">{formatBytes(file.file_size)} · {file.category}</p>
              </div>
              <Badge label={file.file_type.toUpperCase()} />
            </div>
          ))}
        </div>
      )}
    </div>
  )
}