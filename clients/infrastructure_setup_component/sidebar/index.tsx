import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Server } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Infrastructure Setup',
  icon: <Server />,
  goToPath: '',
  subitems: [
    {
      title: 'Overview',
      goToPath: '',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Setup',
      goToPath: '/setup',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Providers',
      goToPath: '/providers',
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
