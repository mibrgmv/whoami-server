import type { Room, RoomPlayer } from './api'

// WebSocket event types (match backend ws_events.go)
export type WSEventType =
  | 'player_joined'
  | 'player_left'
  | 'player_ready'
  | 'game_started'
  | 'player_guess'
  | 'round_ended'
  | 'game_ended'
  | 'room_updated'
  | 'error'

export interface WSEvent<T = unknown> {
  type: WSEventType
  timestamp: string
  payload: T
}

// Payloads
export interface PlayerJoinedPayload {
  player: RoomPlayer
}

export interface PlayerLeftPayload {
  playerId: string
  displayName: string
}

export interface PlayerReadyPayload {
  playerId: string
  displayName: string
  ready: boolean
}

export interface GameStartedPayload {
  roundNumber: number
  wordLength: number
}

export interface PlayerGuessPayload {
  playerId: string
  displayName: string
  guessWord?: string
  result?: string
  attempts: number
  solved: boolean
}

export interface PlayerScore {
  playerId: string
  displayName: string
  result: string
  attempts: number
  score: number
}

export interface RoundEndedPayload {
  roundNumber: number
  targetWord: string
  results: PlayerScore[]
}

export interface GameEndedPayload {
  reason: string
  finalScores: PlayerScore[]
  targetWord?: string
}

export interface RoomUpdatedPayload {
  room: Room
}

export interface ErrorPayload {
  code: string
  message: string
}

// Client -> Server messages
export type WSMessageType = 'guess' | 'ready' | 'start_game' | 'next_round' | 'leave'

export interface WSMessage {
  type: WSMessageType
  payload?: Record<string, unknown>
}

export interface GuessMessage extends WSMessage {
  type: 'guess'
  payload: { word: string }
}

export interface ReadyMessage extends WSMessage {
  type: 'ready'
  payload: { ready: boolean }
}
