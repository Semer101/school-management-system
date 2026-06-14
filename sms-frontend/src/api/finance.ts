import api from './axiosClient'
import type { APIResponse } from '../types/api'
import type { Transaction, Payroll } from '../types/finance'

// ── Student / Parent ─────────────────────────────────────
export const submitReceipt = (data: {
  amount: number
  receipt_id: string
  description?: string
  semester?: string
}) => api.post<APIResponse<Transaction>>('/api/finance/receipt', data)

export const getMyTransactions = () =>
  api.get<APIResponse<Transaction[]>>('/api/finance/transactions')

// ── Parent: Receipt image upload ─────────────────────────
export const uploadReceipt = (formData: FormData) =>
  api.post<APIResponse<Transaction>>('/api/parent/finance/receipts', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })

export const getReceipt = (id: number) =>
  api.get<APIResponse<Transaction>>(`/api/parent/finance/receipts/${id}`)

// ── Admin: Receipt moderation ────────────────────────────
export const listPendingReceipts = (params?: {
  status?: string
  academic_year?: string
  semester?: string
  search?: string
}) => api.get<APIResponse<Transaction[]>>('/api/admin/finance/receipts', { params })

export const approveReceipt = (id: number) =>
  api.patch<APIResponse<Transaction>>(`/api/admin/finance/receipts/${id}/approve`)

export const rejectReceipt = (id: number, notes: string) =>
  api.patch<APIResponse<Transaction>>(`/api/admin/finance/receipts/${id}/reject`, { notes })

// ── Admin ────────────────────────────────────────────────
export const getAllTransactions = (params?: { academic_year?: string; semester?: string; status?: string; student?: string }) =>
  api.get<APIResponse<Transaction[]>>('/api/admin/finance/summary', { params })

export const verifyReceipt = (id: number) =>
  api.patch<APIResponse>(`/api/admin/finance/receipt/${id}/verify`)

export const createPayroll = (data: {
  teacher_id: number
  amount: number
  month: number
  year: number
}) => api.post<APIResponse<Payroll>>('/api/admin/finance/payroll', data)

export const markPayrollPaid = (id: number) =>
  api.patch<APIResponse>(`/api/admin/finance/payroll/${id}/pay`)

export const getPayrolls = (params?: { month?: string; year?: string; department?: string }) =>
  api.get<APIResponse<Payroll[]>>('/api/admin/finance/payroll', { params })

export interface OverduePaymentRow {
  student_id: number
  student_name: string
  student_code: string
  class_name: string
  academic_year: number
  semester_1: 'Paid' | 'Pending' | 'Overdue'
  semester_2: 'Paid' | 'Pending' | 'Overdue'
  semester_3: 'Paid' | 'Pending' | 'Overdue'
}

export const getOverduePayments = () =>
  api.get<APIResponse<OverduePaymentRow[]>>('/api/admin/finance/overdue')

export const sendPaymentReminder = (data: { student_id: number; academic_year: number; semester: string }) =>
  api.post<APIResponse>('/api/admin/finance/remind', data)