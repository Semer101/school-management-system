import { useEffect, useState, type FormEvent } from 'react'
import { getClasses, createClass, archiveClass, getTeachers } from '../../api/admin'
import type { Class, Teacher } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { Table } from '../../components/ui/Table'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Modal } from '../../components/ui/Modal'

export default function ClassesPage() {
  const [classes, setClasses] = useState<Class[]>([])
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [form, setForm] = useState({ name: '', year: new Date().getFullYear(), teacher_id: 0 })
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    Promise.all([getClasses(), getTeachers()])
      .then(([c, t]) => { setClasses(listFromApi(c.data)); setTeachers(listFromApi(t.data)) })
      .finally(() => setLoading(false))
  }, [])

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      const res = await createClass(form)
      const created = res.data.data
      if (created) setClasses((prev) => [...prev, created])
      setModalOpen(false)
      setForm({ name: '', year: new Date().getFullYear(), teacher_id: 0 })
    } finally {
      setSaving(false)
    }
  }

  const handleArchive = async (id: number) => {
    if (!confirm('Archive this class?')) return
    await archiveClass(id)
    setClasses((prev) => prev.filter((c) => c.id !== id))
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-[var(--text-h)]">Classes</h1>
        <Button onClick={() => setModalOpen(true)}>+ New Class</Button>
      </div>

      <Table
        loading={loading}
        keyExtractor={(c) => c.id}
        data={classes}
        columns={[
          { key: 'name', header: 'Name' },
          { key: 'year', header: 'Year' },
          { key: 'teacher', header: 'Homeroom Teacher', render: (c) => c.teacher?.name ?? '—' },
          {
            key: 'actions', header: '',
            render: (c) => (
              <Button size="sm" variant="danger" onClick={() => handleArchive(c.id)}>Archive</Button>
            ),
          },
        ]}
      />

      <Modal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        title="Create Class"
        footer={
          <Button form="class-form" type="submit" loading={saving}>Create</Button>
        }
      >
        <form id="class-form" onSubmit={handleCreate} className="flex flex-col gap-4">
          <Input label="Class Name" placeholder="Grade 10A" value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} required />
          <Input label="Year" type="number" value={form.year}
            onChange={(e) => setForm((f) => ({ ...f, year: Number(e.target.value) }))} required />
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Homeroom Teacher</label>
            <select
              value={form.teacher_id}
              onChange={(e) => setForm((f) => ({ ...f, teacher_id: Number(e.target.value) }))}
              className="w-full px-3 py-2 rounded-lg text-sm bg-[var(--bg)] border border-[var(--border)] text-[var(--text-h)] outline-none focus:border-[var(--accent)]"
              required
            >
              <option value={0} disabled>Select teacher...</option>
              {teachers.map((t) => (
                <option key={t.id} value={t.id}>{t.user?.name ?? `Teacher #${t.id}`}</option>
              ))}
            </select>
          </div>
        </form>
      </Modal>
    </div>
  )
}