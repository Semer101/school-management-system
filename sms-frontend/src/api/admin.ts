import api from './axiosClient'
import type { APIResponse, PaginatedList } from '../types/api'
import type { Student, Teacher, Class, Subject } from '../types/academic'
import type { User, Role } from '../types/user'

// ── Users ────────────────────────────────────────────────
export const registerUser = (data: {
  name: string
  email: string
  password: string
  role: Role
  phone?: string
}) => api.post<APIResponse<User>>('/api/admin/register', data)

// ── Students ─────────────────────────────────────────────
export const getStudents = () =>
  api.get<APIResponse<Student[]>>('/api/admin/students')

export const getStudent = (id: number) =>
  api.get<APIResponse<Student>>(`/api/admin/students/${id}`)

export const createStudent = (data: Partial<Student> & { user_id: number }) =>
  api.post<APIResponse<Student>>('/api/admin/students', data)

export const updateStudent = (id: number, data: Partial<Student>) =>
  api.put<APIResponse<Student>>(`/api/admin/students/${id}`, data)

export const archiveStudent = (id: number) =>
  api.delete<APIResponse>(`/api/admin/students/${id}`)

// ── Teachers ─────────────────────────────────────────────
export const getTeachers = () =>
  api.get<APIResponse<Teacher[]>>('/api/admin/teachers')

export const getTeacher = (id: number) =>
  api.get<APIResponse<Teacher>>(`/api/admin/teachers/${id}`)

export const createTeacher = (data: Partial<Teacher> & { user_id: number }) =>
  api.post<APIResponse<Teacher>>('/api/admin/teachers', data)

export const updateTeacher = (id: number, data: Partial<Teacher>) =>
  api.put<APIResponse<Teacher>>(`/api/admin/teachers/${id}`, data)

export const archiveTeacher = (id: number) =>
  api.delete<APIResponse>(`/api/admin/teachers/${id}`)

// ── Classes ──────────────────────────────────────────────
export const getClasses = () =>
  api.get<APIResponse<Class[]>>('/api/admin/classes')

export const createClass = (data: { name: string; year: number; teacher_id: number }) =>
  api.post<APIResponse<Class>>('/api/admin/classes', data)

export const updateClass = (id: number, data: Partial<Class>) =>
  api.put<APIResponse<Class>>(`/api/admin/classes/${id}`, data)

export const archiveClass = (id: number) =>
  api.delete<APIResponse>(`/api/admin/classes/${id}`)

// ── Subjects ─────────────────────────────────────────────
export const getSubjects = () =>
  api.get<APIResponse<Subject[]>>('/api/admin/subjects')

export const createSubject = (data: { name: string; code: string; teacher_id: number }) =>
  api.post<APIResponse<Subject>>('/api/admin/subjects', data)

export const updateSubject = (id: number, data: Partial<Subject>) =>
  api.put<APIResponse<Subject>>(`/api/admin/subjects/${id}`, data)

export const archiveSubject = (id: number) =>
  api.delete<APIResponse>(`/api/admin/subjects/${id}`)

// ── Enrollment ───────────────────────────────────────────
export const enrollStudent = (student_id: number, subject_id: number) =>
  api.post<APIResponse>('/api/admin/enroll', { student_id, subject_id })

export const unenrollStudent = (student_id: number, subject_id: number) =>
  api.delete<APIResponse>('/api/admin/unenroll', { data: { student_id, subject_id } })

// ── Attendance Summary ───────────────────────────────────
export type AttendanceSummaryRow = {
  student_name: string
  student_code: string
  subject_name: string
  present: number
  absent: number
  late: number
  total: number
  percentage: number
}

export const getAttendanceSummary = () =>
  api.get<APIResponse<AttendanceSummaryRow[] | PaginatedList<AttendanceSummaryRow>>>(
    '/api/admin/attendance/summary'
  )

// ── Admin Locker ─────────────────────────────────────────
export const adminGetLockerFiles = (studentID: number) =>
  api.get<APIResponse>(`/api/admin/locker/student/${studentID}`)

// ── Notifications ────────────────────────────────────────
export const broadcastAnnouncement = (data: {
  title: string
  body: string
  target_roles: string[]
}) => api.post<APIResponse>('/api/admin/notify/broadcast', data)

export const notifyAbsentParents = () =>
  api.post<APIResponse>('/api/admin/notify/absences')
