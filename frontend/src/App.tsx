import { useEffect } from 'react'
import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { AppNavigation } from './components/AppNavigation'
import { HomePage } from './features/home/HomePage'
import { InsightsLayout } from './features/insights/InsightsLayout'
import { KanjiPage } from './features/kanji/KanjiPage'
import { VocabularyPage } from './features/vocabulary/VocabularyPage'
import { QuizPage } from './features/quiz/QuizPage'
import './App.css'

function App() {
  const { pathname, hash } = useLocation()

  useEffect(() => {
    const target = document.getElementById(hash ? hash.slice(1) : 'main-content')
    target?.focus({ preventScroll: true })
    if (hash) target?.scrollIntoView({ block: 'start' })
    else window.scrollTo(0, 0)
  }, [pathname, hash])

  return (
    <>
      <AppNavigation />
      <main className="app" id="main-content" tabIndex={-1}>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/quiz" element={<QuizPage />} />
          <Route path="/insights" element={<InsightsLayout />}>
            <Route index element={<Navigate to="kanji" replace />} />
            <Route path="kanji" element={<KanjiPage />} />
            <Route path="vocabulary" element={<VocabularyPage />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </>
  )
}

export default App
