import { create } from 'zustand'
import { game } from '../api/client'
import type { GameSession, LetterResult, GameStatus, DailyStatus } from '../types/api'
import { GameStatusValues } from '../types/api'
import { updateLetterStates } from '../utils/letterStates'
import { useToastStore } from './toastStore'

interface GameState {
  session: GameSession | null
  currentGuess: string
  isLoading: boolean
  letterStates: Record<string, LetterResult>

  dailyStatus: DailyStatus | null
  dailyStatusFetched: boolean
  randomStatus: GameStatus | null
  randomStatusFetched: boolean

  startGame: (mode: 'daily' | 'random', language?: string) => Promise<void>
  loadGame: (sessionId: string) => Promise<void>
  setCurrentGuess: (guess: string) => void
  addLetter: (letter: string) => void
  removeLetter: () => void
  submitGuess: () => Promise<void>
  reset: () => void
  fetchDailyStatus: () => Promise<void>
  fetchRandomStatus: () => Promise<void>
  clearDailyStatus: () => void
  clearRandomSession: () => void
}

const RANDOM_SESSION_KEY = 'gordle_random_session'

const WORD_LENGTH = 5

const getRandomSessionId = () => localStorage.getItem(RANDOM_SESSION_KEY)
const saveRandomSessionId = (id: string) => localStorage.setItem(RANDOM_SESSION_KEY, id)
const removeRandomSessionId = () => localStorage.removeItem(RANDOM_SESSION_KEY)

export const useGameStore = create<GameState>((set, get) => ({
  session: null,
  currentGuess: '',
  isLoading: false,
  letterStates: {},
  dailyStatus: null,
  dailyStatusFetched: false,
  randomStatus: null,
  randomStatusFetched: false,

  startGame: async (mode, language = 'en') => {
    set({ isLoading: true })
    try {
      // For random mode, try to resume an existing session first
      if (mode === 'random') {
        const savedId = getRandomSessionId()
        if (savedId) {
          try {
            const session = await game.get(savedId)
            if (session.status === GameStatusValues.IN_PROGRESS) {
              let letterStates: Record<string, LetterResult> = {}
              for (const guess of session.guesses) {
                letterStates = updateLetterStates(letterStates, guess)
              }
              set({ session, letterStates, currentGuess: '', isLoading: false })
              return
            }
            get().clearRandomSession()
            set({ randomStatus: session.status })
          } catch {
            get().clearRandomSession()
          }
        }
      }

      const session = await game.start(mode, language)

      if (mode === 'random') {
        saveRandomSessionId(session.sessionId)
        set({ randomStatus: session.status })
      }

      set({
        session,
        currentGuess: '',
        letterStates: {},
        isLoading: false,
      })
    } catch (err) {
      useToastStore.getState().addToast((err as Error).message, 'error', 1500)
      set({ isLoading: false })
    }
  },

  loadGame: async (sessionId) => {
    set({ isLoading: true })
    try {
      const session = await game.get(sessionId)

      let letterStates: Record<string, LetterResult> = {}
      for (const guess of session.guesses) {
        letterStates = updateLetterStates(letterStates, guess)
      }

      set({ session, letterStates, currentGuess: '', isLoading: false })
    } catch (err) {
      useToastStore.getState().addToast((err as Error).message, 'error', 1500)
      set({ isLoading: false })
    }
  },

  setCurrentGuess: (guess) => {
    if (guess.length <= WORD_LENGTH) {
      set({ currentGuess: guess.toUpperCase() })
    }
  },

  addLetter: (letter) => {
    const { currentGuess, session } = get()
    if (session?.status !== GameStatusValues.IN_PROGRESS) return
    if (currentGuess.length < WORD_LENGTH) {
      set({ currentGuess: currentGuess + letter.toUpperCase() })
    }
  },

  removeLetter: () => {
    const { currentGuess } = get()
    if (currentGuess.length > 0) {
      set({ currentGuess: currentGuess.slice(0, -1) })
    }
  },

  submitGuess: async () => {
    const { session, currentGuess } = get()
    if (!session || currentGuess.length !== WORD_LENGTH) return

    set({ isLoading: true })
    try {
      const result = await game.guess(session.sessionId, currentGuess.toLowerCase())

      const newGuesses = [...session.guesses, result.guess]
      const newLetterStates = updateLetterStates(get().letterStates, result.guess)

      const gameOver = result.gameStatus === GameStatusValues.WON ||
        result.gameStatus === GameStatusValues.LOST

      if (gameOver && session.gameMode === 'GAME_MODE_RANDOM') {
        get().clearRandomSession()
        set({ randomStatus: result.gameStatus as GameStatus })
      }

      set({
        session: {
          ...session,
          guesses: newGuesses,
          status: result.gameStatus as GameStatus,
          targetWord: result.targetWord,
          attemptsUsed: newGuesses.length,
        },
        letterStates: newLetterStates,
        currentGuess: '',
        isLoading: false,
      })
    } catch (err) {
      useToastStore.getState().addToast((err as Error).message, 'error', 1500)
      set({ isLoading: false })
    }
  },

  reset: () => set({
    session: null,
    currentGuess: '',
    isLoading: false,
    letterStates: {},
    dailyStatusFetched: false,
    randomStatusFetched: false,
  }),

  fetchDailyStatus: async () => {
    try {
      const status = await game.dailyStatus()
      set({ dailyStatus: status, dailyStatusFetched: true })
    } catch {
      set({ dailyStatusFetched: true })
    }
  },

  fetchRandomStatus: async () => {
    const savedId = getRandomSessionId()
    if (!savedId) {
      set({ randomStatus: null, randomStatusFetched: true })
      return
    }
    try {
      const session = await game.get(savedId)
      if (session.status === GameStatusValues.IN_PROGRESS) {
        set({ randomStatus: session.status, randomStatusFetched: true })
      } else {
        removeRandomSessionId()
        set({ randomStatus: null, randomStatusFetched: true })
      }
    } catch {
      removeRandomSessionId()
      set({ randomStatus: null, randomStatusFetched: true })
    }
  },

  clearDailyStatus: () => set({ dailyStatus: null, dailyStatusFetched: false }),

  clearRandomSession: () => {
    removeRandomSessionId()
    set({ randomStatus: null })
  },
}))
