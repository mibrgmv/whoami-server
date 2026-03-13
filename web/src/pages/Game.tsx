import { useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { GameBoard } from '../components/GameBoard'
import { Keyboard } from '../components/Keyboard'
import { useGameStore } from '../stores/gameStore'
import { useAuthStore } from '../stores/authStore'
import './Game.css'

export function Game() {
  const navigate = useNavigate()
  const { isAuthenticated, loginAsGuest } = useAuthStore()
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

  // Auto-login as guest if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      loginAsGuest()
    }
  }, [isAuthenticated, loginAsGuest])

  // Start game on mount
  useEffect(() => {
    if (isAuthenticated && !session) {
      startGame('random')
    }

    return () => {
      reset()
    }
  }, [isAuthenticated, session, startGame, reset])

  // Keyboard handler
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (session?.status !== 'in_progress') return

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
    startGame('random')
  }

  const isGameOver = session?.status === 'won' || session?.status === 'lost'

  return (
    <div className="game-page">
      <header className="game-header">
        <button className="btn btn-text" onClick={() => navigate('/')}>
          ← Back
        </button>
        <h1>GORDLE</h1>
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
          <GameBoard
            guesses={session.guesses}
            currentGuess={currentGuess}
            maxAttempts={session.max_attempts}
          />

          {isGameOver && (
            <div className="game-result">
              {session.status === 'won' ? (
                <p className="result-text result-won">
                  You won in {session.attempts_used} {session.attempts_used === 1 ? 'try' : 'tries'}!
                </p>
              ) : (
                <p className="result-text result-lost">
                  The word was <strong>{session.target_word?.toUpperCase()}</strong>
                </p>
              )}
              <button className="btn btn-primary" onClick={handlePlayAgain}>
                Play Again
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
