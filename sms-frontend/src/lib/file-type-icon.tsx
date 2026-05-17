import { FileText, Image, Paperclip } from 'lucide-react'

export function FileTypeIcon({ type }: { type: string }) {
  const cls = 'w-5 h-5 shrink-0'
  if (type === 'pdf') return <FileText className={cls + ' text-accent'} />
  if (['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(type)) {
    return <Image className={cls + ' text-accent'} />
  }
  return <Paperclip className={cls + ' text-muted'} />
}
