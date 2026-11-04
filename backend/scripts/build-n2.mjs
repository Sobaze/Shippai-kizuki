// Development-only data build. Generated snapshots are committed and embedded by
// Go; running the app never contacts these services or regenerates questions.
import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = new URL('../internal/quiz/data/', import.meta.url)
const revision = 'b062d4e38c4bdd0950ae1d4ec55f04b176182e03'
const vocabularyURL = `https://raw.githubusercontent.com/stephenmk/yomitan-jlpt-vocab/${revision}/yomitan-jlpt-vocab/term_meta_bank_2.json`
const cacheDirectory = process.env.N2_BUILD_CACHE || join(tmpdir(), 'opencode', 'shippai-n2-source-cache')
const hash = value => createHash('sha256').update(value).digest('hex')
// Reviewed core glosses favour useful modern senses over incidental dictionary
// ordering (e.g. 第 is an ordinal marker, not primarily "residence").
export const coreGlosses = {
  '第': 'Ordinal number (No.)', '営': 'Conduct business; operate', '細': 'Detailed; slender', '革': 'Leather; reform',
  '介': 'Mediate; introduce', '歴': 'History; passage of time', '般': 'General; all',
  '幅': 'Width; breadth', '門': 'Gate', '額': 'Forehead; amount', '算': 'Calculate',
  '岸': 'Shore; bank', '児': 'Child', '換': 'Exchange; replace', '久': 'Long time',
  '陸': 'Land', '卒': 'Graduate', '測': 'Measure; estimate', '練': 'Practise; train',
  '軽': 'Light; slight', '裏': 'Back; reverse side', '臓': 'Internal organs',
  '匹': 'Counter for small animals', '床': 'Floor; bed', '柔': 'Soft; gentle',
  '殿': 'Hall; lord', '荷': 'Baggage; load', '兆': 'Trillion; omen', '召': 'Summon; call',
  '灰': 'Ashes', '冊': 'Counter for books', '軒': 'Counter for houses; eaves',
  '隻': 'Counter for ships', '枚': 'Counter for flat objects',
  '訓': 'Instruction; native Japanese reading', '令': 'Order; command', '署': 'Signature; government office',
}
const exampleOverrides = {
  '仏': ['仏教', 'ぶっきょう'], '府': ['政府', 'せいふ'], '導': ['指導', 'しどう'],
  '設': ['建設', 'けんせつ'], '専': ['専門', 'せんもん'], '歴': ['歴史', 'れきし'],
  '般': ['一般', 'いっぱん'], '伸': ['伸びる', 'のびる'], '卒': ['卒業', 'そつぎょう'],
  '麦': ['小麦', 'こむぎ'], '希': ['希望', 'きぼう'], '営': ['営業', 'えいぎょう'],
  '細': ['細かい', 'こまかい'], '柔': ['柔らかい', 'やわらかい'], '岸': ['海岸', 'かいがん'],
  '低': ['低い', 'ひくい'], '召': ['召す', 'めす'], '第': ['第一', 'だいいち'], '保': ['保護', 'ほご'],
}
const kana = value => /^[\p{Script=Hiragana}ー]+$/u.test(value)
const writtenWord = value => /\p{Script=Han}/u.test(value) && /^[\p{Script=Han}\p{Script=Hiragana}々ー]+$/u.test(value)
const text = html => html.replace(/<[^>]*>/g, ' ').replace(/&#x([0-9a-f]+);/gi, (_, n) => String.fromCodePoint(parseInt(n, 16)))
  .replace(/&#(\d+);/g, (_, n) => String.fromCodePoint(Number(n))).replace(/&amp;/g, '&').replace(/&nbsp;/g, ' ')
  .replace(/&quot;/g, '"').replace(/&#039;|&apos;/g, "'").replace(/\s+/g, ' ').trim()

export function kanjiFromHTML(html) {
  return html.split(/<tr\b[^>]*>/).filter(row => /jl-td-k\s/.test(row)).map(row => {
    const cells = row.split(/<td\b[^>]*>/)
    return text(cells[2]).trim()
  }).filter(value => /^\p{Script=Han}$/u.test(value))
}

export function phoneticKey(value) {
  return value.normalize('NFKC').replace(/ぢ/g, 'じ').replace(/づ/g, 'ず')
    .replace(/([おこそとのほもよろごぞどぼぽょ])う/g, '$1ー')
    .replace(/([えけせてねへめれげぜでべぺ])い/g, '$1ー')
}

export function distance(a, b) {
  let previous = Array.from({ length: b.length + 1 }, (_, i) => i)
  for (let i = 1; i <= a.length; i++) {
    const current = [i]
    for (let j = 1; j <= b.length; j++) current[j] = Math.min(current[j - 1] + 1, previous[j] + 1, previous[j - 1] + (a[i - 1] === b[j - 1] ? 0 : 1))
    previous = current
  }
  return previous[b.length]
}

export function nearbyReadings(reading) {
  const result = new Set()
  const groups = ['かが', 'きぎ', 'くぐ', 'けげ', 'こご', 'さざ', 'しじ', 'すず', 'せぜ', 'そぞ',
    'ただ', 'ちぢ', 'つづ', 'てで', 'とど', 'はばぱ', 'ひびぴ', 'ふぶぷ', 'へべぺ', 'ほぼぽ',
    'ゃや', 'ゅゆ', 'ょよ', 'っつ', 'あいうえお', 'かきくけこ', 'さしすせそ', 'たちつてと',
    'なにぬねの', 'はひふへほ', 'まみむめも', 'らりるれろ', 'がぎぐげご', 'ざじずぜぞ', 'ばびぶべぼ']
  for (let i = 0; i < reading.length; i++) {
    for (const group of groups.filter(group => group.includes(reading[i]))) {
      for (const other of group) if (other !== reading[i]) result.add(reading.slice(0, i) + other + reading.slice(i + 1))
    }
    if (reading.length > 2 && 'いうっ'.includes(reading[i])) result.add(reading.slice(0, i) + reading.slice(i + 1))
  }
  result.delete(reading)
  return [...result]
}

// Conservative synonym families supplement full dictionary-gloss token checks.
// Over-excluding a distractor is preferable to offering another valid meaning.
const families = [
  'protect guard defend defense defence resist shield safeguard',
  'region district area territory province prefecture jurisdiction boundary border frontier',
  'amount quantity measure volume number count total sum',
  'work labor labour job employment occupation business service',
  'change modify alter alteration reform revision revise amend transform replace replacement substitute exchange interchange switch swap',
  'group party faction clique association organization organisation society',
  'make build construct create form construction manufacture produce',
  'tell say speak state mention describe narrate report express explain',
  'lead guide guidance conduct direct direction command govern rule control manage instruction instruct',
  'keep retain preserve save maintain store conserve reserve',
  'join connect attach link tie bind unite combine fasten',
  'help aid assist support supplement supply compensate',
  'money wealth riches fortune currency cash',
  'decrease reduce diminish dwindle lessen shrink',
  'increase grow expand enlarge extend spread',
  'war army military soldier troops warrior arms fight battle',
  'place location spot position site situation',
  'finish end conclusion complete completion final ultimate',
  'start beginning origin source cause foundation base',
  'see look watch view observe inspect examine investigate survey',
  'knowledge learn study research education teach instruction',
  'bright light shine radiance luminous glow',
  'dark shadow shade gloomy darkness',
  'water liquid fluid wet moisture damp',
  'road street path way route track lane',
  'old ancient aged former past previous',
  'rest relax ease calm quiet peaceful peace comfort',
  'remove omit delete abolish eliminate exclude discard reject',
  'return restore revert recover repeat again',
  'skill art craft talent ability capable capacity competent',
  'judge judgment judgement decide decision settle determine establish',
  'body physique form shape figure appearance image pattern',
  'law regulation rule principle standard criterion norm',
  'feeling emotion heart mind thought think consider spirit soul',
  'carry transport convey send dispatch transmit deliver',
  'whole all entire total general universal complete',
  'child children juvenile infancy childhood baby infant newborn boy',
  'wood wooden lumber timber log plank board',
  'thin slender narrow taper dainty detailed precise',
  'soft gentle gentleness tender weakness softness frail weak',
].map(group => group.split(' '))
const stop = new Set('a an the of to in on by for and or with at from be as into its one that something'.split(' '))
export function meaningTokens(meanings) {
  const tokens = new Set(meanings.join(' ').toLowerCase().match(/[a-z]+/g)?.filter(token => !stop.has(token)) || [])
  for (const token of [...tokens]) {
    if (token.endsWith('s')) tokens.add(token.slice(0, -1))
    if (token.endsWith('ing')) { tokens.add(token.slice(0, -3)); tokens.add(token.slice(0, -3) + 'e') }
  }
  for (const family of families) if (family.some(token => tokens.has(token))) family.forEach(token => tokens.add(token))
  return tokens
}
export function disjointMeanings(first, second) {
  const a = meaningTokens(first), b = meaningTokens(second)
  return ![...a].some(token => b.has(token))
}

async function source(url) {
  const path = join(cacheDirectory, hash(url) + '.json')
  try { return JSON.parse(await readFile(path, 'utf8')).body } catch (error) { if (error.code !== 'ENOENT') throw error }
  for (let attempt = 0; attempt < 6; attempt++) {
    try {
      const response = await fetch(url, { signal: AbortSignal.timeout(30000) })
      if (!response.ok) throw new Error(`${response.status}: ${url}`)
      const body = await response.text()
      await writeFile(path, JSON.stringify({ url, body }))
      return body
    } catch (error) {
      if (attempt === 5) throw error
      await new Promise(resolve => setTimeout(resolve, 1000 * (attempt + 1)))
    }
  }
}

async function parallel(values, action) {
  let next = 0, completed = 0
  await Promise.all(Array.from({ length: 4 }, async () => {
    while (next < values.length) {
      const value = values[next++]
      await action(value)
      completed++
      if (completed % 100 === 0) console.log(`Verified ${completed}/${values.length} source requests`)
    }
  }))
}

function gloss(entry) { return entry.meanings[0]?.glosses.slice(0, 3).join('; ') || '' }
function common(variant) { return variant.priorities.some(priority => /^(ichi|news|spec|gai)1$|^nf\d+$/.test(priority)) ? 0 : variant.priorities.length ? 1 : 2 }
function frequency(variant) {
  const ranks = variant.priorities.filter(p => /^nf\d+$/.test(p)).map(p => Number(p.slice(2)))
  return ranks.length ? Math.min(...ranks) : 50
}
function catalogJSON(catalog) {
  const metadata = JSON.stringify({ version: catalog.version, checkedAt: catalog.checkedAt, sources: catalog.sources }, null, 2)
  return metadata.slice(0, -2) + ',\n  "pools": [\n' + catalog.pools.map(pool =>
    `    { "mode": "${pool.mode}", "level": "N2", "questions": [\n` + pool.questions.map(q => `      ${JSON.stringify(q)}`).join(',\n') + '\n    ] }').join(',\n') + '\n  ]\n}\n'
}

export async function build() {
  await mkdir(cacheDirectory, { recursive: true })
  const kanjiList = []
  for (let page = 1; page <= 4; page++) kanjiList.push(...kanjiFromHTML(await source(`https://jlptsensei.com/jlpt-n2-kanji-list/${page === 1 ? '' : `page/${page}/`}`)))
  assert.equal(kanjiList.length, 374, 'N2 source-list size changed; review before rebuilding')
  assert.equal(new Set(kanjiList).size, 374)
  const vocabulary = JSON.parse(await source(vocabularyURL))
  const base = JSON.parse(await readFile(new URL('questions.json', root), 'utf8'))
  const existingWords = new Set(base.pools.filter(pool => pool.mode === 'reading').flatMap(pool => pool.questions.map(q => q.prompt)))
  const candidates = vocabulary.filter(([word, , info]) => info.frequency.displayValue === 'N2' && writtenWord(word) && kana(info.reading) && !existingWords.has(word))
  const characters = [...new Set([...kanjiList, ...candidates.map(([word]) => [...word].find(c => /\p{Script=Han}/u.test(c)))])]
  const dictionaries = new Map(), kanji = new Map()
  console.log(`N2 sources: ${kanjiList.length} kanji; ${vocabulary.length} vocabulary records; ${candidates.length} eligible reading candidates; ${characters.length} dictionary characters`)
  await parallel(characters, async character => dictionaries.set(character, JSON.parse(await source(`https://kanjiapi.dev/v1/words/${encodeURIComponent(character)}`))))
  await parallel(kanjiList, async character => kanji.set(character, JSON.parse(await source(`https://kanjiapi.dev/v1/kanji/${encodeURIComponent(character)}`))))

  const words = new Map(), exclusions = []
  for (const [word, , info] of candidates) {
    const entries = dictionaries.get([...word].find(c => /\p{Script=Han}/u.test(c)))
    const entry = entries.find(entry => entry.variants.some(v => v.written === word && v.pronounced === info.reading) && gloss(entry))
    if (!entry) { exclusions.push({ word, reading: info.reading, reason: 'No exact written/reading dictionary match' }); continue }
    if (words.has(word)) continue // One stable question per written prompt.
    const acceptedReadings = [...new Set(entries.flatMap(e => e.variants.filter(v => v.written === word).map(v => v.pronounced)))]
    words.set(word, { written: word, reading: info.reading, meaning: gloss(entry), acceptedReadings })
  }
  const allReadings = [...new Set([...words.values()].map(word => word.reading))].sort()
  const readingQuestions = [...words.values()].map(word => {
    const accepted = new Set(word.acceptedReadings.map(phoneticKey))
    const nearby = new Set(nearbyReadings(word.reading))
    const options = [...new Set([...nearby, ...allReadings])].filter(reading => !accepted.has(phoneticKey(reading)))
      .map(reading => ({ reading, distance: distance(word.reading, reading), nearby: nearby.has(reading) }))
      .sort((a, b) => a.distance - b.distance || Number(b.nearby) - Number(a.nearby) || hash(word.written + a.reading).localeCompare(hash(word.written + b.reading)))
    const distractors = []
    for (const option of options) {
      if (distractors.some(other => phoneticKey(other) === phoneticKey(option.reading))) continue
      distractors.push(option.reading)
      if (distractors.length === 3) break
    }
    assert.equal(distractors.length, 3)
    return { id: `n2-reading-${hash(word.written).slice(0, 16)}`, mode: 'reading', prompt: word.written, answer: word.reading,
      choices: [word.reading, ...distractors], example: { written: word.written, reading: word.reading, meaning: word.meaning } }
  })

  const meaningEvidence = []
  const answerFor = character => {
    const value = coreGlosses[character] || kanji.get(character).meanings.slice(0, 2).join('; ')
    return value.charAt(0).toUpperCase() + value.slice(1)
  }
  const allMeanings = character => [...kanji.get(character).meanings, answerFor(character)]
  const meaningQuestions = kanjiList.map(character => {
    const meanings = kanji.get(character).meanings
    assert.ok(meanings.length)
    const answer = answerFor(character)
    const others = kanjiList.filter(other => other !== character && disjointMeanings(allMeanings(character), allMeanings(other)))
      .sort((a, b) => hash(character + a).localeCompare(hash(character + b)))
    const distractors = []
    for (const other of others) {
      const choice = answerFor(other)
      if (choice === answer || distractors.includes(choice)) continue
      distractors.push(choice)
      if (distractors.length === 3) break
    }
    assert.equal(distractors.length, 3)
    const examples = dictionaries.get(character).flatMap(entry => entry.variants.filter(v => v.written.includes(character) && writtenWord(v.written) && kana(v.pronounced) && gloss(entry))
      .map(variant => ({ variant, entry })))
    examples.sort((a, b) => common(a.variant) - common(b.variant) || frequency(a.variant) - frequency(b.variant)
      || Number(!words.has(a.variant.written)) - Number(!words.has(b.variant.written))
      || a.variant.written.length - b.variant.written.length || a.variant.written.localeCompare(b.variant.written, 'ja'))
    assert.ok(examples.length, `Missing example for ${character}`)
    const preferred = exampleOverrides[character]
    const selected = preferred ? examples.find(e => e.variant.written === preferred[0] && e.variant.pronounced === preferred[1]) : examples[0]
    assert.ok(selected, `Reviewed example unavailable for ${character}: ${preferred}`)
    const { variant, entry } = selected
    const example = { written: variant.written, reading: variant.pronounced, meaning: gloss(entry) }
    meaningEvidence.push({ kanji: character, meanings, answer, example })
    return { id: `n2-meaning-${character.codePointAt(0).toString(16)}`, mode: 'meaning', prompt: character, answer, choices: [answer, ...distractors], example }
  })
  const sources = {
    'N2-kanji': 'https://jlptsensei.com/jlpt-n2-kanji-list/ (all four pages; 374 N2-specific kanji)',
    'N2-vocabulary': `https://github.com/stephenmk/yomitan-jlpt-vocab/tree/${revision} (Stephen Kraus; Jonathan Waller/Tanos estimates; CC BY-SA 4.0)`,
    dictionary: 'https://kanjiapi.dev/ (KANJIDIC2 / JMdict, EDRDG, CC BY-SA 4.0)',
  }
  const catalog = { version: 1, checkedAt: '2026-11-03', sources, pools: [
    { mode: 'meaning', level: 'N2', questions: meaningQuestions },
    { mode: 'reading', level: 'N2', questions: readingQuestions },
  ] }
  const audit = { version: 1, checkedAt: catalog.checkedAt, vocabularyRevision: revision,
    sourceKanji: kanjiList.join(''), sourceVocabularyRecords: vocabulary.length,
    sourceVocabulary: vocabulary.map(([written, , info]) => ({ written, reading: info.reading, level: info.frequency.displayValue })),
    excludedNonKanjiOrUnsupportedOrExisting: vocabulary.length - candidates.length,
    exclusions, meaningEvidence, readingEvidence: [...words.values()],
  }
  await writeFile(new URL('n2.generated.json', root), catalogJSON(catalog))
  await writeFile(new URL('n2.audit.json', root), JSON.stringify(audit, null, 2) + '\n')
  console.log(`Built ${meaningQuestions.length} N2 meanings and ${readingQuestions.length} dictionary-matched N2 reading questions; ${exclusions.length} unmatched pairs deferred.`)
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await build()
