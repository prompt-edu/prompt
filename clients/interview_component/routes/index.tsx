import { ExtendedRouteObject } from '@/interfaces/extendedRouteObject'
import { Role } from '@tumaet/prompt-shared-state'
import { InterviewDataShell } from '../src/interview/pages/InterviewDataShell'
import { StudentInterviewPage } from '../src/interview/pages/StudentInterview/StudentInterviewPage'
import OverviewPage from '../src/interview/pages/Overview/OverviewPage'
import { ProfileDetailPage } from '../src/interview/pages/ProfileDetail/ProfileDetailPage'
import { MailingPage } from '../src/interview/pages/Mailing/MailingPage'
import { QuestionConfiguration } from '../src/interview/pages/Settings/QuestionConfiguration'
import { InterviewScheduleManagement } from '../src/interview/pages/ScheduleManagement/InterviewScheduleManagement'
import { InterviewParticipantsPage } from '../src/interview/pages/InterviewParticipantsPage/InterviewParticipantsPage'

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
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/manage/details/:studentId',
    element: (
      <InterviewDataShell>
        <ProfileDetailPage />
      </InterviewDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/participants',
    element: <InterviewParticipantsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/mailing',
    element: (
      <InterviewDataShell>
        <MailingPage />
      </InterviewDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/schedule',
    element: <InterviewScheduleManagement />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/settings',
    element: (
      <InterviewDataShell>
        <QuestionConfiguration />
      </InterviewDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
]

export default interviewRoutes
