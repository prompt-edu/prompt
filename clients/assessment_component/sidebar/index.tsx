import '../src/loadStyles'

import { ClipboardList } from 'lucide-react'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Role } from '@tumaet/prompt-shared-state'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Assessment Component',
  icon: <ClipboardList />,
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
      title: 'Tutor Overview',
      goToPath: '/tutors',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Statistics',
      goToPath: '/statistics',
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
