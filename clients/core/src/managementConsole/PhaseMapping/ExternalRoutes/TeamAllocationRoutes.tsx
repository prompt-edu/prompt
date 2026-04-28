import React from 'react'
import { ExtendedRouteObject } from '@tumaet/prompt-shared-state'
import { ExternalRoutes } from './ExternalRoutes'
import { LoadingError } from '../utils/LoadingError'

export const TeamAllocationRoutes = React.lazy(() =>
  import('team_allocation_component/routes')
    .then((module): { default: React.FC } => ({
      default: () => {
        const routes: ExtendedRouteObject[] = module.default || []
        return <ExternalRoutes routes={routes} />
      },
    }))
    .catch((): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load team allocation routes')
        return <LoadingError phaseTitle={'Team Allocation'} />
      },
    })),
)
