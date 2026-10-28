import { useId, useState } from 'react'
import type { KanjiEvidence } from './types'

type KanjiEvidenceTableProps = {
  kanji: string
  evidence: KanjiEvidence[]
  practiceReadings?: boolean
}

function PracticeWord({ word }: { word: KanjiEvidence }) {
  const [showReading, setShowReading] = useState(false)
  const readingId = useId()

  return (
    <div className="kanji-practice-word">
      <span className="kanji-word-reading" id={readingId}>
        {showReading && word.reading ? (
          <ruby>{word.vocabulary}<rp>(</rp><rt>{word.reading}</rt><rp>)</rp></ruby>
        ) : word.vocabulary}
      </span>
      {word.reading ? (
        <button
          type="button"
          className="kanji-furigana-toggle"
          lang="en"
          aria-expanded={showReading}
          aria-controls={readingId}
          aria-label={`${showReading ? 'Hide' : 'Show'} furigana for ${word.vocabulary}`}
          onClick={() => setShowReading((shown) => !shown)}
        >
          {showReading ? 'Hide furigana' : 'Show furigana'}
        </button>
      ) : <small lang="en">Reading not available</small>}
    </div>
  )
}

export function KanjiEvidenceTable({ kanji, evidence, practiceReadings = false }: KanjiEvidenceTableProps) {
  return (
    <div className="kanji-evidence-scroll">
      <table>
        <caption>Vocabulary evidence for <span lang="ja">{kanji}</span></caption>
        <thead>
          <tr>
            <th scope="col">Word</th>
            <th scope="col">Difficulty</th>
            <th scope="col">Reviews</th>
            <th scope="col">Use in score</th>
          </tr>
        </thead>
        <tbody>
          {evidence.map((word) => (
            <tr key={word.vocabulary}>
              <td lang="ja">{practiceReadings ? <PracticeWord word={word} /> : word.vocabulary}</td>
              <td>{word.difficulty}/100</td>
              <td>{word.reviewCount}</td>
              <td className={word.qualifies ? 'kanji-included' : 'kanji-early'}>
                {word.qualifies ? 'Included' : 'Early evidence · excluded'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
