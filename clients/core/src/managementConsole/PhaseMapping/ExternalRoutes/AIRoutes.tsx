import { useAIEnabled } from '@core/network/hooks/useAIEnabled'
import type { ExtendedRouteObject } from '@tumaet/prompt-shared-state'
import { Alert, AlertDescription, AlertTitle } from '@tumaet/prompt-ui-components'
import { Info } from 'lucide-react'
import React from 'react'
import { loadAIRemote } from '../aiRemote'
import { LoadingError } from '../utils/LoadingError'
import { ExternalRoutes } from './ExternalRoutes'

const AIRemoteRoutes = React.lazy(() =>
  loadAIRemote<typeof import('ai_component/routes')>('routes')
    .then((module): { default: React.FC } => ({
      default: () => {
        const routes: ExtendedRouteObject[] = module.default || []
        return <ExternalRoutes routes={routes} />
      },
    }))
    .catch((): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load AI routes')
        return <LoadingError phaseTitle={'AI'} />
      },
    })),
)

const AINotAvailable = () => (
  <Alert>
    <Info className='h-4 w-4' />
    <AlertTitle>AI is not available</AlertTitle>
    <AlertDescription>
      The AI service is turned off or unreachable, so its pages cannot be shown.
    </AlertDescription>
  </Alert>
)

export const AIRoutes = () => {
  const aiEnabled = useAIEnabled()
  if (aiEnabled === undefined) return null
  return aiEnabled ? <AIRemoteRoutes /> : <AINotAvailable />
}
