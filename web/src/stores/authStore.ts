import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { auth } from '../api/client'
import type { LoginRequest, RegisterRequest } from '../types/api'

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  isGuest: boolean
  isAuthenticated: boolean
  hasHydrated: boolean

  login: (data: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<void>
  loginAsGuest: () => Promise<void>
  logout: () => Promise<void>
  setTokens: (accessToken: string, refreshToken: string, isGuest?: boolean) => void
  setHasHydrated: (state: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      refreshToken: null,
      isGuest: false,
      isAuthenticated: false,
      hasHydrated: false,

      setHasHydrated: (state) => {
        set({ hasHydrated: state })
      },

      setTokens: (accessToken, refreshToken, isGuest = false) => {
        set({
          accessToken,
          refreshToken,
          isGuest,
          isAuthenticated: true,
        })
      },

      login: async (data) => {
        const response = await auth.login(data)
        get().setTokens(response.accessToken, response.refreshToken, false)
      },

      register: async (data) => {
        await auth.register(data)
        try {
          await get().login({ username: data.username, password: data.password })
        } catch (err) {
          const message = (err as Error).message || ''
          if (message.includes('email not verified')) {
            throw new Error('email_not_verified')
          }
          throw err
        }
      },

      loginAsGuest: async () => {
        const response = await auth.guest()
        get().setTokens(response.accessToken, '', true)
      },

      logout: async () => {
        const { refreshToken } = get()
        if (refreshToken) {
          try {
            await auth.logout(refreshToken)
          } catch {
            // Ignore logout errors
          }
        }
        set({
          accessToken: null,
          refreshToken: null,
          isGuest: false,
          isAuthenticated: false,
        })
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isGuest: state.isGuest,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state?.isAuthenticated && !state.isGuest && !state.refreshToken) {
          state.accessToken = null
          state.refreshToken = null
          state.isGuest = false
          state.isAuthenticated = false
        }
        state?.setHasHydrated(true)
      },
    }
  )
)
