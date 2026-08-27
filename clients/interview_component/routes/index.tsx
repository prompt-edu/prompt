import { type ExtendedRouteObject, LECTURER_ROLES, Role } from '@tumaet/prompt-shared-state'
import { InterviewDataShell } from '../src/interview/pages/InterviewDataShell'
import { InterviewParticipantsPage } from '../src/interview/pages/InterviewParticipantsPage/InterviewParticipantsPage'
import OverviewPage from '../src/interview/pages/Overview/OverviewPage'
import { ProfileDetailPage } from '../src/interview/pages/ProfileDetail/ProfileDetailPage'
import { InterviewScheduleManagement } from '../src/interview/pages/ScheduleManagement/InterviewScheduleManagement'
import { SettingsPage } from '../src/interview/pages/Settings/SettingsPage'
import { StudentInterviewPage } from '../src/interview/pages/StudentInterview/StudentInterviewPage'

const interviewRoutes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <StudentInterviewPage />,
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/manage',
    element: (
      <InterviewDataShell>
        <OverviewPage />
      </InterviewDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/manage/details/:studentId',
    element: (
      <InterviewDataShell>
        <ProfileDetailPage />
      </InterviewDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/participants',
    element: <InterviewParticipantsPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/schedule',
    element: <InterviewScheduleManagement />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/settings',
    element: (
      <InterviewDataShell>
        <SettingsPage />
      </InterviewDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
]

export default interviewRoutes
