import { useEffect, useState, type FormEvent } from 'react'
import { getAllTransactions, verifyReceipt, createPayroll, getPayrolls, markPayrollPaid } from '../../api/finance'
import { getTeachers } from '../../api/admin'
import type { Transaction, Payroll } from '../../types/finance'
import { listFromApi } from '../../types/api'
import type { Teacher } from '../../types/academic'
import { Table } from '../../components/ui/Table'
import { Badge, statusVariant } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Modal } from '../../components/ui/Modal'
import { Spinner } from '../../components/ui/Spinner'

type Tab = 'transactions' | 'payroll'

export default function AdminFinancePage() {
  const [tab, setTab] = useState<Tab>('transactions')
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [payrolls, setPayrolls] = useState<Payroll[]>([])
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [loading, setLoading] = useState(true)
  const [payrollModal, setPayrollModal] = useState(false)
  const [payrollForm, setPayrollForm] = useState({ teacher_id: 0, amount: '', month: new Date().getMonth() + 1, year: new Date().getFullYear() })
  const [savingPayroll, setSavingPayroll] = useState(false)

  const loadPayrolls = () =>
    getPayrolls().then((r) => setPayrolls(Array.isArray(r.data.data) ? r.data.data : []))

  useEffect(() => {
    Promise.all([getAllTransactions(), getTeachers({ page_size: 50 }), getPayrolls()])
      .then(([t, te, p]) => {
        setTransactions(listFromApi(t.data))
        setTeachers(listFromApi(te.data))
        setPayrolls(Array.isArray(p.data.data) ? p.data.data : [])
      })
      .finally(() => setLoading(false))
  }, [])

  const handleVerify = async (id: number) => {
    await verifyReceipt(id)
    setTransactions((prev) => prev.map((t) => t.id === id ? { ...t, status: 'Verified' } : t))
  }

  const handlePayroll = async (e: FormEvent) => {
    e.preventDefault()
    setSavingPayroll(true)
    try {
      await createPayroll({ ...payrollForm, amount: Number(payrollForm.amount) })
      setPayrollModal(false)
      await loadPayrolls()
    } finally {
      setSavingPayroll(false)
    }
  }

  const handleMarkPaid = async (id: number) => {
    await markPayrollPaid(id)
    await loadPayrolls()
  }

  if (loading) return <Spinner fullPage />

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-[var(--text-h)]">Finance</h1>
        <Button size="sm" onClick={() => setPayrollModal(true)}>+ Payroll</Button>
      </div>

      <div className="flex gap-2">
        {(['transactions', 'payroll'] as Tab[]).map((t) => (
          <button key={t} onClick={() => setTab(t)}
            className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-colors capitalize
              ${tab === t ? 'bg-[var(--accent)] text-white' : 'bg-[var(--code-bg)] text-[var(--text)]'}`}
          >
            {t}
          </button>
        ))}
      </div>

      {tab === 'transactions' && (
        <Table keyExtractor={(t) => t.id} data={transactions} emptyMessage="No transactions."
          columns={[
            { key: 'student', header: 'Student', render: (t) => t.student?.user?.name ?? `#${t.student_id}` },
            { key: 'receipt_id', header: 'Receipt ID' },
            { key: 'amount', header: 'Amount', render: (t) => `ETB ${t.amount.toLocaleString()}` },
            { key: 'type', header: 'Type' },
            { key: 'status', header: 'Status', render: (t) => <Badge label={t.status} variant={statusVariant(t.status)} /> },
            { key: 'created_at', header: 'Date', render: (t) => new Date(t.created_at).toLocaleDateString() },
            {
              key: 'actions', header: '',
              render: (t) => t.status === 'Pending' ? (
                <Button size="sm" variant="secondary" onClick={() => handleVerify(t.id)}>Verify</Button>
              ) : null,
            },
          ]}
        />
      )}

      {tab === 'payroll' && (
        <div className="max-h-[480px] overflow-y-auto rounded-xl border border-[var(--border)]">
          <Table keyExtractor={(p) => p.id} data={payrolls} emptyMessage="No payroll records. Run seed or create one."
            columns={[
              { key: 'teacher', header: 'Teacher', render: (p) => p.teacher?.user?.name ?? `#${p.teacher_id}` },
              { key: 'amount', header: 'Amount', render: (p) => `ETB ${p.amount.toLocaleString()}` },
              { key: 'month', header: 'Month', render: (p) => new Date(0, p.month - 1).toLocaleString('default', { month: 'long' }) },
              { key: 'year', header: 'Year' },
              { key: 'status', header: 'Status', render: (p) => <Badge label={p.status} variant={statusVariant(p.status)} /> },
              {
                key: 'actions', header: '',
                render: (p) => p.status === 'Pending' ? (
                  <Button size="sm" variant="secondary" onClick={() => handleMarkPaid(p.id)}>Mark paid</Button>
                ) : null,
              },
            ]}
          />
        </div>
      )}

      <Modal open={payrollModal} onClose={() => setPayrollModal(false)} title="Create Payroll"
        footer={<Button form="payroll-form" type="submit" loading={savingPayroll}>Create</Button>}
      >
        <form id="payroll-form" onSubmit={handlePayroll} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Teacher</label>
            <select value={payrollForm.teacher_id}
              onChange={(e) => setPayrollForm((f) => ({ ...f, teacher_id: Number(e.target.value) }))}
              className="w-full px-3 py-2 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none focus:border-[var(--accent)]"
              required>
              <option value={0} disabled>Select teacher...</option>
              {teachers.map((t) => <option key={t.id} value={t.id}>{t.user?.name}</option>)}
            </select>
          </div>
          <Input label="Amount (ETB)" type="number" min="0"
            value={payrollForm.amount} onChange={(e) => setPayrollForm((f) => ({ ...f, amount: e.target.value }))} required />
          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-[var(--text-h)]">Month</label>
              <select value={payrollForm.month}
                onChange={(e) => setPayrollForm((f) => ({ ...f, month: Number(e.target.value) }))}
                className="w-full px-3 py-2 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none">
                {Array.from({ length: 12 }, (_, i) => (
                  <option key={i + 1} value={i + 1}>{new Date(0, i).toLocaleString('default', { month: 'long' })}</option>
                ))}
              </select>
            </div>
            <Input label="Year" type="number" value={payrollForm.year}
              onChange={(e) => setPayrollForm((f) => ({ ...f, year: Number(e.target.value) }))} required />
          </div>
        </form>
      </Modal>
    </div>
  )
}
