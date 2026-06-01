import api from './axiosClient'
import type { APIResponse } from '../types/api'

export type TrashEntity = 'students' | 'teachers' | 'classes' | 'subjects' | 'users'

export const listTrash = (entity: TrashEntity, page = 1, pageSize = 20) =>
  api.get<APIResponse<{ total: number; data: unknown[]; entity: string }>>(
    '/api/admin/trash',
    { params: { entity, page, page_size: pageSize } }
  )

export const restoreTrash = (entity: TrashEntity, id: number) =>
  api.post<APIResponse>(`/api/admin/trash/${entity}/${id}/restore`)

export const permanentDelete = (entity: TrashEntity, id: number, password: string) =>
  api.delete<APIResponse>(`/api/admin/trash/${entity}/${id}/permanent`, { data: { password } })
