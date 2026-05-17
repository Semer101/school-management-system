import api from './axiosClient'
import type { APIResponse } from '../types/api'
import type { User } from '../types/user'

export const getMe = () =>
  api.get<APIResponse<User>>('/api/me')

export const changePassword = (current_password: string, new_password: string) =>
  api.put<APIResponse>('/api/me/password', { current_password, new_password })
