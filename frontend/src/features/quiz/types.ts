export type QuizMode = 'meaning' | 'reading' | 'sentence'
export type QuizAttemptKind = 'full' | 'retry'
export type QuizLevel = 'N5' | 'N4' | 'N2'

export type VocabularyExample = {
  written: string
  reading: string
  meaning: string
}

export type QuizQuestion = {
  id: string
  mode: QuizMode
  prompt: string
  answer: string
  choices: string[]
  example: VocabularyExample
  estimatedLevel?: QuizLevel
  exerciseId?: string
  task?: 'reading' | 'spelling'
  context?: { before: string; after: string }
}

export type QuizPool = {
  mode: QuizMode
  level: QuizLevel
  questions: QuizQuestion[]
  version: number
  checkedAt: string
  sources: Record<string, string>
  exerciseId?: string
  title?: string
}

export type QuizAnswer = {
  questionId: string
  selectedAnswer: string
}

export type QuizQuestionCoverage = {
  questionId: string
  attempts: number
  lastPracticedAt: string | null
}

export type QuizCoverage = {
  mode: QuizMode
  level: QuizLevel
  total: number
  attempted: number
  unseen: number
  questions: QuizQuestionCoverage[]
  exerciseId?: string
}

export type QuizCandidate = {
  question: QuizQuestion
  streak: number
  lastPracticedAt: string
}

export type QuizResultItem = {
  question: QuizQuestion
  selectedAnswer: string | null
  correct: boolean
}

export type QuizResult = {
  correctCount: number
  total: number
  items: QuizResultItem[]
}

export type QuizState =
  | { status: 'ready' }
  | { status: 'playing'; questions: QuizQuestion[]; answers: QuizAnswer[] }
  | { status: 'results'; result: QuizResult }

export type QuizRun = {
  id: string
  mode: QuizMode
  kind: QuizAttemptKind
  startedAt: string
  parentAttemptId?: string
}

export type QuizAttemptSubmission = QuizRun & {
  completedAt: string
  questions: QuizQuestion[]
  answers: QuizAnswer[]
}

export type QuizAttemptSummary = {
  id: string
  mode: QuizMode
  kind: QuizAttemptKind
  completedAt: string
  correctCount: number
  total: number
  exerciseId?: string
}

export type SavedQuizAttempt = QuizAttemptSummary & {
  startedAt: string
  parentAttemptId?: string
  result: QuizResult
}
