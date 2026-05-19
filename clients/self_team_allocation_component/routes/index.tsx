import { Role } from '@tumaet/prompt-shared-state'
import { ExtendedRouteObject } from '@tumaet/prompt-shared-state'
import { SelfTeamAllocationPage } from '../src/self_team_allocation/pages/TeamAllocation/SelfTeamAllocationPage'
import { SelfTeamAllocationParticipantsPage } from '../src/self_team_allocation/pages/SelfTeamAllocationParticipantsPage/SelfTeamAllocationParticipantsPage'
import { SettingsPage } from '../src/self_team_allocation/pages/Settings/SettingsPage'

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
