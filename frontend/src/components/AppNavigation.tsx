import { Link, NavLink } from 'react-router-dom'
import './app-navigation.css'

export function AppNavigation() {
  return (
    <header className="app-navigation">
      <a className="skip-link" href="#main-content">Skip to content</a>
      <div className="app-navigation-inner">
        <Link to="/" className="app-brand" aria-label="Shippai Kizuki home">
          <span className="app-brand-mark" aria-hidden="true">気</span>
          <span>Shippai Kizuki<small lang="ja">失敗気づき</small></span>
        </Link>
        <nav className="app-navigation-links" aria-label="Main navigation">
          <NavLink to="/" end>Home</NavLink>
          <NavLink to="/insights">Insights</NavLink>
        </nav>
      </div>
    </header>
  )
}
