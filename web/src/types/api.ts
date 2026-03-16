// Auth types
export interface TokenResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  first_name?: string
  last_name?: string
}

// Game types
export type GameMode = 'GAME_MODE_DAILY' | 'GAME_MODE_RANDOM'
export type GameStatus = 'GAME_STATUS_UNSPECIFIED' | 'GAME_STATUS_IN_PROGRESS' | 'GAME_STATUS_WON' | 'GAME_STATUS_LOST'
export type LetterResult = 'LETTER_RESULT_CORRECT' | 'LETTER_RESULT_PRESENT' | 'LETTER_RESULT_ABSENT'

// Helper constants for easier comparison
export const GameStatusValues = {
  IN_PROGRESS: 'GAME_STATUS_IN_PROGRESS' as GameStatus,
  WON: 'GAME_STATUS_WON' as GameStatus,
  LOST: 'GAME_STATUS_LOST' as GameStatus,
}

export const LetterResultValues = {
  CORRECT: 'LETTER_RESULT_CORRECT' as LetterResult,
  PRESENT: 'LETTER_RESULT_PRESENT' as LetterResult,
  ABSENT: 'LETTER_RESULT_ABSENT' as LetterResult,
}

export interface Guess {
  word: string
  results: LetterResult[]
  attemptNumber: number
}

export interface GameSession {
  sessionId: string
  userId: string
  language: string
  gameMode: GameMode
  gameDate: string
  status: GameStatus
  attemptsUsed: number
  maxAttempts: number
  guesses: Guess[]
  targetWord?: string
  startedAt: string
  completedAt?: string
}

export interface GuessResult {
  guess: Guess
  gameStatus: GameStatus
  targetWord?: string
}

export interface DailyStatus {
  hasPlayedToday: boolean
  sessionId?: string
  status?: GameStatus
}

// Room types
export type RoomMode = 'single_round' | 'marathon'
export type RoomStatus = 'waiting' | 'playing' | 'finished'
export type PlayerStatus = 'waiting' | 'ready' | 'playing' | 'finished'
export type PlayerResult = 'won' | 'lost'

export interface RoomSettings {
  mode: RoomMode
  maxPlayers: number
  timeLimitSecs: number
  showGuesses: boolean
}

export interface Room {
  id: string
  code: string
  hostId: string
  status: RoomStatus
  settings: RoomSettings
  roundNumber: number
  language: string
  createdAt: string
  expiresAt: string
}

export interface RoomPlayer {
  playerId: string
  userId: string
  guestId?: string
  displayName: string
  status: PlayerStatus
  result?: PlayerResult
  currentAttempts: number
  totalScore: number
  guesses: Guess[]
  finishedAt?: string
}

export interface CreateRoomRequest {
  settings?: Partial<RoomSettings>
  language?: string
}

export interface JoinRoomRequest {
  code: string
  display_name: string
}

export interface RoomResponse {
  room: Room
  players: RoomPlayer[]
}
