import { ApplicationDetailsPage } from '@core/managementConsole/applicationAdministration/pages/ApplicationParticipantsPage/components/ApplicationDetailsDialog/ApplicationDetailsPage'
import { EDITOR_ROLES, type ExtendedRouteObject, LECTURER_ROLES } from '@tumaet/prompt-shared-state'
import { ApplicationDataWrapper } from '../../applicationAdministration/components/ApplicationDataWrapper'
import { ApplicationLandingPage } from '../../applicationAdministration/pages/ApplicationLandingPage/ApplicationLandingPage'
import { ApplicationParticipantsPage } from '../../applicationAdministration/pages/ApplicationParticipantsPage/ApplicationParticipantsPage'
import { ApplicationQuestionConfig } from '../../applicationAdministration/pages/ApplicationQuestionConfigPage/ApplicationQuestionConfig'
import { ApplicationConfiguration } from '../../applicationAdministration/pages/ApplicationSettingsPage/ApplicationSettings'
import { ExternalRoutes } from './ExternalRoutes'

const applicationRoutesObjects: ExtendedRouteObject[] = [
  {
    path: '',
    element: (
      <ApplicationDataWrapper>
        <ApplicationLandingPage />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: EDITOR_ROLES,
  },
  {
    path: '/settings',
    element: (
      <ApplicationDataWrapper>
        <ApplicationConfiguration />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/questions',
    element: (
      <ApplicationDataWrapper>
        <ApplicationQuestionConfig />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/participants',
    element: (
      <ApplicationDataWrapper>
        <ApplicationParticipantsPage />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/participants/:participationId',
    element: (
      <ApplicationDataWrapper>
        <ApplicationDetailsPage />
      </ApplicationDataWrapper>
    ),
    requiredPermissions: LECTURER_ROLES,
  },
]

export const ApplicationRoutes = () => {
  return <ExternalRoutes routes={applicationRoutesObjects} />
}
