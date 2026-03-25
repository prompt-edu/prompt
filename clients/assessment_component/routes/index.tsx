import { ExtendedRouteObject } from '@/interfaces/extendedRouteObject'
import { Role } from '@tumaet/prompt-shared-state'

import { EvaluationDataShell } from '../src/assessment/pages/EvaluationDataShell'
import { EvaluationOverviewPage } from '../src/assessment/pages/EvaluationOverviewPage/EvaluationOverviewPage'
import { EvaluationResultsPage } from '../src/assessment/pages/EvaluationOverviewPage/EvaluationResultsPage'
import { SelfEvaluationPage } from '../src/assessment/pages/EvaluationPages/SelfEvaluationPage'
import { PeerEvaluationPage } from '../src/assessment/pages/EvaluationPages/PeerEvaluationPage'
import { TutorEvaluationPage } from '../src/assessment/pages/EvaluationPages/TutorEvaluationPage'

import { AssessmentDataShell } from '../src/assessment/pages/AssessmentDataShell'
import { AssessmentParticipantsPage } from '../src/assessment/pages/AssessmentParticipantsPage/AssessmentParticipantsPage'
import { AssessmentPage } from '../src/assessment/pages/AssessmentPage/AssessmentPage'
import { AssessmentStatisticsPage } from '../src/assessment/pages/AssessmentStatisticsPage/AssessmentStatisticsPage'
import { TutorOverviewPage } from '../src/assessment/pages/TutorOverviewPage/TutorOverviewPage'
import { TutorEvaluationResultsPage } from '../src/assessment/pages/TutorEvaluationResultsPage/TutorEvaluationResultsPage'
import { SchemaConfigurationPage } from '../src/assessment/pages/SchemaConfigurationPage/SchemaConfigurationPage'
import { SettingsPage } from '../src/assessment/pages/SettingsPage/SettingsPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: (
      <EvaluationDataShell>
        <EvaluationOverviewPage />
      </EvaluationDataShell>
    ),
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/self-evaluation',
    element: (
      <EvaluationDataShell>
        <SelfEvaluationPage />
      </EvaluationDataShell>
    ),
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/results',
    element: (
      <EvaluationDataShell>
        <EvaluationResultsPage />
      </EvaluationDataShell>
    ),
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/peer-evaluation/:courseParticipationID',
    element: (
      <EvaluationDataShell>
        <PeerEvaluationPage />
      </EvaluationDataShell>
    ),
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/tutor-evaluation/:courseParticipationID',
    element: (
      <EvaluationDataShell>
        <TutorEvaluationPage />
      </EvaluationDataShell>
    ),
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
      <AssessmentDataShell>
        <AssessmentParticipantsPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_EDITOR],
  },
  {
    path: '/participants/:courseParticipationID',
    element: (
      <AssessmentDataShell>
        <AssessmentPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_EDITOR],
  },
  {
    path: '/statistics',
    element: (
      <AssessmentDataShell>
        <AssessmentStatisticsPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/tutors',
    element: (
      <AssessmentDataShell>
        <TutorOverviewPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/tutors/:tutorId',
    element: (
      <AssessmentDataShell>
        <TutorEvaluationResultsPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/settings',
    element: (
      <AssessmentDataShell>
        <SettingsPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/settings/schema/:schemaId',
    element: (
      <AssessmentDataShell>
        <SchemaConfigurationPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
]

export default routes
