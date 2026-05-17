import { ProvidersPage } from '../src/infrastructure_setup/pages/ProvidersPage'
import { ResourceConfigPage } from '../src/infrastructure_setup/pages/ResourceConfigPage'
import { ExecutionPage } from '../src/infrastructure_setup/pages/ExecutionPage'
import { SetupConfigPage } from '../src/infrastructure_setup/pages/SetupConfigPage'
import { ExtendedRouteObject, Role } from '@tumaet/prompt-shared-state'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <ProvidersPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/setup',
    element: <SetupConfigPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/resource-configs',
    element: <ResourceConfigPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/execution',
    element: <ExecutionPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
]

export default routes
