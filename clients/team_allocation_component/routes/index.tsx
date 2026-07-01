import { ExtendedRouteObject, Role } from '@tumaet/prompt-shared-state'
import { StudentSurveyPage } from '../src/team_allocation/pages/StudentSurvey/StudentSurveyPage'
import { SurveySettingsPage } from '../src/team_allocation/pages/SurveySettings/SurveySettingsPage'
import { TeamAllocationPage } from '../src/team_allocation/pages/TeamAllocation/TeamAllocationPage'
import { TeamAllocationParticipantsPage } from '../src/team_allocation/pages/TeamAllocationParticipantsPage/TeamAllocationParticipantsPage'
import { TeaseConfigPage } from '../src/team_allocation/pages/TeaseConfig/TeaseConfigPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <StudentSurveyPage />,
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/settings',
    element: <SurveySettingsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/tease-config',
    element: <TeaseConfigPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/allocations',
    element: <TeamAllocationPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/participants',
    element: <TeamAllocationParticipantsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
]

export default routes
