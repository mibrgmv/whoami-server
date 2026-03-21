import { useEffect } from 'react'
import { redirectToRegister } from '../stores/authStore'
import './Auth.css'

export function Register() {
  useEffect(() => {
    redirectToRegister()
  }, [])

  return (
    <div className="auth-page">
      <div className="auth-container">
        <p className="auth-message">Redirecting to registration...</p>
      </div>
    </div>
  )
}
