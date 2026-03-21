import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { config } from '../config'
import { useAuthStore } from '../stores/authStore'
import { statistics } from '../api/client'
import type { UserStatistics, GameHistoryItem } from '../types/api'
import './Profile.css'

function keycloakActionUrl(action: string): string {
  const params = new URLSearchParams({
    client_id: config.keycloak.clientId,
    redirect_uri: `${window.location.origin}/oauth/callback`,
    response_type: 'code',
    scope: 'openid',
    kc_action: action,
  })
  return `${config.keycloak.oidcBase}/auth?${params}`
}

function handleKeycloakAction(action: string) {
  sessionStorage.setItem('auth_return_to', '/profile')
  window.location.href = keycloakActionUrl(action)
}

export function Profile() {
  const navigate = useNavigate()
  const { isAuthenticated, isGuest, hasHydrated } = useAuthStore()
  const [stats, setStats] = useState<UserStatistics | null>(null)
  const [history, setHistory] = useState<GameHistoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!hasHydrated) return

    if (!isAuthenticated || isGuest) {
      navigate('/')
      return
    }

    Promise.all([
      statistics.getMyStats(),
      statistics.getMyHistory(10)
    ])
      .then(([statsData, historyData]) => {
        setStats(statsData)
        setHistory(historyData.items || [])
      })
      .catch((err) => setError(err.message || 'Failed to load statistics'))
      .finally(() => setLoading(false))
  }, [hasHydrated, isAuthenticated, isGuest, navigate])

  if (!hasHydrated || loading) {
    return (
      <div className="profile">
        <div className="profile-loading">Loading...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="profile">
        <header className="page-header">
          <button className="page-header-back" onClick={() => navigate('/')}>←</button>
          <h1>Profile</h1>
          <div className="page-header-spacer" />
        </header>
        <div className="profile-error">{error}</div>
      </div>
    )
  }

  const distribution = stats?.guessDistribution
  const maxDistribution = distribution
    ? Math.max(
        distribution.one,
        distribution.two,
        distribution.three,
        distribution.four,
        distribution.five,
        distribution.six,
        1
      )
    : 1

  return (
    <div className="profile">
      <header className="page-header">
        <button className="page-header-back" onClick={() => navigate('/')}>←</button>
        <h1>Profile</h1>
        <div className="page-header-spacer" />
      </header>

      <div className="profile-content">
        <div className="profile-tiles">
          <div className="profile-tile" onClick={() => handleKeycloakAction('UPDATE_PASSWORD')}>
            <div className="profile-tile-name">Password</div>
            <div className="profile-tile-desc">Change</div>
          </div>
          <div className="profile-tile" onClick={() => handleKeycloakAction('UPDATE_EMAIL')}>
            <div className="profile-tile-name">Email</div>
            <div className="profile-tile-desc">Change</div>
          </div>
        </div>

        <div className="stats-cards">
          <div className="stat-card">
            <div className="stat-value">{stats?.gamesPlayed ?? 0}</div>
            <div className="stat-label">Played</div>
          </div>
          <div className="stat-card">
            <div className="stat-value">{stats?.winPercentage ?? 0}%</div>
            <div className="stat-label">Win %</div>
          </div>
          <div className="stat-card">
            <div className="stat-value">{stats?.currentStreak ?? 0}</div>
            <div className="stat-label">Current Streak</div>
          </div>
          <div className="stat-card">
            <div className="stat-value">{stats?.maxStreak ?? 0}</div>
            <div className="stat-label">Max Streak</div>
          </div>
          <div className="stat-card">
            <div className="stat-value">{stats?.averageAttempts ?? 0}</div>
            <div className="stat-label">Avg Attempts</div>
          </div>
        </div>

        <div className="guess-distribution">
          <h2>Guess Distribution</h2>
          <div className="distribution-bars">
            {distribution && (
              <>
                <DistributionBar label="1" value={distribution.one} max={maxDistribution} />
                <DistributionBar label="2" value={distribution.two} max={maxDistribution} />
                <DistributionBar label="3" value={distribution.three} max={maxDistribution} />
                <DistributionBar label="4" value={distribution.four} max={maxDistribution} />
                <DistributionBar label="5" value={distribution.five} max={maxDistribution} />
                <DistributionBar label="6" value={distribution.six} max={maxDistribution} />
              </>
            )}
          </div>
        </div>

        <div className="game-history">
          <h2>Recent Games</h2>
          {history.length === 0 ? (
            <div className="history-empty">No games played yet</div>
          ) : (
            <div className="history-list">
              {history.map((game) => (
                <HistoryItem key={game.historyId} game={game} />
              ))}
            </div>
          )}
        </div>

      </div>
    </div>
  )
}

function DistributionBar({ label, value, max }: { label: string; value: number; max: number }) {
  const percentage = max > 0 ? (value / max) * 100 : 0

  return (
    <div className="distribution-row">
      <div className="distribution-label">{label}</div>
      <div className="distribution-bar-container">
        <div
          className="distribution-bar"
          style={{ width: `${Math.max(percentage, 8)}%` }}
        >
          <span className="distribution-value">{value}</span>
        </div>
      </div>
    </div>
  )
}

function HistoryItem({ game }: { game: GameHistoryItem }) {
  const isWon = game.result === 'won'
  const modeLabel = game.gameMode === 'daily' ? 'Daily'
    : game.gameMode === 'random' ? 'Random'
    : game.gameMode === 'room' ? 'Room'
    : game.gameMode
  const date = new Date(game.gameDate).toLocaleDateString()

  return (
    <div className={`history-item ${isWon ? 'won' : 'lost'}`}>
      <div className="history-item-header">
        <span className="history-mode">{modeLabel}</span>
        <span className="history-date">{date}</span>
      </div>
      <div className="history-item-body">
        <span className="history-word">{game.targetWord.toUpperCase()}</span>
        <span className="history-result">
          {isWon ? `${game.attemptsUsed}/6` : 'X/6'}
        </span>
      </div>
    </div>
  )
}
