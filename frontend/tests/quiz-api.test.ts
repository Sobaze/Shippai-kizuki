import assert from 'node:assert/strict'
import test from 'node:test'
import { getQuizAttempt, getQuizCandidates, getQuizCoverage, getQuizHistory, getQuizPool, saveQuizAttempt } from '../src/features/quiz/api.ts'
import { gradeQuiz } from '../src/features/quiz/quiz.ts'
import { coverageFixture, createMeaningQuiz, fixturePool } from './quiz-fixtures.ts'
import type { QuizAttemptSubmission, SavedQuizAttempt } from '../src/features/quiz/types.ts'

function fixture(id: string): QuizAttemptSubmission {
  const questions = createMeaningQuiz(() => 0)
  return {
    id, mode: 'meaning', kind: 'full',
    startedAt: '2026-11-01T10:00:00.000Z', completedAt: '2026-11-01T10:01:00.000Z',
    questions, answers: questions.map((question) => ({ questionId: question.id, selectedAnswer: question.answer })),
  }
}

test('named exercise pool and coverage requests include their exercise identity', async (t) => {
  const questions = fixturePool('sentence', 'N2')
  const coverage = { mode: 'sentence', level: 'N2', exerciseId: 'exercise-4', total: 12, attempted: 0, unseen: 12,
    questions: questions.map(q => ({ questionId: q.id, attempts: 0, lastPracticedAt: null })) }
  const calls: string[] = []
  t.mock.method(globalThis, 'fetch', async (url: string | URL | Request) => {
    calls.push(String(url))
    return Response.json(String(url).includes('/coverage') ? coverage : { mode: 'sentence', level: 'N2', exerciseId: 'exercise-4', title: 'Exercise 4', questions })
  })
  const signal = new AbortController().signal
  assert.equal((await getQuizPool('sentence', 'N2', signal, 'exercise-4')).questions.length, 12)
  assert.equal((await getQuizCoverage('sentence', 'N2', signal, 'exercise-4')).total, 12)
  assert.deepEqual(calls, ['/api/quiz/questions?mode=sentence&level=N2&exercise=exercise-4', '/api/quiz/coverage?mode=sentence&level=N2&exercise=exercise-4'])
  await assert.rejects(getQuizPool('sentence', 'N2', signal, 'exercise-5'), /mismatched/)
})

test('coverage waits for pending saves and uses selected mode/level with cancellation', async (t) => {
  const input = fixture('coverage-after-save')
  const coverage = coverageFixture('reading', 'N4')
  let complete!: () => void
  const gate = new Promise<void>((resolve) => { complete = resolve })
  const calls: string[] = []
  t.mock.method(globalThis, 'fetch', async (url: string | URL | Request, options?: RequestInit) => {
    calls.push(String(url))
    if (String(url) === '/api/quiz/attempts') {
      await gate
      return Response.json(saved(input))
    }
    assert.equal(options?.cache, 'no-store')
    return Response.json(coverage)
  })
  const write = saveQuizAttempt(input)
  const read = getQuizCoverage('reading', 'N4', new AbortController().signal)
  const cancelled = new AbortController()
  const cancelledRead = getQuizCoverage('meaning', 'N5', cancelled.signal)
  cancelled.abort()
  const rejection = assert.rejects(cancelledRead, { name: 'AbortError' })
  await Promise.resolve()
  assert.deepEqual(calls, ['/api/quiz/attempts'])
  complete()
  await write
  assert.deepEqual(await read, coverage)
  await rejection
  assert.deepEqual(calls, ['/api/quiz/attempts', '/api/quiz/coverage?mode=reading&level=N4'])
})

test('coverage errors and mismatched statistics are not silently treated as unseen', async (t) => {
  const coverage = coverageFixture()
  const signal = new AbortController().signal
  t.mock.method(globalThis, 'fetch', async () => new Response('404 page not found', { status: 404 }))
  await assert.rejects(getQuizCoverage('meaning', 'N5', signal), /Restart the Go backend/)
  t.mock.method(globalThis, 'fetch', async () => new Response('Database unavailable', { status: 500 }))
  await assert.rejects(getQuizCoverage('meaning', 'N5', signal), /Database unavailable/)
  for (const broken of [
    { ...coverage, level: 'N4' },
    { ...coverage, questions: [] },
    { ...coverage, unseen: 25 },
    { ...coverage, questions: coverage.questions.map((entry) => ({ ...entry, attempts: -1 })) },
    { ...coverage, questions: coverage.questions.map((entry) => ({ ...entry, attempts: 1, lastPracticedAt: 'yesterday' })) },
  ]) {
    t.mock.method(globalThis, 'fetch', async () => Response.json(broken))
    await assert.rejects(getQuizCoverage('meaning', 'N5', signal), /incomplete or mismatched/)
  }
})

test('question pools request the selected mode and level, propagate cancellation, and explain stale backends', async (t) => {
  const controller = new AbortController()
  const pool = { mode: 'meaning', level: 'N5', questions: fixturePool('meaning', 'N5'), version: 1, checkedAt: '2026-11-02', sources: {} }
  t.mock.method(globalThis, 'fetch', async (url: string | URL | Request, options?: RequestInit) => {
    assert.equal(url, '/api/quiz/questions?mode=meaning&level=N5')
    assert.equal(options?.signal, controller.signal)
    return Response.json(pool)
  })
  assert.deepEqual(await getQuizPool('meaning', 'N5', controller.signal), pool)
  t.mock.method(globalThis, 'fetch', async () => new Response('404 page not found', { status: 404 }))
  await assert.rejects(getQuizPool('meaning', 'N4', controller.signal), /Restart the Go backend/)
  t.mock.method(globalThis, 'fetch', async () => Response.json(pool))
  await assert.rejects(getQuizPool('meaning', 'N4', controller.signal), /mismatched/)
  t.mock.method(globalThis, 'fetch', async () => Response.json({ ...pool, questions: [] }))
  await assert.rejects(getQuizPool('meaning', 'N5', controller.signal), /incomplete/)
})

test('practice queue waits for saving and respects mode and cancellation', async (t) => {
  const input = fixture('queue-after-save')
  let complete!: () => void
  const gate = new Promise<void>((resolve) => { complete = resolve })
  const calls: string[] = []
  t.mock.method(globalThis, 'fetch', async (url: string | URL | Request) => {
    calls.push(String(url))
    if (String(url) === '/api/quiz/attempts') {
      await gate
      return Response.json(saved(input))
    }
    return Response.json({ candidates: [] })
  })
  const write = saveQuizAttempt(input)
  const queue = getQuizCandidates('reading', new AbortController().signal)
  const cancelled = new AbortController()
  const cancelledQueue = getQuizCandidates('meaning', cancelled.signal)
  cancelled.abort()
  const rejection = assert.rejects(cancelledQueue, { name: 'AbortError' })
  await Promise.resolve()
  assert.deepEqual(calls, ['/api/quiz/attempts'])
  complete()
  await write
  assert.deepEqual(await queue, [])
  await rejection
  assert.deepEqual(calls, ['/api/quiz/attempts', '/api/quiz/practice?mode=reading'])
})

function saved(input: QuizAttemptSubmission): SavedQuizAttempt {
  return {
    id: input.id, mode: input.mode, kind: input.kind, startedAt: input.startedAt, completedAt: input.completedAt,
    correctCount: 10, total: 10, result: gradeQuiz(input.questions, input.answers),
  }
}

test('completed attempts share one pending POST and send snapshots without client-calculated scores', async (t) => {
  const input = fixture('pending-save')
  let complete!: () => void
  const gate = new Promise<void>((resolve) => { complete = resolve })
  const fetch = t.mock.method(globalThis, 'fetch', async (url: string | URL | Request, options?: RequestInit) => {
    assert.equal(url, '/api/quiz/attempts')
    assert.equal(options?.method, 'POST')
    const body = JSON.parse(String(options?.body))
    assert.deepEqual(body, input)
    assert.equal('correctCount' in body, false)
    await gate
    return Response.json(saved(input), { status: 201 })
  })
  const first = saveQuizAttempt(input)
  const duplicate = saveQuizAttempt(input)
  assert.equal(first, duplicate)
  assert.equal(fetch.mock.callCount(), 1)
  complete()
  assert.deepEqual(await first, saved(input))
})

test('failed saves can retry the exact same ID and snapshot', async (t) => {
  const input = fixture('retry-save')
  const bodies: string[] = []
  t.mock.method(globalThis, 'fetch', async (_url: string | URL | Request, options?: RequestInit) => {
    bodies.push(String(options?.body))
    return bodies.length === 1 ? new Response('Database unavailable', { status: 500 }) : Response.json(saved(input))
  })
  await assert.rejects(saveQuizAttempt(input), /Database unavailable/)
  await saveQuizAttempt(input)
  assert.equal(bodies.length, 2)
  assert.equal(bodies[0], bodies[1])
})

test('history waits for a pending save and then requests only the selected mode', async (t) => {
  const input = fixture('history-after-save')
  let complete!: () => void
  const gate = new Promise<void>((resolve) => { complete = resolve })
  const calls: string[] = []
  t.mock.method(globalThis, 'fetch', async (url: string | URL | Request) => {
    calls.push(String(url))
    if (String(url) === '/api/quiz/attempts') {
      await gate
      return Response.json(saved(input), { status: 201 })
    }
    return Response.json({ attempts: [saved(input)] })
  })
  const write = saveQuizAttempt(input)
  const history = getQuizHistory('meaning', new AbortController().signal)
  await Promise.resolve()
  assert.deepEqual(calls, ['/api/quiz/attempts'])
  complete()
  await write
  assert.equal((await history)[0].id, input.id)
  assert.deepEqual(calls, ['/api/quiz/attempts', '/api/quiz/attempts?mode=meaning'])
})

test('saved review uses the detail endpoint and does not create another attempt', async (t) => {
  const input = fixture('saved-review')
  const calls: string[] = []
  t.mock.method(globalThis, 'fetch', async (url: string | URL | Request, options?: RequestInit) => {
    calls.push(String(url))
    assert.equal(options?.method, undefined)
    return Response.json(saved(input))
  })
  assert.deepEqual(await getQuizAttempt(input.id, new AbortController().signal), saved(input))
  assert.deepEqual(calls, ['/api/quiz/attempts/saved-review'])
})
