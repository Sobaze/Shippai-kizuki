import { useEffect, useState } from 'react'
import { getQuizPool } from './api'
import type { QuizLevel, QuizMode, QuizPool } from './types'

type PoolState =
  | { key: string; status: 'ready'; pool: QuizPool }
  | { key: string; status: 'error'; message: string }

export function useQuizPool(mode: QuizMode, level: QuizLevel, exerciseId?: string) {
  const [state, setState] = useState<PoolState | null>(null)
  const [refreshCount, setRefreshCount] = useState(0)
  const key = `${mode}/${level}/${exerciseId ?? ''}/${refreshCount}`

  useEffect(() => {
    const controller = new AbortController()
    getQuizPool(mode, level, controller.signal, exerciseId)
      .then((pool) => {
        if (!controller.signal.aborted) setState({ key, status: 'ready', pool })
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) setState({ key, status: 'error', message: error instanceof Error ? error.message : 'Could not load quiz questions.' })
      })
    return () => controller.abort()
  }, [mode, level, exerciseId, key])

  // Ignore an earlier pool immediately on selection/refresh, even before the
  // effect starts. Aborted or stale responses cannot start the wrong level.
  const current = state?.key === key ? state : null
  return { state: current, reload: () => setRefreshCount((count) => count + 1) }
}
