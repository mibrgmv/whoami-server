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
  const [showRoomSettings, setShowRoomSettings] = useState(false)
  const [showGuesses, setShowGuesses] = useState(true)
  const [dailyStatus, setDailyStatus] = useState<DailyStatus | null>(null)
  const [showDailyHint, setShowDailyHint] = useState(false)

  useEffect(() => {
    if (hasHydrated && isAuthenticated) {
      game.dailyStatus().then(setDailyStatus).catch(() => {})
    }
  }, [hasHydrated, isAuthenticated])

  const handlePlayDaily = () => {
    if (isDailyDisabled) {
      setShowDailyHint(true)
      setTimeout(() => setShowDailyHint(false), 1500)
      return
    }
    navigate('/game/daily')
  }

  const handlePlayRandom = () => {
    navigate('/game')
  }

  const handleCreateRoom = () => {
    if (!isAuthenticated) {
      setShowAuth(true)
      return
    }

    if (isGuest) {
      alert('Guests cannot create rooms. Please log in.')
      return
    }

    setShowRoomSettings(true)
  }

  const handleConfirmCreateRoom = async () => {
    try {
      const code = await createRoom({ settings: { showGuesses } })
      setShowRoomSettings(false)
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

  const handleLogout = () => {
    logout()
    setDailyStatus(null)
  }

  const getDailyButtonText = () => {
    if (!dailyStatus) return 'Daily Challenge'
    if (dailyStatus.hasPlayedToday) {
      if (dailyStatus.status === GameStatusValues.WON) {
        return 'Daily ✓ Completed'
      }
      if (dailyStatus.status === GameStatusValues.LOST) {
        return 'Daily ✗ Try Tomorrow'
      }
      if (dailyStatus.status === GameStatusValues.IN_PROGRESS) {
        return 'Daily - Continue'
      }
    }
    return 'Daily Challenge'
  }

  const isDailyDisabled = !isAuthenticated ||
    isGuest ||
    (dailyStatus?.hasPlayedToday &&
     (dailyStatus.status === GameStatusValues.WON ||
      dailyStatus.status === GameStatusValues.LOST))

  return (
    <div className="home">
      <div className="home-header">
        <div />
        {isAuthenticated && (
          <div className="header-actions">
            {!isGuest && (
              <button
                className="btn btn-text"
                onClick={() => navigate('/profile')}
              >
                Profile
              </button>
            )}
            <div className="user-info">
              {isGuest ? 'Playing as Guest' : 'User'}
            </div>
          </div>
        )}
      </div>

      <h1 className="home-title">GORDLE</h1>
      <p className="home-subtitle">Multiplayer word game</p>

      {error && (
        <div className="error-banner" onClick={clearError}>
          {error}
        </div>
      )}

      <div className="home-actions">
        <div className="divider">
          <span>singleplayer</span>
        </div>

        <div className="play-modes">
          <div
            className="button-with-hint"
            onClick={isDailyDisabled ? handlePlayDaily : undefined}
          >
            <button
              className="btn btn-primary btn-large"
              onClick={!isDailyDisabled ? handlePlayDaily : undefined}
              disabled={isDailyDisabled}
            >
              {getDailyButtonText()}
            </button>
            <div className="hint-container">
              {showDailyHint && (
                <div className="hint-text">
                  Register to play Daily
                </div>
              )}
            </div>
          </div>
          <button
            className="btn btn-secondary btn-large"
            onClick={handlePlayRandom}
            disabled={!isAuthenticated}
          >
            Random Word
          </button>
        </div>

        <div className="divider">
          <span>multiplayer</span>
        </div>

        <div className="multiplayer-actions">
          <button
            className="btn btn-secondary btn-large btn-full"
            onClick={handleCreateRoom}
            disabled={!isAuthenticated || isLoading}
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
              disabled={!isAuthenticated || !roomCode.trim()}
            >
              Join
            </button>
          </div>
        </div>
      </div>

      <div className="home-footer">
        <button className="btn btn-text" onClick={() => setShowAuth(true)}>
          {isAuthenticated ? 'Log out' : 'Log in'}
        </button>
      </div>

      {showAuth && (
        <div className="modal-overlay" onClick={() => setShowAuth(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>{isAuthenticated ? 'Account' : 'Get Started'}</h2>
            <div className="modal-actions">
              {isAuthenticated ? (
                <>
                  <div className="auth-status">
                    {isGuest ? 'Playing as Guest' : 'Logged in'}
                  </div>
                  <button className="btn btn-primary" onClick={() => { handleLogout(); setShowAuth(false); }}>
                    Log out
                  </button>
                  {isGuest && (
                    <>
                      <button
                        className="btn btn-secondary"
                        onClick={() => navigate('/login')}
                      >
                        Switch to Account
                      </button>
                      <button
                        className="btn btn-text"
                        onClick={() => navigate('/register')}
                      >
                        Create account
                      </button>
                    </>
                  )}
                </>
              ) : (
                <>
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
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {showRoomSettings && (
        <div className="modal-overlay" onClick={() => setShowRoomSettings(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>Room Settings</h2>
            <div className="room-settings">
              <label className="toggle-setting">
                <span>Show guesses to all players</span>
                <input
                  type="checkbox"
                  checked={showGuesses}
                  onChange={(e) => setShowGuesses(e.target.checked)}
                />
              </label>
              <p className="setting-hint">
                Players will see each other's results as emoji squares
              </p>
            </div>
            <div className="modal-actions">
              <button
                className="btn btn-primary"
                onClick={handleConfirmCreateRoom}
                disabled={isLoading}
              >
                {isLoading ? 'Creating...' : 'Create Room'}
              </button>
              <button
                className="btn btn-text"
                onClick={() => setShowRoomSettings(false)}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
