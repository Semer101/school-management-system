import api from './axiosClient'
import type { APIResponse } from '../types/api'

export interface DashboardKPIs {
  students: number
  teachers: number
  classes: number
  subjects: number
  present_today: number
  absent_today: number
  pending_transactions: number
}

export interface AnalyticsData {
  kpis: {
    students: number
    teachers: number
    parents: number
    notifications: number
    revenue_etb: number
    pending_etb: number
    payroll_paid: number
    payroll_pending: number
  }
  students_by_grade: { grade: number; count: number }[]
  students_by_stream: { stream: string; count: number }[]
  grade_averages: { subject_name: string; average: number }[]
  attendance_breakdown: { status: string; count: number }[]
  monthly_attendance: { month: string; present: number; absent: number }[]
  promotion_distribution: { status: string; count: number }[]
}

export const getDashboardKPIs = () =>
  api.get<APIResponse<DashboardKPIs>>('/api/admin/dashboard/kpis')

export const getAnalytics = () =>
  api.get<APIResponse<AnalyticsData>>('/api/admin/analytics')
