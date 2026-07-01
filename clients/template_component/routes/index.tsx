import { ExtendedRouteObject, Role } from '@tumaet/prompt-shared-state'
import { OverviewPage } from '../src/OverviewPage'
import { ParticipantsPage } from '../src/template_component/pages/ParticipantsPage'
import { SettingsPage } from '../src/template_component/pages/SettingsPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <OverviewPage />,
    requiredPermissions: [], // empty means no permissions required
  },
  {
    path: '/participants',
    element: <ParticipantsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/settings',
    element: <SettingsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  // Add more routes here as needed
]

export default routes
