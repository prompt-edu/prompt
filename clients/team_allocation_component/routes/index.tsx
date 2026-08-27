import { type ExtendedRouteObject, LECTURER_ROLES, Role } from '@tumaet/prompt-shared-state'
import { StudentSurveyPage } from '../src/team_allocation/pages/StudentSurvey/StudentSurveyPage'
import { SurveySettingsPage } from '../src/team_allocation/pages/SurveySettings/SurveySettingsPage'
import { SurveyStatisticsPage } from '../src/team_allocation/pages/SurveyStatistics/SurveyStatisticsPage'
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
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/tease-config',
    element: <TeaseConfigPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/allocations',
    element: <TeamAllocationPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/participants',
    element: <TeamAllocationParticipantsPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/statistics',
    element: <SurveyStatisticsPage />,
    requiredPermissions: LECTURER_ROLES,
  },
]

export default routes
