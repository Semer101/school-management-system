import { useAuth } from './useAuth'

export function useRole() {
  const { user } = useAuth()
  const role = user?.role ?? null

  return {
    role,
    isAdmin: role === 'Admin',
    isTeacher: role === 'Teacher',
    isStudent: role === 'Student',
    isParent: role === 'Parent',
  }
}