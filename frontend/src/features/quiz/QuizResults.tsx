import { useEffect, useRef } from 'react'
import { describeQuizQuestions } from './quiz'
import type { QuizAttemptKind, QuizResult } from './types'

type QuizResultsProps = {
  result: QuizResult
  kind: QuizAttemptKind
  restartLabel: string
  canRestart: boolean
  onRestart: () => void
  onChooseMode: () => void
  onRetryMissed: () => void
}

export function QuizResults({ result, kind, restartLabel, canRestart, onRestart, onChooseMode, onRetryMissed }: QuizResultsProps) {
  const headingRef = useRef<HTMLHeadingElement>(null)
  const percentage = result.total === 0 ? 0 : Math.round(100 * result.correctCount / result.total)
  const isReading = result.items[0]?.question.mode === 'reading'
  const isSentence = result.items[0]?.question.mode === 'sentence'
  const missedCount = result.items.filter((item) => !item.correct).length

  useEffect(() => {
    headingRef.current?.focus({ preventScroll: true })
    headingRef.current?.scrollIntoView({ block: 'start' })
  }, [])

  return (
    <section className="quiz-results" aria-labelledby="quiz-results-heading">
      <div className="quiz-card quiz-result-summary">
        <div>
          <p className="eyebrow">{kind === 'retry' ? 'Focused retry complete' : 'Quiz complete'} · {describeQuizQuestions(result.items.map((item) => item.question))}</p>
          <h2 id="quiz-results-heading" ref={headingRef} tabIndex={-1}>Your results</h2>
          <p className="quiz-score"><strong>{result.correctCount} / {result.total}</strong> correct · {percentage}%</p>
          {isSentence && <dl className="quiz-task-scores">{(['reading', 'spelling'] as const).map(task => {
            const items = result.items.filter(item => item.question.task === task)
            return items.length > 0 ? <div key={task}><dt>{task === 'reading' ? 'Readings' : 'Spellings'}</dt><dd>{items.filter(item => item.correct).length} / {items.length}</dd></div> : null
          })}</dl>}
          <p className="quiz-note">Review each answer below. This is a practice score, not a JLPT score or a mastery rating.</p>
        </div>
        <div className="quiz-result-actions">
          {missedCount > 0 && <button type="button" className="quiz-primary quiz-retry-missed" onClick={onRetryMissed}>Retry {missedCount} missed {missedCount === 1 ? 'question' : 'questions'}</button>}
          <button type="button" className={`${missedCount > 0 ? 'quiz-secondary' : 'quiz-primary'} quiz-new-attempt`} disabled={!canRestart} onClick={onRestart}>{restartLabel}</button>
          <button type="button" className="quiz-secondary" onClick={onChooseMode}>Change quiz type</button>
        </div>
      </div>

      <ol className="quiz-result-list">
        {result.items.map((item, index) => (
          <li className={`quiz-card quiz-result-item ${item.correct ? 'is-correct' : 'is-incorrect'}`} key={item.question.id}>
            <div className="quiz-result-item-header">
              <span className={isReading || isSentence ? 'quiz-result-word' : 'quiz-result-kanji'} lang="ja">{item.question.prompt}</span>
              <span className="quiz-result-status">Question {index + 1} · {isSentence ? `${item.question.task === 'reading' ? 'Reading' : 'Spelling'} · ` : item.question.estimatedLevel && `Estimated ${item.question.estimatedLevel} · `}{item.correct ? 'Correct' : 'Incorrect'}</span>
            </div>
            {isSentence && item.question.context && <p className="quiz-sentence quiz-sentence-review" lang="ja">{item.question.context.before}<mark className="quiz-target"><u>{item.question.example.written}</u></mark>{item.question.context.after}</p>}
            <dl>
              <div><dt>Your answer</dt><dd lang={isReading || isSentence ? 'ja' : undefined}>{item.selectedAnswer ?? 'Not answered'}</dd></div>
              <div><dt>{isReading || item.question.task === 'reading' ? 'Correct reading' : isSentence ? 'Correct spelling' : 'Correct meaning'}</dt><dd lang={isReading || isSentence ? 'ja' : undefined}>{item.question.answer}</dd></div>
            </dl>
            <div className="quiz-word-feedback">
              <h3>{isSentence ? 'Target expression' : isReading ? 'Word meaning' : 'Example word'}</h3>
              {!isReading && (
                <p className="quiz-example-word" lang="ja">
                  {item.question.example.written}
                  <span className="quiz-example-reading">（{item.question.example.reading}）</span>
                </p>
              )}
              <p className="quiz-example-meaning">{item.question.example.meaning}</p>
            </div>
          </li>
        ))}
      </ol>
    </section>
  )
}
