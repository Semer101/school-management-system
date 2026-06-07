import {
  FileText, Image, Paperclip, FileSpreadsheet,
  Presentation, Archive, File, FileCode,
} from 'lucide-react'

/**
 * File type config maps a file extension to its display properties:
 * icon, background color class, and human-readable label.
 */
const FILE_TYPES: Record<string, { icon: typeof FileText; bg: string; label: string }> = {
  pdf:  { icon: FileText,     bg: 'text-red-500 bg-red-500/10',     label: 'PDF Document' },
  doc:  { icon: FileText,     bg: 'text-blue-500 bg-blue-500/10',   label: 'Word Document' },
  docx: { icon: FileText,     bg: 'text-blue-500 bg-blue-500/10',   label: 'Word Document' },
  xls:  { icon: FileSpreadsheet, bg: 'text-green-600 bg-green-600/10', label: 'Excel Spreadsheet' },
  xlsx: { icon: FileSpreadsheet, bg: 'text-green-600 bg-green-600/10', label: 'Excel Spreadsheet' },
  csv:  { icon: FileSpreadsheet, bg: 'text-green-600 bg-green-600/10', label: 'CSV Spreadsheet' },
  ppt:  { icon: Presentation, bg: 'text-orange-500 bg-orange-500/10', label: 'PowerPoint' },
  pptx: { icon: Presentation, bg: 'text-orange-500 bg-orange-500/10', label: 'PowerPoint' },
  jpg:  { icon: Image,        bg: 'text-purple-500 bg-purple-500/10', label: 'JPEG Image' },
  jpeg: { icon: Image,        bg: 'text-purple-500 bg-purple-500/10', label: 'JPEG Image' },
  png:  { icon: Image,        bg: 'text-purple-500 bg-purple-500/10', label: 'PNG Image' },
  gif:  { icon: Image,        bg: 'text-purple-500 bg-purple-500/10', label: 'GIF Image' },
  webp: { icon: Image,        bg: 'text-purple-500 bg-purple-500/10', label: 'WebP Image' },
  svg:  { icon: Image,        bg: 'text-purple-500 bg-purple-500/10', label: 'SVG Image' },
  zip:  { icon: Archive,      bg: 'text-amber-600 bg-amber-600/10',  label: 'ZIP Archive' },
  rar:  { icon: Archive,      bg: 'text-amber-600 bg-amber-600/10',  label: 'RAR Archive' },
  '7z': { icon: Archive,      bg: 'text-amber-600 bg-amber-600/10',  label: '7-Zip Archive' },
  txt:  { icon: File,         bg: 'text-gray-500 bg-gray-500/10',    label: 'Text File' },
  md:   { icon: FileCode,     bg: 'text-slate-500 bg-slate-500/10',  label: 'Markdown' },
  json: { icon: FileCode,     bg: 'text-slate-500 bg-slate-500/10',  label: 'JSON' },
}

const FALLBACK = { icon: Paperclip, bg: 'text-muted bg-surface-elevated', label: 'File' }

export function getFileTypeInfo(type: string) {
  const key = type.toLowerCase().replace(/^\./, '')
  return FILE_TYPES[key] ?? FALLBACK
}

export function getFileTypeLabel(type: string): string {
  return getFileTypeInfo(type).label
}

/** Displays a colored icon chip for a given file type. */
export function FileTypeIcon({ type, size = 'md' }: { type: string; size?: 'sm' | 'md' }) {
  const info = getFileTypeInfo(type)
  const Icon = info.icon
  const cls = size === 'sm' ? 'w-4 h-4' : 'w-5 h-5'
  return (
    <div className={cn(
      'flex items-center justify-center rounded-lg shrink-0',
      size === 'sm' ? 'w-8 h-8' : 'w-10 h-10',
      info.bg
    )}>
      <Icon className={cls} />
    </div>
  )
}

/** Returns the file extension in uppercase, e.g. "PDF", "DOCX", "PNG". */
export function getFileExtension(type: string): string {
  return type.replace(/^\./, '').toUpperCase()
}

/** Format bytes to a human-readable string */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

/** Format an ISO date string to a readable format */
export function formatDate(dateStr: string): string {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

function cn(...classes: (string | boolean | undefined | null)[]) {
  return classes.filter(Boolean).join(' ')
}