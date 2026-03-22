import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore, redirectToLogin } from '../stores/authStore'
import { useRoomStore } from '../stores/roomStore'
import { room as roomApi } from '../api/client'
import { useGameStore } from '../stores/gameStore'
import { useToastStore } from '../stores/toastStore'
import { GameStatusValues } from '../types/api'
import './Home.css'

export function Home() {
  const navigate = useNavigate()
  const { isAuthenticated, isGuest, username, hasHydrated, loginAsGuest, logout } = useAuthStore()
  const { createRoom, isLoading } = useRoomStore()
  const { dailyStatus, dailyStatusFetched, fetchDailyStatus, clearDailyStatus, randomStatus, randomStatusFetched, fetchRandomStatus } = useGameStore()

  const [roomCode, setRoomCode] = useState('')
  const [isJoining, setIsJoining] = useState(false)
  const [showRoomSettings, setShowRoomSettings] = useState(false)
  const [showAuth, setShowAuth] = useState(false)
  const [showGuesses, setShowGuesses] = useState(true)
  const [roomMode, setRoomMode] = useState<'single_round' | 'marathon'>('single_round')
  const [maxPlayers, setMaxPlayers] = useState(6)
  const [timeLimitSecs, setTimeLimitSecs] = useState(180)

  useEffect(() => {
    if (hasHydrated && isAuthenticated && !isGuest && !dailyStatusFetched) {
      fetchDailyStatus()
    }
  }, [hasHydrated, isAuthenticated, isGuest, dailyStatusFetched, fetchDailyStatus])

  useEffect(() => {
    if (hasHydrated && isAuthenticated && !randomStatusFetched) {
      fetchRandomStatus()
    }
  }, [hasHydrated, isAuthenticated, randomStatusFetched, fetchRandomStatus])

  const handlePlayDaily = () => {
    if (!isAuthenticated || isGuest) return
    navigate('/game/daily')
  }

  const handlePlayRandom = () => {
    if (!isAuthenticated) return
    navigate('/game')
  }

  const isDailyDisabled = !isAuthenticated || isGuest

  const handleCreateRoom = () => {
    if (!isAuthenticated) {
      setShowAuth(true)
      return
    }
    setShowRoomSettings(!showRoomSettings)
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

  const handleJoinRoom = async () => {
    const code = roomCode.trim().toUpperCase()
    if (!code) return
    setIsJoining(true)
    try {
      await roomApi.get(code)
      navigate(`/room/${code}`)
    } catch {
      useToastStore.getState().addToast('Room not found', 'error')
    } finally {
      setIsJoining(false)
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

  const getDailyStatus = () => {
    if (!dailyStatusFetched && isAuthenticated && !isGuest) return null
    if (!dailyStatus?.hasPlayedToday) return null
    if (dailyStatus.status === GameStatusValues.WON) return 'won'
    if (dailyStatus.status === GameStatusValues.LOST) return 'lost'
    if (dailyStatus.status === GameStatusValues.IN_PROGRESS) return 'continue'
    return null
  }

  const dailyDesc = () => {
    const s = getDailyStatus()
    if (s === 'continue') return 'Continue'
    if (s === 'won' || s === 'lost') return 'Come back tomorrow'
    return 'New word everyday'
  }

  return (
    <div className="home">
      <h1 className="home-title">GORDLE</h1>
      <p className="home-subtitle">Multiplayer word game</p>

      <div className="home-cards">

        {/* Multiplayer card — prominent */}
        <div className="card mp-card">
          <div className="card-title">Multiplayer</div>
          <div className="mp-actions">
            <button
              className="btn btn-primary"
              onClick={handleCreateRoom}
              disabled={!isAuthenticated || isGuest || isLoading}
            >
              {isLoading ? 'Creating...' : 'Create Room'}
            </button>
          </div>

          <div className="join-room">
            <input
              type="text"
              placeholder="Room code"
              value={roomCode}
              onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
              onKeyDown={(e) => e.key === 'Enter' && handleJoinRoom()}
              maxLength={6}
              className="input"
            />
            <button
              className="btn btn-secondary"
              onClick={handleJoinRoom}
              disabled={!isAuthenticated || !roomCode.trim() || isJoining}
            >
              {isJoining ? 'Checking...' : 'Join'}
            </button>
          </div>

          {/* Room settings — inline expand */}
          <div className={`collapsible ${showRoomSettings ? 'collapsible--open' : ''}`}>
            <div className="collapsible-inner">
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

                <div className="room-settings-actions">
                  <button
                    className="btn btn-primary"
                    onClick={handleConfirmCreateRoom}
                    disabled={isLoading}
                  >
                    {isLoading ? 'Creating...' : 'Start Room'}
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
          </div>
        </div>

        {/* Solo cards row */}
        <div className="solo-row">
          <div
            className={`card solo-card ${isDailyDisabled ? 'solo-card--disabled' : ''} ${getDailyStatus() === 'continue' ? 'solo-card--active' : ''}`}
            onClick={handlePlayDaily}
          >
            <div className="solo-card-name">Daily</div>
            <div className="solo-card-desc">{dailyDesc()}</div>
          </div>

          <div
            className={`card solo-card ${!isAuthenticated ? 'solo-card--disabled' : ''} ${randomStatus === GameStatusValues.IN_PROGRESS ? 'solo-card--active' : ''}`}
            onClick={handlePlayRandom}
          >
            <div className="solo-card-name">Random</div>
            <div className="solo-card-desc">
              {randomStatus === GameStatusValues.IN_PROGRESS ? 'Continue' : 'Practice with infinite words'}
            </div>
          </div>
        </div>

        {/* Auth card */}
        <div className="card auth-card" onClick={() => setShowAuth(!showAuth)}>
          <div className="auth-card-label">
            {isAuthenticated
              ? (isGuest ? 'Playing as Guest' : username)
              : 'Log in'}
          </div>
          <div className={`collapsible ${showAuth ? 'collapsible--open' : ''}`}>
            <div className="collapsible-inner">
              <div className="auth-card-actions" onClick={(e) => e.stopPropagation()}>
                {!isAuthenticated && (
                  <>
                    <button className="btn btn-primary" onClick={handleGuestLogin}>
                      Play as Guest
                    </button>
                    <button className="btn btn-secondary" onClick={() => redirectToLogin()}>
                      Log in
                    </button>
                  </>
                )}
                {isAuthenticated && isGuest && (
                  <>
                    <button className="btn btn-secondary" onClick={() => redirectToLogin()}>
                      Log in
                    </button>
                    <button className="btn btn-secondary" onClick={handleLogout}>
                      Log out
                    </button>
                  </>
                )}
                {isAuthenticated && !isGuest && (
                  <>
                    <button className="btn btn-secondary" onClick={() => navigate('/profile')}>
                      Profile
                    </button>
                    <button className="btn btn-secondary" onClick={handleLogout}>
                      Log out
                    </button>
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
