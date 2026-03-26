import { useAuthStore, useCourseStore } from '@tumaet/prompt-shared-state'
import { Role, getPermissionString } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import UnauthorizedPage from '@/components/UnauthorizedPage'
import { useKeycloak } from '@core/keycloak/useKeycloak'
import { getCourseParticipation } from '@core/network/queries/courseParticipation'
import { CourseParticipation } from '../interfaces/CourseParticipation'
import { useQuery } from '@tanstack/react-query'
import { ErrorPage } from '@tumaet/prompt-ui-components'

interface PermissionRestrictionProps {
  requiredPermissions: Role[]
  children: React.ReactNode
}

// The server will only return data which the user is allowed to see
// This is only needed if the user has to restrict permission further to not show some pages at all (i.e. settings pages)
export const PermissionRestriction = ({
  requiredPermissions,
  children,
}: PermissionRestrictionProps) => {
  const { permissions } = useAuthStore()
  const { courses, isStudentOfCourse } = useCourseStore()
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const { logout } = useKeycloak()

  const isStudent = courseId !== '' && isStudentOfCourse(courseId ?? '')

  // get the current progression if the user is a student
  const {
    data: fetchedCourseParticipation,
    isPending: isCourseParticipationPending,
    isError: isCourseParticipationError,
    refetch: refetchCourseParticipation,
  } = useQuery<CourseParticipation>({
    queryKey: ['course_participation', courseId],
    queryFn: () => getCourseParticipation(courseId ?? ''),
    enabled: isStudent,
  })

  // This means something /general
  if (!courseId) {
    if (requiredPermissions.length === 0) {
      return <>{permissions.length > 0 ? children : <UnauthorizedPage onLogout={logout} />}</>
    }
    const hasPermission = requiredPermissions.some((role) =>
      permissions.includes(getPermissionString(role)),
    )
    return <>{hasPermission ? children : <UnauthorizedPage onLogout={logout} />}</>
  }

  if (isStudent && isCourseParticipationPending) {
    return <div>Loading...</div>
  }

  if (isStudent && isCourseParticipationError) {
    return <ErrorPage onRetry={refetchCourseParticipation} />
  }

  // in ManagementRoot is verified that this exists
  const course = courses.find((c) => c.id === courseId)

  let hasPermission = true
  if (requiredPermissions.length > 0) {
    hasPermission = requiredPermissions.some((role) => {
      return permissions.includes(getPermissionString(role, course?.name, course?.semesterTag))
    })

    // We need to compare student role with ownCourseIDs -> otherwise we could not hide pages from i.e. instructors
    // set hasPermission to true if the user is a student in the course and the page is accessible for students
    if (requiredPermissions.includes(Role.COURSE_STUDENT) && isStudentOfCourse(courseId)) {
      console.log(phaseId)
      if (phaseId === '') {
        hasPermission = true
      } else if (fetchedCourseParticipation) {
        hasPermission = fetchedCourseParticipation.activeCoursePhases.some(
          (coursePhaseID) => coursePhaseID === phaseId,
        )
      } else {
        hasPermission = false
      }
    }
  }

  return <>{hasPermission ? children : <UnauthorizedPage />}</>
}
