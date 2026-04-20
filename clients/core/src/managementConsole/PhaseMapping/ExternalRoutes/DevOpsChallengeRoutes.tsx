import React from 'react'
import { ExtendedRouteObject } from '@/interfaces/extendedRouteObject'
import { ExternalRoutes } from './ExternalRoutes'
import { LoadingError } from '../utils/LoadingError'

export const DevOpsChallengeRoutes = React.lazy(() =>
  import('devops_challenge_component/routes')
    .then((module): { default: React.FC } => ({
      default: () => {
        const routes: ExtendedRouteObject[] = module.default || []
        return <ExternalRoutes routes={routes} />
      },
    }))
    .catch((): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load challenge routes')
        return <LoadingError phaseTitle={'DevOps Challenge'} />
      },
    })),
)
