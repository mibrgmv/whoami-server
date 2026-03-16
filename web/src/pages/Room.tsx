import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { GameBoard } from '../components/GameBoard'
import { Keyboard } from '../components/Keyboard'
import { useRoomStore } from '../stores/roomStore'
import { useAuthStore } from '../stores/authStore'
import './Room.css'

export function Room() {
  const { code } = useParams<{ code: string }>()
  const navigate = useNavigate()
  const { isAuthenticated, loginAsGuest } = useAuthStore()

  const {
    room,
    players,
    currentPlayer,
    currentGuess,
    letterStates,
    wordLength,
    isLoading,
    error,
    gameResult,
    joinRoom,
    connectWebSocket,
    setReady,
    startGame,
    submitGuess,
    nextRound,
    leaveRoom,
    addLetter,
    removeLetter,
    clearError,
    reset,
  } = useRoomStore()

  const [displayName, setDisplayName] = useState('')
  const [hasJoined, setHasJoined] = useState(false)

  // Auto-login as guest if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      loginAsGuest()
    }
  }, [isAuthenticated, loginAsGuest])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      reset()
    }
  }, [reset])

  // Connect WebSocket after joining
  useEffect(() => {
    if (hasJoined && currentPlayer && code) {
      connectWebSocket(code, currentPlayer.playerId)
    }
  }, [hasJoined, currentPlayer, code, connectWebSocket])

  // Keyboard handler
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (room?.status !== 'playing') return

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
    [room?.status, submitGuess, removeLetter, addLetter]
  )

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  const handleJoin = async () => {
    if (!code || !displayName.trim()) return

    try {
      await joinRoom(code, displayName.trim())
      setHasJoined(true)
    } catch {
      // Error handled in store
    }
  }

  const handleLeave = () => {
    leaveRoom()
    navigate('/')
  }

  const handleCopyCode = () => {
    if (code) {
      navigator.clipboard.writeText(code)
    }
  }

  const isHost = currentPlayer?.userId === room?.hostId
  const isReady = currentPlayer?.status === 'ready'
  const allReady = players.length > 1 && players.every((p) => p.status === 'ready')
  const isPlaying = room?.status === 'playing'
  const isFinished = room?.status === 'finished'

  // Join screen
  if (!hasJoined) {
    return (
      <div className="room-page">
        <header className="room-header">
          <button className="btn btn-text" onClick={() => navigate('/')}>
            ← Back
          </button>
          <h1>Join Room</h1>
          <div style={{ width: 60 }} />
        </header>

        {error && (
          <div className="error-banner" onClick={clearError}>
            {error}
          </div>
        )}

        <div className="join-form">
          <p className="room-code-display">{code}</p>
          <input
            type="text"
            placeholder="Your display name"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            maxLength={20}
            className="input"
          />
          <button
            className="btn btn-primary"
            onClick={handleJoin}
            disabled={!displayName.trim() || isLoading}
          >
            {isLoading ? 'Joining...' : 'Join Room'}
          </button>
        </div>
      </div>
    )
  }

  // Lobby
  if (room?.status === 'waiting') {
    return (
      <div className="room-page">
        <header className="room-header">
          <button className="btn btn-text" onClick={handleLeave}>
            ← Leave
          </button>
          <h1>Lobby</h1>
          <div style={{ width: 60 }} />
        </header>

        {error && (
          <div className="error-banner" onClick={clearError}>
            {error}
          </div>
        )}

        <div className="lobby">
          <div className="room-code-section">
            <p>Room Code</p>
            <button className="room-code-btn" onClick={handleCopyCode}>
              {code} <span>Copy</span>
            </button>
          </div>

          <div className="players-list">
            <h3>Players ({players.length}/{room.settings.maxPlayers})</h3>
            {players.map((player) => (
              <div key={player.playerId} className="player-item">
                <span className="player-name">
                  {player.displayName}
                  {player.userId === room.hostId && ' (Host)'}
                </span>
                <span className={`player-status ${player.status}`}>
                  {player.status === 'ready' ? '✓ Ready' : 'Waiting'}
                </span>
              </div>
            ))}
          </div>

          <div className="lobby-actions">
            <button
              className={`btn ${isReady ? 'btn-secondary' : 'btn-primary'}`}
              onClick={() => setReady(!isReady)}
            >
              {isReady ? 'Not Ready' : 'Ready'}
            </button>

            {isHost && (
              <button
                className="btn btn-primary"
                onClick={startGame}
                disabled={!allReady}
              >
                Start Game
              </button>
            )}
          </div>

          {isHost && !allReady && players.length > 1 && (
            <p className="lobby-hint">Waiting for all players to be ready...</p>
          )}

          {players.length === 1 && (
            <p className="lobby-hint">Waiting for more players to join...</p>
          )}
        </div>
      </div>
    )
  }

  // Game / Results
  return (
    <div className="room-page">
      <header className="room-header">
        <button className="btn btn-text" onClick={handleLeave}>
          ← Leave
        </button>
        <h1>
          {code} - Round {room?.roundNumber || 1}
        </h1>
        <div style={{ width: 60 }} />
      </header>

      {error && (
        <div className="error-banner" onClick={clearError}>
          {error}
        </div>
      )}

      <div className="game-area">
        <div className="main-board">
          <GameBoard
            guesses={currentPlayer?.guesses || []}
            currentGuess={currentGuess}
            maxAttempts={6}
            wordLength={wordLength}
          />
        </div>

        <div className="players-sidebar">
          <h3>Players</h3>
          {players.map((player) => (
            <div
              key={player.playerId}
              className={`player-card ${player.playerId === currentPlayer?.playerId ? 'current' : ''}`}
            >
              <span className="player-name">{player.displayName}</span>
              <span className="player-attempts">
                {player.currentAttempts}/6
              </span>
            </div>
          ))}
        </div>
      </div>

      {gameResult && (
        <div className="round-result">
          <h2>{'targetWord' in gameResult ? 'Round Over!' : 'Game Over!'}</h2>
          {'targetWord' in gameResult && (
            <p>
              The word was: <strong>{gameResult.targetWord?.toUpperCase()}</strong>
            </p>
          )}
          <div className="scores">
            {(gameResult as { results?: Array<{ displayName: string; score: number }> }).results?.map((r, i) => (
              <div key={i} className="score-row">
                <span>{r.displayName}</span>
                <span>{r.score} pts</span>
              </div>
            ))}
          </div>
          {isHost && !isFinished && (
            <button className="btn btn-primary" onClick={nextRound}>
              Next Round
            </button>
          )}
          {isFinished && (
            <button className="btn btn-primary" onClick={handleLeave}>
              Back to Home
            </button>
          )}
        </div>
      )}

      {isPlaying && !gameResult && (
        <Keyboard
          onKey={addLetter}
          onEnter={submitGuess}
          onBackspace={removeLetter}
          letterStates={letterStates}
          disabled={isLoading}
        />
      )}
    </div>
  )
}
