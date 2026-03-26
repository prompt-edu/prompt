import { ExtendedRouteObject } from '@/interfaces/extendedRouteObject'
import { Role } from '@tumaet/prompt-shared-state'
import { ApplicationLandingPage } from '../../applicationAdministration/pages/ApplicationLandingPage/ApplicationLandingPage'
import { ApplicationConfiguration } from '../../applicationAdministration/pages/ApplicationSettingsPage/ApplicationSettings'
import { ApplicationQuestionConfig } from '../../applicationAdministration/pages/ApplicationQuestionConfigPage/ApplicationQuestionConfig'
import { ExternalRoutes } from './ExternalRoutes'
import { ApplicationParticipantsPage } from '../../applicationAdministration/pages/ApplicationParticipantsPage/ApplicationParticipantsPage'
import { ApplicationMailingSettings } from '../../applicationAdministration/pages/Mailing/ApplicationMailingSettings'
import { ApplicationDataWrapper } from '../../applicationAdministration/components/ApplicationDataWrapper'
// eslint-disable-next-line max-len
import { ApplicationDetailsPage } from '@core/managementConsole/applicationAdministration/pages/ApplicationParticipantsPage/components/ApplicationDetailsDialog/ApplicationDetailsPage'

const applicationRoutesObjects: ExtendedRouteObject[] = [
  {
    path: '',
    element: (
      <ApplicationDataWrapper>
        <ApplicationLandingPage />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_EDITOR], // empty means no permissions required
  },
  {
    path: '/settings',
    element: (
      <ApplicationDataWrapper>
        <ApplicationConfiguration />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/questions',
    element: (
      <ApplicationDataWrapper>
        <ApplicationQuestionConfig />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/participants',
    element: (
      <ApplicationDataWrapper>
        <ApplicationParticipantsPage />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/participants/:participationId',
    element: (
      <ApplicationDataWrapper>
        <ApplicationDetailsPage />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/mailing',
    element: (
      <ApplicationDataWrapper>
        <ApplicationMailingSettings />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
]

export const ApplicationRoutes = () => {
  return <ExternalRoutes routes={applicationRoutesObjects} />
}
