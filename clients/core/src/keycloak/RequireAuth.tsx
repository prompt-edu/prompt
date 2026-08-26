import { DarkModeProvider, LoadingPage } from '@tumaet/prompt-ui-components'
import { type ReactNode, useEffect, useRef } from 'react'
import { useKeycloak } from './useKeycloak'

// Gate for every surface that needs a session. Children mount only once a token is
// available, so no request can go out with a missing or expired one.
export const RequireAuth = ({ children }: { children: ReactNode }) => {
  const { isInitialized, isAuthenticated, login } = useKeycloak()
  const loginRequested = useRef(false)

  useEffect(() => {
    if (isAuthenticated) {
      loginRequested.current = false
      return
    }
    if (isInitialized && !loginRequested.current) {
      loginRequested.current = true
      login()
    }
  }, [isInitialized, isAuthenticated, login])

  if (!isInitialized || !isAuthenticated) {
    return (
      <DarkModeProvider>
        <LoadingPage />
      </DarkModeProvider>
    )
  }

  return <>{children}</>
}
