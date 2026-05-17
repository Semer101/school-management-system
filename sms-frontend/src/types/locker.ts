import type { Student } from './academic'

export interface LockerFile {
  id: number
  student_id: number
  student?: Student
  file_name: string
  file_size: number
  file_type: string
  category: 'Certificate' | 'Assignment' | 'Portfolio'
  is_public: boolean
  uploaded_at: string
  created_at: string
}