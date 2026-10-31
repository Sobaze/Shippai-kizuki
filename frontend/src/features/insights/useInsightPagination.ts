import { useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'
import { getPageWindow } from './pageWindow'

export function useInsightPagination(totalItems: number, pageSize: number, ready: boolean, headingId: string) {
  const [searchParams, setSearchParams] = useSearchParams()
  const pagination = getPageWindow(searchParams.get('page'), totalItems, pageSize)
  const previousPage = useRef(pagination.page)

  useEffect(() => {
    if (!ready || !searchParams.has('page')) return
    const canonical = pagination.page === 1 ? null : String(pagination.page)
    if (searchParams.get('page') === canonical && searchParams.getAll('page').length === 1) return

    const params = new URLSearchParams(searchParams)
    if (canonical === null) params.delete('page')
    else params.set('page', canonical)
    setSearchParams(params, { replace: true })
  }, [ready, pagination.page, searchParams, setSearchParams])

  useEffect(() => {
    if (!ready) return
    const changed = previousPage.current !== pagination.page
    previousPage.current = pagination.page
    if (changed) {
      const heading = document.getElementById(headingId)
      heading?.focus({ preventScroll: true })
      heading?.scrollIntoView({ block: 'start' })
    }
  }, [ready, pagination.page, headingId])

  return pagination
}
