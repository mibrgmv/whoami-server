import { useEffect } from 'react'
import { config } from '../config'
import './Auth.css'

export function Register() {
  useEffect(() => {
    window.location.href = `${config.apiBase}/auth/register`
  }, [])

  return (
    <div className="auth-page">
      <div className="auth-container">
        <p className="auth-message">Redirecting to registration...</p>
      </div>
    </div>
  )
}
