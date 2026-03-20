import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { useRoomStore } from '../stores/roomStore'
import { useGameStore } from '../stores/gameStore'
import { useToastStore } from '../stores/toastStore'
import { GameStatusValues } from '../types/api'
import './Home.css'

export function Home() {
  const navigate = useNavigate()
  const { isAuthenticated, isGuest, hasHydrated, loginAsGuest, logout } = useAuthStore()
  const { createRoom, isLoading } = useRoomStore()
  const { dailyStatus, dailyStatusFetched, fetchDailyStatus, clearDailyStatus } = useGameStore()

  const [roomCode, setRoomCode] = useState('')
  const [showAuth, setShowAuth] = useState(false)
  const [showRoomSettings, setShowRoomSettings] = useState(false)
  const [showGuesses, setShowGuesses] = useState(true)
  const [roomMode, setRoomMode] = useState<'single_round' | 'marathon'>('single_round')
  const [maxPlayers, setMaxPlayers] = useState(6)
  const [timeLimitSecs, setTimeLimitSecs] = useState(180)
  const [showDailyHint, setShowDailyHint] = useState(false)

  useEffect(() => {
    if (hasHydrated && isAuthenticated && !isGuest && !dailyStatusFetched) {
      fetchDailyStatus()
    }
  }, [hasHydrated, isAuthenticated, isGuest, dailyStatusFetched, fetchDailyStatus])

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
      useToastStore.getState().addToast('Guests cannot create rooms. Please log in.', 'error')
      return
    }

    setShowRoomSettings(true)
  }

  const handleConfirmCreateRoom = async () => {
    try {
      const code = await createRoom({
        settings: {
          mode: roomMode,
          showGuesses,
          maxPlayers,
          ...(roomMode === 'marathon' ? { timeLimitSecs } : {}),
        },
      })
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
      useToastStore.getState().addToast('Failed to login as guest', 'error')
    }
  }

  const handleLogout = () => {
    logout()
    clearDailyStatus()
  }

  const getDailyButtonText = () => {
    if (!dailyStatusFetched && isAuthenticated && !isGuest) return 'Daily...'
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
              <div className="setting-group">
                <label className="setting-label">Game Mode</label>
                <div className="setting-tabs">
                  <button
                    className={`setting-tab ${roomMode === 'single_round' ? 'active' : ''}`}
                    onClick={() => setRoomMode('single_round')}
                  >
                    Single Round
                  </button>
                  <button
                    className={`setting-tab ${roomMode === 'marathon' ? 'active' : ''}`}
                    onClick={() => setRoomMode('marathon')}
                  >
                    Marathon
                  </button>
                </div>
                <p className="setting-hint">
                  {roomMode === 'single_round'
                    ? 'Everyone guesses the same word'
                    : 'Solve as many words as you can before time runs out'}
                </p>
              </div>

              {roomMode === 'marathon' && (
                <div className="setting-group">
                  <label className="setting-label">Time Limit</label>
                  <div className="setting-tabs">
                    {[60, 120, 180, 300].map((secs) => (
                      <button
                        key={secs}
                        className={`setting-tab ${timeLimitSecs === secs ? 'active' : ''}`}
                        onClick={() => setTimeLimitSecs(secs)}
                      >
                        {secs >= 60 ? `${secs / 60}m` : `${secs}s`}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              <div className="setting-group">
                <label className="setting-label">Max Players</label>
                <div className="setting-tabs">
                  {[2, 3, 4, 5, 6].map((n) => (
                    <button
                      key={n}
                      className={`setting-tab ${maxPlayers === n ? 'active' : ''}`}
                      onClick={() => setMaxPlayers(n)}
                    >
                      {n}
                    </button>
                  ))}
                </div>
              </div>

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
