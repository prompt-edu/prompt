import { ExtendedRouteObject, Role } from '@tumaet/prompt-shared-state'
import { MatchingOverviewPage } from '../src/matching/MatchingOverviewPage'
import { DataExportPage } from '../src/matching/pages/DataExport/DataExportPage'
import { DataImportPage } from '../src/matching/pages/DataImport/DataImportPage'
import { MatchingParticipantsPage } from '../src/matching/pages/MatchingParticipantsPage/MatchingParticipantsPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <MatchingOverviewPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/export',
    element: <DataExportPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/re-import',
    element: <DataImportPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/participants',
    element: <MatchingParticipantsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  // Add more routes here as needed
]

export default routes
