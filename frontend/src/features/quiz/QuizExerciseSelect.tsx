import { useEffect, useState } from 'react'
import { getQuizExercises } from './api'
import type { ExerciseCategory, ExerciseSummary, QuizLevel } from './types'

type ListState =
  | { status: 'loading' }
  | { status: 'ready'; exercises: ExerciseSummary[] }
  | { status: 'error'; message: string }

export function QuizExerciseSelect({ level, category, exerciseId, disabled, onChange }: {
  level: QuizLevel; category: ExerciseCategory; exerciseId: string; disabled: boolean; onChange: (id: string) => void
}) {
  const [state, setState] = useState<ListState>({ status: 'loading' })
  const [refreshCount, setRefreshCount] = useState(0)
  useEffect(() => {
    const controller = new AbortController()
    getQuizExercises(level, controller.signal, category)
      .then(exercises => { if (!controller.signal.aborted) setState({ status: 'ready', exercises }) })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) setState({ status: 'error', message: error instanceof Error ? error.message : 'Could not load exercises.' })
      })
    return () => controller.abort()
  }, [level, category, refreshCount])

  return (
    <div className="quiz-exercise-select">
      <label htmlFor="quiz-exercise">Choose exercise</label>
      <select id="quiz-exercise" value={exerciseId} disabled={disabled || state.status !== 'ready'} onChange={event => onChange(event.target.value)}>
        {state.status !== 'ready' ? <option value={exerciseId}>Loading exercises…</option> : (
          <>
            {!state.exercises.some(exercise => exercise.id === exerciseId) && <option value={exerciseId}>Selected exercise unavailable</option>}
            {state.exercises.some(exercise => exercise.original) && <optgroup label="Original N2 practice">{state.exercises.filter(exercise => exercise.original).map(exercise => <option key={exercise.id} value={exercise.id}>{exercise.title} — {exercise.questionCount} questions</option>)}</optgroup>}
            {state.exercises.some(exercise => !exercise.original) && <optgroup label="Supplied exercise sets">{state.exercises.filter(exercise => !exercise.original).map(exercise => <option key={exercise.id} value={exercise.id}>{exercise.title} — {exercise.questionCount} questions</option>)}</optgroup>}
          </>
        )}
      </select>
      {state.status === 'loading' && <p className="quiz-note" role="status">Loading exercises…</p>}
      {state.status === 'error' && <p className="quiz-error" role="alert">{state.message}</p>}
      {state.status === 'ready' && state.exercises.length === 0 && <p className="quiz-note">No exercises are available at this level.</p>}
      {state.status === 'error' && <button type="button" className="quiz-secondary" disabled={disabled} onClick={() => { setState({ status: 'loading' }); setRefreshCount(count => count + 1) }}>Retry loading exercises</button>}
    </div>
  )
}
