const KEYCLOAK_REALM = import.meta.env.VITE_KEYCLOAK_REALM || 'gordle-realm'
const KEYCLOAK_CLIENT_ID = import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'gordle-public'

export const config = {
  apiBase: import.meta.env.VITE_API_BASE || '/api/v1',

  keycloak: {
    realm: KEYCLOAK_REALM,
    clientId: KEYCLOAK_CLIENT_ID,
    oidcBase: `/realms/${KEYCLOAK_REALM}/protocol/openid-connect`,
    accountUrl: `/realms/${KEYCLOAK_REALM}/account/`,
  },
} as const
