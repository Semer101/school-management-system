import { useMemo, useState, type ReactNode } from 'react'
import {
  ChevronLeft, ChevronRight, Search, X, Inbox,
} from 'lucide-react'
import { cn } from '../../lib/utils'

export interface Column<T> {
  key: string
  header: string
  render?: (row: T) => ReactNode
  filterable?: boolean
}

/**
 * A single filter definition.
 * - `key` is the identifier used internally to track the selected option.
 * - `accessor` lets you pull a value from a (potentially nested) row object so
 *   filters can target derived/joined columns such as `user.role`.
 */
export interface FilterDef<T> {
  key: string
  label: string
  options: { value: string; label: string }[]
  accessor?: (row: T) => string
}

export interface EmptyState {
  icon?: ReactNode
  title?: string
  description?: string
}

interface DataTableProps<T> {
  columns: Column<T>[]
  data: T[]
  keyExtractor: (row: T) => string | number
  loading?: boolean

  /** Plain message shown when the table is empty (kept for back-compat). */
  emptyMessage?: string
  /** Richer empty-state UI: icon + title + description. Overrides `emptyMessage`. */
  emptyState?: EmptyState

  searchPlaceholder?: string
  /**
   * Dotted-path keys to include in the search index, e.g. `['name', 'user.email']`.
   * Use `searchAccessor` instead when the row has nested/derived data.
   */
  searchKeys?: (keyof T | string)[]
  /**
   * Custom function returning a single lower-cased string per row to match
   * against the user's query. Receives the whole row. Takes precedence over
   * `searchKeys` when provided.
   */
  searchAccessor?: (row: T) => string

  filters?: FilterDef<T>[]

  pageSize?: number
  toolbar?: ReactNode
}

export function DataTable<T extends object>({
  columns,
  data,
  keyExtractor,
  loading,
  emptyMessage = 'No data found.',
  emptyState,
  searchPlaceholder = 'Search...',
  searchKeys,
  searchAccessor,
  filters = [],
  pageSize = 10,
  toolbar,
}: DataTableProps<T>) {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [filterValues, setFilterValues] = useState<Record<string, string>>({})

  /**
   * Resolve a dotted path like `user.email` against an arbitrary object.
   * Returns '' if any segment is missing so the call always yields a string.
   */
  const resolvePath = (row: T, path: string): string => {
    const segments = path.split('.')
    let cur: unknown = row
    for (const seg of segments) {
      if (cur && typeof cur === 'object' && seg in (cur as Record<string, unknown>)) {
        cur = (cur as Record<string, unknown>)[seg]
      } else {
        return ''
      }
    }
    return cur == null ? '' : String(cur)
  }

  const getSearchable = (row: T): string => {
    if (searchAccessor) return searchAccessor(row)
    if (searchKeys?.length) {
      return searchKeys.map((k) => resolvePath(row, String(k))).join(' \u0001 ')
    }
    // Fallback: stringify the whole row. Cheap, but at least finds _something_.
    try {
      return JSON.stringify(row)
    } catch {
      return ''
    }
  }

  const filtered = useMemo(() => {
    let rows = [...data]
    const q = search.trim().toLowerCase()
    if (q) {
      rows = rows.filter((row) => getSearchable(row).toLowerCase().includes(q))
    }
    for (const f of filters) {
      const val = filterValues[f.key]
      if (!val) continue
      rows = rows.filter((row) => {
        const v = f.accessor ? f.accessor(row) : resolvePath(row, f.key)
        return v === val
      })
    }
    return rows
  }, [data, search, searchKeys, searchAccessor, filters, filterValues]) // eslint-disable-line react-hooks/exhaustive-deps

  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize))
  const currentPage = Math.min(page, totalPages)
  const paged = filtered.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  const isFiltered = search.trim().length > 0
    || Object.values(filterValues).some((v) => Boolean(v))

  const clearAll = () => {
    setSearch('')
    setFilterValues({})
    setPage(1)
  }

  return (
    <div className="space-y-4">
      {/* Toolbar: search + filters + custom slot */}
      <div className="flex flex-col sm:flex-row gap-3 sm:items-center sm:justify-between">
        <div className="flex flex-wrap gap-2 flex-1">
          <div className="relative min-w-[200px] flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted pointer-events-none" />
            <input
              type="search"
              placeholder={searchPlaceholder}
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1) }}
              className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border border-surface-border bg-surface text-foreground placeholder:text-muted focus:outline-none focus:border-accent/50 focus:ring-1 focus:ring-accent/30 transition-colors"
            />
          </div>

          {filters.map((f) => (
            <select
              key={f.key}
              value={filterValues[f.key] ?? ''}
              onChange={(e) => {
                setFilterValues((v) => ({ ...v, [f.key]: e.target.value }))
                setPage(1)
              }}
              className="px-3 py-2 text-sm rounded-lg border border-surface-border bg-surface text-foreground focus:outline-none focus:border-accent/50 focus:ring-1 focus:ring-accent/30 transition-colors"
            >
              <option value="">{f.label}</option>
              {f.options.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          ))}

          {isFiltered && (
            <button
              type="button"
              onClick={clearAll}
              className="inline-flex items-center gap-1 px-3 py-2 text-xs font-semibold text-muted hover:text-foreground border border-surface-border rounded-lg transition-colors"
            >
              <X className="w-3.5 h-3.5" /> Clear
            </button>
          )}
        </div>
        {toolbar}
      </div>

      {/* Active filter chips (cosmetic) */}
      {isFiltered && (
        <div className="flex flex-wrap gap-1.5 -mt-2">
          {search.trim() && (
            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-accent/10 border border-accent/30 text-accent text-[11px] font-mono">
              search: "{search.trim()}"
              <button
                type="button"
                aria-label="Clear search"
                onClick={() => { setSearch(''); setPage(1) }}
                className="hover:text-accent-hover"
              >
                <X className="w-3 h-3" />
              </button>
            </span>
          )}
          {filters.filter((f) => filterValues[f.key]).map((f) => {
            const opt = f.options.find((o) => o.value === filterValues[f.key])
            return (
              <span
                key={f.key}
                className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-accent/10 border border-accent/30 text-accent text-[11px] font-mono"
              >
                {f.label}: {opt?.label ?? filterValues[f.key]}
                <button
                  type="button"
                  aria-label={`Clear ${f.label}`}
                  onClick={() => {
                    setFilterValues((v) => ({ ...v, [f.key]: '' }))
                    setPage(1)
                  }}
                  className="hover:text-accent-hover"
                >
                  <X className="w-3 h-3" />
                </button>
              </span>
            )
          })}
        </div>
      )}

      {/* Table */}
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
                <td colSpan={columns.length} className="py-10">
                  {emptyState ? (
                    <div className="flex flex-col items-center justify-center text-center px-6 py-4 gap-2">
                      <div className="w-12 h-12 rounded-full bg-surface-elevated border border-surface-border flex items-center justify-center text-muted">
                        {emptyState.icon ?? <Inbox className="w-6 h-6" />}
                      </div>
                      <p className="text-sm font-semibold text-foreground">
                        {emptyState.title ?? 'Nothing to show'}
                      </p>
                      {emptyState.description && (
                        <p className="text-xs text-muted max-w-sm">{emptyState.description}</p>
                      )}
                      {isFiltered && (
                        <button
                          type="button"
                          onClick={clearAll}
                          className="mt-1 text-xs font-semibold text-accent hover:underline"
                        >
                          Clear filters
                        </button>
                      )}
                    </div>
                  ) : (
                    <div className="text-center text-muted">{emptyMessage}</div>
                  )}
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

      {/* Pagination + record count */}
      <div className="flex flex-col sm:flex-row gap-2 sm:items-center sm:justify-between text-sm text-muted">
        <span>
          {filtered.length} record{filtered.length !== 1 ? 's' : ''}
          {isFiltered ? ' (filtered)' : ''}
          {data.length !== filtered.length && !isFiltered ? ` of ${data.length}` : ''}
        </span>
        <div className="flex items-center gap-2">
          <button
            type="button"
            disabled={currentPage <= 1}
            onClick={() => setPage((p) => p - 1)}
            className={cn(
              'p-1.5 rounded border border-surface-border transition-colors',
              currentPage <= 1
                ? 'opacity-40 cursor-not-allowed'
                : 'hover:border-accent/40 hover:text-accent'
            )}
            aria-label="Previous page"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <span className="font-mono text-xs px-2">
            Page {currentPage} of {totalPages}
          </span>
          <button
            type="button"
            disabled={currentPage >= totalPages}
            onClick={() => setPage((p) => p + 1)}
            className={cn(
              'p-1.5 rounded border border-surface-border transition-colors',
              currentPage >= totalPages
                ? 'opacity-40 cursor-not-allowed'
                : 'hover:border-accent/40 hover:text-accent'
            )}
            aria-label="Next page"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  )
}
