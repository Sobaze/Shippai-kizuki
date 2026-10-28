import assert from 'node:assert/strict'
import test from 'node:test'
import { getCachedKanjiDetails, getKanjiDetails, prefetchKanjiDetails } from '../src/features/kanji/api.ts'

function entry(kanji: string) {
  return {
    kanji, meanings: ['example'], onReadings: ['ショ'],
    kunReadings: ['か.く'], nameReadings: [], strokeCount: 10,
  }
}

test('concurrent lookups share one request and successful results are reused', async (t) => {
  let complete!: () => void
  const gate = new Promise<void>((resolve) => { complete = resolve })
  const fetch = t.mock.method(globalThis, 'fetch', async () => {
    await gate
    return Response.json(entry('書'))
  })

  assert.equal(getCachedKanjiDetails('書'), undefined)
  const first = getKanjiDetails('書')
  const second = getKanjiDetails('書')
  assert.equal(first, second)
  assert.equal(fetch.mock.callCount(), 1)

  complete()
  assert.deepEqual(await first, entry('書'))
  assert.deepEqual(await second, entry('書'))
  assert.deepEqual(getCachedKanjiDetails('書'), entry('書'))
  assert.deepEqual(await getKanjiDetails('書'), entry('書'))
  assert.equal(fetch.mock.callCount(), 1)
})

test('failed lookups are not cached and can be retried', async (t) => {
  let attempts = 0
  t.mock.method(globalThis, 'fetch', async () => {
    attempts++
    return attempts === 1
      ? new Response('Dictionary unavailable', { status: 502 })
      : Response.json(entry('日'))
  })

  await assert.rejects(getKanjiDetails('日'), /Dictionary unavailable/)
  assert.equal(getCachedKanjiDetails('日'), undefined)
  assert.deepEqual(await getKanjiDetails('日'), entry('日'))
  assert.equal(attempts, 2)
})

test('hover/focus preloads share a pending lookup with the popup', async (t) => {
  let complete!: () => void
  const gate = new Promise<void>((resolve) => { complete = resolve })
  const fetch = t.mock.method(globalThis, 'fetch', async () => {
    await gate
    return Response.json(entry('読'))
  })

  prefetchKanjiDetails('読')
  prefetchKanjiDetails('読')
  const opened = getKanjiDetails('読')
  assert.equal(fetch.mock.callCount(), 1)
  complete()
  assert.deepEqual(await opened, entry('読'))
  assert.deepEqual(getCachedKanjiDetails('読'), entry('読'))
})

test('different kanji keep separate cache entries', async (t) => {
  const fetch = t.mock.method(globalThis, 'fetch', async (input: string | URL | Request) => {
    const character = decodeURIComponent(String(input).split('/').at(-1)!)
    return Response.json(entry(character))
  })
  const [food, book] = await Promise.all([getKanjiDetails('食'), getKanjiDetails('本')])
  assert.equal(food.kanji, '食')
  assert.equal(book.kanji, '本')
  assert.equal(fetch.mock.callCount(), 2)
})

test('invalid JSON does not poison the cache', async (t) => {
  let attempts = 0
  t.mock.method(globalThis, 'fetch', async () => {
    attempts++
    return attempts === 1
      ? new Response('invalid JSON', { status: 200 })
      : Response.json(entry('館'))
  })
  await assert.rejects(getKanjiDetails('館'), SyntaxError)
  assert.equal(getCachedKanjiDetails('館'), undefined)
  assert.deepEqual(await getKanjiDetails('館'), entry('館'))
})
