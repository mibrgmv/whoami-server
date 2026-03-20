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

  startGame: (mode: 'daily' | 'random', language?: string) => Promise<void>
  loadGame: (sessionId: string) => Promise<void>
  setCurrentGuess: (guess: string) => void
  addLetter: (letter: string) => void
  removeLetter: () => void
  submitGuess: () => Promise<void>
  reset: () => void
  fetchDailyStatus: () => Promise<void>
  clearDailyStatus: () => void
}

export const useGameStore = create<GameState>((set, get) => ({
  session: null,
  currentGuess: '',
  isLoading: false,
  letterStates: {},
  dailyStatus: null,
  dailyStatusFetched: false,

  startGame: async (mode, language = 'en') => {
    set({ isLoading: true })
    try {
      const session = await game.start(mode, language)
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
    if (guess.length <= 5) {
      set({ currentGuess: guess.toUpperCase() })
    }
  },

  addLetter: (letter) => {
    const { currentGuess, session } = get()
    if (session?.status !== GameStatusValues.IN_PROGRESS) return
    if (currentGuess.length < 5) {
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
    if (!session || currentGuess.length !== 5) return

    set({ isLoading: true })
    try {
      const result = await game.guess(session.sessionId, currentGuess.toLowerCase())

      const newGuesses = [...session.guesses, result.guess]
      const newLetterStates = updateLetterStates(get().letterStates, result.guess)

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
  }),

  fetchDailyStatus: async () => {
    try {
      const status = await game.dailyStatus()
      set({ dailyStatus: status, dailyStatusFetched: true })
    } catch {
      set({ dailyStatusFetched: true })
    }
  },

  clearDailyStatus: () => set({ dailyStatus: null, dailyStatusFetched: false }),
}))
