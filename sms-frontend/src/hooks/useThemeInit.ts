import { useEffect } from 'react'
import { useAppSelector } from '../store/hooks'

export function useThemeInit() {
  const theme = useAppSelector((s) => s.ui.theme)
  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }, [theme])
}
