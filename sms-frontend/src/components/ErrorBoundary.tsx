import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  message: string
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, message: '' }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, message: error.message }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Uncaught render error:', error, info)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex flex-col items-center justify-center gap-4 px-4 bg-[var(--bg)]">
          <div className="text-4xl">⚠️</div>
          <h1 className="text-xl font-bold text-[var(--text-h)]">Something went wrong</h1>
          <p className="text-sm text-[var(--text)] max-w-sm text-center">{this.state.message}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 rounded-lg bg-[var(--accent)] text-white text-sm font-medium hover:opacity-90"
          >
            Reload page
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
