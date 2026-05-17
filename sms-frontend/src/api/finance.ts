import api from './axiosClient'
import type { APIResponse } from '../types/api'
import type { Transaction, Payroll } from '../types/finance'

// ── Student / Parent ─────────────────────────────────────
export const submitReceipt = (data: {
  amount: number
  receipt_id: string
  description?: string
}) => api.post<APIResponse<Transaction>>('/api/finance/receipt', data)

export const getMyTransactions = () =>
  api.get<APIResponse<Transaction[]>>('/api/finance/transactions')

// ── Admin ────────────────────────────────────────────────
export const getAllTransactions = () =>
  api.get<APIResponse<Transaction[]>>('/api/admin/finance/summary')

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
