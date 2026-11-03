import { useEffect, useRef, useState } from 'react'
import { createCandidateQuiz, createCoverageQuiz, createExerciseQuiz, createMissedQuiz, QUIZ_LENGTH, recordQuizAnswer } from './quiz'
import { QuizQuestion } from './QuizQuestion'
import { QuizResults } from './QuizResults'
import { QuizAttemptSave } from './QuizAttemptSave'
import { QuizHistory } from './QuizHistory'
import { QuizPractice } from './QuizPractice'
import { QuizCoverage } from './QuizCoverage'
import { getQuizCoverage } from './api'
import { useQuizPool } from './useQuizPool'
import type { QuizCandidate, QuizLevel, QuizMode, QuizRun, QuizState, SavedQuizAttempt } from './types'
import './quiz.css'

export function QuizPage() {
  const [state, setState] = useState<QuizState>({ status: 'ready' })
  const [mode, setMode] = useState<QuizMode>('meaning')
  const [levels, setLevels] = useState<Record<QuizMode, QuizLevel>>({ meaning: 'N2', reading: 'N2', sentence: 'N2' })
  const [run, setRun] = useState<QuizRun | null>(null)
  const [reviewedAttempt, setReviewedAttempt] = useState<SavedQuizAttempt | null>(null)
  const introHeadingRef = useRef<HTMLHeadingElement>(null)
  const startRequest = useRef<AbortController | null>(null)
  const [preparingQuiz, setPreparingQuiz] = useState(false)
  const [startError, setStartError] = useState<string | null>(null)
  const isReading = mode === 'reading'
  const isSentence = mode === 'sentence'
  const exerciseId = isSentence ? 'exercise-4' : undefined
  const selectedLevel = levels[mode]
  const { state: poolState, reload: reloadPool } = useQuizPool(mode, selectedLevel, exerciseId)

  useEffect(() => () => startRequest.current?.abort(), [])

  function cancelPreparation() {
    startRequest.current?.abort()
    startRequest.current = null
    setPreparingQuiz(false)
    setStartError(null)
  }

  async function startQuiz() {
    if (poolState?.status !== 'ready' || startRequest.current) return
    const controller = new AbortController()
    startRequest.current = controller
    setPreparingQuiz(true)
    setStartError(null)
    try {
      // Always refresh before selecting, including a fast restart while the
      // previous completed run is saving. The results view stays mounted.
      const coverage = await getQuizCoverage(mode, selectedLevel, controller.signal, exerciseId)
      if (controller.signal.aborted) return
      const questions = isSentence ? createExerciseQuiz(poolState.pool.questions) : createCoverageQuiz(poolState.pool.questions, coverage)
      setRun({ id: crypto.randomUUID(), mode, kind: 'full', startedAt: new Date().toISOString() })
      setReviewedAttempt(null)
      setState({ status: 'playing', questions, answers: [] })
    } catch (error: unknown) {
      if (!controller.signal.aborted) setStartError(error instanceof Error ? error.message : 'Could not prepare this quiz. Try again.')
    } finally {
      if (startRequest.current === controller) {
        startRequest.current = null
        setPreparingQuiz(false)
      }
    }
  }

  function retryMissed() {
    if (state.status !== 'results') return
    cancelPreparation()
    const questions = createMissedQuiz(state.result)
    if (questions.length === 0) return
    const nextMode = questions[0].mode
    const parentAttemptId = reviewedAttempt?.id ?? run?.id
    setMode(nextMode)
    setRun({ id: crypto.randomUUID(), mode: nextMode, kind: 'retry', startedAt: new Date().toISOString(), parentAttemptId })
    setReviewedAttempt(null)
    setState({ status: 'playing', questions, answers: [] })
  }

  function startCandidateQuiz(candidates: QuizCandidate[]) {
    cancelPreparation()
    const questions = createCandidateQuiz(candidates)
    if (questions.length === 0) return
    setRun({ id: crypto.randomUUID(), mode, kind: 'retry', startedAt: new Date().toISOString() })
    setReviewedAttempt(null)
    setState({ status: 'playing', questions, answers: [] })
  }

  function reviewSaved(attempt: SavedQuizAttempt) {
    cancelPreparation()
    setMode(attempt.mode)
    const level = attempt.result.items[0]?.question.estimatedLevel
    if (level && attempt.result.items.every((item) => item.question.estimatedLevel === level)) {
      setLevels((current) => ({ ...current, [attempt.mode]: level }))
    }
    setRun(null)
    setReviewedAttempt(attempt)
    setState({ status: 'results', result: attempt.result })
  }

  function returnToSetup() {
    cancelPreparation()
    setState({ status: 'ready' })
    setRun(null)
    setReviewedAttempt(null)
    requestAnimationFrame(() => {
      introHeadingRef.current?.focus({ preventScroll: true })
      introHeadingRef.current?.scrollIntoView({ block: 'start' })
    })
  }

  return (
    <>
      <header className="quiz-page-header">
        <p className="eyebrow">Learn from what you miss</p>
        <h1>Japanese practice quiz</h1>
        <p>Connect written forms, readings, and meaning—one short attempt at a time.</p>
      </header>

      {state.status === 'ready' && (
        <section className="quiz-card quiz-intro" aria-labelledby="quiz-intro-heading">
          <p className="quiz-level">Estimated {selectedLevel} · {isSentence ? 'JLPT Kanji exercise' : isReading ? 'Word-reading practice' : 'Meaning practice'}</p>
          <h2 id="quiz-intro-heading" ref={introHeadingRef} tabIndex={-1}>{isSentence ? 'Exercise 4' : isReading ? 'Read kanji in everyday words' : 'A short quiz, one kanji at a time'}</h2>
          <fieldset className="quiz-mode-fieldset" disabled={preparingQuiz}>
            <legend>Choose your practice</legend>
            <div className="quiz-modes">
              <label className={`quiz-mode-option${mode === 'meaning' ? ' is-selected' : ''}`}>
                <input type="radio" name="quiz-mode" value="meaning" checked={mode === 'meaning'} onChange={() => setMode('meaning')} />
                <span><strong>Kanji meanings</strong><small>Recognize a core meaning in English.</small></span>
              </label>
              <label className={`quiz-mode-option${isReading ? ' is-selected' : ''}`}>
                <input type="radio" name="quiz-mode" value="reading" checked={isReading} onChange={() => setMode('reading')} />
                <span><strong>Word readings</strong><small>Read a complete word in kana.</small></span>
              </label>
              <label className={`quiz-mode-option${isSentence ? ' is-selected' : ''}`}>
                <input type="radio" name="quiz-mode" value="sentence" checked={isSentence} onChange={() => setMode('sentence')} />
                <span><strong>JLPT Kanji exercise</strong><small>Readings and spellings in sentences.</small></span>
              </label>
            </div>
          </fieldset>
          {!isSentence && <fieldset className="quiz-mode-fieldset" disabled={preparingQuiz}>
            <legend>Estimated {isReading ? 'vocabulary' : 'kanji'} level</legend>
            <div className="quiz-modes">
              {(['N2', 'N4', 'N5'] as const).map((level) => (
                <label key={level} className={`quiz-mode-option${selectedLevel === level ? ' is-selected' : ''}`}>
                  <input type="radio" name="quiz-level" value={level} checked={selectedLevel === level} onChange={() => setLevels((current) => ({ ...current, [mode]: level }))} />
                  <span><strong>{level}</strong><small>{level === 'N2' ? 'Exam focus' : level === 'N5' ? 'Foundational' : 'Everyday'} {isReading ? 'words' : 'kanji'}</small></span>
                </label>
              ))}
            </div>
            <p className="quiz-note">Each pool is level-specific, not cumulative. N2 covers 374 N2-specific kanji and a dictionary-matched N2 vocabulary list; prerequisite kanji and vocabulary are also expected for the exam.</p>
            {isReading && <p className="quiz-note">These are vocabulary-level estimates, checked separately from the levels of the kanji in each word.</p>}
          </fieldset>}
          {isSentence ? <p>Exercise 4 contains 12 fixed questions: six reading questions and six spelling questions. Sentences stay in the exercise’s original order; answer choices are shuffled. Look at the highlighted, underlined target in each sentence.</p> : <p>Answer {QUIZ_LENGTH} questions with four choices each. Unseen items come first, then those with fewer saved attempts; equally practised items favour the oldest practice. Ties, question order, and answer positions are shuffled.</p>}
          {isSentence && <p className="quiz-note">Adapted from the exercise you supplied, published by JapaneseTest4You on March 14, 2015. This is JLPT-inspired practice; the publication date does not identify an official past exam.</p>}
          {isReading && <p className="quiz-note">Practise complete words. Accepted alternative readings are excluded from wrong choices. Word meanings and readings appear in the results, after you have tried recalling them.</p>}
          <dl className="quiz-intro-facts">
            <div><dt>Questions</dt><dd>{isSentence ? '12 sentence targets' : `${QUIZ_LENGTH} distinct ${isReading ? 'words' : 'kanji'}`}</dd></div>
            <div><dt>Starting pool</dt><dd>{poolState?.status === 'ready' ? `${poolState.pool.questions.length} checked entries` : '—'}</dd></div>
            <div><dt>Pace</dt><dd>Untimed</dd></div>
          </dl>
          {poolState?.status === 'ready' && <QuizCoverage key={`${mode}/${selectedLevel}/${exerciseId ?? ''}`} mode={mode} level={selectedLevel} exerciseId={exerciseId} />}
          <p className="quiz-note">Choose one answer before moving on. Your score, correct answers, and word feedback appear after {isSentence ? 'question 12' : 'question 10'}.</p>
          {!poolState && <p role="status">Loading checked questions…</p>}
          {poolState?.status === 'error' && (
            <>
              <p className="quiz-error" role="alert">{poolState.message}</p>
              <button type="button" className="quiz-secondary" onClick={reloadPool}>Retry loading questions</button>
            </>
          )}
          <button type="button" className="quiz-primary" disabled={poolState?.status !== 'ready' || preparingQuiz} onClick={startQuiz}>{preparingQuiz ? 'Preparing…' : isSentence ? 'Start Exercise 4' : 'Start 10-question quiz'}</button>
        </section>
      )}

      {state.status === 'playing' && (
        <QuizQuestion
          key={state.questions[state.answers.length].id}
          question={state.questions[state.answers.length]}
          kind={run?.kind ?? 'full'}
          position={state.answers.length + 1}
          total={state.questions.length}
          onAnswer={(answer) => setState((current) => recordQuizAnswer(current, answer))}
          onCancel={returnToSetup}
        />
      )}
      {state.status === 'results' && (
        <>
          {run && <QuizAttemptSave key={run.id} run={run} result={state.result} />}
          {reviewedAttempt && <p className="quiz-saved-view">Reviewing saved {reviewedAttempt.kind === 'retry' ? 'retry' : 'quiz'} · {new Date(reviewedAttempt.completedAt).toLocaleString()} · original question snapshot</p>}
          <QuizResults
            result={state.result}
            kind={run?.kind ?? reviewedAttempt?.kind ?? 'full'}
            restartLabel={isSentence ? 'Restart Exercise 4' : `Start another 10-question ${selectedLevel}${isReading ? ' reading' : ''} quiz`}
            canRestart={poolState?.status === 'ready' && !preparingQuiz}
            onRestart={startQuiz}
            onChooseMode={returnToSetup}
            onRetryMissed={retryMissed}
          />
          {poolState?.status === 'error' && <p className="quiz-error" role="alert">{poolState.message} Return to setup to retry loading.</p>}
        </>
      )}

      {preparingQuiz && (
        <p className="quiz-preparation" role="status">
          Checking saved practice to choose your next questions…{' '}
          <button type="button" className="quiz-secondary" onClick={cancelPreparation}>Cancel preparation</button>
        </p>
      )}
      {startError && <p className="quiz-error quiz-preparation" role="alert">{startError} Try starting again after the issue is resolved.</p>}

      {state.status === 'ready' && (
        <>
          <QuizPractice key={`practice-${mode}/${selectedLevel}`} mode={mode} level={selectedLevel} poolQuestions={poolState?.status === 'ready' ? poolState.pool.questions : []} onStart={startCandidateQuiz} />
          <QuizHistory key={mode} mode={mode} onReview={reviewSaved} />
        </>
      )}

      <p className="quiz-session-note">Completed attempts are saved through the Go backend; the results screen shows their save status. Cancelling discards the unfinished run. Quiz scores stay separate from Anki-based difficulty scores.</p>
      <details className="quiz-sources">
        <summary>Question sources and level estimate</summary>
        <p>JLPT Kanji exercise · Exercise 4 uses the sentences, choices, and answer key supplied by you from <a href="https://japanesetest4you.com/japanese-language-proficiency-test-jlpt-n2-kanji-exercise-4/" target="_blank" rel="noreferrer">JapaneseTest4You’s Exercise 04</a>, with added target readings/meanings for feedback. Its N2 label describes the exercise, not an independent vocabulary classification for every target.</p>
        <p>
          This is a small authored practice pool, not an official or complete JLPT list.
          The meaning pools’ kanji level estimates were checked against JLPT Sensei’s <a href="https://jlptsensei.com/jlpt-n5-kanji-list/" target="_blank" rel="noreferrer">N5</a> and <a href="https://jlptsensei.com/jlpt-n4-kanji-list/" target="_blank" rel="noreferrer">N4</a> study lists.
          {' '}The <a href="https://jlptsensei.com/jlpt-n2-kanji-list/" target="_blank" rel="noreferrer">N2 kanji list</a> supplies the 374 N2-specific targets.
          Word-reading levels were checked separately against their <a href="https://jlptsensei.com/jlpt-n5-vocabulary-list/" target="_blank" rel="noreferrer">N5 vocabulary</a> and <a href="https://jlptsensei.com/jlpt-n4-vocabulary-list/" target="_blank" rel="noreferrer">N4 vocabulary</a> lists.
          {' '}N2 vocabulary classifications come from <a href="https://github.com/stephenmk/yomitan-jlpt-vocab" target="_blank" rel="noreferrer">Stephen Kraus’s corrected Jonathan Waller/Tanos list</a> (CC BY-SA 4.0); exact written forms/readings are checked against JMdict. Kana-only, unsupported, and unmatched entries are excluded.
          The <a href="https://www.jlpt.jp/e/faq/#anchor28" target="_blank" rel="noreferrer">official JLPT FAQ</a> explains why test-content lists are no longer published.
        </p>
        <p>
          Core kanji meanings and whole-word readings/glosses were checked using <a href="https://kanjiapi.dev/" target="_blank" rel="noreferrer">kanjiapi.dev</a>,
          based on <a href="https://www.edrdg.org/wiki/KANJIDIC_Project.html" target="_blank" rel="noreferrer">KANJIDIC2</a>
          {' '}and <a href="https://www.edrdg.org/wiki/JMdict-EDICT_Dictionary_Project.html" target="_blank" rel="noreferrer">JMdict / EDRDG</a>
          {' '}(<a href="https://www.edrdg.org/edrdg/licence.html" target="_blank" rel="noreferrer">CC BY-SA 4.0</a>).
          A kanji may have additional meanings depending on context.
        </p>
      </details>
    </>
  )
}
