import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { useRoomStore } from '../stores/roomStore'
import { game } from '../api/client'
import type { DailyStatus } from '../types/api'
import { GameStatusValues } from '../types/api'
import './Home.css'

export function Home() {
  const navigate = useNavigate()
  const { isAuthenticated, isGuest, hasHydrated, loginAsGuest, logout } = useAuthStore()
  const { createRoom, isLoading, error, clearError } = useRoomStore()

  const [roomCode, setRoomCode] = useState('')
  const [showAuth, setShowAuth] = useState(false)
  const [dailyStatus, setDailyStatus] = useState<DailyStatus | null>(null)

  useEffect(() => {
    if (hasHydrated && isAuthenticated) {
      game.dailyStatus().then(setDailyStatus).catch(() => {})
    }
  }, [hasHydrated, isAuthenticated])

  const handlePlayDaily = () => {
    navigate('/game/daily')
  }

  const handlePlayRandom = () => {
    navigate('/game')
  }

  const handleCreateRoom = async () => {
    if (!isAuthenticated) {
      setShowAuth(true)
      return
    }

    if (isGuest) {
      alert('Guests cannot create rooms. Please log in.')
      return
    }

    try {
      const code = await createRoom()
      navigate(`/room/${code}`)
    } catch {
      // Error handled in store
    }
  }

  const handleJoinRoom = () => {
    if (roomCode.trim()) {
      navigate(`/room/${roomCode.trim().toUpperCase()}`)
    }
  }

  const handleGuestLogin = async () => {
    try {
      await loginAsGuest()
      setShowAuth(false)
    } catch {
      alert('Failed to login as guest')
    }
  }

  const getDailyButtonText = () => {
    if (!dailyStatus) return 'Daily Challenge'
    if (dailyStatus.hasPlayedToday) {
      return dailyStatus.status === GameStatusValues.WON ? 'Daily ✓ Completed' : 'Daily ✗ Try Tomorrow'
    }
    return 'Daily Challenge'
  }

  const isDailyDisabled = dailyStatus?.hasPlayedToday ?? false

  return (
    <div className="home">
      <h1 className="home-title">GORDLE</h1>
      <p className="home-subtitle">Multiplayer word game</p>

      {error && (
        <div className="error-banner" onClick={clearError}>
          {error}
        </div>
      )}

      <div className="home-actions">
        <div className="play-modes">
          <button
            className="btn btn-primary btn-large"
            onClick={handlePlayDaily}
            disabled={isDailyDisabled}
          >
            {getDailyButtonText()}
          </button>
          <button
            className="btn btn-secondary btn-large"
            onClick={handlePlayRandom}
          >
            Random Word
          </button>
        </div>

        <div className="divider">
          <span>multiplayer</span>
        </div>

        <button
          className="btn btn-secondary"
          onClick={handleCreateRoom}
          disabled={isLoading}
        >
          {isLoading ? 'Creating...' : 'Create Room'}
        </button>

        <div className="join-room">
          <input
            type="text"
            placeholder="Room code"
            value={roomCode}
            onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
            maxLength={6}
            className="input"
          />
          <button
            className="btn btn-secondary"
            onClick={handleJoinRoom}
            disabled={!roomCode.trim()}
          >
            Join
          </button>
        </div>
      </div>

      <div className="home-footer">
        {isAuthenticated ? (
          <button className="btn btn-text" onClick={logout}>
            {isGuest ? 'Playing as Guest' : 'Log out'}
          </button>
        ) : (
          <button className="btn btn-text" onClick={() => setShowAuth(true)}>
            Log in / Register
          </button>
        )}
      </div>

      {showAuth && !isAuthenticated && (
        <div className="modal-overlay" onClick={() => setShowAuth(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>Get Started</h2>
            <div className="modal-actions">
              <button className="btn btn-primary" onClick={handleGuestLogin}>
                Play as Guest
              </button>
              <button
                className="btn btn-secondary"
                onClick={() => navigate('/login')}
              >
                Log in
              </button>
              <button
                className="btn btn-text"
                onClick={() => navigate('/register')}
              >
                Create account
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
