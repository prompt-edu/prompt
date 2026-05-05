import { useKeycloak } from '../keycloak/useKeycloak'
import { useAuthStore, useCourseStore } from '@tumaet/prompt-shared-state'
import { UnauthorizedPage } from '@tumaet/prompt-ui-components'
import { AppSidebar } from './layout/Sidebar/AppSidebar'
import { EmptyPage } from './shared/components/EmptyPage'
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
  LoadingPage,
  ErrorPage,
  Separator,
} from '@tumaet/prompt-ui-components'
import React, { useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Course } from '@tumaet/prompt-shared-state'
import { getAllCourses } from '../network/queries/course'
import { DarkModeProvider } from '@tumaet/prompt-ui-components'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import CourseNotFound from './shared/components/CourseNotFound'
import { Breadcrumbs } from './layout/Breadcrumbs/Breadcrumbs'
import { getOwnCourseIDs } from '@core/network/queries/ownCourseIDs'
import { NavUserMenu } from './layout/Sidebar/CourseSwitchSidebar/components/NavUserMenu'
import { Footer } from '@core/publicPages/shared/components/Footer'

export const ManagementRoot = ({ children }: { children?: React.ReactNode }) => {
  const { keycloak, logout } = useKeycloak()
  const { user, permissions } = useAuthStore()
  const { courseId } = useParams<{ courseId: string }>()
  const hasChildren = React.Children.count(children) > 0
  const path = useLocation().pathname

  const {
    setCourses,
    setOwnCourseIDs,
    getSelectedCourseID,
    setSelectedCourseID,
    removeSelectedCourseID,
  } = useCourseStore()
  const navigate = useNavigate()

  // getting the courses
  const {
    data: fetchedCourses,
    isPending,
    isError: isCourseError,
    refetch: refetchCourses,
  } = useQuery<Course[]>({
    queryKey: ['courses'],
    queryFn: () => getAllCourses(),
    enabled: !!keycloak,
  })

  // getting the course ids of the course a user is enrolled in
  const {
    data: fetchedOwnCourseIDs,
    isPending: isOwnCourseIdPending,
    isError: isOwnCourseIdError,
    refetch: refetchOwnCourseIds,
  } = useQuery<string[]>({
    queryKey: ['own_courses'],
    queryFn: () => getOwnCourseIDs(),
    enabled: !!keycloak,
  })

  const refetch = () => {
    refetchOwnCourseIds()
    refetchCourses()
  }
  const isLoading = !(keycloak && user) || isPending || isOwnCourseIdPending
  const isError = isCourseError || isOwnCourseIdError
  const courseExists = fetchedCourses?.some((course) => course.id === courseId)

  useEffect(() => {
    if (fetchedCourses) {
      // Spreading into a new array forces an immutable update.
      setCourses([...fetchedCourses])
    }
  }, [fetchedCourses, setCourses])

  useEffect(() => {
    if (fetchedOwnCourseIDs) {
      setOwnCourseIDs([...fetchedOwnCourseIDs])
    }
  }, [fetchedOwnCourseIDs, setOwnCourseIDs])

  useEffect(() => {
    if (!fetchedCourses) return
    if (path === '/management') {
      const retrievedCourseID = getSelectedCourseID()
      const selectedCourse = fetchedCourses.find((course) => course.id === retrievedCourseID)
      if (retrievedCourseID && selectedCourse !== undefined) {
        navigate(`/management/course/${retrievedCourseID}`)
      } else if (fetchedCourses.length === 1) {
        // Single-course users should go straight to their course overview.
        navigate(`/management/course/${fetchedCourses[0].id}`)
      } else if (fetchedCourses.length > 1) {
        // Multi-course users should see the actual management landing page.
        navigate('/management/courses')
      } else {
        removeSelectedCourseID()
      }
    } else if (
      [
        '/management/courses',
        '/management/course-templates',
        '/management/course-archive',
      ].includes(path) ||
      (courseId && !courseExists)
    ) {
      removeSelectedCourseID()
    } else if (courseId && courseExists) {
      setSelectedCourseID(courseId)
    }
  }, [
    path,
    fetchedCourses,
    courseId,
    courseExists,
    navigate,
    getSelectedCourseID,
    removeSelectedCourseID,
    setSelectedCourseID,
  ])

  if (isError) {
    return <ErrorPage onRetry={() => refetch()} onLogout={() => logout()} />
  }

  if (isLoading || !keycloak) {
    return (
      <DarkModeProvider>
        <LoadingPage />
      </DarkModeProvider>
    )
  }

  // Check if the user has at least some Prompt rights
  if (permissions.length === 0 && fetchedCourses && fetchedCourses.length === 0) {
    return <UnauthorizedPage onLogout={logout} />
  }

  return (
    <DarkModeProvider>
      <SidebarProvider
        style={
          {
            '--sidebar-width': 'calc(var(--sidebar-width-icon) + 222px + 1px)',
          } as React.CSSProperties
        }
      >
        <AppSidebar />
        <SidebarInset>
          {courseId && !courseExists && <CourseNotFound courseId={courseId || ''} />}
          <header className='sticky top-0 z-10 flex h-14 shrink-0 items-center justify-between gap-4 border-b bg-background px-4'>
            <div className='flex min-w-0 items-center gap-2'>
              <SidebarTrigger className='-ml-1' />
              <Separator orientation='vertical' className='mr-2 h-4' />
              <Breadcrumbs />
            </div>
            <NavUserMenu onLogout={() => logout()} />
          </header>
          <div
            id='management-children'
            className='flex min-h-0 flex-1 flex-col gap-4 overflow-auto p-6'
          >
            {hasChildren ? children : <EmptyPage />}
          </div>
          <Footer />
        </SidebarInset>
      </SidebarProvider>
    </DarkModeProvider>
  )
}
