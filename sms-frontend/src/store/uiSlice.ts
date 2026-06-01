import { createSlice, type PayloadAction } from '@reduxjs/toolkit'

export type ThemeMode = 'light' | 'dark'

interface UiState {
  sidebarCollapsed: boolean
  theme: ThemeMode
}

const stored = localStorage.getItem('sms_theme') as ThemeMode | null
const initialState: UiState = {
  sidebarCollapsed: false,
  theme: stored === 'dark' ? 'dark' : 'light',
}

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    setSidebarCollapsed(state, action: PayloadAction<boolean>) {
      state.sidebarCollapsed = action.payload
    },
    toggleSidebar(state) {
      state.sidebarCollapsed = !state.sidebarCollapsed
    },
    setTheme(state, action: PayloadAction<ThemeMode>) {
      state.theme = action.payload
      localStorage.setItem('sms_theme', action.payload)
      document.documentElement.classList.toggle('dark', action.payload === 'dark')
    },
    toggleTheme(state) {
      const next = state.theme === 'light' ? 'dark' : 'light'
      state.theme = next
      localStorage.setItem('sms_theme', next)
      document.documentElement.classList.toggle('dark', next === 'dark')
    },
  },
})

export const { setSidebarCollapsed, toggleSidebar, setTheme, toggleTheme } = uiSlice.actions
export default uiSlice.reducer
