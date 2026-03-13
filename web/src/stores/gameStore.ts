import { create } from 'zustand'
import { game } from '../api/client'
import type { GameSession, Guess, LetterResult, GameStatus } from '../types/api'

interface GameState {
  session: GameSession | null
  currentGuess: string
  isLoading: boolean
  error: string | null

  // Keyboard state: which letters are in which state
  letterStates: Record<string, LetterResult>

  // Actions
  startGame: (mode: 'daily' | 'random', language?: string) => Promise<void>
  loadGame: (sessionId: string) => Promise<void>
  setCurrentGuess: (guess: string) => void
  addLetter: (letter: string) => void
  removeLetter: () => void
  submitGuess: () => Promise<void>
  clearError: () => void
  reset: () => void
}

function updateLetterStates(
  current: Record<string, LetterResult>,
  guess: Guess
): Record<string, LetterResult> {
  const updated = { ...current }
  const word = guess.word.toUpperCase()

  for (let i = 0; i < word.length; i++) {
    const letter = word[i]
    const result = guess.results[i]

    // Only upgrade: absent -> present -> correct
    if (result === 'correct') {
      updated[letter] = 'correct'
    } else if (result === 'present' && updated[letter] !== 'correct') {
      updated[letter] = 'present'
    } else if (result === 'absent' && !updated[letter]) {
      updated[letter] = 'absent'
    }
  }

  return updated
}

export const useGameStore = create<GameState>((set, get) => ({
  session: null,
  currentGuess: '',
  isLoading: false,
  error: null,
  letterStates: {},

  startGame: async (mode, language = 'en') => {
    set({ isLoading: true, error: null })
    try {
      const session = await game.start(mode, language)
      set({
        session,
        currentGuess: '',
        letterStates: {},
        isLoading: false,
      })
    } catch (err) {
      set({ error: (err as Error).message, isLoading: false })
    }
  },

  loadGame: async (sessionId) => {
    set({ isLoading: true, error: null })
    try {
      const session = await game.get(sessionId)

      // Rebuild letter states from existing guesses
      let letterStates: Record<string, LetterResult> = {}
      for (const guess of session.guesses) {
        letterStates = updateLetterStates(letterStates, guess)
      }

      set({ session, letterStates, currentGuess: '', isLoading: false })
    } catch (err) {
      set({ error: (err as Error).message, isLoading: false })
    }
  },

  setCurrentGuess: (guess) => {
    if (guess.length <= 5) {
      set({ currentGuess: guess.toUpperCase() })
    }
  },

  addLetter: (letter) => {
    const { currentGuess, session } = get()
    if (session?.status !== 'in_progress') return
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

    set({ isLoading: true, error: null })
    try {
      const result = await game.guess(session.session_id, currentGuess.toLowerCase())

      const newGuesses = [...session.guesses, result.guess]
      const newLetterStates = updateLetterStates(get().letterStates, result.guess)

      set({
        session: {
          ...session,
          guesses: newGuesses,
          status: result.game_status as GameStatus,
          target_word: result.target_word,
          attempts_used: newGuesses.length,
        },
        letterStates: newLetterStates,
        currentGuess: '',
        isLoading: false,
      })
    } catch (err) {
      set({ error: (err as Error).message, isLoading: false })
    }
  },

  clearError: () => set({ error: null }),

  reset: () => set({
    session: null,
    currentGuess: '',
    isLoading: false,
    error: null,
    letterStates: {},
  }),
}))
