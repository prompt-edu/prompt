import '../src/styles.css'

import { type ExtendedRouteObject, Role } from '@tumaet/prompt-shared-state'
import { lazy, Suspense } from 'react'

const OverviewPage = lazy(() => import('../src/presentation/pages/OverviewPage'))
const FeedbackWorkspacePage = lazy(() => import('../src/presentation/pages/FeedbackWorkspacePage'))
const SchedulePage = lazy(() => import('../src/presentation/pages/SchedulePage'))
const SettingsPage = lazy(() => import('../src/presentation/pages/SettingsPage'))

const fallback = <div className='p-6 text-muted-foreground'>Loading presentations…</div>

const routes: ExtendedRouteObject[] = [
  {
    path: '',
    element: (
      <Suspense fallback={fallback}>
        <OverviewPage />
      </Suspense>
    ),
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/presentations/:presentationId',
    element: (
      <Suspense fallback={fallback}>
        <FeedbackWorkspacePage />
      </Suspense>
    ),
    requiredPermissions: [
      Role.PROMPT_ADMIN,
      Role.COURSE_LECTURER,
      Role.COURSE_EDITOR,
      Role.COURSE_STUDENT,
    ],
  },
  {
    path: '/schedule',
    element: (
      <Suspense fallback={fallback}>
        <SchedulePage />
      </Suspense>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
  {
    path: '/settings',
    element: (
      <Suspense fallback={fallback}>
        <SettingsPage />
      </Suspense>
    ),
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  },
]

export default routes
