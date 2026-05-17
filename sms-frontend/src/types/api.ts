/** Matches sms-backend helpers.APIResponse */
export interface APIResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
  error?: string
}

/** Backend list endpoints return either a bare array or { total, data: T[] } */
export type PaginatedList<T> = { total: number; data: T[] }

export function listFromApi<T>(body: APIResponse<PaginatedList<T> | T[]>): T[] {
  const payload = body.data
  if (!payload) return []
  if (Array.isArray(payload)) return payload
  return payload.data ?? []
}
