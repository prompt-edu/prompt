import { createRoot } from 'react-dom/client'

import * as Sentry from '@sentry/react'
import { env } from '@/env'

import App from './App'

// Initialize Sentry if enabled and DSN is provided
if (env.SENTRY_ENABLED === 'true' && env.SENTRY_DSN_CLIENT) {
  Sentry.init({
    dsn: env.SENTRY_DSN_CLIENT,
    environment: env.ENVIRONMENT || 'development',
    sendDefaultPii: true,
  })
}

const rootElement = document.getElementById('root')

const root = createRoot(rootElement!)

root.render(<App />)
