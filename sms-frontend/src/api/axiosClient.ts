import axios from 'axios'
import type { APIResponse } from '../types/api'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
  withCredentials: true, // sends sms_refresh HttpOnly cookie
  timeout: 10000,        // 10 s — prevents indefinite spinner on cold-start / network issues
})

// Attach Bearer token to every request
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let isRefreshing = false
let refreshQueue: Array<(token: string) => void> = []

// Auto-refresh on 401
api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config

    if (error.response?.status === 401 && !original._retry) {
      original._retry = true

      if (isRefreshing) {
        // Queue requests while refresh is in-flight
        return new Promise((resolve) => {
          refreshQueue.push((token) => {
            original.headers.Authorization = `Bearer ${token}`
            resolve(api(original))
          })
        })
      }

      isRefreshing = true

      try {
        const base = import.meta.env.VITE_API_BASE_URL ?? ''
        const refreshUrl = base ? `${base}/api/token/refresh` : '/api/token/refresh'
        const { data } = await axios.post<APIResponse<{ access_token: string }>>(
          refreshUrl,
          {},
          { withCredentials: true }
        )
        const newToken: string = data.data?.access_token ?? ''
        if (!newToken) throw new Error('missing access token')
        localStorage.setItem('access_token', newToken)

        refreshQueue.forEach((cb) => cb(newToken))
        refreshQueue = []

        original.headers.Authorization = `Bearer ${newToken}`
        return api(original)
      } catch {
        localStorage.removeItem('access_token')
        window.location.href = '/login'
        return Promise.reject(error)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  }
)

export default api
