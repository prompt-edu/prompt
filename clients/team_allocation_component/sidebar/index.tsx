import { LECTURER_ROLES, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Users2 } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Team Allocation',
  icon: <Users2 />,
  goToPath: '',
  subitems: [
    {
      title: 'Participants',
      goToPath: '/participants',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Tease Configuration',
      goToPath: '/tease-config',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Allocations',
      goToPath: '/allocations',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Survey Statistics',
      goToPath: '/statistics',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Settings',
      goToPath: '/settings',
      requiredPermissions: LECTURER_ROLES,
    },
  ],
}

export default sidebarItems
