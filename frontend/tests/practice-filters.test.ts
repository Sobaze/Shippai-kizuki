import assert from 'node:assert/strict'
import test from 'node:test'
import { matchesPracticeFilter, parsePracticeFilter } from '../src/features/insights/practiceFilter.ts'

test('URL filters accept supported options and safely default to All', () => {
  for (const value of ['all', 'practice', 'limited', 'no-extra'] as const) {
    assert.equal(parsePracticeFilter(value), value)
  }
  assert.equal(parsePracticeFilter(null), 'all')
  assert.equal(parsePracticeFilter('unknown'), 'all')
})

test('practice suggestions exclude high-scoring but limited-evidence items', () => {
  assert.equal(matchesPracticeFilter('suggested_practice', 'practice'), true)
  for (const status of ['limited_evidence', 'limited_review_evidence', 'limited_cross_word_evidence', 'no_extra_practice']) {
    assert.equal(matchesPracticeFilter(status, 'practice'), false)
  }
})

test('limited evidence includes vocabulary and both kanji evidence limitations', () => {
  for (const status of ['limited_evidence', 'limited_review_evidence', 'limited_cross_word_evidence']) {
    assert.equal(matchesPracticeFilter(status, 'limited'), true)
    assert.equal(matchesPracticeFilter(status, 'all'), true)
    assert.equal(matchesPracticeFilter(status, 'no-extra'), false)
  }
  assert.equal(matchesPracticeFilter('suggested_practice', 'limited'), false)
  assert.equal(matchesPracticeFilter('no_extra_practice', 'limited'), false)
  assert.equal(matchesPracticeFilter('no_extra_practice', 'no-extra'), true)
})

test('filtering precedes the display limit so later matches remain visible', () => {
  const entries = [
    ...Array.from({ length: 24 }, () => ({ recommendation: 'limited_cross_word_evidence' })),
    { recommendation: 'suggested_practice' },
  ]
  const visible = entries.filter((entry) => matchesPracticeFilter(entry.recommendation, 'practice')).slice(0, 24)
  assert.equal(visible.length, 1)
  assert.equal(visible[0].recommendation, 'suggested_practice')
})
