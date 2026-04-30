import { Server } from 'lucide-react'
import { SidebarMenuItemProps, Role } from '@tumaet/prompt-shared-state'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Infrastructure Setup',
  icon: <Server />,
  goToPath: '',
  subitems: [
    {
      title: 'Providers',
      goToPath: '',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Resource Configs',
      goToPath: '/resource-configs',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Execution',
      goToPath: '/execution',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
  ],
}

export default sidebarItems
