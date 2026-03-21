import { useEffect } from 'react'
import { redirectToLogin } from '../stores/authStore'
import './Auth.css'

export function Login() {
  useEffect(() => {
    redirectToLogin()
  }, [])

  return (
    <div className="auth-page">
      <div className="auth-container">
        <p className="auth-message">Redirecting to login...</p>
      </div>
    </div>
  )
}
