import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Puzzle } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Matching',
  icon: <Puzzle />,
  requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  goToPath: '',
  subitems: [
    {
      title: 'Participants',
      goToPath: '/participants',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
  ],
}

export default sidebarItems
