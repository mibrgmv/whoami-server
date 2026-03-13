import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { auth } from '../api/client'
import type { LoginRequest, RegisterRequest } from '../types/api'

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  isGuest: boolean
  isAuthenticated: boolean

  login: (data: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<void>
  loginAsGuest: () => Promise<void>
  logout: () => Promise<void>
  setTokens: (accessToken: string, refreshToken: string, isGuest?: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      refreshToken: null,
      isGuest: false,
      isAuthenticated: false,

      setTokens: (accessToken, refreshToken, isGuest = false) => {
        localStorage.setItem('access_token', accessToken)
        set({
          accessToken,
          refreshToken,
          isGuest,
          isAuthenticated: true,
        })
      },

      login: async (data) => {
        const response = await auth.login(data)
        get().setTokens(response.access_token, response.refresh_token, false)
      },

      register: async (data) => {
        await auth.register(data)
        await get().login({ username: data.username, password: data.password })
      },

      loginAsGuest: async () => {
        const response = await auth.guest()
        get().setTokens(response.access_token, response.refresh_token, true)
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
        localStorage.removeItem('access_token')
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
    }
  )
)
