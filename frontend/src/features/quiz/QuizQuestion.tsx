import { useEffect, useRef, useState } from 'react'
import type { QuizAnswer, QuizAttemptKind, QuizQuestion as Question } from './types'

type QuizQuestionProps = {
  question: Question
  kind: QuizAttemptKind
  position: number
  total: number
  onAnswer: (answer: QuizAnswer) => void
  onCancel: () => void
}

export function QuizQuestion({ question, kind, position, total, onAnswer, onCancel }: QuizQuestionProps) {
  const [selectedAnswer, setSelectedAnswer] = useState<string | null>(null)
  const isSentence = question.mode === 'sentence'
  const isReading = question.mode === 'reading' || (isSentence && question.task === 'reading')
  const sectionRef = useRef<HTMLElement>(null)

  useEffect(() => {
    sectionRef.current?.focus({ preventScroll: true })
    sectionRef.current?.scrollIntoView({ block: 'start' })
  }, [])

  return (
    <section className="quiz-card quiz-question" ref={sectionRef} tabIndex={-1} aria-labelledby="quiz-question-heading">
      <div className="quiz-progress-header">
        <div className="quiz-progress-count">
          <p>{kind === 'retry' ? 'Retry question' : 'Question'} {position} of {total}</p>
          <span>{position - 1} answered</span>
        </div>
        <button type="button" className="quiz-cancel" onClick={onCancel}>Cancel {isSentence ? 'exercise' : 'quiz'}</button>
      </div>
      <progress className="quiz-progress" value={position - 1} max={total} aria-label="Answered questions" />
      <h2 id="quiz-question-heading">
        <span className="quiz-prompt">{isSentence ? (isReading ? 'How is the underlined expression read?' : 'Which spelling fits the underlined expression?') : isReading ? 'How is this word read?' : 'Which is a meaning of this kanji?'}</span>
        {isSentence && question.context ? (
          <span className="quiz-sentence" lang="ja">{question.context.before}<mark className="quiz-target"><u>{question.prompt}</u></mark>{question.context.after}</span>
        ) : <span className={isReading ? 'quiz-word' : 'quiz-glyph'} lang="ja">{question.prompt}</span>}
      </h2>

      <form onSubmit={(event) => {
        event.preventDefault()
        if (selectedAnswer !== null) onAnswer({ questionId: question.id, selectedAnswer })
      }}>
        <fieldset className="quiz-choice-fieldset">
          <legend>{isReading ? 'Select one reading' : isSentence ? 'Select one spelling' : 'Select one meaning'}</legend>
          <div className="quiz-choices">
            {question.choices.map((choice) => (
              <label className={`quiz-choice${selectedAnswer === choice ? ' is-selected' : ''}`} key={choice}>
                <input
                  type="radio"
                  name="answer"
                  value={choice}
                  checked={selectedAnswer === choice}
                  onChange={() => setSelectedAnswer(choice)}
                />
                <span lang={isReading || isSentence ? 'ja' : undefined}>{choice}</span>
              </label>
            ))}
          </div>
        </fieldset>
        <div className="quiz-question-actions">
          <p>Answers are reviewed after the final question.</p>
          <button className="quiz-primary" type="submit" disabled={selectedAnswer === null}>
            {position === total ? (kind === 'retry' ? 'Finish retry' : isSentence ? 'Finish exercise' : 'Finish quiz') : 'Next question →'}
          </button>
        </div>
      </form>
    </section>
  )
}
