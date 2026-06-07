import { useEffect, useState } from 'react'
import { getStudents } from '../../api/admin'
import { getStudentPublicFiles } from '../../api/locker'
import type { Student } from '../../types/academic'
import { listFromApi } from '../../types/api'
import type { LockerFile } from '../../types/locker'
import { FolderOpen, Calendar } from 'lucide-react'
import { EmptyState } from '../../components/ui/EmptyState'
import { Spinner } from '../../components/ui/Spinner'
import { Badge } from '../../components/ui/Badge'
import {
  FileTypeIcon, getFileExtension, getFileTypeLabel, formatBytes, formatDate,
} from '../../lib/file-type-icon'

export default function StudentLockerPage() {
  const [students, setStudents] = useState<Student[]>([])
  const [selectedId, setSelectedId] = useState(0)
  const [files, setFiles] = useState<LockerFile[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => { getStudents({ page_size: 50 }).then((r) => setStudents(listFromApi(r.data))) }, [])

  const loadFiles = async (id: number) => {
    setSelectedId(id)
    if (id === 0) { setFiles([]); return }
    setLoading(true)
    try {
      const res = await getStudentPublicFiles(id)
      setFiles(res.data.data ?? [])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <h1 className="text-xl font-bold text-foreground">Student Locker</h1>

      <div className="flex flex-col gap-1.5">
        <label className="text-sm font-medium text-foreground">Select Student</label>
        <select value={selectedId} onChange={(e) => loadFiles(Number(e.target.value))}
          className="w-full max-w-xs px-3 py-2 rounded-lg text-sm bg-surface border border-surface-border text-foreground outline-none focus:border-accent">
          <option value={0}>Select student...</option>
          {students.map((s) => <option key={s.id} value={s.id}>{s.user?.name ?? s.student_code}</option>)}
        </select>
      </div>

      {loading ? (
        <Spinner />
      ) : selectedId === 0 ? null : files.length === 0 ? (
        <EmptyState icon={FolderOpen} title="No public files" description="This student has no public locker files." />
      ) : (
        <div className="space-y-2">
          {files.map((file) => {
            const ext = getFileExtension(file.file_type)
            const typeLabel = getFileTypeLabel(file.file_type)
            return (
              <div
                key={file.id}
                className="flex items-center gap-4 px-4 py-3.5 bg-surface border border-surface-border rounded-xl hover:border-surface-border/80 transition-colors"
              >
                <FileTypeIcon type={file.file_type} size="md" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-0.5">
                    <p className="text-sm font-medium text-foreground truncate">{file.file_name}</p>
                    <Badge label={ext} />
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted">
                    <span>{typeLabel}</span>
                    <span className="text-border">·</span>
                    <span>{formatBytes(file.file_size)}</span>
                    <span className="text-border">·</span>
                    <span>{file.category}</span>
                    <span className="text-border">·</span>
                    <span className="flex items-center gap-1">
                      <Calendar className="w-3 h-3" />
                      {formatDate(file.uploaded_at || file.created_at)}
                    </span>
                  </div>
                </div>
                <Badge label={file.file_type.toUpperCase()} />
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}