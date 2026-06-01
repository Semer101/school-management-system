import type { User } from './user'

export interface Class {
  id: number
  name: string
  grade_level?: number
  section?: string
  stream?: string
  status?: string
  year: number
  teacher_id: number
  teacher?: Teacher
  students?: Student[]
  created_at: string
}

export interface Subject {
  id: number
  name: string
  code: string
  grade_level?: number
  stream?: string
  status?: string
  teacher_id: number
  teacher?: Teacher
  created_at: string
}

export interface Student {
  id: number
  user_id: number
  user?: User
  class_id: number
  class?: Class
  parent_id: number
  student_code: string
  parent_name: string
  parent_email: string
  parent_phone: string
  date_of_birth: string
  enrolled_at: string
  stream?: string
  grade_level?: number
  promotion_status?: string
  academic_year?: number
}

export interface Teacher {
  id: number
  user_id: number
  user?: User
  teacher_code: string
  qualification: string
  joined_at: string
}

export interface Enrollment {
  id: number
  student_id: number
  subject_id: number
  student?: Student
  subject?: Subject
}

export interface Attendance {
  id: number
  student_id: number
  subject_id?: number | null
  date: string
  status: 'Present' | 'Absent' | 'Late'
  student?: Student
  subject?: Subject
}

export interface AttendancePercentage {
  student_id: number
  subject_id: number
  subject_name: string
  present: number
  total: number
  percentage: number
}

export interface Grade {
  id: number
  student_id: number
  subject_id: number
  score: number
  grade_type: string
  term: string
  remarks: string
  student?: Student
  subject?: Subject
  created_at: string
}

export interface ReportCard {
  student: Student
  grades: Grade[]
  attendance: AttendancePercentage[]
  average: number
  term: string
  year: number
}