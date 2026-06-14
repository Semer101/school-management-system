import { useEffect, useState, useRef, type ChangeEvent, type FormEvent } from 'react'
import { useRole } from '../../hooks/useRole'
import { submitReceipt, getMyTransactions, uploadReceipt } from '../../api/finance'
import { getMyChildren } from '../../api/parent'
import { submitParentReceipt, getParentTransactions } from '../../api/parent'
import type { Transaction } from '../../types/finance'
import type { Student } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { Table } from '../../components/ui/Table'
import { Badge, statusVariant } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Spinner } from '../../components/ui/Spinner'

const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/webp']
const MAX_FILE_SIZE = 5 * 1024 * 1024

export default function MyFinancePage() {
  const { isParent } = useRole()
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [children, setChildren] = useState<Student[]>([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState({ student_id: 0, amount: '', receipt_id: '', semester: '', description: '' })
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [fileError, setFileError] = useState('')
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  const fetchTx = async () => {
    const res = isParent ? await getParentTransactions() : await getMyTransactions()
    setTransactions(listFromApi(res.data))
  }

  useEffect(() => {
    async function load() {
      setLoading(true)
      try {
        if (isParent) {
          const childrenRes = await getMyChildren()
          setChildren(listFromApi(childrenRes.data))
        }
        await fetchTx()
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [isParent])

  const handleFileSelect = (e: ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    setFileError('')
    if (!f) return setFile(null)

    if (!ACCEPTED_TYPES.includes(f.type)) {
      setFileError('Only JPG, JPEG, PNG, and WEBP images are supported.')
      setFile(null)
      setPreview(null)
      return
    }
    if (f.size > MAX_FILE_SIZE) {
      setFileError('File size must be less than 5 MB.')
      setFile(null)
      setPreview(null)
      return
    }

    setFile(f)
    setPreview(URL.createObjectURL(f))
  }

  const removeFile = () => {
    setFile(null)
    if (preview) URL.revokeObjectURL(preview)
    setPreview(null)
    setFileError('')
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(''); setMessage(''); setSaving(true)

    try {
      if (isParent) {
        if (!file) {
          setError('Receipt image is required.')
          setSaving(false)
          return
        }
        if (!form.description.trim()) {
          setError('Description is required.')
          setSaving(false)
          return
        }
        // Image upload via multipart FormData
        const fd = new FormData()
        fd.append('receipt', file)
        fd.append('student_id', String(form.student_id))
        fd.append('amount', form.amount)
        fd.append('receipt_id', form.receipt_id)
        fd.append('description', form.description)
        if (form.semester) {
          fd.append('semester', form.semester)
        }
        const res = await uploadReceipt(fd)
        const created = res.data.data
        if (created) setTransactions((prev) => [created, ...prev])
      } else {
        // Plain text receipt (existing path)
        const res = isParent
          ? await submitParentReceipt({
              amount: Number(form.amount),
              receipt_id: form.receipt_id,
              semester: form.semester || undefined,
              description: form.description,
            })
          : await submitReceipt({
              amount: Number(form.amount),
              receipt_id: form.receipt_id,
              semester: form.semester || undefined,
              description: form.description,
            })
        const created = res.data.data
        if (created) setTransactions((prev) => [created, ...prev])
      }

      setMessage('Receipt submitted. Pending admin verification.')
      setForm({ student_id: 0, amount: '', receipt_id: '', semester: '', description: '' })
      removeFile()
    } catch (err: unknown) {
      const httpErr = err as { response?: { data?: { error?: string } } }
      setError(httpErr?.response?.data?.error ?? 'Submission failed.')
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <Spinner fullPage />

  const apiBase = import.meta.env.VITE_API_BASE_URL ?? ''

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <h1 className="text-xl font-bold text-foreground">Finance</h1>

      {/* Submit receipt */}
      <div className="bg-surface border border-surface-border rounded-2xl p-6">
        <h2 className="text-base font-semibold text-foreground mb-4">Submit Payment Receipt</h2>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          {isParent && children.length > 0 && (
            <label className="text-sm font-medium text-foreground">
              Student
              <select
                value={form.student_id}
                onChange={(e) => setForm((f) => ({ ...f, student_id: Number(e.target.value) }))}
                className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface text-sm focus:outline-none focus:border-accent/50"
                required
              >
                <option value={0}>Select a student...</option>
                {children.map((c) => (
                  <option key={c.id} value={c.id}>{c.user?.name ?? `Student #${c.id}`}</option>
                ))}
              </select>
            </label>
          )}

          <div className="grid grid-cols-2 gap-4">
            <Input label="Amount (ETB)" type="number" min="0" step="0.01"
              value={form.amount} onChange={(e) => setForm((f) => ({ ...f, amount: e.target.value }))} required />
            <Input label="Bank Receipt / Tx ID" placeholder="TXN-XXXXXXXX"
              value={form.receipt_id} onChange={(e) => setForm((f) => ({ ...f, receipt_id: e.target.value }))} required />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input label="Description" placeholder="Tuition for Semester 1"
              value={form.description} onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))} required />
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-foreground">Semester</label>
              <select
                value={form.semester}
                onChange={(e) => setForm((f) => ({ ...f, semester: e.target.value }))}
                className="w-full px-3 py-2 rounded-lg border border-surface-border bg-surface text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-accent/40 focus:border-accent"
              >
                <option value="">Other / Non-semester fee</option>
                <option value="Semester 1">Semester 1</option>
                <option value="Semester 2">Semester 2</option>
                <option value="Semester 3">Semester 3</option>
              </select>
            </div>
          </div>

          {/* Image upload */}
          {isParent && (
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">Upload Receipt Image (JPG / PNG / WEBP) *</label>

              {!preview ? (
                <div
                  onClick={() => fileInputRef.current?.click()}
                  className="border-2 border-dashed border-surface-border rounded-xl p-8 text-center cursor-pointer hover:border-accent/50 transition-colors"
                >
                  <p className="text-sm text-muted">Click to select a receipt image</p>
                  <p className="text-xs text-muted/60 mt-1">Max 5 MB — JPG, JPEG, PNG, WEBP</p>
                </div>
              ) : (
                <div className="relative inline-block">
                  <img
                    src={preview}
                    alt="Receipt preview"
                    className="max-h-48 rounded-xl border border-surface-border object-contain"
                  />
                  <button
                    type="button"
                    onClick={removeFile}
                    className="absolute -top-2 -right-2 w-6 h-6 rounded-full bg-danger text-white text-xs font-bold flex items-center justify-center hover:bg-red-600"
                  >
                    ×
                  </button>
                </div>
              )}

              <input
                ref={fileInputRef}
                type="file"
                accept=".jpg,.jpeg,.png,.webp"
                onChange={handleFileSelect}
                className="hidden"
              />

              {file && (
                <p className="text-xs text-muted font-mono">{file.name} ({(file.size / 1024).toFixed(1)} KB)</p>
              )}
              {fileError && <p className="text-xs text-danger">{fileError}</p>}
            </div>
          )}

          {error && <p className="text-sm text-danger">{error}</p>}
          {message && <p className="text-sm text-success font-semibold">{message}</p>}

          <Button type="submit" loading={saving}>Submit Receipt</Button>
        </form>
      </div>

      {/* Transaction history */}
      <div>
        <h2 className="text-base font-semibold text-foreground mb-3">Transaction History</h2>
        <Table
          keyExtractor={(t) => t.id}
          data={transactions}
          emptyMessage="No transactions yet."
          columns={[
            { key: 'receipt_id', header: 'Receipt ID' },
            {
              key: 'receipt_image_url',
              header: 'Receipt',
              render: (t) =>
                t.receipt_image_url ? (
                  <a
                    href={`${apiBase}${t.receipt_image_url}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-accent underline decoration-dotted text-xs"
                  >
                    View Image
                  </a>
                ) : (
                  <span className="text-muted text-xs">N/A</span>
                ),
            },
            { key: 'amount', header: 'Amount', render: (t) => `ETB ${t.amount.toLocaleString()}` },
            { key: 'type', header: 'Type' },
            {
              key: 'status',
              header: 'Status',
              render: (t) => (
                <div className="flex flex-col gap-0.5">
                  <Badge label={t.status} variant={statusVariant(t.status)} />
                  {t.status === 'Rejected' && t.rejection_notes && (
                    <span className="text-[10px] text-danger/80 max-w-[160px] line-clamp-2" title={t.rejection_notes}>
                      {t.rejection_notes}
                    </span>
                  )}
                </div>
              ),
            },
            { key: 'created_at', header: 'Date', render: (t) => new Date(t.created_at).toLocaleDateString() },
          ]}
        />
      </div>
    </div>
  )
}