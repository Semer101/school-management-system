import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Navbar } from './Navbar'
import { useAppSelector, useAppDispatch } from '../../store/hooks'
import { setSidebarCollapsed } from '../../store/uiSlice'

export function AppShell() {
  const sidebarCollapsed = useAppSelector((state) => state.ui.sidebarCollapsed)
  const dispatch = useAppDispatch()

  return (
    <div className="flex min-h-screen app-grid-bg">
      {/* Sidebar Navigation */}
      <Sidebar />

      {/* Mobile Backdrop Overlay */}
      {!sidebarCollapsed && (
        <div
          className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm lg:hidden transition-opacity duration-300"
          onClick={() => dispatch(setSidebarCollapsed(true))}
        />
      )}

      {/* Main Content Area */}
      <div className="flex flex-col flex-1 min-w-0 min-h-screen">
        <Navbar />
        <main className="flex-1 p-4 sm:p-6 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
