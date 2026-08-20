import { LECTURER_ROLES, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Puzzle } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Matching',
  icon: <Puzzle />,
  requiredPermissions: LECTURER_ROLES,
  goToPath: '',
  subitems: [
    {
      title: 'Participants',
      goToPath: '/participants',
      requiredPermissions: LECTURER_ROLES,
    },
  ],
}

export default sidebarItems
