import type { QuizAttemptSubmission, QuizAttemptSummary, QuizCandidate, QuizCoverage, QuizLevel, QuizMode, QuizPool, SavedQuizAttempt } from './types'

const pendingSaves = new Map<string, Promise<SavedQuizAttempt>>()

export async function getQuizCoverage(mode: QuizMode, level: QuizLevel, signal: AbortSignal, exerciseId?: string): Promise<QuizCoverage> {
  // A just-completed run may still be saving after its results view closes.
  await Promise.allSettled([...pendingSaves.values()])
  signal.throwIfAborted()
  const response = await fetch(`/api/quiz/coverage?mode=${mode}&level=${level}${exerciseId ? `&exercise=${encodeURIComponent(exerciseId)}` : ''}`, { signal, cache: 'no-store' })
  if (!response.ok) {
    if (response.status === 404) throw new Error('Quiz coverage is unavailable. Restart the Go backend to load the coverage route, then retry.')
    throw new Error((await response.text()).trim() || 'Could not load saved quiz coverage.')
  }
  const coverage: QuizCoverage = await response.json()
  if (coverage.mode !== mode || coverage.level !== level || coverage.exerciseId !== exerciseId || !Array.isArray(coverage.questions)
    || !Number.isInteger(coverage.total) || coverage.total < (mode === 'sentence' ? 1 : 10) || coverage.questions.length !== coverage.total
    || !Number.isInteger(coverage.attempted) || coverage.attempted < 0 || coverage.attempted > coverage.total
    || coverage.unseen !== coverage.total - coverage.attempted
    || new Set(coverage.questions.map((entry) => entry.questionId)).size !== coverage.total
    || coverage.questions.some((entry) => !Number.isInteger(entry.attempts) || entry.attempts < 0
      || (entry.attempts === 0 ? entry.lastPracticedAt !== null : !entry.lastPracticedAt || !Number.isFinite(Date.parse(entry.lastPracticedAt))))) {
    throw new Error('The backend returned incomplete or mismatched quiz coverage.')
  }
  return coverage
}

export async function getQuizPool(mode: QuizMode, level: QuizLevel, signal: AbortSignal, exerciseId?: string): Promise<QuizPool> {
  const response = await fetch(`/api/quiz/questions?mode=${mode}&level=${level}${exerciseId ? `&exercise=${encodeURIComponent(exerciseId)}` : ''}`, { signal, cache: 'no-store' })
  if (!response.ok) {
    if (response.status === 404) throw new Error('Quiz questions are unavailable. Restart the Go backend to load the new question-pool route, then retry.')
    throw new Error((await response.text()).trim() || 'Could not load quiz questions.')
  }
  const pool: QuizPool = await response.json()
  if (pool.mode !== mode || pool.level !== level || pool.exerciseId !== exerciseId || !Array.isArray(pool.questions) || pool.questions.length < (mode === 'sentence' ? 1 : 10)) {
    throw new Error('The backend returned an incomplete or mismatched question pool.')
  }
  return pool
}

export function saveQuizAttempt(input: QuizAttemptSubmission): Promise<SavedQuizAttempt> {
  const pending = pendingSaves.get(input.id)
  if (pending) return pending

  // Finish a completed-attempt write even if its result view closes. The stable
  // ID also makes an explicit retry safe if the first response was lost.
  const request = postAttempt(input).finally(() => pendingSaves.delete(input.id))
  pendingSaves.set(input.id, request)
  return request
}

async function postAttempt(input: QuizAttemptSubmission): Promise<SavedQuizAttempt> {
  const response = await fetch('/api/quiz/attempts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!response.ok) {
    throw new Error((await response.text()).trim() || 'Could not save this attempt. Check that the Go backend is running.')
  }
  return response.json()
}

export async function getQuizHistory(mode: QuizMode, signal: AbortSignal): Promise<QuizAttemptSummary[]> {
  // If the result view just closed, let its save finish before reading history.
  await Promise.allSettled([...pendingSaves.values()])
  signal.throwIfAborted()
  const response = await fetch(`/api/quiz/attempts?mode=${mode}`, { signal, cache: 'no-store' })
  if (!response.ok) throw new Error((await response.text()).trim() || 'Could not load quiz history.')
  const result: { attempts: QuizAttemptSummary[] } = await response.json()
  return result.attempts
}

export async function getQuizAttempt(id: string, signal: AbortSignal): Promise<SavedQuizAttempt> {
  const response = await fetch(`/api/quiz/attempts/${encodeURIComponent(id)}`, { signal, cache: 'no-store' })
  if (!response.ok) throw new Error((await response.text()).trim() || 'Could not load this saved attempt.')
  return response.json()
}

export async function getQuizCandidates(mode: QuizMode, signal: AbortSignal): Promise<QuizCandidate[]> {
  await Promise.allSettled([...pendingSaves.values()])
  signal.throwIfAborted()
  const response = await fetch(`/api/quiz/practice?mode=${mode}`, { signal, cache: 'no-store' })
  if (!response.ok) throw new Error((await response.text()).trim() || 'Could not load practice candidates.')
  const result: { candidates: QuizCandidate[] } = await response.json()
  return result.candidates
}
