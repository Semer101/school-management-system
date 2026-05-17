import { useEffect, useState, type FormEvent } from 'react'
import { getSubjects, createSubject, archiveSubject, getTeachers } from '../../api/admin'
import type { Subject, Teacher } from '../../types/academic'
import { listFromApi } from '../../types/api'
import { Table } from '../../components/ui/Table'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Modal } from '../../components/ui/Modal'

export default function SubjectsPage() {
  const [subjects, setSubjects] = useState<Subject[]>([])
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [form, setForm] = useState({ name: '', code: '', teacher_id: 0 })
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    Promise.all([getSubjects(), getTeachers()])
      .then(([s, t]) => { setSubjects(listFromApi(s.data)); setTeachers(listFromApi(t.data)) })
      .finally(() => setLoading(false))
  }, [])

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      const res = await createSubject(form)
      const created = res.data.data
      if (created) setSubjects((prev) => [...prev, created])
      setModalOpen(false)
      setForm({ name: '', code: '', teacher_id: 0 })
    } finally {
      setSaving(false)
    }
  }

  const handleArchive = async (id: number) => {
    if (!confirm('Archive this subject?')) return
    await archiveSubject(id)
    setSubjects((prev) => prev.filter((s) => s.id !== id))
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-[var(--text-h)]">Subjects</h1>
        <Button onClick={() => setModalOpen(true)}>+ New Subject</Button>
      </div>

      <Table
        loading={loading}
        keyExtractor={(s) => s.id}
        data={subjects}
        columns={[
          { key: 'code', header: 'Code' },
          { key: 'name', header: 'Name' },
          { key: 'teacher', header: 'Teacher', render: (s) => s.teacher?.name ?? '—' },
          {
            key: 'actions', header: '',
            render: (s) => (
              <Button size="sm" variant="danger" onClick={() => handleArchive(s.id)}>Archive</Button>
            ),
          },
        ]}
      />

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title="Create Subject"
        footer={<Button form="subject-form" type="submit" loading={saving}>Create</Button>}
      >
        <form id="subject-form" onSubmit={handleCreate} className="flex flex-col gap-4">
          <Input label="Subject Name" placeholder="Mathematics" value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} required />
          <Input label="Code" placeholder="MATH101" value={form.code}
            onChange={(e) => setForm((f) => ({ ...f, code: e.target.value }))} required />
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-[var(--text-h)]">Teacher</label>
            <select value={form.teacher_id}
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