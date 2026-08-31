import {
  getPermissionString,
  Role,
  useAuthStore,
  useCourseStore,
} from '@tumaet/prompt-shared-state'

// Sending and resending are restricted to lecturers/admins; course editors may
// only edit drafts. Mirrors the send-role gating enforced by the server.
const sendRoles = [Role.PROMPT_ADMIN, Role.PROMPT_LECTURER, Role.COURSE_LECTURER]

export const useCanSendMailCampaign = (courseID: string | undefined): boolean => {
  const { permissions } = useAuthStore()
  const { courses } = useCourseStore()
  const course = courses.find((c) => c.id === courseID)
  return sendRoles.some((role) =>
    permissions.includes(getPermissionString(role, course?.name, course?.semesterTag)),
  )
}
