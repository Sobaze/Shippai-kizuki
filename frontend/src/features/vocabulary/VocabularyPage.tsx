import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { PracticeFilters } from '../insights/PracticeFilters'
import { Pagination } from '../insights/Pagination'
import { useInsightPagination } from '../insights/useInsightPagination'
import { matchesPracticeFilter, parsePracticeFilter } from '../insights/practiceFilter'
import type { PracticeFilter } from '../insights/practiceFilter'
import { getVocabularyStats } from './api'
import type { PracticeRecommendation, VocabularyStats } from './types'
import './vocabulary.css'

type VocabularyState =
  | { status: 'loading' }
  | { status: 'success'; words: VocabularyStats[] }
  | { status: 'error'; message: string }

const recommendationLabels: Record<PracticeRecommendation, string> = {
  limited_evidence: 'Limited evidence',
  suggested_practice: 'Suggested practice',
  no_extra_practice: 'No extra practice suggested',
}

export function VocabularyPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const filter = parsePracticeFilter(searchParams.get('filter'))
  const [state, setState] = useState<VocabularyState>({ status: 'loading' })
  const [reloadCount, setReloadCount] = useState(0)
  const words = state.status === 'success' ? state.words : []
  const filteredWords = words.filter((word) => matchesPracticeFilter(word.recommendation, filter))
  const pagination = useInsightPagination(filteredWords.length, 50, state.status === 'success', 'vocabulary-heading')

  useEffect(() => {
    const controller = new AbortController()

    getVocabularyStats(controller.signal)
      .then((words) => {
        if (!controller.signal.aborted) {
          setState({ status: 'success', words })
        }
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) {
          setState({
            status: 'error',
            message: error instanceof Error ? error.message : 'Could not load vocabulary scores.',
          })
        }
      })

    return () => controller.abort()
  }, [reloadCount])

  function refreshScores() {
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
    <section className="vocabulary" aria-labelledby="vocabulary-heading">
      <div className="vocabulary-header">
        <div>
          <h2 id="vocabulary-heading" tabIndex={-1}>What could use more practice?</h2>
          <p>Vocabulary difficulty · Highest scores first</p>
        </div>
        <button type="button" onClick={refreshScores} disabled={state.status === 'loading'}>
          {state.status === 'loading' ? 'Loading…' : 'Refresh scores'}
        </button>
      </div>

      <p className="vocabulary-note">
        Based on saved answer history. Higher scores mean more recorded difficulty,
        not a measurement of mastery.
      </p>

      {state.status === 'loading' && <p role="status">Calculating vocabulary scores…</p>}
      {state.status === 'error' && <p className="vocabulary-error" role="alert">{state.message}</p>}
      {state.status === 'success' && (
        <PracticeFilters
          label="Filter vocabulary by practice suggestion"
          value={filter}
          recommendations={words.map((word) => word.recommendation)}
          onChange={changeFilter}
        />
      )}
      {state.status === 'success' && (
        words.length === 0 ? (
          <p className="insights-empty">No saved answers yet. <Link to="/">Import JP Vocabulary on Home</Link> to get started.</p>
        ) : filteredWords.length === 0 ? (
          <p className="insights-empty" role="status">No vocabulary matches this filter. Choose All or another suggestion.</p>
        ) : (
          <>
            <Pagination pagination={pagination} label="Vocabulary pages, top" />
            <div className="vocabulary-table-container">
              <table className="vocabulary-table">
                <caption>
                  Showing {pagination.from.toLocaleString()}–{pagination.to.toLocaleString()} of {filteredWords.length.toLocaleString()} matching words
                  {' · '}{filteredWords.reduce((sum, word) => sum + word.reviewCount, 0).toLocaleString()} answers across matching words
                </caption>
                <thead>
                  <tr>
                    <th scope="col">Vocabulary</th>
                    <th scope="col">Difficulty</th>
                    <th scope="col">Evidence</th>
                    <th scope="col">Suggestion</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredWords.slice(pagination.startIndex, pagination.endIndex).map((word) => (
                    <tr key={word.vocabulary}>
                      <td lang="ja">{word.vocabulary}</td>
                      <td className="difficulty-score">{word.difficulty}<span> / 100</span></td>
                      <td>
                        <strong>{word.reviewCount} reviews</strong>
                        <small>
                          Again {word.ratingCounts.again} · Hard {word.ratingCounts.hard}
                          {' · '}Good {word.ratingCounts.good} · Easy {word.ratingCounts.easy}
                        </small>
                      </td>
                      <td>
                        <span className={`practice-label practice-${word.recommendation}`}>
                          {recommendationLabels[word.recommendation]}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <Pagination pagination={pagination} label="Vocabulary pages, bottom" />
          </>
        )
      )}
    </section>
  )
}
