import type { ReactNode } from 'react'

interface Column<T> {
  key: string
  header: string
  render?: (row: T) => ReactNode
}

interface TableProps<T> {
  columns: Column<T>[]
  data: T[]
  keyExtractor: (row: T) => string | number
  loading?: boolean
  emptyMessage?: string
}

export function Table<T>({
  columns,
  data,
  keyExtractor,
  loading,
  emptyMessage = 'No data found.',
}: TableProps<T>) {
  return (
    <div className="w-full overflow-x-auto rounded-xl border border-[var(--border)]">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className="bg-[var(--code-bg)]">
            {columns.map((col) => (
              <th
                key={col.key}
                className="text-left px-4 py-3 font-semibold text-[var(--text-h)] border-b border-[var(--border)] whitespace-nowrap"
              >
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr>
              <td colSpan={columns.length} className="text-center py-10">
                <span className="inline-block w-5 h-5 border-2 border-[var(--accent)] border-t-transparent rounded-full animate-spin" />
              </td>
            </tr>
          ) : data.length === 0 ? (
            <tr>
              <td
                colSpan={columns.length}
                className="text-center py-10 text-[var(--text)]"
              >
                {emptyMessage}
              </td>
            </tr>
          ) : (
            data.map((row) => (
              <tr
                key={keyExtractor(row)}
                className="border-b border-[var(--border)] hover:bg-[var(--code-bg)] transition-colors"
              >
                {columns.map((col) => (
                  <td key={col.key} className="px-4 py-3 text-[var(--text)]">
                    {col.render
                      ? col.render(row)
                      : String((row as Record<string, unknown>)[col.key] ?? '')}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}