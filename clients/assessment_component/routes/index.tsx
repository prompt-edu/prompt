import '../src/loadStyles'

import { type ExtendedRouteObject, Role } from '@tumaet/prompt-shared-state'
import { AssessmentType } from '../src/assessment/interfaces/assessmentType'
import { AssessmentDataShell } from '../src/assessment/pages/AssessmentDataShell'
import { AssessmentPage } from '../src/assessment/pages/AssessmentPage/AssessmentPage'
import { AssessmentParticipantsPage } from '../src/assessment/pages/AssessmentParticipantsPage/AssessmentParticipantsPage'
import { AssessmentStatisticsPage } from '../src/assessment/pages/AssessmentStatisticsPage/AssessmentStatisticsPage'
import { EvaluationDataShell } from '../src/assessment/pages/EvaluationDataShell'
import { EvaluationOverviewPage } from '../src/assessment/pages/EvaluationOverviewPage/EvaluationOverviewPage'
import { EvaluationResultsPage } from '../src/assessment/pages/EvaluationOverviewPage/EvaluationResultsPage'
import { PeerEvaluationPage } from '../src/assessment/pages/EvaluationPages/PeerEvaluationPage'
import { SelfEvaluationPage } from '../src/assessment/pages/EvaluationPages/SelfEvaluationPage'
import { TutorEvaluationPage } from '../src/assessment/pages/EvaluationPages/TutorEvaluationPage'
import { EvaluationParticipantResultsPage } from '../src/assessment/pages/EvaluationParticipantResultsPage/EvaluationParticipantResultsPage'
import { EvaluationParticipantsOverviewPage } from '../src/assessment/pages/EvaluationParticipantResultsPage/EvaluationParticipantsOverviewPage'
import { SchemaConfigurationPage } from '../src/assessment/pages/SchemaConfigurationPage/SchemaConfigurationPage'
import { SettingsPage } from '../src/assessment/pages/SettingsPage/SettingsPage'
import { TutorEvaluationResultsPage } from '../src/assessment/pages/TutorEvaluationResultsPage/TutorEvaluationResultsPage'
import { TutorOverviewPage } from '../src/assessment/pages/TutorOverviewPage/TutorOverviewPage'

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
    path: '/self-evaluations',
    element: (
      <AssessmentDataShell>
        <EvaluationParticipantsOverviewPage assessmentType={AssessmentType.SELF} />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/self-evaluations/:courseParticipationID',
    element: (
      <AssessmentDataShell>
        <EvaluationParticipantResultsPage assessmentType={AssessmentType.SELF} />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/peer-evaluations',
    element: (
      <AssessmentDataShell>
        <EvaluationParticipantsOverviewPage assessmentType={AssessmentType.PEER} />
      </AssessmentDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/peer-evaluations/:courseParticipationID',
    element: (
      <AssessmentDataShell>
        <EvaluationParticipantResultsPage assessmentType={AssessmentType.PEER} />
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
