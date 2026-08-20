import { type ExtendedRouteObject, LECTURER_ROLES } from '@tumaet/prompt-shared-state'
import { ParticipantsPage } from '../src/example_component/pages/ParticipantsPage'
import { SettingsPage } from '../src/example_component/pages/SettingsPage'
import { OverviewPage } from '../src/OverviewPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <OverviewPage />,
    requiredPermissions: [], // empty means no permissions required
  },
  {
    path: '/participants',
    element: <ParticipantsPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/settings',
    element: <SettingsPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  // Add more routes here as needed
]

export default routes
