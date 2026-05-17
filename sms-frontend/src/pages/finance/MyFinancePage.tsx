import { useEffect, useState, type FormEvent } from 'react'
import { useRole } from '../../hooks/useRole'
import { submitReceipt, getMyTransactions } from '../../api/finance'
import { submitParentReceipt, getParentTransactions } from '../../api/parent'
import type { Transaction } from '../../types/finance'
import { listFromApi } from '../../types/api'
import { Table } from '../../components/ui/Table'
import { Badge, statusVariant } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Spinner } from '../../components/ui/Spinner'

export default function MyFinancePage() {
  const { isParent } = useRole()
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState({ amount: '', receipt_id: '', description: '' })
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const fetchTx = async () => {
    const res = isParent ? await getParentTransactions() : await getMyTransactions()
    setTransactions(listFromApi(res.data))
  }

  useEffect(() => {
    fetchTx().finally(() => setLoading(false))
  }, [])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(''); setMessage(''); setSaving(true)
    try {
      const payload = { amount: Number(form.amount), receipt_id: form.receipt_id, description: form.description }
      const res = isParent ? await submitParentReceipt(payload) : await submitReceipt(payload)
      const created = res.data.data
      if (created) setTransactions((prev) => [created, ...prev])
      setMessage('Receipt submitted. Pending admin verification.')
      setForm({ amount: '', receipt_id: '', description: '' })
    } catch (err: unknown) {
      setError(
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Submission failed.'
      )
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <Spinner fullPage />

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <h1 className="text-xl font-bold text-[var(--text-h)]">Finance</h1>

      {/* Submit receipt */}
      <div className="bg-[var(--bg)] border border-[var(--border)] rounded-2xl p-6">
        <h2 className="text-base font-semibold text-[var(--text-h)] mb-4">Submit Payment Receipt</h2>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="grid grid-cols-2 gap-4">
            <Input label="Amount (ETB)" type="number" min="0" step="0.01"
              value={form.amount} onChange={(e) => setForm((f) => ({ ...f, amount: e.target.value }))} required />
            <Input label="Bank Receipt / Tx ID" placeholder="TXN-XXXXXXXX"
              value={form.receipt_id} onChange={(e) => setForm((f) => ({ ...f, receipt_id: e.target.value }))} required />
          </div>
          <Input label="Description (optional)" placeholder="Tuition for Term 1"
            value={form.description} onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))} />
          {error && <p className="text-sm text-red-500">{error}</p>}
          {message && <p className="text-sm text-green-600">{message}</p>}
          <Button type="submit" loading={saving}>Submit Receipt</Button>
        </form>
      </div>

      {/* Transaction history */}
      <div>
        <h2 className="text-base font-semibold text-[var(--text-h)] mb-3">Transaction History</h2>
        <Table keyExtractor={(t) => t.id} data={transactions}
          emptyMessage="No transactions yet."
          columns={[
            { key: 'receipt_id', header: 'Receipt ID' },
            { key: 'amount', header: 'Amount', render: (t) => `ETB ${t.amount.toLocaleString()}` },
            { key: 'type', header: 'Type' },
            { key: 'status', header: 'Status', render: (t) => <Badge label={t.status} variant={statusVariant(t.status)} /> },
            { key: 'created_at', header: 'Date', render: (t) => new Date(t.created_at).toLocaleDateString() },
          ]}
        />
      </div>
    </div>
  )
}