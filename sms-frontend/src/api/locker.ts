import api from './axiosClient'
import type { APIResponse } from '../types/api'
import type { LockerFile } from '../types/locker'

export const uploadLockerFile = (formData: FormData) =>
  api.post<APIResponse<LockerFile>>('/api/locker/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })

export const getMyLockerFiles = () =>
  api.get<APIResponse<LockerFile[]>>('/api/locker/my-files')

export const deleteLockerFile = (fileID: number) =>
  api.delete<APIResponse>(`/api/locker/files/${fileID}`)

export const toggleFileVisibility = (fileID: number, is_public: boolean) =>
  api.patch<APIResponse>(`/api/locker/files/${fileID}/visibility`, { is_public })

// Teacher view
export const getStudentPublicFiles = (studentID: number) =>
  api.get<APIResponse<LockerFile[]>>(`/api/locker/student/${studentID}/public`)
