import {
  EDITOR_ROLES,
  type ExtendedRouteObject,
  LECTURER_ROLES,
  Role,
} from '@tumaet/prompt-shared-state'
import { ParticipantsPage } from '../src/certificate/pages/ParticipantsPage'
import { SettingsPage } from '../src/certificate/pages/SettingsPage'
import { StudentOverviewPage } from '../src/certificate/pages/StudentOverviewPage'

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
    requiredPermissions: EDITOR_ROLES,
  },
  {
    path: '/settings',
    element: <SettingsPage />,
    requiredPermissions: LECTURER_ROLES,
  },
]

export default routes
