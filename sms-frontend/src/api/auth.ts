import api from './axiosClient'
import type { APIResponse } from '../types/api'
import type { User } from '../types/user'

export interface LoginResponse {
  access_token: string
  user: User
}

export const login = (email: string, password: string) =>
  api.post<APIResponse<LoginResponse>>('/api/login', { email, password })

export const logout = () =>
  api.post<APIResponse>('/api/logout')

export const refreshToken = () =>
  api.post<APIResponse<{ access_token: string }>>('/api/token/refresh')
