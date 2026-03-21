import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuthStore, redirectToLogin } from '../stores/authStore'
import { auth } from '../api/client'
import './Auth.css'

export function OAuthCallback() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setTokens } = useAuthStore()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const code = searchParams.get('code')
    const errorParam = searchParams.get('error')

    if (errorParam) {
      if (searchParams.get('error_description') === 'authentication_expired') {
        navigate('/')
        return
      }
      setError(searchParams.get('error_description') || 'Authentication failed')
      return
    }

    if (!code) {
      setError('No authorization code received')
      return
    }

    auth.exchangeCode(code)
      .then((tokens) => {
        setTokens(tokens.access_token, tokens.refresh_token, false)
        const returnTo = sessionStorage.getItem('auth_return_to')
        sessionStorage.removeItem('auth_return_to')
        navigate(returnTo || '/')
      })
      .catch((err) => {
        setError((err as Error).message || 'Failed to complete sign-in')
      })
  }, [searchParams, setTokens, navigate])

  if (error) {
    return (
      <div className="auth-page">
        <div className="auth-container">
          <h1>Sign-in failed</h1>
          <div className="auth-error">{error}</div>
          <button className="btn btn-primary" onClick={() => redirectToLogin()}>
            Back to Login
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="auth-page">
      <div className="auth-container">
        <p className="auth-message">Completing sign-in...</p>
      </div>
    </div>
  )
}
