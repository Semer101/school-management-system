import { useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import {
  CheckCircle2,
  AlertTriangle,
  TrendingUp,
  Sparkles,
  HelpCircle,
  ArrowLeft,
  BookOpen,
  Award,
  AlertCircle,
} from 'lucide-react'
import {
  getStudents,
  promoteStudent,
  getPromotionPreview,
  getStudent,
  getStudentEnrollmentStatus,
} from '../../api/admin'
import { getStudentGrades } from '../../api/academics'
import type { Student, Grade } from '../../types/academic'
import { DataTable } from '../../components/ui/DataTable'
import { Button } from '../../components/ui/Button'
import { AlertModal } from '../../components/ui/AlertModal'
import { PageHeader } from '../../components/ui/PageHeader'
import { Badge } from '../../components/ui/Badge'
import { Spinner } from '../../components/ui/Spinner'

interface PreviewData {
  promotion_status: string
  failed_subjects: number
  can_promote: boolean
  grade_level: number
  stream: string
}

export default function PromotionPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const [students, setStudents] = useState<Student[]>([])
  const [loading, setLoading] = useState(true)
  const [promoteId, setPromoteId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)

  // Selected Student State
  const [selectedStudentId, setSelectedStudentId] = useState<number | null>(null)
  const [studentDetails, setStudentDetails] = useState<Student | null>(null)
  const [enrolledSubjects, setEnrolledSubjects] = useState<any[]>([])
  const [studentGrades, setStudentGrades] = useState<Grade[]>([])
  const [loadingDetails, setLoadingDetails] = useState(false)
  const [previewData, setPreviewData] = useState<PreviewData | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)

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
        setStudents([])
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

  // Detect studentId query parameter
  useEffect(() => {
    const params = new URLSearchParams(location.search)
    const sId = params.get('studentId')
    if (sId) {
      const numId = Number(sId)
      setSelectedStudentId(numId)
      loadStudentPromotionDetails(numId)
    } else {
      setSelectedStudentId(null)
      setStudentDetails(null)
      setEnrolledSubjects([])
      setStudentGrades([])
      setPreviewData(null)
      setValidationError(null)
    }
  }, [location.search])

  const loadStudentPromotionDetails = async (id: number) => {
    setLoadingDetails(true)
    setValidationError(null)
    setPreviewData(null)
    try {
      const [stuRes, enrollRes, gradesRes] = await Promise.all([
        getStudent(id),
        getStudentEnrollmentStatus(id),
        getStudentGrades(id),
      ])

      const studentData = stuRes.data.data
      setStudentDetails(studentData ?? null)

      const enrolled = (enrollRes.data.data ?? []).filter((r: any) => r.enrolled)
      setEnrolledSubjects(enrolled)
      setStudentGrades(gradesRes.data.data ?? [])

      // Load preview or capture validation error
      try {
        const previewRes = await getPromotionPreview(id)
        setPreviewData(previewRes.data.data as unknown as PreviewData)
      } catch (err: any) {
        const errMsg = err.response?.data?.error ?? 'Student is ineligible for promotion.'
        setValidationError(errMsg)
      }
    } catch (err) {
      console.error(err)
      setValidationError('Failed to load student academic details.')
    } finally {
      setLoadingDetails(false)
    }
  }

  const handlePromote = async () => {
    const targetId = selectedStudentId || promoteId
    if (!targetId) return

    setSaving(true)
    try {
      const res = await promoteStudent(targetId)
      setPromoteId(null)
      fetchStudents()
      setAlertState({
        open: true,
        title: 'Success',
        message: res.data.message || 'Promotion processed successfully.',
        type: 'success',
      })
      // Clear URL and selection to go back to list
      navigate('/admin/promotion')
    } catch (err: any) {
      const errMsg = err.response?.data?.error ?? 'Promotion failed'
      setAlertState({ open: true, title: 'Promotion Failed', message: errMsg, type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  // Calculate semester average scores per subject
  const subjectGradesMap = enrolledSubjects.map((sub) => {
    const subGrades = studentGrades.filter((g) => g.subject_id === sub.subject_id)

    const getSemGrades = (sem: string) => {
      const semGrades = subGrades.filter((g) => g.semester === sem)
      const midterm = semGrades.find((g) => g.grade_type === 'Midterm' || g.grade_type === 'Exam' || g.semester === sem && (g as any).type === 'Midterm')?.score
      const final = semGrades.find((g) => g.grade_type === 'Final' || g.semester === sem && (g as any).type === 'Final')?.score

      const otherAvg =
        semGrades.length > 0
          ? semGrades.reduce((sum, g) => sum + g.score, 0) / semGrades.length
          : null

      return { midterm, final, avg: otherAvg }
    }

    const sem1 = getSemGrades('Semester 1')
    const sem2 = getSemGrades('Semester 2')
    const sem3 = getSemGrades('Semester 3')

    const completedSems = [sem1.avg, sem2.avg, sem3.avg].filter((v) => v !== null) as number[]
    const overallAvg =
      completedSems.length > 0
        ? completedSems.reduce((sum, v) => sum + v, 0) / completedSems.length
        : 0

    return {
      ...sub,
      sem1,
      sem2,
      sem3,
      overallAvg,
      status: overallAvg >= 50 ? 'Pass' : 'Fail',
    }
  })

  // Check individual rules locally
  const isEnrolledYearValid =
    studentDetails &&
    studentDetails.enrolled_at &&
    new Date(studentDetails.enrolled_at).getFullYear() <= (studentDetails.academic_year ?? 2025)

  const isSemesterCompletionValid = !validationError || !validationError.includes('missing grades')

  // Detailed Student Assessment View
  if (selectedStudentId) {
    if (loadingDetails) {
      return (
        <div className="flex justify-center items-center h-[50vh]">
          <Spinner />
        </div>
      )
    }

    if (!studentDetails) {
      return (
        <div className="space-y-4">
          <Button variant="ghost" onClick={() => navigate('/admin/promotion')}>
            <ArrowLeft className="w-4 h-4 mr-2" /> Back
          </Button>
          <div className="p-6 text-center text-muted bg-surface rounded-xl border border-surface-border">
            Student profile not found.
          </div>
        </div>
      )
    }

    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between gap-3">
          <Button variant="ghost" onClick={() => navigate('/admin/promotion')}>
            <ArrowLeft className="w-4 h-4 mr-2" /> Back to Students
          </Button>
          <h1 className="text-xl font-bold text-foreground">Promotion Evaluation</h1>
        </div>

        {/* Student Profile Info */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="md:col-span-2 space-y-6">
            <div className="p-6 bg-surface border border-surface-border rounded-2xl shadow-sm space-y-4">
              <div className="flex justify-between items-start border-b border-surface-border pb-4">
                <div>
                  <h2 className="text-lg font-bold text-foreground">{studentDetails.user?.name}</h2>
                  <p className="text-xs text-muted font-mono mt-1">Code: {studentDetails.student_code}</p>
                </div>
                <div className="flex gap-2">
                  <Badge label={`Grade ${studentDetails.grade_level}`} variant="accent" />
                  <Badge label={studentDetails.stream || 'Common Stream'} variant="default" />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-muted block text-xs">Current Academic Year</span>
                  <span className="font-medium text-foreground">{studentDetails.academic_year}</span>
                </div>
                <div>
                  <span className="text-muted block text-xs">Enrollment Year</span>
                  <span className="font-medium text-foreground">
                    {studentDetails.enrolled_at ? new Date(studentDetails.enrolled_at).getFullYear() : '—'}
                  </span>
                </div>
                <div>
                  <span className="text-muted block text-xs">Email</span>
                  <span className="font-medium text-foreground">{studentDetails.parent_email || '—'}</span>
                </div>
                <div>
                  <span className="text-muted block text-xs">Phone</span>
                  <span className="font-medium text-foreground">{studentDetails.parent_phone || '—'}</span>
                </div>
              </div>
            </div>

            {/* Enrolled Subjects & Semester Grades Details */}
            <div className="bg-surface border border-surface-border rounded-2xl overflow-hidden shadow-sm">
              <div className="px-6 py-4 border-b border-surface-border flex items-center justify-between bg-surface-elevated/40">
                <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                  <BookOpen className="w-4 h-4 text-accent" /> Enrolled Subjects & Semester Grades
                </h3>
                <span className="text-xs text-muted font-mono">{subjectGradesMap.length} Subjects Enrolled</span>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm text-left border-collapse">
                  <thead>
                    <tr className="bg-surface-elevated/20 text-xs font-mono uppercase tracking-wider text-muted border-b border-surface-border">
                      <th className="px-6 py-3 font-semibold">Subject</th>
                      <th className="px-6 py-3 font-semibold">Sem 1 Avg</th>
                      <th className="px-6 py-3 font-semibold">Sem 2 Avg</th>
                      <th className="px-6 py-3 font-semibold">Sem 3 Avg</th>
                      <th className="px-6 py-3 font-semibold">Final Avg</th>
                      <th className="px-6 py-3 font-semibold">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {subjectGradesMap.length === 0 ? (
                      <tr>
                        <td colSpan={6} className="text-center py-8 text-muted">
                          No subjects enrolled.
                        </td>
                      </tr>
                    ) : (
                      subjectGradesMap.map((sub) => (
                        <tr key={sub.subject_id} className="border-b border-surface-border hover:bg-surface-elevated/20 last:border-0">
                          <td className="px-6 py-4">
                            <div className="font-medium text-foreground">{sub.subject_name}</div>
                            <div className="text-[10px] text-muted font-mono">{sub.subject_code}</div>
                          </td>
                          <td className="px-6 py-4 text-muted">
                            {sub.sem1.avg !== null ? `${sub.sem1.avg.toFixed(1)}%` : '—'}
                          </td>
                          <td className="px-6 py-4 text-muted">
                            {sub.sem2.avg !== null ? `${sub.sem2.avg.toFixed(1)}%` : '—'}
                          </td>
                          <td className="px-6 py-4 text-muted">
                            {sub.sem3.avg !== null ? `${sub.sem3.avg.toFixed(1)}%` : '—'}
                          </td>
                          <td className="px-6 py-4 font-semibold text-foreground">
                            {sub.overallAvg.toFixed(1)}%
                          </td>
                          <td className="px-6 py-4">
                            <Badge
                              label={sub.status}
                              variant={sub.status === 'Pass' ? 'success' : 'danger'}
                            />
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          {/* Right Column: Promotion Checks & Rules */}
          <div className="space-y-6">
            {/* Outcomes Placement Card */}
            <div className="p-6 bg-surface border border-surface-border rounded-2xl shadow-sm space-y-4">
              <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                <Award className="w-4 h-4 text-accent" /> Outcome Placement
              </h3>

              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 rounded-lg border border-surface-border bg-surface/30">
                  <p className="text-[10px] text-muted uppercase tracking-wider font-mono">Current Placement</p>
                  <p className="mt-1 font-semibold text-foreground text-sm">Grade {studentDetails.grade_level}</p>
                  <p className="text-[10px] text-muted">{studentDetails.stream || 'Common Curriculum'}</p>
                </div>
                <div className="p-3 rounded-lg border border-accent/20 bg-accent/5">
                  <p className="text-[10px] text-accent uppercase tracking-wider font-semibold font-mono">Next Placement</p>
                  <p className="mt-1 font-semibold text-accent text-sm">
                    {previewData?.promotion_status === 'repeat'
                      ? `Grade ${studentDetails.grade_level} (Repeat)`
                      : `Grade ${Math.min(12, (studentDetails.grade_level ?? 9) + 1)}`}
                  </p>
                  <p className="text-[10px] text-muted">
                    {previewData?.promotion_status === 'repeat'
                      ? studentDetails.stream || 'Common Stream'
                      : (studentDetails.grade_level ?? 9) + 1 >= 11
                      ? previewData?.stream || studentDetails.stream || 'Natural/Social stream'
                      : 'Common Stream'}
                  </p>
                </div>
              </div>
            </div>

            {/* Validation Rules Checklist Card */}
            <div className="p-6 bg-surface border border-surface-border rounded-2xl shadow-sm space-y-4">
              <h3 className="text-sm font-semibold text-foreground">Promotion Requirements Checklist</h3>
              <div className="space-y-3.5 text-sm">
                <div className="flex items-start gap-3">
                  {isEnrolledYearValid ? (
                    <CheckCircle2 className="w-5 h-5 text-success shrink-0 mt-0.5" />
                  ) : (
                    <AlertTriangle className="w-5 h-5 text-danger shrink-0 mt-0.5" />
                  )}
                  <div>
                    <p className="font-medium text-foreground">Valid Enrollment Date</p>
                    <p className="text-xs text-muted mt-0.5">
                      Enrolled year must be less than or equal to current academic year.
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-3">
                  {isSemesterCompletionValid ? (
                    <CheckCircle2 className="w-5 h-5 text-success shrink-0 mt-0.5" />
                  ) : (
                    <AlertTriangle className="w-5 h-5 text-danger shrink-0 mt-0.5" />
                  )}
                  <div>
                    <p className="font-medium text-foreground">Complete Semester Grades</p>
                    <p className="text-xs text-muted mt-0.5">
                      Student must have grades recorded for Semester 1, 2, and 3 for all enrolled subjects.
                    </p>
                  </div>
                </div>
              </div>
            </div>

            {/* Results & Status Alert Banner */}
            {validationError ? (
              <div className="p-4 rounded-xl border border-danger-border bg-danger-bg text-danger flex items-start gap-3">
                <AlertCircle className="w-5 h-5 shrink-0 mt-0.5" />
                <div className="text-xs space-y-1">
                  <p className="font-bold">Ineligible for Promotion</p>
                  <p className="leading-relaxed">{validationError}</p>
                </div>
              </div>
            ) : previewData ? (
              <div
                className={`p-4 rounded-xl border flex items-start gap-3
                  ${
                    previewData.promotion_status === 'repeat'
                      ? 'border-danger-border bg-danger-bg text-danger'
                      : previewData.promotion_status === 'conditional'
                      ? 'border-warning-border bg-warning-bg text-warning'
                      : 'border-success-border bg-success-bg text-success'
                  }`}
              >
                {previewData.promotion_status === 'repeat' ? (
                  <AlertTriangle className="w-5 h-5 shrink-0 mt-0.5" />
                ) : (
                  <CheckCircle2 className="w-5 h-5 shrink-0 mt-0.5" />
                )}
                <div className="text-xs space-y-1">
                  <p className="font-bold uppercase tracking-wider font-mono">
                    Evaluation Result: {previewData.promotion_status}
                  </p>
                  <p className="leading-relaxed">
                    {previewData.promotion_status === 'repeat'
                      ? `Student has failed ${previewData.failed_subjects} subjects (< 50% average) and is required to repeat Grade ${studentDetails.grade_level}.`
                      : previewData.promotion_status === 'conditional'
                      ? `Student has failed ${previewData.failed_subjects} subject(s). Promoted conditionally with mandatory retakes.`
                      : 'Eligible for grade advancement. Academic performance requirements met successfully.'}
                  </p>
                </div>
              </div>
            ) : null}

            {/* Big Action Buttons */}
            <div className="flex gap-2">
              <Button
                className="flex-1 justify-center py-2.5"
                variant="primary"
                loading={saving}
                disabled={!!validationError}
                onClick={handlePromote}
              >
                <TrendingUp className="w-4 h-4 mr-2" /> Confirm & Execute Promotion
              </Button>
            </div>
          </div>
        </div>

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

  // Standard Student list view
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
                <Button size="sm" variant="ghost" onClick={() => navigate(`/admin/promotion?studentId=${s.id}`)}>
                  <HelpCircle className="w-3.5 h-3.5 mr-1" /> Evaluate Promotion
                </Button>
              </div>
            ),
          },
        ]}
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
