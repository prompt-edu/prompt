import '../src/styles.css'

import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Presentation } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Presentations',
  icon: <Presentation />,
  goToPath: '',
  requiredPermissions: [
    Role.PROMPT_ADMIN,
    Role.COURSE_LECTURER,
    Role.COURSE_EDITOR,
    Role.COURSE_STUDENT,
  ],
  subitems: [
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

export default sidebarItems
