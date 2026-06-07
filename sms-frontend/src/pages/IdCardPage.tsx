import { useEffect, useRef, useState } from 'react'
import { useAuth } from '../hooks/useAuth'
import { useRole } from '../hooks/useRole'
import { Button } from '../components/ui/Button'
import { PageHeader } from '../components/ui/PageHeader'
import { Spinner } from '../components/ui/Spinner'
import { QRCodeSVG } from 'qrcode.react'
import { Printer, Download } from 'lucide-react'
import { getMyChildren } from '../api/parent'
import { listFromApi } from '../types/api'
import type { Student } from '../types/academic'
const API_BASE = import.meta.env.VITE_API_BASE_URL ?? ''

interface CardUser {
  id: number
  name: string
  email: string
  role: string
  code?: string
  avatar_url?: string
}

function IdCard({ user, student }: { user: CardUser; student?: Student }) {
  const profileUrl = user.avatar_url ? `${API_BASE}${user.avatar_url}` : null
  const qrData = JSON.stringify({
    id: user.id,
    name: user.name,
    role: user.role,
    code: user.code,
    email: user.email,
  })

  return (
    <div className="id-card w-[340px] bg-white rounded-2xl shadow-xl border border-gray-200 overflow-hidden print:shadow-none print:border-2">
      {/* Header stripe */}
      <div className="bg-gradient-to-r from-accent to-cyan-600 px-5 py-4 text-white text-center">
        <div className="flex items-center justify-center gap-2 mb-1">
          <div className="w-8 h-8 rounded-lg bg-white/20 border border-white/30 flex items-center justify-center">
            <span className="text-white font-mono text-sm font-bold">S</span>
          </div>
          <p className="text-sm font-bold tracking-wide">SMS Portal</p>
        </div>
        <p className="text-[10px] uppercase tracking-widest opacity-80">Ethiopia G9–12</p>
      </div>

      {/* Photo */}
      <div className="flex justify-center -mt-10">
        {profileUrl ? (
          <img
            src={profileUrl}
            alt={user.name}
            className="w-20 h-20 rounded-full border-4 border-white shadow-md object-cover bg-gray-100"
          />
        ) : (
          <div className="w-20 h-20 rounded-full border-4 border-white shadow-md bg-accent/10 flex items-center justify-center text-accent text-2xl font-bold">
            {user.name[0]?.toUpperCase()}
          </div>
        )}
      </div>

      {/* Details */}
      <div className="px-5 py-4 text-center space-y-1.5">
        <h3 className="text-base font-bold text-gray-900">{user.name}</h3>
        <p className="text-xs font-semibold uppercase tracking-wider text-accent">{user.role}</p>
        {user.code && (
          <p className="text-xs font-mono text-gray-500">ID: {user.code}</p>
        )}
        <p className="text-xs text-gray-500">{user.email}</p>
        {student && (
          <>
            <div className="border-t border-gray-100 my-2" />
            <p className="text-xs text-gray-600">
              Grade {student.grade_level} · {student.stream || '—'}
            </p>
            {student.class && (
              <p className="text-[11px] text-gray-400">{student.class.name}</p>
            )}
          </>
        )}
      </div>

      {/* QR code */}
      <div className="flex justify-center pb-4">
        <div className="bg-white p-1.5 rounded-lg border border-gray-200">
          <QRCodeSVG value={qrData} size={80} level="M" />
        </div>
      </div>

      {/* Footer */}
      <div className="bg-gray-50 px-5 py-2.5 text-center border-t border-gray-200">
        <p className="text-[9px] uppercase tracking-widest text-gray-400">
          Valid for Academic Year 2025
        </p>
      </div>
    </div>
  )
}

export default function IdCardPage() {
  const { user: authUser } = useAuth()
  const { role } = useRole()
  const printRef = useRef<HTMLDivElement>(null)
  const [children, setChildren] = useState<Student[]>([])
  const [selectedChild, setSelectedChild] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (role !== 'Parent') return
    getMyChildren()
      .then((res) => {
        const kids = listFromApi(res.data)
        setChildren(kids)
        if (kids.length > 0) setSelectedChild(kids[0].id)
      })
      .finally(() => setLoading(false))
  }, [role])

  if (loading) return <Spinner fullPage />

  if (!authUser || !role) return null

  const buildUser = (overrides?: Partial<CardUser>): CardUser => ({
    id: authUser.id,
    name: authUser.name,
    email: authUser.email,
    role: overrides?.role ?? role,
    code: overrides?.code,
    avatar_url: authUser.avatar_url,
  })

  const handlePrint = () => {
    const printWindow = window.open('', '_blank')
    if (!printWindow) { window.print(); return }

    const cardEl = printRef.current?.querySelector('.id-card')
    if (!cardEl) return

    printWindow.document.write(`
      <html>
        <head>
          <title>ID Card</title>
          <style>
            body { margin: 0; display: flex; justify-content: center; align-items: center; min-height: 100vh; background: #fff; }
            .id-card { width: 340px; background: white; border-radius: 16px; border: 2px solid #e5e7eb; overflow: hidden; font-family: system-ui, sans-serif; }
            .id-card * { box-sizing: border-box; }
            @media print {
              @page { margin: 0; size: 95mm 140mm; }
              body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
            }
          </style>
        </head>
        <body>${cardEl.outerHTML}</body>
      </html>
    `)
    printWindow.document.close()
    printWindow.focus()
    setTimeout(() => { printWindow.print() }, 500)
  }

  const pdfCard = () => {
    const cardEl = printRef.current?.querySelector('.id-card')
    if (!cardEl) return

    const win = window.open('', '_blank')
    if (!win) return

    win.document.write(`
      <html>
        <head>
          <title>ID Card</title>
          <style>
            body { margin: 0; display: flex; justify-content: center; align-items: center; min-height: 100vh; background: #1e1e2e; }
            .id-card { width: 340px; background: white; border-radius: 16px; border: 2px solid #e5e7eb; overflow: hidden; font-family: system-ui, sans-serif; }
            @media print {
              @page { margin: 0; size: 95mm 140mm; }
              body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
            }
          </style>
        </head>
        <body>${cardEl.outerHTML}</body>
      </html>
    `)
    win.document.close()
    win.focus()
    setTimeout(() => { win.print() }, 500)
  }

  const currentStudent = children.find((c) => c.id === selectedChild)

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <PageHeader
        title="Identification Card"
        subtitle="Generate and print your official school ID card"
        action={
          <div className="flex gap-2">
            <Button variant="secondary" size="sm" onClick={handlePrint}>
              <Printer className="w-4 h-4 mr-1" /> Print
            </Button>
            <Button variant="secondary" size="sm" onClick={pdfCard}>
              <Download className="w-4 h-4 mr-1" /> PDF
            </Button>
          </div>
        }
      />

      {/* Child selector for Parent */}
      {role === 'Parent' && children.length > 0 && (
        <div className="bg-surface border border-surface-border rounded-xl p-4">
          <label className="text-sm font-medium text-foreground">Select Child</label>
          <select
            value={selectedChild ?? ''}
            onChange={(e) => setSelectedChild(Number(e.target.value))}
            className="w-full mt-1 px-3 py-2 rounded-lg border border-surface-border bg-surface text-sm focus:outline-none focus:border-accent/50"
          >
            {children.map((c) => (
              <option key={c.id} value={c.id}>{c.user?.name ?? `Student #${c.id}`}</option>
            ))}
          </select>
        </div>
      )}

      {/* ID Card Display */}
      <div ref={printRef} className="flex justify-center py-4">
        {role === 'Admin' && <IdCard user={buildUser({ role: 'Admin' })} />}
        {role === 'Teacher' && <IdCard user={buildUser({ role: 'Teacher' })} />}
        {role === 'Student' && <IdCard user={buildUser({ role: 'Student' })} />}
        {role === 'Parent' && currentStudent && (
          <IdCard
            user={buildUser({ role: 'Parent', code: currentStudent.student_code })}
            student={currentStudent}
          />
        )}
        {role === 'Parent' && children.length === 0 && (
          <p className="text-sm text-muted">No linked children found.</p>
        )}
      </div>
    </div>
  )
}