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
      await axios.post(refreshUrl, {}, { withCredentials: true })
      refreshQueue.forEach((cb) => cb())
      refreshQueue = []
      return api(original)
    } catch {
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
