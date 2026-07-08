import { type ExtendedRouteObject, Role } from '@tumaet/prompt-shared-state'
import { SelfTeamAllocationParticipantsPage } from '../src/self_team_allocation/pages/SelfTeamAllocationParticipantsPage/SelfTeamAllocationParticipantsPage'
import { SettingsPage } from '../src/self_team_allocation/pages/Settings/SettingsPage'
import { SelfTeamAllocationPage } from '../src/self_team_allocation/pages/TeamAllocation/SelfTeamAllocationPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <SelfTeamAllocationPage />,
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/participants',
    element: <SelfTeamAllocationParticipantsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/settings',
    element: <SettingsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
]

export default routes
