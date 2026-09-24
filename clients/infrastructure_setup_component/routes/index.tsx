import { type ExtendedRouteObject, Role } from '@tumaet/prompt-shared-state'
import { ConfigurationPage } from '../src/infrastructure_setup/pages/ConfigurationPage'
import { ParticipantsPage } from '../src/infrastructure_setup/pages/ParticipantsPage'
import { ProvisioningPage } from '../src/infrastructure_setup/pages/ProvisioningPage'
import { StudentResourcesPage } from '../src/infrastructure_setup/pages/StudentResourcesPage'

// The phase's API carries external provider credentials, so it admits admins and
// lecturers only; editors are deliberately left out of the management pages too.
const MANAGER_ROLES = [Role.PROMPT_ADMIN, Role.COURSE_LECTURER]

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <StudentResourcesPage />,
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/participants',
    element: <ParticipantsPage />,
    requiredPermissions: MANAGER_ROLES,
  },
  {
    path: '/configuration',
    element: <ConfigurationPage />,
    requiredPermissions: MANAGER_ROLES,
  },
  {
    path: '/provisioning',
    element: <ProvisioningPage />,
    requiredPermissions: MANAGER_ROLES,
  },
]

export default routes
