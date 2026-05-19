import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Role, useAuthStore, useCourseStore } from '@tumaet/prompt-shared-state'
import { InsideSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/InsideSidebarMenuItem'
import { getPermissionString } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import { CourseParticipation } from '@core/managementConsole/shared/interfaces/CourseParticipation'
import { getCourseParticipation } from '@core/network/queries/courseParticipation'
import { useQuery } from '@tanstack/react-query'
import { ErrorPage } from '@tumaet/prompt-ui-components'

interface ExternalSidebarProps {
  rootPath: string
  title?: string
  sidebarElement: SidebarMenuItemProps
  coursePhaseID?: string
}

export const ExternalSidebarComponent: React.FC<ExternalSidebarProps> = ({
  title,
  rootPath,
  sidebarElement,
  coursePhaseID,
}: ExternalSidebarProps) => {
  // Example of using a custom hook
  const { permissions } = useAuthStore() // Example of calling your custom hook
  const { courses, isStudentOfCourse } = useCourseStore()
  const courseId = useParams<{ courseId: string }>().courseId

  const course = courses.find((c) => c.id === courseId)

  // get the current progression if the user is a student
  const {
    data: fetchedCourseParticipation,
    isError: isCourseParticipationError,
    refetch: refetchCourseParticipation,
  } = useQuery<CourseParticipation>({
    queryKey: ['course_participation', courseId],
    queryFn: () => getCourseParticipation(courseId ?? ''),
  })

  let hasComponentPermission = false
  if (sidebarElement.requiredPermissions && sidebarElement.requiredPermissions.length > 0) {
    // checks if user has access through keycloak roles
    hasComponentPermission = sidebarElement.requiredPermissions.some((role) => {
      return permissions.includes(getPermissionString(role, course?.name, course?.semesterTag))
    })

    // case that user is only student
    if (
      !hasComponentPermission &&
      coursePhaseID && // some sidebar items (i.e. Mailing are not connected to a phase)
      sidebarElement.requiredPermissions.includes(Role.COURSE_STUDENT) &&
      isStudentOfCourse(courseId ?? '') &&
      fetchedCourseParticipation
    ) {
      hasComponentPermission = fetchedCourseParticipation.activeCoursePhases.some(
        (phaseID) => phaseID === coursePhaseID,
      )
    }
  } else {
    // no permissions required
    hasComponentPermission = true
  }

  // we ignore this error if the user has access anyway
  if (isCourseParticipationError && !hasComponentPermission) {
    return (
      <>
        <ErrorPage
          message='Failed to get the course participation data'
          onRetry={refetchCourseParticipation}
        />
      </>
    )
  }

  return (
    <>
      {hasComponentPermission && (
        <InsideSidebarMenuItem
          title={title || sidebarElement.title}
          icon={sidebarElement.icon}
          goToPath={rootPath + sidebarElement.goToPath}
          subitems={
            sidebarElement.subitems
              ?.filter((subitem) => {
                const hasPermission = subitem.requiredPermissions?.some((role) => {
                  return permissions.includes(
                    getPermissionString(role, course?.name, course?.semesterTag),
                  )
                })
                if (subitem.requiredPermissions && !hasPermission) {
                  return false
                } else {
                  return true
                }
              })
              .map((subitem) => {
                return {
                  title: subitem.title,
                  goToPath: rootPath + subitem.goToPath,
                }
              }) || []
          }
        />
      )}
    </>
  )
}
