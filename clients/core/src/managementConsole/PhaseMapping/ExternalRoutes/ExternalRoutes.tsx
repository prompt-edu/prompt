import { ExtendedRouteObject } from '@tumaet/prompt-shared-state'
import { Route, Routes } from 'react-router-dom'
import ErrorBoundary from '../../shared/components/ErrorBoundary'
import { PermissionRestriction } from '../../shared/components/PermissionRestriction'

interface ExternalRoutesProps {
  routes: ExtendedRouteObject[]
}

export const ExternalRoutes: React.FC<ExternalRoutesProps> = ({ routes }: ExternalRoutesProps) => {
  return (
    <>
      <Routes>
        {routes.map((route, index) => (
          <Route
            key={index}
            path={route.path}
            element={
              <PermissionRestriction requiredPermissions={route.requiredPermissions || []}>
                <ErrorBoundary fallback={<div>Route loading failed</div>}>
                  {route.element}
                </ErrorBoundary>
              </PermissionRestriction>
            }
          />
        ))}
      </Routes>
    </>
  )
}
