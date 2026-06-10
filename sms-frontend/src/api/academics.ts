import api from './axiosClient'
import type { APIResponse } from '../types/api'
import type { Attendance, Grade, ReportCard } from '../types/academic'

// ── Attendance ───────────────────────────────────────────
export const recordAttendance = (data: {
  student_id: number
  date: string
  status: 'Present' | 'Absent' | 'Late'
  notes?: string
}) => api.post<APIResponse<Attendance>>('/api/academics/attendance', data)

export const getClassAttendance = (classID: number, date: string) =>
  api.get<APIResponse<Attendance[]>>(`/api/academics/attendance/class/${classID}`, { params: { date } })

export type AttendanceStats = {
  student_id: string
  overall_percentage: number
  total_days: number
  attended: number
  by_month: { month: string; total: number; present: number; percentage: number }[]
}

export const getAttendancePercentage = (studentID: number) =>
  api.get<APIResponse<AttendanceStats>>(`/api/academics/attendance/${studentID}`)

// ── Grades ───────────────────────────────────────────────
export const bulkGradeEntry = (grades: {
  student_id: number
  subject_id: number
  score: number
  grade_type: string
  semester: string
  remarks?: string
}[]) => api.post<APIResponse>('/api/academics/grades/bulk', { grades })

export const getSubjectGrades = (subjectID: number) =>
  api.get<APIResponse<Grade[]>>(`/api/academics/grades/subject/${subjectID}`)

export const getStudentGrades = (studentID: number, semester?: string, subjectID?: string) => {
  const params = new URLSearchParams()
  if (semester) params.append('semester', semester)
  if (subjectID) params.append('subject_id', subjectID)
  const query = params.toString() ? `?${params.toString()}` : ''
  return api.get<APIResponse<Grade[]>>(`/api/academics/grades/student/${studentID}${query}`)
}

// ── Report Card ──────────────────────────────────────────
export const getReportCard = (studentID: number) =>
  api.get<APIResponse<ReportCard>>(`/api/academics/reportcard/${studentID}`)

export const downloadReportCard = (studentID: number) =>
  api.get(`/api/academics/reportcard/${studentID}/pdf`, { responseType: 'blob' })

export const getTeacherDashboardKPIs = () =>
  api.get<APIResponse<{ students: number; classes: number; subjects: number; attendance_rate: number }>>('/api/academics/dashboard/kpis')
