import React from 'react'
import { ExtendedRouteObject } from '@tumaet/prompt-shared-state'
import { ExternalRoutes } from './ExternalRoutes'
import { LoadingError } from '../utils/LoadingError'

export const IntroCourseDeveloperRoutes = React.lazy(() =>
  import('intro_course_developer_component/routes')
    .then((module): { default: React.FC } => ({
      default: () => {
        const routes: ExtendedRouteObject[] = module.default || []
        return <ExternalRoutes routes={routes} />
      },
    }))
    .catch((error): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load intro course developer routes')
        console.warn(error)
        return <LoadingError phaseTitle={'Intro Course Developer'} />
      },
    })),
)
