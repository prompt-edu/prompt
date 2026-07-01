import { Role, SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Settings } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'

export const CourseSettingsSidebar = ({ rootPath, title }: { rootPath: string; title: string }) => {
  const courseConfiguratorSidebarItems: SidebarMenuItemProps = {
    title: 'Settings',
    icon: <Settings />,
    goToPath: '/settings',
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
