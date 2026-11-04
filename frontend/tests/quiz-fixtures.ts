import { readFileSync } from 'node:fs'
import { createQuiz } from '../src/features/quiz/quiz.ts'
import type { QuizCoverage, QuizLevel, QuizMode, QuizQuestion, VocabularyExample } from '../src/features/quiz/types.ts'

// Tests use the canonical backend data; the browser loads these pools via HTTP.
const data: { pools: { mode: QuizMode; level: QuizLevel; questions: QuizQuestion[] }[] } = JSON.parse(
  readFileSync(new URL('../../backend/internal/quiz/data/questions.json', import.meta.url), 'utf8'),
)
data.pools.push(...JSON.parse(readFileSync(new URL('../../backend/internal/quiz/data/n2.generated.json', import.meta.url), 'utf8')).pools)
data.pools.push(...JSON.parse(readFileSync(new URL('../../backend/internal/quiz/data/exercises.json', import.meta.url), 'utf8')).pools)
export function fixturePool(mode: QuizMode, level: QuizLevel): QuizQuestion[] {
  return data.pools.find((pool) => pool.mode === mode && pool.level === level)!.questions.map((question) => ({
    ...question,
    estimatedLevel: level,
  }))
}
export const n4MeaningQuestions = fixturePool('meaning', 'N4')
export const wordReadingQuestions = fixturePool('reading', 'N5')
export const quizVocabulary: Record<string, VocabularyExample> = Object.fromEntries(
  data.pools.flatMap((pool) => pool.questions.map((question) => [question.example.written, question.example])),
)
export const createMeaningQuiz = (random: () => number = Math.random) => createQuiz(n4MeaningQuestions, random)
export const createReadingQuiz = (random: () => number = Math.random) => createQuiz(wordReadingQuestions, random)

export function coverageFixture(mode: QuizMode = 'meaning', level: QuizLevel = 'N5'): QuizCoverage {
  const questions = fixturePool(mode, level)
  return { mode, level, total: questions.length, attempted: 0, unseen: questions.length,
    questions: questions.map((question) => ({ questionId: question.id, attempts: 0, lastPracticedAt: null })) }
}
