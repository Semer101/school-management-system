import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AnimatePresence } from 'framer-motion'
import { ProtectedRoute } from './components/layout/ProtectedRoute'
import { AppShell } from './components/layout/AppShell'
import { PageTransition } from './components/motion/PageTransition'

import LoginPage from './pages/LoginPage'
import DashboardPage from './pages/dashboard/DashboardPage'
import ProfilePage from './pages/profile/ProfilePage'
import StudentsPage from './pages/admin/StudentsPage'
import PromotionPage from './pages/admin/PromotionPage'
import AnalyticsPage from './pages/admin/AnalyticsPage'
import TrashPage from './pages/admin/TrashPage'
import ParentsPage from './pages/admin/ParentsPage'
import AdminsPage from './pages/admin/AdminsPage'
import ForgotPasswordPage from './pages/ForgotPasswordPage'
import IdCardPage from './pages/IdCardPage'
import TeachersPage from './pages/admin/TeachersPage'
import ClassesPage from './pages/admin/ClassesPage'
import SubjectsPage from './pages/admin/SubjectsPage'
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

function AnimatedRoutes() {
  const location = useLocation()

  return (
    <AnimatePresence mode="wait">
      <Routes location={location} key={location.pathname}>
        <Route path="/login" element={<PageTransition><LoginPage /></PageTransition>} />
        <Route path="/forgot-password" element={<PageTransition><ForgotPasswordPage /></PageTransition>} />
        <Route path="/" element={<Navigate to="/dashboard" replace />} />

        <Route element={<ProtectedRoute><AppShell /></ProtectedRoute>}>
          <Route path="/dashboard" element={<PageTransition><DashboardPage /></PageTransition>} />
          <Route path="/profile" element={<PageTransition><ProfilePage /></PageTransition>} />
          <Route path="/notifications" element={<PageTransition><NotificationsPage /></PageTransition>} />

          <Route path="/admin/students" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><StudentsPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/promotion" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><PromotionPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/admins" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><AdminsPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/parents" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><ParentsPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/analytics" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><AnalyticsPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/trash" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><TrashPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/teachers" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><TeachersPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/classes" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><ClassesPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/subjects" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><SubjectsPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/attendance-summary" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><AttendanceSummary /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/notify" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><NotifyPage /></PageTransition></ProtectedRoute>} />
          <Route path="/admin/finance" element={<ProtectedRoute allowedRoles={['Admin']}><PageTransition><AdminFinancePage /></PageTransition></ProtectedRoute>} />

          <Route path="/academics/attendance" element={<ProtectedRoute allowedRoles={['Admin', 'Teacher', 'Student']}><PageTransition><AttendancePage /></PageTransition></ProtectedRoute>} />
          <Route path="/academics/grades" element={<ProtectedRoute allowedRoles={['Admin', 'Teacher', 'Student']}><PageTransition><GradesPage /></PageTransition></ProtectedRoute>} />
          <Route path="/academics/reportcard" element={<ProtectedRoute allowedRoles={['Student', 'Parent']}><PageTransition><ReportCardPage /></PageTransition></ProtectedRoute>} />
          <Route path="/id-card" element={<ProtectedRoute allowedRoles={['Admin', 'Teacher', 'Student', 'Parent']}><PageTransition><IdCardPage /></PageTransition></ProtectedRoute>} />

          <Route path="/finance" element={<ProtectedRoute allowedRoles={['Parent']}><PageTransition><MyFinancePage /></PageTransition></ProtectedRoute>} />
          <Route path="/locker" element={<ProtectedRoute allowedRoles={['Student']}><PageTransition><MyLockerPage /></PageTransition></ProtectedRoute>} />
          <Route path="/locker/student" element={<ProtectedRoute allowedRoles={['Admin', 'Teacher']}><PageTransition><StudentLockerPage /></PageTransition></ProtectedRoute>} />

          <Route path="/parent/children" element={<ProtectedRoute allowedRoles={['Parent']}><PageTransition><ChildrenPage /></PageTransition></ProtectedRoute>} />
          <Route path="/parent/children/:id" element={<ProtectedRoute allowedRoles={['Parent']}><PageTransition><ChildDetailPage /></PageTransition></ProtectedRoute>} />
        </Route>

        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </AnimatePresence>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AnimatedRoutes />
      </BrowserRouter>
    </QueryClientProvider>
  )
}
