import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { PracticeFilters } from '../insights/PracticeFilters'
import { Pagination } from '../insights/Pagination'
import { useInsightPagination } from '../insights/useInsightPagination'
import { matchesPracticeFilter, parsePracticeFilter } from '../insights/practiceFilter'
import type { PracticeFilter } from '../insights/practiceFilter'
import { getKanjiStats } from './api'
import { KanjiGrid } from './KanjiGrid'
import type { KanjiStats } from './types'
import './kanji.css'

type KanjiState =
  | { status: 'loading' }
  | { status: 'success'; kanji: KanjiStats[] }
  | { status: 'error'; message: string }

export function KanjiPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const filter = parsePracticeFilter(searchParams.get('filter'))
  const [state, setState] = useState<KanjiState>({ status: 'loading' })
  const [reloadCount, setReloadCount] = useState(0)
  const kanji = state.status === 'success' ? state.kanji : []
  const filteredKanji = kanji.filter((entry) => matchesPracticeFilter(entry.recommendation, filter))
  const pagination = useInsightPagination(filteredKanji.length, 24, state.status === 'success', 'kanji-heading')

  useEffect(() => {
    const controller = new AbortController()

    getKanjiStats(controller.signal)
      .then((kanji) => {
        if (!controller.signal.aborted) {
          setState({ status: 'success', kanji })
        }
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) {
          setState({
            status: 'error',
            message: error instanceof Error ? error.message : 'Could not load kanji insights.',
          })
        }
      })

    return () => controller.abort()
  }, [reloadCount])

  function refreshKanji() {
    setState({ status: 'loading' })
    setReloadCount((count) => count + 1)
  }

  function changeFilter(next: PracticeFilter) {
    if (next === filter) return
    const params = new URLSearchParams(searchParams)
    if (next === 'all') params.delete('filter')
    else params.set('filter', next)
    params.delete('page')
    setSearchParams(params)
  }

  return (
    <section className="kanji-insights" aria-labelledby="kanji-heading">
      <div className="kanji-header">
        <div>
          <p className="kanji-eyebrow">Patterns across vocabulary</p>
          <h2 id="kanji-heading" tabIndex={-1}>Kanji worth a closer look</h2>
        </div>
        <button type="button" onClick={refreshKanji} disabled={state.status === 'loading'}>
          {state.status === 'loading' ? 'Loading…' : 'Refresh kanji'}
        </button>
      </div>

      <p className="kanji-note">
        Difficulty in words containing a kanji is a study clue, not proof that the kanji
        caused a failed review. Evidence comes from your saved vocabulary answer history.
      </p>

      {state.status === 'loading' && <p role="status">Looking for kanji patterns…</p>}
      {state.status === 'error' && <p className="kanji-error" role="alert">{state.message}</p>}
      {state.status === 'success' && (
        <PracticeFilters
          label="Filter kanji by practice suggestion"
          value={filter}
          recommendations={kanji.map((entry) => entry.recommendation)}
          onChange={changeFilter}
        />
      )}
      {state.status === 'success' && (
        kanji.length === 0 ? (
          <p className="insights-empty">No kanji evidence yet. <Link to="/">Import JP Vocabulary on Home</Link> to get started.</p>
        ) : (
          <>
            <p className="kanji-note" role="status">Showing {pagination.totalItems === 0 ? '0' : `${pagination.from.toLocaleString()}–${pagination.to.toLocaleString()}`} of {filteredKanji.length.toLocaleString()} matching kanji · {kanji.length.toLocaleString()} total</p>
            <Pagination pagination={pagination} label="Kanji pages, top" />
            {filteredKanji.length === 0 && <p className="insights-empty">No kanji match this filter. Choose All or another suggestion.</p>}
            {/* Result changes reset popup selection and expanded card state. */}
            <KanjiGrid
              key={`${filter}:${pagination.page}`}
              entries={filteredKanji.slice(pagination.startIndex, pagination.endIndex)}
            />
            <Pagination pagination={pagination} label="Kanji pages, bottom" />
          </>
        )
      )}
    </section>
  )
}
