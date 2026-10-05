import { type ExtendedRouteObject, LECTURER_ROLES, Role } from '@tumaet/prompt-shared-state'
import { Navigate, useParams } from 'react-router-dom'
import { InterviewDataShell } from '../src/interview/pages/InterviewDataShell'
import { InterviewParticipantsPage } from '../src/interview/pages/InterviewParticipantsPage/InterviewParticipantsPage'
import { ProfileDetailPage } from '../src/interview/pages/ProfileDetail/ProfileDetailPage'
import { InterviewScheduleManagement } from '../src/interview/pages/ScheduleManagement/InterviewScheduleManagement'
import { SettingsPage } from '../src/interview/pages/Settings/SettingsPage'
import { StudentInterviewPage } from '../src/interview/pages/StudentInterview/StudentInterviewPage'

const LegacyManageRedirect = () => {
  const { courseId, phaseId, studentId } = useParams<{
    courseId: string
    phaseId: string
    studentId?: string
  }>()
  const participantsPath = `/management/course/${courseId}/${phaseId}/participants`
  return <Navigate to={studentId ? `${participantsPath}/${studentId}` : participantsPath} replace />
}

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
    path: '/participants',
    element: (
      <InterviewDataShell>
        <InterviewParticipantsPage />
      </InterviewDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/participants/:studentId',
    element: (
      <InterviewDataShell>
        <ProfileDetailPage />
      </InterviewDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/manage',
    element: <LegacyManageRedirect />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/manage/:studentId',
    element: <LegacyManageRedirect />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/manage/details/:studentId',
    element: <LegacyManageRedirect />,
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
