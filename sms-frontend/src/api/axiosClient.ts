import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
  withCredentials: true,
  timeout: 30000,
  headers: {
    'X-Requested-With': 'XMLHttpRequest',
  },
})

let isRefreshing = false
let refreshQueue: Array<() => void> = []

// Add access token from localStorage to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('sms_access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config
    if (!original || error.response?.status !== 401 || original._retry) {
      return Promise.reject(error)
    }

    if (original.url?.includes('/api/login') || original.url?.includes('/api/token/refresh')) {
      return Promise.reject(error)
    }

    original._retry = true

    if (isRefreshing) {
      return new Promise((resolve) => {
        refreshQueue.push(() => resolve(api(original)))
      })
    }

    isRefreshing = true
    try {
      const base = import.meta.env.VITE_API_BASE_URL ?? ''
      const refreshUrl = base ? `${base}/api/token/refresh` : '/api/token/refresh'
      const res = await axios.post(refreshUrl, {}, { withCredentials: true })
      
      // Store new access token
      if (res.data?.data?.access_token) {
        localStorage.setItem('sms_access_token', res.data.data.access_token)
      }
      
      refreshQueue.forEach((cb) => cb())
      refreshQueue = []
      return api(original)
    } catch {
      localStorage.removeItem('sms_access_token')
      if (!window.location.pathname.startsWith('/login')) {
        window.location.href = '/login'
      }
      return Promise.reject(error)
    } finally {
      isRefreshing = false
    }
  }
)

export default api
