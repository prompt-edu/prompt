import React from 'react'
import { ExtendedRouteObject } from '@tumaet/prompt-shared-state'
import { ExternalRoutes } from './ExternalRoutes'
import { LoadingError } from '../utils/LoadingError'

/** We use this style with a separate loading file for better performance */
/** It would be possible to have one loading script and pass the import path as variable */
/** but this requires a dictionary for static compilation + leads to re-renders every time */
/** Hence this way allows for better UI expierence */
export const TemplateRoutes = React.lazy(() =>
  import('template_component/routes')
    .then((module): { default: React.FC } => ({
      default: () => {
        const routes: ExtendedRouteObject[] = module.default || []
        return <ExternalRoutes routes={routes} />
      },
    }))
    .catch((): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load template routes')
        return <LoadingError phaseTitle={'Template'} />
      },
    })),
)
