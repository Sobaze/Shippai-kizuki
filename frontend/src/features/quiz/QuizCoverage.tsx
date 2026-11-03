import { useEffect, useState } from 'react'
import { getQuizCoverage } from './api'
import type { QuizCoverage as Coverage, QuizLevel, QuizMode } from './types'

type CoverageState =
  | { status: 'loading' }
  | { status: 'ready'; coverage: Coverage }
  | { status: 'error'; message: string }

export function QuizCoverage({ mode, level, exerciseId }: { mode: QuizMode; level: QuizLevel; exerciseId?: string }) {
  const [state, setState] = useState<CoverageState>({ status: 'loading' })
  const [refreshCount, setRefreshCount] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    getQuizCoverage(mode, level, controller.signal, exerciseId)
      .then((coverage) => {
        if (!controller.signal.aborted) setState({ status: 'ready', coverage })
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) setState({ status: 'error', message: error instanceof Error ? error.message : 'Could not load coverage.' })
      })
    return () => controller.abort()
  }, [mode, level, exerciseId, refreshCount])

  return (
    <section className="quiz-coverage" aria-labelledby="quiz-coverage-heading">
      <div className="quiz-history-header">
        <h3 id="quiz-coverage-heading">Your {level} {mode === 'sentence' ? 'exercise' : mode === 'reading' ? 'word-reading' : 'meaning'} coverage</h3>
        <button type="button" className="quiz-secondary" disabled={state.status === 'loading'} onClick={() => {
          setState({ status: 'loading' })
          setRefreshCount((count) => count + 1)
        }}>Refresh coverage</button>
      </div>
      {state.status === 'loading' && <p role="status">Loading saved coverage…</p>}
      {state.status === 'error' && <p className="quiz-error" role="alert">{state.message}</p>}
      {state.status === 'ready' && (
        <>
          <p className="quiz-coverage-count"><strong>{state.coverage.attempted} / {state.coverage.total}</strong> attempted · {state.coverage.unseen} not yet attempted</p>
          <progress className="quiz-progress" value={state.coverage.attempted} max={state.coverage.total} aria-label={`${level} ${mode} pool items attempted`} />
        </>
      )}
      <p className="quiz-note">Counts distinct items in this current pool from all saved completed quizzes and retries, including older matching IDs. Coverage measures practice exposure, not correctness or mastery. Cancelled and unsaved runs do not count.</p>
    </section>
  )
}
