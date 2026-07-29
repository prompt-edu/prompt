import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { env, Role } from '@tumaet/prompt-shared-state'
import { Toaster } from '@tumaet/prompt-ui-components'
import { Suspense } from 'react'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { KeycloakProvider } from './keycloak/KeycloakProvider'
import CourseConfiguratorPage from './managementConsole/courseConfigurator/CourseConfiguratorPage'
import { CourseOverview } from './managementConsole/courseOverview/CourseOverviewPage'
import { CourseSettingsPage } from './managementConsole/courseSettings/CourseSettingsPage'
import { CourseUserManagementPage } from './managementConsole/courseUserManagement/pages/CourseUserManagementPage'
import { ManagementRoot } from './managementConsole/ManagementConsole'
import { ExampleRoutes } from './managementConsole/PhaseMapping/ExternalRoutes/ExampleRoutes'
import { PhaseRouterMapping } from './managementConsole/PhaseMapping/PhaseRouterMapping'
import { ActiveCoursesPage } from './managementConsole/pages/ActiveCoursesPage'
import { AdminPrivacyPage } from './managementConsole/pages/AdminPrivacyPage'
import { ArchivedCoursesPage } from './managementConsole/pages/ArchivedCoursesPage'
import { StudentNoteTagsPage } from './managementConsole/pages/InstructorNoteTagsPage'
import { PrivacyDataDeletionPage } from './managementConsole/pages/PrivacyDataDeletionPage'
import { PrivacyDataExportPage } from './managementConsole/pages/PrivacyDataExportPage'
import { PrivacyOverviewPage } from './managementConsole/pages/PrivacyOverviewPage'
import { StudentDetailPage } from './managementConsole/pages/StudentDetailPage'
import { StudentsPage } from './managementConsole/pages/StudentsPage'
import { SystemStatusPage } from './managementConsole/pages/SystemStatusPage/SystemStatusPage'
import { TemplateCoursesPage } from './managementConsole/pages/TemplateCoursesPage'
import { PermissionRestriction } from './managementConsole/shared/components/PermissionRestriction'
import { ApplicationLoginPage } from './publicPages/application/ApplicationLoginPage'
import { ApplicationAuthenticated } from './publicPages/application/pages/ApplicationAuthenticated/ApplicationAuthenticated'
import { LandingPage } from './publicPages/landingPage/LandingPage'
import AboutPage from './publicPages/legalPages/AboutPage'
import ImprintPage from './publicPages/legalPages/Imprint'
import PrivacyPage from './publicPages/legalPages/Privacy'
import { parseURL } from './utils/parseURL'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
})

const keycloakUrl = parseURL(env.KEYCLOAK_HOST)
const keycloakRealmName = env.KEYCLOAK_REALM_NAME

export const App = () => {
  return (
    <KeycloakProvider keycloakRealmName={keycloakRealmName} keycloakUrl={keycloakUrl}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            <Route path='/about' element={<AboutPage />} />
            <Route path='/privacy' element={<PrivacyPage />} />
            <Route path='/imprint' element={<ImprintPage />} />
            <Route path='/' element={<LandingPage />} />
            <Route path='/apply/:phaseId' element={<ApplicationLoginPage />} />
            <Route path='/apply/:phaseId/authenticated' element={<ApplicationAuthenticated />} />
            <Route path='/management' element={<ManagementRoot />} />
            <Route
              path='/management/courses'
              element={
                <ManagementRoot>
                  <ActiveCoursesPage />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course-templates'
              element={
                <ManagementRoot>
                  <TemplateCoursesPage />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course-archive'
              element={
                <ManagementRoot>
                  <ArchivedCoursesPage />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course/:courseId'
              element={
                <ManagementRoot>
                  <CourseOverview />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/students'
              element={
                <ManagementRoot>
                  <PermissionRestriction
                    requiredPermissions={[Role.PROMPT_ADMIN, Role.PROMPT_LECTURER]}
                  >
                    <StudentsPage />
                  </PermissionRestriction>
                </ManagementRoot>
              }
            />
            <Route
              path='/management/students/:studentId'
              element={
                <ManagementRoot>
                  <PermissionRestriction requiredPermissions={[Role.PROMPT_ADMIN]}>
                    <StudentDetailPage />
                  </PermissionRestriction>
                </ManagementRoot>
              }
            />
            <Route
              path='/management/system-status'
              element={
                <ManagementRoot>
                  <PermissionRestriction requiredPermissions={[Role.PROMPT_ADMIN]}>
                    <SystemStatusPage />
                  </PermissionRestriction>
                </ManagementRoot>
              }
            />
            <Route
              path='/management/student-note-tags'
              element={
                <ManagementRoot>
                  <PermissionRestriction requiredPermissions={[Role.PROMPT_ADMIN]}>
                    <StudentNoteTagsPage />
                  </PermissionRestriction>
                </ManagementRoot>
              }
            />
            <Route
              path='/management/privacy'
              element={
                <ManagementRoot>
                  <PrivacyOverviewPage />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/privacy/data-export'
              element={
                <ManagementRoot>
                  <PrivacyDataExportPage />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/privacy/data-deletion'
              element={
                <ManagementRoot>
                  <PrivacyDataDeletionPage />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/admin/privacy'
              element={
                <ManagementRoot>
                  <PermissionRestriction requiredPermissions={[Role.PROMPT_ADMIN]}>
                    <AdminPrivacyPage />
                  </PermissionRestriction>
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course/:courseId/configurator'
              element={
                <ManagementRoot>
                  <PermissionRestriction
                    requiredPermissions={[
                      Role.PROMPT_ADMIN,
                      Role.COURSE_LECTURER,
                      Role.COURSE_EDITOR,
                    ]}
                  >
                    <CourseConfiguratorPage />
                  </PermissionRestriction>
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course/:courseId/settings'
              element={
                <ManagementRoot>
                  <PermissionRestriction
                    requiredPermissions={[Role.PROMPT_ADMIN, Role.COURSE_LECTURER]}
                  >
                    <CourseSettingsPage />
                  </PermissionRestriction>
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course/:courseId/user-management'
              element={
                <ManagementRoot>
                  <PermissionRestriction
                    requiredPermissions={[Role.PROMPT_ADMIN, Role.COURSE_LECTURER]}
                  >
                    <CourseUserManagementPage />
                  </PermissionRestriction>
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course/:courseId/:phaseId/*'
              element={
                <ManagementRoot>
                  <PhaseRouterMapping />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course/:courseId/example_component/*'
              element={
                <ManagementRoot>
                  <Suspense fallback={<div>Fallback</div>}>
                    <ExampleRoutes />
                  </Suspense>
                </ManagementRoot>
              }
            />
          </Routes>
          <Toaster />
        </BrowserRouter>
      </QueryClientProvider>
    </KeycloakProvider>
  )
}

export default App
