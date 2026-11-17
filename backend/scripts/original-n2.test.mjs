import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const directory = new URL('../internal/quiz/data/', import.meta.url)
const data = JSON.parse(await readFile(new URL('original-n2.json', directory), 'utf8'))
const audit = JSON.parse(await readFile(new URL('n2.audit.json', directory), 'utf8'))

test('original kanji pilot uses N2-classified targets with three readings and three spellings per set', () => {
  const pairs = new Set(audit.sourceVocabulary.filter(word => word.level === 'N2').map(word => `${word.written}/${word.reading}`))
  const dictionary = new Map(audit.readingEvidence.map(word => [word.written, word]))
  const keys = [[3, 2, 4, 1, 2, 4], [3, 1, 4, 3, 2, 3]]
  const pools = data.pools.filter(pool => pool.category === 'kanji')
  assert.equal(pools.length, 2)
  pools.forEach((pool, index) => {
    assert.equal(pool.questions.length, 6)
    assert.equal(pool.questions.filter(q => q.task === 'reading').length, 3)
    assert.equal(pool.questions.filter(q => q.task === 'spelling').length, 3)
    assert.deepEqual(pool.questions.map(q => q.choices.indexOf(q.answer) + 1), keys[index])
    for (const q of pool.questions) {
      assert.ok(pairs.has(`${q.example.written}/${q.example.reading}`), `Target is outside the N2 source: ${q.example.written}`)
      assert.ok(q.explanation && q.context && q.context.before + q.prompt + q.context.after)
      assert.equal(q.task === 'reading' ? q.example.reading : q.example.written, q.answer)
      if (q.task === 'reading') {
        const entry = dictionary.get(q.example.written)
        assert.ok(entry?.acceptedReadings.includes(q.answer))
        assert.equal(q.choices.filter(choice => entry.acceptedReadings.includes(choice)).length, 1)
      }
    }
  })
})

test('three original comprehension passages have grounded evidence, stable passages, and four unique choices', () => {
  const pools = data.pools.filter(pool => pool.category === 'comprehension')
  assert.equal(pools.length, 3)
  const keys = [[2, 3], [1, 4], [2, 4]]
  pools.forEach((pool, index) => {
    assert.equal(pool.questions.length, 2)
    assert.deepEqual(pool.questions.map(q => q.choices.indexOf(q.answer) + 1), keys[index])
    assert.deepEqual(pool.questions[0].passage, pool.questions[1].passage)
    for (const q of pool.questions) {
      assert.equal(q.task, 'comprehension')
      assert.ok([...q.passage.text].length >= 250)
      assert.ok(q.passage.text.includes(q.evidence))
      assert.ok(q.explanation.length >= 80)
      assert.equal(q.choices.length, 4)
      assert.equal(new Set(q.choices).size, 4)
      assert.equal(q.choices.filter(choice => choice === q.answer).length, 1)
      assert.equal(q.context, undefined)
    }
  })
  assert.equal(data.pools.flatMap(pool => pool.questions).length, 18)
})
