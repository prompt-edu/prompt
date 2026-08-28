import { LECTURER_ROLES, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Construction } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'ExampleComponent',
  icon: <Construction />,
  goToPath: '',
  subitems: [
    {
      title: 'Participants',
      goToPath: '/participants',
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
