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
  const has = (role: Role) =>
    permissions.includes(
      role === Role.PROMPT_ADMIN
        ? getPermissionString(role)
        : getPermissionString(role, course?.name, course?.semesterTag),
    )
  const isAdmin = has(Role.PROMPT_ADMIN)
  const isLecturer = isAdmin || has(Role.COURSE_LECTURER)
  const isEditor = isLecturer || has(Role.COURSE_EDITOR)

  return { isAdmin, isLecturer, isEditor, isStaff: isEditor }
}
