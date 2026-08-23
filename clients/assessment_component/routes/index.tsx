import '../src/loadStyles'

import {
  EDITOR_ROLES,
  type ExtendedRouteObject,
  LECTURER_ROLES,
  Role,
} from '@tumaet/prompt-shared-state'
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
    requiredPermissions: EDITOR_ROLES,
  },
  {
    path: '/participants/:courseParticipationID',
    element: (
      <AssessmentDataShell>
        <AssessmentPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: EDITOR_ROLES,
  },
  {
    path: '/statistics',
    element: (
      <AssessmentDataShell>
        <AssessmentStatisticsPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/tutors',
    element: (
      <AssessmentDataShell>
        <TutorOverviewPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/tutors/:tutorId',
    element: (
      <AssessmentDataShell>
        <TutorEvaluationResultsPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/self-evaluations',
    element: (
      <AssessmentDataShell>
        <EvaluationParticipantsOverviewPage assessmentType={AssessmentType.SELF} />
      </AssessmentDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/self-evaluations/:courseParticipationID',
    element: (
      <AssessmentDataShell>
        <EvaluationParticipantResultsPage assessmentType={AssessmentType.SELF} />
      </AssessmentDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/peer-evaluations',
    element: (
      <AssessmentDataShell>
        <EvaluationParticipantsOverviewPage assessmentType={AssessmentType.PEER} />
      </AssessmentDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/peer-evaluations/:courseParticipationID',
    element: (
      <AssessmentDataShell>
        <EvaluationParticipantResultsPage assessmentType={AssessmentType.PEER} />
      </AssessmentDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/settings',
    element: (
      <AssessmentDataShell>
        <SettingsPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/settings/schema/:schemaId',
    element: (
      <AssessmentDataShell>
        <SchemaConfigurationPage />
      </AssessmentDataShell>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
]

export default routes
