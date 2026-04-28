import { createRoot } from 'react-dom/client'

import * as Sentry from '@sentry/react'
import { env } from '@tumaet/prompt-shared-state'

import App from './App'

declare const __LOCAL_SENTRY_DSN_CLIENT__: string
declare const __LOCAL_SENTRY_ENABLED__: string

const sentryEnv = env as typeof env & {
  SENTRY_DSN_CLIENT?: string
  SENTRY_ENABLED?: string
}

// Initialize Sentry if enabled and DSN is provided
const sentryEnabled = (sentryEnv.SENTRY_ENABLED || __LOCAL_SENTRY_ENABLED__) === 'true'
const sentryDsn = sentryEnv.SENTRY_DSN_CLIENT || __LOCAL_SENTRY_DSN_CLIENT__

if (sentryEnabled && sentryDsn) {
  Sentry.init({
    dsn: sentryDsn,
    environment: env.ENVIRONMENT || 'development',
    sendDefaultPii: true,
  })
}

const rootElement = document.getElementById('root')

const root = createRoot(rootElement!)

root.render(<App />)
