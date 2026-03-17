import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { statistics } from '../api/client'
import type { UserStatistics } from '../types/api'
import './Profile.css'

export function Profile() {
  const navigate = useNavigate()
  const { isAuthenticated, isGuest, hasHydrated } = useAuthStore()
  const [stats, setStats] = useState<UserStatistics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!hasHydrated) return

    if (!isAuthenticated || isGuest) {
      navigate('/')
      return
    }

    statistics.getMyStats()
      .then(setStats)
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
        <div className="profile-header">
          <button className="btn btn-text" onClick={() => navigate('/')}>
            Back
          </button>
        </div>
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
      <div className="profile-header">
        <button className="btn btn-text" onClick={() => navigate('/')}>
          Back
        </button>
        <h1>Profile</h1>
        <div />
      </div>

      <div className="profile-content">
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
