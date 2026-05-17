import { createContext, useState, useEffect, type ReactNode } from 'react'
import type { User } from '../types/user'
import { login as apiLogin, logout as apiLogout } from '../api/auth'
import { getMe } from '../api/me'

interface AuthContextValue {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue>({
  user: null,
  loading: true,
  login: async () => {},
  logout: async () => {},
})

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  // Restore session on mount
  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token) {
      setLoading(false)
      return
    }
    getMe()
      .then((res) => setUser(res.data.data ?? null))
      .catch(() => {
        localStorage.removeItem('access_token')
        setUser(null)
      })
      .finally(() => setLoading(false))
  }, [])

  const login = async (email: string, password: string) => {
    const res = await apiLogin(email, password)
    const payload = res.data.data
    if (!payload) throw new Error('Login failed')
    localStorage.setItem('access_token', payload.access_token)
    setUser(payload.user)
  }

  const logout = async () => {
    try {
      await apiLogout()
    } finally {
      localStorage.removeItem('access_token')
      setUser(null)
    }
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}