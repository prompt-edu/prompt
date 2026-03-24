import { createContext, type ReactNode, useContext } from 'react'

import type { SettingsPageController } from '../hooks/useSettingsPageController'

const SettingsPageContext = createContext<SettingsPageController | undefined>(undefined)

interface SettingsPageProviderProps {
  value: SettingsPageController
  children: ReactNode
}

export const SettingsPageProvider = ({ value, children }: SettingsPageProviderProps) => {
  return <SettingsPageContext.Provider value={value}>{children}</SettingsPageContext.Provider>
}

export const useSettingsPageContext = (): SettingsPageController => {
  const context = useContext(SettingsPageContext)
  if (!context) {
    throw new Error('useSettingsPageContext must be used within SettingsPageProvider')
  }

  return context
}

