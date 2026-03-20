import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import './Auth.css'

async function exchangeCodeForTokens(code: string): Promise<{
  access_token: string
  refresh_token: string
}> {
  const redirectUri = `${window.location.origin}/oauth/callback`
  const body = new URLSearchParams({
    grant_type: 'authorization_code',
    client_id: 'gordle-public',
    code,
    redirect_uri: redirectUri,
  })

  const tokenUrl = `/realms/gordle-realm/protocol/openid-connect/token`
  const response = await fetch(tokenUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body,
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error_description || 'Failed to exchange authorization code')
  }

  return response.json()
}

export function OAuthCallback() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setTokens } = useAuthStore()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const code = searchParams.get('code')
    const errorParam = searchParams.get('error')

    if (errorParam) {
      setError(searchParams.get('error_description') || 'Authentication failed')
      return
    }

    if (!code) {
      setError('No authorization code received')
      return
    }

    exchangeCodeForTokens(code)
      .then((tokens) => {
        setTokens(tokens.access_token, tokens.refresh_token, false)
        navigate('/')
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
          <button className="btn btn-primary" onClick={() => navigate('/login')}>
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
