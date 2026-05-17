import api from './axiosClient'
import type { APIResponse } from '../types/api'
import type { AttendancePercentage, Grade, ReportCard, Student } from '../types/academic'
import type { Transaction } from '../types/finance'

export const getMyChildren = () =>
  api.get<APIResponse<Student[]>>('/api/parent/children')

export const getChildAttendance = (studentID: number) =>
  api.get<APIResponse<AttendancePercentage[]>>(`/api/parent/attendance/${studentID}`)

export const getChildGrades = (studentID: number) =>
  api.get<APIResponse<Grade[]>>(`/api/parent/grades/${studentID}`)

export const getChildReportCard = (studentID: number) =>
  api.get<APIResponse<ReportCard>>(`/api/parent/reportcard/${studentID}`)

export const downloadChildReportCard = (studentID: number) =>
  api.get(`/api/parent/reportcard/${studentID}/pdf`, { responseType: 'blob' })

export const submitParentReceipt = (data: {
  amount: number
  receipt_id: string
  description?: string
}) => api.post<APIResponse<Transaction>>('/api/parent/finance/receipt', data)

export const getParentTransactions = () =>
  api.get<APIResponse<Transaction[]>>('/api/parent/finance/transactions')
