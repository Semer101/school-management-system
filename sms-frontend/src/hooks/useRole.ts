import { useAppSelector } from '../store/hooks'

export function useRole() {
  const user = useAppSelector((s) => s.auth.user)
  const role = user?.role ?? null

  return {
    role,
    isAdmin: role === 'Admin',
    isTeacher: role === 'Teacher',
    isStudent: role === 'Student',
    isParent: role === 'Parent',
  }
}
