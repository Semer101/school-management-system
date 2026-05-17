import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from './context/AuthContext'
import { ProtectedRoute } from './components/layout/ProtectedRoute'
import { AppShell } from './components/layout/AppShell'

// Pages
import LoginPage from './pages/LoginPage'
import DashboardPage from './pages/dashboard/DashboardPage'
import ProfilePage from './pages/profile/ProfilePage'
import UsersPage from './pages/admin/UsersPage'
import StudentsPage from './pages/admin/StudentsPage'
import TeachersPage from './pages/admin/TeachersPage'
import ClassesPage from './pages/admin/ClassesPage'
import SubjectsPage from './pages/admin/SubjectsPage'
import EnrollmentPage from './pages/admin/EnrollmentPage'
import AttendanceSummary from './pages/admin/AttendanceSummary'
import NotifyPage from './pages/admin/NotifyPage'
import AdminFinancePage from './pages/finance/AdminFinancePage'
import AttendancePage from './pages/academics/AttendancePage'
import GradesPage from './pages/academics/GradesPage'
import ReportCardPage from './pages/academics/ReportCardPage'
import MyFinancePage from './pages/finance/MyFinancePage'
import MyLockerPage from './pages/locker/MyLockerPage'
import StudentLockerPage from './pages/locker/StudentLockerPage'
import NotificationsPage from './pages/notifications/NotificationsPage'
import ChildrenPage from './pages/parent/ChildrenPage'
import ChildDetailPage from './pages/parent/ChildDetailPage'

const queryClient = new QueryClient()

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            {/* Public */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<Navigate to="/dashboard" replace />} />

            {/* Protected — all authenticated users */}
            <Route element={<ProtectedRoute><AppShell /></ProtectedRoute>}>
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/profile" element={<ProfilePage />} />
              <Route path="/notifications" element={<NotificationsPage />} />

              {/* Admin only */}
              <Route path="/admin/users" element={
                <ProtectedRoute allowedRoles={['Admin']}><UsersPage /></ProtectedRoute>
              } />
              <Route path="/admin/students" element={
                <ProtectedRoute allowedRoles={['Admin']}><StudentsPage /></ProtectedRoute>
              } />
              <Route path="/admin/teachers" element={
                <ProtectedRoute allowedRoles={['Admin']}><TeachersPage /></ProtectedRoute>
              } />
              <Route path="/admin/classes" element={
                <ProtectedRoute allowedRoles={['Admin']}><ClassesPage /></ProtectedRoute>
              } />
              <Route path="/admin/subjects" element={
                <ProtectedRoute allowedRoles={['Admin']}><SubjectsPage /></ProtectedRoute>
              } />
              <Route path="/admin/enrollment" element={
                <ProtectedRoute allowedRoles={['Admin']}><EnrollmentPage /></ProtectedRoute>
              } />
              <Route path="/admin/attendance-summary" element={
                <ProtectedRoute allowedRoles={['Admin']}><AttendanceSummary /></ProtectedRoute>
              } />
              <Route path="/admin/notify" element={
                <ProtectedRoute allowedRoles={['Admin']}><NotifyPage /></ProtectedRoute>
              } />
              <Route path="/admin/finance" element={
                <ProtectedRoute allowedRoles={['Admin']}><AdminFinancePage /></ProtectedRoute>
              } />

              {/* Teacher + Admin + Student */}
              <Route path="/academics/attendance" element={
                <ProtectedRoute allowedRoles={['Admin', 'Teacher', 'Student']}><AttendancePage /></ProtectedRoute>
              } />
              <Route path="/academics/grades" element={
                <ProtectedRoute allowedRoles={['Admin', 'Teacher', 'Student']}><GradesPage /></ProtectedRoute>
              } />
              <Route path="/academics/reportcard" element={
                <ProtectedRoute allowedRoles={['Student']}><ReportCardPage /></ProtectedRoute>
              } />

              {/* Student + Parent */}
              <Route path="/finance" element={
                <ProtectedRoute allowedRoles={['Student', 'Parent']}><MyFinancePage /></ProtectedRoute>
              } />

              {/* Student locker */}
              <Route path="/locker" element={
                <ProtectedRoute allowedRoles={['Student']}><MyLockerPage /></ProtectedRoute>
              } />

              {/* Teacher + Admin locker view */}
              <Route path="/locker/student" element={
                <ProtectedRoute allowedRoles={['Admin', 'Teacher']}><StudentLockerPage /></ProtectedRoute>
              } />

              {/* Parent */}
              <Route path="/parent/children" element={
                <ProtectedRoute allowedRoles={['Parent']}><ChildrenPage /></ProtectedRoute>
              } />
              <Route path="/parent/children/:id" element={
                <ProtectedRoute allowedRoles={['Parent']}><ChildDetailPage /></ProtectedRoute>
              } />
            </Route>

            {/* Catch-all */}
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  )
}