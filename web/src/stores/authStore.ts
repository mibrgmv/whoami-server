import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { auth } from '../api/client'
import { config } from '../config'
import { useGameStore } from './gameStore'

interface AuthState {
  accessToken: string | null
  username: string | null
  isGuest: boolean
  isAuthenticated: boolean
  hasHydrated: boolean

  init: () => Promise<void>
  loginAsGuest: () => Promise<void>
  logout: () => void
  clearAuth: () => void
  setGuestToken: (accessToken: string) => void
  setHasHydrated: (state: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      username: null,
      isGuest: false,
      isAuthenticated: false,
      hasHydrated: false,

      setHasHydrated: (state) => {
        set({ hasHydrated: state })
      },

      setGuestToken: (accessToken) => {
        set({
          accessToken,
          username: null,
          isGuest: true,
          isAuthenticated: true,
        })
      },

      clearAuth: () => {
        set({
          accessToken: null,
          username: null,
          isGuest: false,
          isAuthenticated: false,
        })
      },

      init: async () => {
        try {
          const me = await auth.me()
          if (me.isAuthenticated && !me.isGuest) {
            if (get().isGuest) {
              useGameStore.getState().clearRandomSession()
            }
            set({
              accessToken: null,
              username: me.username || null,
              isGuest: false,
              isAuthenticated: true,
            })
          } else if (!me.isAuthenticated) {
            const state = get()
            if (!state.isGuest) {
              get().clearAuth()
            }
          }
        } catch {
          if (!get().isGuest) {
            get().clearAuth()
          }
        }
      },

      loginAsGuest: async () => {
        const response = await auth.guest()
        get().setGuestToken(response.access_token)
      },

      logout: () => {
        const { isGuest } = get()
        get().clearAuth()
        useGameStore.getState().clearRandomSession()
        useGameStore.getState().clearDailyStatus()
        if (!isGuest) {
          window.location.href = `${config.apiBase}/auth/logout`
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        username: state.username,
        isGuest: state.isGuest,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        state?.init().finally(() => {
          state?.setHasHydrated(true)
        })
      },
    }
  )
)
