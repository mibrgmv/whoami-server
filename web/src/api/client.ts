import { config } from '../config'
import type {
  GuestAuthResponse,
  GameSession,
  GuessResult,
  DailyStatus,
  CreateRoomRequest,
  Room,
  RoomResponse,
  RoomPlayer,
  UserStatistics,
  GameHistoryResponse,
} from '../types/api'

class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

let isRefreshing = false
let refreshPromise: Promise<string> | null = null

async function refreshAccessToken(): Promise<string> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise
  }

  isRefreshing = true
  refreshPromise = (async () => {
    try {
      const { useAuthStore } = await import('../stores/authStore')
      const store = useAuthStore.getState()

      const { refreshToken } = store
      if (!refreshToken) {
        throw new Error('No refresh token available')
      }

      const data = await auth.refresh(refreshToken)
      store.setTokens(data.access_token, data.refresh_token, store.isGuest)
      return data.access_token
    } catch (error) {
      const { useAuthStore } = await import('../stores/authStore')
      const store = useAuthStore.getState()
      store.clearAuth()
      throw error
    } finally {
      isRefreshing = false
      refreshPromise = null
    }
  })()

  return refreshPromise
}

function getAccessToken(): string | null {
  const stored = localStorage.getItem('auth-storage')
  if (!stored) return null
  try {
    const parsed = JSON.parse(stored)
    return parsed.state?.accessToken || null
  } catch {
    return null
  }
}

async function request<T>(
  endpoint: string,
  options: RequestInit = {},
  isRetry = false
): Promise<T> {
  const token = getAccessToken()

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  }

  if (token) {
    ;(headers as Record<string, string>)['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${config.apiBase}${endpoint}`, {
    ...options,
    headers,
  })

  if (response.status === 401 && !isRetry && !endpoint.includes('/auth/')) {
    try {
      await refreshAccessToken()
      return request<T>(endpoint, options, true)
    } catch (error) {
      const errorData = await response.json().catch(() => ({ error: 'Unauthorized' }))
      throw new ApiError(response.status, errorData.error || errorData.message || 'Unauthorized')
    }
  }

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }))
    throw new ApiError(response.status, error.error || error.message || 'Request failed')
  }

  if (response.status === 204) {
    return {} as T
  }

  return response.json()
}

async function keycloakPost<T>(path: string, body: URLSearchParams): Promise<T> {
  const response = await fetch(`${config.keycloak.oidcBase}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body,
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error_description || error.error || 'Keycloak request failed')
  }

  if (response.status === 204) {
    return {} as T
  }

  return response.json()
}

// Auth API
export const auth = {
  guest: (): Promise<GuestAuthResponse> =>
    request('/auth/guest', { method: 'POST' }),

  exchangeCode: (code: string): Promise<{ access_token: string; refresh_token: string }> =>
    keycloakPost('/token', new URLSearchParams({
      grant_type: 'authorization_code',
      client_id: config.keycloak.clientId,
      code,
      redirect_uri: `${window.location.origin}/oauth/callback`,
    })),

  refresh: (refreshToken: string): Promise<{ access_token: string; refresh_token: string }> =>
    keycloakPost('/token', new URLSearchParams({
      grant_type: 'refresh_token',
      client_id: config.keycloak.clientId,
      refresh_token: refreshToken,
    })),

  logout: async (refreshToken: string): Promise<void> => {
    await keycloakPost('/logout', new URLSearchParams({
      client_id: config.keycloak.clientId,
      refresh_token: refreshToken,
    }))
  },
}

// Game API
export const game = {
  start: (mode: 'daily' | 'random', language = 'en'): Promise<GameSession> =>
    request('/games', {
      method: 'POST',
      body: JSON.stringify({ mode: `GAME_MODE_${mode.toUpperCase()}`, language }),
    }),

  get: (sessionId: string): Promise<GameSession> =>
    request(`/games/${sessionId}`),

  guess: (sessionId: string, word: string): Promise<GuessResult> =>
    request(`/games/${sessionId}/guesses`, {
      method: 'POST',
      body: JSON.stringify({ word }),
    }),

  dailyStatus: (language = 'en'): Promise<DailyStatus> =>
    request(`/games/daily/status?language=${language}`),

  validateWord: (word: string, language = 'en'): Promise<{ is_valid: boolean }> =>
    request(`/games/validate/${word}?language=${language}`),
}

// Room API
export const room = {
  create: (data: CreateRoomRequest = {}): Promise<{ room: Room }> =>
    request('/rooms', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  get: (code: string): Promise<RoomResponse> =>
    request(`/rooms/${code}`),

  join: (code: string, display_name: string): Promise<{ room: Room; player: RoomPlayer }> =>
    request(`/rooms/${code}/join`, {
      method: 'POST',
      body: JSON.stringify({ display_name }),
    }),

  wsUrl: (code: string, playerId: string): string => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    return `${protocol}//${host}${config.apiBase}/rooms/${code}/ws?player_id=${playerId}`
  },
}

// Statistics API
export const statistics = {
  getMyStats: (): Promise<UserStatistics> =>
    request('/statistics/me'),

  getMyHistory: (pageSize = 10, pageToken = ''): Promise<GameHistoryResponse> => {
    const params = new URLSearchParams()
    if (pageSize) params.set('pageSize', String(pageSize))
    if (pageToken) params.set('pageToken', pageToken)
    return request(`/statistics/me/history?${params}`)
  },
}

export { ApiError }
