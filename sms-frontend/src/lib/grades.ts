import type { Grade } from '../types/academic'

export function gradeTypeLabel(g: Grade): string {
  return g.grade_type || g.type || '—'
}
