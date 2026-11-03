import { useEffect, useRef, useState } from 'react'
import { getQuizAttempt, getQuizHistory } from './api'
import { exerciseTitle } from './quiz'
import type { QuizAttemptSummary, QuizMode, SavedQuizAttempt } from './types'

type HistoryState =
  | { status: 'loading' }
  | { status: 'ready'; attempts: QuizAttemptSummary[] }
  | { status: 'error'; message: string }

export function QuizHistory({ mode, onReview }: { mode: QuizMode; onReview: (attempt: SavedQuizAttempt) => void }) {
  const [state, setState] = useState<HistoryState>({ status: 'loading' })
  const [refreshCount, setRefreshCount] = useState(0)
  const [openingId, setOpeningId] = useState<string | null>(null)
  const [detailError, setDetailError] = useState<string | null>(null)
  const requestRef = useRef<AbortController | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    requestRef.current = controller
    getQuizHistory(mode, controller.signal)
      .then((attempts) => {
        if (!controller.signal.aborted) setState({ status: 'ready', attempts })
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) setState({ status: 'error', message: error instanceof Error ? error.message : 'Could not load history.' })
      })
    return () => requestRef.current?.abort()
  }, [mode, refreshCount])

  async function reviewAttempt(id: string) {
    requestRef.current?.abort()
    const controller = new AbortController()
    requestRef.current = controller
    setOpeningId(id)
    setDetailError(null)
    try {
      const attempt = await getQuizAttempt(id, controller.signal)
      if (!controller.signal.aborted) onReview(attempt)
    } catch (error: unknown) {
      if (!controller.signal.aborted) setDetailError(error instanceof Error ? error.message : 'Could not load the attempt.')
    } finally {
      if (!controller.signal.aborted) setOpeningId(null)
    }
  }

  return (
    <section className="quiz-card quiz-history" aria-labelledby="quiz-history-heading">
      <div className="quiz-history-header">
        <div>
          <h2 id="quiz-history-heading">Recent {mode === 'sentence' ? 'JLPT Kanji exercise' : mode === 'reading' ? 'word-reading' : 'meaning'} attempts</h2>
          <p className="quiz-note">Latest 20 completed attempts in this mode. Full quizzes and focused retries are labelled separately.</p>
        </div>
        <button type="button" className="quiz-secondary" disabled={state.status === 'loading' || openingId !== null} onClick={() => {
          setState({ status: 'loading' })
          setDetailError(null)
          setRefreshCount((count) => count + 1)
        }}>Refresh history</button>
      </div>
      {state.status === 'loading' && <p role="status">Loading saved quiz history…</p>}
      {state.status === 'error' && <p className="quiz-error" role="alert">{state.message} You can still practise above.</p>}
      {detailError && <p className="quiz-error" role="alert">{detailError}</p>}
      {state.status === 'ready' && (
        state.attempts.length === 0 ? <p className="quiz-note">No completed attempts saved for this mode yet.</p> : (
          <div className="quiz-history-scroll">
            <table>
              <thead><tr><th scope="col">Completed</th><th scope="col">Practice</th><th scope="col">Score</th><th scope="col">Review</th></tr></thead>
              <tbody>
                {state.attempts.map((attempt) => (
                  <tr key={attempt.id}>
                    <td>{new Date(attempt.completedAt).toLocaleString()}</td>
                    <td>{mode === 'sentence' ? `${exerciseTitle(attempt.exerciseId)} · ` : ''}{attempt.kind === 'retry' ? 'Focused retry' : mode === 'sentence' ? 'Full exercise' : 'Full quiz'}</td>
                    <td>{attempt.correctCount} / {attempt.total}</td>
                    <td><button type="button" className="quiz-secondary" disabled={openingId !== null} onClick={() => reviewAttempt(attempt.id)} aria-label={`Review attempt from ${new Date(attempt.completedAt).toLocaleString()}`}>
                      {openingId === attempt.id ? 'Loading…' : 'Review'}
                    </button></td>
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
