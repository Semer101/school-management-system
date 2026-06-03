import { useEffect, useState, type FormEvent } from 'react'
import {
  getAllTransactions,
  verifyReceipt,
  createPayroll,
  getPayrolls,
  markPayrollPaid,
  getOverduePayments,
  sendPaymentReminder,
  type OverduePaymentRow
} from '../../api/finance'
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
import { DollarSign, UserCheck, TrendingUp, AlertCircle, FileSpreadsheet, Send } from 'lucide-react'
import { GlassCard } from '../../components/ui/GlassCard'

type Tab = 'transactions' | 'payroll' | 'reminders'

export default function AdminFinancePage() {
  const [tab, setTab] = useState<Tab>('transactions')
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [payrolls, setPayrolls] = useState<Payroll[]>([])
  const [overdues, setOverdues] = useState<OverduePaymentRow[]>([])
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [loading, setLoading] = useState(true)

  // Filters for Transactions
  const [txYear, setTxYear] = useState('')
  const [txSem, setTxSem] = useState('')
  const [txStatus, setTxStatus] = useState('')
  const [txStudent, setTxStudent] = useState('')

  // Filters for Payroll
  const [pyMonth, setPyMonth] = useState('')
  const [pyYear, setPyYear] = useState('')
  const [pyDept, setPyDept] = useState('')

  // Modal form for creating payroll
  const [payrollModal, setPayrollModal] = useState(false)
  const [payrollForm, setPayrollForm] = useState({ teacher_id: 0, amount: '', month: new Date().getMonth() + 1, year: new Date().getFullYear() })
  const [savingPayroll, setSavingPayroll] = useState(false)

  // Reminders states
  const [reminderMessage, setReminderMessage] = useState('')
  const [reminderError, setReminderError] = useState('')
  const [reminderLoading, setReminderLoading] = useState<Record<string, boolean>>({})

  const loadData = () => {
    setLoading(true)
    Promise.all([
      getAllTransactions({
        academic_year: txYear || undefined,
        semester: txSem || undefined,
        status: txStatus || undefined,
        student: txStudent || undefined
      }),
      getTeachers({ page_size: 50 }),
      getPayrolls({
        month: pyMonth || undefined,
        year: pyYear || undefined,
        department: pyDept || undefined
      }),
      getOverduePayments()
    ])
      .then(([t, te, p, o]) => {
        setTransactions(listFromApi(t.data))
        setTeachers(listFromApi(te.data))
        setPayrolls(Array.isArray(p.data.data) ? p.data.data : [])
        setOverdues(Array.isArray(o.data.data) ? o.data.data : [])
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    async function load() {
      setLoading(true)
      try {
        const [t, te, p, o] = await Promise.all([
          getAllTransactions({
            academic_year: txYear || undefined,
            semester: txSem || undefined,
            status: txStatus || undefined,
            student: txStudent || undefined
          }),
          getTeachers({ page_size: 50 }),
          getPayrolls({
            month: pyMonth || undefined,
            year: pyYear || undefined,
            department: pyDept || undefined
          }),
          getOverduePayments()
        ])
        setTransactions(listFromApi(t.data))
        setTeachers(listFromApi(te.data))
        setPayrolls(Array.isArray(p.data.data) ? p.data.data : [])
        setOverdues(Array.isArray(o.data.data) ? o.data.data : [])
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [txYear, txSem, txStatus, txStudent, pyMonth, pyYear, pyDept])

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
      loadData()
    } finally {
      setSavingPayroll(false)
    }
  }

  const handleMarkPaid = async (id: number) => {
    await markPayrollPaid(id)
    loadData()
  }

  const handleSendReminder = async (studentId: number, year: number, semester: string) => {
    const key = `${studentId}-${semester}`
    setReminderLoading(prev => ({ ...prev, [key]: true }))
    setReminderMessage('')
    setReminderError('')
    try {
      await sendPaymentReminder({ student_id: studentId, academic_year: year, semester })
      setReminderMessage(`Tuition fee reminder sent successfully for ${semester}.`)
    } catch {
      setReminderError('Failed to send payment reminder.')
    } finally {
      setReminderLoading(prev => ({ ...prev, [key]: false }))
    }
  }

  const downloadCSV = (headers: string[], data: (string | number)[][], filename: string) => {
    const csvContent = [
      headers.join(','),
      ...data.map(row => row.map(val => `"${String(val).replace(/"/g, '""')}"`).join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.setAttribute('href', url)
    link.setAttribute('download', filename)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const exportTransactionsCSV = () => {
    const headers = ['Student', 'Student Code', 'Receipt ID', 'Amount', 'Type', 'Status', 'Academic Year', 'Semester', 'Date']
    const data = transactions.map(t => [
      t.student?.user?.name || '—',
      t.student?.student_code || '—',
      t.receipt_id,
      t.amount,
      t.type,
      t.status,
      t.academic_year || '—',
      t.semester || '—',
      new Date(t.created_at).toLocaleDateString()
    ])
    downloadCSV(headers, data, `student_payments_${new Date().toISOString().split('T')[0]}.csv`)
  }

  const exportPayrollCSV = () => {
    const headers = ['Teacher', 'Amount', 'Month', 'Year', 'Status', 'Department']
    const data = payrolls.map(p => [
      p.teacher?.user?.name || '—',
      p.amount,
      new Date(0, p.month - 1).toLocaleString('default', { month: 'long' }),
      p.year,
      p.status,
      p.teacher?.department || '—'
    ])
    downloadCSV(headers, data, `staff_payroll_${new Date().toISOString().split('T')[0]}.csv`)
  }

  const exportOverdueCSV = () => {
    const headers = ['Student', 'Student Code', 'Class', 'Academic Year', 'Semester 1 Status', 'Semester 2 Status', 'Semester 3 Status']
    const data = overdues.map(o => [
      o.student_name,
      o.student_code,
      o.class_name,
      o.academic_year,
      o.semester_1,
      o.semester_2,
      o.semester_3
    ])
    downloadCSV(headers, data, `overdue_payments_${new Date().toISOString().split('T')[0]}.csv`)
  }

  if (loading && transactions.length === 0 && payrolls.length === 0) return <Spinner fullPage />

  // Statistics Calculations (based on overall list)
  const totalVerified = transactions.filter(t => t.status === 'Verified').reduce((sum, t) => sum + t.amount, 0)
  const totalPending = transactions.filter(t => t.status === 'Pending').reduce((sum, t) => sum + t.amount, 0)
  const totalPaidPayroll = payrolls.filter(p => p.status === 'Paid').reduce((sum, p) => sum + p.amount, 0)
  const netCashFlow = totalVerified - totalPaidPayroll

  // Active Payroll summary (based on filtered list)
  const payrollPaidSum = payrolls.filter(p => p.status === 'Paid').reduce((sum, p) => sum + p.amount, 0)
  const payrollPendingSum = payrolls.filter(p => p.status === 'Pending').reduce((sum, p) => sum + p.amount, 0)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-[var(--text-h)]">Finance</h1>
        <Button size="sm" onClick={() => setPayrollModal(true)}>+ Payroll</Button>
      </div>

      {/* KPI Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <GlassCard className="p-4 flex items-center gap-3">
          <div className="p-3 rounded-xl bg-green-500/10 text-green-500">
            <TrendingUp className="w-5 h-5" />
          </div>
          <div>
            <p className="text-xs text-muted">Total Revenue</p>
            <h3 className="text-lg font-bold text-foreground">ETB {totalVerified.toLocaleString()}</h3>
          </div>
        </GlassCard>
        <GlassCard className="p-4 flex items-center gap-3">
          <div className="p-3 rounded-xl bg-yellow-500/10 text-yellow-500">
            <DollarSign className="w-5 h-5" />
          </div>
          <div>
            <p className="text-xs text-muted">Pending Revenue</p>
            <h3 className="text-lg font-bold text-foreground">ETB {totalPending.toLocaleString()}</h3>
          </div>
        </GlassCard>
        <GlassCard className="p-4 flex items-center gap-3">
          <div className="p-3 rounded-xl bg-red-500/10 text-red-500">
            <AlertCircle className="w-5 h-5" />
          </div>
          <div>
            <p className="text-xs text-muted">Payroll Expenses</p>
            <h3 className="text-lg font-bold text-foreground">ETB {totalPaidPayroll.toLocaleString()}</h3>
          </div>
        </GlassCard>
        <GlassCard className="p-4 flex items-center gap-3">
          <div className="p-3 rounded-xl bg-blue-500/10 text-blue-500">
            <UserCheck className="w-5 h-5" />
          </div>
          <div>
            <p className="text-xs text-muted">Net Cash Flow</p>
            <h3 className={`text-lg font-bold ${netCashFlow >= 0 ? 'text-green-500' : 'text-red-500'}`}>
              ETB {netCashFlow.toLocaleString()}
            </h3>
          </div>
        </GlassCard>
      </div>

      {/* Tabs */}
      <div className="flex gap-2">
        {([
          { id: 'transactions', label: 'Student Payments' },
          { id: 'payroll', label: 'Staff Payroll' },
          { id: 'reminders', label: 'Overdue & Reminders' }
        ] as const).map((t) => (
          <button key={t.id} onClick={() => setTab(t.id)}
            className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-colors
              ${tab === t.id ? 'bg-[var(--accent)] text-white' : 'bg-[var(--code-bg)] text-[var(--text)]'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Transactions Tab */}
      {tab === 'transactions' && (
        <div className="space-y-4">
          <div className="flex flex-wrap gap-3 items-end bg-surface border border-surface-border p-4 rounded-xl">
            <label className="text-xs text-muted flex flex-col gap-1">
              Academic Year
              <select value={txYear} onChange={(e) => setTxYear(e.target.value)}
                className="px-3 py-1.5 rounded-lg border border-surface-border bg-surface text-sm">
                <option value="">All</option>
                {[2024, 2025, 2026, 2027].map(y => <option key={y} value={String(y)}>{y}</option>)}
              </select>
            </label>
            <label className="text-xs text-muted flex flex-col gap-1">
              Semester
              <select value={txSem} onChange={(e) => setTxSem(e.target.value)}
                className="px-3 py-1.5 rounded-lg border border-surface-border bg-surface text-sm">
                <option value="">All</option>
                {['Semester 1', 'Semester 2', 'Semester 3'].map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </label>
            <label className="text-xs text-muted flex flex-col gap-1">
              Status
              <select value={txStatus} onChange={(e) => setTxStatus(e.target.value)}
                className="px-3 py-1.5 rounded-lg border border-surface-border bg-surface text-sm">
                <option value="">All</option>
                {['Pending', 'Verified', 'Rejected'].map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </label>
            <label className="text-xs text-muted flex flex-col gap-1 flex-1 min-w-[200px]">
              Student Name / Code
              <input type="text" value={txStudent} onChange={(e) => setTxStudent(e.target.value)}
                placeholder="Search..."
                className="px-3 py-1.5 rounded-lg border border-surface-border bg-surface text-sm w-full" />
            </label>
            <Button size="sm" variant="secondary" onClick={exportTransactionsCSV} className="flex items-center gap-1.5">
              <FileSpreadsheet className="w-4 h-4" /> Export CSV
            </Button>
          </div>

          <Table keyExtractor={(t) => t.id} data={transactions} emptyMessage="No transactions."
            columns={[
              { key: 'student', header: 'Student', render: (t) => t.student?.user?.name ?? `#${t.student_id}` },
              { key: 'receipt_id', header: 'Receipt ID' },
              { key: 'amount', header: 'Amount', render: (t) => `ETB ${t.amount.toLocaleString()}` },
              { key: 'academic_year', header: 'Academic Year', render: (t) => t.academic_year || '—' },
              { key: 'semester', header: 'Semester', render: (t) => t.semester || '—' },
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
        </div>
      )}

      {/* Payroll Tab */}
      {tab === 'payroll' && (
        <div className="space-y-4">
          <div className="flex justify-between items-center bg-surface border border-surface-border p-4 rounded-xl flex-wrap gap-3">
            <div className="flex gap-4 text-sm font-medium">
              <div>Total Paid: <span className="text-green-500 font-bold">ETB {payrollPaidSum.toLocaleString()}</span></div>
              <div className="border-l border-surface-border pr-2" />
              <div>Total Pending: <span className="text-yellow-500 font-bold">ETB {payrollPendingSum.toLocaleString()}</span></div>
            </div>
          </div>

          <div className="flex flex-wrap gap-3 items-end bg-surface border border-surface-border p-4 rounded-xl">
            <label className="text-xs text-muted flex flex-col gap-1">
              Month
              <select value={pyMonth} onChange={(e) => setPyMonth(e.target.value)}
                className="px-3 py-1.5 rounded-lg border border-surface-border bg-surface text-sm">
                <option value="">All</option>
                {Array.from({ length: 12 }, (_, i) => (
                  <option key={i + 1} value={String(i + 1)}>{new Date(0, i).toLocaleString('default', { month: 'long' })}</option>
                ))}
              </select>
            </label>
            <label className="text-xs text-muted flex flex-col gap-1 flex-1 max-w-[120px]">
              Year
              <input type="number" value={pyYear} onChange={(e) => setPyYear(e.target.value)}
                placeholder="Year..."
                className="px-3 py-1.5 rounded-lg border border-surface-border bg-surface text-sm w-full" />
            </label>
            <label className="text-xs text-muted flex flex-col gap-1 flex-1 min-w-[180px]">
              Department
              <select value={pyDept} onChange={(e) => setPyDept(e.target.value)}
                className="px-3 py-1.5 rounded-lg border border-surface-border bg-surface text-sm w-full">
                <option value="">All</option>
                {['Mathematics', 'Physics', 'Chemistry', 'Biology', 'English', 'Amharic', 'Social Studies', 'Civics', 'ICT'].map(d => (
                  <option key={d} value={d}>{d}</option>
                ))}
              </select>
            </label>
            <Button size="sm" variant="secondary" onClick={exportPayrollCSV} className="flex items-center gap-1.5">
              <FileSpreadsheet className="w-4 h-4" /> Export CSV
            </Button>
          </div>

          <div className="max-h-[480px] overflow-y-auto rounded-xl border border-[var(--border)] bg-surface">
            <Table keyExtractor={(p) => p.id} data={payrolls} emptyMessage="No payroll records matching current filters."
              columns={[
                { key: 'teacher', header: 'Teacher', render: (p) => p.teacher?.user?.name ?? `#${p.teacher_id}` },
                { key: 'dept', header: 'Department', render: (p) => p.teacher?.department ?? '—' },
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
        </div>
      )}

      {/* Reminders Tab */}
      {tab === 'reminders' && (
        <div className="space-y-4">
          <div className="flex justify-between items-center bg-surface border border-surface-border p-4 rounded-xl flex-wrap gap-3">
            <div className="text-sm font-medium text-muted">
              Overdue tuition payment monitoring and system reminder alerts.
            </div>
            <Button size="sm" variant="secondary" onClick={exportOverdueCSV} className="flex items-center gap-1.5">
              <FileSpreadsheet className="w-4 h-4" /> Export CSV
            </Button>
          </div>

          {reminderMessage && <p className="text-sm text-green-600 font-semibold">{reminderMessage}</p>}
          {reminderError && <p className="text-sm text-red-500 font-semibold">{reminderError}</p>}

          <Table keyExtractor={(o) => o.student_id} data={overdues} emptyMessage="No student overdue data found."
            columns={[
              { key: 'student_name', header: 'Student' },
              { key: 'student_code', header: 'Code' },
              { key: 'class_name', header: 'Class' },
              { key: 'academic_year', header: 'Academic Year' },
              {
                key: 'semester_1', header: 'Semester 1',
                render: (o) => (
                  <div className="flex items-center gap-2">
                    <Badge label={o.semester_1} variant={o.semester_1 === 'Paid' ? 'success' : o.semester_1 === 'Pending' ? 'warning' : 'danger'} />
                    {o.semester_1 === 'Overdue' && (
                      <Button size="sm" variant="secondary"
                        loading={reminderLoading[`${o.student_id}-Semester 1`]}
                        onClick={() => handleSendReminder(o.student_id, o.academic_year, 'Semester 1')}
                        className="flex items-center gap-1 py-0.5 px-1.5 text-xs text-accent border border-accent/20 bg-accent/5 hover:bg-accent/10"
                      >
                        <Send className="w-3 h-3" /> Remind
                      </Button>
                    )}
                  </div>
                )
              },
              {
                key: 'semester_2', header: 'Semester 2',
                render: (o) => (
                  <div className="flex items-center gap-2">
                    <Badge label={o.semester_2} variant={o.semester_2 === 'Paid' ? 'success' : o.semester_2 === 'Pending' ? 'warning' : 'danger'} />
                    {o.semester_2 === 'Overdue' && (
                      <Button size="sm" variant="secondary"
                        loading={reminderLoading[`${o.student_id}-Semester 2`]}
                        onClick={() => handleSendReminder(o.student_id, o.academic_year, 'Semester 2')}
                        className="flex items-center gap-1 py-0.5 px-1.5 text-xs text-accent border border-accent/20 bg-accent/5 hover:bg-accent/10"
                      >
                        <Send className="w-3 h-3" /> Remind
                      </Button>
                    )}
                  </div>
                )
              },
              {
                key: 'semester_3', header: 'Semester 3',
                render: (o) => (
                  <div className="flex items-center gap-2">
                    <Badge label={o.semester_3} variant={o.semester_3 === 'Paid' ? 'success' : o.semester_3 === 'Pending' ? 'warning' : 'danger'} />
                    {o.semester_3 === 'Overdue' && (
                      <Button size="sm" variant="secondary"
                        loading={reminderLoading[`${o.student_id}-Semester 3`]}
                        onClick={() => handleSendReminder(o.student_id, o.academic_year, 'Semester 3')}
                        className="flex items-center gap-1 py-0.5 px-1.5 text-xs text-accent border border-accent/20 bg-accent/5 hover:bg-accent/10"
                      >
                        <Send className="w-3 h-3" /> Remind
                      </Button>
                    )}
                  </div>
                )
              }
            ]}
          />
        </div>
      )}

      {/* Create Payroll Modal */}
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
