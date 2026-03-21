import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuthStore, redirectToLogin } from '../stores/authStore'
import { auth } from '../api/client'
import './Auth.css'

function safeReturnTo(): string {
  const returnTo = sessionStorage.getItem('auth_return_to')
  sessionStorage.removeItem('auth_return_to')
  if (returnTo && returnTo.startsWith('/') && !returnTo.startsWith('//')) {
    return returnTo
  }
  return '/'
}

export function OAuthCallback() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setTokens } = useAuthStore()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const code = searchParams.get('code')
    const errorParam = searchParams.get('error')
    const kcActionStatus = searchParams.get('kc_action_status')

    if (errorParam) {
      if (searchParams.get('error_description') === 'authentication_expired') {
        navigate('/')
        return
      }
      setError(searchParams.get('error_description') || 'Authentication failed')
      return
    }

    if (!code) {
      navigate(safeReturnTo())
      return
    }

    if (kcActionStatus) {
      auth.exchangeCode(code)
        .then((tokens) => {
          setTokens(tokens.access_token, tokens.refresh_token, false)
        })
        .catch(() => {
          // Token refresh failed, not critical
        })
        .finally(() => {
          navigate(safeReturnTo())
        })
      return
    }

    auth.exchangeCode(code)
      .then((tokens) => {
        setTokens(tokens.access_token, tokens.refresh_token, false)
        navigate(safeReturnTo())
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
