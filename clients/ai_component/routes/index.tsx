import { type ExtendedRouteObject, Role } from '@tumaet/prompt-shared-state'
import { AICallsPage } from '../src/ai_component/pages/AICallsPage'

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: <AICallsPage />,
    requiredPermissions: [Role.PROMPT_ADMIN],
  },
]

export default routes
