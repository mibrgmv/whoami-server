import type {
  TokenResponse,
  GuestAuthResponse,
  LoginRequest,
  RegisterRequest,
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

const API_BASE = '/api/v1'

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

      const response = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken }),
      })

      if (!response.ok) {
        throw new Error('Refresh token failed')
      }

      const data: { accessToken: string; refreshToken: string } = await response.json()

      store.setTokens(data.accessToken, data.refreshToken, store.isGuest)

      return data.accessToken
    } catch (error) {
      const { useAuthStore } = await import('../stores/authStore')
      const store = useAuthStore.getState()
      await store.logout()
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

  const response = await fetch(`${API_BASE}${endpoint}`, {
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

// Auth API
export const auth = {
  login: (data: LoginRequest): Promise<TokenResponse> =>
    request('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  register: (data: RegisterRequest): Promise<{ id: string; message: string }> =>
    request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  guest: (): Promise<GuestAuthResponse> =>
    request('/auth/guest', {
      method: 'POST',
    }),

  refresh: (refreshToken: string): Promise<TokenResponse> =>
    request('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
    }),

  logout: (refreshToken: string): Promise<void> =>
    request('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
    }),
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
    return `${protocol}//${host}/api/v1/rooms/${code}/ws?player_id=${playerId}`
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
