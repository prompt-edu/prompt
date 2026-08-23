import { LECTURER_ROLES, Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Users2 } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Team Allocation',
  icon: <Users2 />,
  goToPath: '',
  requiredPermissions: [
    Role.PROMPT_ADMIN,
    Role.COURSE_LECTURER,
    Role.COURSE_EDITOR,
    Role.COURSE_STUDENT,
  ],
  subitems: [
    {
      title: 'Participants',
      goToPath: '/participants',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Survey Settings',
      goToPath: '/settings',
      requiredPermissions: LECTURER_ROLES,
    },
  ],
}

export default sidebarItems
