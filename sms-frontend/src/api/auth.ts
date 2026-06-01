import api from './axiosClient'

export const forgotPassword = (email: string) =>
  api.post('/api/forgot-password', { email })

export const resetPasswordWithOTP = (email: string, otp: string, new_password: string) =>
  api.post('/api/reset-password', { email, otp, new_password })
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
