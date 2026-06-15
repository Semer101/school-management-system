import { StrictMode, useRef, useEffect } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from 'react-redux'
import { store } from './store'
import { initializeAuth } from './store/authSlice'
import './index.css'
import 'flowbite'
import App from './App'
import { ErrorBoundary } from './components/ErrorBoundary'

function Bootstrap() {
  // Guard against StrictMode double-fire in React 18
  const initialized = useRef(false)

  useEffect(() => {
    const theme = localStorage.getItem('sms_theme')
    document.documentElement.classList.toggle('dark', theme === 'dark')

    if (!initialized.current) {
      initialized.current = true
      store.dispatch(initializeAuth())
    }
  }, [])

  return <App />
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary>
      <Provider store={store}>
        <Bootstrap />
      </Provider>
    </ErrorBoundary>
  </StrictMode>
)