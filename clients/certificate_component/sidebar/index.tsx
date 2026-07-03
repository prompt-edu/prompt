import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Award } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Certificate',
  icon: <Award />,
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
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_EDITOR],
    },
    {
      title: 'Settings',
      goToPath: '/settings',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
  ],
}

export default sidebarItems
