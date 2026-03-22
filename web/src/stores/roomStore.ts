import { create } from 'zustand'
import { room as roomApi } from '../api/client'
import type { Room, RoomPlayer, LetterResult, CreateRoomRequest } from '../types/api'
import { RoomStatusValues, PlayerStatusValues } from '../types/api'
import { updateLetterStates } from '../utils/letterStates'
import { useToastStore } from './toastStore'
import type {
  WSEvent,
  PlayerJoinedPayload,
  PlayerLeftPayload,
  PlayerReadyPayload,
  GameStartedPayload,
  PlayerAttemptPayload,
  PlayerGuessPayload,
  RoundEndedPayload,
  GameEndedPayload,
  ErrorPayload,
} from '../types/ws'

function resultToEmoji(result: string): string {
  return result.split('').map(r => {
    if (r === 'G') return '🟩'
    if (r === 'Y') return '🟨'
    return '⬜'
  }).join('')
}

const MAX_RECONNECT_ATTEMPTS = 5

interface RoomState {
  room: Room | null
  players: RoomPlayer[]
  currentPlayer: RoomPlayer | null
  currentGuess: string
  letterStates: Record<string, LetterResult>
  wordLength: number
  isLoading: boolean
  gameResult: RoundEndedPayload | GameEndedPayload | null

  // WebSocket
  ws: WebSocket | null
  isWsConnected: boolean
  wsReconnectAttempts: number

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
  gameResult: null,
  ws: null,
  isWsConnected: false,
  wsReconnectAttempts: 0,

  createRoom: async (settings) => {
    set({ isLoading: true })
    try {
      const response = await roomApi.create(settings)
      const code = response.room.code

      const fullRoom = await roomApi.get(code)
      const currentPlayer = fullRoom.players.find(
        (p) => p.playerId === response.room.hostId
      ) || null

      if (currentPlayer) {
        localStorage.setItem(`room_${code}_player`, currentPlayer.playerId)
      }

      set({
        room: fullRoom.room,
        players: fullRoom.players,
        currentPlayer,
        isLoading: false,
      })

      return code
    } catch (err) {
      useToastStore.getState().addToast((err as Error).message, 'error')
      set({ isLoading: false })
      throw err
    }
  },

  joinRoom: async (code, displayName) => {
    set({ isLoading: true })
    try {
      const response = await roomApi.join(code, displayName)

      localStorage.setItem(`room_${code}_player`, response.player.playerId)

      set({
        room: response.room,
        currentPlayer: response.player,
        isLoading: false,
      })

      const fullRoom = await roomApi.get(code)
      set({ players: fullRoom.players })
    } catch (err) {
      useToastStore.getState().addToast((err as Error).message, 'error')
      set({ isLoading: false })
      throw err
    }
  },

  connectWebSocket: (code, playerId) => {
    const { ws: existingWs } = get()

    // Don't connect if there's already an open or connecting socket
    if (existingWs && (existingWs.readyState === WebSocket.OPEN || existingWs.readyState === WebSocket.CONNECTING)) {
      return
    }

    // Clean up any closing socket without triggering reconnect
    if (existingWs) {
      existingWs.onclose = null
      existingWs.onerror = null
      existingWs.close()
    }

    const url = roomApi.wsUrl(code, playerId)
    const ws = new WebSocket(url)

    ws.onopen = () => {
      set({ isWsConnected: true, wsReconnectAttempts: 0 })
    }

    ws.onmessage = (event) => {
      try {
        const wsEvent = JSON.parse(event.data) as WSEvent
        get().handleWSMessage(wsEvent)
      } catch {
        // Ignore malformed messages
      }
    }

    ws.onerror = () => {
      // Only show error if this is still the active socket
      if (get().ws === ws) {
        set({ isWsConnected: false })
      }
    }

    ws.onclose = (event) => {
      // Ignore if this is a stale socket
      if (get().ws !== ws) return

      set({ ws: null, isWsConnected: false })

      if (event.code !== 1000) {
        const attempts = get().wsReconnectAttempts
        if (attempts >= MAX_RECONNECT_ATTEMPTS) {
          useToastStore.getState().addToast('Connection lost. Please refresh the page.', 'error')
          return
        }
        const delay = Math.min(2000 * Math.pow(2, attempts), 30000)
        set({ wsReconnectAttempts: attempts + 1 })
        setTimeout(() => {
          const state = get()
          if (!state.isWsConnected && !state.ws && state.room) {
            get().connectWebSocket(state.room.code, playerId)
          }
        }, delay)
      }
    }

    set({ ws })
  },

  disconnect: () => {
    const { ws } = get()
    if (ws) {
      ws.close()
      set({ ws: null, isWsConnected: false, wsReconnectAttempts: 0 })
    }
  },

  handleWSMessage: (event) => {
    switch (event.type) {
      case 'player_joined': {
        const payload = event.payload as PlayerJoinedPayload
        useToastStore.getState().addToast(`${payload.player.displayName} joined`)
        set((state) => ({ players: [...state.players, payload.player] }))
        break
      }

      case 'player_left': {
        const payload = event.payload as PlayerLeftPayload
        useToastStore.getState().addToast(`${payload.displayName} left`)
        set((state) => ({
          players: state.players.filter((p) => p.playerId !== payload.playerId),
          currentPlayer: state.currentPlayer?.playerId === payload.playerId ? null : state.currentPlayer,
        }))
        break
      }

      case 'player_ready': {
        const payload = event.payload as PlayerReadyPayload
        set((state) => {
          const newStatus = payload.ready ? PlayerStatusValues.READY : PlayerStatusValues.WAITING
          const updatedPlayers = state.players.map((p) =>
            p.playerId === payload.playerId ? { ...p, status: newStatus } : p
          )
          const isCurrentPlayer = state.currentPlayer?.playerId === payload.playerId
          return {
            players: updatedPlayers,
            currentPlayer: isCurrentPlayer && state.currentPlayer
              ? { ...state.currentPlayer, status: newStatus }
              : state.currentPlayer,
          }
        })
        break
      }

      case 'game_started': {
        const payload = event.payload as GameStartedPayload
        set((state) => {
          const updatedPlayers = state.players.map((p) => ({
            ...p,
            status: PlayerStatusValues.PLAYING,
            guesses: [],
            currentAttempts: 0,
          }))
          return {
            wordLength: payload.wordLength,
            currentGuess: '',
            letterStates: {},
            gameResult: null,
            players: updatedPlayers,
            currentPlayer: state.currentPlayer
              ? { ...state.currentPlayer!, status: PlayerStatusValues.PLAYING, guesses: [], currentAttempts: 0 }
              : null,
            room: state.room ? { ...state.room, status: RoomStatusValues.PLAYING, roundNumber: payload.roundNumber } : null,
          }
        })
        break
      }

      case 'player_attempt': {
        const payload = event.payload as PlayerAttemptPayload
        set((state) => {
          const isCurrentPlayer = state.currentPlayer?.playerId === payload.playerId
          const updatedPlayers = state.players.map((p) =>
            p.playerId === payload.playerId
              ? { ...p, currentAttempts: payload.attempts }
              : p
          )
          return {
            players: updatedPlayers,
            currentPlayer: isCurrentPlayer && state.currentPlayer
              ? { ...state.currentPlayer, currentAttempts: payload.attempts }
              : state.currentPlayer,
          }
        })
        break
      }

      case 'player_guess': {
        const payload = event.payload as PlayerGuessPayload
        const state = get()
        const isCurrentPlayer = state.currentPlayer?.playerId === payload.playerId

        if (!isCurrentPlayer) {
          const emoji = resultToEmoji(payload.result)
          useToastStore.getState().addToast(`${payload.displayName}  ${emoji}`)
        }

        const newGuess = {
          word: payload.guessWord,
          attemptNumber: payload.attempts,
          results: payload.result.split('').map(r => {
            if (r === 'G') return 'LETTER_RESULT_CORRECT' as const
            if (r === 'Y') return 'LETTER_RESULT_PRESENT' as const
            return 'LETTER_RESULT_ABSENT' as const
          })
        }

        set((state) => {
          const updatedPlayers = state.players.map((p) =>
            p.playerId === payload.playerId
              ? { ...p, guesses: [...p.guesses, newGuess] }
              : p
          )

          const newLetterStates = isCurrentPlayer
            ? updateLetterStates(state.letterStates, newGuess)
            : state.letterStates

          return {
            players: updatedPlayers,
            currentPlayer: isCurrentPlayer && state.currentPlayer
              ? { ...state.currentPlayer, guesses: [...state.currentPlayer.guesses, newGuess] }
              : state.currentPlayer,
            letterStates: newLetterStates,
            currentGuess: isCurrentPlayer ? '' : state.currentGuess,
            isLoading: false,
          }
        })
        break
      }

      case 'round_ended': {
        const payload = event.payload as RoundEndedPayload
        set((state) => {
          const updatedPlayers = state.players.map((p) => ({ ...p, status: PlayerStatusValues.FINISHED }))
          return {
            gameResult: payload,
            players: updatedPlayers,
            currentPlayer: state.currentPlayer
              ? { ...state.currentPlayer!, status: PlayerStatusValues.FINISHED }
              : null,
          }
        })
        break
      }

      case 'game_ended': {
        const payload = event.payload as GameEndedPayload
        set((state) => ({
          gameResult: payload,
          room: state.room ? { ...state.room, status: RoomStatusValues.FINISHED } : null,
        }))
        break
      }

      case 'room_updated': {
        const payload = event.payload as { room: Room }
        set({ room: payload.room })
        break
      }

      case 'error': {
        const payload = event.payload as ErrorPayload
        useToastStore.getState().addToast(payload.message, 'error')
        set({ isLoading: false })
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
      set({ isLoading: true })
    }
  },

  nextRound: () => {
    const { ws } = get()
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'next_round' }))
    }
  },

  leaveRoom: () => {
    const { ws, room } = get()
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'leave' }))
    }
    if (room?.code) {
      localStorage.removeItem(`room_${room.code}_player`)
    }
    get().disconnect()
    get().reset()
  },

  // Local actions
  addLetter: (letter) => {
    const { currentGuess, wordLength, room, gameResult } = get()
    if (room?.status !== RoomStatusValues.PLAYING || gameResult) return
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

  reset: () => {
    const { ws } = get()
    if (ws) {
      ws.close()
    }
    set({
      room: null,
      players: [],
      currentPlayer: null,
      currentGuess: '',
      letterStates: {},
      wordLength: 5,
      isLoading: false,
      gameResult: null,
      ws: null,
      isWsConnected: false,
      wsReconnectAttempts: 0,
    })
  },
}))
