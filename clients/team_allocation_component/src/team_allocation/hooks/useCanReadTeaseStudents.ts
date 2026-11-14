import {
  getPermissionString,
  Role,
  useAuthStore,
  useCourseStore,
} from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

export const useCanReadTeaseStudents = (): boolean => {
  const { courseId } = useParams<{ courseId: string }>()
  const { courses } = useCourseStore()
  const { permissions } = useAuthStore()
  const course = courses.find((c) => c.id === courseId)

  return (
    permissions.includes(getPermissionString(Role.PROMPT_ADMIN)) ||
    permissions.includes(
      getPermissionString(Role.COURSE_LECTURER, course?.name, course?.semesterTag),
    )
  )
}
