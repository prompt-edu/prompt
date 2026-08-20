import { type ReactNode, StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

export const mountRemote = (rootId: string, app: ReactNode) => {
  const container = document.getElementById(rootId)
  if (!container) {
    throw new Error(`mountRemote: no element with id '${rootId}'`)
  }

  createRoot(container).render(<StrictMode>{app}</StrictMode>)
}
