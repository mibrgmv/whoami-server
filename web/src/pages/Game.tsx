import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { GameBoard } from '../components/GameBoard'
import { Keyboard } from '../components/Keyboard'
import { useGameStore } from '../stores/gameStore'
import { useAuthStore } from '../stores/authStore'
import { useKeyboardInput } from '../hooks/useKeyboardInput'
import { useErrorProgress } from '../hooks/useErrorProgress'
import { GameStatusValues } from '../types/api'
import './Game.css'

export function Game() {
  const navigate = useNavigate()
  const location = useLocation()
  const isDaily = location.pathname === '/game/daily'

  const { isAuthenticated, hasHydrated } = useAuthStore()
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

  const errorProgress = useErrorProgress(error, 1500, clearError)
  const isGameOver = session?.status === GameStatusValues.WON || session?.status === GameStatusValues.LOST

  useKeyboardInput({
    onLetter: addLetter,
    onEnter: submitGuess,
    onBackspace: removeLetter,
    disabled: session?.status !== GameStatusValues.IN_PROGRESS,
  })

  useEffect(() => {
    if (hasHydrated && !isAuthenticated) {
      navigate('/')
    }
  }, [hasHydrated, isAuthenticated, navigate])

  useEffect(() => {
    if (hasHydrated && isAuthenticated && !session) {
      startGame(isDaily ? 'daily' : 'random')
    }

    return () => {
      reset()
    }
  }, [hasHydrated, isAuthenticated, isDaily])

  const handlePlayAgain = () => {
    if (isDaily) {
      navigate('/')
    } else {
      reset()
      startGame('random')
    }
  }

  return (
    <div className="game-page">
      <header className="game-header">
        <button className="btn btn-text" onClick={() => navigate('/')}>
          ← Back
        </button>
        <h1>{isDaily ? 'DAILY' : 'GORDLE'}</h1>
        <div style={{ width: 60 }} />
      </header>

      <div className="error-container">
        {error && (
          <div className="error-banner-game" onClick={clearError}>
            {error}
            <div
              className="error-progress"
              style={{ width: `${errorProgress}%` }}
            />
          </div>
        )}
      </div>

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
