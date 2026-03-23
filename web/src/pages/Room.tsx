import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { GameBoard } from '../components/GameBoard'
import { Keyboard } from '../components/Keyboard'
import { useRoomStore } from '../stores/roomStore'
import { useAuthStore } from '../stores/authStore'
import { useToastStore } from '../stores/toastStore'
import { useKeyboardInput } from '../hooks/useKeyboardInput'
import { QRCodeSVG } from 'qrcode.react'
import { room as roomApi } from '../api/client'
import { RoomStatusValues, PlayerStatusValues } from '../types/api'
import type { LetterResult } from '../types/api'
import { updateLetterStates } from '../utils/letterStates'
import './Room.css'

export function Room() {
  const { code } = useParams<{ code: string }>()
  const navigate = useNavigate()
  const { isAuthenticated, isGuest, username } = useAuthStore()

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
          let letterStates: Record<string, LetterResult> = {}
          for (const guess of player.guesses) {
            letterStates = updateLetterStates(letterStates, guess)
          }
          useRoomStore.setState({
            room: fullRoom.room,
            players: fullRoom.players,
            currentPlayer: player,
            letterStates,
          })
        } else {
          localStorage.removeItem(`room_${code}_player`)
        }
      } catch {
        localStorage.removeItem(`room_${code}_player`)
        useToastStore.getState().addToast('Room not found or expired', 'error')
        navigate('/')
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
    const name = isGuest ? displayName.trim() : username
    if (!code || !name) return

    try {
      await joinRoom(code, name)
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

  // Auto-join for logged-in (non-guest) users — skip if restoring existing session
  useEffect(() => {
    if (!hasJoined && !currentPlayer && !isLoading && isAuthenticated && !isGuest && username && code) {
      const savedPlayerId = localStorage.getItem(`room_${code}_player`)
      if (!savedPlayerId) {
        handleJoin()
      }
    }
  }, [hasJoined, currentPlayer, isLoading, isAuthenticated, isGuest, username, code])

  // ─── Join screen (guests only) ───
  if (!hasJoined) {
    return (
      <div className="room-page">
        <header className="page-header">
          <button className="page-header-back" onClick={() => navigate('/')}>←</button>
        </header>

        <div className="room-content">
          <div className="room-card join-card">
            <p className="join-card-code">{code}</p>
            <input
              type="text"
              placeholder="Your display name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleJoin()}
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
      </div>
    )
  }

  // ─── Lobby ───
  if (room?.status === RoomStatusValues.WAITING) {
    return (
      <div className="room-page">
        <header className="page-header">
          <button className="page-header-back" onClick={handleLeave}>←</button>
        </header>

        <div className="room-content">
          {/* Invite card */}
          <div className="room-card invite-card">
            <div className="invite-qr">
              <QRCodeSVG
                value={`${window.location.origin}/room/${code}`}
                size={140}
                bgColor="transparent"
                fgColor="currentColor"
              />
            </div>
            <button className="invite-code-btn" onClick={handleCopyCode}>
              {code}
              <div className="invite-code-hint">
                {copiedCode ? 'Copied!' : 'Tap to copy'}
              </div>
            </button>
          </div>

          {/* Players card */}
          <div className="room-card">
            <div className="players-card-header">
              <div className="room-card-title">Players</div>
              <span>{players.length}/{room.settings.maxPlayers}</span>
            </div>
            {players.map((player) => (
              <div key={player.playerId} className="player-row">
                <span className="player-name">
                  {player.displayName}
                  {player.playerId === room.hostId && (
                    <span className="player-tag"> · Host</span>
                  )}
                </span>
                <span className={`player-status ${player.status === PlayerStatusValues.READY ? 'ready' : ''}`}>
                  {player.status === PlayerStatusValues.READY ? 'Ready' : 'Waiting'}
                </span>
              </div>
            ))}
          </div>

          {/* Settings card */}
          <div className="room-card settings-card">
            <div className="room-card-title">Settings</div>
            <div className="settings-rows">
              <div className="settings-row">
                <span>Max Players</span>
                <span>{room.settings.maxPlayers}</span>
              </div>
              <div className="settings-row">
                <span>Show Guesses</span>
                <span>{room.settings.showGuesses ? 'Yes' : 'No'}</span>
              </div>
            </div>
          </div>

          {/* Actions */}
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

          {!isWsConnected && (
            <p className="lobby-hint">Connecting...</p>
          )}
        </div>
      </div>
    )
  }

  // ─── Playing / Finished ───
  return (
    <div className="room-page">
      <header className="page-header">
        <button className="page-header-back" onClick={handleLeave}>←</button>
      </header>

      <div className="game-area">
        <div className="players-bar">
          {players.map((player) => (
            <div
              key={player.playerId}
              className={`player-chip ${player.playerId === currentPlayer?.playerId ? 'current' : ''}`}
            >
              <span>{player.displayName}</span>
              <span className="player-chip-attempts">{player.currentAttempts}/6</span>
            </div>
          ))}
        </div>

        <div className="main-board">
          <GameBoard
            guesses={currentPlayer?.guesses || []}
            currentGuess={currentGuess}
            maxAttempts={6}
            wordLength={wordLength}
          />
        </div>

        {gameResult && (
          <div className="round-result">
            <div className="round-result-card">
              <h2>{'targetWord' in gameResult ? 'Round Over!' : 'Game Over!'}</h2>
              {'targetWord' in gameResult && (
                <p>
                  The word was <strong>{gameResult.targetWord?.toUpperCase()}</strong>
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
            </div>
            <div className="round-result-actions">
              {isHost && isFinished && (
                <button className="btn btn-primary" onClick={nextRound} disabled={!isWsConnected}>
                  Play Again
                </button>
              )}
              {isFinished && (
                <button className="btn btn-secondary" onClick={handleLeave}>
                  Home
                </button>
              )}
            </div>
          </div>
        )}
      </div>

      <Keyboard
        onKey={addLetter}
        onEnter={submitGuess}
        onBackspace={removeLetter}
        letterStates={letterStates}
        disabled={!canPlay}
      />
    </div>
  )
}
