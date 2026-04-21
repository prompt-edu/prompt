import { SidebarMenuItemProps } from '@/interfaces/sidebar'
import { Role } from '@tumaet/prompt-shared-state'
import { GitBranch } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'DevOps Challenge',
  icon: <GitBranch />,
  goToPath: '',
  requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_STUDENT],
  subitems: [
    {
      title: 'Results Overview',
      goToPath: '/results-overview',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Mailing',
      goToPath: '/mailing',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
  ],
}

export default sidebarItems
