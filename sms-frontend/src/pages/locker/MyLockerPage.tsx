import { useEffect, useState, useRef } from 'react'
import { getMyLockerFiles, uploadLockerFile, deleteLockerFile, toggleFileVisibility } from '../../api/locker'
import type { LockerFile } from '../../types/locker'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { EmptyState } from '../../components/ui/EmptyState'
import { Spinner } from '../../components/ui/Spinner'

const CATEGORIES = ['Certificate', 'Assignment', 'Portfolio'] as const

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export default function MyLockerPage() {
  const [files, setFiles] = useState<LockerFile[]>([])
  const [loading, setLoading] = useState(true)
  const [uploading, setUploading] = useState(false)
  const [category, setCategory] = useState<typeof CATEGORIES[number]>('Assignment')
  const fileRef = useRef<HTMLInputElement>(null)

  const fetchFiles = () =>
    getMyLockerFiles().then((r) => setFiles(r.data.data ?? [])).finally(() => setLoading(false))

  useEffect(() => { fetchFiles() }, [])

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('category', category)
      const res = await uploadLockerFile(formData)
      const created = res.data.data
      if (created) setFiles((prev) => [created, ...prev])
    } finally {
      setUploading(false)
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this file?')) return
    await deleteLockerFile(id)
    setFiles((prev) => prev.filter((f) => f.id !== id))
  }

  const handleToggle = async (file: LockerFile) => {
    await toggleFileVisibility(file.id, !file.is_public)
    setFiles((prev) => prev.map((f) => f.id === file.id ? { ...f, is_public: !f.is_public } : f))
  }

  if (loading) return <Spinner fullPage />

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <h1 className="text-xl font-bold text-[var(--text-h)]">My Locker</h1>

      {/* Upload */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <h2 className="text-base font-semibold text-[var(--text-h)] mb-4">Upload File</h2>
        <div className="flex items-end gap-3 flex-wrap">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Category</label>
            <select value={category} onChange={(e) => setCategory(e.target.value as typeof category)}
              className="px-3 py-2 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none focus:border-[var(--accent)]">
              {CATEGORIES.map((c) => <option key={c}>{c}</option>)}
            </select>
          </div>
          <input type="file" ref={fileRef} onChange={handleUpload} className="hidden" />
          <Button variant="secondary" loading={uploading} onClick={() => fileRef.current?.click()}>
            📎 Choose File
          </Button>
        </div>
      </div>

      {/* File list */}
      {files.length === 0 ? (
        <EmptyState icon="🗂️" title="Your locker is empty" description="Upload certificates, assignments, or portfolio items." />
      ) : (
        <div className="space-y-2">
          {files.map((file) => (
            <div key={file.id} className="flex items-center justify-between px-4 py-3 bg-[var(--bg)] border border-[var(--border)] rounded-xl">
              <div className="flex items-center gap-3 min-w-0">
                <span className="text-lg shrink-0">
                  {file.file_type === 'pdf' ? '📄' : ['jpg', 'jpeg', 'png'].includes(file.file_type) ? '🖼️' : '📎'}
                </span>
                <div className="min-w-0">
                  <p className="text-sm font-medium text-[var(--text-h)] truncate">{file.file_name}</p>
                  <p className="text-xs text-[var(--text)]">{formatBytes(file.file_size)} · {file.category}</p>
                </div>
              </div>
              <div className="flex items-center gap-2 shrink-0 ml-3">
                <Badge label={file.is_public ? 'Public' : 'Private'} variant={file.is_public ? 'success' : 'default'} />
                <Button size="sm" variant="ghost" onClick={() => handleToggle(file)}>
                  {file.is_public ? '🔒' : '👁️'}
                </Button>
                <Button size="sm" variant="danger" onClick={() => handleDelete(file.id)}>✕</Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}