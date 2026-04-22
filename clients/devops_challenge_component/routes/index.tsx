import { ExtendedRouteObject } from '@/interfaces/extendedRouteObject'
import { Role } from '@tumaet/prompt-shared-state'
import { DevOpsChallengePage } from '../src/devops_challenge/DevOpsChallengePage'
import { DevOpsChallengeDataShell } from '../src/devops_challenge/DevOpsChallengeDataShell'
import { ResultsOverviewPage } from '../src/devops_challenge/pages/ResultsOverviewPage/ResultsOverviewPage'
import { MailingPage } from '../src/devops_challenge/pages/Mailing/MailingPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: (
      <DevOpsChallengeDataShell>
        <DevOpsChallengePage />
      </DevOpsChallengeDataShell>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_STUDENT],
  },
  {
    path: '/results-overview',
    element: <ResultsOverviewPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/mailing',
    element: <MailingPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
]

export default routes
