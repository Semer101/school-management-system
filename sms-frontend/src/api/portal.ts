import axios from 'axios'
import api from './axiosClient'
import type { APIResponse } from '../types/api'

export type PortalContext = {
  academic_year: number
  active_semester: string
}

export type HealthStatus = {
  status: string
  service: string
}

export const getPortalContext = () =>
  api.get<APIResponse<PortalContext>>('/api/portal/context')

export const checkHealth = async (): Promise<'online' | 'offline'> => {
  try {
    const base = import.meta.env.VITE_API_BASE_URL ?? ''
    const url = base ? `${base}/health` : '/health'
    const res = await axios.get<HealthStatus>(url, { timeout: 5000 })
    return res.data?.status === 'ok' ? 'online' : 'offline'
  } catch {
    return 'offline'
  }
}
