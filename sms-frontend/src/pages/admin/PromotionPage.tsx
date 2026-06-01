import { useEffect, useState } from 'react'
import { CheckCircle2, AlertTriangle, TrendingUp, Sparkles, HelpCircle } from 'lucide-react'
import {
  getStudents,
  promoteStudent,
  getPromotionPreview,
} from '../../api/admin'
import type { Student } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { DataTable } from '../../components/ui/DataTable'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import { ConfirmModal } from '../../components/ui/ConfirmModal'
import { AlertModal } from '../../components/ui/AlertModal'
import { PageHeader } from '../../components/ui/PageHeader'
import { Badge } from '../../components/ui/Badge'

interface PreviewData {
  promotion_status: string
  failed_subjects: number
  can_promote: boolean
  grade_level: number
  stream: string
}

export default function PromotionPage() {
  const [students, setStudents] = useState<Student[]>([])
  const [loading, setLoading] = useState(true)
  const [promoteId, setPromoteId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)

  // Preview state
  const [previewStudent, setPreviewStudent] = useState<Student | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewData, setPreviewData] = useState<PreviewData | null>(null)

  const [alertState, setAlertState] = useState<{
    open: boolean
    title: string
    message: string
    type: 'success' | 'error'
  }>({
    open: false,
    title: '',
    message: '',
    type: 'success',
  })

  const fetchStudents = async () => {
    setLoading(true)
    try {
      const res = await getStudents({ page_size: 100 })
      const body = res.data
      const payload = body.data
      if (Array.isArray(payload)) {
        setStudents(payload as Student[])
      } else if (payload && typeof payload === 'object' && 'data' in payload) {
        setStudents((payload as { data: Student[] }).data)
      } else {
        setStudents(listFromApi(body))
      }
    } catch {
      setStudents([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchStudents()
  }, [])

  const handleOpenPreview = async (student: Student) => {
    setPreviewStudent(student)
    setPreviewLoading(true)
    setPreviewData(null)
    try {
      const res = await getPromotionPreview(student.id)
      const data = res.data.data
      setPreviewData(data as unknown as PreviewData)
    } catch (err: unknown) {
      const errMsg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Eligibility check failed'
      setAlertState({ open: true, title: 'Error', message: errMsg, type: 'error' })
      setPreviewStudent(null)
    } finally {
      setPreviewLoading(false)
    }
  }

  const handlePromote = async () => {
    const targetId = promoteId || previewStudent?.id
    if (!targetId) return

    setSaving(true)
    try {
      const res = await promoteStudent(targetId)
      setPromoteId(null)
      setPreviewStudent(null)
      fetchStudents()
      setAlertState({
        open: true,
        title: 'Success',
        message: res.data.message || 'Promotion processed successfully.',
        type: 'success',
      })
    } catch (err: unknown) {
      const errMsg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Promotion failed'
      setAlertState({ open: true, title: 'Promotion Failed', message: errMsg, type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <PageHeader
        title="Student Promotion"
        subtitle="Manage year-end grade advancement and academic year transitions"
      />

      <div className="mb-6 p-4 rounded-xl border border-surface-border bg-surface/40 backdrop-blur-md flex items-start gap-3">
        <Sparkles className="w-5 h-5 text-accent shrink-0 mt-0.5" />
        <div className="text-sm text-muted leading-relaxed">
          <span className="font-semibold text-foreground">Promotion Rules:</span> Midterm and Final averages are calculated. 
          A student with <span className="text-danger font-medium">3+ failed subjects</span> repeats the current grade; 
          <span className="text-warning font-medium">1–2 failures</span> result in a conditional promotion with mandatory retakes; 
          all other students advance normally. Auto-enrollment into next-grade subjects is processed instantly upon promotion.
        </div>
      </div>

      <DataTable
        loading={loading}
        data={students}
        keyExtractor={(s) => s.id}
        searchPlaceholder="Search student name, code..."
        filters={[
          {
            key: 'stream',
            label: 'All streams',
            options: [
              { value: 'Natural Science', label: 'Natural Science' },
              { value: 'Social Science', label: 'Social Science' },
            ],
          },
          {
            key: 'grade_level',
            label: 'All grades',
            options: [9, 10, 11, 12].map((g) => ({
              value: String(g),
              label: `Grade ${g}`,
            })),
          },
        ]}
        columns={[
          { key: 'student_code', header: 'Code' },
          { key: 'user', header: 'Name', render: (s) => s.user?.name ?? '—' },
          { key: 'grade_level', header: 'Grade', render: (s) => `Grade ${s.grade_level ?? 9}` },
          { key: 'stream', header: 'Stream', render: (s) => s.stream || 'Common' },
          { key: 'academic_year', header: 'Academic Year', render: (s) => s.academic_year || '—' },
          {
            key: 'promotion_status',
            header: 'Status',
            render: (s) => (
              <Badge
                label={s.promotion_status ?? 'normal'}
                variant={
                  s.promotion_status === 'repeat'
                    ? 'danger'
                    : s.promotion_status === 'conditional'
                    ? 'warning'
                    : 'success'
                }
              />
            ),
          },
          {
            key: 'actions',
            header: 'Actions',
            render: (s) => (
              <div className="flex gap-2">
                <Button size="sm" variant="ghost" onClick={() => handleOpenPreview(s)}>
                  <HelpCircle className="w-3.5 h-3.5 mr-1" /> Check Eligibility
                </Button>
                <Button size="sm" variant="primary" onClick={() => setPromoteId(s.id)}>
                  <TrendingUp className="w-3.5 h-3.5 mr-1" /> Promote
                </Button>
              </div>
            ),
          },
        ]}
      />

      {/* Promotion Eligibility Preview Modal */}
      <Modal
        open={!!previewStudent}
        onClose={() => setPreviewStudent(null)}
        title={`Promotion Eligibility Check`}
      >
        {previewLoading ? (
          <div className="flex justify-center py-8">
            <span className="w-8 h-8 border-3 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        ) : previewData ? (
          <div className="space-y-4">
            <div className="p-4 rounded-lg bg-surface-elevated/40 border border-surface-border flex items-start gap-3">
              {previewData.promotion_status === 'repeat' ? (
                <AlertTriangle className="w-6 h-6 text-danger shrink-0 mt-0.5" />
              ) : previewData.promotion_status === 'conditional' ? (
                <AlertTriangle className="w-6 h-6 text-warning shrink-0 mt-0.5" />
              ) : (
                <CheckCircle2 className="w-6 h-6 text-success shrink-0 mt-0.5" />
              )}
              <div>
                <p className="text-sm font-semibold text-foreground">
                  Student: {previewStudent?.user?.name} ({previewStudent?.student_code})
                </p>
                <div className="mt-2 space-y-1 text-sm text-muted">
                  <p>
                    Eligibility Status:{' '}
                    <span
                      className={`font-semibold capitalize ${
                        previewData.promotion_status === 'repeat'
                          ? 'text-danger'
                          : previewData.promotion_status === 'conditional'
                          ? 'text-warning'
                          : 'text-success'
                      }`}
                    >
                      {previewData.promotion_status}
                    </span>
                  </p>
                  <p>Failed Subjects: <span className="font-semibold text-foreground">{previewData.failed_subjects}</span></p>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="p-3 rounded-lg border border-surface-border bg-surface/30">
                <p className="text-xs text-muted uppercase tracking-wider">Current Placement</p>
                <p className="mt-1 font-semibold text-foreground text-sm">Grade {previewStudent?.grade_level}</p>
                <p className="text-xs text-muted font-mono">{previewStudent?.stream || 'Common Stream'}</p>
              </div>
              <div className="p-3 rounded-lg border border-accent/20 bg-accent/5">
                <p className="text-xs text-accent uppercase tracking-wider font-semibold">Outcome Placement</p>
                <p className="mt-1 font-semibold text-accent text-sm">
                  {previewData.promotion_status === 'repeat'
                    ? `Grade ${previewStudent?.grade_level} (Repeat)`
                    : `Grade ${Math.min(12, (previewStudent?.grade_level ?? 9) + 1)}`}
                </p>
                <p className="text-xs text-muted font-mono">
                  {previewData.promotion_status === 'repeat'
                    ? previewStudent?.stream || 'Common Stream'
                    : previewData.grade_level >= 11
                    ? previewData.stream || 'Natural/Social stream'
                    : 'Common Stream'}
                </p>
              </div>
            </div>

            <div className="text-xs text-muted italic">
              {previewData.promotion_status === 'repeat'
                ? 'Student has 3 or more failed subjects and is required to repeat the current grade.'
                : previewData.promotion_status === 'conditional'
                ? 'Student will be promoted conditionally. Failed subjects will be registered as retakes.'
                : 'Student has passed all subjects and is eligible for normal promotion.'}
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button variant="ghost" onClick={() => setPreviewStudent(null)}>
                Close
              </Button>
              <Button loading={saving} onClick={handlePromote}>
                Confirm & Promote
              </Button>
            </div>
          </div>
        ) : (
          <div className="text-center text-muted py-6">Failed to retrieve preview data.</div>
        )}
      </Modal>

      {/* Confirmation Modal */}
      <ConfirmModal
        open={!!promoteId}
        onClose={() => setPromoteId(null)}
        title="Confirm Student Promotion"
        message="Are you sure you want to run promotion? This checks current year averages, updates the student grade level/status, clears old subject enrollments, and auto-enrolls them in new subjects."
        confirmLabel="Promote"
        loading={saving}
        onConfirm={handlePromote}
      />

      <AlertModal
        open={alertState.open}
        onClose={() => setAlertState({ ...alertState, open: false })}
        title={alertState.title}
        message={alertState.message}
        type={alertState.type}
      />
    </div>
  )
}
