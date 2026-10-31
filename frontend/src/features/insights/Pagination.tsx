import { Link, useLocation, useSearchParams } from 'react-router-dom'
import { getPageNumbers } from './pageWindow'
import type { PageWindow } from './pageWindow'

type PaginationProps = {
  pagination: PageWindow
  label: string
}

export function Pagination({ pagination, label }: PaginationProps) {
  const { pathname } = useLocation()
  const [searchParams] = useSearchParams()
  if (pagination.totalItems === 0) return null

  function destination(page: number) {
    const params = new URLSearchParams(searchParams)
    if (page === 1) params.delete('page')
    else params.set('page', String(page))
    return { pathname, search: params.toString() ? `?${params}` : '' }
  }

  return (
    <nav className="pagination" aria-label={label}>
      <p className="pagination-summary">Page {pagination.page.toLocaleString()} of {pagination.totalPages.toLocaleString()}</p>
      {pagination.totalPages > 1 && (
        <div className="pagination-controls">
          {pagination.page === 1 ? (
            <span className="pagination-previous pagination-disabled" aria-disabled="true">← Previous</span>
          ) : (
            <Link className="pagination-previous" to={destination(pagination.page - 1)} aria-label="Previous page">← Previous</Link>
          )}

          <div className="pagination-pages">
            {getPageNumbers(pagination.page, pagination.totalPages).map((page, index) => (
              page === null ? <span className="pagination-gap" key={`gap-${index}`} aria-hidden="true">…</span> : (
                <Link
                  key={page}
                  to={destination(page)}
                  aria-label={`Page ${page}`}
                  aria-current={page === pagination.page ? 'page' : undefined}
                >{page}</Link>
              )
            ))}
          </div>

          {pagination.page === pagination.totalPages ? (
            <span className="pagination-next pagination-disabled" aria-disabled="true">Next →</span>
          ) : (
            <Link className="pagination-next" to={destination(pagination.page + 1)} aria-label="Next page">Next →</Link>
          )}
        </div>
      )}
    </nav>
  )
}
