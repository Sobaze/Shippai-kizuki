import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'
import { coreGlosses, disjointMeanings, distance, kanjiFromHTML, nearbyReadings, phoneticKey } from './build-n2.mjs'

const catalog = JSON.parse(await readFile(new URL('../quiz_n2.generated.json', import.meta.url), 'utf8'))
const audit = JSON.parse(await readFile(new URL('../quiz_n2.audit.json', import.meta.url), 'utf8'))

test('source parser handles the source website’s optional table closing tags', () => {
  const html = '<tr class=jl-row><td>1<td class="jl-td-k align-middle"><a>党</a><td>tou<td>nakama<td>party<tr class=jl-row><td>2<td class="jl-td-k align-middle"><a>協</a><td>kyou<td>cooperation'
  assert.deepEqual(kanjiFromHTML(html), ['党', '協'])
  assert.deepEqual(kanjiFromHTML('<tr><td>irrelevant</td></tr>'), [])
})

test('distractor checks reject synonyms, sound-equivalent spellings, and valid alternate readings', () => {
  assert.equal(disjointMeanings(['child', 'juvenile'], ['infancy', 'childhood']), false)
  assert.equal(disjointMeanings(['exchange', 'substitute'], ['change', 'reformation']), false)
  assert.equal(disjointMeanings(['lumber'], ['plank', 'board']), false)
  assert.equal(disjointMeanings(['salt'], ['forest']), true)
  assert.equal(phoneticKey('ちず'), phoneticKey('ちづ'))
  assert.equal(phoneticKey('こう'), phoneticKey('こー'))
  assert.equal(distance('せいとう', 'せんとう'), 1)
  const nearby = nearbyReadings('しゅうかん')
  assert.ok(nearby.includes('じゅうかん'))
  assert.ok(nearby.includes('しゅかん'))
  assert.ok(!nearby.includes('しゅうかん'))
})

test('all 374 N2-specific kanji have dictionary examples and non-overlapping meaning choices', () => {
  const pool = catalog.pools.find(pool => pool.mode === 'meaning')
  assert.equal(pool.questions.length, 374)
  assert.deepEqual(new Set(pool.questions.map(q => q.prompt)), new Set([...audit.sourceKanji]))
  const evidence = new Map(audit.meaningEvidence.map(entry => [entry.kanji, entry]))
  const choiceEvidence = new Map(audit.meaningEvidence.map(entry => [entry.answer, entry]))
  for (const question of pool.questions) {
    const source = evidence.get(question.prompt)
    const expected = coreGlosses[question.prompt] || source.meanings.slice(0, 2).join('; ')
    assert.equal(question.answer, expected.charAt(0).toUpperCase() + expected.slice(1))
    assert.equal(question.answer, source.answer)
    assert.deepEqual(question.example, source.example)
    assert.ok(question.example.written.includes(question.prompt))
    assert.equal(question.choices.length, 4)
    assert.equal(new Set(question.choices).size, 4)
    assert.equal(question.choices.filter(choice => choice === question.answer).length, 1)
    for (const choice of question.choices.filter(choice => choice !== question.answer)) {
      const other = choiceEvidence.get(choice)
      assert.ok(other, `${question.prompt}: distractor without dictionary evidence`)
      assert.ok(disjointMeanings([...source.meanings, question.answer], [...other.meanings, other.answer]), `${question.prompt}: conflicting choice ${choice}`)
    }
  }
})

test('all N2 reading questions have independent level evidence and exclude every recorded alternative reading', () => {
  const pool = catalog.pools.find(pool => pool.mode === 'reading')
  assert.equal(pool.questions.length, 1640)
  assert.equal(new Set(pool.questions.map(q => q.id)).size, 1640)
  assert.equal(new Set(pool.questions.map(q => q.prompt)).size, 1640)
  assert.equal(audit.sourceVocabularyRecords, 1856)
  const classified = new Set(audit.sourceVocabulary.filter(entry => entry.level === 'N2').map(entry => `${entry.written}/${entry.reading}`))
  const evidence = new Map(audit.readingEvidence.map(entry => [entry.written, entry]))
  for (const question of pool.questions) {
    const source = evidence.get(question.prompt)
    assert.ok(classified.has(`${question.prompt}/${question.answer}`), `No N2 classification: ${question.prompt}`)
    assert.ok(source.acceptedReadings.includes(question.answer))
    assert.deepEqual(question.example, { written: source.written, reading: source.reading, meaning: source.meaning })
    assert.equal(new Set(question.choices).size, 4)
    const valid = new Set(source.acceptedReadings.map(phoneticKey))
    assert.equal(question.choices.filter(choice => valid.has(phoneticKey(choice))).length, 1, `Ambiguous answer: ${question.prompt}`)
    for (const choice of question.choices) assert.match(choice, /^[\p{Script=Hiragana}ー]+$/u)
  }
  assert.deepEqual(audit.exclusions, [{ word: '嚙る', reading: 'かじる', reason: 'No exact written/reading dictionary match' }])
})
