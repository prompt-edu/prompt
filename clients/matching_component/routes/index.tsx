import { type ExtendedRouteObject, LECTURER_ROLES } from '@tumaet/prompt-shared-state'
import { MatchingOverviewPage } from '../src/matching/MatchingOverviewPage'
import { DataExportPage } from '../src/matching/pages/DataExport/DataExportPage'
import { DataImportPage } from '../src/matching/pages/DataImport/DataImportPage'
import { MatchingParticipantsPage } from '../src/matching/pages/MatchingParticipantsPage/MatchingParticipantsPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <MatchingOverviewPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/export',
    element: <DataExportPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/re-import',
    element: <DataImportPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  {
    path: '/participants',
    element: <MatchingParticipantsPage />,
    requiredPermissions: LECTURER_ROLES,
  },
  // Add more routes here as needed
]

export default routes
