import assert from 'node:assert/strict'
import test from 'node:test'
import { n4MeaningQuestions, wordReadingQuestions, quizVocabulary, createMeaningQuiz, createReadingQuiz, coverageFixture, fixturePool } from './quiz-fixtures.ts'
import { candidatesForLevel, createCandidateQuiz, createCoverageQuiz, createExerciseQuiz, createQuiz, createMissedQuiz, describeQuizQuestions, gradeQuiz, QUIZ_LENGTH, recordQuizAnswer } from '../src/features/quiz/quiz.ts'
import type { QuizState } from '../src/features/quiz/types.ts'

test('Exercise 4 retains all twelve contextual targets, source order, answer keys, and task identities', () => {
  const pool = fixturePool('sentence', 'N2')
  const snapshot = JSON.stringify(pool)
  const key = [1, 3, 4, 3, 1, 3, 1, 3, 4, 3, 4, 2]
  assert.equal(pool.length, 12)
  pool.forEach((question, index) => assert.equal(question.choices[key[index] - 1], question.answer))
  const questions = createExerciseQuiz(pool, () => 0)
  assert.deepEqual(questions.map(q => q.id), pool.map(q => q.id))
  assert.equal(questions.filter(q => q.task === 'reading').length, 6)
  assert.equal(questions.filter(q => q.task === 'spelling').length, 6)
  for (const question of questions) {
    assert.ok(question.context?.before && question.context?.after)
    assert.equal(question.exerciseId, 'exercise-4')
    assert.deepEqual(new Set(question.choices), new Set(pool.find(q => q.id === question.id)!.choices))
  }
  assert.equal(describeQuizQuestions(questions), 'JLPT Kanji exercise · Exercise 4')
  assert.equal(JSON.stringify(pool), snapshot)
  let state: QuizState = { status: 'playing', questions, answers: [] }
  questions.forEach((question, index) => {
    state = recordQuizAnswer(state, { questionId: question.id, selectedAnswer: question.answer })
    assert.equal(state.status, index === 11 ? 'results' : 'playing')
  })
  const wrong = gradeQuiz(questions, [])
  const retry = createMissedQuiz(wrong, () => 0)
  assert.equal(retry.length, 12)
  assert.deepEqual(new Set(retry.map(q => q.id)), new Set(questions.map(q => q.id)))
  assert.ok(retry.every(q => q.context && q.task && q.exerciseId === 'exercise-4'))
})

test('large N2 pools achieve full first-pass coverage with level-preserving snapshots', () => {
  for (const mode of ['meaning', 'reading'] as const) {
    const pool = fixturePool(mode, 'N2')
    assert.equal(pool.length, mode === 'meaning' ? 374 : 1640)
    const coverage = coverageFixture(mode, 'N2')
    const byId = new Map(coverage.questions.map(entry => [entry.questionId, entry]))
    const seen = new Set<string>()
    for (let run = 0; run < Math.ceil(pool.length / 10); run++) {
      const questions = createCoverageQuiz(pool, coverage, () => 0.5)
      assert.equal(questions.length, 10)
      assert.equal(new Set(questions.map(q => q.id)).size, 10)
      const unseenBefore = pool.length - seen.size
      assert.equal(questions.filter(q => !seen.has(q.id)).length, Math.min(unseenBefore, 10))
      for (const question of questions) {
        assert.equal(question.estimatedLevel, 'N2')
        seen.add(question.id)
        const entry = byId.get(question.id)!
        entry.attempts++
        entry.lastPracticedAt = new Date(Date.UTC(2026, 10, 1, 12, run)).toISOString()
      }
      coverage.attempted = seen.size
      coverage.unseen = pool.length - seen.size
    }
    assert.equal(seen.size, pool.length)
    assert.equal(describeQuizQuestions(pool), `Estimated N2 ${mode === 'reading' ? 'word readings' : 'meanings'}`)
  }
})

test('selected-level queues keep N2 focused while all-level data and legacy matching IDs remain available', () => {
  const n2 = fixturePool('reading', 'N2')[0]
  const n5 = fixturePool('reading', 'N5')[0]
  const legacy = { ...n2, estimatedLevel: undefined }
  const candidates = [n5, legacy, { ...n2, id: 'another-n2-target' }].map(question => ({ question, streak: 0, lastPracticedAt: '2026-11-01T12:00:00Z' }))
  const focused = candidatesForLevel(candidates, 'N2', [n2])
  assert.deepEqual(focused.map(c => c.question.id), [n2.id, 'another-n2-target'])
  assert.equal(candidates.length, 3)
  assert.equal(describeQuizQuestions([n2, n5]), 'Mixed N5/N2 word readings')
  assert.equal(describeQuizQuestions([n2, { ...fixturePool('reading', 'N4')[0] }]), 'Mixed N4/N2 word readings')
})

test('coverage selection prioritizes unseen items in every mode and level without mutating snapshots', () => {
  for (const mode of ['meaning', 'reading'] as const) for (const level of ['N5', 'N4'] as const) {
    const pool = fixturePool(mode, level)
    const coverage = coverageFixture(mode, level)
    coverage.attempted = 14
    coverage.unseen = 10
    coverage.questions.slice(0, 14).forEach((entry) => { entry.attempts = 1; entry.lastPracticedAt = '2026-11-01T12:00:00Z' })
    const snapshot = JSON.stringify({ pool, coverage })
    const questions = createCoverageQuiz(pool, coverage, () => 0)
    assert.deepEqual(new Set(questions.map((question) => question.id)), new Set(pool.slice(14).map((question) => question.id)))
    assert.equal(questions.length, 10)
    for (const question of questions) {
      const source = pool.find((entry) => entry.id === question.id)!
      assert.deepEqual(new Set(question.choices), new Set(source.choices))
      assert.equal(question.estimatedLevel, level)
      assert.deepEqual(question.example, source.example)
    }
    assert.equal(JSON.stringify({ pool, coverage }), snapshot)
  }
})

test('fewer attempts outrank recency; recency breaks count ties before final question shuffling', () => {
  const pool = fixturePool('meaning', 'N5')
  const coverage = coverageFixture()
  coverage.questions.forEach((entry, index) => {
    entry.attempts = index < 16 ? 4 : index < 20 ? 1 : 0
    entry.lastPracticedAt = entry.attempts ? new Date(Date.UTC(2026, 10, 1, 12, index)).toISOString() : null
  })
  coverage.attempted = 20
  coverage.unseen = 4
  const questions = createCoverageQuiz(pool, coverage, () => 0.5)
  assert.deepEqual(new Set(questions.map((question) => question.id)), new Set([...pool.slice(16), ...pool.slice(0, 2)].map((question) => question.id)))
  assert.notDeepEqual(questions.slice(0, 4).map((question) => question.id), pool.slice(20).map((question) => question.id))
})

test('tied candidates are randomized and three completed selections cover the 24-item pool', () => {
  const pool = fixturePool('meaning', 'N5')
  const coverage = coverageFixture()
  const ids = (random: () => number) => createCoverageQuiz(pool, coverage, random).map((question) => question.id)
  assert.notDeepEqual(new Set(ids(() => 0)), new Set(ids(() => 0.999)))
  const visited = new Set<string>()
  for (let run = 0; run < 3; run++) {
    const questions = createCoverageQuiz(pool, coverage, () => 0.5)
    assert.equal(new Set(questions.map((question) => question.id)).size, 10)
    if (run < 2) assert.ok(questions.every((question) => !visited.has(question.id)))
    for (const question of questions) {
      visited.add(question.id)
      const entry = coverage.questions.find((item) => item.questionId === question.id)!
      entry.attempts++
      entry.lastPracticedAt = new Date(Date.UTC(2026, 10, 1, 12, run)).toISOString()
    }
    coverage.attempted = visited.size
    coverage.unseen = coverage.total - visited.size
  }
  assert.equal(visited.size, 24)
  assert.ok(coverage.questions.every((entry) => entry.attempts >= 1 && entry.attempts <= 2))
  const otherLevel = coverageFixture('meaning', 'N4')
  assert.throws(() => createCoverageQuiz(pool, otherLevel), /does not match/)
  assert.throws(() => createCoverageQuiz(pool, { ...coverage, questions: coverage.questions.slice(1) }), /does not match/)
})

test('candidate sessions select oldest ten before shuffling and support small queues', () => {
  const questions = createMeaningQuiz(() => 0)
  const candidates = Array.from({ length: 12 }, (_, index) => ({
    question: { ...questions[index % questions.length], id: `candidate-${index}` },
    streak: index % 2,
    lastPracticedAt: new Date(index * 1000).toISOString(),
  }))
  const original = JSON.stringify(candidates)
  const quiz = createCandidateQuiz(candidates, () => 0)
  assert.equal(quiz.length, 10)
  assert.deepEqual(new Set(quiz.map((question) => question.id)), new Set(candidates.slice(0, 10).map(({ question }) => question.id)))
  assert.notDeepEqual(quiz.map((question) => question.id), candidates.slice(0, 10).map(({ question }) => question.id))
  for (const question of quiz) {
    const source = candidates.find((candidate) => candidate.question.id === question.id)!.question
    assert.deepEqual(new Set(question.choices), new Set(source.choices))
    assert.equal(question.answer, source.answer)
    assert.deepEqual(question.example, source.example)
  }
  assert.equal(JSON.stringify(candidates), original)
  assert.equal(createCandidateQuiz(candidates.slice(0, 1)).length, 1)
  assert.deepEqual(createCandidateQuiz([]), [])
})

test('checked pool contains distinct kanji and exactly one declared answer per four choices', () => {
  assert.ok(n4MeaningQuestions.length >= QUIZ_LENGTH)
  assert.equal(new Set(n4MeaningQuestions.map((question) => question.id)).size, n4MeaningQuestions.length)
  assert.equal(new Set(n4MeaningQuestions.map((question) => question.prompt)).size, n4MeaningQuestions.length)
  for (const question of n4MeaningQuestions) {
    assert.equal([...question.prompt].length, 1)
    const choices = question.choices
    assert.equal(choices.length, 4)
    assert.equal(new Set(choices.map((choice) => choice.toLowerCase().trim())).size, 4)
    assert.ok(question.example.written.includes(question.prompt))
    assert.ok(question.example.reading && question.example.meaning)
  }
})

test('quiz creation draws ten without replacement, shuffles choices, and preserves the source pool', () => {
  const original = JSON.stringify(n4MeaningQuestions)
  const quiz = createMeaningQuiz(() => 0)
  assert.equal(quiz.length, QUIZ_LENGTH)
  assert.equal(new Set(quiz.map((question) => question.id)).size, QUIZ_LENGTH)
  for (const question of quiz) {
    assert.ok(n4MeaningQuestions.some((entry) => entry.id === question.id))
    assert.equal(new Set(question.choices).size, 4)
    assert.equal(question.choices.filter((choice) => choice === question.answer).length, 1)
  }
  assert.deepEqual(quiz, createMeaningQuiz(() => 0))
  assert.notDeepEqual(quiz.map((question) => question.id), createMeaningQuiz(() => 0.999).map((question) => question.id))
  assert.equal(JSON.stringify(n4MeaningQuestions), original)
})

test('each meaning level draws only from its own pool and retains level snapshots through retries', () => {
  for (const level of ['N5', 'N4'] as const) {
    const pool = fixturePool('meaning', level)
    const quiz = createQuiz(pool, () => 0.5)
    assert.equal(quiz.length, 10)
    assert.ok(quiz.every((question) => question.estimatedLevel === level && pool.some((entry) => entry.id === question.id)))
    assert.equal(describeQuizQuestions(quiz), `Estimated ${level} meanings`)
    const retry = createMissedQuiz(gradeQuiz(quiz, []))
    assert.ok(retry.every((question) => question.estimatedLevel === level))
  }
  assert.equal(describeQuizQuestions([fixturePool('meaning', 'N5')[0], fixturePool('meaning', 'N4')[0]]), 'Mixed N5/N4 meanings')
  const legacy = { ...n4MeaningQuestions[0] }
  delete legacy.estimatedLevel
  assert.equal(describeQuizQuestions([legacy]), 'Kanji meanings')
  assert.equal(describeQuizQuestions(wordReadingQuestions), 'Estimated N5 word readings')
  assert.throws(() => createQuiz(n4MeaningQuestions.slice(0, 9)), /at least 10/)
})

test('reading levels retain vocabulary classification in quizzes and mixed/legacy retries', () => {
  for (const level of ['N5', 'N4'] as const) {
    const pool = fixturePool('reading', level)
    const questions = createQuiz(pool, () => 0)
    assert.equal(pool.length, 24)
    assert.equal(new Set(questions.map((question) => question.id)).size, 10)
    assert.ok(questions.every((question) => question.estimatedLevel === level && question.mode === 'reading' && pool.some((entry) => entry.id === question.id)))
    assert.equal(describeQuizQuestions(questions), `Estimated ${level} word readings`)
    const retry = createMissedQuiz(gradeQuiz(questions, []))
    assert.ok(retry.every((question) => question.estimatedLevel === level))
  }
  const mixed = [fixturePool('reading', 'N5')[0], fixturePool('reading', 'N4')[0]]
  assert.equal(describeQuizQuestions(mixed), 'Mixed N5/N4 word readings')
  const legacy = { ...mixed[0] }
  delete legacy.estimatedLevel
  assert.equal(describeQuizQuestions([legacy]), 'Word readings')
  assert.equal(describeQuizQuestions([legacy, ...mixed]), 'Word readings')
  const queue = createCandidateQuiz(mixed.map((question) => ({ question, streak: 0, lastPracticedAt: '2026-11-01T12:00:00Z' })))
  assert.deepEqual(new Set(queue.map((question) => question.estimatedLevel)), new Set(['N5', 'N4']))
})

test('invalid, duplicate, and stale submissions cannot skip a question', () => {
  const questions = createMeaningQuiz(() => 0)
  const state: QuizState = { status: 'playing', questions, answers: [] }
  assert.equal(recordQuizAnswer(state, { questionId: questions[0].id, selectedAnswer: 'not an option' }), state)
  assert.equal(recordQuizAnswer(state, { questionId: questions[1].id, selectedAnswer: questions[1].answer }), state)

  const answer = { questionId: questions[0].id, selectedAnswer: questions[0].answer }
  const next = recordQuizAnswer(state, answer)
  assert.equal(next.status, 'playing')
  if (next.status !== 'playing') return
  assert.equal(next.answers.length, 1)
  assert.equal(state.answers.length, 0)
  assert.equal(recordQuizAnswer(next, answer), next)
})

test('ten answers complete a quiz and retain every selected answer in question order', () => {
  const questions = createMeaningQuiz(() => 0.5)
  let state: QuizState = { status: 'playing', questions, answers: [] }
  const selected = questions.map((question, index) => index % 2 === 0 ? question.answer : question.choices.find((choice) => choice !== question.answer)!)
  for (let index = 0; index < questions.length; index++) {
    state = recordQuizAnswer(state, { questionId: questions[index].id, selectedAnswer: selected[index] })
    assert.equal(state.status, index === questions.length - 1 ? 'results' : 'playing')
  }
  if (state.status !== 'results') return
  assert.equal(state.result.correctCount, 5)
  assert.equal(state.result.total, 10)
  assert.deepEqual(state.result.items.map((item) => item.selectedAnswer), selected)
  assert.deepEqual(state.result.items.map((item) => item.question.id), questions.map((question) => question.id))
  assert.equal(recordQuizAnswer(state, { questionId: questions[9].id, selectedAnswer: selected[9] }), state)
})

test('grading handles all-correct, all-incorrect, and unanswered questions without inflating scores', () => {
  const questions = createMeaningQuiz(() => 0)
  const correct = questions.map((question) => ({ questionId: question.id, selectedAnswer: question.answer }))
  assert.equal(gradeQuiz(questions, correct).correctCount, 10)
  assert.equal(gradeQuiz(questions, [...correct, ...correct]).correctCount, 10)
  const wrong = questions.map((question) => ({ questionId: question.id, selectedAnswer: question.choices.find((choice) => choice !== question.answer)! }))
  assert.equal(gradeQuiz(questions, wrong).correctCount, 0)
  const unanswered = gradeQuiz(questions, [])
  assert.equal(unanswered.correctCount, 0)
  assert.ok(unanswered.items.every((item) => item.selectedAnswer === null))
})

test('vocabulary feedback keeps a written word, kana reading, and meaning together', () => {
  for (const [written, example] of Object.entries(quizVocabulary)) {
    assert.equal(example.written, written)
    assert.match(example.reading, /^[\p{Script=Hiragana}ー]+$/u)
    assert.ok(example.meaning.trim().length > 0)
  }
  assert.equal(quizVocabulary['新しい'].reading, 'あたらしい')
  assert.equal(quizVocabulary['新聞'].reading, 'しんぶん')
})

test('reading quizzes draw distinct whole words with four kana options and stable feedback', () => {
  const original = JSON.stringify({ wordReadingQuestions, quizVocabulary })
  const questions = createReadingQuiz(() => 0.999)
  assert.equal(questions.length, QUIZ_LENGTH)
  assert.equal(new Set(questions.map((question) => question.id)).size, QUIZ_LENGTH)
  assert.equal(new Set(questions.map((question) => question.prompt)).size, QUIZ_LENGTH)
  assert.equal(new Set(wordReadingQuestions.map((question) => question.example.written)).size, wordReadingQuestions.length)
  assert.ok(questions.some((question) => question.prompt === '新しい'))
  assert.ok(questions.some((question) => question.prompt === '新聞'))
  for (const question of questions) {
    assert.equal(question.mode, 'reading')
    assert.equal(question.prompt, question.example.written)
    assert.equal(question.answer, question.example.reading)
    assert.equal(new Set(question.choices).size, 4)
    assert.equal(question.choices.filter((choice) => choice === question.answer).length, 1)
    for (const choice of question.choices) assert.match(choice, /^[\p{Script=Hiragana}ー]+$/u)
  }
  assert.deepEqual(questions, createReadingQuiz(() => 0.999))
  assert.equal(JSON.stringify({ wordReadingQuestions, quizVocabulary }), original)
})

test('reading attempts use the shared progression and retain word/reading/meaning results', () => {
  const questions = createReadingQuiz(() => 0)
  let state: QuizState = { status: 'playing', questions, answers: [] }
  for (let index = 0; index < questions.length; index++) {
    const question = questions[index]
    const selectedAnswer = index % 2 === 0 ? question.answer : question.choices.find((choice) => choice !== question.answer)!
    state = recordQuizAnswer(state, { questionId: question.id, selectedAnswer })
  }
  assert.equal(state.status, 'results')
  if (state.status !== 'results') return
  assert.equal(state.result.correctCount, 5)
  assert.ok(state.result.items.every((item) => item.question.mode === 'reading'))
  assert.deepEqual(state.result.items.map((item) => item.question.example), questions.map((question) => question.example))
  const study = questions.find((question) => question.prompt === '勉強')
  assert.ok(study, 'the deterministic fixture includes the long-vowel question')
  assert.equal(gradeQuiz([study], [{ questionId: study.id, selectedAnswer: 'べんきょ' }]).correctCount, 0)
  assert.equal(gradeQuiz([study], [{ questionId: study.id, selectedAnswer: 'べんきょう' }]).correctCount, 1)
})

test('missed-item retries use only missed snapshots and leave the original result unchanged', () => {
  for (const questions of [createMeaningQuiz(() => 0), createReadingQuiz(() => 0)]) {
    const answers = questions.map((question, index) => ({
      questionId: question.id,
      selectedAnswer: index < 3 ? question.choices.find((choice) => choice !== question.answer)! : question.answer,
    }))
    const original = gradeQuiz(questions, answers)
    const snapshot = JSON.stringify(original)
    const retry = createMissedQuiz(original, () => 0)
    assert.equal(retry.length, 3)
    assert.deepEqual(new Set(retry.map((question) => question.id)), new Set(questions.slice(0, 3).map((question) => question.id)))
    for (const question of retry) {
      const saved = original.items.find((item) => item.question.id === question.id)!.question
      assert.equal(question.prompt, saved.prompt)
      assert.equal(question.mode, saved.mode)
      assert.equal(question.answer, saved.answer)
      assert.deepEqual(question.example, saved.example)
      assert.deepEqual(new Set(question.choices), new Set(saved.choices))
    }
    assert.equal(JSON.stringify(original), snapshot)
    const retried = gradeQuiz(retry, retry.map((question) => ({ questionId: question.id, selectedAnswer: question.answer })))
    assert.equal(retried.correctCount, 3)
    assert.equal(retried.total, 3)
    assert.equal(original.correctCount, 7)
    assert.equal(original.total, 10)
  }
})

test('perfect attempts have no missed-item retry and one miss produces one question', () => {
  const questions = createMeaningQuiz(() => 0)
  const answers = questions.map((question) => ({ questionId: question.id, selectedAnswer: question.answer }))
  assert.deepEqual(createMissedQuiz(gradeQuiz(questions, answers)), [])
  answers[0].selectedAnswer = questions[0].choices.find((choice) => choice !== questions[0].answer)!
  assert.equal(createMissedQuiz(gradeQuiz(questions, answers)).length, 1)
})
