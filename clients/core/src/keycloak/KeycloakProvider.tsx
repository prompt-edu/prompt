import { useAuthStore } from '@tumaet/prompt-shared-state'
import { jwtDecode } from 'jwt-decode'
import Keycloak from 'keycloak-js'
import type React from 'react'
import { createContext, type ReactNode, useCallback, useEffect, useRef, useState } from 'react'

const TOKEN_KEY = 'jwt_token'
const REFRESH_TOKEN_KEY = 'refreshToken'

interface KeycloakContextType {
  isInitialized: boolean
  isAuthenticated: boolean
  isLoggingOut: boolean
  initError: Error | null
  login: (redirectUri?: string) => void
  logout: () => Promise<void>
  forceTokenRefresh: () => Promise<void>
}

export const KeycloakContext = createContext<KeycloakContextType>({
  isInitialized: false,
  isAuthenticated: false,
  isLoggingOut: false,
  initError: null,
  login: () => {},
  logout: async () => {},
  forceTokenRefresh: async () => {},
})

const parseJwt = (token: string) => {
  try {
    return jwtDecode<{
      given_name: string
      family_name: string
      email: string
      preferred_username: string
      matriculation_number: string
      university_login: string
    }>(token)
  } catch {
    return null
  }
}

// keycloak-js rejects with an Error for transport failures but with the raw OIDC
// payload for a refused grant, so neither shape can be rendered on its own.
const toInitError = (reason: unknown): Error => {
  if (reason instanceof Error) return reason
  if (typeof reason === 'string') return new Error(reason)

  const payload = reason as { error?: string; error_description?: string } | null
  const description = [payload?.error, payload?.error_description].filter(Boolean).join(': ')
  return new Error(description || 'Keycloak could not be reached.')
}

const detachSessionCallbacks = (instance: Keycloak) => {
  instance.onAuthSuccess = undefined
  instance.onAuthRefreshSuccess = undefined
  instance.onAuthRefreshError = undefined
  instance.onAuthLogout = undefined
  instance.onTokenExpired = undefined
}

export const KeycloakProvider: React.FC<{
  keycloakUrl: string
  keycloakRealmName: string
  children: ReactNode
}> = ({ keycloakUrl, keycloakRealmName, children }) => {
  const keycloakRef = useRef<Keycloak | undefined>(undefined)
  const initStarted = useRef(false)
  const [isInitialized, setIsInitialized] = useState(false)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [isLoggingOut, setIsLoggingOut] = useState(false)
  const [initError, setInitError] = useState<Error | null>(null)
  const { setUser, setPermissions, clearUser, clearPermissions, setLogoutFunction } = useAuthStore()

  const clearLocalSession = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    clearUser()
    clearPermissions()
    setIsAuthenticated(false)
  }, [clearUser, clearPermissions])

  const storeSession = useCallback(
    (instance: Keycloak) => {
      const token = instance.token
      const decodedJwt = token ? parseJwt(token) : null
      if (!token || !decodedJwt) {
        clearLocalSession()
        return
      }

      localStorage.setItem(TOKEN_KEY, token)
      localStorage.setItem(REFRESH_TOKEN_KEY, instance.refreshToken ?? '')

      setUser({
        firstName: decodedJwt.given_name || '',
        lastName: decodedJwt.family_name || '',
        email: decodedJwt.email || '',
        username: decodedJwt.preferred_username || '',
        matriculationNumber: decodedJwt.matriculation_number || '',
        universityLogin: decodedJwt.university_login || '',
      })
      setPermissions(instance.resourceAccess?.['prompt-server']?.roles || [])
      setIsAuthenticated(true)
    },
    [setUser, setPermissions, clearLocalSession],
  )

  // Also drops the in-memory tokens, so the tab cannot silently refresh itself back
  // into an authenticated state after a failed refresh or a logout in another tab.
  const clearSession = useCallback(() => {
    keycloakRef.current?.clearToken()
    clearLocalSession()
  }, [clearLocalSession])

  const attachSessionCallbacks = useCallback(
    (instance: Keycloak) => {
      instance.onAuthSuccess = () => storeSession(instance)
      instance.onAuthRefreshSuccess = () => storeSession(instance)
      instance.onAuthRefreshError = () => clearSession()
      instance.onAuthLogout = () => clearLocalSession()
      instance.onTokenExpired = () => {
        void instance.updateToken(5).catch(() => undefined)
      }
    },
    [storeSession, clearSession, clearLocalSession],
  )

  useEffect(() => {
    if (initStarted.current) return
    initStarted.current = true

    const instance = new Keycloak({
      realm: keycloakRealmName,
      url: keycloakUrl,
      clientId: 'prompt-client',
    })
    keycloakRef.current = instance
    attachSessionCallbacks(instance)

    const storedToken = localStorage.getItem(TOKEN_KEY)
    const storedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)

    instance
      .init({
        checkLoginIframe: false,
        ...(storedToken && storedRefreshToken
          ? { token: storedToken, refreshToken: storedRefreshToken }
          : {}),
      })
      // A rejected or expired refresh token and an unreachable or misconfigured
      // realm both end here, and only the first can be resolved by logging in
      // again. Keeping the reason lets the caller show it once it stops retrying.
      .catch((error: unknown) => {
        setInitError(toInitError(error))
        clearSession()
      })
      .finally(() => setIsInitialized(true))
  }, [keycloakRealmName, keycloakUrl, attachSessionCallbacks, clearSession])

  // Another tab logged out: the shared tokens are gone, so this tab must not keep
  // rendering an authenticated UI (the shared axios reads localStorage per request).
  useEffect(() => {
    const onStorage = (event: StorageEvent) => {
      const tokensCleared = event.key === null
      const tokenRemoved =
        (event.key === TOKEN_KEY || event.key === REFRESH_TOKEN_KEY) && event.newValue === null
      if (tokensCleared || tokenRemoved) {
        clearSession()
      }
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [clearSession])

  const login = useCallback((redirectUri?: string) => {
    void keycloakRef.current?.login({ redirectUri: redirectUri ?? window.location.href })
  }, [])

  // isLoggingOut stops RequireAuth from answering the cleared session with a login redirect,
  // which would replace the logout redirect and let the still-active Keycloak SSO session
  // sign the user straight back in. The refresh callbacks are detached so a refresh landing
  // mid-redirect can neither write tokens back nor touch the auth state, and reattached if
  // the logout fails so a later login works without a reload. Keeps the in-memory
  // id token, which Keycloak needs as id_token_hint for the logout URL.
  const logout = useCallback(async () => {
    const instance = keycloakRef.current
    if (!instance) return

    detachSessionCallbacks(instance)
    setIsLoggingOut(true)
    clearLocalSession()

    try {
      await instance.logout({ redirectUri: window.location.origin })
    } catch (error) {
      attachSessionCallbacks(instance)
      setIsLoggingOut(false)
      throw error
    }
  }, [attachSessionCallbacks, clearLocalSession])

  const forceTokenRefresh = useCallback(async () => {
    const instance = keycloakRef.current
    if (!instance) {
      throw new Error('Keycloak instance is not initialized.')
    }
    await instance.updateToken(-1)
  }, [])

  useEffect(() => {
    setLogoutFunction(logout)
  }, [logout, setLogoutFunction])

  return (
    <KeycloakContext.Provider
      value={{
        isInitialized,
        isAuthenticated,
        isLoggingOut,
        initError,
        login,
        logout,
        forceTokenRefresh,
      }}
    >
      {children}
    </KeycloakContext.Provider>
  )
}
