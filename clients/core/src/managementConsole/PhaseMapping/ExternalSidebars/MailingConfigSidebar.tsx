import { Role } from '@tumaet/prompt-shared-state'
import { Mail } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'

export const MailingConfigSidebar = ({ rootPath, title }: { rootPath: string; title: string }) => {
  const courseConfiguratorSidebarItems: SidebarMenuItemProps = {
    title: 'Mailing Settings',
    icon: <Mail />,
    goToPath: '/mailing',
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  }
  return (
    <ExternalSidebarComponent
      title={title}
      rootPath={rootPath}
      sidebarElement={courseConfiguratorSidebarItems}
    />
  )
}
