import { DarkModeProvider, ErrorPage, LoadingPage } from '@tumaet/prompt-ui-components'
import { type ReactNode, useEffect, useRef, useState } from 'react'
import { useKeycloak } from './useKeycloak'

// Survives the login redirect, so a session that never comes back authenticated
// cannot bounce the browser between here and Keycloak forever.
const LOGIN_ATTEMPTS_KEY = 'login_attempts'
const MAX_LOGIN_ATTEMPTS = 3

const readLoginAttempts = () => Number(sessionStorage.getItem(LOGIN_ATTEMPTS_KEY)) || 0

// Gate for every surface that needs a session. Children mount only once a token is
// available, so no request can go out with a missing or expired one.
export const RequireAuth = ({ children }: { children: ReactNode }) => {
  const { isInitialized, isAuthenticated, initError, login } = useKeycloak()
  const loginRequested = useRef(false)
  const [loginAttemptsExhausted, setLoginAttemptsExhausted] = useState(false)

  useEffect(() => {
    if (isAuthenticated) {
      sessionStorage.removeItem(LOGIN_ATTEMPTS_KEY)
      loginRequested.current = false
      return
    }
    if (!isInitialized || loginRequested.current) return
    loginRequested.current = true

    const attempts = readLoginAttempts()
    if (attempts >= MAX_LOGIN_ATTEMPTS) {
      setLoginAttemptsExhausted(true)
      return
    }
    sessionStorage.setItem(LOGIN_ATTEMPTS_KEY, String(attempts + 1))
    login()
  }, [isInitialized, isAuthenticated, login])

  if (loginAttemptsExhausted) {
    return (
      <DarkModeProvider>
        <ErrorPage
          title='Login failed'
          description='We could not sign you in. Please try again, and contact the course organizers if this keeps happening.'
          message={initError?.message}
          onRetry={() => {
            sessionStorage.removeItem(LOGIN_ATTEMPTS_KEY)
            login()
          }}
        />
      </DarkModeProvider>
    )
  }

  if (!isInitialized || !isAuthenticated) {
    return (
      <DarkModeProvider>
        <LoadingPage />
      </DarkModeProvider>
    )
  }

  return <>{children}</>
}
