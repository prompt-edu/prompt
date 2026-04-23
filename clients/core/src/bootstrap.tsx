import { createRoot } from 'react-dom/client'

import * as Sentry from '@sentry/react'
import { env } from '@tumaet/prompt-shared-state'

import App from './App'

// Initialize Sentry if DSN is provided
if (env.SENTRY_DSN_CLIENT) {
  Sentry.init({
    dsn: env.SENTRY_DSN_CLIENT,
    environment: env.ENVIRONMENT || 'development',
    sendDefaultPii: true,
  })
}

const rootElement = document.getElementById('root')

const root = createRoot(rootElement!)

root.render(<App />)
