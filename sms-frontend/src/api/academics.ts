import api from './axiosClient'
import type { APIResponse } from '../types/api'
import type { Attendance, AttendancePercentage, Grade, ReportCard } from '../types/academic'

// ── Attendance ───────────────────────────────────────────
export const recordAttendance = (data: {
  student_id: number
  subject_id: number
  date: string
  status: 'Present' | 'Absent' | 'Late'
}) => api.post<APIResponse<Attendance>>('/api/academics/attendance', data)

export const getClassAttendance = (classID: number) =>
  api.get<APIResponse<Attendance[]>>(`/api/academics/attendance/class/${classID}`)

export const getAttendancePercentage = (studentID: number) =>
  api.get<APIResponse<AttendancePercentage[]>>(`/api/academics/attendance/${studentID}`)

// ── Grades ───────────────────────────────────────────────
export const bulkGradeEntry = (grades: {
  student_id: number
  subject_id: number
  score: number
  grade_type: string
  term: string
  remarks?: string
}[]) => api.post<APIResponse>('/api/academics/grades/bulk', { grades })

export const getSubjectGrades = (subjectID: number) =>
  api.get<APIResponse<Grade[]>>(`/api/academics/grades/subject/${subjectID}`)

export const getStudentGrades = (studentID: number) =>
  api.get<APIResponse<Grade[]>>(`/api/academics/grades/student/${studentID}`)

// ── Report Card ──────────────────────────────────────────
export const getReportCard = (studentID: number) =>
  api.get<APIResponse<ReportCard>>(`/api/academics/reportcard/${studentID}`)

export const downloadReportCard = (studentID: number) =>
  api.get(`/api/academics/reportcard/${studentID}/pdf`, { responseType: 'blob' })
