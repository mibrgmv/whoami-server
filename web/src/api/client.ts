import type {
  TokenResponse,
  LoginRequest,
  RegisterRequest,
  GameSession,
  GuessResult,
  DailyStatus,
  CreateRoomRequest,
  Room,
  RoomResponse,
  RoomPlayer,
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

async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = localStorage.getItem('access_token')

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

  guest: (): Promise<TokenResponse> =>
    request('/auth/guest', {
      method: 'POST',
    }),

  refresh: (refresh_token: string): Promise<TokenResponse> =>
    request('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token }),
    }),

  logout: (refresh_token: string): Promise<void> =>
    request('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refresh_token }),
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

export { ApiError }
