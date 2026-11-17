// Offline content build: expand the user's shared passages into the same
// immutable per-target sentence snapshots used by Exercise 4.
import assert from 'node:assert/strict'
import { readFile, writeFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'

const directory = new URL('../internal/quiz/data/', import.meta.url)

export function compileExercises(source) {
  assert.equal(source.version, 1)
  assert.ok(source.source && source.suppliedAt && Array.isArray(source.exercises))
  const seen = new Set()
  const sources = {}
  const pools = source.exercises.map(exercise => {
    assert.ok(Number.isInteger(exercise.number) && exercise.number > 0 && exercise.number !== 4 && !seen.has(exercise.number), 'Exercise numbers must be unique; Exercise 4 is retained separately')
    seen.add(exercise.number)
    const exerciseId = `exercise-${exercise.number}`
    sources[exerciseId] = `${source.source}; JLPT N2 Kanji Exercise ${String(exercise.number).padStart(2, '0')}`
    assert.ok(exercise.questions.length >= 1 && exercise.questions.length <= 12)
    const questions = exercise.questions.map((question, index) => {
      const label = `${exerciseId}, question ${index + 1}`
      assert.equal(question.number, index + 1, `${label}: source numbering must be sequential`)
      const passage = exercise.passages[question.passage]
      assert.ok(typeof passage === 'string' && question.target, `${label}: missing passage/target`)
      assert.ok(question.task === 'reading' || question.task === 'spelling', `${label}: unknown task`)
      const occurrence = question.occurrence ?? 1
      assert.ok(Number.isInteger(occurrence) && occurrence > 0)
      if (passage.split(question.target).length > 2) assert.ok(question.occurrence, `${label}: choose the target occurrence explicitly`)
      let position = -1
      for (let i = 0; i < occurrence; i++) {
        position = passage.indexOf(question.target, position < 0 ? 0 : position + question.target.length)
        assert.ok(position >= 0, `${label}: target occurrence not found`)
      }
      const choices = question.choices.map(choice => choice.trim())
      assert.equal(choices.length, 4, label)
      assert.equal(new Set(choices).size, 4, label)
      assert.ok(choices.every(Boolean), label)
      assert.ok(Number.isInteger(question.correct) && question.correct >= 1 && question.correct <= 4, label)
      const answer = choices[question.correct - 1]
      assert.match(question.reading, /^[\p{Script=Hiragana}ー]+$/u, label)
      assert.ok(question.written && question.meaning, label)
      assert.equal(question.target, question.task === 'reading' ? question.written : question.reading, `${label}: target/word mismatch`)
      assert.equal(answer, question.task === 'reading' ? question.reading : question.written, `${label}: source key/feedback mismatch`)
      return {
        id: `${exerciseId}-${question.task}-${question.number}`, mode: 'sentence', exerciseId, task: question.task,
        prompt: question.target,
        context: { before: passage.slice(0, position), after: passage.slice(position + question.target.length) },
        answer, choices, example: { written: question.written, reading: question.reading, meaning: question.meaning },
      }
    })
    return { mode: 'sentence', level: 'N2', exerciseId, title: `Exercise ${exercise.number}`, questions }
  })
  return { version: 1, checkedAt: source.suppliedAt, sources, pools }
}

function formatCatalog(catalog) {
  const metadata = JSON.stringify({ version: catalog.version, checkedAt: catalog.checkedAt, sources: catalog.sources }, null, 2)
  return metadata.slice(0, -2) + ',\n  "pools": [\n' + catalog.pools.map(pool =>
    `    { "mode": "sentence", "level": "N2", "exerciseId": "${pool.exerciseId}", "title": "${pool.title}", "questions": [\n` +
    pool.questions.map(question => `      ${JSON.stringify(question)}`).join(',\n') + '\n    ] }').join(',\n') + '\n  ]\n}\n'
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const source = JSON.parse(await readFile(new URL('exercises.supplied.json', directory), 'utf8'))
  const catalog = compileExercises(source)
  await writeFile(new URL('exercises.additional.generated.json', directory), formatCatalog(catalog))
  console.log(`Built ${catalog.pools.length} additional exercises, ${catalog.pools.reduce((count, pool) => count + pool.questions.length, 0)} targets; original Exercise 4 retained separately.`)
}
