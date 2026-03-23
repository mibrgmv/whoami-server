import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { GameBoard } from '../components/GameBoard'
import { Keyboard } from '../components/Keyboard'
import { useGameStore } from '../stores/gameStore'
import { useAuthStore } from '../stores/authStore'
import { useKeyboardInput } from '../hooks/useKeyboardInput'
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
    startGame,
    addLetter,
    removeLetter,
    submitGuess,
    reset,
  } = useGameStore()

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
    reset()
    if (hasHydrated && isAuthenticated) {
      startGame(isDaily ? 'daily' : 'random')
    }
  }, [hasHydrated, isAuthenticated, isDaily])

  useEffect(() => {
    return () => reset()
  }, [])

  const handlePlayAgain = () => {
    if (isDaily) {
      navigate('/')
    } else {
      useGameStore.getState().clearRandomSession()
      reset()
      startGame('random')
    }
  }

  return (
    <div className="game-page">
      <header className="page-header">
        <button className="page-header-back" onClick={() => navigate('/')}>←</button>
      </header>

      <h1 className="page-title">{isDaily ? 'DAILY' : 'RANDOM'}</h1>

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

          <div className="game-spacer" />

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
