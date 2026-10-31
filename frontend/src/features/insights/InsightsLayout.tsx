import { Link, NavLink, Outlet, useLocation } from 'react-router-dom'
import './insights.css'

export function InsightsLayout() {
  const { search } = useLocation()
  const tabParams = new URLSearchParams(search)
  tabParams.delete('page')
  const tabSearch = tabParams.toString() ? `?${tabParams}` : ''
  return (
    <>
      <header className="insights-header">
        <p className="eyebrow">From review history to focused practice</p>
        <h1>Your learning insights</h1>
        <p>Explore individual words or patterns across their kanji.</p>
        <Link className="insights-guide-link" to="/#scoring-guide">How scoring works →</Link>
      </header>
      <nav className="insights-tabs" aria-label="Insight views">
        <NavLink to={`/insights/kanji${tabSearch}`}>Kanji</NavLink>
        <NavLink to={`/insights/vocabulary${tabSearch}`}>Vocabulary</NavLink>
      </nav>
      <Outlet />
    </>
  )
}
