import assert from 'node:assert/strict'
import test from 'node:test'
import { getPageNumbers, getPageWindow } from '../src/features/insights/pageWindow.ts'

test('pages produce exact, non-overlapping slices including a partial final page', () => {
  const data = Array.from({ length: 53 }, (_, index) => index)
  const pages = ['1', '2', '3'].map((page) => getPageWindow(page, data.length, 24))
  assert.deepEqual(pages.map(({ from, to }) => [from, to]), [[1, 24], [25, 48], [49, 53]])
  const combined = pages.flatMap((page) => data.slice(page.startIndex, page.endIndex))
  assert.deepEqual(combined, data)
  assert.ok(pages.every((page) => page.totalPages === 3))
})

test('invalid URL pages default to the first page; out-of-range pages clamp to the last', () => {
  for (const raw of [null, '0', '-1', '1.5', 'abc', '01', 'Infinity', '999999999999999999999']) {
    assert.equal(getPageWindow(raw, 1080, 24).page, 1)
  }
  assert.equal(getPageWindow('999', 1080, 24).page, 45)
  assert.equal(getPageWindow('2', 1741, 50).from, 51)
  assert.equal(getPageWindow('35', 1741, 50).to, 1741)
})

test('empty filters produce an empty slice and a safe first-page state', () => {
  const page = getPageWindow('5', 0, 24)
  assert.equal(page.page, 1)
  assert.equal(page.totalPages, 1)
  assert.equal(page.startIndex, 0)
  assert.equal(page.endIndex, 0)
  assert.equal(page.from, 0)
  assert.equal(page.to, 0)
})

test('filtered matches are paginated independently of their original positions', () => {
  const data = Array.from({ length: 200 }, (_, index) => ({ index, matches: index % 3 === 0 }))
  const filtered = data.filter((entry) => entry.matches)
  const page = getPageWindow('2', filtered.length, 24)
  assert.equal(filtered.slice(page.startIndex, page.endIndex)[0].index, 72)
  assert.equal(page.totalPages, 3)
})

test('page numbers retain first/current/last pages and collapse long gaps', () => {
  assert.deepEqual(getPageNumbers(1, 3), [1, 2, 3])
  assert.deepEqual(getPageNumbers(1, 45), [1, 2, 3, 4, null, 45])
  assert.deepEqual(getPageNumbers(23, 45), [1, null, 22, 23, 24, null, 45])
  assert.deepEqual(getPageNumbers(45, 45), [1, null, 42, 43, 44, 45])
  for (let page = 1; page <= 45; page++) {
    const links = getPageNumbers(page, 45)
    const numbers = links.filter((value) => value !== null)
    assert.ok(links.length <= 7)
    assert.ok(numbers.includes(1) && numbers.includes(page) && numbers.includes(45))
    assert.equal(new Set(numbers).size, numbers.length)
  }
})
