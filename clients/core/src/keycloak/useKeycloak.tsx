import { useContext, useEffect, useCallback } from 'react'
import Keycloak from 'keycloak-js'
import { KeycloakContext } from './KeycloakProvider'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import { jwtDecode } from 'jwt-decode'

// Helper function to decode JWT safely
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

export const useKeycloak = (): {
  keycloak: Keycloak | undefined
  logout: () => void
  forceTokenRefresh: () => Promise<void>
} => {
  const context = useContext(KeycloakContext)
  const { setUser, setPermissions, clearUser, clearPermissions, setLogoutFunction } = useAuthStore()

  if (!context) {
    throw new Error('useKeycloak must be used within a KeycloakProvider')
  }

  const { keycloakUrl, keycloakRealmName, keycloakValue } = context

  // re-set app state from incoming token
  const syncStateFromToken = useCallback(
    (keycloak: Keycloak) => {
      localStorage.setItem('jwt_token', keycloak.token ?? '')
      localStorage.setItem('refreshToken', keycloak.refreshToken ?? '')

      if (keycloak.token) {
        const decodedJwt = parseJwt(keycloak.token)
        if (decodedJwt) {
          setUser({
            firstName: decodedJwt.given_name || '',
            lastName: decodedJwt.family_name || '',
            email: decodedJwt.email || '',
            username: decodedJwt.preferred_username || '',
            matriculationNumber: decodedJwt.matriculation_number || '',
            universityLogin: decodedJwt.university_login || '',
          })
        } else {
          clearUser()
        }
        const resourceRoles = keycloak.resourceAccess?.['prompt-server']?.roles || []
        setPermissions(resourceRoles)
      }
    },
    [setUser, clearUser, setPermissions],
  )

  const initializeKeycloak = useCallback(() => {
    const keycloak = new Keycloak({
      realm: keycloakRealmName,
      url: keycloakUrl,
      clientId: 'prompt-client',
    })

    keycloak.onTokenExpired = () => {
      keycloak
        .updateToken(5)
        .then(() => syncStateFromToken(keycloak))
        .catch(() => {
          clearUser()
          clearPermissions()
          alert('Session expired. Please log in again.')
          keycloak.logout({ redirectUri: window.location.origin })
        })
    }

    void keycloak
      .init({ onLoad: 'login-required' })
      .then(() => {
        context.keycloakValue = keycloak // Update context dynamically
        syncStateFromToken(keycloak)
      })
      .catch((err) => {
        clearUser()
        clearPermissions()
        alert(`Authentication error: ${err.message}`)
      })

    return keycloak
  }, [context, clearUser, clearPermissions, syncStateFromToken, keycloakRealmName, keycloakUrl])

  useEffect(() => {
    if (!keycloakValue) {
      initializeKeycloak()
    }
  }, [keycloakValue, initializeKeycloak])

  const logout = useCallback(async () => {
    if (keycloakValue) {
      await keycloakValue.logout({ redirectUri: window.location.origin }) // Use provided URI or fallback
    }
    clearUser()
    clearPermissions()
    localStorage.removeItem('jwt_token')
    localStorage.removeItem('refreshToken')
  }, [clearPermissions, clearUser, keycloakValue])

  useEffect(() => {
    setLogoutFunction(logout) // Inject the logout function into the store
  }, [logout, setLogoutFunction])

  const forceTokenRefresh = (): Promise<void> => {
    return new Promise((resolve, reject) => {
      if (keycloakValue) {
        keycloakValue
          .updateToken(1000000) // Force immediate refresh
          .then(() => {
            // re-sync all token-derived state so newly added roles apply without a reload
            syncStateFromToken(keycloakValue)
            resolve() // Resolve the promise on success
          })
          .catch((err) => {
            console.error('Token refresh failed', err)
            clearUser()
            clearPermissions()
            keycloakValue.logout({ redirectUri: window.location.origin })
            reject(err) // Reject the promise on failure
          })
      } else {
        console.warn('Keycloak instance is not initialized.')
        reject(new Error('Keycloak instance is not initialized.'))
      }
    })
  }

  return { keycloak: keycloakValue, logout, forceTokenRefresh }
}
