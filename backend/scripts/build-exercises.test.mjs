import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'
import { compileExercises } from './build-exercises.mjs'

const directory = new URL('../internal/quiz/data/', import.meta.url)
const source = JSON.parse(await readFile(new URL('exercises.supplied.json', directory), 'utf8'))
const generated = JSON.parse(await readFile(new URL('exercises.additional.generated.json', directory), 'utf8'))
const original = JSON.parse(await readFile(new URL('exercises.json', directory), 'utf8'))

// Independently transcribed from the answer keys supplied in the conversation.
const keys = {
  1: [1, 1, 3, 4, 2, 1, 2, 1, 2, 3, 4, 1],
  2: [3, 4, 3, 4, 1, 3, 3, 2, 4, 3, 2, 2],
  3: [4, 3, 2, 2, 4, 2, 4, 3, 3, 2, 2, 1],
  4: [1, 3, 4, 3, 1, 3, 1, 3, 4, 3, 4, 2],
  5: [3, 1, 4, 3, 2, 4, 3, 1, 2, 1, 3, 2],
  6: [3, 2, 4, 3, 2, 1, 2, 3, 1, 4, 3, 1],
  7: [1, 4, 1, 3, 3, 3, 4, 1, 2, 1],
  8: [1, 3, 4, 3, 2, 1, 3, 3, 2, 1],
  9: [4, 1, 2, 4, 2, 3, 1, 2, 4, 1],
  10: [2, 1, 2, 3, 3, 2, 1, 4, 4, 1],
  11: [4, 2, 3, 3, 1, 2, 4, 1, 3, 2],
}

test('all eleven exercises retain the supplied lengths, answer positions, and task counts', () => {
  assert.deepEqual(generated, compileExercises(source))
  const pools = [...original.pools, ...generated.pools]
  assert.equal(pools.length, 11)
  assert.equal(new Set(pools.map(pool => pool.exerciseId)).size, 11)
  let total = 0, readings = 0, spellings = 0
  for (const pool of pools) {
    const number = Number(pool.exerciseId.replace('exercise-', ''))
    assert.equal(pool.title, `Exercise ${number}`)
    assert.equal(pool.questions.length, number <= 6 ? 12 : 10)
    assert.deepEqual(pool.questions.map(question => question.choices.indexOf(question.answer) + 1), keys[number])
    for (const question of pool.questions) {
      assert.equal(question.exerciseId, pool.exerciseId)
      assert.equal(question.choices.length, 4)
      assert.equal(new Set(question.choices).size, 4)
      assert.equal(question.task === 'reading' ? question.example.reading : question.example.written, question.answer)
      total++
      if (question.task === 'reading') readings++
      if (question.task === 'spelling') spellings++
    }
  }
  assert.equal(total, 122)
  assert.equal(readings, 68)
  assert.equal(spellings, 54)
})

test('shared passages remain complete, with the requested target occurrence preserved', () => {
  for (const exercise of source.exercises) {
    const pool = generated.pools.find(pool => pool.exerciseId === `exercise-${exercise.number}`)
    exercise.questions.forEach((question, index) => {
      const rendered = pool.questions[index]
      assert.equal(rendered.context.before + rendered.prompt + rendered.context.after, exercise.passages[question.passage])
      assert.equal(rendered.prompt, question.target)
      assert.equal(rendered.id, `${pool.exerciseId}-${question.task}-${question.number}`)
    })
  }
  const seven = generated.pools.find(pool => pool.exerciseId === 'exercise-7')
  assert.equal(new Set(seven.questions.slice(0, 4).map(q => q.context.before + q.prompt + q.context.after)).size, 1)
  const eight = generated.pools.find(pool => pool.exerciseId === 'exercise-8')
  assert.ok(eight.questions[2].context.after.includes('いそがしく'), 'only the selected occurrence is highlighted; the rest of the passage remains intact')
  assert.equal(eight.questions[2].context.before.includes('いそがしく'), false)
})

test('content build rejects missing targets, ambiguous occurrences, and mismatched answer keys', () => {
  for (const mutate of [
    data => { data.exercises[0].questions[0].target = 'not in the sentence' },
    data => { data.exercises[0].questions[0].correct = 2 },
    data => { data.exercises[0].questions[0].passage = 99 },
    data => { data.exercises[0].number = 4 },
    data => { delete data.exercises.find(exercise => exercise.number === 8).questions[2].occurrence },
  ]) {
    const broken = structuredClone(source)
    mutate(broken)
    assert.throws(() => compileExercises(broken))
  }
})
