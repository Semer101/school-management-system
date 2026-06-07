import { useEffect, useState, useRef } from 'react'
import { getMyLockerFiles, uploadLockerFile, deleteLockerFile, toggleFileVisibility } from '../../api/locker'
import type { LockerFile } from '../../types/locker'
import { Button } from '../../components/ui/Button'
import { Badge } from '../../components/ui/Badge'
import { FolderOpen, Paperclip, Lock, Eye, X, Calendar } from 'lucide-react'
import { EmptyState } from '../../components/ui/EmptyState'
import { Spinner } from '../../components/ui/Spinner'
import {
  FileTypeIcon, getFileExtension, getFileTypeLabel, formatBytes, formatDate,
} from '../../lib/file-type-icon'

const CATEGORIES = ['Certificate', 'Assignment', 'Portfolio'] as const

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
    setFiles((prev) => prev.map((f) => f.id === file.id ? { ...f, is_public: !file.is_public } : f))
  }

  if (loading) return <Spinner fullPage />

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <h1 className="text-xl font-bold text-foreground">My Locker</h1>

      {/* Upload */}
      <div className="bg-surface border border-surface-border rounded-2xl p-6">
        <h2 className="text-base font-semibold text-foreground mb-4">Upload File</h2>
        <div className="flex items-end gap-3 flex-wrap">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">Category</label>
            <select value={category} onChange={(e) => setCategory(e.target.value as typeof category)}
              className="px-3 py-2 rounded-lg text-sm bg-surface border border-surface-border text-foreground outline-none focus:border-accent">
              {CATEGORIES.map((c) => <option key={c}>{c}</option>)}
            </select>
          </div>
          <input type="file" ref={fileRef} onChange={handleUpload} className="hidden" />
          <Button variant="secondary" loading={uploading} onClick={() => fileRef.current?.click()}>
            <Paperclip className="w-4 h-4" />
            Choose File
          </Button>
          <span className="text-xs text-muted ml-2">
            Accepted: PDF, Word, Excel, Images, Text
          </span>
        </div>
      </div>

      {/* File list */}
      {files.length === 0 ? (
        <EmptyState icon={FolderOpen} title="Your locker is empty" description="Upload certificates, assignments, or portfolio items." />
      ) : (
        <div className="space-y-2">
          {files.map((file) => {
            const ext = getFileExtension(file.file_type)
            const typeLabel = getFileTypeLabel(file.file_type)
            return (
              <div
                key={file.id}
                className="flex items-center gap-4 px-4 py-3.5 bg-surface border border-surface-border rounded-xl hover:border-surface-border/80 transition-colors group"
              >
                {/* File type icon */}
                <FileTypeIcon type={file.file_type} size="md" />

                {/* File details */}
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

                {/* Actions */}
                <div className="flex items-center gap-2 shrink-0 ml-3">
                  <Badge
                    label={file.is_public ? 'Public' : 'Private'}
                    variant={file.is_public ? 'success' : 'default'}
                  />
                  <Button size="sm" variant="ghost" onClick={() => handleToggle(file)} aria-label={file.is_public ? 'Make private' : 'Make public'}>
                    {file.is_public ? <Lock className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </Button>
                  <Button size="sm" variant="danger" onClick={() => handleDelete(file.id)} aria-label="Delete file">
                    <X className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}