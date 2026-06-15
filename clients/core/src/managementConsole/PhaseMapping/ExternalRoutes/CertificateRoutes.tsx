import React from 'react'
import { ExtendedRouteObject } from '@tumaet/prompt-shared-state'
import { ExternalRoutes } from './ExternalRoutes'
import { LoadingError } from '../utils/LoadingError'

export const CertificateRoutes = React.lazy(() =>
  import('certificate_component/routes')
    .then((module): { default: React.FC } => ({
      default: () => {
        const routes: ExtendedRouteObject[] = module.default || []
        return <ExternalRoutes routes={routes} />
      },
    }))
    .catch((err): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load certificate routes', err)
        return <LoadingError phaseTitle={'Certificate'} />
      },
    })),
)
