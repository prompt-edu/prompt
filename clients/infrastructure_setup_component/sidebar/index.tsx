import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Server } from 'lucide-react'

// The same roles as the management routes: the phase's API admits admins and lecturers.
const MANAGER_ROLES = [Role.PROMPT_ADMIN, Role.COURSE_LECTURER]

const sidebarItems: SidebarMenuItemProps = {
  title: 'Infrastructure Setup',
  icon: <Server />,
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
      requiredPermissions: MANAGER_ROLES,
    },
    {
      title: 'Configuration',
      goToPath: '/configuration',
      requiredPermissions: MANAGER_ROLES,
    },
    {
      title: 'Provisioning',
      goToPath: '/provisioning',
      requiredPermissions: MANAGER_ROLES,
    },
  ],
}

export default sidebarItems
