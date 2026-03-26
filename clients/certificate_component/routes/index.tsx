import { ExtendedRouteObject } from '@/interfaces/extendedRouteObject'
import { Role } from '@tumaet/prompt-shared-state'

import { StudentOverviewPage } from '../src/certificate/pages/StudentOverviewPage'
import { ParticipantsPage } from '../src/certificate/pages/ParticipantsPage'
import { SettingsPage } from '../src/certificate/pages/SettingsPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <StudentOverviewPage />,
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
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_EDITOR],
  },
  {
    path: '/settings',
    element: <SettingsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
]

export default routes
