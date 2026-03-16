import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import './Auth.css'

export function Login() {
  const navigate = useNavigate()
  const { login, loginAsGuest } = useAuthStore()

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setIsLoading(true)

    try {
      await login({ username, password })
      navigate('/')
    } catch (err) {
      setError((err as Error).message || 'Login failed')
    } finally {
      setIsLoading(false)
    }
  }

  const handleGuestLogin = async () => {
    setIsLoading(true)
    try {
      await loginAsGuest()
      navigate('/')
    } catch (err) {
      setError((err as Error).message || 'Guest login failed')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-container">
        <h1>Login</h1>

        {error && <div className="auth-error">{error}</div>}

        <form onSubmit={handleSubmit} className="auth-form">
          <input
            type="text"
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="input"
            required
            autoComplete="username"
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="input"
            required
            autoComplete="current-password"
          />
          <button
            type="submit"
            className="btn btn-primary"
            disabled={isLoading}
          >
            {isLoading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <div className="auth-divider">
          <span>or</span>
        </div>

        <button
          className="btn btn-secondary"
          onClick={handleGuestLogin}
          disabled={isLoading}
        >
          Play as Guest
        </button>

        <p className="auth-link">
          Don't have an account? <Link to="/register">Register</Link>
        </p>

        <Link to="/" className="btn btn-text">
          ← Back to Home
        </Link>
      </div>
    </div>
  )
}
