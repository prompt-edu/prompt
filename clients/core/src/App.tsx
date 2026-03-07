import { Suspense } from 'react'
import { LandingPage } from './publicPages/landingPage/LandingPage'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { KeycloakProvider } from './keycloak/KeycloakProvider'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { ManagementRoot } from './managementConsole/ManagementConsole'
import { TemplateRoutes } from './managementConsole/PhaseMapping/ExternalRoutes/TemplateRoutes'
import { PhaseRouterMapping } from './managementConsole/PhaseMapping/PhaseRouterMapping'
import PrivacyPage from './publicPages/legalPages/Privacy'
import ImprintPage from './publicPages/legalPages/Imprint'
import AboutPage from './publicPages/legalPages/AboutPage'
import { CourseOverview } from './managementConsole/courseOverview/CourseOverviewPage'
import { ApplicationLoginPage } from './publicPages/application/ApplicationLoginPage'
import { ApplicationAuthenticated } from './publicPages/application/pages/ApplicationAuthenticated/ApplicationAuthenticated'
import { Toaster } from '@tumaet/prompt-ui-components'
import CourseConfiguratorPage from './managementConsole/courseConfigurator/CourseConfiguratorPage'
import { PermissionRestriction } from './managementConsole/shared/components/PermissionRestriction'
import { Role } from '@tumaet/prompt-shared-state'
import { env } from '@/env'
import { parseURL } from './utils/parseURL'
import { CourseSettingsPage } from './managementConsole/courseSettings/CourseSettingsPage'
import { ActiveCoursesPage } from './managementConsole/pages/ActiveCoursesPage'
import { TemplateCoursesPage } from './managementConsole/pages/TemplateCoursesPage'
import { ArchivedCoursesPage } from './managementConsole/pages/ArchivedCoursesPage'
import { StudentsPage } from './managementConsole/pages/StudentsPage'
import { StudentDetailPage } from './managementConsole/pages/StudentDetailPage'
import { StudentNoteTagsPage } from './managementConsole/pages/InstructorNoteTagsPage'

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
              path='/management/course_templates'
              element={
                <ManagementRoot>
                  <TemplateCoursesPage />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course_archive'
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
                  <StudentsPage />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/students/:studentId'
              element={
                <ManagementRoot>
                  <StudentDetailPage />
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
              path='/management/course/:courseId/:phaseId/*'
              element={
                <ManagementRoot>
                  <PhaseRouterMapping />
                </ManagementRoot>
              }
            />
            <Route
              path='/management/course/:courseId/template_component/*'
              element={
                <ManagementRoot>
                  <Suspense fallback={<div>Fallback</div>}>
                    <TemplateRoutes />
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
