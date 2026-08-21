import {
  getPermissionString,
  Role,
  useAuthStore,
  useCourseStore,
} from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

export const useCoursePhaseId = (): string => {
  const { phaseId, coursePhaseId } = useParams<{ phaseId: string; coursePhaseId: string }>()
  return phaseId ?? coursePhaseId ?? ''
}

export const usePresentationAccess = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const permissions = useAuthStore((state) => state.permissions)
  const course = useCourseStore((state) => state.courses.find((item) => item.id === courseId))
  const isStudentOfCourse = useCourseStore((state) => state.isStudentOfCourse)
  const has = (role: Role) =>
    permissions.includes(
      role === Role.PROMPT_ADMIN || role === Role.PROMPT_LECTURER
        ? getPermissionString(role)
        : getPermissionString(role, course?.name, course?.semesterTag),
    )
  // A PROMPT lecturer is staff on every course, and the server treats them as such, so the
  // client has to as well. Otherwise they would be routed through the student-only endpoints.
  const isAdmin = has(Role.PROMPT_ADMIN)
  const isLecturer = isAdmin || has(Role.PROMPT_LECTURER) || has(Role.COURSE_LECTURER)
  const isEditor = isLecturer || has(Role.COURSE_EDITOR)

  return {
    isAdmin,
    isLecturer,
    isEditor,
    isStaff: isEditor,
    // Matches the server's manager roles: schedule and settings are course level, so a PROMPT
    // lecturer evaluating presentations still cannot change them.
    canManagePhase: isAdmin || has(Role.COURSE_LECTURER),
    isStudent: isStudentOfCourse(courseId ?? ''),
  }
}
