import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { GameBoard } from '../components/GameBoard'
import { Keyboard } from '../components/Keyboard'
import { useRoomStore } from '../stores/roomStore'
import { useAuthStore } from '../stores/authStore'
import { useKeyboardInput } from '../hooks/useKeyboardInput'
import { QRCodeSVG } from 'qrcode.react'
import { room as roomApi } from '../api/client'
import { RoomStatusValues, PlayerStatusValues } from '../types/api'
import './Room.css'

export function Room() {
  const { code } = useParams<{ code: string }>()
  const navigate = useNavigate()
  const { isAuthenticated } = useAuthStore()

  const {
    room,
    players,
    currentPlayer,
    currentGuess,
    letterStates,
    wordLength,
    isLoading,
    gameResult,
    isWsConnected,
    joinRoom,
    connectWebSocket,
    setReady,
    startGame,
    submitGuess,
    nextRound,
    leaveRoom,
    addLetter,
    removeLetter,
    reset,
  } = useRoomStore()

  const [displayName, setDisplayName] = useState('')
  const [hasJoined, setHasJoined] = useState(false)
  const [copiedCode, setCopiedCode] = useState(false)

  const isHost = currentPlayer?.playerId === room?.hostId
  const isReady = currentPlayer?.status === PlayerStatusValues.READY
  const allReady = players.length > 1 && players.every((p) => p.status === PlayerStatusValues.READY)
  const isPlaying = room?.status === RoomStatusValues.PLAYING
  const isFinished = room?.status === RoomStatusValues.FINISHED

  const canPlay = isPlaying &&
    !gameResult &&
    currentPlayer?.status === PlayerStatusValues.PLAYING

  useKeyboardInput({
    onLetter: addLetter,
    onEnter: submitGuess,
    onBackspace: removeLetter,
    disabled: !canPlay,
  })

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/')
      return
    }

    const restoreRoomState = async () => {
      if (!code || room || isLoading) return

      const savedPlayerId = localStorage.getItem(`room_${code}_player`)
      if (!savedPlayerId) return

      try {
        const fullRoom = await roomApi.get(code)
        const player = fullRoom.players.find(p => p.playerId === savedPlayerId)

        if (player) {
          useRoomStore.setState({
            room: fullRoom.room,
            players: fullRoom.players,
            currentPlayer: player,
          })
        }
      } catch {
      }
    }

    restoreRoomState()
  }, [isAuthenticated, navigate, code, room, isLoading])

  useEffect(() => {
    return () => {
      reset()
    }
  }, [reset])

  useEffect(() => {
    if (currentPlayer && code && !hasJoined) {
      setHasJoined(true)
    }
  }, [currentPlayer, code, hasJoined])

  useEffect(() => {
    if (hasJoined && currentPlayer && code && !isWsConnected) {
      connectWebSocket(code, currentPlayer.playerId)
    }
  }, [hasJoined, currentPlayer?.playerId, code, isWsConnected, connectWebSocket])

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

  const handleCopyCode = async () => {
    if (code) {
      try {
        await navigator.clipboard.writeText(code)
        setCopiedCode(true)
        setTimeout(() => setCopiedCode(false), 2000)
      } catch {
        setCopiedCode(false)
      }
    }
  }

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

  if (room?.status === RoomStatusValues.WAITING) {
    return (
      <div className="room-page">
        <header className="room-header">
          <button className="btn btn-text" onClick={handleLeave}>
            ← Leave
          </button>
          <h1>Lobby</h1>
          <div style={{ width: 60 }} />
        </header>

        <div className="lobby">
          <div className="room-code-section">
            <div className="qr-code">
              <QRCodeSVG
                value={`${window.location.origin}/room/${code}`}
                size={160}
                bgColor="transparent"
                fgColor="currentColor"
              />
            </div>
            <button className="room-code-btn" onClick={handleCopyCode}>
              {code} <span>{copiedCode ? '✓ Copied' : 'Copy'}</span>
            </button>
          </div>

          <div className="players-list">
            <h3>Players ({players.length}/{room.settings.maxPlayers})</h3>
            {players.map((player) => (
              <div key={player.playerId} className="player-item">
                <span className="player-name">
                  {player.displayName}
                  {player.playerId === room.hostId && ' (Host)'}
                </span>
                <span className={`player-status ${player.status === PlayerStatusValues.READY ? 'ready' : ''}`}>
                  {player.status === PlayerStatusValues.READY ? '✓ Ready' : 'Waiting'}
                </span>
              </div>
            ))}
          </div>

          <div className="lobby-actions">
            <button
              className={`btn ${isReady ? 'btn-secondary' : 'btn-primary'}`}
              onClick={() => setReady(!isReady)}
              disabled={!isWsConnected}
            >
              {isReady ? 'Not Ready' : 'Ready'}
            </button>

            {isHost && (
              <button
                className="btn btn-primary"
                onClick={startGame}
                disabled={!allReady || !isWsConnected}
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

          {!isWsConnected && (
            <p className="lobby-hint">Connecting...</p>
          )}
        </div>
      </div>
    )
  }

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
          {isHost && isFinished && (
            <button className="btn btn-primary" onClick={nextRound} disabled={!isWsConnected}>
              Play Again
            </button>
          )}
          {isFinished && (
            <button className="btn btn-primary" onClick={handleLeave}>
              Back to Home
            </button>
          )}
        </div>
      )}

      {canPlay && (
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
