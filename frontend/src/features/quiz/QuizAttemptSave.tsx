import { useEffect, useState } from 'react'
import { saveQuizAttempt } from './api'
import type { QuizAttemptSubmission, QuizResult, QuizRun, SavedQuizAttempt } from './types'

type SaveState =
  | { status: 'saving' }
  | { status: 'saved'; attempt: SavedQuizAttempt }
  | { status: 'error'; message: string }

export function QuizAttemptSave({ run, result }: { run: QuizRun; result: QuizResult }) {
  const [input] = useState<QuizAttemptSubmission>(() => ({
    ...run,
    completedAt: new Date().toISOString(),
    questions: result.items.map((item) => item.question),
    answers: result.items.flatMap((item) => item.selectedAnswer === null ? [] : [{
      questionId: item.question.id, selectedAnswer: item.selectedAnswer,
    }]),
  }))
  const [state, setState] = useState<SaveState>({ status: 'saving' })
  const [retryCount, setRetryCount] = useState(0)

  useEffect(() => {
    let active = true
    saveQuizAttempt(input)
      .then((attempt) => {
        if (active) setState({ status: 'saved', attempt })
      })
      .catch((error: unknown) => {
        if (active) setState({ status: 'error', message: error instanceof Error ? error.message : 'Could not save this quiz.' })
      })
    return () => { active = false }
  }, [input, retryCount])

  return (
    <div className="quiz-save-status">
      {state.status === 'saving' && <p role="status">Saving completed attempt…</p>}
      {state.status === 'saved' && <p className="quiz-save-success" role="status">Saved · {new Date(state.attempt.completedAt).toLocaleString()} · available in recent history</p>}
      {state.status === 'error' && (
        <>
          <p className="quiz-error" role="alert">{state.message}</p>
          <button type="button" className="quiz-secondary" onClick={() => {
            setState({ status: 'saving' })
            setRetryCount((count) => count + 1)
          }}>Retry saving</button>
        </>
      )}
    </div>
  )
}
