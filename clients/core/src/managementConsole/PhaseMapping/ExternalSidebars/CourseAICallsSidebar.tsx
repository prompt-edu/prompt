import { useAIEnabled } from '@core/network/hooks/useAIEnabled'
import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Sparkles } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'

export const CourseAICallsSidebar = ({ rootPath, title }: { rootPath: string; title: string }) => {
  const aiEnabled = useAIEnabled()

  if (!aiEnabled) {
    return null
  }

  const aiCallsSidebarItem: SidebarMenuItemProps = {
    title: 'AI Calls',
    icon: <Sparkles />,
    goToPath: '/ai-calls',
    requiredPermissions: [Role.PROMPT_ADMIN],
  }

  return (
    <ExternalSidebarComponent
      title={title}
      rootPath={rootPath}
      sidebarElement={aiCallsSidebarItem}
    />
  )
}
