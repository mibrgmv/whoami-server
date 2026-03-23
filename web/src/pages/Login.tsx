import { useEffect } from 'react'
import { config } from '../config'
import './Auth.css'

export function Login() {
  useEffect(() => {
    window.location.href = `${config.apiBase}/auth/login`
  }, [])

  return (
    <div className="auth-page">
      <div className="auth-container">
        <p className="auth-message">Redirecting to login...</p>
      </div>
    </div>
  )
}
