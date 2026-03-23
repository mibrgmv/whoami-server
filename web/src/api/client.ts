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

async function request<T>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T> {
  const { useAuthStore } = await import('../stores/authStore')
  const store = useAuthStore.getState()

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  }

  if (store.isGuest && store.accessToken) {
    ;(headers as Record<string, string>)['Authorization'] = `Bearer ${store.accessToken}`
  }

  const response = await fetch(`${config.apiBase}${endpoint}`, {
    ...options,
    headers,
    credentials: 'include',
  })

  if (response.status === 401 && !endpoint.includes('/auth/')) {
    store.clearAuth()
    window.location.href = `${config.apiBase}/auth/login`
    throw new ApiError(401, 'Unauthorized')
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

export const auth = {
  guest: (): Promise<GuestAuthResponse> =>
    request('/auth/guest', { method: 'POST' }),

  me: (): Promise<{ username: string; email: string; isGuest: boolean; isAuthenticated: boolean }> =>
    request('/auth/me'),
}

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
