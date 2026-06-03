import type { Student } from './academic'
import type { Teacher } from './academic'

export interface Transaction {
  id: number
  student_id: number
  student?: Student
  amount: number
  receipt_id: string
  type: 'Tuition' | 'Expense' | 'Payroll'
  status: 'Pending' | 'Verified' | 'Rejected'
  description: string
  created_by: number
  verified_by: number
  verified_at: string | null
  academic_year: number
  semester: string
  created_at: string
  updated_at: string
}

export interface Payroll {
  id: number
  teacher_id: number
  teacher?: Teacher
  amount: number
  month: number
  year: number
  status: 'Pending' | 'Paid'
  paid_at: string | null
  created_at: string
}