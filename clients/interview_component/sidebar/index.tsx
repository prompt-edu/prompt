import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Mic } from 'lucide-react'

const interviewSidebarItems: SidebarMenuItemProps = {
  title: 'Interview',
  icon: <Mic />,
  goToPath: '',
  requiredPermissions: [
    Role.PROMPT_ADMIN,
    Role.COURSE_LECTURER,
    Role.COURSE_EDITOR,
    Role.COURSE_STUDENT,
  ],
  subitems: [
    {
      title: 'Manage Interviews',
      goToPath: '/manage',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Participants',
      goToPath: '/participants',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Schedule',
      goToPath: '/schedule',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Settings',
      goToPath: '/settings',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
  ],
}

export default interviewSidebarItems
