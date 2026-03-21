import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { config } from '../config'
import { auth } from '../api/client'

export function redirectToLogin(): void {
  const params = new URLSearchParams({
    client_id: config.keycloak.clientId,
    redirect_uri: `${window.location.origin}/oauth/callback`,
    response_type: 'code',
    scope: 'openid',
  })
  window.location.href = `${config.keycloak.oidcBase}/auth?${params}`
}

export function redirectToRegister(): void {
  const params = new URLSearchParams({
    client_id: config.keycloak.clientId,
    redirect_uri: `${window.location.origin}/oauth/callback`,
    response_type: 'code',
    scope: 'openid',
  })
  window.location.href = `${config.keycloak.oidcBase}/registrations?${params}`
}

function parseUsername(token: string): string | null {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.preferred_username || null
  } catch {
    return null
  }
}

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  username: string | null
  isGuest: boolean
  isAuthenticated: boolean
  hasHydrated: boolean

  loginAsGuest: () => Promise<void>
  logout: () => Promise<void>
  clearAuth: () => void
  setTokens: (accessToken: string, refreshToken: string, isGuest?: boolean) => void
  setHasHydrated: (state: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      refreshToken: null,
      username: null,
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
          username: parseUsername(accessToken),
          isGuest,
          isAuthenticated: true,
        })
      },

      clearAuth: () => {
        set({
          accessToken: null,
          refreshToken: null,
          username: null,
          isGuest: false,
          isAuthenticated: false,
        })
      },

      loginAsGuest: async () => {
        const response = await auth.guest()
        get().setTokens(response.accessToken, '', true)
      },

      logout: async () => {
        const { refreshToken, isGuest } = get()
        if (refreshToken) {
          try {
            await auth.logout(refreshToken)
          } catch {
            // Ignore logout errors
          }
        }
        get().clearAuth()

        if (!isGuest) {
          const params = new URLSearchParams({
            client_id: config.keycloak.clientId,
            post_logout_redirect_uri: window.location.origin,
          })
          window.location.href = `${config.keycloak.oidcBase}/logout?${params}`
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        username: state.username,
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
