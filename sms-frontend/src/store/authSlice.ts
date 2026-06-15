import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import type { User } from '../types/user'
import { login as apiLogin, logout as apiLogout } from '../api/auth'
import { getMe } from '../api/me'
import api from '../api/axiosClient'

interface AuthState {
  user: User | null
  loading: boolean
  initialized: boolean
  error: string | null
}

const initialState: AuthState = {
  user: null,
  loading: false,
  initialized: false,
  error: null,
}

export const initializeAuth = createAsyncThunk(
  'auth/initialize',
  async (_, { rejectWithValue }) => {
    // Check if there's an access token first
    const accessToken = localStorage.getItem('sms_access_token')
    if (!accessToken) {
      return rejectWithValue('no session')
    }

    // Try to get user info with existing token
    try {
      const res = await getMe()
      return res.data.data ?? null
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status
      // If 401, try to refresh token
      if (status === 401) {
        try {
          const refreshRes = await api.post('/api/token/refresh', {}, { withCredentials: true })
          // Store new access token
          if (refreshRes.data?.data?.access_token) {
            localStorage.setItem('sms_access_token', refreshRes.data.data.access_token)
          }
          // Try getMe again
          const res = await getMe()
          return res.data.data ?? null
        } catch {
          localStorage.removeItem('sms_access_token')
          return rejectWithValue('session expired')
        }
      }
      return rejectWithValue('session expired')
    }
  }
)

export const login = createAsyncThunk(
  'auth/login',
  async ({ email, password }: { email: string; password: string }, { rejectWithValue }) => {
    try {
      const res = await apiLogin(email, password)
      const payload = res.data.data
      if (!payload?.user) return rejectWithValue('Login failed')
      // Store access token for Authorization header
      if (payload.access_token) {
        localStorage.setItem('sms_access_token', payload.access_token)
      }
      return payload.user
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Invalid email or password.'
      return rejectWithValue(msg)
    }
  }
)

export const logout = createAsyncThunk('auth/logout', async () => {
  try {
    await apiLogout()
  } finally {
    localStorage.removeItem('sms_access_token')
  }
})

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    clearError(state) {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(initializeAuth.pending, (state) => {
        state.loading = true
      })
      .addCase(initializeAuth.fulfilled, (state, action) => {
        state.user = action.payload
        state.loading = false
        state.initialized = true
      })
      .addCase(initializeAuth.rejected, (state) => {
        state.user = null
        state.loading = false
        state.initialized = true
      })
      .addCase(login.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(login.fulfilled, (state, action) => {
        state.user = action.payload
        state.loading = false
        state.initialized = true
      })
      .addCase(login.rejected, (state, action) => {
        state.loading = false
        state.error = (action.payload as string) ?? 'Login failed'
      })
      .addCase(logout.fulfilled, (state) => {
        state.user = null
        state.error = null
      })
  },
})

export const { clearError } = authSlice.actions
export default authSlice.reducer
