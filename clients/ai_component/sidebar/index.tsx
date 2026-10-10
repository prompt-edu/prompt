import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Sparkles } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'AI Calls',
  icon: <Sparkles />,
  goToPath: '',
  requiredPermissions: [Role.PROMPT_ADMIN],
}

export default sidebarItems
