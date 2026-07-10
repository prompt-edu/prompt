import type { ExtendedRouteObject } from '@tumaet/prompt-shared-state'
import React from 'react'
import { LoadingError } from '../utils/LoadingError'
import { ExternalRoutes } from './ExternalRoutes'

export const InfrastructureSetupRoutes = React.lazy(() =>
  import('infrastructure_setup_component/routes')
    .then((module): { default: React.FC } => ({
      default: () => {
        const routes: ExtendedRouteObject[] = module.default || []
        return <ExternalRoutes routes={routes} />
      },
    }))
    .catch((): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load infrastructure setup routes')
        return <LoadingError phaseTitle={'Infrastructure Setup'} />
      },
    })),
)
