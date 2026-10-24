import { useEffect, useState } from 'react'
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
  const [state, setState] = useState<VocabularyState>({ status: 'loading' })
  const [reloadCount, setReloadCount] = useState(0)

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

  return (
    <section className="vocabulary" aria-labelledby="vocabulary-heading">
      <div className="vocabulary-header">
        <div>
          <h2 id="vocabulary-heading">What could use more practice?</h2>
          <p>Vocabulary difficulty · Highest scores first</p>
        </div>
        <button type="button" onClick={refreshScores} disabled={state.status === 'loading'}>
          {state.status === 'loading' ? 'Loading…' : 'Refresh scores'}
        </button>
      </div>

      <p className="vocabulary-note">
        Demo histories, not Anki data. Higher scores mean more recorded difficulty,
        not a measurement of mastery.
      </p>

      <details className="scoring-explanation">
        <summary>How is the score calculated?</summary>
        <p>
          Again = 1, Hard = 0.5, Good = 0, Easy = 0. Add the weights, divide by
          the review count, multiply by 100, then round to a whole number.
        </p>
        <p>
          At least 3 reviews and a score of 50 or more suggests extra practice.
          Fewer than 3 reviews means limited evidence, regardless of the score.
          These are experimental rules; all reviews currently have equal influence,
          regardless of age.
        </p>
      </details>

      {state.status === 'loading' && <p role="status">Calculating vocabulary scores…</p>}
      {state.status === 'error' && <p className="vocabulary-error" role="alert">{state.message}</p>}
      {state.status === 'success' && (
        state.words.length === 0 ? (
          <p>No reviews yet. Record a review to see vocabulary scores.</p>
        ) : (
          <div className="vocabulary-table-container">
            <table className="vocabulary-table">
              <caption>
                {state.words.length} words · {state.words.reduce((sum, word) => sum + word.reviewCount, 0)} reviews
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
                {state.words.map((word) => (
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
        )
      )}
    </section>
  )
}
