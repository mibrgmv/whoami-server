import { create } from 'zustand'
import { room as roomApi } from '../api/client'
import type { Room, RoomPlayer, LetterResult, CreateRoomRequest } from '../types/api'
import type {
  WSEvent,
  PlayerJoinedPayload,
  PlayerLeftPayload,
  PlayerReadyPayload,
  GameStartedPayload,
  PlayerGuessPayload,
  RoundEndedPayload,
  GameEndedPayload,
  ErrorPayload,
} from '../types/ws'

interface RoomState {
  room: Room | null
  players: RoomPlayer[]
  currentPlayer: RoomPlayer | null
  currentGuess: string
  letterStates: Record<string, LetterResult>
  wordLength: number
  isLoading: boolean
  error: string | null
  gameResult: RoundEndedPayload | GameEndedPayload | null

  // WebSocket
  ws: WebSocket | null

  // Actions
  createRoom: (settings?: CreateRoomRequest) => Promise<string>
  joinRoom: (code: string, displayName: string) => Promise<void>
  connectWebSocket: (code: string, playerId: string) => void
  disconnect: () => void

  // Game actions (sent via WebSocket)
  setReady: (ready: boolean) => void
  startGame: () => void
  submitGuess: () => void
  nextRound: () => void
  leaveRoom: () => void

  // Local actions
  addLetter: (letter: string) => void
  removeLetter: () => void
  clearError: () => void
  reset: () => void

  // Internal
  handleWSMessage: (event: WSEvent) => void
}

export const useRoomStore = create<RoomState>((set, get) => ({
  room: null,
  players: [],
  currentPlayer: null,
  currentGuess: '',
  letterStates: {},
  wordLength: 5,
  isLoading: false,
  error: null,
  gameResult: null,
  ws: null,

  createRoom: async (settings) => {
    set({ isLoading: true, error: null })
    try {
      const response = await roomApi.create(settings)
      set({ room: response.room, isLoading: false })
      return response.room.code
    } catch (err) {
      set({ error: (err as Error).message, isLoading: false })
      throw err
    }
  },

  joinRoom: async (code, displayName) => {
    set({ isLoading: true, error: null })
    try {
      const response = await roomApi.join(code, displayName)
      set({
        room: response.room,
        currentPlayer: response.player,
        isLoading: false,
      })

      // Fetch full room data with all players
      const fullRoom = await roomApi.get(code)
      set({ players: fullRoom.players })
    } catch (err) {
      set({ error: (err as Error).message, isLoading: false })
      throw err
    }
  },

  connectWebSocket: (code, playerId) => {
    const { ws: existingWs } = get()
    if (existingWs) {
      existingWs.close()
    }

    const url = roomApi.wsUrl(code, playerId)
    const ws = new WebSocket(url)

    ws.onopen = () => {
      console.log('WebSocket connected')
    }

    ws.onmessage = (event) => {
      try {
        const wsEvent = JSON.parse(event.data) as WSEvent
        get().handleWSMessage(wsEvent)
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      set({ error: 'Connection error' })
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      set({ ws: null })
    }

    set({ ws })
  },

  disconnect: () => {
    const { ws } = get()
    if (ws) {
      ws.close()
      set({ ws: null })
    }
  },

  handleWSMessage: (event) => {
    const { players, currentPlayer } = get()

    switch (event.type) {
      case 'player_joined': {
        const payload = event.payload as PlayerJoinedPayload
        set({ players: [...players, payload.player] })
        break
      }

      case 'player_left': {
        const payload = event.payload as PlayerLeftPayload
        set({ players: players.filter((p) => p.player_id !== payload.player_id) })
        break
      }

      case 'player_ready': {
        const payload = event.payload as PlayerReadyPayload
        set({
          players: players.map((p) =>
            p.player_id === payload.player_id
              ? { ...p, status: payload.ready ? 'ready' : 'waiting' }
              : p
          ),
        })
        break
      }

      case 'game_started': {
        const payload = event.payload as GameStartedPayload
        set({
          wordLength: payload.word_length,
          currentGuess: '',
          letterStates: {},
          gameResult: null,
          players: players.map((p) => ({ ...p, status: 'playing', guesses: [] })),
          room: get().room ? { ...get().room!, status: 'playing', round_number: payload.round_number } : null,
        })
        break
      }

      case 'player_guess': {
        const payload = event.payload as PlayerGuessPayload
        set({
          players: players.map((p) =>
            p.player_id === payload.player_id
              ? { ...p, current_attempts: payload.attempts }
              : p
          ),
        })

        // If it's our own guess, update letter states
        if (currentPlayer && payload.player_id === currentPlayer.player_id && payload.result) {
          // Parse result to update letter states
          // Result format depends on backend implementation
        }
        break
      }

      case 'round_ended': {
        const payload = event.payload as RoundEndedPayload
        set({
          gameResult: payload,
          players: players.map((p) => ({ ...p, status: 'finished' })),
        })
        break
      }

      case 'game_ended': {
        const payload = event.payload as GameEndedPayload
        set({
          gameResult: payload,
          room: get().room ? { ...get().room!, status: 'finished' } : null,
        })
        break
      }

      case 'room_updated': {
        const payload = event.payload as { room: Room }
        set({ room: payload.room })
        break
      }

      case 'error': {
        const payload = event.payload as ErrorPayload
        set({ error: payload.message })
        break
      }
    }
  },

  // WebSocket actions
  setReady: (ready) => {
    const { ws } = get()
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'ready', payload: { ready } }))
    }
  },

  startGame: () => {
    const { ws } = get()
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'start_game' }))
    }
  },

  submitGuess: () => {
    const { ws, currentGuess, wordLength } = get()
    if (ws && ws.readyState === WebSocket.OPEN && currentGuess.length === wordLength) {
      ws.send(JSON.stringify({ type: 'guess', payload: { word: currentGuess.toLowerCase() } }))
      set({ currentGuess: '' })
    }
  },

  nextRound: () => {
    const { ws } = get()
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'next_round' }))
    }
  },

  leaveRoom: () => {
    const { ws } = get()
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'leave' }))
    }
    get().disconnect()
    get().reset()
  },

  // Local actions
  addLetter: (letter) => {
    const { currentGuess, wordLength, room } = get()
    if (room?.status !== 'playing') return
    if (currentGuess.length < wordLength) {
      set({ currentGuess: currentGuess + letter.toUpperCase() })
    }
  },

  removeLetter: () => {
    const { currentGuess } = get()
    if (currentGuess.length > 0) {
      set({ currentGuess: currentGuess.slice(0, -1) })
    }
  },

  clearError: () => set({ error: null }),

  reset: () => set({
    room: null,
    players: [],
    currentPlayer: null,
    currentGuess: '',
    letterStates: {},
    wordLength: 5,
    isLoading: false,
    error: null,
    gameResult: null,
    ws: null,
  }),
}))
