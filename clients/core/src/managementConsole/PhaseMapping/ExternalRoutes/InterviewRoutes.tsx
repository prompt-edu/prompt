import React from 'react'
import { ExtendedRouteObject } from '@tumaet/prompt-shared-state'
import { ExternalRoutes } from './ExternalRoutes'
import { LoadingError } from '../utils/LoadingError'

export const InterviewRoutes = React.lazy(() =>
  import('interview_component/routes')
    .then((module): { default: React.FC } => ({
      default: () => {
        const routes: ExtendedRouteObject[] = module.default || []
        return <ExternalRoutes routes={routes} />
      },
    }))
    .catch((error): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load interview routes')
        console.warn(error)
        return <LoadingError phaseTitle={'Interview'} />
      },
    })),
)
