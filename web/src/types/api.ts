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
export type GameMode = 'daily' | 'random'
export type GameStatus = 'in_progress' | 'won' | 'lost'
export type LetterResult = 'correct' | 'present' | 'absent'

export interface Guess {
  word: string
  results: LetterResult[]
  attempt_number: number
}

export interface GameSession {
  session_id: string
  user_id: string
  language: string
  game_mode: GameMode
  game_date: string
  status: GameStatus
  attempts_used: number
  max_attempts: number
  guesses: Guess[]
  target_word?: string
  started_at: string
  completed_at?: string
}

export interface GuessResult {
  guess: Guess
  game_status: GameStatus
  target_word?: string
}

export interface DailyStatus {
  has_played_today: boolean
  session_id?: string
  status?: GameStatus
}

// Room types
export type RoomMode = 'single_round' | 'marathon'
export type RoomStatus = 'waiting' | 'playing' | 'finished'
export type PlayerStatus = 'waiting' | 'ready' | 'playing' | 'finished'
export type PlayerResult = 'won' | 'lost'

export interface RoomSettings {
  mode: RoomMode
  max_players: number
  time_limit_secs: number
  show_guesses: boolean
}

export interface Room {
  id: string
  code: string
  host_id: string
  status: RoomStatus
  settings: RoomSettings
  round_number: number
  language: string
  created_at: string
  expires_at: string
}

export interface RoomPlayer {
  player_id: string
  user_id: string
  guest_id?: string
  display_name: string
  status: PlayerStatus
  result?: PlayerResult
  current_attempts: number
  total_score: number
  guesses: Guess[]
  finished_at?: string
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
