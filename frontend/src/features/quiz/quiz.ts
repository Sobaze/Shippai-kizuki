import type { QuizAnswer, QuizCandidate, QuizCoverage, QuizLevel, QuizQuestion, QuizResult, QuizState } from './types'

export const QUIZ_LENGTH = 10

export function exerciseTitle(id?: string): string {
  return id ? id.replace(/^exercise-/, 'Exercise ') : 'Mixed exercises'
}

function shuffled<T>(values: readonly T[], random: () => number): T[] {
  const result = [...values]
  for (let index = result.length - 1; index > 0; index--) {
    const other = Math.floor(random() * (index + 1))
    const value = result[index]
    result[index] = result[other]
    result[other] = value
  }
  return result
}

export function createQuiz(pool: readonly QuizQuestion[], random: () => number = Math.random): QuizQuestion[] {
  if (pool.length < QUIZ_LENGTH) throw new Error('This pool needs at least 10 checked questions.')
  return shuffled(pool, random).slice(0, QUIZ_LENGTH).map((question) => ({
    ...question,
    choices: shuffled(question.choices, random),
  }))
}

export function createExerciseQuiz(pool: readonly QuizQuestion[], random: () => number = Math.random): QuizQuestion[] {
  if (pool.length === 0 || pool.some(question => question.mode !== 'sentence' || !question.context || !question.task)) throw new Error('This exercise is incomplete.')
  return pool.map(question => ({ ...question, choices: shuffled(question.choices, random) }))
}

export function createCoverageQuiz(pool: readonly QuizQuestion[], coverage: QuizCoverage, random: () => number = Math.random): QuizQuestion[] {
  const counts = new Map(coverage.questions.map((entry) => [entry.questionId, entry]))
  if (coverage.total !== pool.length || pool.some((question) => question.mode !== coverage.mode
    || question.estimatedLevel !== coverage.level || !counts.has(question.id))) {
    throw new Error('Coverage does not match this question pool. Reload questions and try again.')
  }
  // Stable sort keeps the initial shuffle for exact ties, instead of always
  // favouring the first authored questions. Rank before selecting ten.
  const ranked = shuffled(pool, random).sort((a, b) => {
    const first = counts.get(a.id)!
    const second = counts.get(b.id)!
    return first.attempts - second.attempts
      || (first.lastPracticedAt ? Date.parse(first.lastPracticedAt) : 0) - (second.lastPracticedAt ? Date.parse(second.lastPracticedAt) : 0)
  })
  return createQuiz(ranked.slice(0, QUIZ_LENGTH), random)
}

export function describeQuizQuestions(questions: QuizQuestion[]): string {
  if (questions[0]?.mode === 'sentence') {
    const exercises = new Set(questions.map(question => question.exerciseId))
    return `JLPT Kanji exercise · ${exerciseTitle(exercises.size === 1 ? questions[0].exerciseId : undefined)}`
  }
  const label = questions[0]?.mode === 'reading' ? 'word readings' : 'meanings'
  if (questions.length === 0 || questions.some((question) => !question.estimatedLevel)) return label === 'word readings' ? 'Word readings' : 'Kanji meanings'
  const levels = new Set(questions.map((question) => question.estimatedLevel))
  const levelLabel = [...levels].sort((a, b) => Number(b!.slice(1)) - Number(a!.slice(1))).join('/')
  return levels.size === 1 ? `Estimated ${levelLabel} ${label}` : `Mixed ${levelLabel} ${label}`
}

export function createMissedQuiz(result: QuizResult, random: () => number = Math.random): QuizQuestion[] {
  const missed = result.items.filter((item) => !item.correct).map((item) => item.question)
  return shuffled(missed, random).map((question) => ({
    ...question,
    choices: shuffled(question.choices, random),
  }))
}

// The API orders candidates oldest first. Select before shuffling so newer
// practice does not crowd out items that have waited longer.
export function createCandidateQuiz(candidates: QuizCandidate[], random: () => number = Math.random): QuizQuestion[] {
  return shuffled(candidates.slice(0, QUIZ_LENGTH), random).map(({ question }) => ({
    ...question,
    choices: shuffled(question.choices, random),
  }))
}

export function candidatesForLevel(candidates: QuizCandidate[], level: QuizLevel, poolQuestions: QuizQuestion[]): QuizCandidate[] {
  const currentIds = new Set(poolQuestions.map((question) => question.id))
  return candidates.filter(({ question }) => question.estimatedLevel === level || currentIds.has(question.id))
}

export function gradeQuiz(questions: QuizQuestion[], answers: QuizAnswer[]): QuizResult {
  const items = questions.map((question) => {
    const selectedAnswer = answers.find((answer) => answer.questionId === question.id)?.selectedAnswer ?? null
    return { question, selectedAnswer, correct: selectedAnswer === question.answer }
  })
  return { items, total: items.length, correctCount: items.filter((item) => item.correct).length }
}

export function recordQuizAnswer(state: QuizState, answer: QuizAnswer): QuizState {
  if (state.status !== 'playing') return state
  const question = state.questions[state.answers.length]
  // Guard against an old form being submitted twice or an invalid answer value.
  if (!question || question.id !== answer.questionId || !question.choices.includes(answer.selectedAnswer)) return state

  const answers = [...state.answers, answer]
  if (answers.length === state.questions.length) {
    return { status: 'results', result: gradeQuiz(state.questions, answers) }
  }
  return { ...state, answers }
}
