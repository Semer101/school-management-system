import { useMemo, useState, type ReactNode } from 'react'
import { ChevronLeft, ChevronRight, Search } from 'lucide-react'
import { cn } from '../../lib/utils'

export interface Column<T> {
  key: string
  header: string
  render?: (row: T) => ReactNode
  filterable?: boolean
}

interface DataTableProps<T> {
  columns: Column<T>[]
  data: T[]
  keyExtractor: (row: T) => string | number
  loading?: boolean
  emptyMessage?: string
  searchPlaceholder?: string
  searchKeys?: (keyof T | string)[]
  filters?: { key: string; label: string; options: { value: string; label: string }[] }[]
  pageSize?: number
  toolbar?: ReactNode
}

export function DataTable<T extends object>({
  columns,
  data,
  keyExtractor,
  loading,
  emptyMessage = 'No data found.',
  searchPlaceholder = 'Search...',
  searchKeys,
  filters = [],
  pageSize = 10,
  toolbar,
}: DataTableProps<T>) {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [filterValues, setFilterValues] = useState<Record<string, string>>({})

  const filtered = useMemo(() => {
    let rows = [...data]
    const q = search.trim().toLowerCase()
    if (q) {
      rows = rows.filter((row) => {
        if (searchKeys?.length) {
          return searchKeys.some((k) => {
            const v = String((row as Record<string, unknown>)[k as string] ?? '').toLowerCase()
            return v.includes(q)
          })
        }
        return JSON.stringify(row).toLowerCase().includes(q)
      })
    }
    for (const f of filters) {
      const val = filterValues[f.key]
      if (val) rows = rows.filter((row) => String((row as Record<string, unknown>)[f.key] ?? '') === val)
    }
    return rows
  }, [data, search, searchKeys, filters, filterValues])

  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize))
  const currentPage = Math.min(page, totalPages)
  const paged = filtered.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row gap-3 sm:items-center sm:justify-between">
        <div className="flex flex-wrap gap-2 flex-1">
          <div className="relative min-w-[200px] flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted" />
            <input
              type="search"
              placeholder={searchPlaceholder}
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1) }}
              className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border border-surface-border bg-surface text-foreground placeholder:text-muted focus:outline-none focus:border-accent/50"
            />
          </div>
          {filters.map((f) => (
            <select
              key={f.key}
              value={filterValues[f.key] ?? ''}
              onChange={(e) => { setFilterValues((v) => ({ ...v, [f.key]: e.target.value })); setPage(1) }}
              className="px-3 py-2 text-sm rounded-lg border border-surface-border bg-surface text-foreground"
            >
              <option value="">{f.label}</option>
              {f.options.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          ))}
        </div>
        {toolbar}
      </div>

      <div className="w-full max-h-[60vh] overflow-auto rounded-xl border border-surface-border">
        <table className="w-full text-sm border-collapse min-w-[600px]">
          <thead className="sticky top-0 z-10">
            <tr className="bg-surface-elevated/95 backdrop-blur">
              {columns.map((col) => (
                <th
                  key={col.key}
                  className="text-left px-4 py-3 font-semibold text-foreground border-b border-surface-border whitespace-nowrap font-mono text-xs uppercase tracking-wider"
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
                  <span className="inline-block w-5 h-5 border-2 border-accent border-t-transparent rounded-full animate-spin" />
                </td>
              </tr>
            ) : paged.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="text-center py-10 text-muted">
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              paged.map((row) => (
                <tr
                  key={keyExtractor(row)}
                  className="border-b border-surface-border hover:bg-surface-elevated/40 transition-colors"
                >
                  {columns.map((col) => (
                    <td key={col.key} className="px-4 py-3 text-muted">
                      {col.render ? col.render(row) : String((row as Record<string, unknown>)[col.key] ?? '')}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between text-sm text-muted">
        <span>
          {filtered.length} record{filtered.length !== 1 ? 's' : ''}
          {search || Object.values(filterValues).some(Boolean) ? ' (filtered)' : ''}
        </span>
        <div className="flex items-center gap-2">
          <button
            type="button"
            disabled={currentPage <= 1}
            onClick={() => setPage((p) => p - 1)}
            className={cn('p-1.5 rounded border border-surface-border', currentPage <= 1 && 'opacity-40')}
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <span className="font-mono text-xs">
            {currentPage} / {totalPages}
          </span>
          <button
            type="button"
            disabled={currentPage >= totalPages}
            onClick={() => setPage((p) => p + 1)}
            className={cn('p-1.5 rounded border border-surface-border', currentPage >= totalPages && 'opacity-40')}
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  )
}
