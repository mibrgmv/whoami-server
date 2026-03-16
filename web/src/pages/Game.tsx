import { useEffect, useCallback } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { GameBoard } from '../components/GameBoard'
import { Keyboard } from '../components/Keyboard'
import { useGameStore } from '../stores/gameStore'
import { useAuthStore } from '../stores/authStore'
import { GameStatusValues } from '../types/api'
import './Game.css'

export function Game() {
  const navigate = useNavigate()
  const location = useLocation()
  const isDaily = location.pathname === '/game/daily'

  const { isAuthenticated, hasHydrated, loginAsGuest } = useAuthStore()
  const {
    session,
    currentGuess,
    letterStates,
    isLoading,
    error,
    startGame,
    addLetter,
    removeLetter,
    submitGuess,
    clearError,
    reset,
  } = useGameStore()

  // Auto-login as guest if not authenticated (only after hydration)
  useEffect(() => {
    if (hasHydrated && !isAuthenticated) {
      loginAsGuest()
    }
  }, [hasHydrated, isAuthenticated, loginAsGuest])

  // Start game on mount (only after hydration and authentication)
  useEffect(() => {
    if (hasHydrated && isAuthenticated && !session) {
      startGame(isDaily ? 'daily' : 'random')
    }

    return () => {
      reset()
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hasHydrated, isAuthenticated, isDaily])

  // Keyboard handler
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (session?.status !== GameStatusValues.IN_PROGRESS) return

      if (e.key === 'Enter') {
        e.preventDefault()
        submitGuess()
      } else if (e.key === 'Backspace') {
        e.preventDefault()
        removeLetter()
      } else if (/^[a-zA-Z]$/.test(e.key)) {
        addLetter(e.key)
      }
    },
    [session?.status, submitGuess, removeLetter, addLetter]
  )

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  const handlePlayAgain = () => {
    if (isDaily) {
      navigate('/')
    } else {
      reset()
      startGame('random')
    }
  }

  const isGameOver = session?.status === GameStatusValues.WON || session?.status === GameStatusValues.LOST

  return (
    <div className="game-page">
      <header className="game-header">
        <button className="btn btn-text" onClick={() => navigate('/')}>
          ← Back
        </button>
        <h1>{isDaily ? 'DAILY' : 'GORDLE'}</h1>
        <div style={{ width: 60 }} />
      </header>

      {error && (
        <div className="error-banner" onClick={clearError}>
          {error}
        </div>
      )}

      {isLoading && !session && <div className="loading">Loading...</div>}

      {session && (
        <>
          {isDaily && (
            <div className="game-date">
              {new Date(session.gameDate).toLocaleDateString('en-US', {
                weekday: 'long',
                month: 'long',
                day: 'numeric',
              })}
            </div>
          )}

          <GameBoard
            guesses={session.guesses}
            currentGuess={currentGuess}
            maxAttempts={session.maxAttempts}
          />

          {isGameOver && (
            <div className="game-result">
              {session.status === GameStatusValues.WON ? (
                <p className="result-text result-won">
                  You won in {session.attemptsUsed} {session.attemptsUsed === 1 ? 'try' : 'tries'}!
                </p>
              ) : (
                <p className="result-text result-lost">
                  The word was <strong>{session.targetWord?.toUpperCase()}</strong>
                </p>
              )}
              <button className="btn btn-primary" onClick={handlePlayAgain}>
                {isDaily ? 'Back to Home' : 'Play Again'}
              </button>
            </div>
          )}

          <Keyboard
            onKey={addLetter}
            onEnter={submitGuess}
            onBackspace={removeLetter}
            letterStates={letterStates}
            disabled={isGameOver || isLoading}
          />
        </>
      )}
    </div>
  )
}
