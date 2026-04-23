import { Puzzle } from 'lucide-react'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Role } from '@tumaet/prompt-shared-state'

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
