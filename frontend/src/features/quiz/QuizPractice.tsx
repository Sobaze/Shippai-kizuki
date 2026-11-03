import { useEffect, useState } from 'react'
import { getQuizCandidates } from './api'
import { candidatesForLevel, exerciseTitle, QUIZ_LENGTH } from './quiz'
import type { QuizCandidate, QuizLevel, QuizMode, QuizQuestion } from './types'

type PracticeState =
  | { status: 'loading' }
  | { status: 'ready'; candidates: QuizCandidate[] }
  | { status: 'error'; message: string }

export function QuizPractice({ mode, level, poolQuestions, onStart }: { mode: QuizMode; level: QuizLevel; poolQuestions: QuizQuestion[]; onStart: (candidates: QuizCandidate[]) => void }) {
  const [state, setState] = useState<PracticeState>({ status: 'loading' })
  const [refreshCount, setRefreshCount] = useState(0)
  const [showAllLevels, setShowAllLevels] = useState(false)
  const [visibleCount, setVisibleCount] = useState(24)
  const candidates = state.status === 'ready' ? (showAllLevels ? state.candidates : candidatesForLevel(state.candidates, level, poolQuestions)) : []

  useEffect(() => {
    const controller = new AbortController()
    getQuizCandidates(mode, controller.signal)
      .then((candidates) => {
        if (!controller.signal.aborted) setState({ status: 'ready', candidates })
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) setState({ status: 'error', message: error instanceof Error ? error.message : 'Could not load practice candidates.' })
      })
    return () => controller.abort()
  }, [mode, refreshCount])

  return (
    <section className="quiz-card quiz-practice" aria-labelledby="quiz-practice-heading">
      <div className="quiz-history-header">
        <h2 id="quiz-practice-heading">Your {showAllLevels ? 'all-level' : level} {mode === 'sentence' ? 'JLPT Kanji exercise' : mode === 'reading' ? 'word-reading' : 'meaning'} practice queue</h2>
        <button type="button" className="quiz-secondary" disabled={state.status === 'loading'} onClick={() => {
          setState({ status: 'loading' })
          setVisibleCount(24)
          setRefreshCount((count) => count + 1)
        }}>Refresh queue</button>
      </div>
      <p>A wrong answer adds an item at 0/2. Two consecutive correct answers in separate saved attempts clear it for now. A later mistake brings it back.</p>
      <label className="quiz-note"><input type="checkbox" checked={showAllLevels} onChange={(event) => { setShowAllLevels(event.target.checked); setVisibleCount(24) }} /> Include other levels and older saved items</label>
      <p className="quiz-note">{showAllLevels ? 'Showing this mode’s entire saved queue.' : `Focused on estimated ${level}; matching legacy IDs also count.`} Full quizzes and retries both count. Unasked items and cancelled runs leave progress unchanged. Queue sessions are saved as focused retries.</p>
      {state.status === 'loading' && <p role="status">Loading practice queue…</p>}
      {state.status === 'error' && <p className="quiz-error" role="alert">{state.message}</p>}
      {state.status === 'ready' && (candidates.length === 0 ? (
        <p className="quiz-note">No items waiting in this {showAllLevels ? 'mode' : 'level'}. Complete a quiz to discover what needs practice.</p>
      ) : (
        <>
          <p>{candidates.length} items waiting. Practise up to {QUIZ_LENGTH}, least recently practised first; question order and choices are shuffled.</p>
          <button type="button" className="quiz-primary" onClick={() => onStart(candidates)}>
            Practise {Math.min(QUIZ_LENGTH, candidates.length)} queued {candidates.length === 1 ? 'item' : 'items'}
          </button>
          <ul className="quiz-candidate-list">
            {candidates.slice(0, visibleCount).map((candidate) => (
              <li key={candidate.question.id}>
                <strong lang="ja">{candidate.question.prompt}</strong>
                {candidate.question.task && <small>{exerciseTitle(candidate.question.exerciseId)} · {candidate.question.task === 'reading' ? 'Reading' : 'Spelling'}</small>}
                {candidate.question.estimatedLevel && <small>Estimated {candidate.question.estimatedLevel}</small>}
                <span>{candidate.streak}/2 consecutive correct</span>
                <small>Last practised {new Date(candidate.lastPracticedAt).toLocaleString()}</small>
              </li>
            ))}
          </ul>
          {candidates.length > visibleCount && <button type="button" className="quiz-secondary" onClick={() => setVisibleCount((count) => count + 24)}>Show 24 more ({visibleCount} of {candidates.length} shown)</button>}
        </>
      ))}
    </section>
  )
}
