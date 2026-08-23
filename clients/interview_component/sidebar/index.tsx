import { LECTURER_ROLES, Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
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
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Participants',
      goToPath: '/participants',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Schedule',
      goToPath: '/schedule',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Settings',
      goToPath: '/settings',
      requiredPermissions: LECTURER_ROLES,
    },
  ],
}

export default interviewSidebarItems
