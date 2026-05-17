export type Role = 'Admin' | 'Teacher' | 'Student' | 'Parent'

export interface User {
  id: number
  name: string
  email: string
  role: Role
  phone: string
  avatar_url: string
  is_active: boolean
  created_at: string
  updated_at: string
}