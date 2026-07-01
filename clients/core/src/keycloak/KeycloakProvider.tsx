import type Keycloak from 'keycloak-js'
import React, { createContext, ReactNode, useState } from 'react'

interface KeycloakContextType {
  keycloakUrl: string
  keycloakRealmName: string
  keycloakValue: Keycloak | undefined
  setKeycloakValue: (keycloak: Keycloak) => void
}

export const KeycloakContext = createContext<KeycloakContextType>({
  keycloakUrl: '',
  keycloakRealmName: '',
  keycloakValue: undefined,
  setKeycloakValue: () => {},
})

export const KeycloakProvider: React.FC<{
  keycloakUrl: string
  keycloakRealmName: string
  children: ReactNode
}> = ({ keycloakUrl, keycloakRealmName, children }) => {
  const [keycloakValue, setKeycloakValue] = useState<Keycloak>()

  return (
    <KeycloakContext.Provider
      value={{ keycloakUrl, keycloakRealmName, keycloakValue, setKeycloakValue }}
    >
      {children}
    </KeycloakContext.Provider>
  )
}
